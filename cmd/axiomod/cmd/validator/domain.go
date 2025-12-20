package validator

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// domainCmd represents the validator domain command
var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Validate domain boundaries and dependencies",
	Long: `Validate that domain boundaries are respected according to defined rules.

This check helps ensure modules do not improperly depend on each other.
Rules are typically defined within the code or a configuration file.

Example:
  axiomod validator domain
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Validating domain boundaries...")

		// Perform validation
		issues, err := ValidateDomainBoundaries()
		if err != nil {
			fmt.Printf("Domain boundary validation error: %v\n", err)
			os.Exit(1)
		}

		if len(issues) > 0 {
			fmt.Printf("Found %d domain boundary violations:\n", len(issues))
			for i, issue := range issues {
				fmt.Printf("%d. %s\n", i+1, issue)
			}
			os.Exit(1)
		} else {
			fmt.Println("Domain boundary validation passed successfully.")
		}
	},
}

// NewDomainCmd returns the validator domain command.
func NewDomainCmd() *cobra.Command {
	// Add flags if needed, e.g., for configuration
	// domainCmd.Flags().StringP("config", "c", "", "Path to domain rules configuration")
	return domainCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(domainCmd)
}
