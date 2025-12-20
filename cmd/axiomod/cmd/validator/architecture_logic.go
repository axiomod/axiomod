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

// Configuration defines the dependency rules between packages
type Configuration struct {
	// Map of package layers to their allowed dependencies
	AllowedDependencies map[string][]string `json:"allowedDependencies"`
	// Exceptions are specific files that can bypass the rules
	Exceptions []string `json:"exceptions,omitempty"`
	// PatternRules defines wildcard patterns for dependencies
	PatternRules []PatternRule `json:"patternRules,omitempty"`
	// DomainRules defines rules specific to domains
	DomainRules DomainRuleSet `json:"domainRules,omitempty"`
}

// PatternRule defines a pattern-based rule for dependencies
type PatternRule struct {
	// Pattern is the wildcard pattern for source packages
	Pattern string `json:"pattern"`
	// AllowedToImport is the wildcard pattern for allowed imports
	AllowedToImport string `json:"allowedToImport"`
	// Explanation provides context for this rule
	Explanation string `json:"explanation"`
}

// DomainRuleSet defines rules for domain dependencies
type DomainRuleSet struct {
	// AllowCrossDomainDependencies indicates if direct domain-to-domain dependencies are allowed
	AllowCrossDomainDependencies bool `json:"allowCrossDomainDependencies"`
	// AllowedCrossDomainImports lists specific cross-domain dependencies that are allowed
	AllowedCrossDomainImports []DomainDependency `json:"allowedCrossDomainImports,omitempty"`
	// AllowedInternalStructure defines what imports are allowed within a domain
	AllowedInternalStructure map[string][]string `json:"allowedInternalStructure,omitempty"`
}

// DomainDependency defines a specific allowed cross-domain dependency
type DomainDependency struct {
	// Source domain
	Source string `json:"source"`
	// Target domain
	Target string `json:"target"`
	// Optional explanation
	Explanation string `json:"explanation,omitempty"`
}

// ValidationSummary holds information about validation results
type ValidationSummary struct {
	// Total number of files tested
	FilesChecked int
	// Total number of imports checked
	ImportsChecked int
	// Total number of violations found
	TotalViolations int
	// Violations by rule category
	ViolationsByCategory map[string]int
	// Violations by source package
	ViolationsBySource map[string]int
	// Violations by target package
	ViolationsByTarget map[string]int
}

// Violation represents a single architecture rule violation
type Violation struct {
	// Source package with the illegal import
	Source string
	// Target package being imported illegally
	Target string
	// Full file path where violation was found
	FilePath string
	// Line number where violation was found
	Line int
	// Category of violation
	Category string
}

// getDefaultConfig returns the default configuration with predefined dependency rules
func getDefaultConfig() Configuration {
	return Configuration{
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
		},
		PatternRules: []PatternRule{
			{
				Pattern:         "internal/*/entity",
				AllowedToImport: "pkg/*",
				Explanation:     "Entities can import utility packages",
			},
			{
				Pattern:         "internal/*/repository",
				AllowedToImport: "platform/ent",
				Explanation:     "Repositories can import the Ent ORM",
			},
			{
				Pattern:         "internal/*/delivery/*",
				AllowedToImport: "platform/server",
				Explanation:     "Delivery layers can import server components",
			},
		},
		DomainRules: DomainRuleSet{
			AllowCrossDomainDependencies: false,
			AllowedCrossDomainImports: []DomainDependency{
				{
					Source:      "examples/example",
					Target:      "platform",
					Explanation: "Example domain can import platform components",
				},
			},
			AllowedInternalStructure: map[string][]string{
				"entity":                     {},
				"repository":                 {"entity"},
				"usecase":                    {"entity", "repository", "service"},
				"service":                    {"entity", "repository"},
				"delivery/http":              {"usecase", "entity", "middleware"},
				"delivery/grpc":              {"usecase", "entity"},
				"infrastructure/persistence": {"entity", "repository"},
				"infrastructure/cache":       {"entity"},
				"infrastructure/messaging":   {"entity"},
			},
		},
	}
}

// RunArchitectureValidation validates the architecture of the codebase
func RunArchitectureValidation(rootDir string, configPath string) (bool, error) {
	fmt.Println("Running architecture validation...")

	// Load configuration
	config := loadConfiguration(configPath)

	// Start the validation process
	violations, summary := validateArchitecture(rootDir, config)

	// Print summary report before violations
	printSummaryReport(summary)

	if summary.TotalViolations > 0 {
		fmt.Println("\n❌ Architecture violations found:")
		for _, v := range violations {
			fmt.Printf("  - %s\n", v)
		}

		// Print the summary report again after violations for better visibility
		fmt.Println("\nArchitecture Validation Summary:")
		printSummaryReport(summary)
		return false, fmt.Errorf("architecture validation failed with %d violations", summary.TotalViolations)
	}

	fmt.Println("\n✅ Architecture validation passed!")
	return true, nil
}

// printSummaryReport prints a summary report of architecture validation
func printSummaryReport(summary ValidationSummary) {
	fmt.Printf("  Files checked: %d\n", summary.FilesChecked)
	fmt.Printf("  Imports checked: %d\n", summary.ImportsChecked)
	fmt.Printf("  Total violations: %d\n", summary.TotalViolations)

	if summary.TotalViolations > 0 {
		fmt.Println("\nViolations by rule category:")
		for category, count := range summary.ViolationsByCategory {
			fmt.Printf("  - %s: %d\n", category, count)
		}

		fmt.Println("\nTop violating source packages:")
		printTopViolations(summary.ViolationsBySource, 5)

		fmt.Println("\nTop violated target packages:")
		printTopViolations(summary.ViolationsByTarget, 5)
	}
}

// printTopViolations prints the top N violating packages
func printTopViolations(violations map[string]int, topN int) {
	// Convert map to slice for sorting
	type packageViolation struct {
		pkg   string
		count int
	}

	var items []packageViolation
	for pkg, count := range violations {
		items = append(items, packageViolation{pkg, count})
	}

	// Sort by violation count (descending)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].count < items[j].count {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Print top N (or all if less than N)
	count := topN
	if len(items) < count {
		count = len(items)
	}

	for i := 0; i < count; i++ {
		fmt.Printf("  - %s: %d\n", items[i].pkg, items[i].count)
	}
}

// loadConfiguration loads dependency rules from a JSON config file if specified
// otherwise returns the default configuration
func loadConfiguration(configPath string) Configuration {
	if configPath == "" {
		// Check for config file in default locations
		potentialPaths := []string{
			"architecture-rules.json",
			".architecture-rules.json",
			filepath.Join("configs", "architecture-rules.json"),
		}

		for _, path := range potentialPaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	// If no config file is found or specified, use default configuration
	if configPath == "" {
		return getDefaultConfig()
	}

	// Load config from file
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Warning: Could not read config file %s: %v\nUsing default configuration.\n", configPath, err)
		return getDefaultConfig()
	}

	var config Configuration
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Warning: Could not parse config file %s: %v\nUsing default configuration.\n", configPath, err)
		return getDefaultConfig()
	}

	return config
}

// isWildcardMatch checks if a string matches a wildcard pattern
// The pattern can contain '*' which matches any sequence of characters
func isWildcardMatch(pattern, str string) bool {
	// If pattern contains no wildcards, do direct comparison
	if !strings.Contains(pattern, "*") {
		return pattern == str
	}

	// Split pattern by '*'
	parts := strings.Split(pattern, "*")

	// If the pattern starts with '*', the first part is empty
	// so we'll start matching from the beginning
	startIdx := 0

	// Check if the string starts with the first pattern part (if not empty)
	if parts[0] != "" {
		if !strings.HasPrefix(str, parts[0]) {
			return false
		}
		startIdx = len(parts[0])
	}

	// For each middle part of the pattern
	for i := 1; i < len(parts)-1; i++ {
		part := parts[i]
		// Skip empty parts (consecutive '*')
		if part == "" {
			continue
		}

		// Find this part in the remaining string
		idx := strings.Index(str[startIdx:], part)
		if idx == -1 {
			return false // Part not found
		}

		// Move startIdx past the found part
		startIdx += idx + len(part)
	}

	// Check if the string ends with the last pattern part (if not empty)
	lastPart := parts[len(parts)-1]
	if lastPart != "" {
		return strings.HasSuffix(str, lastPart)
	}

	// If the pattern ends with '*', we've matched all parts
	return true
}

// isDomainPath checks if a path is within a domain package
func isDomainPath(path string) bool {
	return strings.HasPrefix(path, "domain/") || strings.Contains(path, "/examples/example/")
}

// extractDomainName extracts the domain name from a path like "domain/example/model"
func extractDomainName(path string) string {
	if strings.HasPrefix(path, "domain/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[1]
		}
	} else if strings.Contains(path, "/examples/example/") {
		return "example"
	}
	return ""
}

// checkDomainDependency checks if a domain dependency is allowed
func checkDomainDependency(sourceDomain, targetDomain string, config Configuration) bool {
	// Same domain is always allowed
	if sourceDomain == targetDomain {
		return true
	}

	// If cross-domain dependencies are allowed in general
	if config.DomainRules.AllowCrossDomainDependencies {
		return true
	}

	// Check specific allowed cross-domain dependencies
	for _, dep := range config.DomainRules.AllowedCrossDomainImports {
		if dep.Source == sourceDomain && dep.Target == targetDomain {
			return true
		}
	}

	return false
}

// checkInternalDomainStructure checks if within a domain, the file dependencies follow allowed patterns
func checkInternalDomainStructure(sourceFile, targetFile string, config Configuration) bool {
	// Extract file types (e.g., "model", "service", etc.) from paths
	// Assuming paths like "domain/example/model.go" or "domain/example/service/user_service.go"
	sourceType := getFileType(sourceFile)
	targetType := getFileType(targetFile)

	if sourceType == "" || targetType == "" {
		// If we can't determine types, default to allowing the dependency
		return true
	}

	// Check if this internal structure dependency is allowed
	allowedTargets, exists := config.DomainRules.AllowedInternalStructure[sourceType]
	if !exists {
		// If no specific rule, default to allowing
		return true
	}

	// Check if the target type is in the allowed list
	for _, allowed := range allowedTargets {
		if allowed == targetType || isWildcardMatch(allowed, targetType) {
			return true
		}
	}

	return false
}

// getFileType extracts the type of file from a path
// For example, from "domain/example/model.go" extracts "model"
// From "domain/example/service/user_service.go" extracts "service"
func getFileType(path string) string {
	// Remove the .go extension if present
	path = strings.TrimSuffix(path, ".go")

	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return ""
	}

	// If it's a file directly in the domain root, like model.go
	if len(parts) == 3 {
		return parts[2]
	}

	// If it's in a subdirectory, like service/user_service.go
	return parts[3]
}

// checkLayerDependency checks if a dependency is allowed based on direct matches or pattern rules
// Returns a boolean indicating if allowed and a category string for violation reporting
func checkLayerDependency(currentLayer, importedLayer string, config Configuration) (bool, string) {
	// Handle domain-specific rules
	if isDomainPath(currentLayer) && isDomainPath(importedLayer) {
		sourceDomain := extractDomainName(currentLayer)
		targetDomain := extractDomainName(importedLayer)

		// Check if cross-domain dependency is allowed
		if !checkDomainDependency(sourceDomain, targetDomain, config) {
			return false, "cross-domain-dependency"
		}

		// If it's within the same domain, check internal structure
		if sourceDomain == targetDomain {
			if !checkInternalDomainStructure(currentLayer, importedLayer, config) {
				return false, "domain-internal-structure"
			}
		}
	}

	// Check direct allowed dependencies
	if allowedLayers, exists := config.AllowedDependencies[currentLayer]; exists {
		for _, layer := range allowedLayers {
			// Check exact match
			if layer == importedLayer {
				return true, ""
			}
			// Check wildcard match
			if strings.Contains(layer, "*") && isWildcardMatch(layer, importedLayer) {
				return true, ""
			}
		}
	}

	// Check wildcard source layer match
	for wildcardLayer, allowedLayers := range config.AllowedDependencies {
		if strings.Contains(wildcardLayer, "*") && isWildcardMatch(wildcardLayer, currentLayer) {
			for _, layer := range allowedLayers {
				// Check exact match
				if layer == importedLayer {
					return true, ""
				}
				// Check wildcard match
				if strings.Contains(layer, "*") && isWildcardMatch(layer, importedLayer) {
					return true, ""
				}
			}
		}
	}

	// Check pattern rules
	for _, rule := range config.PatternRules {
		if isWildcardMatch(rule.Pattern, currentLayer) && isWildcardMatch(rule.AllowedToImport, importedLayer) {
			return true, ""
		}
	}

	return false, "layer-dependency"
}

func validateArchitecture(rootDir string, config Configuration) ([]string, ValidationSummary) {
	var violations []string
	var structuralViolations []Violation

	// Initialize summary
	summary := ValidationSummary{
		ViolationsByCategory: make(map[string]int),
		ViolationsBySource:   make(map[string]int),
		ViolationsByTarget:   make(map[string]int),
	}

	// Walk through all Go files in the project
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Check if this file is exempt from validation
		for _, exception := range config.Exceptions {
			if strings.Contains(path, exception) {
				return nil
			}
		}

		// Increment file counter
		summary.FilesChecked++

		// Determine the package from the file path
		relPath, err := filepath.Rel(rootDir, filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Convert path separators to '/' for consistency across platforms
		relPath = filepath.ToSlash(relPath)

		// Determine current layer or package
		currentLayer := relPath

		// For old structure compatibility
		if !strings.Contains(relPath, "/") {
			currentLayer = relPath
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Check each import statement
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")

			// Increment import counter
			summary.ImportsChecked++

			// Only check project imports
			if !strings.Contains(importPath, "internal/") && !strings.Contains(importPath, "github.com/axiomod/axiomod/") {
				continue
			}

			// Extract the layer/package from the import path
			var importedLayer string
			if strings.HasPrefix(importPath, "github.com/axiomod/axiomod/") {
				importedLayer = strings.TrimPrefix(importPath, "github.com/axiomod/axiomod/")
			} else {
				continue // Skip internal/ and external imports
			}

			// Check if this dependency is allowed
			allowed, category := checkLayerDependency(currentLayer, importedLayer, config)
			if !allowed {
				// Get position information for the import
				pos := fset.Position(imp.Pos())

				// Record the violation
				violation := fmt.Sprintf("%s:%d: %s imports %s (not allowed)", path, pos.Line, currentLayer, importedLayer)
				violations = append(violations, violation)

				// Record detailed violation for statistics
				structuralViolations = append(structuralViolations, Violation{
					Source:   currentLayer,
					Target:   importedLayer,
					FilePath: path,
					Line:     pos.Line,
					Category: category,
				})

				// Update summary counters
				summary.TotalViolations++
				summary.ViolationsByCategory[category]++
				summary.ViolationsBySource[currentLayer]++
				summary.ViolationsByTarget[importedLayer]++
			}
		}

		return nil
	})

	if err != nil {
		violations = append(violations, fmt.Sprintf("Error during validation: %v", err))
	}

	return violations, summary
}
