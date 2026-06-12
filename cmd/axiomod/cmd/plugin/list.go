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

		pluginDir := "plugins"
		if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
			fmt.Println("Plugin directory not found.")
			return
		}

		// A plugin is any directory under plugins/ containing a plugin.go
		// (e.g. audit, auth/ldap, cache/redis). The built-in database/auth
		// plugins live in plugins/builtin_plugins.go.
		var found []string
		err := filepath.WalkDir(pluginDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() || path == pluginDir {
				return nil
			}
			if d.Name() == "example_plugin" {
				return filepath.SkipDir
			}
			if _, statErr := os.Stat(filepath.Join(path, "plugin.go")); statErr == nil {
				rel, relErr := filepath.Rel(pluginDir, path)
				if relErr == nil {
					found = append(found, filepath.ToSlash(rel))
				}
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error scanning plugin directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Installed plugins:")
		for _, name := range found {
			fmt.Printf("- %s\n", name)
		}
		fmt.Println("- builtin (mysql, postgres, jwt, keycloak, casdoor, casbin in plugins/builtin_plugins.go)")

		if len(found) == 0 {
			fmt.Println("No standalone plugins installed.")
		} else {
			fmt.Printf("\nTotal standalone plugins: %d\n", len(found))
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
