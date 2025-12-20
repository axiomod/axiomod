package main

import (
	"fmt"
	"os"
	"os/exec"
)

// This script runs all tests in the project
func main() {
	fmt.Println("Running tests for Enterprise Go Macroservice Framework...")

	// Run unit tests
	fmt.Println("\nRunning unit tests...")
	unitTestCmd := exec.Command("go", "test", "./...")
	unitTestCmd.Stdout = os.Stdout
	unitTestCmd.Stderr = os.Stderr
	if err := unitTestCmd.Run(); err != nil {
		fmt.Printf("Unit tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAll tests passed successfully!")
	fmt.Println("\nEnterprise Go Macroservice Framework is ready for use.")
}
