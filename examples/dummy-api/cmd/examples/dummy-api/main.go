package main

import (
	"context"
	"examples/dummy-api/internal"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/worker"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/axiomod/axiomod/platform/server"
	"github.com/axiomod/axiomod/plugins"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to config file (default: config/service_default.yaml)")
	flag.Parse()

	// Determine config path
	resolvedConfigPath := *configPath
	if resolvedConfigPath == "" {
		resolvedConfigPath = "config/service_default.yaml"
	}

	// Load initial config just for logger setup (optional, can use default)
	tempCfg, err := config.Load(resolvedConfigPath)
	if err != nil {
		fmt.Printf("Warning: could not load config for initial logger setup: %v\n", err)
		// Use default logger config or handle error
	}

	// Setup initial logger (can be replaced by FX provided logger later)
	initialLogger, err := observability.NewLogger(tempCfg)
	if err != nil {
		fmt.Printf("Error creating initial logger: %v\n", err)
		os.Exit(1)
	}
	initialLogger.Info("Starting application", zap.String("configPath", resolvedConfigPath))

	// Create application with dependencies
	app := fx.New(
		// Provide the configuration
		fx.Provide(func() (*config.Config, error) {
			return config.Load(resolvedConfigPath)
		}),

		// Core platform modules
		observability.Module,
		server.Module,
		plugins.Module,
		worker.Module,
		internal.Module,

		// Register HTTP and gRPC servers
		fx.Invoke(
			server.RegisterHTTPServer,
			server.RegisterGRPCServer,
		),

		// Register lifecycle hooks
		fx.Invoke(func(lc fx.Lifecycle, logger *observability.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Starting application")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Stopping application")
					return nil
				},
			})
		}),
	)

	// Start the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		initialLogger.Info("Received signal, shutting down...")
		cancel()
	}()

	// Start and wait for context cancellation
	if err := app.Start(ctx); err != nil {
		initialLogger.Fatal("Error starting application", zap.Error(err))
	}
	<-ctx.Done()

	// Stop the application
	initialLogger.Info("Initiating graceful shutdown...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second) // Add timeout
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		initialLogger.Error("Error stopping application", zap.Error(err))
		os.Exit(1)
	}
	initialLogger.Info("Application stopped gracefully")
}
