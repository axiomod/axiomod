package generate

import (
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code for the project",
	Long: `Generate code for the Go Macroservice project.

This command has subcommands for generating different types of code.

Example:
  axiomod generate module --name=user
  axiomod generate service --name=auth
  axiomod generate handler --name=product
`,
}

// NewGenerateCmd returns the generate command.
func NewGenerateCmd() *cobra.Command {
	return generateCmd
}
