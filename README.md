[![codecov](https://codecov.io/github/fortio/multicurl/branch/main/graph/badge.svg?token=LONYZDFQ7C)](https://codecov.io/github/fortio/multicurl)

# multicurl

Fetches a URL from all the IPs of a given host. Optionally repeat until an expected result code is obtained from all addresses.
It will also print information about certificates, including the shortest expiration found.

## Installation

If you have a recent go installation already:
```shell
CGO_ENABLED=0 go install fortio.org/multicurl@latest
```

Or get on of the [binary releases](https://github.com/fortio/multicurl/releases)

Or using the docker image
```shell
docker run fortio/multicurl http://debug.fortio.org/test
```

Or using brew (mac)
```shell
brew install fortio/tap/multicurl
```

## Usage

multicurl https://debug.fortio.org/test

Use `-4` for ipv4 only, `-6` for ipv6 only, otherwise it'll try all of them.

<!-- generate using
LOGGER_CONSOLE_COLOR=false LOGGER_IGNORE_CLI_MODE=true \
go run . help | expand | fold -s -w 92 | sed -e "s/ $//" -e "s/</\&lt;/"
-->
```
flags:
  -4    Use only IPv4
  -6    Use only IPv6
  -H key:value
        Additional http header(s). Multiple key:value pairs can be passed using multiple -H.
  -I file
        IP address file to use instead of resolving the URL, use - for stdin
  -X string
        HTTP method to use, default is GET unless -d is set which defaults to POST
  -cacert file
        Path to a custom CA certificate file to use instead of system ones.
  -cert file
        Path to a custom client certificate file for mTLS.
  -cert-expiry days
        Certificate expiry error threshold in days (default 7)
  -d string
        Payload to POST, use @filename to read from file
  -expected int
        Expected HTTP return code, 0 means any and non 200s will be warning otherwise if
set any different code is an error
  -i    Include response headers in output
  -insecure
        Skip verification of server certificate (insecure TLS)
  -json
        JSON output of summary results
  -key file
        Path to a custom client key file for mTLS.
  -logger-force-color
        Force color output even if stderr isn't a terminal
  -logger-no-color
        Prevent colorized output even if stderr is a terminal
  -loglevel level
        log level, one of [Debug Verbose Info Warning Error Critical Fatal] (default Info)
  -n int
        Max number of IPs to use/try (0 means all the ones found)
  -nobar
        Disable display of progress bar (or spinner when no content-length)
  -o file name pattern
        Output file name pattern, e.g "out-%.html" where % will be replaced by the ip,
default is stdout, use "none" for no output (in combination with -json for instance)
  -quiet
        Quiet mode, sets loglevel to Error (quietly) to reduces the output
  -relookup
        Re-lookup the URL between each repeat
  -repeat int
        Max number of times to retry on errors if positive, default is 0 (no retry),
negative is retry until -total-timeout
  -repeat-delay duration
        Delay between retries (default 5s)
  -request-timeout duration
        HTTP method (default 3s)
  -total-timeout duration
        HTTP method (default 30s)
```

Note that `-relookup` works better on CGO_ENABLED=0 built binary, otherwise the OS library caches the results.

Note that `-H Host:xxx https://yyyy/` is a special header and using that will be the same as querying `https://xxx/` using the IPs of `yyy` (convenient to test a virtual host against a LoadBalancer or ingress name before the DNS is updated)

See also [multicurl.txtar](multicurl.txtar) for examples (tests)

### Example

#### Regular
```
$ multicurl -expected 301 -repeat 2 -n 2 -relookup debug.fortio.org
11:49:20 I Fortio multicurl dev  go1.19.5 arm64 darwin, using resolver ip, GET debug.fortio.org
11:49:20 I Resolved ip debug.fortio.org:http to port 80 and 6 addresses [18.222.136.83 192.9.142.5 192.9.227.83 2600:1f16:9c6:b400:282c:a766:6cab:4e82 2603:c024:c00a:d144:6663:5896:7efb:fbf3 2603:c024:c00a:d144:7cd0:4951:7106:96b8] - keeping first 2
11:49:20 E 1: Status 200 "200 OK" from 18.222.136.83
Φορτίο version 1.40.1 h1:D1H+5aOnauTr4WTnopHl1MhSZt/l0Asi3ZEqkpBwT0c= go1.19.5 arm64 linux (in fortio.org/proxy 1.8.1)
Debug server on a1 up for 37h43m3s
Request from 216.194.105.159:4496

GET / HTTP/1.1

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
Connection: close
User-Agent: fortio.org/multicurl-dev

body:


11:49:20 E 2: Status 200 "200 OK" from 192.9.142.5
Φορτίο version 1.40.1 h1:D1H+5aOnauTr4WTnopHl1MhSZt/l0Asi3ZEqkpBwT0c= go1.19.5 arm64 linux (in fortio.org/proxy 1.8.1)
Debug server on oa1 up for 23h55m4.5s
Request from 216.194.105.159:4497

GET / HTTP/1.1

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
Connection: close
User-Agent: fortio.org/multicurl-dev

body:


11:49:20 E [1] 2 errors (0 warnings)
11:49:25 I Resolved ip debug.fortio.org:http to port 80 and 6 addresses [18.222.136.83 192.9.142.5 192.9.227.83 2603:c024:c00a:d144:7cd0:4951:7106:96b8 2600:1f16:9c6:b400:282c:a766:6cab:4e82 2603:c024:c00a:d144:6663:5896:7efb:fbf3] - keeping first 2
11:49:26 E 1: Status 200 "200 OK" from 18.222.136.83
Φορτίο version 1.40.1 h1:D1H+5aOnauTr4WTnopHl1MhSZt/l0Asi3ZEqkpBwT0c= go1.19.5 arm64 linux (in fortio.org/proxy 1.8.1)
Debug server on a1 up for 37h43m8.5s
Request from 216.194.105.159:4624

GET / HTTP/1.1

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
Connection: close
User-Agent: fortio.org/multicurl-dev

body:


11:49:26 E 2: Status 200 "200 OK" from 192.9.142.5
Φορτίο version 1.40.1 h1:D1H+5aOnauTr4WTnopHl1MhSZt/l0Asi3ZEqkpBwT0c= go1.19.5 arm64 linux (in fortio.org/proxy 1.8.1)
Debug server on oa1 up for 23h55m10s
Request from 216.194.105.159:4625

GET / HTTP/1.1

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
Connection: close
User-Agent: fortio.org/multicurl-dev

body:


11:49:26 E [2] 2 errors (0 warnings)
11:49:31 I Resolved ip debug.fortio.org:http to port 80 and 6 addresses [192.9.227.83 18.222.136.83 192.9.142.5 2600:1f16:9c6:b400:282c:a766:6cab:4e82 2603:c024:c00a:d144:6663:5896:7efb:fbf3 2603:c024:c00a:d144:7cd0:4951:7106:96b8] - keeping first 2
11:49:31 E 1: Status 200 "200 OK" from 192.9.227.83
Φορτίο version 1.40.1 h1:D1H+5aOnauTr4WTnopHl1MhSZt/l0Asi3ZEqkpBwT0c= go1.19.5 amd64 linux (in fortio.org/proxy 1.8.1)
Debug server on ol1 up for 37h36m2.9s
Request from 216.194.105.159:4784

GET / HTTP/1.1

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
Connection: close
User-Agent: fortio.org/multicurl-dev

body:


11:49:31 E 2: Status 200 "200 OK" from 18.222.136.83
Φορτίο version 1.40.1 h1:D1H+5aOnauTr4WTnopHl1MhSZt/l0Asi3ZEqkpBwT0c= go1.19.5 arm64 linux (in fortio.org/proxy 1.8.1)
Debug server on a1 up for 37h43m14.2s
Request from 216.194.105.157:4528

GET / HTTP/1.1

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
Connection: close
User-Agent: fortio.org/multicurl-dev

body:


11:49:31 E [3] 2 errors (0 warnings)
11:49:31 E Reached max repeat 2
11:49:31 I Total iterations: 3, errors: 6, warnings 0
exit status 2
```

#### Certificate information

```bash
% multicurl -4 -cert-expiry 60 https://debug.fortio.org > /dev/null
```
Yields
```
12:46:54 I Fortio multicurl dev  go1.19.6 arm64 darwin, using resolver ip4, GET https://debug.fortio.org
12:46:55 I Resolved ip4 debug.fortio.org:https to port 443 and 3 addresses [192.9.142.5 18.222.136.83 192.9.227.83]
12:46:55 I 1: Status 200 "200 OK" from 192.9.142.5
12:46:55 I Certificate "CN=debug.fortio.org" expires in 52 days
12:46:55 I Certificate "CN=R3,O=Let's Encrypt,C=US" expires in 925 days
12:46:55 I Certificate "CN=ISRG Root X1,O=Internet Security Research Group,C=US" expires in 575 days
12:46:55 I 2: Status 200 "200 OK" from 18.222.136.83
12:46:55 I Certificate "CN=debug.fortio.org" expires in 51 days
12:46:55 I Certificate "CN=R3,O=Let's Encrypt,C=US" expires in 925 days
12:46:55 I Certificate "CN=ISRG Root X1,O=Internet Security Research Group,C=US" expires in 575 days
12:46:55 I 3: Status 200 "200 OK" from 192.9.227.83
12:46:55 I Certificate "CN=debug.fortio.org" expires in 52 days
12:46:55 I Certificate "CN=R3,O=Let's Encrypt,C=US" expires in 925 days
12:46:55 I Certificate "CN=ISRG Root X1,O=Internet Security Research Group,C=US" expires in 575 days
12:46:55 I [1] 0 errors (0 warnings)
12:46:55 E Shortest cert expiry is 2023-04-26 02:13:27 +0000 UTC (51.2 days from now)
12:46:55 I Total iterations: 1, errors: 0, warnings 0
exit status 1
```

### JSON output

```bash
multicurl -json -quiet -o none https://debug.fortio.org | jq
```
Yields
 ```json
{
  "Errors": 0,
  "Warnings": 0,
  "Addresses": [
    "2603:c024:c00a:d144:6663:5896:7efb:fbf3",
    "2603:c024:c00a:d144:7cd0:4951:7106:96b8",
    "2600:1f16:9c6:b400:282c:a766:6cab:4e82",
    "192.9.142.5",
    "192.9.227.83",
    "18.222.136.83"
  ],
  "Codes": {
    "18.222.136.83:443": 200,
    "192.9.142.5:443": 200,
    "192.9.227.83:443": 200,
    "[2600:1f16:9c6:b400:282c:a766:6cab:4e82]:443": 200,
    "[2603:c024:c00a:d144:6663:5896:7efb:fbf3]:443": 200,
    "[2603:c024:c00a:d144:7cd0:4951:7106:96b8]:443": 200
  },
  "Sizes": {
    "18.222.136.83:443": 339,
    "192.9.142.5:443": 340,
    "192.9.227.83:443": 340,
    "[2600:1f16:9c6:b400:282c:a766:6cab:4e82]:443": 369,
    "[2603:c024:c00a:d144:6663:5896:7efb:fbf3]:443": 370,
    "[2603:c024:c00a:d144:7cd0:4951:7106:96b8]:443": 370
  },
  "Iterations": 1,
  "ShortestCertExpiry": "2023-04-26T02:13:27Z"
}
```

Note the handy `ShortestCertExpiry` entry.


ps: this started as https://pkg.go.dev/github.com/fortio/multicurl and now is available under https://pkg.go.dev/fortio.org/multicurl


Note: If you have iCloud Private Relay on and IpV6 and build with go1.24+, that the resolution of http URLs gets cached and made to the wrong same ipv6 instead of the 3 ipv4 of http://debug.fortio.org for instance, it works fine with any other go version or OS combination - or using `-tags netgo` - the binary we build use 1.23 so don't have that issue either way.
