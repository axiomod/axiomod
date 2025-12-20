package core

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format Go source code",
	Long: `Run gofmt on the project source code to ensure consistent formatting.

Example:
  axiomod fmt
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Formatting Go source code...")

		// Prepare the gofmt command
		// -w flag writes the result back to the source file
		// ./... applies the command recursively to all subdirectories
		gofmtCmd := exec.Command("gofmt", "-w", ".")

		// Set the output to the standard output and error streams
		gofmtCmd.Stdout = os.Stdout
		gofmtCmd.Stderr = os.Stderr

		// Run the command
		fmt.Printf("Executing: %s\n", gofmtCmd.String())
		err := gofmtCmd.Run()

		if err != nil {
			fmt.Printf("Code formatting failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nGo source code formatted successfully.")
	},
}

// NewFmtCmd returns the fmt command.
func NewFmtCmd() *cobra.Command {
	return fmtCmd
}
