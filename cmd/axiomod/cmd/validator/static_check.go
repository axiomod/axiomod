package validator

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// staticCheckCmd represents the validator static-check command
var staticCheckCmd = &cobra.Command{
	Use:   "static-check",
	Short: "Run staticcheck static analyzer",
	Long: `Run the staticcheck static analyzer on the project.

Example:
  axiomod validator static-check
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running staticcheck...")

		// Check if staticcheck is installed
		_, err := exec.LookPath("staticcheck")
		if err != nil {
			fmt.Println("staticcheck not found. Installing...")
			installCmd := exec.Command("go", "install", "honnef.co/go/tools/cmd/staticcheck@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				fmt.Printf("Failed to install staticcheck: %v\n", err)
				os.Exit(1)
			}
		}

		// Run staticcheck
		checkCmd := exec.Command("staticcheck", "./...")
		checkCmd.Stdout = os.Stdout
		checkCmd.Stderr = os.Stderr
		if err := checkCmd.Run(); err != nil {
			fmt.Printf("staticcheck failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("staticcheck completed successfully.")
	},
}

// NewStaticCheckCmd returns the validator static-check command.
func NewStaticCheckCmd() *cobra.Command {
	return staticCheckCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(staticCheckCmd)
}
