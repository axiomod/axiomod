package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// removeCmd represents the plugin remove command
var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an installed plugin",
	Long: `Remove an installed plugin from the Axiomod framework.

Example:
  axiomod plugin remove myplugin
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pluginName := args[0]
		fmt.Printf("Removing plugin: %s\n", pluginName)

		// Define plugin directory
		pluginDir := filepath.Join("plugins", pluginName)

		// Check if plugin directory exists
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			fmt.Printf("Plugin %s not found.\n", pluginName)
			return
		}

		// Remove the plugin directory
		fmt.Println("Removing plugin directory...")
		err := os.RemoveAll(pluginDir)
		if err != nil {
			fmt.Printf("Error removing plugin directory %s: %v\n", pluginDir, err)
			os.Exit(1)
		}

		// Update the main plugin registration file (e.g., plugins/builtin_plugins.go)
		// This part requires more complex code modification (AST parsing or simple text removal)
		// For simplicity, we will just print a message here.
		fmt.Println("\nPlugin directory removed.")
		fmt.Println("Please manually unregister the plugin from plugins/builtin_plugins.go")
		fmt.Println("and remove its options from cmd/axiomod-server/fx_options.go")

		fmt.Printf("\nPlugin %s removed successfully (manual unregistration required).\n", pluginName)
	},
}

// NewRemoveCmd returns the plugin remove command.
func NewRemoveCmd() *cobra.Command {
	return removeCmd
}

func init() {
	// Add subcommands to the parent pluginCmd
	pluginCmd.AddCommand(removeCmd)
}
