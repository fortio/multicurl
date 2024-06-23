all: lint check test-local-image

test: check test-local-image
	go test -race ./...

lint: .golangci.yml
	golangci-lint run

check: .goreleaser.yaml
	goreleaser check

local-image:
	CGO_ENABLED=0 GOOS=linux go build -a .
	docker build -t fortio/multicurl:local -f Dockerfile .

test-local-image: local-image
	docker run --rm fortio/multicurl:local -4 https://debug.fortio.org/build-test

.golangci.yml: Makefile
	curl -fsS -o .golangci.yml https://raw.githubusercontent.com/fortio/workflows/main/golangci.yml

.goreleaser.yaml: Makefile
	curl -fsS -o .goreleaser.yaml https://raw.githubusercontent.com/fortio/workflows/main/goreleaser.yaml # same use branch for testing instead of main in #38


.PHONY: lint check all local-image test-local-image
