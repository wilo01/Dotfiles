# hlp - Developer Utility CLI
BINARY_NAME := hlp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build install clean test tidy run help

all: build

## Build the binary
build: tidy
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/hlp

## Install to GOPATH/bin
install: tidy
	go install $(LDFLAGS) ./cmd/hlp

## Clean build artifacts
clean:
	rm -rf bin/
	go clean

## Run tests
test:
	go test -v ./...

## Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## Tidy dependencies
tidy:
	go mod tidy

## Run the CLI
run: build
	./bin/$(BINARY_NAME)

## Format code
fmt:
	go fmt ./...

## Lint code (requires golangci-lint)
lint:
	golangci-lint run

## Show help
help:
	@echo "hlp - Developer Utility CLI"
	@echo ""
	@echo "Usage:"
	@echo "  make build    - Build the binary"
	@echo "  make install  - Install to GOPATH/bin"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make tidy     - Tidy dependencies"
	@echo "  make run      - Build and run"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Lint code"
	@echo "  make help     - Show this help"
