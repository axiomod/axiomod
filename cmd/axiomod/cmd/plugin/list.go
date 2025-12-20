package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// listCmd represents the plugin list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	Long: `List all installed plugins in the Axiomod framework.

Example:
  axiomod plugin list
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing installed plugins...")

		// Define plugin directory
		pluginDir := filepath.Join("internal", "plugins")

		// Check if plugin directory exists
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			fmt.Println("Plugin directory not found.")
			return
		}

		// Read plugin directory
		entries, err := os.ReadDir(pluginDir)
		if err != nil {
			fmt.Printf("Error reading plugin directory: %v\n", err)
			os.Exit(1)
		}

		// Filter and display plugins
		fmt.Println("Installed plugins:")
		pluginCount := 0
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "example_plugin" {
				// Skip the example_plugin and non-directories
				pluginPath := filepath.Join(pluginDir, entry.Name())

				// Check if it has a plugin.go file (simple heuristic)
				if _, err := os.Stat(filepath.Join(pluginPath, entry.Name()+".go")); err == nil {
					fmt.Printf("- %s\n", entry.Name())
					pluginCount++
				}
			}
		}

		if pluginCount == 0 {
			fmt.Println("No plugins installed.")
		} else {
			fmt.Printf("\nTotal plugins: %d\n", pluginCount)
		}
	},
}

// NewListCmd returns the plugin list command.
func NewListCmd() *cobra.Command {
	return listCmd
}

func init() {
	// Add subcommands to the parent pluginCmd
	pluginCmd.AddCommand(listCmd)
}
