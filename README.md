[![codecov](https://codecov.io/github/fortio/multicurl/branch/main/graph/badge.svg?token=LONYZDFQ7C)](https://codecov.io/github/fortio/multicurl)

# multicurl

Fetches a URL from all the IPs of a given host

## Installation
```
go install github.com/fortio/multicurl@latest
```

Or the [binary releases](https://github.com/fortio/multicurl/releases)

Or using the docker image
```
docker run fortio/multicurl http://debug.fortio.org/test
```

## Usage

multicurl https://debug.fortio.org/test

Use `-4` for ipv4 only, `-6` for ipv6 only, otherwise it'll try all of them.

Relevant flags (some extra are from fortio library but not used/relevant)

```
flags:
flags:
  -4	Use only IPv4
  -6	Use only IPv6
  -i	Include response headers in output
  -s	Quiet mode (sets log level to warning quietly)
  -loglevel value
    	loglevel, one of [Debug Verbose Info Warning Error Critical Fatal] (default Info)
  -method string
    	HTTP method (default "GET")
  -request-timeout duration
    	HTTP method (default 3s)
  -total-timeout duration
    	HTTP method (default 30s)
  -version
    	Show full version info and exit.
```

See also [multicurl.txtar](multicurl.txtar) for examples (tests)
