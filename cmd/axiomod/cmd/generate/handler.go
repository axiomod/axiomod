package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// generateHandlerCmd represents the generate handler command
var generateHandlerCmd = &cobra.Command{
	Use:   "handler --name=[name]",
	Short: "Generate a new HTTP handler",
	Long: `Generate a new HTTP handler with associated files.

Example:
  axiomod generate handler --name=product
`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("Error: name flag is required")
			os.Exit(1)
		}

		fmt.Printf("Generating HTTP handler: %s\n", name)

		// Define paths
		modulePath := filepath.Join("internal", "examples", name)
		handlerPath := filepath.Join(modulePath, "delivery", "http")
		servicePath := filepath.Join(modulePath, "service")
		entityPath := filepath.Join(modulePath, "entity")

		// Create directories if they don't exist
		os.MkdirAll(handlerPath, 0755)
		os.MkdirAll(servicePath, 0755)
		os.MkdirAll(entityPath, 0755)

		// Define template data
		data := struct {
			ModuleName      string
			ModuleNameTitle string
			EntityName      string
			EntityNameLower string
			ServiceName     string
			HandlerName     string
		}{
			ModuleName:      name,
			ModuleNameTitle: strings.Title(name),
			EntityName:      strings.Title(name), // Assuming entity name matches module name
			EntityNameLower: name,
			ServiceName:     strings.Title(name) + "Service",
			HandlerName:     strings.Title(name) + "Handler",
		}

		// Generate handler file
		handlerTemplate := `package http

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"axiomod/internal/examples/{{.ModuleName}}/service"
)

// {{.HandlerName}} handles HTTP requests for the {{.ModuleName}} module.
type {{.HandlerName}} struct {
	logger  *zap.Logger
	// Add service dependency here
	// service *service.{{.ServiceName}}
}

// New{{.HandlerName}} creates a new {{.HandlerName}}.
func New{{.HandlerName}}(logger *zap.Logger /*, service *service.{{.ServiceName}}*/) *{{.HandlerName}} {
	return &{{.HandlerName}}{
		logger:  logger,
		// service: service,
	}
}

// RegisterRoutes registers the handler routes with the Fiber app.
func (h *{{.HandlerName}}) RegisterRoutes(app *fiber.App) {
	// Define routes for the {{.ModuleName}} module
	group := app.Group("/{{.ModuleName}}")

	group.Get("/", h.handleGet{{.ModuleNameTitle}})
	// Add more routes here (POST, PUT, DELETE, etc.)
}

// handleGet{{.ModuleNameTitle}} handles GET requests for {{.ModuleName}}.
func (h *{{.HandlerName}}) handleGet{{.ModuleNameTitle}}(c *fiber.Ctx) error {
	 h.logger.Info("Handling GET /{{.ModuleName}}")

	// Example: Call service method
	// data, err := h.service.GetData(c.Context())
	// if err != nil {
	// 	 h.logger.Error("Failed to get data", zap.Error(err))
	// 	 return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve data"})
	// }

	// return c.JSON(data)
	 return c.Status(http.StatusOK).JSON(fiber.Map{"message": "GET /{{.ModuleName}} endpoint reached"})
}

// Add more handler methods here
`
		generateFile(handlerTemplate, filepath.Join(handlerPath, name+"_handler.go"), data)

		// Generate basic service file (if it doesn't exist)
		serviceFilePath := filepath.Join(servicePath, name+"_domain_service.go")
		if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) {
			serviceTemplate := `package service

import (
	"context"

	"go.uber.org/zap"
	// Import repository and entity if needed
	// "axiomod/internal/examples/{{.ModuleName}}/entity"
	// "axiomod/internal/examples/{{.ModuleName}}/repository"
)

// {{.ServiceName}} defines the interface for the {{.ModuleName}} service.
type {{.ServiceName}} interface {
	// Define service methods here
	GetData(ctx context.Context) (string, error)
}

// {{.ModuleName}}Service implements the {{.ServiceName}} interface.
type {{.ModuleName}}Service struct {
	logger *zap.Logger
	// Add repository dependency here
	// repo repository.{{.EntityName}}Repository
}

// New{{.ServiceName}} creates a new {{.ModuleName}}Service.
func New{{.ServiceName}}(logger *zap.Logger /*, repo repository.{{.EntityName}}Repository*/) {{.ServiceName}} {
	return &{{.ModuleName}}Service{
		logger: logger,
		// repo: repo,
	}
}

// GetData is an example service method.
func (s *{{.ModuleName}}Service) GetData(ctx context.Context) (string, error) {
	 s.logger.Info("Getting data in {{.ModuleName}} service")
	// Implement logic here, potentially calling the repository
	 return "Data from {{.ModuleName}} service", nil
}
`
			generateFile(serviceTemplate, serviceFilePath, data)
		} else {
			fmt.Printf("Service file already exists: %s\n", serviceFilePath)
		}

		// Generate basic entity file (if it doesn't exist)
		entityFilePath := filepath.Join(entityPath, name+".go")
		if _, err := os.Stat(entityFilePath); os.IsNotExist(err) {
			entityTemplate := `package entity

import "time"

// {{.EntityName}} represents the core entity for the {{.ModuleName}} module.
type {{.EntityName}} struct {
	ID        string    ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}
`
			generateFile(entityTemplate, entityFilePath, data)
		} else {
			fmt.Printf("Entity file already exists: %s\n", entityFilePath)
		}

		fmt.Printf("\nHTTP handler %s generated successfully.\n", name)
		fmt.Println("\nRemember to:")
		fmt.Println("1. Implement the actual logic in the handler and service.")
		fmt.Println("2. Define the entity structure properly.")
		fmt.Println("3. Add the service and handler to your dependency injection setup (e.g., FX module).")
		fmt.Println("4. Register the handler routes in your main server setup.")
	},
}

// generateFile creates a file from a template.
func generateFile(tmplContent, filePath string, data interface{}) {
	tmpl, err := template.New(filepath.Base(filePath)).Parse(tmplContent)
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", filepath.Base(filePath), err)
		os.Exit(1)
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filePath, err)
		os.Exit(1)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("Error executing template %s: %v\n", filepath.Base(filePath), err)
		os.Exit(1)
	}
	fmt.Printf("Generated file: %s\n", filePath)
}

// NewGenerateHandlerCmd returns the generate handler command.
func NewGenerateHandlerCmd() *cobra.Command {
	generateHandlerCmd.Flags().StringP("name", "n", "", "Name of the handler (required)")
	generateHandlerCmd.MarkFlagRequired("name")
	return generateHandlerCmd
}

func init() {
	// Add subcommands to the parent generateCmd
	generateCmd.AddCommand(generateHandlerCmd)
}
