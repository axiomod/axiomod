// Package kafka provides a plugin that manages a Kafka producer built on
// framework/kafka, with configuration via plugin settings.
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	fwkafka "github.com/axiomod/axiomod/framework/kafka"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/plugins"

	"go.uber.org/zap"
)

// Plugin manages a Kafka producer.
type Plugin struct {
	logger *observability.Logger

	producerConfig *fwkafka.ProducerConfig
	producer       *fwkafka.Producer
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "kafka"
}

// Initialize configures the plugin from its settings.
//
// Settings:
//   - brokers ([]string or []interface{}): broker addresses, default ["localhost:9092"]
//   - clientId (string): client ID, default "go-axiomod"
//   - retries (int): max produce retries, default 3
//   - timeout (string): Go duration for dial/read/write timeouts, default "10s"
func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, h *health.Health) error {
	settings = plugins.NormalizeSettings(settings)
	p.logger = logger
	p.producerConfig = fwkafka.DefaultProducerConfig()

	if brokers := toStringSlice(settings["brokers"]); len(brokers) > 0 {
		p.producerConfig.Brokers = brokers
	}
	if clientID, ok := settings["clientid"].(string); ok && clientID != "" {
		p.producerConfig.ClientID = clientID
	}
	if retries, ok := settings["retries"].(int); ok && retries > 0 {
		p.producerConfig.Retries = retries
	}
	if timeout, ok := settings["timeout"].(string); ok && timeout != "" {
		parsed, err := time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("kafka plugin: invalid timeout %q: %w", timeout, err)
		}
		p.producerConfig.Timeout = parsed
	}

	if h != nil {
		h.RegisterCheck(p.Name(), p.ping)
	}

	return nil
}

// Start creates the producer and connects to the brokers.
func (p *Plugin) Start() error {
	producer, err := fwkafka.NewProducer(p.logger, p.producerConfig)
	if err != nil {
		return fmt.Errorf("kafka plugin: failed to create producer: %w", err)
	}
	p.producer = producer

	if p.logger != nil {
		p.logger.Info("Kafka plugin started", zap.Strings("brokers", p.producerConfig.Brokers))
	}
	return nil
}

// Stop closes the producer.
func (p *Plugin) Stop() error {
	if p.producer != nil {
		err := p.producer.Close()
		p.producer = nil
		return err
	}
	return nil
}

// Producer returns the underlying producer, or nil before Start.
func (p *Plugin) Producer() *fwkafka.Producer {
	return p.producer
}

// Publish publishes a message via the managed producer.
func (p *Plugin) Publish(ctx context.Context, topic, key string, value []byte) error {
	if p.producer == nil {
		return fmt.Errorf("kafka plugin: not started")
	}
	return p.producer.Publish(ctx, topic, key, value)
}

// ping reports whether the producer has been started.
func (p *Plugin) ping() error {
	if p.producer == nil {
		return fmt.Errorf("kafka plugin: producer not started")
	}
	return nil
}

// toStringSlice converts a settings value into a []string, accepting both
// []string and the []interface{} produced by YAML decoding.
func toStringSlice(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return nil
	}
}
