package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"axiomod/internal/platform/observability"

	"go.uber.org/zap"
)

// Common errors
var (
	ErrTopicEmpty     = errors.New("topic cannot be empty")
	ErrPayloadEmpty   = errors.New("payload cannot be empty")
	ErrPublishTimeout = errors.New("publish timeout")
)

// Event represents a message to be published or consumed
type Event struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Payload   json.RawMessage   `json:"payload"`
	Timestamp time.Time         `json:"timestamp"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// Publisher defines the interface for publishing events
type Publisher interface {
	// Publish publishes an event to the specified topic
	Publish(ctx context.Context, topic string, payload []byte, headers map[string]string) error

	// Close closes the publisher
	Close() error
}

// Consumer defines the interface for consuming events
type Consumer interface {
	// Subscribe subscribes to the specified topics and calls the handler for each event
	Subscribe(ctx context.Context, topics []string, handler func(ctx context.Context, event Event) error) error

	// Close closes the consumer
	Close() error
}

// EventBus provides a simple in-memory event bus for testing
type EventBus struct {
	subscribers map[string][]func(ctx context.Context, event Event) error
	mu          sync.RWMutex
	logger      *observability.Logger
}

// NewEventBus creates a new in-memory event bus
func NewEventBus(logger *observability.Logger) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(ctx context.Context, event Event) error),
		logger:      logger,
	}
}

// Publish publishes an event to the specified topic
func (b *EventBus) Publish(ctx context.Context, topic string, payload []byte, headers map[string]string) error {
	if topic == "" {
		return ErrTopicEmpty
	}
	if len(payload) == 0 {
		return ErrPayloadEmpty
	}

	event := Event{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
		Headers:   headers,
	}

	b.mu.RLock()
	subscribers := b.subscribers[topic]
	b.mu.RUnlock()

	for _, handler := range subscribers {
		go func(h func(ctx context.Context, event Event) error) {
			if err := h(ctx, event); err != nil {
				b.logger.Error("Failed to handle event",
					zap.String("topic", topic),
					zap.String("id", event.ID),
					zap.Error(err),
				)
			}
		}(handler)
	}

	return nil
}

// Subscribe subscribes to the specified topics and calls the handler for each event
func (b *EventBus) Subscribe(ctx context.Context, topics []string, handler func(ctx context.Context, event Event) error) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, topic := range topics {
		b.subscribers[topic] = append(b.subscribers[topic], handler)
		b.logger.Info("Subscribed to topic", zap.String("topic", topic))
	}

	return nil
}

// Close closes the event bus
func (b *EventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = make(map[string][]func(ctx context.Context, event Event) error)
	return nil
}
