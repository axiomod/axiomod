package validator

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// securityCmd represents the validator security command
var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Run gosec security scanner",
	Long: `Run the gosec security scanner to identify potential security vulnerabilities.

Example:
  axiomod validator security
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running gosec security scanner...")

		// Check if gosec is installed
		_, err := exec.LookPath("gosec")
		if err != nil {
			fmt.Println("gosec not found. Installing...")
			installCmd := exec.Command("go", "install", "github.com/securego/gosec/v2/cmd/gosec@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				fmt.Printf("Failed to install gosec: %v\n", err)
				os.Exit(1)
			}
		}

		// Run gosec
		secCmd := exec.Command("gosec", "./...")
		secCmd.Stdout = os.Stdout
		secCmd.Stderr = os.Stderr
		if err := secCmd.Run(); err != nil {
			fmt.Printf("gosec failed: %v\n", err)
			os.Exit(1) // Security issues should likely cause a failure
		}

		fmt.Println("gosec security scan completed successfully.")
	},
}

// NewSecurityCmd returns the validator security command.
func NewSecurityCmd() *cobra.Command {
	return securityCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(securityCmd)
}
