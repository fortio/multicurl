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
	"context"
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

	"fortio.org/fortio/log"
	"fortio.org/fortio/version"
)

// Config object for MultiCurl to avoid passing too many parameters.
type Config struct {
	URL            string
	ResolveType    string // `ip4`, `ip6`, or `ip` for both.
	Method         string
	RequestTimeout time.Duration
	IncludeHeaders bool
	Headers        http.Header
	HostOverride   string
}

var (
	libShortVersion string
	libLongVersion  string
)

func init() {
	libShortVersion, libLongVersion, _ = version.FromBuildInfoPath("github.com/fortio/multicurl")
}

// MultiCurl is the main function of the multicurl tool. timeout is per request/ip.
func MultiCurl(ctx context.Context, cfg *Config) int {
	log.Infof("Fortio multicurl %s, using resolver %s, %s %s", libLongVersion, cfg.ResolveType, cfg.Method, cfg.URL)
	numErrors := 0
	if len(cfg.URL) == 0 {
		return log.FErrf("Unexpected empty url")
	}
	urlString := URLAddScheme(cfg.URL)
	// Parse the url, extract components.
	url, err := url.Parse(urlString)
	if err != nil {
		return log.FErrf("Bad url %q : %v", urlString, err)
	}
	host := url.Hostname()
	port := url.Port()
	if port == "" {
		port = url.Scheme // ie http / https which turns into 80 / 443 later
		log.LogVf("No port specified, using %s", port)
	}
	log.LogVf("Resolving %s host %s port %s", cfg.ResolveType, host, port)
	portNum, addrs, err := ResolveAll(ctx, host, port, cfg.ResolveType)
	if err != nil {
		return 1 // already logged
	}
	plural := ""
	if len(addrs) != 1 {
		plural = "es"
	}
	log.Infof("Resolved %s %s:%s to port %d and %d address%s %v", cfg.ResolveType, host, port, portNum, len(addrs), plural, addrs)
	req, err := http.NewRequestWithContext(ctx, cfg.Method, urlString, nil)
	req.Header = cfg.Headers
	req.Host = cfg.HostOverride
	if err != nil {
		return log.FErrf("Error creating request: %v", err)
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DisableKeepAlives = true
	cli := http.Client{
		Transport: tr,
		Timeout:   cfg.RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	numWarnings := 0
	for idx, addr := range addrs {
		i := idx + 1 // humans count from 1
		log.LogVf("%d: Using %s", i, addr)
		aStr := IPPortString(addr, portNum)
		tr.DialContext = func(ctx context.Context, network, oAddr string) (net.Conn, error) {
			log.LogVf("%d: DialContext %s %s -> %s", i, network, oAddr, aStr)
			d := net.Dialer{}
			return d.DialContext(ctx, "tcp", aStr)
		}
		resp, err := cli.Do(req)
		if err != nil {
			log.Errf("%d: Error fetching %s: %v", i, addr, err)
			numErrors++
			continue
		}
		level := log.Info
		if resp.StatusCode != http.StatusOK {
			level = log.Warning
			numWarnings++
		}
		log.Logf(level, "%d: Status %d %q from %s", i, resp.StatusCode, resp.Status, addr)
		if cfg.IncludeHeaders {
			DumpResponseDetails(os.Stdout, resp)
		}
		data, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			log.Errf("%d: Error reading body from %s: %v", i, addr, err)
			numErrors++
		}
		os.Stdout.Write(data)
	}
	level := log.Info
	if numWarnings > 0 {
		level = log.Warning
	}
	if numErrors > 0 {
		level = log.Error
	}
	log.Logf(level, "Total %d %s (%d %s)", numErrors, Plural(numErrors, "error"), numWarnings, Plural(numWarnings, "warning"))
	return numErrors
}

func Plural(i int, noun string) string {
	if i == 1 {
		return noun
	}
	return noun + "s"
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

func ResolveAll(ctx context.Context, host, portString, resolveType string) (int, []net.IP, error) {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		log.Debugf("host %s looks like an IPv6, stripping []", host)
		host = host[1 : len(host)-1]
	}
	port, err := net.LookupPort("tcp", portString)
	if err != nil {
		log.Errf("Unable to resolve port %q: %v", portString, err)
		return 0, nil, err
	}
	isAddr := net.ParseIP(host)
	if isAddr != nil {
		log.LogVf("Resolved %s:%d already an IP as addr", host, port)
		return port, []net.IP{isAddr}, nil
	}
	addrs, err := net.DefaultResolver.LookupIP(ctx, resolveType, host)
	if err != nil {
		log.Errf("Unable to lookup %q: %v", host, err)
	}
	return port, addrs, err
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
		Headers: make(http.Header, 1),
	}
	cfg.Headers.Set("User-Agent", "fortio.org/multicurl-"+libShortVersion)
	return &cfg
}
