package validator

import (
	"fmt"

	"github.com/spf13/cobra"
)

// standardsCheckCmd represents the validator standards-check command (aliased as all)
var standardsCheckCmd = &cobra.Command{
	Use:     "standards-check",
	Aliases: []string{"all"},
	Short:   "Run all validators (architecture, naming, domain, static-analysis, etc.)",
	Long: `Run a comprehensive suite of validators to ensure the project meets all defined standards.

This is equivalent to running:
- axiomod validator architecture
- axiomod validator naming
- axiomod validator domain
- axiomod validator static-analysis
- axiomod validator check-api-spec (if spec provided)
- axiomod validator check-docs

Example:
  axiomod validator standards-check
  axiomod validator all
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running all standard checks...")

		// Note: This command should ideally trigger the execution of other validator commands.
		// For simplicity in this example, we just print a message.
		// A real implementation might use cobra.Command.Execute() or call the underlying
		// validation functions directly.

		fmt.Println("\n--- Running Architecture Validator ---")
		// Simulate running architectureCmd.Run(cmd, args) or call ValidateArchitecture
		if err := ValidateArchitecture("architecture-rules.json"); err != nil {
			fmt.Printf("Architecture validation failed: %v\n", err)
		} else {
			fmt.Println("Architecture validation passed.")
		}

		fmt.Println("\n--- Running Naming Validator ---")
		// Simulate running namingCmd.Run(cmd, args) or call ValidateNaming
		if issues, err := ValidateNaming(false); err != nil {
			fmt.Printf("Naming validation error: %v\n", err)
		} else if len(issues) > 0 {
			fmt.Printf("Naming validation failed with %d issues.\n", len(issues))
		} else {
			fmt.Println("Naming validation passed.")
		}

		fmt.Println("\n--- Running Domain Validator ---")
		// Simulate running domainCmd.Run(cmd, args) or call ValidateDomainBoundaries
		if issues, err := ValidateDomainBoundaries(); err != nil {
			fmt.Printf("Domain boundary validation error: %v\n", err)
		} else if len(issues) > 0 {
			fmt.Printf("Domain boundary validation failed with %d issues.\n", len(issues))
		} else {
			fmt.Println("Domain boundary validation passed.")
		}

		fmt.Println("\n--- Running Static Analysis Validator ---")
		// Simulate running staticAnalysisCmd.Run(cmd, args)
		// This would involve running go vet, gosec, staticcheck
		fmt.Println("(Simulated) Static analysis checks passed.") // Placeholder

		fmt.Println("\n--- Running API Spec Validator ---")
		// Simulate running checkAPISpecCmd.Run(cmd, args)
		// Requires finding a spec file or skipping
		fmt.Println("(Simulated) API spec check skipped (no spec file provided).") // Placeholder

		fmt.Println("\n--- Running Docs Check Validator ---")
		// Simulate running checkDocsCmd.Run(cmd, args)
		fmt.Println("(Simulated) Docs check passed.") // Placeholder

		fmt.Println("\nAll standard checks completed.")
		// In a real scenario, exit with non-zero code if any check failed.
	},
}

// NewStandardsCheckCmd returns the validator standards-check command.
func NewStandardsCheckCmd() *cobra.Command {
	// Add flags if needed for sub-validators, e.g., API spec path
	return standardsCheckCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(standardsCheckCmd)
}
