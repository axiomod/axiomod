package plugin

import (
	"github.com/spf13/cobra"
)

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage Axiomod plugins",
	Long: `Manage Axiomod plugins.

This command has subcommands for listing, installing, and removing plugins.

Example:
  axiomod plugin list
  axiomod plugin install <plugin_source>
  axiomod plugin remove <plugin_name>
`,
}

// NewPluginCmd returns the plugin command.
func NewPluginCmd() *cobra.Command {
	return pluginCmd
}
