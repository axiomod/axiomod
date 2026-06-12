package core

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	testUnitOnly        bool
	testIntegrationOnly bool
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test [target]",
	Short: "Run tests for the project",
	Long: `Run tests for the project.

--unit restricts the run to ./tests/unit/...; --integration runs
./tests/integration/... with RUN_INTEGRATION_TESTS=true so gated tests
execute. Without flags, all packages (or the given target) are tested.

Example:
  axiomod test
  axiomod test ./examples/example/...
  axiomod test --unit
  axiomod test --integration
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running tests...")

		if testUnitOnly && testIntegrationOnly {
			fmt.Println("Error: --unit and --integration are mutually exclusive.")
			os.Exit(1)
		}

		// Default to running all tests if no specific package is provided
		target := "./..."
		if len(args) > 0 {
			target = args[0]
		}

		env := os.Environ()
		switch {
		case testUnitOnly:
			target = "./tests/unit/..."
		case testIntegrationOnly:
			target = "./tests/integration/..."
			env = append(env, "RUN_INTEGRATION_TESTS=true")
		}

		// -v provides verbose output, -cover shows test coverage
		goTestCmd := exec.Command("go", "test", "-v", "-cover", target)
		goTestCmd.Env = env
		goTestCmd.Stdout = os.Stdout
		goTestCmd.Stderr = os.Stderr

		fmt.Printf("Executing: %s\n", goTestCmd.String())
		if err := goTestCmd.Run(); err != nil {
			fmt.Printf("Tests failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nTests completed successfully.")
	},
}

// NewTestCmd returns the test command.
func NewTestCmd() *cobra.Command {
	testCmd.Flags().BoolVar(&testUnitOnly, "unit", false, "run only ./tests/unit/...")
	testCmd.Flags().BoolVar(&testIntegrationOnly, "integration", false,
		"run only ./tests/integration/... with RUN_INTEGRATION_TESTS=true")
	return testCmd
}
