package validator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunStaticAnalysisValidation runs all static analysis tools (vet, gosec, staticcheck)
func RunStaticAnalysisValidation(dir string) (bool, error) {
	fmt.Println("Running static analysis tools...")

	// Run go vet
	vetSuccess, vetErr := RunGoVet(dir)

	// Run staticcheck
	staticCheckSuccess, staticCheckErr := RunStaticCheck(dir)

	// Run gosec
	securitySuccess, securityErr := RunSecurityCheck(dir)

	// If any of the checks failed, return failure
	if !vetSuccess || !staticCheckSuccess || !securitySuccess {
		return false, fmt.Errorf("static analysis failed: %v %v %v", vetErr, staticCheckErr, securityErr)
	}

	fmt.Println("\n✅ All static analysis checks passed!")
	return true, nil
}

// RunGoVet runs go vet on the specified directory
func RunGoVet(dir string) (bool, error) {
	fmt.Println("\n🔍 Running go vet...")

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ go vet found issues")
		return false, err
	}

	fmt.Println("✅ go vet passed")
	return true, nil
}

// RunStaticCheck runs staticcheck on the specified directory
func RunStaticCheck(dir string) (bool, error) {
	fmt.Println("\n🔍 Running staticcheck...")

	// Check if staticcheck is installed
	_, err := exec.LookPath("staticcheck")
	if err != nil {
		fmt.Println("staticcheck is not installed. Installing...")
		installCmd := exec.Command("go", "install", "honnef.co/go/tools/cmd/staticcheck@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		err = installCmd.Run()
		if err != nil {
			return false, fmt.Errorf("failed to install staticcheck: %v", err)
		}
	}

	cmd := exec.Command("staticcheck", "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("❌ staticcheck found issues")
		return false, err
	}

	fmt.Println("✅ staticcheck passed")
	return true, nil
}

// RunSecurityCheck runs gosec on the specified directory
func RunSecurityCheck(dir string) (bool, error) {
	fmt.Println("\n🔍 Running gosec security scanner...")

	// Check if gosec is installed
	_, err := exec.LookPath("gosec")
	if err != nil {
		fmt.Println("gosec is not installed. Installing...")
		installCmd := exec.Command("go", "install", "github.com/securego/gosec/v2/cmd/gosec@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		err = installCmd.Run()
		if err != nil {
			return false, fmt.Errorf("failed to install gosec: %v", err)
		}
	}

	cmd := exec.Command("gosec", "-quiet", "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("❌ gosec found security issues")
		return false, err
	}

	fmt.Println("✅ gosec security check passed")
	return true, nil
}

// RunAPISpecCheck checks API spec against standards
func RunAPISpecCheck(dir string) (bool, error) {
	fmt.Println("\n🔍 Checking API spec against standards...")

	// Look for OpenAPI spec files
	var specFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for OpenAPI spec files
		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".json") {
			// Check if file contains OpenAPI content
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			contentStr := string(content)
			if (filepath.Ext(path) == ".yaml" && (strings.Contains(contentStr, "openapi:") ||
				strings.Contains(contentStr, "swagger:"))) ||
				(filepath.Ext(path) == ".json" && (strings.Contains(contentStr, "\"openapi\"") ||
					strings.Contains(contentStr, "\"swagger\""))) {
				specFiles = append(specFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return false, fmt.Errorf("error searching for API spec files: %v", err)
	}

	if len(specFiles) == 0 {
		fmt.Println("⚠️ No OpenAPI spec files found")
		return true, nil
	}

	// Check if spectral is installed
	_, err = exec.LookPath("spectral")
	if err != nil {
		fmt.Println("spectral is not installed. Skipping API spec validation.")
		fmt.Println("To install spectral: npm install -g @stoplight/spectral-cli")
		return true, nil
	}

	// Validate each spec file
	allPassed := true
	for _, specFile := range specFiles {
		fmt.Printf("Validating %s...\n", specFile)
		cmd := exec.Command("spectral", "lint", specFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			fmt.Printf("❌ API spec validation failed for %s\n", specFile)
			allPassed = false
		} else {
			fmt.Printf("✅ API spec validation passed for %s\n", specFile)
		}
	}

	if !allPassed {
		return false, fmt.Errorf("API spec validation failed")
	}

	return true, nil
}

// RunDocsCheck checks if code changes have documentation updates
func RunDocsCheck(dir string) (bool, error) {
	fmt.Println("\n🔍 Checking if code changes have documentation updates...")

	// Check if git is available
	_, err := exec.LookPath("git")
	if err != nil {
		fmt.Println("⚠️ git is not available, skipping documentation check")
		return true, nil
	}

	// Get list of changed files in the last commit
	cmd := exec.Command("git", "diff", "--name-only", "HEAD~1", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("⚠️ Could not get changed files, skipping documentation check")
		return true, nil
	}

	changedFiles := strings.Split(string(output), "\n")

	// Check if any Go files were changed
	codeChanged := false
	for _, file := range changedFiles {
		if filepath.Ext(file) == ".go" && !strings.Contains(file, "_test.go") {
			codeChanged = true
			break
		}
	}

	if !codeChanged {
		fmt.Println("✅ No code changes detected, documentation check not needed")
		return true, nil
	}

	// Check if any documentation files were changed
	docsChanged := false
	for _, file := range changedFiles {
		if filepath.Ext(file) == ".md" ||
			strings.Contains(file, "docs/") ||
			strings.Contains(file, "documentation/") {
			docsChanged = true
			break
		}
	}

	if !docsChanged {
		fmt.Println("❌ Code changes detected but no documentation updates found")
		fmt.Println("Please update relevant documentation for your code changes")
		return false, fmt.Errorf("documentation updates missing")
	}

	fmt.Println("✅ Documentation updates found for code changes")
	return true, nil
}

// RunAllValidators runs all validators
func RunAllValidators(dir string, configPath string, sqlPath string, apiPath string) (bool, error) {
	fmt.Println("Running all validators...")

	// Track overall success
	allPassed := true
	var failedChecks []string

	// Run architecture validation
	fmt.Println("\n=== Architecture Validation ===")
	archSuccess, _ := RunArchitectureValidation(dir, configPath)
	if !archSuccess {
		allPassed = false
		failedChecks = append(failedChecks, "architecture")
	}

	// Run naming validation
	fmt.Println("\n=== Naming Convention Validation ===")
	namingSuccess, _ := RunNamingValidation(dir, sqlPath, apiPath, false)
	if !namingSuccess {
		allPassed = false
		failedChecks = append(failedChecks, "naming")
	}

	// Run domain validation
	fmt.Println("\n=== Domain Boundary Validation ===")
	domainSuccess, _ := RunDomainValidation(filepath.Join(dir, "internal"), configPath)
	if !domainSuccess {
		allPassed = false
		failedChecks = append(failedChecks, "domain")
	}

	// Run static analysis
	fmt.Println("\n=== Static Analysis ===")
	staticSuccess, _ := RunStaticAnalysisValidation(dir)
	if !staticSuccess {
		allPassed = false
		failedChecks = append(failedChecks, "static-analysis")
	}

	// Run API spec check
	fmt.Println("\n=== API Spec Validation ===")
	apiSpecSuccess, _ := RunAPISpecCheck(dir)
	if !apiSpecSuccess {
		allPassed = false
		failedChecks = append(failedChecks, "api-spec")
	}

	// Run docs check
	fmt.Println("\n=== Documentation Check ===")
	docsSuccess, _ := RunDocsCheck(dir)
	if !docsSuccess {
		allPassed = false
		failedChecks = append(failedChecks, "docs")
	}

	// Print summary
	fmt.Println("\n=== Validation Summary ===")
	if allPassed {
		fmt.Println("✅ All validation checks passed!")
		return true, nil
	} else {
		fmt.Println("❌ Some validation checks failed:")
		for _, check := range failedChecks {
			fmt.Printf("  - %s\n", check)
		}
		return false, fmt.Errorf("validation failed: %s", strings.Join(failedChecks, ", "))
	}
}
