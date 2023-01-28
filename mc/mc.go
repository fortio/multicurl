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
package mc

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"fortio.org/fortio/log"
)

// MultiCurl is the main function of the multicurl tool. timeout is per request/ip.
func MultiCurl(ctx context.Context, timeout time.Duration, method, urlString, resolveType string) int {
	numErrors := 0
	if len(urlString) == 0 {
		return log.FErrf("Unexpected empty url")
	}
	urlString = URLAddScheme(urlString)
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
	log.Infof("Resolving %s host %s port %s", resolveType, host, port)
	portNum, addrs, err := ResolveAll(ctx, host, port, resolveType)
	if err != nil {
		return 1 // already logged
	}
	plural := ""
	if len(addrs) != 1 {
		plural = "es"
	}
	log.Infof("Resolved %s %s:%s to port %d and %d address%s %v", resolveType, host, port, portNum, len(addrs), plural, addrs)
	req, err := http.NewRequestWithContext(ctx, method, urlString, nil)
	if err != nil {
		return log.FErrf("Error creating request: %v", err)
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DisableKeepAlives = true
	cli := http.Client{
		Transport: tr,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	numWarnings := 0
	for i, addr := range addrs {
		log.Infof("%d: Using %s", i, addr)
		aStr := IPPortString(addr, portNum)
		tr.DialContext = func(ctx context.Context, network, oAddr string) (net.Conn, error) {
			log.Infof("%d: DialContext %s %s -> %s", i, network, oAddr, aStr)
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
		log.Logf(level, "%d: Status %d (%s) from %s", i, resp.StatusCode, resp.Status, addr)
		data, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			log.Errf("%d: Error reading body from %s: %v", i, addr, err)
			numErrors++
		}
		os.Stdout.Write(data)
	}
	level := log.Info
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
	log.Infof("Assuming http:// on missing scheme for %q", url)
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
