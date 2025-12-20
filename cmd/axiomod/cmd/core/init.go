package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new Go Macroservice project",
	Long: `Initialize a new Go Macroservice project with the recommended structure.

Example:
  axiomod init myservice
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		fmt.Printf("Initializing new Go Macroservice project: %s\n", projectName)

		// Create project directory
		if err := os.MkdirAll(projectName, 0755); err != nil {
			fmt.Printf("Error creating project directory: %v\n", err)
			os.Exit(1)
		}

		// Change to project directory
		if err := os.Chdir(projectName); err != nil {
			fmt.Printf("Error changing to project directory: %v\n", err)
			os.Exit(1)
		}

		// Initialize Go module
		modInit := exec.Command("go", "mod", "init", projectName)
		modInit.Stdout = os.Stdout
		modInit.Stderr = os.Stderr
		if err := modInit.Run(); err != nil {
			fmt.Printf("Error initializing Go module: %v\n", err)
			os.Exit(1)
		}

		// Add replace directive for local development (optional, but helpful for testing)
		// This points to the parent directory where axiomod is located
		replaceCmd := exec.Command("go", "mod", "edit", "-replace", "github.com/axiomod/axiomod=../")
		replaceCmd.Stdout = os.Stdout
		replaceCmd.Stderr = os.Stderr
		if err := replaceCmd.Run(); err != nil {
			fmt.Printf("Warning: could not add direct replace directive: %v\n", err)
		}

		// Create directory structure
		dirs := []string{
			"cmd/" + projectName,
			"cmd/" + projectName,
			"internal/domain",
			"internal/usecase",
			"internal/infrastructure",
			"tests/unit",
			"tests/integration",
			"docs",
			"scripts",
			"migrations",
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}

		// Create basic files
		createBasicFiles(projectName)

		fmt.Printf("\nProject %s initialized successfully!\n", projectName)
		fmt.Println("\nNext steps:")
		fmt.Println("1. cd " + projectName)
		fmt.Println("2. go mod tidy")
		fmt.Println("3. Update framework/config/service_default.yaml with your settings (e.g., database DSN)")
		fmt.Println("4. axiomod migrate create initial_schema")
		fmt.Println("5. axiomod migrate up")
		fmt.Println("6. go run ./cmd/" + projectName)
	},
}

// createBasicFiles creates basic files for the project
func createBasicFiles(projectName string) {
	// Create main.go
	mainContent := fmt.Sprintf(`package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/worker"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/axiomod/axiomod/platform/server"
	"github.com/axiomod/axiomod/plugins"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to config file (default: config/service_default.yaml)")
	flag.Parse()

	// Determine config path
	resolvedConfigPath := *configPath
	 if resolvedConfigPath == "" {
		 resolvedConfigPath = "config/service_default.yaml"
	 }

	// Load initial config just for logger setup (optional, can use default)
	 tempCfg, err := config.Load(resolvedConfigPath)
	 if err != nil {
		 fmt.Printf("Warning: could not load config for initial logger setup: %%v\n", err)
		// Use default logger config or handle error
	}

	// Setup initial logger (can be replaced by FX provided logger later)
	 initialLogger, err := observability.NewLogger(tempCfg)
	 if err != nil {
		 fmt.Printf("Error creating initial logger: %%v\n", err)
		 os.Exit(1)
	}
	 initialLogger.Info("Starting application", zap.String("configPath", resolvedConfigPath))

	// Create application with dependencies
	 app := fx.New(
		// Provide the configuration
		fx.Provide(func() (*config.Config, error) {
			return config.Load(resolvedConfigPath)
		}),

		// Core platform modules
		observability.Module,
		server.Module,
		plugins.Module,
		worker.Module,

		// Register lifecycle hooks
		fx.Invoke(func(lc fx.Lifecycle, logger *observability.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Starting application")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Stopping application")
					return nil
				},
			})
		}),
	)

	// Start the application
	 ctx, cancel := context.WithCancel(context.Background())
	 defer cancel()

	// Handle signals
	 go func() {
		 sigCh := make(chan os.Signal, 1)
		 signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		 <-sigCh
		 initialLogger.Info("Received signal, shutting down...")
		 cancel()
	 }()

	// Start and wait for context cancellation
	 if err := app.Start(ctx); err != nil {
		 initialLogger.Fatal("Error starting application", zap.Error(err))
	}
	 <-ctx.Done()

	// Stop the application
	 initialLogger.Info("Initiating graceful shutdown...")
	 stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second) // Add timeout
	 defer stopCancel()
	 if err := app.Stop(stopCtx); err != nil {
		 initialLogger.Error("Error stopping application", zap.Error(err))
		 os.Exit(1)
	}
	 initialLogger.Info("Application stopped gracefully")
}
`)

	err := os.WriteFile(filepath.Join("cmd", projectName, "main.go"), []byte(mainContent), 0644)
	if err != nil {
		fmt.Printf("Error creating main.go: %v\n", err)
		os.Exit(1)
	}

	// Create config.yaml
	configContent := `app:
  name: "%s"
  environment: development
  version: 1.0.0
  debug: true

http:
  host: 0.0.0.0
  port: 8080
  readTimeout: 30
  writeTimeout: 30

grpc:
  host: 0.0.0.0
  port: 50051

observability:
  logLevel: info
  logFormat: text
  tracingEnabled: false
  tracingURL: "http://localhost:14268/api/traces"
  metricsEnabled: true
  metricsPort: 9090

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "axiomod"
  sslMode: "disable"
`
	err = os.MkdirAll(filepath.Join("config"), 0755)
	if err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}
	err = os.WriteFile(filepath.Join("config", "service_default.yaml"), []byte(fmt.Sprintf(configContent, projectName)), 0644)
	if err != nil {
		fmt.Printf("Error creating service_default.yaml: %v\n", err)
		os.Exit(1)
	}

	// Create README.md
	// Escape backticks for Go string literal
	readmeContent := fmt.Sprintf(`# %s

A Go Macroservice application built with the Axiomod framework.

## Features

- Modular architecture with clean separation of concerns
- Dependency injection using Uber FX
- Observability with logging, metrics, and tracing
- Pluggable components for flexibility
- HTTP and gRPC API support
- Comprehensive testing

## Getting Started

### Prerequisites

- Go 1.21+
- Docker (optional, for containerization and dependencies like Postgres/Jaeger)

### Building

`+"```bash"+`
# Build the project
go build -o bin/%s ./cmd/%s
`+"```"+`

### Running

`+"```bash"+`
# Ensure dependencies (e.g., PostgreSQL, Jaeger) are running

# Run with default configuration
./bin/%s

# Run with custom configuration
./bin/%s --config=path/to/config.yaml
`+"```"+`

## Project Structure

The project follows a clean architecture approach with the following structure:

- `+"`cmd/%s`"+`: Application entry point
- `+"`internal/domain`"+`: Business entities and rules
- `+"`internal/usecase`"+`: Application-specific business rules
- `+"`internal/infrastructure`"+`: Implementation details (DB repositories, external APIs)
- `+"`tests`"+`: Unit and integration tests
- `+"`docs`"+`: Documentation
- `+"`scripts`"+`: Build and deployment scripts
- `+"`migrations`"+`: Database migration files

## License

This project is licensed under the MIT License - see the LICENSE file for details.
`, strings.Title(projectName), projectName, projectName, projectName, projectName, projectName)

	err = os.WriteFile("README.md", []byte(readmeContent), 0644)
	if err != nil {
		fmt.Printf("Error creating README.md: %v\n", err)
		os.Exit(1)
	}

	// Create .gitignore
	gitignoreContent := `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# IDE files
.idea/
.vscode/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Config files with sensitive information
*.env
service_config.yaml
`
	err = os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)
	if err != nil {
		fmt.Printf("Error creating .gitignore: %v\n", err)
		os.Exit(1)
	}

	// Create basic Makefile
	makefileContent := `GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=%s
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
`
	err = os.WriteFile("Makefile", []byte(fmt.Sprintf(makefileContent, projectName, projectName)), 0644)
	if err != nil {
		fmt.Printf("Error creating Makefile: %v\n", err)
		os.Exit(1)
	}

	// Create basic LICENSE file (MIT)
	licenseContent := `MIT License

Copyright (c) %d Your Name or Company

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
	err = os.WriteFile("LICENSE", []byte(fmt.Sprintf(licenseContent, time.Now().Year())), 0644)
	if err != nil {
		fmt.Printf("Error creating LICENSE: %v\n", err)
		os.Exit(1)
	}
}

// NewInitCmd returns the init command.
func NewInitCmd() *cobra.Command {
	return initCmd
}
