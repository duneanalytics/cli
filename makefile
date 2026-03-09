.PHONY: all setup build run lint test

VERSION        ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//')
COMMIT         ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "unknown")
DATE           ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
AMPLITUDE_KEY  ?= $(DUNE_CLI_AMPLITUDE_KEY)
LDFLAGS         = -s -w \
                  -X main.version=$(VERSION) \
                  -X main.commit=$(COMMIT) \
                  -X main.date=$(DATE) \
                  -X main.amplitudeKey=$(AMPLITUDE_KEY)

all: lint test build

setup: bin/golangci-lint
	go mod download

build: lint
	go build -ldflags '$(LDFLAGS)' -o dune-cli ./cmd

run:
	go run -ldflags '$(LDFLAGS)' ./cmd $(ARGS)

bin:
	mkdir -p bin

bin/golangci-lint: bin
	GOBIN=$(PWD)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

lint: bin/golangci-lint
	go fmt ./...
	go vet ./...
	bin/golangci-lint -c .golangci.yml run ./...
	go mod tidy

test:
	go test -timeout=10s -race -cover -bench=. -benchmem ./...
