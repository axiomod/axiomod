package main

import (
	"github.com/axiomod/axiomod/framework/worker"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/axiomod/axiomod/platform/server"
	"github.com/axiomod/axiomod/plugins"

	"go.uber.org/fx"
)

// getModuleOptions returns all the fx.Option instances for the application modules
func getModuleOptions() []fx.Option {
	return []fx.Option{
		// Core platform modules
		observability.Module,
		server.Module,
		plugins.Module,
		worker.Module,

		// Domain modules
		// Add your domain modules here, for example:
		// example.Module,

		// Register constructors for any additional dependencies
		fx.Provide(
		// Add your providers here
		),

		// Register invocations for any startup hooks
		fx.Invoke(
			// Register HTTP and gRPC servers
			server.RegisterHTTPServer,
			server.RegisterGRPCServer,
		),
	}
}
