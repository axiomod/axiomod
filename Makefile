BINARY_NAME=axiomod-server
CLI_NAME=axiomod

.PHONY: all build build-cli clean test deps lint fmt help docker

all: build build-cli

VERSION := v1.2.0
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X github.com/axiomod/axiomod/framework/version.Version=$(VERSION) \
           -X github.com/axiomod/axiomod/framework/version.GitCommit=$(COMMIT) \
           -X github.com/axiomod/axiomod/framework/version.BuildDate=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/axiomod-server
	@echo "Built $(BINARY_NAME) version $(VERSION)"

build-cli:
	go build -ldflags "$(LDFLAGS)" -o bin/$(CLI_NAME) ./cmd/axiomod
	@echo "Built $(CLI_NAME) version $(VERSION)"

clean:
	go clean
	rm -rf bin/
	@echo "Cleaned build artifacts"

test:
	go test -v ./...

deps:
	go mod tidy
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

docker:
	docker build -t axiomod/server:latest .

help:
	@echo "Available commands:"
	@echo "  make all          - Build both server and CLI"
	@echo "  make build        - Build the server binary"
	@echo "  make build-cli    - Build the CLI tool"
	@echo "  make test         - Run all tests"
	@echo "  make clean        - Remove binaries and build artifacts"
	@echo "  make deps         - Install dependencies"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format Go code"
	@echo "  make docker       - Build Docker image for the server"
