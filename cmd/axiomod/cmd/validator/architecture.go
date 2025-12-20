package validator

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// architectureCmd represents the validator architecture command
var architectureCmd = &cobra.Command{
	Use:   "architecture",
	Short: "Validate project architecture against defined rules",
	Long: `Validate the project architecture by checking dependencies against rules defined in architecture-rules.json.

Example:
  axiomod validator architecture
  axiomod validator architecture --config=path/to/rules.json
`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")
		fmt.Println("Validating project architecture...")

		// Default config path
		if configPath == "" {
			configPath = "architecture-rules.json"
		}

		// Check if config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("Architecture rules file not found: %s\n", configPath)
			fmt.Println("Please create an architecture-rules.json file or specify a path using --config.")
			os.Exit(1)
		}

		fmt.Printf("Using rules file: %s\n", configPath)

		// Load rules and perform validation
		err := ValidateArchitecture(configPath)
		if err != nil {
			fmt.Printf("Architecture validation failed:\n%v\n", err)
			os.Exit(1)
		}

		fmt.Println("Architecture validation passed successfully.")
	},
}

// NewArchitectureCmd returns the validator architecture command.
func NewArchitectureCmd() *cobra.Command {
	architectureCmd.Flags().StringP("config", "c", "", "Path to the architecture rules JSON file")
	return architectureCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(architectureCmd)
}
