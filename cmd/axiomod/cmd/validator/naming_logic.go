package validator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidationResult stores the result of a validation check
type ValidationResult struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Expected    string `json:"expected"`
	Description string `json:"description"`
}

// ValidationResults holds all errors and warnings
type ValidationResults struct {
	Errors   []ValidationResult `json:"errors"`
	Warnings []ValidationResult `json:"warnings"`
}

// NamingValidationSummary stores statistical information about the validation
type NamingValidationSummary struct {
	// Files and components checked
	FilesChecked       int
	PackagesChecked    int
	FunctionsChecked   int
	VariablesChecked   int
	TypesChecked       int
	StructsChecked     int
	EndpointsChecked   int
	TablesChecked      int
	ColumnsChecked     int
	SchemaTypesChecked int

	// Violation statistics
	TotalErrors    int
	TotalWarnings  int
	ErrorsByType   map[string]int
	WarningsByType map[string]int
	ErrorsByFile   map[string]int
}

// NamingValidator holds validation state
type NamingValidator struct {
	results ValidationResults
}

// NewNamingValidator creates a new validator instance
func NewNamingValidator() *NamingValidator {
	return &NamingValidator{
		results: ValidationResults{
			Errors:   make([]ValidationResult, 0),
			Warnings: make([]ValidationResult, 0),
		},
	}
}

// AddError adds an error to the validation results
func (v *NamingValidator) AddError(result *ValidationResult) {
	v.results.Errors = append(v.results.Errors, *result)
}

// AddWarning adds a warning to the validation results
func (v *NamingValidator) AddWarning(result *ValidationResult) {
	v.results.Warnings = append(v.results.Warnings, *result)
}

// HasErrors returns true if there are errors
func (v *NamingValidator) HasErrors() bool {
	return len(v.results.Errors) > 0
}

// GetErrorCount returns the number of errors
func (v *NamingValidator) GetErrorCount() int {
	return len(v.results.Errors)
}

// GetWarningCount returns the number of warnings
func (v *NamingValidator) GetWarningCount() int {
	return len(v.results.Warnings)
}

// RunNamingValidation validates naming conventions in the codebase
func RunNamingValidation(dirPath string, sqlPath string, apiPath string, jsonOutput bool) (bool, error) {
	fmt.Println("⏳ Running naming convention checks...")

	// Create validator instance
	validator := NewNamingValidator()

	// Initialize validation summary
	summary := NamingValidationSummary{
		ErrorsByType:   make(map[string]int),
		WarningsByType: make(map[string]int),
		ErrorsByFile:   make(map[string]int),
	}

	// Resolve relative paths to absolute paths for consistency
	rootDir, err := filepath.Abs(dirPath)
	if err != nil {
		return false, fmt.Errorf("error resolving path: %v", err)
	}

	// Validate Go code (files, packages, variables, functions)
	validateGoCode(rootDir, validator, &summary)

	// Validate API endpoint naming in handlers
	apiDir := apiPath
	if !filepath.IsAbs(apiDir) {
		apiDir = filepath.Join(rootDir, apiDir)
	}
	validateAPIEndpoints(apiDir, validator, &summary)

	// Validate database naming (tables, columns, etc.)
	sqlDir := sqlPath
	if !filepath.IsAbs(sqlDir) {
		sqlDir = filepath.Join(rootDir, sqlDir)
	}
	validateDatabaseNaming(sqlDir, validator, &summary)

	// Validate Ent schemas
	entDir := filepath.Join(rootDir, filepath.FromSlash("platform/ent/schema"))
	validateEntSchemas(entDir, validator, &summary)

	// Update summary counters
	summary.TotalErrors = validator.GetErrorCount()
	summary.TotalWarnings = validator.GetWarningCount()

	// Count errors by type and file
	for _, err := range validator.results.Errors {
		summary.ErrorsByType[err.Type]++
		summary.ErrorsByFile[filepath.Base(err.File)]++
	}

	// Count warnings by type
	for _, warning := range validator.results.Warnings {
		summary.WarningsByType[warning.Type]++
	}

	// Print summary report first
	if !jsonOutput {
		fmt.Println("\nNaming Convention Validation Summary:")
		printNamingSummaryReport(&summary)
	}

	// Report results
	if jsonOutput {
		outputJSON(validator)
	} else {
		outputConsole(validator)
	}

	// Print summary report again after violations
	if summary.TotalErrors > 0 && !jsonOutput {
		fmt.Println("\nNaming Convention Validation Summary:")
		printNamingSummaryReport(&summary)
	}

	// Return success if there are no errors
	if validator.HasErrors() {
		return false, fmt.Errorf("naming convention validation failed with %d errors", validator.GetErrorCount())
	}

	return true, nil
}

// printNamingSummaryReport prints a summary of the validation results
func printNamingSummaryReport(summary *NamingValidationSummary) {
	fmt.Printf("  Files checked: %d\n", summary.FilesChecked)

	if summary.PackagesChecked > 0 {
		fmt.Printf("  Packages checked: %d\n", summary.PackagesChecked)
	}
	if summary.FunctionsChecked > 0 {
		fmt.Printf("  Functions checked: %d\n", summary.FunctionsChecked)
	}
	if summary.TypesChecked > 0 {
		fmt.Printf("  Types checked: %d\n", summary.TypesChecked)
	}
	if summary.StructsChecked > 0 {
		fmt.Printf("  Structs checked: %d\n", summary.StructsChecked)
	}
	if summary.VariablesChecked > 0 {
		fmt.Printf("  Variables checked: %d\n", summary.VariablesChecked)
	}
	if summary.TablesChecked > 0 {
		fmt.Printf("  Database tables checked: %d\n", summary.TablesChecked)
	}
	if summary.ColumnsChecked > 0 {
		fmt.Printf("  Database columns checked: %d\n", summary.ColumnsChecked)
	}
	if summary.EndpointsChecked > 0 {
		fmt.Printf("  API endpoints checked: %d\n", summary.EndpointsChecked)
	}
	if summary.SchemaTypesChecked > 0 {
		fmt.Printf("  Schema types checked: %d\n", summary.SchemaTypesChecked)
	}

	fmt.Printf("  Total errors: %d\n", summary.TotalErrors)
	fmt.Printf("  Total warnings: %d\n", summary.TotalWarnings)

	if summary.TotalErrors > 0 {
		fmt.Println("\nErrors by type:")
		printTopNamingViolations(summary.ErrorsByType)

		fmt.Println("\nTop files with naming violations:")
		printTopNamingViolations(summary.ErrorsByFile)
	}

	if summary.TotalWarnings > 0 {
		fmt.Println("\nWarnings by type:")
		printTopNamingViolations(summary.WarningsByType)
	}
}

// printTopNamingViolations prints violations in descending order
func printTopNamingViolations(violations map[string]int) {
	// Convert map to slice for sorting
	type violation struct {
		item  string
		count int
	}

	var items []violation
	for item, count := range violations {
		items = append(items, violation{item, count})
	}

	// Sort by count (descending)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].count < items[j].count {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Print top 5 items (or all if less than 5)
	count := len(items)
	if count > 5 {
		count = 5
	}

	for i := 0; i < count; i++ {
		fmt.Printf("  - %s: %d\n", items[i].item, items[i].count)
	}
}

func outputConsole(validator *NamingValidator) {
	// Print warnings
	if validator.GetWarningCount() > 0 {
		fmt.Println("\n⚠️  Warnings:")
		for _, warning := range validator.results.Warnings {
			fmt.Printf("  • %s:%d:%d: %s '%s' should be %s: %s\n",
				warning.File, warning.Line, warning.Column,
				warning.Type, warning.Name, warning.Expected, warning.Description)
		}
	}

	// Print errors
	if validator.HasErrors() {
		fmt.Println("\n❌ Naming convention errors:")
		for _, err := range validator.results.Errors {
			fmt.Printf("  • %s:%d:%d: %s '%s' should be %s: %s\n",
				err.File, err.Line, err.Column,
				err.Type, err.Name, err.Expected, err.Description)
		}
		fmt.Printf("\nFound %d naming convention errors.\n", validator.GetErrorCount())
		fmt.Println("Please fix these issues to comply with the project naming standards.")
	} else {
		fmt.Println("\n✅ All naming conventions checks passed!")
	}

	if validator.GetWarningCount() > 0 {
		fmt.Printf("Found %d warnings (these won't fail the build).\n", validator.GetWarningCount())
	}
}

func outputJSON(validator *NamingValidator) {
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(validator.results, "", "  ")
	if err != nil {
		fmt.Printf("Error generating JSON output: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}

func validateGoCode(dirPath string, validator *NamingValidator, summary *NamingValidationSummary) {
	fmt.Println("Checking Go code naming conventions...")

	// Create a new file set
	fset := token.NewFileSet()

	// Walk through directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and generated code
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/ent/") && !strings.Contains(path, "/ent/schema/") {
			return nil
		}

		// Check only Go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			summary.FilesChecked++
			validateGoFile(fset, path, validator, summary)
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}
}

func validateGoFile(fset *token.FileSet, path string, validator *NamingValidator, summary *NamingValidationSummary) {
	// Parse the Go file
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file %s: %v\n", path, err)
		return
	}

	// Validate package name
	summary.PackagesChecked++
	validatePackageName(fset, node, path, validator)

	// Check for exported functions without proper PascalCase
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			summary.FunctionsChecked++
			validateFunctionName(fset, fn, path, validator)
		}
	}

	// Validate variable declarations
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.VAR || x.Tok == token.CONST {
				for _, spec := range x.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							summary.VariablesChecked++
							validateVariableName(fset, name, path, validator)
						}
					}
				}
			}
		case *ast.TypeSpec:
			// Check struct and interface names
			summary.TypesChecked++
			validateTypeName(fset, x, path, validator)

			// Check struct field names if this is a struct
			if st, ok := x.Type.(*ast.StructType); ok {
				summary.StructsChecked++
				for _, field := range st.Fields.List {
					for _, name := range field.Names {
						validateStructFieldName(fset, name, x.Name.String(), path, validator)
					}
				}
			}
		}
		return true
	})

	// Check file name
	validateFileName(path, validator)
}

func validatePackageName(fset *token.FileSet, node *ast.File, path string, validator *NamingValidator) {
	name := node.Name.String()

	// Exception for versioned API packages (like v1, v2) under internal/handler/
	if regexp.MustCompile(`^v\d+$`).MatchString(name) && strings.Contains(path, "/internal/delivery/") {
		return // Skip validation for API version packages
	}

	// Package names should be lowercase, single words
	if !regexp.MustCompile(`^[a-z]+$`).MatchString(name) {
		pos := fset.Position(node.Name.Pos())
		validator.AddError(&ValidationResult{
			File:        path,
			Line:        pos.Line,
			Column:      pos.Column,
			Type:        "Package",
			Name:        name,
			Expected:    "lowercase, single word",
			Description: "Package names should be lowercase, single words without underscores",
		})
	}

	// Allow specific plural package names that are conventional in this codebase
	allowedPluralPackages := map[string]bool{
		"errors":   true,
		"events":   true,
		"adapters": true,
		"mocks":    true,
		"metrics":  true,
		"stats":    true,
		"status":   true,
	}

	// Package should be singular, not plural (with exceptions)
	if strings.HasSuffix(name, "s") && !allowedPluralPackages[name] {
		pos := fset.Position(node.Name.Pos())
		validator.AddWarning(&ValidationResult{
			File:        path,
			Line:        pos.Line,
			Column:      pos.Column,
			Type:        "Package",
			Name:        name,
			Expected:    "singular form",
			Description: "Package names should use singular form, not plural",
		})
	}
}

func validateFunctionName(fset *token.FileSet, fn *ast.FuncDecl, path string, validator *NamingValidator) {
	name := fn.Name.String()

	// Special handling for test functions
	if strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark") || strings.HasPrefix(name, "Example") {
		// Test, Benchmark, and Example functions often follow patterns like Test[Type]_[Method]
		// This is a common and acceptable Go convention
		isValidTestName := regexp.MustCompile(`^(Test|Benchmark|Example)([A-Z][a-zA-Z0-9]*)?(_[A-Za-z0-9]+)*$`).MatchString(name)
		if !isValidTestName {
			pos := fset.Position(fn.Name.Pos())
			validator.AddWarning(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Test function",
				Name:        name,
				Expected:    "Test[Type]_[Method] or similar pattern",
				Description: "Test functions should follow patterns like Test[Type]_[Method]",
			})
		}
		return // Skip other checks for test functions
	}

	// For exported functions (starting with uppercase)
	if ast.IsExported(name) {
		// Exported functions should follow PascalCase
		if !regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$").MatchString(name) {
			pos := fset.Position(fn.Name.Pos())
			validator.AddError(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Exported function",
				Name:        name,
				Expected:    "PascalCase",
				Description: "Exported functions should use PascalCase",
			})
		}
	} else {
		// Unexported functions should follow camelCase
		if !regexp.MustCompile("^[a-z][a-zA-Z0-9]*$").MatchString(name) {
			pos := fset.Position(fn.Name.Pos())
			validator.AddError(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Unexported function",
				Name:        name,
				Expected:    "camelCase",
				Description: "Unexported functions should use camelCase",
			})
		}
	}
}

func validateVariableName(fset *token.FileSet, name *ast.Ident, path string, validator *NamingValidator) {
	varName := name.String()

	// For exported variables (starting with uppercase)
	if ast.IsExported(varName) {
		// Exported variables should follow PascalCase
		if !regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$").MatchString(varName) {
			pos := fset.Position(name.Pos())
			validator.AddError(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Exported variable",
				Name:        varName,
				Expected:    "PascalCase",
				Description: "Exported variables should use PascalCase",
			})
		}
	} else {
		// Unexported variables should follow camelCase
		if !regexp.MustCompile("^[a-z][a-zA-Z0-9]*$").MatchString(varName) && varName != "_" {
			pos := fset.Position(name.Pos())
			validator.AddError(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Unexported variable",
				Name:        varName,
				Expected:    "camelCase",
				Description: "Unexported variables should use camelCase",
			})
		}
	}

	// Check boolean variable naming convention, but exclude variables starting with Err or err
	// as they're typically error variables, not booleans
	if strings.HasPrefix(varName, "Err") || strings.HasPrefix(varName, "err") {
		return // Skip boolean naming checks for error variables
	}

	if !strings.Contains(path, "_test.go") {
		shouldBeBooleanPrefix := strings.HasPrefix(varName, "is") ||
			strings.HasPrefix(varName, "has") ||
			strings.HasPrefix(varName, "should") ||
			strings.HasPrefix(varName, "can") ||
			strings.HasPrefix(varName, "must")

		shouldBeBooleanSuffix := strings.HasSuffix(varName, "Enabled") ||
			strings.HasSuffix(varName, "Disabled") ||
			strings.HasSuffix(varName, "Active") ||
			strings.HasSuffix(varName, "Valid")

		if shouldBeBooleanPrefix || shouldBeBooleanSuffix {
			// This is just a warning, as we can't be sure without type information
			pos := fset.Position(name.Pos())
			validator.AddWarning(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Boolean variable",
				Name:        varName,
				Expected:    "boolean type",
				Description: "Variable name suggests it's a boolean, verify the type is correct",
			})
		}
	}
}

func validateTypeName(fset *token.FileSet, typeSpec *ast.TypeSpec, path string, validator *NamingValidator) {
	typeName := typeSpec.Name.String()

	// All types should be PascalCase
	if !regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$").MatchString(typeName) {
		pos := fset.Position(typeSpec.Name.Pos())
		validator.AddError(&ValidationResult{
			File:        path,
			Line:        pos.Line,
			Column:      pos.Column,
			Type:        "Type",
			Name:        typeName,
			Expected:    "PascalCase",
			Description: "Type names should use PascalCase",
		})
	}

	// Check for interface names that don't follow the -er convention
	if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
		// Interfaces that define a single method often use the -er suffix
		// This is just a guideline, not a strict rule, so we'll make it a warning
		if !strings.HasSuffix(typeName, "er") && !strings.HasSuffix(typeName, "Service") &&
			!strings.HasSuffix(typeName, "Repository") && !strings.HasSuffix(typeName, "Factory") &&
			!strings.HasSuffix(typeName, "Provider") && !strings.HasSuffix(typeName, "Manager") {
			pos := fset.Position(typeSpec.Name.Pos())
			validator.AddWarning(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Interface",
				Name:        typeName,
				Expected:    "name with -er suffix or standard pattern",
				Description: "Consider using -er suffix for interfaces that define a single behavior",
			})
		}
	}
}

func validateStructFieldName(fset *token.FileSet, name *ast.Ident, structName string, path string, validator *NamingValidator) {
	fieldName := name.String()

	// For exported fields (starting with uppercase)
	if ast.IsExported(fieldName) {
		// Exported fields should follow PascalCase
		if !regexp.MustCompile("^[A-Z][a-zA-Z0-9]*$").MatchString(fieldName) {
			pos := fset.Position(name.Pos())
			validator.AddError(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Exported struct field",
				Name:        fieldName,
				Expected:    "PascalCase",
				Description: fmt.Sprintf("Exported fields in struct %s should use PascalCase", structName),
			})
		}
	} else {
		// Unexported fields should follow camelCase
		if !regexp.MustCompile("^[a-z][a-zA-Z0-9]*$").MatchString(fieldName) {
			pos := fset.Position(name.Pos())
			validator.AddError(&ValidationResult{
				File:        path,
				Line:        pos.Line,
				Column:      pos.Column,
				Type:        "Unexported struct field",
				Name:        fieldName,
				Expected:    "camelCase",
				Description: fmt.Sprintf("Unexported fields in struct %s should use camelCase", structName),
			})
		}
	}
}

func validateFileName(path string, validator *NamingValidator) {
	// Get just the filename without directory
	filename := filepath.Base(path)

	// Skip test files
	if strings.HasSuffix(filename, "_test.go") {
		return
	}

	// File names should be snake_case.go
	if !regexp.MustCompile(`^[a-z][a-z0-9_]*\.go$`).MatchString(filename) {
		validator.AddError(&ValidationResult{
			File:        path,
			Line:        1,
			Column:      1,
			Type:        "File name",
			Name:        filename,
			Expected:    "snake_case.go",
			Description: "File names should use snake_case",
		})
	}
}

func validateAPIEndpoints(apiDir string, validator *NamingValidator, summary *NamingValidationSummary) {
	fmt.Println("Checking API endpoint naming conventions...")

	// Walk through the API handlers directory
	err := filepath.Walk(apiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only check Go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			// Parse the file to find route definitions
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				return nil
			}

			// Look for route definitions in the file
			scanner := bufio.NewScanner(strings.NewReader(string(content)))
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()

				// Look for route definitions like app.Get("/api/v1/users", ...)
				// or router.Post("/api/v1/users", ...)
				if routeRegex := regexp.MustCompile(`\.(Get|Post|Put|Delete|Patch)\s*\(\s*"([^"]+)"`); routeRegex.MatchString(line) {
					matches := routeRegex.FindStringSubmatch(line)
					if len(matches) >= 3 {
						method := matches[1]
						endpoint := matches[2]
						summary.EndpointsChecked++

						// Validate the endpoint
						validateAPIEndpoint(path, lineNum, method, endpoint, validator)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking API directory: %v\n", err)
	}
}

func validateAPIEndpoint(file string, line int, method string, endpoint string, validator *NamingValidator) {
	// API endpoints should follow RESTful conventions
	// They should be lowercase with hyphens as separators
	if !regexp.MustCompile(`^/[a-z0-9/-]+$`).MatchString(endpoint) {
		validator.AddError(&ValidationResult{
			File:        file,
			Line:        line,
			Column:      1,
			Type:        "API endpoint",
			Name:        endpoint,
			Expected:    "lowercase with hyphens",
			Description: "API endpoints should be lowercase with hyphens as separators",
		})
	}

	// Check for trailing slashes (except for root endpoint)
	if endpoint != "/" && strings.HasSuffix(endpoint, "/") {
		validator.AddWarning(&ValidationResult{
			File:        file,
			Line:        line,
			Column:      1,
			Type:        "API endpoint",
			Name:        endpoint,
			Expected:    "no trailing slash",
			Description: "API endpoints should not have trailing slashes",
		})
	}

	// Check for proper versioning format
	if strings.Contains(endpoint, "/v") {
		versionRegex := regexp.MustCompile(`/v\d+/`)
		if !versionRegex.MatchString(endpoint) {
			validator.AddWarning(&ValidationResult{
				File:        file,
				Line:        line,
				Column:      1,
				Type:        "API version",
				Name:        endpoint,
				Expected:    "/v{number}/",
				Description: "API version should follow the format /v{number}/",
			})
		}
	}

	// Check for proper resource naming in RESTful endpoints
	parts := strings.Split(strings.Trim(endpoint, "/"), "/")
	for i, part := range parts {
		// Skip version parts like "v1"
		if regexp.MustCompile(`^v\d+$`).MatchString(part) {
			continue
		}

		// Resource names should be plural for collections
		if i == len(parts)-1 || (i < len(parts)-1 && !regexp.MustCompile(`^\d+$`).MatchString(parts[i+1])) {
			if !strings.HasSuffix(part, "s") && part != "health" && part != "status" && part != "login" && part != "logout" {
				validator.AddWarning(&ValidationResult{
					File:        file,
					Line:        line,
					Column:      1,
					Type:        "API resource",
					Name:        part,
					Expected:    "plural form",
					Description: "RESTful resource names should use plural form for collections",
				})
			}
		}
	}
}

func validateDatabaseNaming(sqlDir string, validator *NamingValidator, summary *NamingValidationSummary) {
	fmt.Println("Checking database naming conventions...")

	// Check if the SQL directory exists
	if _, err := os.Stat(sqlDir); os.IsNotExist(err) {
		fmt.Printf("SQL directory %s does not exist, skipping database naming checks\n", sqlDir)
		return
	}

	// Walk through the SQL migration files
	err := filepath.Walk(sqlDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only check SQL files
		if !info.IsDir() && (strings.HasSuffix(path, ".sql") || strings.HasSuffix(path, ".up.sql")) {
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				return nil
			}

			// Parse SQL file to find table and column definitions
			scanner := bufio.NewScanner(strings.NewReader(string(content)))
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()

				// Look for CREATE TABLE statements
				tableRegex := regexp.MustCompile("CREATE\\s+TABLE\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?[\"']?([a-zA-Z0-9_]+)[\"']?")
				if tableRegex.MatchString(line) {
					matches := tableRegex.FindStringSubmatch(line)
					if len(matches) >= 2 {
						tableName := matches[1]
						summary.TablesChecked++
						validateTableName(path, lineNum, tableName, validator)
					}
				}

				// Look for column definitions
				columnRegex := regexp.MustCompile("[\"']?([a-zA-Z0-9_]+)[\"']?\\s+(varchar|int|bigint|text|boolean|timestamp|date|numeric|jsonb|uuid)")
				if columnRegex.MatchString(line) {
					matches := columnRegex.FindStringSubmatch(line)
					if len(matches) >= 3 {
						columnName := matches[1]
						columnType := matches[2]
						summary.ColumnsChecked++
						validateColumnName(path, lineNum, columnName, columnType, validator)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking SQL directory: %v\n", err)
	}
}

func validateTableName(file string, line int, tableName string, validator *NamingValidator) {
	// Table names should be snake_case and plural
	if !regexp.MustCompile(`^[a-z][a-z0-9_]*s$`).MatchString(tableName) {
		// Some tables are allowed to be singular
		allowedSingularTables := map[string]bool{
			"schema_migration":      true,
			"schema_info":           true,
			"schema_history":        true,
			"flyway_schema_history": true,
		}

		if !allowedSingularTables[tableName] {
			validator.AddError(&ValidationResult{
				File:        file,
				Line:        line,
				Column:      1,
				Type:        "Table name",
				Name:        tableName,
				Expected:    "snake_case and plural",
				Description: "Table names should use snake_case and be plural",
			})
		}
	}
}

func validateColumnName(file string, line int, columnName string, columnType string, validator *NamingValidator) {
	// Column names should be snake_case
	if !regexp.MustCompile(`^[a-z][a-z0-9_]*$`).MatchString(columnName) {
		validator.AddError(&ValidationResult{
			File:        file,
			Line:        line,
			Column:      1,
			Type:        "Column name",
			Name:        columnName,
			Expected:    "snake_case",
			Description: "Column names should use snake_case",
		})
	}

	// Check for proper suffix for ID columns
	if strings.HasSuffix(columnName, "id") && columnName != "id" {
		if !strings.HasSuffix(columnName, "_id") {
			validator.AddWarning(&ValidationResult{
				File:        file,
				Line:        line,
				Column:      1,
				Type:        "ID column",
				Name:        columnName,
				Expected:    "suffix with _id",
				Description: "Foreign key columns should end with _id",
			})
		}
	}

	// Check for proper prefix for boolean columns
	if columnType == "boolean" {
		booleanPrefixes := []string{"is_", "has_", "can_", "should_", "must_"}
		hasProperPrefix := false
		for _, prefix := range booleanPrefixes {
			if strings.HasPrefix(columnName, prefix) {
				hasProperPrefix = true
				break
			}
		}

		if !hasProperPrefix {
			validator.AddWarning(&ValidationResult{
				File:        file,
				Line:        line,
				Column:      1,
				Type:        "Boolean column",
				Name:        columnName,
				Expected:    "prefix with is_, has_, can_, etc.",
				Description: "Boolean columns should have a descriptive prefix like is_, has_, can_",
			})
		}
	}

	// Check for timestamp columns
	if columnType == "timestamp" {
		timestampSuffixes := []string{"_at", "_date", "_time"}
		hasProperSuffix := false
		for _, suffix := range timestampSuffixes {
			if strings.HasSuffix(columnName, suffix) {
				hasProperSuffix = true
				break
			}
		}

		if !hasProperSuffix && columnName != "created" && columnName != "updated" && columnName != "deleted" {
			validator.AddWarning(&ValidationResult{
				File:        file,
				Line:        line,
				Column:      1,
				Type:        "Timestamp column",
				Name:        columnName,
				Expected:    "suffix with _at, _date, or _time",
				Description: "Timestamp columns should have a descriptive suffix like _at, _date, or _time",
			})
		}
	}
}

func validateEntSchemas(entDir string, validator *NamingValidator, summary *NamingValidationSummary) {
	fmt.Println("Checking Ent schema naming conventions...")

	// Check if the Ent schema directory exists
	if _, err := os.Stat(entDir); os.IsNotExist(err) {
		fmt.Printf("Ent schema directory %s does not exist, skipping Ent schema naming checks\n", entDir)
		return
	}

	// Walk through the Ent schema files
	err := filepath.Walk(entDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only check Go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Parse the Go file
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				fmt.Printf("Error parsing file %s: %v\n", path, err)
				return nil
			}

			// Look for struct types that define Ent schemas
			for _, decl := range node.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								// Check if this is an Ent schema type
								isEntSchema := false
								for _, field := range structType.Fields.List {
									if len(field.Names) == 0 {
										// This is an embedded field, check if it's ent.Schema
										if ident, ok := field.Type.(*ast.SelectorExpr); ok {
											if ident.Sel.Name == "Schema" {
												isEntSchema = true
												break
											}
										}
									}
								}

								if isEntSchema {
									summary.SchemaTypesChecked++
									validateEntSchemaName(fset, typeSpec, path, validator)
								}
							}
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking Ent schema directory: %v\n", err)
	}
}

func validateEntSchemaName(fset *token.FileSet, typeSpec *ast.TypeSpec, path string, validator *NamingValidator) {
	schemaName := typeSpec.Name.String()

	// Ent schema names should be singular and PascalCase
	if !regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`).MatchString(schemaName) {
		pos := fset.Position(typeSpec.Name.Pos())
		validator.AddError(&ValidationResult{
			File:        path,
			Line:        pos.Line,
			Column:      pos.Column,
			Type:        "Ent schema",
			Name:        schemaName,
			Expected:    "PascalCase",
			Description: "Ent schema names should use PascalCase",
		})
	}

	// Ent schema names should be singular
	if strings.HasSuffix(schemaName, "s") {
		pos := fset.Position(typeSpec.Name.Pos())
		validator.AddError(&ValidationResult{
			File:        path,
			Line:        pos.Line,
			Column:      pos.Column,
			Type:        "Ent schema",
			Name:        schemaName,
			Expected:    "singular form",
			Description: "Ent schema names should be singular, not plural",
		})
	}
}
