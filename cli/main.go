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
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"time"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/multicurl/mc"
)

// -- Support for multiple instances of -H flag on cmd line.
type headersFlagList struct{}

func (f *headersFlagList) String() string {
	return ""
}

func (f *headersFlagList) Set(value string) error {
	return config.AddAndValidateExtraHeader(value)
}

// -- end of functions for -H support

var config = mc.NewConfig()

// Main is the main function for the multicurl tool so it can be called from testscript.
// Note that we could use the (new in 1.39) log.Fatalf that doesn't panic for cli tools but
// it calling os.Exit directly means it doesn't work with the code coverage from `testscript`.
func Main() int {
	ipv4 := flag.Bool("4", false, "Use only IPv4")
	ipv6 := flag.Bool("6", false, "Use only IPv6")
	inclHeaders := flag.Bool("i", false, "Include response headers in output")
	method := flag.String("X", "", "HTTP method to use, default is GET unless -d is set which defaults to POST")
	totalTimeout := flag.Duration("total-timeout", 30*time.Second, "HTTP method")
	requestTimeout := flag.Duration("request-timeout", 3*time.Second, "HTTP method")
	var headersFlags headersFlagList
	flag.Var(&headersFlags, "H",
		"Additional http header(s). Multiple `key:value` pairs can be passed using multiple -H.")
	output := flag.String("o", "", "Output `file name pattern`, e.g \"out-%.html\" where % will be replaced by the ip, "+
		"default is stdout, use \"none\" for no output (in combination with -json for instance)")
	data := flag.String("d", "", "Payload to POST, use @filename to read from file")
	ipInput := flag.String("I", "", "IP address `file` to use instead of resolving the URL, use - for stdin")
	expected := flag.Int("expected", 0,
		"Expected HTTP return code, 0 means any and non 200s will be warning otherwise if set any different code is an error")
	repeat := flag.Int("repeat", 0,
		"Max number of times to retry on errors if positive, default is 0 (no retry), negative is retry until -total-timeout")
	retryDelay := flag.Duration("repeat-delay", 5*time.Second, "Delay between retries")
	maxIPs := flag.Int("n", 0, "Max number of IPs to use/try (0 means all the ones found)")
	relookup := flag.Bool("relookup", false, "Re-lookup the URL between each repeat")
	expiryThreshold := flag.Float64("cert-expiry", 7, "Certificate expiry error threshold in `days`")
	caCertFlag := flag.String("cacert", "",
		"Path to a custom CA certificate `file` to use instead of system ones.")
	insecure := flag.Bool("insecure", false, "Skip verification of server certificate (insecure TLS)")
	certFlag := flag.String("cert", "", "Path to a custom client certificate `file` for mTLS.")
	keyFlag := flag.String("key", "", "Path to a custom client key `file` for mTLS.")
	jsonFlag := flag.Bool("json", false, "JSON output of summary results")
	noBarFlag := flag.Bool("nobar", false, "Disable display of progress bar (or spinner when no content-length)")

	cli.ProgramName = "Fortio multicurl"
	cli.ArgsHelp = "url"
	cli.MinArgs = 1
	cli.Main()
	resolveType := "ip"
	if !(*ipv4 && *ipv6) {
		if *ipv4 {
			resolveType = "ip4"
		}
		if *ipv6 {
			resolveType = "ip6"
		}
	}
	url := flag.Arg(0)
	ctx, cncl := context.WithTimeout(context.Background(), *totalTimeout)
	defer cncl()
	config.RequestTimeout = *requestTimeout
	config.Method = *method
	config.URL = url
	config.ResolveType = resolveType
	config.IncludeHeaders = *inclHeaders
	config.OutputPattern = *output
	config.IPFile = *ipInput
	config.ExpectedCode = *expected
	config.MaxRepeat = *repeat
	config.RepeatDelay = *retryDelay
	config.MaxIPs = *maxIPs
	config.ReLookup = *relookup
	config.CertExpiryError = mc.Dur(*expiryThreshold)
	config.Insecure = *insecure
	config.CAFile = *caCertFlag
	config.Cert = *certFlag
	config.Key = *keyFlag
	config.NoProgressBar = *noBarFlag
	if *data != "" {
		if config.Method == "" {
			config.Method = http.MethodPost
		}
		config.Payload = payload(*data)
		if config.Payload == nil {
			return 1 // error already logged
		}
	}
	if config.Method == "" {
		config.Method = http.MethodGet
	}
	log.Debugf("Config: %+v", config)
	exitCode, results := mc.MultiCurl(ctx, config)
	log.Debugf("Results: %+v", results)
	log.Infof("Total iterations: %d, errors: %d, warnings %d", results.Iterations, results.Errors, results.Warnings)
	if *jsonFlag {
		j, _ := json.MarshalIndent(results, "", "  ") //nolint:errchkjson // https://github.com/breml/errchkjson/issues/22
		os.Stdout.Write(append(j, '\n'))
	}
	return exitCode
}

func payload(dataStr string) []byte {
	if dataStr[0] != '@' {
		return []byte(dataStr)
	}
	fname := dataStr[1:]
	data, err := os.ReadFile(fname)
	if err != nil {
		log.FErrf("Unable to read payload from file %q: %v", fname, err)
		return nil
	}
	log.Infof("Read %d bytes from %q as payload", len(data), fname)
	return data
}
