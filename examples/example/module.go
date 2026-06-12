package example

import (
	"context"
	"time"

	"github.com/axiomod/axiomod/examples/example/delivery/grpc"
	"github.com/axiomod/axiomod/examples/example/delivery/http"
	"github.com/axiomod/axiomod/examples/example/infrastructure/persistence"
	"github.com/axiomod/axiomod/examples/example/repository"
	"github.com/axiomod/axiomod/examples/example/service"
	"github.com/axiomod/axiomod/examples/example/usecase"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/database"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/platform/ent"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	grpc_go "google.golang.org/grpc" // Renamed import to avoid conflict
)

// Module provides the fx options for the example module
var Module = fx.Options(
	// Repository selection: in-memory when no database plugin is enabled
	// (zero-infrastructure boot); otherwise Ent or plain SQL according to
	// database.orm — see ProvideRepository.
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

// ProvideRepository selects the repository implementation:
//
//   - no database plugin enabled: thread-safe in-memory repository, so the
//     server boots with zero external infrastructure;
//   - postgres or mysql enabled: a database-backed repository — Ent (the
//     framework default) or plain database/sql, chosen by database.orm.
func ProvideRepository(
	cfg *config.Config,
	logger *observability.Logger,
	metrics *observability.Metrics,
	h *health.Health,
) (repository.ExampleRepository, error) {
	if !cfg.Plugins.Enabled["postgres"] && !cfg.Plugins.Enabled["mysql"] {
		return persistence.NewExampleMemoryRepository(), nil
	}

	// A database plugin is enabled: connect through the framework pool so
	// slow-query logging, metrics, and the health check are registered.
	db, err := database.Connect(cfg, logger, metrics, h)
	if err != nil {
		return nil, err
	}

	if cfg.Database.ORM == "sql" {
		return persistence.NewExampleSQLRepository(db.GetDB(), logger), nil
	}

	// Default: Ent.
	client, err := ent.NewClientFromDB(cfg.Database.Driver, db.GetDB())
	if err != nil {
		return nil, err
	}
	repo := persistence.NewExampleEntRepository(client, logger)

	// Ensure the schema exists before serving traffic.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := repo.Migrate(ctx); err != nil {
		return nil, err
	}
	return repo, nil
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
