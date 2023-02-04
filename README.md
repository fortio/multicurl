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
  -4	Use only IPv4
  -6	Use only IPv6
  -H key:value
    	Additional http header(s). Multiple key:value pairs can be passed using multiple -H.
  -I file
    	IP address file to use instead of resolving the URL, use - for stdin
  -X string
    	HTTP method to use, default is GET unless -d is set which defaults to POST
  -d string
    	Payload to POST, use @filename to read from file
  -expected int
    	Expected HTTP return code, 0 means any and non 200s will be warning otherwise
      if set any different code is an error
  -i	Include response headers in output
  -loglevel value
    	loglevel, one of [Debug Verbose Info Warning Error Critical Fatal] (default Info)
  -n int
    	Max number of IPs to use/try (0 means all the ones found)
  -o file name pattern
    	Output file name pattern, e.g "out-%.html" where % will be replaced by the ip, default is stdout
  -relookup
    	Re-lookup the URL between each repeat
  -repeat int
    	Max number of times to retry on errors if positive, default is 0 (no retry), 
      negative is retry until -total-timeout
  -repeat-delay duration
    	Delay between retries (default 5s)
  -request-timeout duration
    	HTTP method (default 3s)
  -s	Quiet mode (sets log level to warning quietly)
  -total-timeout duration
    	HTTP method (default 30s)
  -version
    	Show full version info and exit.
```

Note that `-relookup` works better on CGO_ENABLED=0 built binary, otherwise the OS library caches the results.

See also [multicurl.txtar](multicurl.txtar) for examples (tests)

### Example

```
$ multicurl -i https://debug.fortio.org
17:35:52 I Fortio multicurl 1.1.0 h1:LUqSvzZCem9zhawlHnHVBS8ijTCvleQaQn8l7ibugvU= go1.19.5 arm64 darwin, using resolver ip, GET https://debug.fortio.org
17:35:52 I Resolving ip host debug.fortio.org port https
17:35:52 I Resolved ip debug.fortio.org:https to port 443 and 6 addresses [2600:1f16:9c6:b400:282c:a766:6cab:4e82 2603:c024:c00a:d144:7cd0:4951:7106:96b8 2603:c024:c00a:d144:230c:a364:9794:317b 192.9.142.5 192.9.227.83 18.222.136.83]
17:35:52 I 0: Using 2600:1f16:9c6:b400:282c:a766:6cab:4e82
17:35:52 I 0: DialContext tcp debug.fortio.org:443 -> [2600:1f16:9c6:b400:282c:a766:6cab:4e82]:443
17:35:52 I 0: Status 200 "200 OK" from 2600:1f16:9c6:b400:282c:a766:6cab:4e82
HTTP/2.0 200 OK
Content-Type: text/plain; charset=UTF-8
Date: Sun, 29 Jan 2023 01:35:52 GMT

Φορτίο version 1.40.0 h1:jSDO/jGcyC/qTpMZZ84EZbn9BQawsWM9/RMQ9s6Cn3w= go1.19.5 arm64 linux (in fortio.org/proxy 1.8.0)
Debug server on a1 up for 22h55m36.7s
Request from [2600:1700:...ipv6 masked...]:59546 https TLS_AES_128_GCM_SHA256

GET / HTTP/2.0

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
User-Agent: Go-http-client/2.0

body:


17:35:52 I 1: Using 2603:c024:c00a:d144:7cd0:4951:7106:96b8
17:35:52 I 1: DialContext tcp debug.fortio.org:443 -> [2603:c024:c00a:d144:7cd0:4951:7106:96b8]:443
17:35:52 I 1: Status 200 "200 OK" from 2603:c024:c00a:d144:7cd0:4951:7106:96b8
HTTP/2.0 200 OK
Content-Type: text/plain; charset=UTF-8
Date: Sun, 29 Jan 2023 01:35:52 GMT

Φορτίο version 1.40.0 h1:jSDO/jGcyC/qTpMZZ84EZbn9BQawsWM9/RMQ9s6Cn3w= go1.19.5 amd64 linux (in fortio.org/proxy 1.8.0)
Debug server on l1 up for 22h49m20.5s
Request from [2600:1700:...ipv6 masked...]]:59548 https TLS_AES_128_GCM_SHA256

GET / HTTP/2.0

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
User-Agent: Go-http-client/2.0

body:


17:35:52 I 2: Using 2603:c024:c00a:d144:230c:a364:9794:317b
17:35:52 I 2: DialContext tcp debug.fortio.org:443 -> [2603:c024:c00a:d144:230c:a364:9794:317b]:443
17:35:52 I 2: Status 200 "200 OK" from 2603:c024:c00a:d144:230c:a364:9794:317b
HTTP/2.0 200 OK
Content-Type: text/plain; charset=UTF-8
Date: Sun, 29 Jan 2023 01:35:52 GMT

Φορτίο version 1.40.0 h1:jSDO/jGcyC/qTpMZZ84EZbn9BQawsWM9/RMQ9s6Cn3w= go1.19.5 amd64 linux (in fortio.org/proxy 1.8.0)
Debug server on l2 up for 22h49m2.2s
Request from [2600:1700::...ipv6 masked...]]:59549 https TLS_AES_128_GCM_SHA256

GET / HTTP/2.0

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
User-Agent: Go-http-client/2.0

body:


17:35:52 I 3: Using 192.9.142.5
17:35:52 I 3: DialContext tcp debug.fortio.org:443 -> 192.9.142.5:443
17:35:52 I 3: Status 200 "200 OK" from 192.9.142.5
HTTP/2.0 200 OK
Content-Type: text/plain; charset=UTF-8
Date: Sun, 29 Jan 2023 01:35:52 GMT

Φορτίο version 1.40.0 h1:jSDO/jGcyC/qTpMZZ84EZbn9BQawsWM9/RMQ9s6Cn3w= go1.19.5 amd64 linux (in fortio.org/proxy 1.8.0)
Debug server on l2 up for 22h49m2.2s
Request from 99..ipv4-masked...:59550 https TLS_AES_128_GCM_SHA256

GET / HTTP/2.0

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
User-Agent: Go-http-client/2.0

body:


17:35:52 I 4: Using 192.9.227.83
17:35:52 I 4: DialContext tcp debug.fortio.org:443 -> 192.9.227.83:443
17:35:52 I 4: Status 200 "200 OK" from 192.9.227.83
HTTP/2.0 200 OK
Content-Type: text/plain; charset=UTF-8
Date: Sun, 29 Jan 2023 01:35:53 GMT

Φορτίο version 1.40.0 h1:jSDO/jGcyC/qTpMZZ84EZbn9BQawsWM9/RMQ9s6Cn3w= go1.19.5 amd64 linux (in fortio.org/proxy 1.8.0)
Debug server on l1 up for 22h49m20.7s
Request from 99..ipv4-masked...:59551 https TLS_AES_128_GCM_SHA256

GET / HTTP/2.0

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
User-Agent: Go-http-client/2.0

body:


17:35:52 I 5: Using 18.222.136.83
17:35:52 I 5: DialContext tcp debug.fortio.org:443 -> 18.222.136.83:443
17:35:53 I 5: Status 200 "200 OK" from 18.222.136.83
HTTP/2.0 200 OK
Content-Type: text/plain; charset=UTF-8
Date: Sun, 29 Jan 2023 01:35:53 GMT

Φορτίο version 1.40.0 h1:jSDO/jGcyC/qTpMZZ84EZbn9BQawsWM9/RMQ9s6Cn3w= go1.19.5 arm64 linux (in fortio.org/proxy 1.8.0)
Debug server on a1 up for 22h55m37.3s
Request from 99..ipv4-masked...:59552 https TLS_AES_128_GCM_SHA256

GET / HTTP/2.0

headers:

Host: debug.fortio.org
Accept-Encoding: gzip
User-Agent: Go-http-client/2.0

body:


17:35:53 I Total 0 errors (0 warnings)
```
