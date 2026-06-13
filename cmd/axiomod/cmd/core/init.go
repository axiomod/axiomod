package core

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/axiomod/axiomod/framework/utils"
	"github.com/axiomod/axiomod/framework/version"

	"github.com/spf13/cobra"
)

// frameworkModulePath is the canonical module path of the Axiomod framework.
const frameworkModulePath = "github.com/axiomod/axiomod"

// devPseudoVersion is the placeholder version used together with a replace
// directive when scaffolding against a local framework checkout.
const devPseudoVersion = "v0.0.0-00010101000000-000000000000"

var (
	initDevMode       bool
	initFrameworkPath string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new Axiomod service project",
	Long: `Initialize a new Axiomod service project with the recommended structure.

The generated go.mod pins the framework version this CLI was built from, so
"go mod tidy && go build ./..." works out of the box.

Use --dev to develop against a local framework checkout: the framework is
resolved from --framework-path, or discovered by walking up from the current
directory, and wired in with a replace directive.

Examples:
  axiomod init myservice
  axiomod init myservice --dev --framework-path /path/to/axiomod
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		fmt.Printf("Initializing new Axiomod project: %s\n", projectName)

		// Resolve the framework location before changing directories.
		frameworkPath := ""
		if initDevMode {
			resolved, err := resolveFrameworkPath(initFrameworkPath)
			if err != nil {
				fmt.Printf("Error resolving framework path for --dev: %v\n", err)
				os.Exit(1)
			}
			frameworkPath = resolved
		}

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

		// Create directory structure
		dirs := []string{
			"cmd/" + projectName,
			"configs",
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

		// Write go.mod with pinned requirements (and a replace in --dev mode).
		if err := writeGoMod(projectName, frameworkPath); err != nil {
			fmt.Printf("Error creating go.mod: %v\n", err)
			os.Exit(1)
		}

		// Create basic files
		createBasicFiles(projectName)

		fmt.Printf("\nProject %s initialized successfully!\n", projectName)
		fmt.Println("\nNext steps:")
		fmt.Println("1. cd " + projectName)
		fmt.Println("2. go mod tidy")
		fmt.Println("3. Update configs/service_default.yaml with your settings (e.g., database DSN)")
		fmt.Println("4. axiomod migrate create initial_schema")
		fmt.Println("5. axiomod migrate up")
		fmt.Println("6. go run ./cmd/" + projectName)
	},
}

// resolveFrameworkPath returns the absolute path of a local framework
// checkout: the explicit flag value when given, otherwise the first ancestor
// of the working directory whose go.mod declares the framework module.
func resolveFrameworkPath(explicit string) (string, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", fmt.Errorf("invalid --framework-path %q: %w", explicit, err)
		}
		if !isFrameworkRoot(abs) {
			return "", fmt.Errorf("%s does not contain the %s go.mod", abs, frameworkModulePath)
		}
		return abs, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if isFrameworkRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s checkout found above the current directory; pass --framework-path", frameworkModulePath)
		}
		dir = parent
	}
}

// isFrameworkRoot reports whether dir holds the framework's go.mod.
func isFrameworkRoot(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "module "+frameworkModulePath)
}

// writeGoMod renders the project's go.mod with pinned dependency versions.
// In dev mode the framework is required at a pseudo-version and replaced with
// the local checkout.
func writeGoMod(projectName, frameworkPath string) error {
	frameworkVersion := version.Version
	if frameworkPath != "" {
		frameworkVersion = devPseudoVersion
	} else if !isSemver(frameworkVersion) {
		// CLI built without ldflags: there is no tagged version to pin.
		fmt.Println("Warning: CLI has no release version baked in; run 'go get " +
			frameworkModulePath + "@latest' inside the project, or use --dev for a local checkout.")
		frameworkVersion = devPseudoVersion
	}

	var b strings.Builder
	fmt.Fprintf(&b, "module %s\n\n", projectName)
	fmt.Fprintf(&b, "go %s\n\n", goDirectiveVersion())
	b.WriteString("require (\n")
	fmt.Fprintf(&b, "\t%s %s\n", frameworkModulePath, frameworkVersion)
	for _, dep := range []string{"go.uber.org/fx", "go.uber.org/zap"} {
		if v := buildDependencyVersion(dep); v != "" {
			fmt.Fprintf(&b, "\t%s %s\n", dep, v)
		}
	}
	b.WriteString(")\n")

	// Bootstrap pins: the framework's dependency graph contains an old
	// unpruned module (go-grpc-middleware) whose test imports make
	// cloud.google.com/go/compute/metadata ambiguous during the first
	// `go mod tidy`. These pins resolve it; tidy rewrites them afterwards.
	b.WriteString("\nrequire (\n")
	b.WriteString("\tcloud.google.com/go v0.112.1 // indirect\n")
	b.WriteString("\tgolang.org/x/oauth2 v0.32.0 // indirect\n")
	b.WriteString(")\n")

	if frameworkPath != "" {
		fmt.Fprintf(&b, "\nreplace %s => %s\n", frameworkModulePath, frameworkPath)
	}

	return os.WriteFile("go.mod", []byte(b.String()), 0644)
}

// isSemver reports whether s looks like a tagged semantic version (vX.Y.Z).
func isSemver(s string) bool {
	if len(s) < 2 || s[0] != 'v' {
		return false
	}
	return s[1] >= '0' && s[1] <= '9'
}

// minGoDirective is the minimum "go" directive for generated projects; the
// framework module itself requires this language version.
const minGoDirective = "1.25"

// goDirectiveVersion derives the go.mod "go" directive from the running
// toolchain (e.g. "go1.25.11" -> "1.25"), floored at the framework's own
// language requirement.
func goDirectiveVersion() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return minGoDirective
	}
	derived := parts[0] + "." + parts[1]

	var dMaj, dMin, mMaj, mMin int
	if _, err := fmt.Sscanf(derived, "%d.%d", &dMaj, &dMin); err != nil {
		return minGoDirective
	}
	if _, err := fmt.Sscanf(minGoDirective, "%d.%d", &mMaj, &mMin); err != nil {
		return derived
	}
	if dMaj < mMaj || (dMaj == mMaj && dMin < mMin) {
		return minGoDirective
	}
	return derived
}

// buildDependencyVersion looks up the version of a dependency baked into this
// CLI binary, keeping generated projects aligned with the framework's own
// dependency graph.
func buildDependencyVersion(modulePath string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath {
			return dep.Version
		}
	}
	return ""
}

// writeGoFile formats Go source with go/format before writing, so generated
// code is always gofmt-clean.
func writeGoFile(path string, src []byte) error {
	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("generated %s does not compile-format: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0644)
}

// createBasicFiles creates basic files for the project
func createBasicFiles(projectName string) {
	// Create main.go
	mainContent := `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	grpcpkg "github.com/axiomod/axiomod/framework/grpc"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/framework/worker"
	"github.com/axiomod/axiomod/platform/server"
	"github.com/axiomod/axiomod/plugins"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "", "path to config file (default: search configs/service_default.yaml)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	app := fx.New(
		fx.Provide(func() *config.Config { return cfg }),

		// Core framework modules
		observability.Module,
		middleware.Module,
		auth.Module,
		health.Module,
		grpcpkg.Module,
		server.Module,
		plugins.Module,
		worker.Module,

		// Domain modules
		// Add your domain modules here, for example:
		// example.Module,

		// Start the HTTP and gRPC servers
		fx.Invoke(server.RegisterHTTPServer, server.RegisterGRPCServer),

		fx.Invoke(func(lc fx.Lifecycle, logger *observability.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Application started", zap.String("service", cfg.App.Name))
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Application stopping")
					return nil
				},
			})
		}),
	)

	startCtx, cancelStart := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStart()
	if err := app.Start(startCtx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	stopCtx, cancelStop := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStop()
	if err := app.Stop(stopCtx); err != nil {
		fmt.Fprintf(os.Stderr, "error during shutdown: %v\n", err)
		os.Exit(1)
	}
}
`

	if err := writeGoFile(filepath.Join("cmd", projectName, "main.go"), []byte(mainContent)); err != nil {
		fmt.Printf("Error creating main.go: %v\n", err)
		os.Exit(1)
	}

	// Create configs/service_default.yaml following the framework conventions.
	configContent := `app:
  name: "%s"
  environment: "development"
  version: "0.1.0"
  debug: true

observability:
  logLevel: "info"
  logFormat: "json"
  tracingEnabled: false
  tracingExporterType: "stdout" # Options: jaeger, otlp, stdout
  tracingSamplerRatio: 1.0
  metricsEnabled: true # served at /metrics on the HTTP port

http:
  host: "0.0.0.0"
  port: 8080
  readTimeout: 10
  writeTimeout: 10

grpc:
  host: "0.0.0.0"
  port: 9090

database:
  driver: "postgres" # Options: postgres, mysql
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "%s"
  sslMode: "disable"
  orm: "ent" # Options: ent (default), sql

auth:
  jwt:
    # DEV PLACEHOLDER -- override in every deployment, e.g. via
    # APP_AUTH_JWT_SECRETKEY. Minimum 32 bytes; startup fails otherwise.
    secretKey: "dev-only-insecure-secret-change-me-now"
    tokenDuration: 60 # minutes

plugins:
  enabled:
    postgres: false # enable when a database is reachable
    mysql: false
    jwt: false
  settings: {}
`
	if err := os.WriteFile(filepath.Join("configs", "service_default.yaml"),
		[]byte(fmt.Sprintf(configContent, projectName, projectName)), 0644); err != nil {
		fmt.Printf("Error creating service_default.yaml: %v\n", err)
		os.Exit(1)
	}

	// Create Dockerfile
	dockerfileContent := fmt.Sprintf(`# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /out/%[1]s ./cmd/%[1]s

# --- Runtime stage ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /out/%[1]s /app/%[1]s
COPY configs/ /app/configs/

USER app

# HTTP (Prometheus metrics at /metrics) and gRPC
EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
  CMD wget -qO- http://127.0.0.1:8080/live || exit 1

ENTRYPOINT ["/app/%[1]s", "-config", "/app/configs/service_default.yaml"]
`, projectName)
	if err := os.WriteFile("Dockerfile", []byte(dockerfileContent), 0644); err != nil {
		fmt.Printf("Error creating Dockerfile: %v\n", err)
		os.Exit(1)
	}

	// Create README.md
	readmeContent := fmt.Sprintf(`# %s

A Go service built with the [Axiomod](https://github.com/axiomod/axiomod) framework.

## Features

- Modular architecture with clean separation of concerns
- Dependency injection using Uber FX
- Observability with logging, metrics, and tracing
- Pluggable components for flexibility
- HTTP and gRPC API support

## Getting Started

### Prerequisites

- Go 1.25+
- Docker (optional, for containerization and dependencies like Postgres/Jaeger)

### Building

`+"```bash"+`
go mod tidy
go build -o bin/%s ./cmd/%s
`+"```"+`

### Running

`+"```bash"+`
# Run with the default configuration (configs/service_default.yaml)
./bin/%s

# Run with a custom configuration
./bin/%s -config=path/to/config.yaml
`+"```"+`

Health probes are served at /live and /ready; Prometheus metrics at /metrics.

## Project Structure

- `+"`cmd/%s`"+`: Application entry point
- `+"`configs`"+`: Configuration files (service_default.yaml)
- `+"`internal/domain`"+`: Business entities and rules
- `+"`internal/usecase`"+`: Application-specific business rules
- `+"`internal/infrastructure`"+`: Implementation details (DB repositories, external APIs)
- `+"`tests`"+`: Unit and integration tests
- `+"`migrations`"+`: Database migration files

## License

This project is licensed under the MIT License - see the LICENSE file for details.
`, utils.TitleCase(projectName), projectName, projectName, projectName, projectName, projectName)

	if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
		fmt.Printf("Error creating README.md: %v\n", err)
		os.Exit(1)
	}

	// Create .gitignore
	gitignoreContent := `# Binaries
*.exe
*.dll
*.so
*.dylib
bin/

# Test artifacts
*.test
*.out

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
service_local.yaml
`
	if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0644); err != nil {
		fmt.Printf("Error creating .gitignore: %v\n", err)
		os.Exit(1)
	}

	// Create basic Makefile
	makefileContent := fmt.Sprintf(`BINARY_NAME=%s

.PHONY: all build clean test deps lint fmt help

all: build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)
	@echo "Built $(BINARY_NAME) binary"

clean:
	go clean
	rm -rf bin/
	@echo "Cleaned build artifacts"

test:
	go test -v ./...

deps:
	go mod tidy
	go mod download
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

help:
	@echo "Available commands:"
	@echo "  make build        - Build the service binary"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make deps         - Install dependencies and tools"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format Go code"
`, projectName)
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		fmt.Printf("Error creating Makefile: %v\n", err)
		os.Exit(1)
	}

	// Create basic LICENSE file (MIT)
	licenseContent := fmt.Sprintf(`MIT License

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
`, time.Now().Year())
	if err := os.WriteFile("LICENSE", []byte(licenseContent), 0644); err != nil {
		fmt.Printf("Error creating LICENSE: %v\n", err)
		os.Exit(1)
	}
}

// NewInitCmd returns the init command.
func NewInitCmd() *cobra.Command {
	initCmd.Flags().BoolVar(&initDevMode, "dev", false,
		"develop against a local framework checkout via a replace directive")
	initCmd.Flags().StringVar(&initFrameworkPath, "framework-path", "",
		"path to the local framework checkout (with --dev; default: discovered from parent directories)")
	return initCmd
}
