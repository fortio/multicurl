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

package cli

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"fortio.org/fortio/log"
	"fortio.org/fortio/version"
	"github.com/fortio/multicurl/mc"
)

var (
	fullVersion = flag.Bool("version", false, "Show full version info and exit.")
	shortV      string
)

func usage(msg string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "Fortio multicurl %s usage:\n\t%s [flags] url\nflags:\n",
		shortV,
		os.Args[0])
	flag.PrintDefaults()
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg, args...)
		fmt.Fprintln(os.Stderr)
	}
}

// Main is the main function for the multicurl tool so it can be called from testscript.
// Note that we could use the (new in 1.39) log.Fatalf that doesn't panic for cli tools but
// it calling os.Exit directly means it doesn't work with the code coverage from `testscript`.
func Main() int {
	ipv4 := flag.Bool("4", false, "Use only IPv4")
	ipv6 := flag.Bool("6", false, "Use only IPv6")
	inclHeaders := flag.Bool("i", false, "Include response headers in output")
	method := flag.String("method", http.MethodGet, "HTTP method")
	totalTimeout := flag.Duration("total-timeout", 30*time.Second, "HTTP method")
	requestTimeout := flag.Duration("request-timeout", 3*time.Second, "HTTP method")
	quietFlag := flag.Bool("s", false, "Quiet mode (sets log level to warning quietly)")
	flag.CommandLine.Usage = func() { usage("") }
	log.SetFlagDefaultsForClientTools()
	sV, longV, fullV := version.FromBuildInfo()
	shortV = sV
	flag.Parse()
	if *quietFlag {
		log.SetLogLevelQuiet(log.Warning)
	}

	if *fullVersion {
		fmt.Print(fullV)
		return 0
	}
	resolveType := "ip"
	if !(*ipv4 && *ipv6) {
		if *ipv4 {
			resolveType = "ip4"
		}
		if *ipv6 {
			resolveType = "ip6"
		}
	}
	numArgs := len(flag.Args())
	if numArgs != 1 {
		usage("Need 1 argument (url), got %d (%v)", numArgs, flag.Args())
		return 1
	}
	url := flag.Args()[0]
	log.Infof("Fortio multicurl %s, using resolver %s, %s %s", longV, resolveType, *method, url)
	ctx, cncl := context.WithTimeout(context.Background(), *totalTimeout)
	defer cncl()
	return mc.MultiCurl(ctx, &mc.Config{
		RequestTimeout: *requestTimeout,
		Method:         *method,
		URL:            url,
		ResolveType:    resolveType,
		IncludeHeaders: *inclHeaders,
	})
}