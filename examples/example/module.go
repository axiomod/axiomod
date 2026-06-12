package example

import (
	"github.com/axiomod/axiomod/examples/example/delivery/grpc"
	"github.com/axiomod/axiomod/examples/example/delivery/http"
	"github.com/axiomod/axiomod/examples/example/infrastructure/persistence"
	"github.com/axiomod/axiomod/examples/example/repository"
	"github.com/axiomod/axiomod/examples/example/service"
	"github.com/axiomod/axiomod/examples/example/usecase"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/framework/observability"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	grpc_go "google.golang.org/grpc" // Renamed import to avoid conflict
)

// Module provides the fx options for the example module
var Module = fx.Options(
	// Provide the repository: in-memory by default; switch to a database
	// implementation (Ent or plain SQL per database.orm) by enabling the
	// postgres/mysql plugin and providing a *database.DB — see ProvideRepository.
	fx.Provide(persistence.NewExampleMemoryRepository),
	fx.Provide(ProvideRepository),

	// Provide use cases
	fx.Provide(usecase.NewCreateExampleUseCase),
	fx.Provide(usecase.NewGetExampleUseCase),
	fx.Provide(usecase.NewUpdateExampleUseCase),
	fx.Provide(usecase.NewDeleteExampleUseCase),
	fx.Provide(usecase.NewListExamplesUseCase),

	// Provide domain services
	fx.Provide(service.NewExampleDomainService),

	// Provide HTTP handlers
	fx.Provide(http.NewExampleHandler),

	// Provide gRPC services
	fx.Provide(grpc.NewExampleGRPCService),

	// Register HTTP routes
	fx.Invoke(registerHTTPRoutes),

	// Register gRPC services
	fx.Invoke(registerGRPCServices),
)

// ProvideRepository selects the repository implementation. The in-memory
// repository keeps the server bootable with zero infrastructure; database
// implementations are constructed by domain wiring when a DB is available
// (see NewExampleEntRepository / NewExampleSQLRepository).
func ProvideRepository(memory *persistence.ExampleMemoryRepository, cfg *config.Config) repository.ExampleRepository {
	return memory
}

// registerHTTPRoutes registers the HTTP routes for the example module using
// the framework's logging and JWT-auth middleware.
func registerHTTPRoutes(
	app *fiber.App,
	handler *http.ExampleHandler,
	authMiddleware *middleware.AuthMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
) {
	api := app.Group("/api/v1")

	// Apply middleware
	api.Use(loggingMiddleware.Handle())
	api.Use(authMiddleware.Handle())

	// Register routes
	handler.RegisterRoutes(api)
}

// registerGRPCServices registers the gRPC services for the example module
func registerGRPCServices(server *grpc_go.Server, service *grpc.ExampleGRPCService, logger *observability.Logger) {
	grpc.RegisterExampleServiceServer(server, service)
	logger.Info("Registered example gRPC service")
}
