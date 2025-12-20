package example

import (
	"axiomod/internal/examples/example/delivery/grpc"
	"axiomod/internal/examples/example/delivery/http"
	"axiomod/internal/examples/example/delivery/http/middleware"
	"axiomod/internal/examples/example/infrastructure/persistence"
	"axiomod/internal/examples/example/repository"
	"axiomod/internal/examples/example/service"
	"axiomod/internal/examples/example/usecase"
	"axiomod/internal/platform/observability"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	grpc_go "google.golang.org/grpc" // Renamed import to avoid conflict
)

// Module provides the fx options for the example module
var Module = fx.Options(
	// Provide repositories
	fx.Provide(persistence.NewExampleMemoryRepository),
	fx.Provide(func(repo *persistence.ExampleMemoryRepository) repository.ExampleRepository {
		return repo
	}),

	// Provide use cases
	fx.Provide(usecase.NewCreateExampleUseCase),
	fx.Provide(usecase.NewGetExampleUseCase),

	// Provide domain services
	fx.Provide(service.NewExampleDomainService),

	// Provide HTTP handlers and middleware
	fx.Provide(middleware.NewAuthMiddleware),
	fx.Provide(middleware.NewLoggingMiddleware),
	fx.Provide(http.NewExampleHandler),

	// Provide gRPC services
	fx.Provide(grpc.NewExampleGRPCService),

	// Register HTTP routes
	fx.Invoke(registerHTTPRoutes),

	// Register gRPC services
	fx.Invoke(registerGRPCServices),
)

// registerHTTPRoutes registers the HTTP routes for the example module
func registerHTTPRoutes(app *fiber.App, handler *http.ExampleHandler, authMiddleware *middleware.AuthMiddleware, loggingMiddleware *middleware.LoggingMiddleware) {
	api := app.Group("/api/v1")

	// Apply middleware
	api.Use(loggingMiddleware.Handle())
	api.Use(authMiddleware.Handle())

	// Register routes
	 handler.RegisterRoutes(api)
}

// registerGRPCServices registers the gRPC services for the example module
func registerGRPCServices(server *grpc_go.Server, service *grpc.ExampleGRPCService, logger *observability.Logger) {
	// In a real implementation, we would register the gRPC service with the server
	// For example: pb.RegisterExampleServiceServer(server, service)
	logger.Info("Registered example gRPC service")
}
