package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// generateModuleCmd represents the generate module command
var generateModuleCmd = &cobra.Command{
	Use:   "module --name=[name]",
	Short: "Generate a new module with basic structure",
	Long: `Generate a new module with a basic directory structure and placeholder files.

Example:
  axiomod generate module --name=user
`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("Error: name flag is required")
			os.Exit(1)
		}

		fmt.Printf("Generating module: %s\n", name)

		// Define paths
		moduleBasePath := filepath.Join("internal", "examples") // Changed from internal/modules to internal/examples
		modulePath := filepath.Join(moduleBasePath, name)
		entityPath := filepath.Join(modulePath, "entity")
		repositoryPath := filepath.Join(modulePath, "repository")
		usecasePath := filepath.Join(modulePath, "usecase")
		servicePath := filepath.Join(modulePath, "service")
		deliveryHTTPPath := filepath.Join(modulePath, "delivery", "http")
		deliveryGRPCPath := filepath.Join(modulePath, "delivery", "grpc")
		infraPersistencePath := filepath.Join(modulePath, "infrastructure", "persistence")
		infraCachePath := filepath.Join(modulePath, "infrastructure", "cache")
		infraMessagingPath := filepath.Join(modulePath, "infrastructure", "messaging")

		// Create directories
		dirs := []string{
			modulePath,
			entityPath,
			repositoryPath,
			usecasePath,
			servicePath,
			deliveryHTTPPath,
			deliveryGRPCPath,
			infraPersistencePath,
			infraCachePath,
			infraMessagingPath,
		}
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}

		// Define template data
		data := struct {
			ModuleName      string
			ModuleNameTitle string
			EntityName      string
			EntityNameLower string
			RepositoryName  string
			ServiceName     string
			HandlerName     string
			GRPCServiceName string
		}{
			ModuleName:      name,
			ModuleNameTitle: strings.Title(name),
			EntityName:      strings.Title(name),
			EntityNameLower: name,
			RepositoryName:  strings.Title(name) + "Repository",
			ServiceName:     strings.Title(name) + "Service",
			HandlerName:     strings.Title(name) + "Handler",
			GRPCServiceName: strings.Title(name) + "GRPCService",
		}

		// Generate placeholder files
		generateFile(entityTemplate, filepath.Join(entityPath, name+".go"), data)
		generateFile(repositoryTemplate, filepath.Join(repositoryPath, name+"_repository.go"), data)
		generateFile(usecaseTemplate, filepath.Join(usecasePath, "create_"+name+".go"), data)
		generateFile(serviceTemplate, filepath.Join(servicePath, name+"_domain_service.go"), data)
		generateFile(handlerTemplate, filepath.Join(deliveryHTTPPath, name+"_handler.go"), data)
		generateFile(grpcServiceTemplate, filepath.Join(deliveryGRPCPath, name+"_grpc_service.go"), data)
		generateFile(persistenceTemplate, filepath.Join(infraPersistencePath, name+"_memory_repository.go"), data)
		generateFile(moduleFileTemplate, filepath.Join(modulePath, "module.go"), data)

		fmt.Printf("\nModule %s generated successfully in %s\n", name, moduleBasePath)
		fmt.Println("\nRemember to:")
		fmt.Println("1. Implement the actual logic in the generated files.")
		fmt.Println("2. Add the module to your main application setup (e.g., FX options).")
	},
}

// Templates (simplified placeholders)
const entityTemplate = `package entity

import "time"

// {{.EntityName}} represents the core entity for the {{.ModuleName}} module.
type {{.EntityName}} struct {
	ID        string    ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}
`

const repositoryTemplate = `package repository

import (
	"context"
	"axiomod/internal/examples/{{.ModuleName}}/entity"
)

// {{.RepositoryName}} defines the interface for data access operations for {{.EntityName}}.
type {{.RepositoryName}} interface {
	Create(ctx context.Context, {{.EntityNameLower}} *entity.{{.EntityName}}) error
	GetByID(ctx context.Context, id string) (*entity.{{.EntityName}}, error)
	// Add other methods like Update, Delete, List, etc.
}
`

const usecaseTemplate = `package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"axiomod/internal/examples/{{.ModuleName}}/entity"
	"axiomod/internal/examples/{{.ModuleName}}/repository"
)

// Create{{.EntityName}}UseCase handles the creation of a new {{.EntityName}}.
type Create{{.EntityName}}UseCase struct {
	logger *zap.Logger
	 repo   repository.{{.RepositoryName}}
}

// NewCreate{{.EntityName}}UseCase creates a new Create{{.EntityName}}UseCase.
func NewCreate{{.EntityName}}UseCase(logger *zap.Logger, repo repository.{{.RepositoryName}}) *Create{{.EntityName}}UseCase {
	return &Create{{.EntityName}}UseCase{
		logger: logger,
		 repo:   repo,
	}
}

// Execute creates a new {{.EntityName}}.
func (uc *Create{{.EntityName}}UseCase) Execute(ctx context.Context, name string) (*entity.{{.EntityName}}, error) {
	 uc.logger.Info("Creating new {{.EntityNameLower}}", zap.String("name", name))

	 now := time.Now()
	 new{{.EntityName}} := &entity.{{.EntityName}}{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	 if err := uc.repo.Create(ctx, new{{.EntityName}}); err != nil {
		 uc.logger.Error("Failed to create {{.EntityNameLower}}", zap.Error(err))
		 return nil, err
	}

	 uc.logger.Info("{{.EntityNameLower}} created successfully", zap.String("id", new{{.EntityName}}.ID))
	 return new{{.EntityName}}, nil
}
`

const serviceTemplate = `package service

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
	// repo repository.{{.RepositoryName}}
}

// New{{.ServiceName}} creates a new {{.ModuleName}}Service.
func New{{.ServiceName}}(logger *zap.Logger /*, repo repository.{{.RepositoryName}}*/) {{.ServiceName}} {
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

const handlerTemplate = `package http

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	// "axiomod/internal/examples/{{.ModuleName}}/service"
)

// {{.HandlerName}} handles HTTP requests for the {{.ModuleName}} module.
type {{.HandlerName}} struct {
	logger  *zap.Logger
	// Add service dependency here
	// service service.{{.ServiceName}}
}

// New{{.HandlerName}} creates a new {{.HandlerName}}.
func New{{.HandlerName}}(logger *zap.Logger /*, service service.{{.ServiceName}}*/) *{{.HandlerName}} {
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
	 return c.Status(http.StatusOK).JSON(fiber.Map{"message": "GET /{{.ModuleName}} endpoint reached"})
}

// Add more handler methods here
`

const grpcServiceTemplate = `package grpc

import (
	"context"

	"go.uber.org/zap"
	// Import generated protobuf code
	// pb "axiomod/gen/proto/{{.ModuleName}}/v1"
	// "axiomod/internal/examples/{{.ModuleName}}/service"
)

// {{.GRPCServiceName}} implements the gRPC service for the {{.ModuleName}} module.
type {{.GRPCServiceName}} struct {
	// pb.Unimplemented{{.ModuleNameTitle}}ServiceServer // Embed the unimplemented server
	logger *zap.Logger
	// Add service dependency here
	// service service.{{.ServiceName}}
}

// New{{.GRPCServiceName}} creates a new {{.GRPCServiceName}}.
func New{{.GRPCServiceName}}(logger *zap.Logger /*, service service.{{.ServiceName}}*/) *{{.GRPCServiceName}} {
	return &{{.GRPCServiceName}}{
		logger: logger,
		// service: service,
	}
}

// Example RPC method
// func (s *{{.GRPCServiceName}}) Get{{.EntityName}}(ctx context.Context, req *pb.Get{{.EntityName}}Request) (*pb.Get{{.EntityName}}Response, error) {
// 	 s.logger.Info("Handling Get{{.EntityName}} gRPC request", zap.String("id", req.GetId()))
// 	 // Call service logic
// 	 return &pb.Get{{.EntityName}}Response{ /* ... */ }, nil
// }
`

const persistenceTemplate = `package persistence

import (
	"context"
	"fmt"
	"sync"

	"axiomod/internal/examples/{{.ModuleName}}/entity"
	"axiomod/internal/examples/{{.ModuleName}}/repository"
)

// InMemory{{.RepositoryName}} is an in-memory implementation of {{.RepositoryName}}.
type InMemory{{.RepositoryName}} struct {
	mu    sync.RWMutex
	store map[string]*entity.{{.EntityName}}
}

// NewInMemory{{.RepositoryName}} creates a new InMemory{{.RepositoryName}}.
func NewInMemory{{.RepositoryName}}() repository.{{.RepositoryName}} {
	return &InMemory{{.RepositoryName}}{
		store: make(map[string]*entity.{{.EntityName}}),
	}
}

// Create saves a new {{.EntityName}} in memory.
func (r *InMemory{{.RepositoryName}}) Create(ctx context.Context, {{.EntityNameLower}} *entity.{{.EntityName}}) error {
	 r.mu.Lock()
	 defer r.mu.Unlock()

	 if _, exists := r.store[{{.EntityNameLower}}.ID]; exists {
		 return fmt.Errorf("{{.EntityNameLower}} with ID %s already exists", {{.EntityNameLower}}.ID)
	}
	 r.store[{{.EntityNameLower}}.ID] = {{.EntityNameLower}}
	 return nil
}

// GetByID retrieves a {{.EntityName}} by ID from memory.
func (r *InMemory{{.RepositoryName}}) GetByID(ctx context.Context, id string) (*entity.{{.EntityName}}, error) {
	 r.mu.RLock()
	 defer r.mu.RUnlock()

	 {{.EntityNameLower}}, exists := r.store[id]
	 if !exists {
		 return nil, fmt.Errorf("{{.EntityNameLower}} with ID %s not found", id)
	}
	 return {{.EntityNameLower}}, nil
}

// Add other methods like Update, Delete, List, etc.
`

const moduleFileTemplate = `package {{.ModuleName}}

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"axiomod/internal/examples/{{.ModuleName}}/delivery/http"
	"axiomod/internal/examples/{{.ModuleName}}/delivery/grpc"
	"axiomod/internal/examples/{{.ModuleName}}/infrastructure/persistence"
	"axiomod/internal/examples/{{.ModuleName}}/repository"
	"axiomod/internal/examples/{{.ModuleName}}/service"
	"axiomod/internal/examples/{{.ModuleName}}/usecase"
)

// Module provides the FX module for the {{.ModuleName}} example.
var Module = fx.Module(
	"{{.ModuleName}}_example",
	fx.Provide(
		// Persistence
		 persistence.NewInMemory{{.RepositoryName}},
		 // Provide other persistence layers (e.g., Ent) here
		 // fx.Annotate(
		 // 	 persistence.NewEnt{{.RepositoryName}},
		 // 	 fx.As(new(repository.{{.RepositoryName}})),
		 // ),

		// Usecases
		 usecase.NewCreate{{.EntityName}}UseCase,
		 // Add other use cases here

		// Domain Services
		 service.New{{.ServiceName}},
		 // fx.Annotate(
		 // 	 service.New{{.ServiceName}},
		 // 	 fx.As(new(service.{{.ServiceName}})),
		 // ),

		// Delivery
		 http.New{{.HandlerName}},
		 grpc.New{{.GRPCServiceName}},
	),
	fx.Invoke(registerHooks),
)

// registerHooks registers hooks for the module, such as HTTP routes.
func registerHooks(lc fx.Lifecycle, logger *zap.Logger, handler *http.{{.HandlerName}} /*, grpcServer *grpc.{{.GRPCServiceName}}*/) {
	 logger.Info("Registering {{.ModuleName}} module hooks")
	 // Register HTTP routes (assuming a Fiber app is provided elsewhere)
	 // This requires the Fiber app instance to be available in the FX container.
	 // Example: app.Get("/{{.ModuleName}}", handler.HandleGet)

	 // Register gRPC service (assuming a gRPC server is provided elsewhere)
	 // Example: pb.Register{{.ModuleNameTitle}}ServiceServer(grpcServerInstance, grpcServer)
}
`

// NewGenerateModuleCmd returns the generate module command.
func NewGenerateModuleCmd() *cobra.Command {
	generateModuleCmd.Flags().StringP("name", "n", "", "Name of the module (required)")
	generateModuleCmd.MarkFlagRequired("name")
	return generateModuleCmd
}

func init() {
	// Add subcommands to the parent generateCmd
	generateCmd.AddCommand(generateModuleCmd)
}
