package core

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests for the project",
	Long: `Run unit and integration tests for the Go Macroservice project.

Example:
  axiomod test
  axiomod test ./examples/example/...
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running tests...")

		// Default to running all tests if no specific package is provided
		target := "./..."
		if len(args) > 0 {
			target = args[0]
		}

		// Prepare the go test command
		// -v flag provides verbose output
		// -cover flag shows test coverage
		goTestCmd := exec.Command("go", "test", "-v", "-cover", target)

		// Set the output to the standard output and error streams
		goTestCmd.Stdout = os.Stdout
		goTestCmd.Stderr = os.Stderr

		// Run the command
		fmt.Printf("Executing: %s\n", goTestCmd.String())
		err := goTestCmd.Run()

		if err != nil {
			fmt.Printf("Tests failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nTests completed successfully.")
	},
}

// NewTestCmd returns the test command.
func NewTestCmd() *cobra.Command {
	return testCmd
}
