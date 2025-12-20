GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=examples/dummy-api
CLI_NAME=axiomod

.PHONY: all build clean test deps lint fmt help

all: build

build: 
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)
	@echo "Built $(BINARY_NAME) binary"

build-cli:
	$(GOBUILD) -o bin/$(CLI_NAME) ./cmd/$(CLI_NAME)
	@echo "Built $(CLI_NAME) binary"

clean:
	$(GOCLEAN)
	rm -rf bin/
	@echo "Cleaned build artifacts"

test: 
	$(GOTEST) -v ./...

deps:
	$(GOMOD) tidy
	$(GOMOD) download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Installed dependencies and tools"

lint:
	@echo "Running linters..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

help:
	@echo "Available commands:"
	@echo "  make build        - Build the main service binary"
	@echo "  make build-cli    - Build the Axiomod CLI binary"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make deps         - Install dependencies and tools"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format Go code"
%!(EXTRA string=examples/dummy-api)