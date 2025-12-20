package core

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Run linters on the project",
	Long: `Run linters on the Go Macroservice project to ensure code quality.

Example:
  axiomod lint
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running linters...")

		// Check if golangci-lint is installed
		_, err := exec.LookPath("golangci-lint")
		if err != nil {
			fmt.Println("golangci-lint not found. Installing...")
			installCmd := exec.Command("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				fmt.Printf("Failed to install golangci-lint: %v\n", err)
				os.Exit(1)
			}
		}

		// Run golangci-lint
		lintCmd := exec.Command("golangci-lint", "run", "./...")
		lintCmd.Stdout = os.Stdout
		lintCmd.Stderr = os.Stderr

		fmt.Printf("Executing: %s\n", lintCmd.String())
		if err := lintCmd.Run(); err != nil {
			fmt.Printf("Linting failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nLinting completed successfully.")
	},
}

// NewLintCmd returns the lint command.
func NewLintCmd() *cobra.Command {
	return lintCmd
}
