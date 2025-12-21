package worker

import (
	"context"

	"go.uber.org/fx"
)

// Module provides the fx options for the worker module
var Module = fx.Options(
	fx.Provide(New),
	fx.Invoke(RegisterWorker),
)

// RegisterWorker registers the worker with the fx lifecycle
func RegisterWorker(lc fx.Lifecycle, w *Worker) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			w.StopAll()
			return nil
		},
	})
}
