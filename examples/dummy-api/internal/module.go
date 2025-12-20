package internal

import (
	"examples/dummy-api/internal/delivery/http"

	"github.com/axiomod/axiomod/platform/server"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"dummy_api",
	fx.Provide(
		http.NewDummyHandler,
	),
	fx.Invoke(registerRoutes),
)

func registerRoutes(handler *http.DummyHandler, s *server.HTTPServer) {
	// Register routes with the framework's Fiber app
	handler.RegisterRoutes(s.App)
}
