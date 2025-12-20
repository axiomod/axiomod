package validator

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// namingCmd represents the validator naming command
var namingCmd = &cobra.Command{
	Use:   "naming",
	Short: "Validate naming conventions across the project",
	Long: `Validate naming conventions for Go code, API endpoints, and database schemas.

Example:
  axiomod validator naming
  axiomod validator naming --fix
`,
	Run: func(cmd *cobra.Command, args []string) {
		fix, _ := cmd.Flags().GetBool("fix")
		fmt.Println("Validating naming conventions...")

		if fix {
			fmt.Println("Fix mode enabled - will attempt to correct naming issues")
		}

		// Perform validation
		issues, err := ValidateNaming(fix)
		if err != nil {
			fmt.Printf("Naming validation error: %v\n", err)
			os.Exit(1)
		}

		if len(issues) > 0 {
			fmt.Printf("Found %d naming convention issues:\n", len(issues))
			for i, issue := range issues {
				fmt.Printf("%d. %s\n", i+1, issue)
			}

			if !fix {
				fmt.Println("\nRun with --fix flag to attempt automatic fixes")
				os.Exit(1)
			} else {
				fmt.Println("\nAttempted to fix issues. Please review changes.")
			}
		} else {
			fmt.Println("Naming validation passed successfully.")
		}
	},
}

// NewNamingCmd returns the validator naming command.
func NewNamingCmd() *cobra.Command {
	namingCmd.Flags().BoolP("fix", "f", false, "Attempt to fix naming issues automatically")
	return namingCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(namingCmd)
}
