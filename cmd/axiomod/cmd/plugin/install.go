package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// installCmd represents the plugin install command
var installCmd = &cobra.Command{
	Use:   "install [source]",
	Short: "Install a new plugin",
	Long: `Install a new plugin from a specified source (e.g., Git repository).

Example:
  axiomod plugin install github.com/example/myplugin
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pluginSource := args[0]
		fmt.Printf("Installing plugin from: %s\n", pluginSource)

		// Determine plugin name from source (simple example)
		pluginName := filepath.Base(pluginSource)

		// Define plugin directory
		pluginDir := filepath.Join("plugins", pluginName)

		// Check if plugin already exists
		if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
			fmt.Printf("Plugin %s already exists or directory check failed.\n", pluginName)
			// Optionally offer to update or reinstall
			return
		}

		// Create plugin directory
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			fmt.Printf("Error creating plugin directory %s: %v\n", pluginDir, err)
			os.Exit(1)
		}

		// Clone or download the plugin source code
		// Example using git clone
		fmt.Println("Cloning plugin source...")
		gitCmd := exec.Command("git", "clone", pluginSource, pluginDir)
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr

		if err := gitCmd.Run(); err != nil {
			fmt.Printf("Error cloning plugin source: %v\n", err)
			// Clean up potentially partially created directory
			os.RemoveAll(pluginDir)
			os.Exit(1)
		}

		fmt.Println("Please manually register the plugin in plugins/builtin_plugins.go")
		fmt.Println("Example registration:")
		fmt.Printf("import _ \"github.com/axiomod/axiomod/plugins/%s\"\n", pluginName)
		fmt.Println("// Add plugin options to the Fx application in cmd/axiomod-server/fx_options.go")

		fmt.Printf("\nPlugin %s installed successfully (manual registration required).\n", pluginName)
	},
}

// NewInstallCmd returns the plugin install command.
func NewInstallCmd() *cobra.Command {
	return installCmd
}

func init() {
	// Add subcommands to the parent pluginCmd
	pluginCmd.AddCommand(installCmd)
}
