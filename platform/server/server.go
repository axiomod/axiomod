package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	grpc_pkg "github.com/axiomod/axiomod/framework/grpc"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/middleware"
	"github.com/axiomod/axiomod/framework/observability"
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
	// Expose the Fiber app so domain modules can register routes via the
	// documented registerHTTPRoutes(app *fiber.App, ...) pattern.
	fx.Provide(func(s *HTTPServer) *fiber.App { return s.App }),
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
			// Bind synchronously so that startup fails fast if the port is
			// unavailable, then serve in the background.
			addr := fmt.Sprintf("%s:%d", server.Config.HTTP.Host, server.Config.HTTP.Port)
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("failed to bind HTTP server on %s: %w", addr, err)
			}
			server.Logger.Info("Starting HTTP server", zap.String("address", addr))
			go func() {
				if err := server.App.Listener(ln); err != nil && err != http.ErrServerClosed {
					server.Logger.Error("HTTP server terminated with error", zap.Error(err))
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
			// The listener is already bound in grpc.NewServer, so bind
			// failures surface during construction. Serve errors are logged
			// inside Server.Start.
			go func() {
				_ = server.Start()
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.Stop()
			return nil
		},
	})
}
