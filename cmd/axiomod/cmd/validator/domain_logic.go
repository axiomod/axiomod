package validator

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ArchitectureRules represents the architecture rules from JSON file
type ArchitectureRules struct {
	AllowedDependencies map[string][]string `json:"allowedDependencies"`
	Exceptions          []string            `json:"exceptions"`
	DomainRules         DomainRules         `json:"domainRules"`
}

// DomainRules represents domain-specific rules
type DomainRules struct {
	AllowCrossDomainDependencies bool                    `json:"allowCrossDomainDependencies"`
	AllowedCrossDomainImports    []CrossDomainDependency `json:"allowedCrossDomainImports"`
}

// CrossDomainDependency represents an allowed dependency between domains
type CrossDomainDependency struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	Explanation string `json:"explanation,omitempty"`
}

// Import represents a single import statement in a Go file
type Import struct {
	Path string
	File string
}

// DomainValidationSummary holds summary data for the validation
type DomainValidationSummary struct {
	FilesScanned       int
	ImportsChecked     int
	TotalViolations    int
	ViolationsBySource map[string]int
	ViolationsByTarget map[string]int
	ViolationsByType   map[string]int
}

// ViolationDetail holds details about a specific violation
type ViolationDetail struct {
	Source        string
	Target        string
	ViolationType string
}

// RunDomainValidation validates domain boundaries in the codebase
func RunDomainValidation(path string, configPath string) (bool, error) {
	fmt.Println("Domain Boundary Validator")
	fmt.Println("========================")

	// Check if the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, fmt.Errorf("error: Path '%s' does not exist", path)
	}

	fmt.Printf("Validating domain boundaries in: %s\n", path)

	// Load architecture rules
	rules, err := loadDomainRules(configPath)
	if err != nil {
		return false, fmt.Errorf("error loading architecture rules: %v", err)
	}

	// Find all Go files in the given path
	var goFiles []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip exceptions
		for _, exception := range rules.Exceptions {
			if strings.Contains(path, exception) {
				return nil
			}
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("error walking directory: %v", err)
	}

	fmt.Printf("Found %d Go files to check\n", len(goFiles))

	// Initialize summary
	summary := DomainValidationSummary{
		FilesScanned:       len(goFiles),
		ViolationsBySource: make(map[string]int),
		ViolationsByTarget: make(map[string]int),
		ViolationsByType:   make(map[string]int),
	}

	// Extract imports from all Go files
	imports := make(map[string][]Import)
	for _, file := range goFiles {
		fileImports, err := extractImports(file)
		if err != nil {
			fmt.Printf("Error extracting imports from %s: %v\n", file, err)
			continue
		}

		summary.ImportsChecked += len(fileImports)

		// Identify the domain/module of the current file
		module := identifyModule(file, rules.AllowedDependencies)
		if module == "" {
			continue // Skip files that don't match any module
		}

		imports[module] = append(imports[module], fileImports...)
	}

	// Validate imports against rules
	validationErrors, violationDetails := validateImportsWithDetails(imports, rules)
	summary.TotalViolations = len(validationErrors)

	// Count violations by source, target and type
	for _, detail := range violationDetails {
		summary.ViolationsBySource[detail.Source]++
		summary.ViolationsByTarget[detail.Target]++
		summary.ViolationsByType[detail.ViolationType]++
	}

	// Print the summary report before details
	printDomainSummaryReport(summary)

	if len(validationErrors) > 0 {
		fmt.Println("\n❌ Domain boundary violations found:")
		for _, err := range validationErrors {
			fmt.Println(err)
		}

		// Print the summary again after violations
		fmt.Println("\nDomain Boundary Validation Summary:")
		printDomainSummaryReport(summary)

		return false, fmt.Errorf("domain validation failed with %d violations", summary.TotalViolations)
	}

	fmt.Println("\n✅ Domain boundaries are valid!")
	return true, nil
}

// printDomainSummaryReport prints a summary of the validation
func printDomainSummaryReport(summary DomainValidationSummary) {
	fmt.Printf("  Files scanned: %d\n", summary.FilesScanned)
	fmt.Printf("  Imports checked: %d\n", summary.ImportsChecked)
	fmt.Printf("  Total violations: %d\n", summary.TotalViolations)

	if summary.TotalViolations > 0 {
		fmt.Println("\nViolations by type:")
		printTopDomainViolations(summary.ViolationsByType)

		fmt.Println("\nTop violating source modules:")
		printTopDomainViolations(summary.ViolationsBySource)

		fmt.Println("\nTop violated target modules:")
		printTopDomainViolations(summary.ViolationsByTarget)
	}
}

// printTopDomainViolations prints the violations in descending order
func printTopDomainViolations(violations map[string]int) {
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

	// Print all items (limit to top 5 if there are many)
	count := len(items)
	if count > 5 {
		count = 5
	}

	for i := 0; i < count; i++ {
		fmt.Printf("  - %s: %d\n", items[i].item, items[i].count)
	}
}

// loadDomainRules loads the architecture rules from the JSON file
func loadDomainRules(configPath string) (*ArchitectureRules, error) {
	if configPath == "" {
		// Check for config file in default locations
		paths := []string{
			"architecture-rules.json",
			".architecture-rules.json",
			filepath.Join("configs", "architecture-rules.json"),
		}

		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	if configPath == "" {
		return getDefaultDomainRules(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var rules ArchitectureRules
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}

	return &rules, nil
}

// getDefaultDomainRules returns default architecture rules
func getDefaultDomainRules() *ArchitectureRules {
	return &ArchitectureRules{
		AllowedDependencies: map[string][]string{
			// Entity can only depend on itself
			"entity": {},
			// Repository can depend on entity
			"repository": {"entity"},
			// Usecase can depend on entity, repository, and service
			"usecase": {"entity", "repository", "service"},
			// Service can depend on entity and repository
			"service": {"entity", "repository"},
			// HTTP delivery can depend on usecase, entity, and middleware
			"delivery/http": {"usecase", "entity", "middleware"},
			// gRPC delivery can depend on usecase and entity
			"delivery/grpc": {"usecase", "entity"},
			// Infrastructure persistence can depend on entity and repository
			"infrastructure/persistence": {"entity", "repository"},
			// Infrastructure cache can depend on entity
			"infrastructure/cache": {"entity"},
			// Infrastructure messaging can depend on entity
			"infrastructure/messaging": {"entity"},
			// Platform can depend on config
			"platform/*": {"config"},
			// Plugins can depend on platform
			"plugins/*": {"platform/*"},

			// Full paths for internal modules
			"internal/example/entity":                     {},
			"internal/example/repository":                 {"internal/example/entity"},
			"internal/example/usecase":                    {"internal/example/entity", "internal/example/repository", "internal/example/service"},
			"internal/example/service":                    {"internal/example/entity", "internal/example/repository"},
			"internal/example/delivery/http":              {"internal/example/usecase", "internal/example/entity", "internal/example/delivery/http/middleware"},
			"internal/example/delivery/grpc":              {"internal/example/usecase", "internal/example/entity"},
			"internal/example/infrastructure/persistence": {"internal/example/entity", "internal/example/repository"},
			"internal/example/infrastructure/cache":       {"internal/example/entity"},
			"internal/example/infrastructure/messaging":   {"internal/example/entity"},
		},
		DomainRules: DomainRules{
			AllowCrossDomainDependencies: false,
			AllowedCrossDomainImports: []CrossDomainDependency{
				{
					Source:      "internal/example",
					Target:      "internal/platform",
					Explanation: "Example domain can import platform components",
				},
			},
		},
	}
}

// extractImports extracts all imports from a Go file
func extractImports(filePath string) ([]Import, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var imports []Import
	for _, imp := range node.Imports {
		// Remove quotes from import path
		path := strings.Trim(imp.Path.Value, "\"")

		// Only consider internal project imports
		if strings.Contains(path, "/internal/") || strings.Contains(path, "axiomod/") {
			imports = append(imports, Import{
				Path: path,
				File: filePath,
			})
		}
	}

	return imports, nil
}

// identifyModule identifies which module a file belongs to
func identifyModule(filePath string, allowedDeps map[string][]string) string {
	// Get relative path from working directory
	workDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	relPath, err := filepath.Rel(workDir, filePath)
	if err != nil {
		return ""
	}

	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	// Check if the file is in internal/example structure
	if strings.Contains(relPath, "internal/example/") {
		parts := strings.Split(relPath, "/")
		if len(parts) >= 4 {
			// For deeper paths like internal/example/delivery/http
			if len(parts) >= 5 && (parts[3] == "delivery" || parts[3] == "infrastructure") {
				return "internal/example/" + parts[3] + "/" + parts[4]
			}
			// For paths like internal/example/entity, internal/example/repository, etc.
			return "internal/example/" + parts[3]
		}
		return "internal/example"
	} else if strings.Contains(relPath, "internal/platform/") {
		parts := strings.Split(relPath, "/")
		if len(parts) >= 4 {
			return "internal/platform/" + parts[3]
		}
		return "internal/platform"
	} else if strings.Contains(relPath, "internal/plugins/") {
		parts := strings.Split(relPath, "/")
		if len(parts) >= 4 {
			return "internal/plugins/" + parts[3]
		}
		return "internal/plugins"
	}

	// Legacy code paths - check for simple module names
	for module := range allowedDeps {
		if !strings.Contains(module, "/") {
			// Simple module (legacy code)
			if strings.Contains(relPath, "internal/"+module+"/") ||
				strings.Contains(relPath, "/"+module+"/") {
				return module
			}
		}
	}

	return ""
}

// validateImportsWithDetails validates imports and returns errors and detailed violation info
func validateImportsWithDetails(imports map[string][]Import, rules *ArchitectureRules) ([]string, []ViolationDetail) {
	var errors []string
	var details []ViolationDetail

	for module, moduleImports := range imports {
		for _, imp := range moduleImports {
			// Convert import path to module format
			importModule := convertImportToModule(imp.Path)
			if importModule == "" {
				continue
			}

			// Skip self imports
			if importModule == module {
				continue
			}

			// Special case for domain-to-domain dependencies
			if (strings.HasPrefix(module, "internal/example/") && strings.HasPrefix(importModule, "internal/example/")) ||
				(strings.HasPrefix(module, "domain/") && strings.HasPrefix(importModule, "domain/")) {
				if !validateDomainDependency(module, importModule, rules) {
					errors = append(errors, fmt.Sprintf("Cross-domain dependency not allowed: '%s' should not import '%s' (in file %s)",
						module, importModule, imp.File))
					details = append(details, ViolationDetail{
						Source:        module,
						Target:        importModule,
						ViolationType: "cross-domain-dependency",
					})
					continue
				}
			} else {
				// Check if the import is allowed based on general rules
				allowed := false
				if allowedImports, ok := rules.AllowedDependencies[module]; ok {
					for _, allowedImport := range allowedImports {
						// Check direct match
						if importModule == allowedImport {
							allowed = true
							break
						}

						// Check pattern match (e.g., shared/* matches shared/errors, shared/logger, etc.)
						if strings.Contains(allowedImport, "*") {
							prefix := strings.Split(allowedImport, "*")[0]
							if strings.HasPrefix(importModule, prefix) {
								allowed = true
								break
							}
						}
					}
				}

				if !allowed {
					errors = append(errors, fmt.Sprintf("Module '%s' is not allowed to import '%s' (in file %s)",
						module, importModule, imp.File))
					details = append(details, ViolationDetail{
						Source:        module,
						Target:        importModule,
						ViolationType: "layer-dependency",
					})
				}
			}
		}
	}

	return errors, details
}

// validateDomainDependency checks if a domain-to-domain dependency is allowed
func validateDomainDependency(source, target string, rules *ArchitectureRules) bool {
	// If cross-domain dependencies are generally allowed
	if rules.DomainRules.AllowCrossDomainDependencies {
		return true
	}

	// Check specific allowed cross-domain dependencies
	for _, dep := range rules.DomainRules.AllowedCrossDomainImports {
		if dep.Source == source && dep.Target == target {
			return true
		}

		// Check if source is a subdomain of the allowed source
		if strings.HasPrefix(source, dep.Source+"/") && dep.Target == target {
			return true
		}

		// Check if target is a subdomain of the allowed target
		if source == dep.Source && strings.HasPrefix(target, dep.Target+"/") {
			return true
		}
	}

	return false
}

// convertImportToModule converts an import path to a module name
func convertImportToModule(importPath string) string {
	if strings.Contains(importPath, "/internal/example/") {
		parts := strings.Split(importPath, "/internal/example/")
		if len(parts) >= 2 {
			subParts := strings.Split(parts[1], "/")
			if len(subParts) >= 2 && (subParts[0] == "delivery" || subParts[0] == "infrastructure") {
				return "internal/example/" + subParts[0] + "/" + subParts[1]
			}
			return "internal/example/" + subParts[0]
		}
	} else if strings.Contains(importPath, "/internal/platform/") {
		parts := strings.Split(importPath, "/internal/platform/")
		if len(parts) >= 2 {
			subParts := strings.Split(parts[1], "/")
			return "internal/platform/" + subParts[0]
		}
	} else if strings.Contains(importPath, "/internal/plugins/") {
		parts := strings.Split(importPath, "/internal/plugins/")
		if len(parts) >= 2 {
			subParts := strings.Split(parts[1], "/")
			return "internal/plugins/" + subParts[0]
		}
	} else if strings.Contains(importPath, "/internal/") {
		parts := strings.Split(importPath, "/internal/")
		if len(parts) >= 2 {
			subParts := strings.Split(parts[1], "/")
			if len(subParts) >= 1 {
				return subParts[0]
			}
		}
	}

	return ""
}
