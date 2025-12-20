package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"

	"go.uber.org/fx"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
	}()

	// Create and start the application using fx
	app := fx.New(
		// Provide the configuration
		fx.Provide(func() (*config.Config, error) {
			return config.Load(*configPath)
		}),

		// Register all modules
		fx.Options(
			getModuleOptions()...,
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
	if err := app.Start(ctx); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Wait for context cancellation (from signal handler)
	<-ctx.Done()

	// Stop the application gracefully
	stopCtx, stopCancel := context.WithTimeout(context.Background(), app.StopTimeout())
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop application gracefully: %v\n", err)
		os.Exit(1)
	}
}
