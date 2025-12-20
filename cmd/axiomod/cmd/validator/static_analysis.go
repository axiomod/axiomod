package validator

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// staticAnalysisCmd represents the validator static-analysis command
var staticAnalysisCmd = &cobra.Command{
	Use:   "static-analysis",
	Short: "Run all static analysis tools (vet, gosec, staticcheck)",
	Long: `Run a comprehensive set of static analysis tools on the project.
This includes go vet, gosec, and staticcheck.

Example:
  axiomod validator static-analysis
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running all static analysis tools...")

		// Run go vet
		fmt.Println("\n--- Running go vet ---")
		vetCmd := exec.Command("go", "vet", "./...")
		vetCmd.Stdout = os.Stdout
		vetCmd.Stderr = os.Stderr
		if err := vetCmd.Run(); err != nil {
			fmt.Printf("go vet failed: %v\n", err)
			// Decide if this should be a fatal error
		}

		// Run gosec
		fmt.Println("\n--- Running gosec ---")
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
		secCmd := exec.Command("gosec", "./...")
		secCmd.Stdout = os.Stdout
		secCmd.Stderr = os.Stderr
		if err := secCmd.Run(); err != nil {
			fmt.Printf("gosec failed: %v\n", err)
			// Decide if this should be a fatal error
		}

		// Run staticcheck
		fmt.Println("\n--- Running staticcheck ---")
		// Check if staticcheck is installed
		_, err = exec.LookPath("staticcheck")
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
		checkCmd := exec.Command("staticcheck", "./...")
		checkCmd.Stdout = os.Stdout
		checkCmd.Stderr = os.Stderr
		if err := checkCmd.Run(); err != nil {
			fmt.Printf("staticcheck failed: %v\n", err)
			// Decide if this should be a fatal error
		}

		fmt.Println("\nStatic analysis completed.")
	},
}

// NewStaticAnalysisCmd returns the validator static-analysis command.
func NewStaticAnalysisCmd() *cobra.Command {
	return staticAnalysisCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(staticAnalysisCmd)
}
