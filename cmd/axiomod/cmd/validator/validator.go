package validator

import (
	"github.com/spf13/cobra"
)

// validatorCmd represents the validator command
var validatorCmd = &cobra.Command{
	Use:   "validator",
	Short: "Run various validators on the project",
	Long: `Run various validators on the Go Macroservice project to ensure code quality, 
consistency, and adherence to architectural rules.

Includes checks for:
- Architecture rules
- Naming conventions
- Domain boundaries
- Static analysis (vet, staticcheck)
- Security vulnerabilities (gosec)
- API specification standards
- Documentation updates

Example:
  axiomod validator all
  axiomod validator architecture
  axiomod validator naming --fix
`,
}

// NewValidatorCmd returns the validator command.
func NewValidatorCmd() *cobra.Command {
	// Add subcommands in their respective files using init()
	return validatorCmd
}
