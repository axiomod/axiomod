BINARY_NAME=axiomod-server
CLI_NAME=axiomod

.PHONY: all build build-cli cleaner test deps lint fmt help

all: build build-cli

build:
	go build -o bin/$(BINARY_NAME) ./cmd/axiomod-server
	@echo "Built $(BINARY_NAME)"

build-cli:
	go build -o bin/$(CLI_NAME) ./cmd/axiomod
	@echo "Built $(CLI_NAME)"

clean:
	go clean
	rm -rf bin/
	@echo "Cleaned build artifacts"

test:
	go test -v ./internal/... ./tests/...

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
	@echo "  make build        - Build the server binary"
	@echo "  make build-cli    - Build the CLI tool"
	@echo "  make test         - Run all tests"
	@echo "  make clean        - Remove binaries"
	@echo "  make deps         - Install dependencies"
	@echo "  make lint         - Run linters"
