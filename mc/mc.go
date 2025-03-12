// Copyright 2023 Fortio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Multicurl package is a library for the multicurl tool to fetch a url from all its IPs.
// Some of this code is based on the fortio code.
// https://github.com/fortio/fortio/blob/master/fnet/network.go
package mc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/progressbar"
	"fortio.org/version"
)

// Config object for MultiCurl to avoid passing too many parameters.
type Config struct {
	// URL is the url to access.
	URL string
	// ResolveType is a filter for the IPs to use, `ip4`, `ip6`, or `ip` for both.
	ResolveType string
	// Method is the HTTP method to use, defaults to GET. (do use POST/PUT/... if passing a Payload)
	Method string
	// RequestTimeout is the timeout for a single request to succeed by.
	// Pass `context.WithTimeout(context.Background(), totalTimeout)` as context to MultiCurl() for a total timeout.
	RequestTimeout time.Duration
	// IncludeHeaders if true will include the response headers in the output.
	IncludeHeaders bool
	// Headers are the headers to use for the request. Must be initialized. Or call NewConfig().
	Headers http.Header
	// HostOverride is the host/authority to use for the request. If empty, the host from the URL is used.
	// This will also change the ServerName used for TLS handshake (sni) so in essence passing HostOverride
	// is the same as passing the IPs of the server of the url and using the name from HostOverride as the url.
	HostOverride string
	// OutputPattern is the pattern to use for the output file names, must contain a % which will get replaced by
	// the IP of the target. If empty or "-", output is written to stdout. If "none" no output is written.
	OutputPattern string
	// Payload to send or nil if none.
	Payload []byte
	// Source file of the IPs to use instead of resolving the host IPs. Use "-" to read from stdin.
	IPFile string
	// Expected http result code: other codes will count as errors. 0 (default) treats non 200 as warnings.
	ExpectedCode int
	// Repeat until no errors. 0 (default) means no repeat. -1 means repeat until no errors (context timeout still applies.
	// a positive number means repeat that at most that many times.
	MaxRepeat int
	// Delay between repeats. NewConfig will set this to 5 seconds as initial value.
	RepeatDelay time.Duration
	// Limit the number of IPs to use. 0 (default) means no limit.
	MaxIPs int
	// Re-Lookup between iterations. False by default. doesn't apply if IPFile is set.
	ReLookup bool
	// Cert expiration error threshold (if a cert is found expiring sooner than this).
	CertExpiryError time.Duration
	// Path of an alternate CA file to use for TLS validation (instead of system).
	CAFile string
	// Insecure will continue despite certificate validation errors.
	Insecure bool
	// Client certificate file path to provide to server for mutual TLS.
	Cert string
	// Client certificate key file path to provide to server for mutual TLS.
	Key string
	// Don't show progress bar (or spinner).
	NoProgressBar bool
	// extracted host
	host string
	// extracted port string
	port string
	// extracted port int
	portNum int
	// now (at start)
	now time.Time
}

// ResultStats is the details of the MultCurl run when any request is made at all.
type ResultStats struct {
	// Number of errors (if any request is made at all)
	Errors int
	// Number of warnings, ie non 200 responses
	Warnings int
	// Addresses queried (keys of Codes and Sizes)
	Addresses []string
	// http result code for that address (maps to Warnings)
	Codes map[string]int
	// Size of the response from that address
	Sizes map[string]int
	// Iterations done
	Iterations int
	// Shortest certificate expiration found
	ShortestCertExpiry *time.Time `json:"ShortestCertExpiry,omitempty"`
}

var (
	libShortVersion string
	libLongVersion  string
)

func init() {
	libShortVersion, libLongVersion, _ = version.FromBuildInfoPath("fortio.org/multicurl")
}

// MultiCurl is the main function of the multicurl tool. timeout is per request/ip.
// Returns 0 if all is successful, the number of errors otherwise.
// ResultStats is the details of the run (see ResultStats).
func MultiCurl(ctx context.Context, cfg *Config) (int, ResultStats) { //nolint:funlen // lots of needed validation/setup
	cfg.now = time.Now()
	log.Infof("Fortio multicurl %s, using resolver %s, %s %s", libLongVersion, cfg.ResolveType, cfg.Method, cfg.URL)
	result := ResultStats{
		Codes: make(map[string]int),
		Sizes: make(map[string]int),
	}
	if cfg.OutputPattern != "" && cfg.OutputPattern != "-" &&
		cfg.OutputPattern != "none" && !strings.Contains(cfg.OutputPattern, "%") {
		return log.FErrf("Output pattern must contain %% or be \"none\" or \"-\""), result
	}
	if len(cfg.URL) == 0 {
		return log.FErrf("Unexpected empty url"), result
	}
	urlString := URLAddScheme(cfg.URL)
	// Parse the url, extract components.
	url, err := url.Parse(urlString)
	if err != nil {
		return log.FErrf("Bad url %q : %v", urlString, err), result
	}
	cfg.host = url.Hostname()
	cfg.port = url.Port()
	if cfg.port == "" {
		cfg.port = url.Scheme // ie http / https which turns into 80 / 443 later
		log.LogVf("No port specified, using %s", cfg.port)
	}
	cfg.portNum, err = net.LookupPort("tcp", cfg.port)
	if err != nil {
		return log.FErrf("Unable to resolve port %q: %v", cfg.port, err), result
	}
	addrs, err := Resolve(ctx, cfg)
	if err != nil {
		return log.FErrf("Unable to resolve %s host %s: %v", cfg.ResolveType, cfg.host, err), result
	}
	req, err := http.NewRequestWithContext(ctx, cfg.Method, urlString, nil)
	req.Header = cfg.Headers
	req.Host = cfg.HostOverride
	if err != nil {
		return log.FErrf("Error creating request: %v", err), result
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DisableKeepAlives = true
	ca, err := GetCA(cfg.CAFile)
	if err != nil {
		return log.FErrf("%s", err), result
	}
	certs, err := GetCertificate(cfg.Cert, cfg.Key)
	if err != nil {
		return log.FErrf("LoadX509KeyPair error for cert %v / key %v: %v", cfg.Cert, cfg.Key, err), result
	}
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: cfg.Insecure, //nolint:gosec // on purpose with the flag/config
		RootCAs:            ca,
		Certificates:       certs,
		ServerName:         cfg.HostOverride,
	}
	hcli := http.Client{
		Transport: tr,
		Timeout:   cfg.RequestTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	result.Iterations = 1
	lastIterErrors := 0
	lastIterWarnings := 0
	for {
		lastIterErrors = 0
		lastIterWarnings = 0
		for idx, addr := range addrs {
			// humans start counting at 1
			nErr, nWarn := oneRequest(idx+1, cfg, &result, addr, req, tr, hcli)
			lastIterErrors += nErr
			lastIterWarnings += nWarn
			aStr := addr.String()
			if result.Iterations == 1 {
				// only save the addresses list once
				result.Addresses = append(result.Addresses, aStr)
			}
		}
		result.Errors += lastIterErrors
		result.Warnings += lastIterWarnings
		level := log.Info
		if result.Warnings > 0 {
			level = log.Warning
		}
		if result.Errors > 0 {
			level = log.Error
		}
		log.Logf(level, "[%d] %d %s (%d %s)", result.Iterations,
			lastIterErrors, cli.Plural(lastIterErrors, "error"),
			lastIterWarnings, cli.Plural(lastIterWarnings, "warning"))
		if lastIterErrors == 0 {
			break
		}
		if cfg.MaxRepeat >= 0 && result.Iterations > cfg.MaxRepeat {
			log.Errf("Reached max repeat %d", cfg.MaxRepeat)
			break
		}
		log.LogVf("Sleeping for %v before next iteration", cfg.RepeatDelay)
		select {
		case <-ctx.Done():
			log.Errf("Interrupted/total timeout reached")
			return lastIterErrors, result
		case <-time.After(cfg.RepeatDelay):
			// normal pause
		}
		if cfg.ReLookup && cfg.IPFile == "" {
			log.LogVf("Re-resolving %s host %s", cfg.ResolveType, cfg.host)
			addrs, err = Resolve(ctx, cfg)
			if err != nil {
				return log.FErrf("Unable to re-resolve %s host %s: %v", cfg.ResolveType, cfg.host, err), result
			}
		}
		result.Iterations++
	}
	if !checkCertExpiry(cfg, &result) {
		lastIterErrors++
	}
	return lastIterErrors, result
}

func GetCertificate(cert, key string) ([]tls.Certificate, error) {
	if cert == "" || key == "" {
		return nil, nil
	}
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	return []tls.Certificate{c}, nil
}

func GetCA(caFile string) (*x509.CertPool, error) {
	if caFile == "" {
		return nil, nil //nolint:nilnil // nil is perfectly valid pointer value for using default/system CAs
	}
	caCertPool := x509.NewCertPool()
	// Load CA cert from file
	caCert, err := os.ReadFile(caFile)
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if err != nil || !ok { // merge the two errors, mostly for coverage
		return nil, fmt.Errorf("can't read CA file: %w", err)
	}
	return caCertPool, nil
}

// checkCertExpiry returns true if expiry is ok (below error threshold).
func checkCertExpiry(cfg *Config, result *ResultStats) (good bool) {
	good = true
	if result.ShortestCertExpiry == nil {
		return
	}
	expiry := result.ShortestCertExpiry.Sub(cfg.now)
	level := log.Info
	if expiry < cfg.CertExpiryError {
		level = log.Error
		good = false
	}
	log.Logf(level, "Shortest cert expiry is %s (%.1f days from now)", result.ShortestCertExpiry, Days(expiry))
	return
}

func oneRequest(i int, cfg *Config, result *ResultStats, addr net.IP,
	req *http.Request, tr *http.Transport, cli http.Client,
) (int, int) {
	numWarnings := 0
	numErrors := 0
	log.LogVf("%d: Using %s", i, addr)
	if cfg.Payload != nil {
		// need to reset the body for each request
		log.LogVf("Using payload of %d bytes", len(cfg.Payload))
		req.Body = io.NopCloser(bytes.NewReader(cfg.Payload))
		req.ContentLength = int64(len(cfg.Payload)) // avoid chunked encoding, we already know the size
	}
	var out io.Writer
	switch cfg.OutputPattern {
	case "", "-":
		out = bufio.NewWriter(os.Stdout)
	case "none":
		out = io.Discard
	default:
		fname := Filename(cfg, addr)
		f, err := os.Create(fname)
		if err != nil {
			log.Errf("Error creating file %s: %v", fname, err)
			return 1, 0
		}
		defer f.Close()
		out = bufio.NewWriter(f)
		log.Infof("%d: Writing to %s", i, fname)
	}
	aStr := IPPortString(addr, cfg.portNum)
	tr.DialContext = func(ctx context.Context, network, oAddr string) (net.Conn, error) {
		log.LogVf("%d: DialContext %s %s -> %s", i, network, oAddr, aStr)
		d := net.Dialer{}
		c, err := d.DialContext(ctx, network, aStr)
		if c != nil {
			log.LogVf("%d: DialContext %v", i, c.LocalAddr())
		}
		return c, err
	}
	resp, err := cli.Do(req) //nolint:bodyclose // we do close it below
	req.Body = io.NopCloser(bytes.NewReader(cfg.Payload))
	if err != nil {
		log.Errf("%d: Error fetching %s: %v", i, addr, err)
		result.Codes[aStr] = -1
		return 1, 0
	}
	level := log.Info
	if cfg.ExpectedCode > 0 {
		if resp.StatusCode != cfg.ExpectedCode {
			level = log.Error
			numErrors++
		}
	} else if resp.StatusCode != http.StatusOK {
		level = log.Warning
		numWarnings++
	}
	log.Logf(level, "%d: Status %d %q from %s", i, resp.StatusCode, resp.Status, addr)
	if resp.TLS != nil {
		// Print certificate expiration date
		for _, cert := range resp.TLS.PeerCertificates {
			if result.ShortestCertExpiry == nil || cert.NotAfter.Before(*result.ShortestCertExpiry) {
				result.ShortestCertExpiry = &cert.NotAfter
			}
			durDays := Days(cert.NotAfter.Sub(cfg.now))
			log.Infof("Certificate %q expires in %.0f days", cert.Subject, durDays)
		}
	}
	if cfg.IncludeHeaders {
		DumpResponseDetails(out, resp)
	}
	reader := resp.Body
	if !cfg.NoProgressBar {
		bar := progressbar.NewBar()
		bar.Prefix = fmt.Sprintf("%2d ", i)
		bar.NoAnsi = !log.Color
		reader = progressbar.NewAutoReader(bar, resp.Body, resp.ContentLength)
	}
	data, err := io.ReadAll(reader)
	_ = reader.Close() // will close resp.Body too when using the progressbar wrapper.
	if err != nil {
		log.Errf("%d: Error reading body from %s: %v", i, addr, err)
		numErrors++
	}
	_, _ = out.Write(data)
	if f, ok := out.(*bufio.Writer); ok {
		f.Flush()
	}
	// will be the last iteration's results
	result.Codes[aStr] = resp.StatusCode
	result.Sizes[aStr] = len(data)
	return numErrors, numWarnings
}

func URLAddScheme(url string) string {
	log.LogVf("URLSchemeCheck %q", url)
	lcURL := strings.ToLower(url)
	if strings.HasPrefix(lcURL, "https://") {
		return url
	}
	if strings.HasPrefix(lcURL, "http://") {
		return url
	}
	log.LogVf("Assuming http:// on missing scheme for %q", url)
	return "http://" + url
}

func Resolve(ctx context.Context, cfg *Config) ([]net.IP, error) {
	if cfg.IPFile != "" {
		addrs, err := ReadIPs(cfg.IPFile)
		if err != nil {
			return nil, err // already logged
		}
		n := len(addrs)
		log.Infof("Resolved %s %s:%s to port %d and %d %s %v - from file %s",
			cfg.ResolveType, cfg.host, cfg.port, cfg.portNum, n, cli.PluralExt(n, "address", "es"), addrs, cfg.IPFile)
		return addrs, nil
	}
	log.LogVf("Resolving %s host %s (port %s -> %d)", cfg.ResolveType, cfg.host, cfg.port, cfg.portNum)
	addrs, err := ResolveAll(ctx, cfg.host, cfg.ResolveType)
	if err != nil {
		return nil, err
	}
	n := len(addrs)
	if cfg.MaxIPs > 0 && n > cfg.MaxIPs {
		log.Infof("Resolved %s %s:%s to port %d and %d %s %v - keeping first %d",
			cfg.ResolveType, cfg.host, cfg.port, cfg.portNum, n, cli.PluralExt(n, "address", "es"), addrs, cfg.MaxIPs)
		addrs = addrs[:cfg.MaxIPs]
	} else {
		log.Infof("Resolved %s %s:%s to port %d and %d %s %v",
			cfg.ResolveType, cfg.host, cfg.port, cfg.portNum, n, cli.PluralExt(n, "address", "es"), addrs)
	}
	return addrs, nil
}

func ResolveAll(ctx context.Context, host, resolveType string) ([]net.IP, error) {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		log.Debugf("host %s looks like an IPv6, stripping []", host)
		host = host[1 : len(host)-1]
	}
	isAddr := net.ParseIP(host)
	if isAddr != nil {
		log.LogVf("Resolved %s already an IP as addr", host)
		return []net.IP{isAddr}, nil
	}
	addrs, err := net.DefaultResolver.LookupIP(ctx, resolveType, host)
	if err != nil {
		log.Errf("Unable to lookup %q: %v", host, err)
	}
	return addrs, err
}

func IPPortString(ip net.IP, port int) string {
	ipstr := ip.String()
	if strings.Contains(ipstr, ":") {
		ipstr = "[" + ipstr + "]"
	}
	return ipstr + ":" + strconv.Itoa(port)
}

// DumpResponseDetails sort of reconstitutes the server's response (but not really as go
// processes it and the raw response isn't available - use fortio curl fast client for exact bytes).
func DumpResponseDetails(w io.Writer, r *http.Response) {
	fmt.Fprintf(w, "%s %s\n", r.Proto, r.Status)
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		for _, h := range r.Header[name] {
			fmt.Fprintf(w, "%s: %s\n", name, h)
		}
	}
	fmt.Fprintln(w)
}

// AddAndValidateExtraHeader collects extra headers (see cli/main.go for example).
// Inspired/borrowed from fortio/fhttp.
func (cfg *Config) AddAndValidateExtraHeader(hdr string) error {
	s := strings.SplitN(hdr, ":", 2)
	if len(s) != 2 {
		return fmt.Errorf("invalid extra header '%s', expecting Key: Value", hdr)
	}
	key := strings.TrimSpace(s[0])
	// No TrimSpace for the value, so we can set empty "" vs just whitespace " " which
	// will get trimmed later but treated differently: not emitted vs emitted empty for User-Agent.
	value := s[1]
	switch strings.ToLower(key) {
	case "host":
		log.LogVf("Will be setting special Host header to %s", value)
		cfg.HostOverride = strings.TrimSpace(value) // This one needs to be trimmed
	case "user-agent":
		// To remove you must set to empty string as otherwise std go client adds its own
		log.LogVf("User-Agent being Set to %q", value)
		cfg.Headers.Set(key, value)
	default:
		log.LogVf("Setting regular extra header %s: %s", key, value)
		cfg.Headers.Add(key, value)
		log.Debugf("headers now %+v", cfg.Headers)
	}
	return nil
}

func NewConfig() *Config {
	cfg := Config{
		Headers:         make(http.Header, 1),
		RepeatDelay:     5 * time.Second,
		CertExpiryError: Dur(7), // 7 days default to complain about cert expiry
	}
	cfg.Headers.Set("User-Agent", "fortio.org/multicurl-"+libShortVersion)
	return &cfg
}

func Filename(cfg *Config, addr net.IP) string {
	aStr := addr.String()
	return strings.Replace(cfg.OutputPattern, "%", aStr, 1)
}

func ReadIPs(filename string) ([]net.IP, error) {
	var file io.ReadCloser
	if filename == "-" {
		log.Infof("Using stdin for list of IPs to connect to")
		file = os.Stdin
	} else {
		log.Infof("Using content of %q to resolve IPs", filename)
		var err error
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var addrs []net.IP
	for {
		line, _, err := reader.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		}
		ipStr := strings.TrimSpace(string(line))
		if len(ipStr) == 0 || strings.HasPrefix(ipStr, "#") {
			continue
		}
		if strings.HasPrefix(ipStr, "[") && strings.HasSuffix(ipStr, "]") {
			ipStr = ipStr[1 : len(ipStr)-1]
		}
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, fmt.Errorf("unable to parse IP %q", ipStr)
		}
		addrs = append(addrs, ip)
	}
	return addrs, nil
}

// Days returns the number of days in the duration (floating point).
func Days(d time.Duration) float64 {
	return d.Seconds() / 86400.
}

// Dur returns a duration corresponding to the number of days passed.
func Dur(days float64) time.Duration {
	return time.Duration(days * 24 * float64(time.Hour))
}
