package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/axiomod/axiomod/platform/observability"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// Common errors
var (
	ErrInvalidConfig = errors.New("invalid kafka configuration")
	ErrNotConnected  = errors.New("not connected to kafka")
)

// Producer is a Kafka producer
type Producer struct {
	producer sarama.SyncProducer
	logger   *observability.Logger
	config   *ProducerConfig
}

// ProducerConfig contains configuration for the Kafka producer
type ProducerConfig struct {
	Brokers  []string
	ClientID string
	Retries  int
	Timeout  time.Duration
}

// DefaultProducerConfig returns the default producer configuration
func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		Brokers:  []string{"localhost:9092"},
		ClientID: "go-axiomod",
		Retries:  3,
		Timeout:  time.Second * 10,
	}
}

// NewProducer creates a new Kafka producer
func NewProducer(logger *observability.Logger, config *ProducerConfig) (*Producer, error) {
	if config == nil {
		config = DefaultProducerConfig()
	}

	if len(config.Brokers) == 0 {
		return nil, ErrInvalidConfig
	}

	// Create Sarama config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = config.Retries
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.ClientID = config.ClientID
	saramaConfig.Net.DialTimeout = config.Timeout
	saramaConfig.Net.ReadTimeout = config.Timeout
	saramaConfig.Net.WriteTimeout = config.Timeout

	// Create producer
	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		logger.Error("Failed to create Kafka producer", zap.Error(err))
		return nil, err
	}

	logger.Info("Created Kafka producer", zap.Strings("brokers", config.Brokers))

	return &Producer{
		producer: producer,
		logger:   logger,
		config:   config,
	}, nil
}

// Publish publishes a message to a topic
func (p *Producer) Publish(ctx context.Context, topic string, key string, value []byte) error {
	if p.producer == nil {
		return ErrNotConnected
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}

	// Add context deadline if available
	if deadline, ok := ctx.Deadline(); ok {
		msg.Metadata = deadline
	}

	// Publish message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("Failed to publish message",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	p.logger.Debug("Published message",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	if p.producer == nil {
		return nil
	}

	if err := p.producer.Close(); err != nil {
		p.logger.Error("Failed to close Kafka producer", zap.Error(err))
		return err
	}

	p.logger.Info("Closed Kafka producer")
	return nil
}

// Consumer is a Kafka consumer
type Consumer struct {
	consumer sarama.ConsumerGroup
	logger   *observability.Logger
	config   *ConsumerConfig
	handlers map[string]MessageHandler
}

// ConsumerConfig contains configuration for the Kafka consumer
type ConsumerConfig struct {
	Brokers   []string
	GroupID   string
	ClientID  string
	Topics    []string
	Offset    int64
	MinBytes  int
	MaxBytes  int
	MaxWait   time.Duration
	Timeout   time.Duration
	Processor MessageProcessor
}

// MessageProcessor processes messages from Kafka
type MessageProcessor interface {
	Process(ctx context.Context, message *Message) error
}

// MessageHandler handles messages from Kafka
type MessageHandler func(ctx context.Context, message *Message) error

// Message represents a Kafka message
type Message struct {
	Topic     string
	Key       string
	Value     []byte
	Partition int32
	Offset    int64
	Timestamp time.Time
	Headers   map[string]string
}

// DefaultConsumerConfig returns the default consumer configuration
func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:  []string{"localhost:9092"},
		GroupID:  "go-axiomod",
		ClientID: "go-axiomod",
		Topics:   []string{},
		Offset:   sarama.OffsetNewest,
		MinBytes: 1,
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
		Timeout:  time.Second * 10,
	}
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(logger *observability.Logger, config *ConsumerConfig) (*Consumer, error) {
	if config == nil {
		config = DefaultConsumerConfig()
	}

	if len(config.Brokers) == 0 {
		return nil, ErrInvalidConfig
	}

	if len(config.Topics) == 0 {
		return nil, ErrInvalidConfig
	}

	// Create Sarama config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = config.Offset
	saramaConfig.Consumer.MaxWaitTime = config.MaxWait
	saramaConfig.Consumer.Fetch.Min = int32(config.MinBytes)
	saramaConfig.Consumer.Fetch.Max = int32(config.MaxBytes)
	saramaConfig.ClientID = config.ClientID
	saramaConfig.Net.DialTimeout = config.Timeout
	saramaConfig.Net.ReadTimeout = config.Timeout
	saramaConfig.Net.WriteTimeout = config.Timeout

	// Create consumer group
	consumer, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		logger.Error("Failed to create Kafka consumer", zap.Error(err))
		return nil, err
	}

	logger.Info("Created Kafka consumer",
		zap.Strings("brokers", config.Brokers),
		zap.String("group", config.GroupID),
		zap.Strings("topics", config.Topics),
	)

	return &Consumer{
		consumer: consumer,
		logger:   logger,
		config:   config,
		handlers: make(map[string]MessageHandler),
	}, nil
}

// RegisterHandler registers a handler for a topic
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	if c.consumer == nil {
		return ErrNotConnected
	}

	// Create consumer handler
	handler := &consumerHandler{
		logger:    c.logger,
		handlers:  c.handlers,
		processor: c.config.Processor,
	}

	// Start consuming
	go func() {
		for {
			if err := c.consumer.Consume(ctx, c.config.Topics, handler); err != nil {
				c.logger.Error("Error from consumer", zap.Error(err))
			}

			// Check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				c.logger.Info("Stopping Kafka consumer", zap.Error(ctx.Err()))
				return
			}
		}
	}()

	c.logger.Info("Started Kafka consumer",
		zap.String("group", c.config.GroupID),
		zap.Strings("topics", c.config.Topics),
	)

	return nil
}

// Close closes the consumer
func (c *Consumer) Close() error {
	if c.consumer == nil {
		return nil
	}

	if err := c.consumer.Close(); err != nil {
		c.logger.Error("Failed to close Kafka consumer", zap.Error(err))
		return err
	}

	c.logger.Info("Closed Kafka consumer")
	return nil
}

// consumerHandler implements sarama.ConsumerGroupHandler
type consumerHandler struct {
	logger    *observability.Logger
	handlers  map[string]MessageHandler
	processor MessageProcessor
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// Create message
		message := &Message{
			Topic:     msg.Topic,
			Partition: msg.Partition,
			Offset:    msg.Offset,
			Timestamp: msg.Timestamp,
			Headers:   make(map[string]string),
		}

		// Set key if available
		if msg.Key != nil {
			message.Key = string(msg.Key)
		}

		// Set value if available
		if msg.Value != nil {
			message.Value = msg.Value
		}

		// Set headers if available
		for _, header := range msg.Headers {
			message.Headers[string(header.Key)] = string(header.Value)
		}

		// Process message
		var err error
		if h.processor != nil {
			err = h.processor.Process(session.Context(), message)
		} else if handler, ok := h.handlers[msg.Topic]; ok {
			err = handler(session.Context(), message)
		} else {
			h.logger.Warn("No handler for topic", zap.String("topic", msg.Topic))
		}

		if err != nil {
			h.logger.Error("Failed to process message",
				zap.String("topic", msg.Topic),
				zap.String("key", message.Key),
				zap.Int32("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
				zap.Error(err),
			)
		} else {
			// Mark message as processed
			session.MarkMessage(msg, "")

			h.logger.Debug("Processed message",
				zap.String("topic", msg.Topic),
				zap.String("key", message.Key),
				zap.Int32("partition", msg.Partition),
				zap.Int64("offset", msg.Offset),
			)
		}
	}

	return nil
}
