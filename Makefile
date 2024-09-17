all: clean lint check test-local-image

OS:=$(shell go env GOOS)

test: test-local-image
ifeq ($(OS),windows)
	@echo "Skipping test on windows, issue with -- and testscript"
else
	go test -race ./...
endif

lint: .golangci.yml
	golangci-lint run

check: .goreleaser.yaml
	goreleaser check

multicurl: # normal one with the bundle through fortio/cli
	CGO_ENABLED=0 GOOS=linux go build -a .

# Will fail because of missing bundle, on purpose, to confirm the negative build tag works.
no-bundle-failing-test: build_no_tls_fallback test-local-image

build_no_tls_fallback:
	CGO_ENABLED=0 GOOS=linux go build -a -tags no_tls_fallback .

clean:
	rm -f multicurl

# on a mac with docker... run `make OS=linux` to test local docker.
test-local-image: multicurl
ifeq ($(OS),linux)
	docker build -t fortio/multicurl:local -f Dockerfile .
	docker run --rm fortio/multicurl:local -4 https://debug.fortio.org/build-test
endif

.golangci.yml: Makefile
	curl -fsS -o .golangci.yml https://raw.githubusercontent.com/fortio/workflows/main/golangci.yml

.goreleaser.yaml: Makefile
	curl -fsS -o .goreleaser.yaml https://raw.githubusercontent.com/fortio/workflows/main/goreleaser.yaml # same use branch for testing instead of main in #38

.PHONY: lint check all local-image test-local-image no-bundle-failing-test build_no_tls_fallback clean
