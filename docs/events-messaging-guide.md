# Events & Messaging Guide

The Axiomod framework uses Kafka for event-driven communication, powered by the `IBM/sarama` library. This guide covers how to produce and consume messages.

## 1. Configuration

Kafka settings are managed in your application configuration.

```yaml
kafka:
  brokers:
    - localhost:9092
  clientId: axiomod-service
  groupId: axiomod-group
```

## 2. Producing Messages

The `kafka.Producer` provides a simple way to publish messages to a topic.

### Initialization

Inject the `*kafka.Producer` into your use cases or services:

```go
type MyUseCase struct {
    producer *kafka.Producer
}

func NewMyUseCase(producer *kafka.Producer) *MyUseCase {
    return &MyUseCase{producer: producer}
}
```

### Publishing

```go
func (uc *MyUseCase) Execute(ctx context.Context) error {
    topic := "user-created"
    key := "user-123"
    value := []byte(`{"id": "user-123", "name": "John Doe"}`)
    
    return uc.producer.Publish(ctx, topic, key, value)
}
```

## 3. Consuming Messages

The `kafka.Consumer` allows you to register handlers for specific topics and process them asynchronously.

### Registering Handlers

You can register handlers using the `RegisterHandler` method.

```go
func RegisterKafkaHandlers(consumer *kafka.Consumer) {
    consumer.RegisterHandler("user-created", func(ctx context.Context, msg *kafka.Message) error {
        fmt.Printf("Received message: %s\n", string(msg.Value))
        return nil
    })
}
```

### Starting the Consumer

The `kafka` module automatically manages the lifecycle of the consumer. Just include `kafka.Module` in your Fx application options and ensure your configuration is correct.

```go
fx.New(
    kafka.Module,
    // ...
)
```

## 4. Message Structure

The `kafka.Message` struct provides access to the message payload and metadata:

- `Topic`: The source topic.
- `Key`: The message key.
- `Value`: The raw byte payload.
- `Partition/Offset`: Tracking information.
- `Headers`: Custom metadata headers.

## 5. Error Handling & Retries

- **Producer**: Configurable retries are available in `ProducerConfig`.
- **Consumer**: If a handler returns an error, it is logged, but the offset is not marked as processed by default (depending on your `MessageProcessor` or `MessageHandler` logic). The framework's default handler marks the offset as processed only if no error is returned.
