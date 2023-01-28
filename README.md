[![codecov](https://codecov.io/github/fortio/multicurl/branch/main/graph/badge.svg?token=LONYZDFQ7C)](https://codecov.io/github/fortio/multicurl)

# multicurl

Fetches a URL from all the IPs of a given host

## Installation

go install github.com/fortio/multicurl@latest

## Usage

multicurl https://debug.fortio.org/test

Use `-4` for ipv4 only, `-6` for ipv6 only, otherwise it'll try all of them.

Relevant flags (some extra are from fortio library but not used/relevant)
```
flags:
  -4	Only IPv4
  -6	Only IPv6
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
