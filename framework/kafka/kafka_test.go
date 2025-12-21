package kafka

import (
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/stretchr/testify/assert"
)

func TestKafkaProducerConfig(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		cfg := DefaultProducerConfig()
		assert.Equal(t, []string{"localhost:9092"}, cfg.Brokers)
		assert.Equal(t, "go-axiomod", cfg.ClientID)
		assert.Equal(t, 3, cfg.Retries)
		assert.Equal(t, 10*time.Second, cfg.Timeout)
	})

	t.Run("NewProducer Validation", func(t *testing.T) {
		logger, _ := observability.NewLogger(&config.Config{})

		// Should fail with empty brokers if we were to pass empty config
		// But NewProducer fills defaults if nil.
		// Let's pass a config with empty brokers explicitly
		cfg := &ProducerConfig{
			Brokers: []string{},
		}
		_, err := NewProducer(logger, cfg)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidConfig, err)
	})
}

func TestKafkaConsumerConfig(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		cfg := DefaultConsumerConfig()
		assert.Equal(t, []string{"localhost:9092"}, cfg.Brokers)
		assert.Equal(t, "go-axiomod", cfg.GroupID)
	})

	t.Run("NewConsumer Validation", func(t *testing.T) {
		logger, _ := observability.NewLogger(&config.Config{})

		cfg := &ConsumerConfig{
			Brokers: []string{"localhost:9092"},
			Topics:  []string{}, // Empty topics
		}
		_, err := NewConsumer(logger, cfg)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidConfig, err)

		cfg.Topics = []string{"topic1"}
		cfg.Brokers = []string{} // Empty brokers
		_, err = NewConsumer(logger, cfg)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidConfig, err)
	})
}
