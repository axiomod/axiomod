package core

import (
	"github.com/spf13/cobra"

	configcmd "github.com/axiomod/axiomod/cmd/axiomod/cmd/core/config"
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

// NewConfigCmd returns the config command with its subcommands attached.
func NewConfigCmd() *cobra.Command {
	configCmd.AddCommand(configcmd.NewConfigValidateCmd())
	configCmd.AddCommand(configcmd.NewConfigDiffCmd())
	return configCmd
}
