.PHONY: all setup build lint test

all: lint test build

setup: bin/golangci-lint
	go mod download

dune-cli: lint
	go build -o dune-cli cmd/main.go

build: dune-cli

bin:
	mkdir -p bin

bin/golangci-lint: bin
	GOBIN=$(PWD)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2

lint: bin/golangci-lint
	go fmt ./...
	go vet ./...
	bin/golangci-lint -c .golangci.yml run ./...
	go mod tidy

test:
	go test -timeout=10s -race -cover -bench=. -benchmem ./...
