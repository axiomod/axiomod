package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	grpc_pkg "github.com/axiomod/axiomod/framework/grpc"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/gofiber/adaptor/v2"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger" // Import Fiber logger
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the fx options for the server module
var Module = fx.Options(
	fx.Provide(NewHTTPServer),
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	App    *fiber.App
	Config *config.Config
	Logger *observability.Logger
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(cfg *config.Config, obsLogger *observability.Logger, metrics *observability.Metrics, metricsMid *middleware.MetricsMiddleware, tracingMid *middleware.TracingMiddleware, h *health.Health) *HTTPServer {
	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		AppName:      cfg.App.Name,
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(compress.New())
	// Use Fiber's logger middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// Add metrics middleware
	app.Use(metricsMid.Handle())

	// Add tracing middleware
	app.Use(tracingMid.Handle())

	// Add health check endpoint (liveness)
	app.Get("/live", adaptor.HTTPHandlerFunc(h.Handler()))

	// Add readiness probe
	app.Get("/ready", adaptor.HTTPHandlerFunc(h.Handler()))

	// Add legacy health check for backward compatibility
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})

	// Add metrics endpoint
	app.Get("/metrics", adaptor.HTTPHandler(metrics.Handler))

	return &HTTPServer{
		App:    app,
		Config: cfg,
		Logger: obsLogger, // Use the observability logger for internal logging
	}
}

// RegisterHTTPServer registers the HTTP server with the fx lifecycle
func RegisterHTTPServer(lc fx.Lifecycle, server *HTTPServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start the server in a goroutine
			go func() {
				addr := fmt.Sprintf("%s:%d", server.Config.HTTP.Host, server.Config.HTTP.Port)
				server.Logger.Info("Starting HTTP server", zap.String("address", addr))
				if err := server.App.Listen(addr); err != nil && err != http.ErrServerClosed {
					server.Logger.Error("Failed to start HTTP server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.Logger.Info("Stopping HTTP server")
			return server.App.Shutdown()
		},
	})
}

// RegisterGRPCServer registers the gRPC server with the fx lifecycle
func RegisterGRPCServer(lc fx.Lifecycle, server *grpc_pkg.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.Start(); err != nil && err != http.ErrServerClosed {
					// logger is internal to server
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.Stop()
			return nil
		},
	})
}
