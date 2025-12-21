package kafka

import (
	"context"

	"go.uber.org/fx"
)

// Module provides the fx options for the kafka module
var Module = fx.Options(
	fx.Provide(DefaultProducerConfig),
	fx.Provide(NewProducer),
	fx.Provide(DefaultConsumerConfig),
	fx.Provide(NewConsumer),
	fx.Invoke(RegisterProducerLifecycle),
	fx.Invoke(RegisterConsumerLifecycle),
)

// RegisterProducerLifecycle registers lifecycle hooks for the Kafka producer
func RegisterProducerLifecycle(lc fx.Lifecycle, producer *Producer) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return producer.Close()
		},
	})
}

// RegisterConsumerLifecycle registers lifecycle hooks for the Kafka consumer
func RegisterConsumerLifecycle(lc fx.Lifecycle, consumer *Consumer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start consumer in background
			return consumer.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return consumer.Close()
		},
	})
}
