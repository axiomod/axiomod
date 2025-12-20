package validator

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// checkAPISpecCmd represents the validator check-api-spec command
var checkAPISpecCmd = &cobra.Command{
	Use:   "check-api-spec",
	Short: "Check API spec against standards",
	Long: `Check the API specification (e.g., OpenAPI/Swagger) against defined standards and best practices.

Example:
  axiomod validator check-api-spec --spec=docs/api/openapi.yaml
`,
	Run: func(cmd *cobra.Command, args []string) {
		specPath, _ := cmd.Flags().GetString("spec")
		fmt.Println("Checking API specification...")

		if specPath == "" {
			fmt.Println("Error: --spec flag pointing to the API specification file is required.")
			os.Exit(1)
		}

		// Check if spec file exists
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			fmt.Printf("API specification file not found: %s\n", specPath)
			os.Exit(1)
		}

		fmt.Printf("Using API specification: %s\n", specPath)

		// Perform API spec validation
		err := CheckAPISpec(specPath)
		if err != nil {
			fmt.Printf("API specification validation failed:\n%v\n", err)
			os.Exit(1)
		}

		fmt.Println("API specification validation passed successfully.")
	},
}

// NewCheckAPISpecCmd returns the validator check-api-spec command.
func NewCheckAPISpecCmd() *cobra.Command {
	checkAPISpecCmd.Flags().StringP("spec", "s", "", "Path to the API specification file (e.g., OpenAPI yaml/json)")
	checkAPISpecCmd.MarkFlagRequired("spec")
	return checkAPISpecCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(checkAPISpecCmd)
}
