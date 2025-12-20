package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start an interactive session with the Axiomod CLI",
	Long: `Start an interactive session where you can run Axiomod commands without prefixing them.

Example:
  axiomod interactive
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting interactive Axiomod session...")
		fmt.Println("Type 'exit' or 'quit' to leave.")

		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("axiomod> ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				continue
			}

			input = strings.TrimSpace(input)

			if input == "exit" || input == "quit" {
				fmt.Println("Exiting interactive session.")
				break
			}

			if input == "" {
				continue
			}

			// Split the input into command and arguments
			parts := strings.Fields(input)
			command := parts[0]
			commandArgs := parts[1:]

			// Find the command in Cobra
			subCmd, _, err := cmd.Root().Find(parts)
			if err != nil || subCmd == cmd { // Don't allow running 'interactive' recursively
				fmt.Printf("Unknown command: %s\n", command)
				continue
			}

			// Execute the command
			subCmd.SetArgs(commandArgs)
			if err := subCmd.Execute(); err != nil {
				fmt.Printf("Error executing command '%s': %v\n", input, err)
			}
			fmt.Println() // Add a newline for better separation
		}
	},
}

// NewInteractiveCmd returns the interactive command.
func NewInteractiveCmd() *cobra.Command {
	return interactiveCmd
}
