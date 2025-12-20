package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// dockerizeCmd represents the dockerize command
var dockerizeCmd = &cobra.Command{
	Use:   "dockerize",
	Short: "Create a Dockerfile and build a Docker image for the application",
	Long: `Create a Dockerfile for the Go Macroservice application and build a Docker image.

This command generates a multi-stage Dockerfile optimized for Go applications
and then builds a Docker image tagged as 'go-axiomod:latest'.

Example:
  axiomod dockerize
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating Dockerfile...")

		// Dockerfile content (multi-stage build)
		dockerfileContent := `
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Statically link to produce a smaller binary
# Use -tags netgo to ensure static linking of network libraries
# Use -ldflags "-s -w" to strip debug information
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -ldflags="-s -w" -o /axiomod-server ./cmd/axiomod-server

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /axiomod-server .

# Copy configuration files (adjust path if needed)
COPY framework/config/service_default.yaml /config/config.yaml

# Expose port (adjust if your service uses a different port)
EXPOSE 8080

# Command to run the application
# Pass the config file path
CMD ["./axiomod-server", "--config=/config/config.yaml"]
`

		// Write Dockerfile
		err := os.WriteFile("Dockerfile", []byte(strings.TrimSpace(dockerfileContent)), 0644)
		if err != nil {
			fmt.Printf("Error writing Dockerfile: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Dockerfile generated successfully.")

		fmt.Println("\nBuilding Docker image...")
		dockerCmd := exec.Command("docker", "build", "-t", "go-axiomod:latest", ".")
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		if err := dockerCmd.Run(); err != nil {
			fmt.Printf("Docker build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nDocker image go-axiomod:latest built successfully.")
	},
}

// NewDockerizeCmd returns the dockerize command.
func NewDockerizeCmd() *cobra.Command {
	return dockerizeCmd
}
