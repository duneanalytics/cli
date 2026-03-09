.PHONY: all setup build run lint lint-timezone test

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

lint: bin/golangci-lint lint-timezone
	go fmt ./...
	go vet ./...
	bin/golangci-lint -c .golangci.yml run ./...
	go mod tidy

# All binary entrypoints must set time.Local = time.UTC in init() to ensure
# time.Now() always returns UTC.
lint-timezone:
	@fail=0; \
	for f in $$(find . -name main.go -path '*/cmd/*/main.go'); do \
		if ! grep -q 'time\.Local = time\.UTC' "$$f"; then \
			echo "ERROR: $$f missing 'time.Local = time.UTC' in init()"; \
			fail=1; \
		fi; \
	done; \
	if [ "$$fail" -eq 1 ]; then exit 1; fi

test:
	go test -timeout=10s -race -cover -bench=. -benchmem ./...
