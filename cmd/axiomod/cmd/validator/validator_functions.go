package validator

import (
	"fmt"
	"os"
)

// ValidateArchitecture validates the architecture of the codebase against the rules defined in the config file
// This function is called by the architecture command and standards-check command
func ValidateArchitecture(configPath string) error {
	// Get current working directory
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Run the validation
	passed, err := RunArchitectureValidation(rootDir, configPath)
	if err != nil {
		return err
	}

	if !passed {
		return fmt.Errorf("architecture validation failed")
	}

	return nil
}

// CheckAPISpec validates an API specification file against standards
// This function is called by the check-api-spec command
func CheckAPISpec(specPath string) error {
	// Check if the file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("API specification file not found: %s", specPath)
	}

	// In a real implementation, this would validate the OpenAPI/Swagger spec
	// using a library like go-swagger or openapi-validator
	fmt.Printf("Validating API specification: %s\n", specPath)
	fmt.Println("API specification validation is a placeholder in this version.")

	// For demonstration purposes, we'll just return success
	return nil
}

// ValidateNaming checks naming conventions across the codebase
// This function is called by the naming command and standards-check command
func ValidateNaming(fixIssues bool) ([]string, error) {
	// Get current working directory
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// In a real implementation, this would scan the codebase for naming convention violations
	fmt.Printf("Validating naming conventions in: %s\n", rootDir)

	// For demonstration purposes, we'll just return an empty list of issues
	return []string{}, nil
}

// ValidateDomainBoundaries checks that domain boundaries are respected
// This function is called by the domain command and standards-check command
func ValidateDomainBoundaries() ([]string, error) {
	// Get current working directory
	rootDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// In a real implementation, this would analyze imports to ensure domain boundaries are respected
	fmt.Printf("Validating domain boundaries in: %s\n", rootDir)

	// For demonstration purposes, we'll just return an empty list of issues
	return []string{}, nil
}
