package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configDiffCmd represents the config diff command
var configDiffCmd = &cobra.Command{
	Use:   "diff [env1] [env2]",
	Short: "Compare configuration between environments",
	Long: `Compare configuration between two environments.

Example:
  axiomod config diff dev prod
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		env1 := args[0]
		env2 := args[1]

		fmt.Printf("Comparing configuration between %s and %s...\n", env1, env2)

		// Check if config directory exists
		configDir := "framework/config"
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			fmt.Println("No config directory found.")
			return
		}

		// Look for environment-specific config files
		env1File := filepath.Join(configDir, fmt.Sprintf("config.%s.yaml", env1))
		env2File := filepath.Join(configDir, fmt.Sprintf("config.%s.yaml", env2))

		if _, err := os.Stat(env1File); os.IsNotExist(err) {
			fmt.Printf("Config file for environment %s not found: %s\n", env1, env1File)
			return
		}

		if _, err := os.Stat(env2File); os.IsNotExist(err) {
			fmt.Printf("Config file for environment %s not found: %s\n", env2, env2File)
			return
		}

		// Load config for env1
		v1 := viper.New()
		v1.SetConfigFile(env1File)
		if err := v1.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config file %s: %v\n", env1File, err)
			os.Exit(1)
		}

		// Load config for env2
		v2 := viper.New()
		v2.SetConfigFile(env2File)
		if err := v2.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config file %s: %v\n", env2File, err)
			os.Exit(1)
		}

		// Get all settings
		settings1 := v1.AllSettings()
		settings2 := v2.AllSettings()

		// Compare settings
		fmt.Printf("\nDifferences between %s and %s:\n\n", env1, env2)

		// Get all keys from both configs
		keys := make(map[string]bool)
		for k := range flattenMap(settings1, "") {
			keys[k] = true
		}
		for k := range flattenMap(settings2, "") {
			keys[k] = true
		}

		// Sort keys for consistent output
		sortedKeys := make([]string, 0, len(keys))
		for k := range keys {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// Compare values for each key
		diffCount := 0
		for _, k := range sortedKeys {
			val1 := v1.Get(k)
			val2 := v2.Get(k)

			if !reflect.DeepEqual(val1, val2) {
				fmt.Printf("Key: %s\n", k)
				fmt.Printf("  %s: %v\n", env1, val1)
				fmt.Printf("  %s: %v\n\n", env2, val2)
				diffCount++
			}
		}

		if diffCount == 0 {
			fmt.Printf("No differences found between %s and %s configurations.\n", env1, env2)
		} else {
			fmt.Printf("Found %d differences between %s and %s configurations.\n", diffCount, env1, env2)
		}
	},
}

// NewConfigDiffCmd returns the config diff command
func NewConfigDiffCmd() *cobra.Command {
	return configDiffCmd
}

// flattenMap converts a nested map to a flat map with dot-separated keys
func flattenMap(m map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		if nested, ok := v.(map[string]interface{}); ok {
			// If value is a nested map, recursively flatten it
			for nestedK, nestedV := range flattenMap(nested, key) {
				result[nestedK] = nestedV
			}
		} else {
			// Otherwise, add the key-value pair to the result
			result[key] = v
		}
	}

	return result
}
