# Makefile for Go Macroservice Framework

# Variables
APP_NAME := axiomod
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH) -X main.GoVersion=$(GO_VERSION)"

# Go related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(shell find . -type f -name "*.go" -not -path "./vendor/*")

# Docker related variables
DOCKER_IMAGE := axiomod
DOCKER_TAG := latest

# Environment
ENV ?= dev

# Default target
.PHONY: all
all: clean build test

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(GOBIN)
	@go build $(LDFLAGS) -o $(GOBIN)/$(APP_NAME) ./cmd/axiomod
	@echo "Build complete: $(GOBIN)/$(APP_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	@$(GOBIN)/$(APP_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(GOBIN)
	@go clean
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v -cover ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(GOBASE)/coverage
	@go test -v -coverprofile=$(GOBASE)/coverage/coverage.out ./...
	@go tool cover -html=$(GOBASE)/coverage/coverage.out -o $(GOBASE)/coverage/coverage.html
	@echo "Coverage report generated: $(GOBASE)/coverage/coverage.html"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@gofmt -s -w $(GOFILES)
	@echo "Formatting complete"

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container
.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Generate code (entities, repositories, etc.)
.PHONY: generate
generate:
	@echo "Generating code..."
	@go generate ./...
	@echo "Code generation complete"

# Create a new migration
.PHONY: migrate-create
migrate-create:
	@echo "Creating migration..."
	@./bin/axiomod migrate create $(name)

# Apply migrations
.PHONY: migrate-up
migrate-up:
	@echo "Applying migrations..."
	@./bin/axiomod migrate up

# Rollback migrations
.PHONY: migrate-down
migrate-down:
	@echo "Rolling back migrations..."
	@./bin/axiomod migrate down

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Dependencies installed"

# Build CLI tool
.PHONY: build-cli
build-cli:
	@echo "Building axiomod CLI tool..."
	@mkdir -p $(GOBIN)
	@go build $(LDFLAGS) -o $(GOBIN)/axiomod ./cmd/axiomod
	@echo "CLI tool built: $(GOBIN)/axiomod"

# Deploy to environment
.PHONY: deploy
deploy:
	@echo "Deploying to $(ENV) environment..."
	@./bin/axiomod deploy $(ENV)

# Check status
.PHONY: status
status:
	@echo "Checking status..."
	@./bin/axiomod status --env $(ENV)

# View logs
.PHONY: logs
logs:
	@echo "Viewing logs..."
	@./bin/axiomod logs --env $(ENV) $(service)

# Health check
.PHONY: healthcheck
healthcheck:
	@echo "Running health check..."
	@./bin/axiomod healthcheck --env $(ENV)

# Help target
.PHONY: help
help:
	@echo "Go Macroservice Framework Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all             Build, test, and lint the application"
	@echo "  build           Build the application"
	@echo "  build-cli       Build the axiomod CLI tool"
	@echo "  clean           Clean build artifacts"
	@echo "  deps            Install dependencies"
	@echo "  deploy          Deploy to environment (ENV=dev|staging|prod)"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-run      Run Docker container"
	@echo "  fmt             Format code"
	@echo "  generate        Generate code"
	@echo "  healthcheck     Run health check"
	@echo "  help            Show this help message"
	@echo "  lint            Run linter"
	@echo "  logs            View logs (service=service-name)"
	@echo "  migrate-create  Create a new migration (name=migration_name)"
	@echo "  migrate-down    Rollback migrations"
	@echo "  migrate-up      Apply migrations"
	@echo "  run             Run the application"
	@echo "  status          Check status"
	@echo "  test            Run tests"
	@echo "  test-coverage   Run tests with coverage report"
