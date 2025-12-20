package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"axiomod/internal/examples/example/entity"
	"axiomod/internal/platform/observability"

	"go.uber.org/zap"
)

// ExampleEvent represents an event related to an Example entity
type ExampleEvent struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// EventType constants
const (
	ExampleCreatedEvent = "example.created"
	ExampleUpdatedEvent = "example.updated"
	ExampleDeletedEvent = "example.deleted"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, topic string, event []byte) error
}

// ExampleEventPublisher publishes events related to Example entities
type ExampleEventPublisher struct {
	publisher EventPublisher
	logger    *observability.Logger
}

// NewExampleEventPublisher creates a new ExampleEventPublisher
func NewExampleEventPublisher(publisher EventPublisher, logger *observability.Logger) *ExampleEventPublisher {
	return &ExampleEventPublisher{
		publisher: publisher,
		logger:    logger,
	}
}

// PublishCreated publishes an ExampleCreatedEvent
func (p *ExampleEventPublisher) PublishCreated(ctx context.Context, example *entity.Example) error {
	return p.publishEvent(ctx, ExampleCreatedEvent, example)
}

// PublishUpdated publishes an ExampleUpdatedEvent
func (p *ExampleEventPublisher) PublishUpdated(ctx context.Context, example *entity.Example) error {
	return p.publishEvent(ctx, ExampleUpdatedEvent, example)
}

// PublishDeleted publishes an ExampleDeletedEvent
func (p *ExampleEventPublisher) PublishDeleted(ctx context.Context, id string) error {
	event := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}

	return p.publishEventData(ctx, ExampleDeletedEvent, id, event)
}

// publishEvent publishes an event with the Example entity as data
func (p *ExampleEventPublisher) publishEvent(ctx context.Context, eventType string, example *entity.Example) error {
	return p.publishEventData(ctx, eventType, example.ID, example)
}

// publishEventData publishes an event with the given data
func (p *ExampleEventPublisher) publishEventData(ctx context.Context, eventType string, id string, data interface{}) error {
	// Marshal the data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		p.logger.Error("Failed to marshal event data", zap.Error(err))
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Create the event
	event := ExampleEvent{
		Type:      eventType,
		ID:        id,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		Data:      dataBytes,
	}

	// Marshal the event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal event", zap.Error(err))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish the event
	topic := fmt.Sprintf("examples.%s", eventType)
	if err := p.publisher.Publish(ctx, topic, eventBytes); err != nil {
		p.logger.Error("Failed to publish event",
			zap.String("topic", topic),
			zap.String("type", eventType),
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info("Published event",
		zap.String("topic", topic),
		zap.String("type", eventType),
		zap.String("id", id))
	return nil
}
