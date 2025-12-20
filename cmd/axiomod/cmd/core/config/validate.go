package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configValidateCmd represents the config validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration files",
	Long: `Validate configuration files for your project.

Example:
  axiomod config validate
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Validating configuration files...")

		// Check if config directory exists
		configDir := "internal/framework/config"
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			fmt.Println("No config directory found.")
			return
		}

		// Get config files
		files, err := filepath.Glob(filepath.Join(configDir, "*.yaml"))
		if err != nil {
			fmt.Printf("Error finding config files: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No config files found")
			return
		}

		// Validate each config file
		validFiles := 0
		invalidFiles := 0

		for _, file := range files {
			fmt.Printf("Validating config file: %s\n", filepath.Base(file))

			// Create a new viper instance for each file
			v := viper.New()
			v.SetConfigFile(file)

			// Try to read the config file
			if err := v.ReadInConfig(); err != nil {
				fmt.Printf("Error reading config file %s: %v\n", filepath.Base(file), err)
				invalidFiles++
				continue
			}

			// If we get here, the file is valid YAML
			fmt.Printf("Config file %s is valid\n", filepath.Base(file))
			validFiles++
		}

		fmt.Printf("\nValidation complete: %d valid files, %d invalid files\n", validFiles, invalidFiles)

		if invalidFiles > 0 {
			os.Exit(1)
		}
	},
}

// NewConfigValidateCmd returns the config validate command
func NewConfigValidateCmd() *cobra.Command {
	return configValidateCmd
}
