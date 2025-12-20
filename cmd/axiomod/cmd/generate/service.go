package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// generateServiceCmd represents the generate service command
var generateServiceCmd = &cobra.Command{
	Use:   "service --name=[name]",
	Short: "Generate a new domain service",
	Long: `Generate a new domain service interface and basic implementation.

Example:
  axiomod generate service --name=auth
`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("Error: name flag is required")
			os.Exit(1)
		}

		fmt.Printf("Generating domain service: %s\n", name)

		// Define paths (assuming it belongs to an existing module)
		// TODO: Add flag to specify module or improve discovery
		moduleName := name // Assuming service name matches module for simplicity
		modulePath := filepath.Join("internal", "examples", moduleName)
		servicePath := filepath.Join(modulePath, "service")
		repositoryPath := filepath.Join(modulePath, "repository")

		// Create directories if they don't exist
		os.MkdirAll(servicePath, 0755)
		os.MkdirAll(repositoryPath, 0755) // Ensure repo dir exists for import

		// Define template data
		data := struct {
			ModuleName      string
			ModuleNameTitle string
			ServiceName     string
			RepositoryName  string
			EntityName      string
		}{
			ModuleName:      moduleName,
			ModuleNameTitle: strings.Title(moduleName),
			ServiceName:     strings.Title(name) + "Service",
			RepositoryName:  strings.Title(moduleName) + "Repository", // Assuming repo name convention
			EntityName:      strings.Title(moduleName),                // Assuming entity name convention
		}

		// Generate service file
		serviceFilePath := filepath.Join(servicePath, name+"_domain_service.go")
		if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) {
			serviceTemplate := `package service

import (
	"context"

	"go.uber.org/zap"
	// Import repository and entity if needed
	// "github.com/axiomod/axiomod/examples/{{.ModuleName}}/entity"
	"github.com/axiomod/axiomod/examples/{{.ModuleName}}/repository"
)

// {{.ServiceName}} defines the interface for the {{.ModuleName}} service.
type {{.ServiceName}} interface {
	// Define service methods here
	ProcessData(ctx context.Context, data string) error
}

// {{.ModuleName}}Service implements the {{.ServiceName}} interface.
type {{.ModuleName}}Service struct {
	logger *zap.Logger
	 repo   repository.{{.RepositoryName}}
}

// New{{.ServiceName}} creates a new {{.ModuleName}}Service.
func New{{.ServiceName}}(logger *zap.Logger, repo repository.{{.RepositoryName}}) {{.ServiceName}} {
	return &{{.ModuleName}}Service{
		logger: logger,
		 repo:   repo,
	}
}

// ProcessData is an example service method.
func (s *{{.ModuleName}}Service) ProcessData(ctx context.Context, data string) error {
	 s.logger.Info("Processing data in {{.ModuleName}} service", zap.String("data", data))
	// Implement logic here, potentially calling the repository
	// Example: entity, err := s.repo.GetByID(ctx, data)
	 return nil
}
`
			generateFile(serviceTemplate, serviceFilePath, data)
		} else {
			fmt.Printf("Service file already exists: %s\n", serviceFilePath)
		}

		// Generate basic repository file (if it doesn't exist)
		repositoryFilePath := filepath.Join(repositoryPath, name+"_repository.go")
		if _, err := os.Stat(repositoryFilePath); os.IsNotExist(err) {
			repositoryTemplate := `package repository

import (
	"context"
	"github.com/axiomod/axiomod/examples/{{.ModuleName}}/entity"
)

// {{.RepositoryName}} defines the interface for data access operations for {{.EntityName}}.
type {{.RepositoryName}} interface {
	// Define repository methods here
	GetByID(ctx context.Context, id string) (*entity.{{.EntityName}}, error)
	Save(ctx context.Context, entity *entity.{{.EntityName}}) error
}
`
			generateFile(repositoryTemplate, repositoryFilePath, data)
		} else {
			fmt.Printf("Repository file already exists: %s\n", repositoryFilePath)
		}

		fmt.Printf("\nDomain service %s generated successfully.", name)
		fmt.Println("\nRemember to:")
		fmt.Println("1. Implement the actual logic in the service and repository.")
		fmt.Println("2. Add the service and repository implementation to your dependency injection setup.")
	},
}

// NewGenerateServiceCmd returns the generate service command.
func NewGenerateServiceCmd() *cobra.Command {
	generateServiceCmd.Flags().StringP("name", "n", "", "Name of the service (required)")
	generateServiceCmd.MarkFlagRequired("name")
	return generateServiceCmd
}

func init() {
	// Add subcommands to the parent generateCmd
	generateCmd.AddCommand(generateServiceCmd)
}
