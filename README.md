[![codecov](https://codecov.io/github/fortio/multicurl/branch/main/graph/badge.svg?token=LONYZDFQ7C)](https://codecov.io/github/fortio/multicurl)

# multicurl

Fetches a URL from all the IPs of a given host. Optionally repeat until an expected result code is obtained from all addresses.

## Installation
```shell
CGO_ENABLED=0 go install github.com/fortio/multicurl@latest
```

Or the [binary releases](https://github.com/fortio/multicurl/releases)

Or using the docker image
```shell
docker run fortio/multicurl http://debug.fortio.org/test
```

## Usage

multicurl https://debug.fortio.org/test

Use `-4` for ipv4 only, `-6` for ipv6 only, otherwise it'll try all of them.

Relevant flags (some extra are from fortio library but not used/relevant)

```
flags:
  -4    Use only IPv4
  -6    Use only IPv6
  -H key:value
        Additional http header(s). Multiple key:value pairs can be passed using
multiple -H.
  -I file
        IP address file to use instead of resolving the URL, use - for stdin
  -X string
        HTTP method to use, default is GET unless -d is set which defaults to
POST
  -d string
        Payload to POST, use @filename to read from file
  -expected int
        Expected HTTP return code, 0 means any and non 200s will be warning
otherwise if set any different code is an error
  -i    Include response headers in output
  -loglevel value
        loglevel, one of [Debug Verbose Info Warning Error Critical Fatal]
(default Info)
  -n int
        Max number of IPs to use/try (0 means all the ones found)
  -o file name pattern
        Output file name pattern, e.g "out-%.html" where % will be replaced by
the ip, default is stdout
  -relookup
        Re-lookup the URL between each repeat
  -repeat int
        Max number of times to retry on errors if positive, default is 0 (no
retry), negative is retry until -total-timeout
  -repeat-delay duration
        Delay between retries (default 5s)
  -request-timeout duration
        HTTP method (default 3s)
  -s    Quiet mode (sets log level to warning quietly)
  -total-timeout duration
        HTTP method (default 30s)
  -version
        Show full version info and exit.
```

Note that `-relookup` works better on CGO_ENABLED=0 built binary, otherwise the OS library caches the results.

See also [multicurl.txtar](multicurl.txtar) for examples (tests)

### Example

```
 -expected 301 -repeat 2 -n 2 -relookup debug.fortio.org
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
