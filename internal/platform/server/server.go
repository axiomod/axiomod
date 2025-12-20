package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"axiomod/internal/framework/config"
	"axiomod/internal/platform/observability"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger" // Import Fiber logger
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Module provides the fx options for the server module
var Module = fx.Options(
	fx.Provide(NewHTTPServer),
	fx.Provide(NewGRPCServer),
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	App    *fiber.App
	Config *config.Config
	Logger *observability.Logger
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(cfg *config.Config, obsLogger *observability.Logger, metrics *observability.Metrics) *HTTPServer {
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

	// Add health check endpoint (liveness)
	app.Get("/live", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "alive"})
	})

	// Add readiness probe
	app.Get("/ready", func(c *fiber.Ctx) error {
		// In a real app, this should check DB connections, etc.
		return c.JSON(map[string]string{"status": "ready"})
	})

	// Add legacy health check for backward compatibility
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})

	// Add metrics endpoint
	app.Get("/metrics", func(c *fiber.Ctx) error {
		// Create a handler that will serve the metrics
		return c.SendStatus(fiber.StatusOK)
		// Note: In a real implementation, we would need to properly adapt the http.Handler to fiber
	})

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

// GRPCServer represents the gRPC server
type GRPCServer struct {
	Server *grpc.Server
	Config *config.Config
	Logger *observability.Logger
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(cfg *config.Config, logger *observability.Logger) *GRPCServer {
	// Create a new gRPC server
	server := grpc.NewServer()

	// Enable reflection for development
	if cfg.App.Debug {
		reflection.Register(server)
	}

	return &GRPCServer{
		Server: server,
		Config: cfg,
		Logger: logger,
	}
}

// RegisterGRPCServer registers the gRPC server with the fx lifecycle
func RegisterGRPCServer(lc fx.Lifecycle, server *GRPCServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start the server in a goroutine
			go func() {
				addr := fmt.Sprintf("%s:%d", server.Config.GRPC.Host, server.Config.GRPC.Port)
				server.Logger.Info("Starting gRPC server", zap.String("address", addr))
				lis, err := net.Listen("tcp", addr)
				if err != nil {
					server.Logger.Error("Failed to listen", zap.Error(err))
					return
				}
				if err := server.Server.Serve(lis); err != nil {
					server.Logger.Error("Failed to start gRPC server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.Logger.Info("Stopping gRPC server")
			server.Server.GracefulStop()
			return nil
		},
	})
}
