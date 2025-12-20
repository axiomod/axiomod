package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Go Macroservice application",
	Long: `Build the main Go Macroservice application binary.

Example:
  axiomod build
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building the application...")

		// Determine the main package path
		// Assuming the main service is named 'axiomod-server'
		mainPackage := "./cmd/axiomod-server"

		// Determine the output path
		outputPath := "bin/axiomod-server"

		// Create the output directory if it doesn't exist
		outputDir := filepath.Dir(outputPath)
		_ = os.MkdirAll(outputDir, 0755)

		// Prepare the go build command
		// -o flag specifies the output file path
		// -v flag provides verbose output
		goCmd := exec.Command("go", "build", "-v", "-o", outputPath, mainPackage)

		// Set the output to the standard output and error streams
		goCmd.Stdout = os.Stdout
		goCmd.Stderr = os.Stderr

		// Run the command
		fmt.Printf("Executing: %s\n", goCmd.String())

		startTime := time.Now()

		err := goCmd.Run()

		duration := time.Since(startTime)
		fmt.Printf("\nBuild finished in %s\n", duration)

		if err != nil {
			fmt.Printf("Build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Application built successfully: %s\n", outputPath)
	},
}

// NewBuildCmd returns the build command.
func NewBuildCmd() *cobra.Command {
	// Add flags or initialization specific to buildCmd here if needed in the future
	return buildCmd
}

// Removed the init() function that tried to access rootCmd
// func init() {
// 	rootCmd.AddCommand(buildCmd)
// }
