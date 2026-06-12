package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var dockerizeTag string

// dockerizeCmd represents the dockerize command
var dockerizeCmd = &cobra.Command{
	Use:   "dockerize",
	Short: "Create a Dockerfile and build a Docker image for the application",
	Long: `Create a multi-stage Dockerfile for the application and build a Docker
image. The image tag defaults to "<module-name>:latest".

Example:
  axiomod dockerize
  axiomod dockerize --tag=myservice:v1.0.0
`,
	Run: func(cmd *cobra.Command, args []string) {
		module, err := moduleName()
		if err != nil {
			fmt.Printf("Error reading go.mod: %v\n", err)
			os.Exit(1)
		}

		mainPkg, binaryName, err := detectMainPackage(module)
		if err != nil {
			fmt.Printf("Error locating main package: %v\n", err)
			os.Exit(1)
		}

		tag := dockerizeTag
		if tag == "" {
			tag = binaryName + ":latest"
		}

		fmt.Println("Generating Dockerfile...")

		dockerfileContent := fmt.Sprintf(`# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/%[1]s %[2]s

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
`, binaryName, mainPkg)

		if err := os.WriteFile("Dockerfile", []byte(dockerfileContent), 0644); err != nil {
			fmt.Printf("Error writing Dockerfile: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Dockerfile generated successfully.")

		fmt.Printf("\nBuilding Docker image %s ...\n", tag)
		dockerCmd := exec.Command("docker", "build", "-t", tag, ".")
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		if err := dockerCmd.Run(); err != nil {
			fmt.Printf("Docker build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nDocker image %s built successfully.\n", tag)
	},
}

// moduleName returns the module path's last element from go.mod in the
// current directory.
func moduleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			return filepath.Base(modulePath), nil
		}
	}
	return "", fmt.Errorf("no module directive found in go.mod")
}

// detectMainPackage finds the application entry point: cmd/<module> for
// scaffolded projects, falling back to cmd/axiomod-server for the framework
// repository, then any single directory under cmd/.
func detectMainPackage(module string) (pkg, binary string, err error) {
	candidates := []string{
		filepath.Join("cmd", module),
		filepath.Join("cmd", "axiomod-server"),
	}
	for _, dir := range candidates {
		if info, statErr := os.Stat(dir); statErr == nil && info.IsDir() {
			return "./" + filepath.ToSlash(dir), filepath.Base(dir), nil
		}
	}

	entries, readErr := os.ReadDir("cmd")
	if readErr != nil {
		return "", "", fmt.Errorf("no cmd/ directory found: %w", readErr)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return "./" + filepath.ToSlash(filepath.Join("cmd", entry.Name())), entry.Name(), nil
		}
	}
	return "", "", fmt.Errorf("no main package found under cmd/")
}

// NewDockerizeCmd returns the dockerize command.
func NewDockerizeCmd() *cobra.Command {
	dockerizeCmd.Flags().StringVarP(&dockerizeTag, "tag", "t", "",
		"image tag (default \"<module-name>:latest\")")
	return dockerizeCmd
}
