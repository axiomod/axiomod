package core

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage configuration for your project.

Examples:
  axiomod config validate
  axiomod config diff dev prod`,
}

// NewConfigCmd returns the config command
func NewConfigCmd() *cobra.Command {
	return configCmd
}
