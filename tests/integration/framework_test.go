package integration

import (
	"context"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/worker"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/axiomod/axiomod/platform/server"
	"github.com/axiomod/axiomod/plugins"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestFrameworkBootstrap(t *testing.T) {
	// Simple integration test to ensure all modules can be provided and injected
	// without violating dependencies or failing initialization.

	t.Run("Full Bootstrap", func(t *testing.T) {
		app := fx.New(
			fx.Provide(func() (*config.Config, error) {
				return &config.Config{
					App:           config.AppConfig{Name: "integration-test"},
					HTTP:          config.HTTPConfig{Port: 0}, // Random port
					Observability: config.ObservabilityConfig{MetricsEnabled: false},
				}, nil
			}),
			observability.Module,
			server.Module,
			plugins.Module,
			worker.Module,

			// Invoke to trigger start/stop hooks
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return nil
					},
				})
			}),

			// Suppress Fx logs in test
			fx.NopLogger,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := app.Start(ctx)
		assert.NoError(t, err)

		err = app.Stop(ctx)
		assert.NoError(t, err)
	})
}
