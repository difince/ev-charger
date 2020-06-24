.PHONY: build test clean docker
GO = CGO_ENABLED=0 GO111MODULE=on go

BINARY=ev-charger

VERSION=$(shell cat ./VERSION)
GOFLAGS=-ldflags "-X github.com/edgexfoundry/ev-charger.Version=$(VERSION)"

build:
	$(GO) build $(GOFLAGS) -o cmd/$(BINARY) ./cmd

GIT_SHA=$(shell git rev-parse HEAD)

docker:
	docker build \
		-f cmd/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-ev-charger-ds:$(GIT_SHA) \
		-t edgexfoundry/docker-ev-charger-ds:$(VERSION)-dev \
		.

test:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]
	./bin/test-attribution-txt.sh
	./bin/test-go-mod-tidy.sh

clean:
	rm -f cmd/$(BINARY)
