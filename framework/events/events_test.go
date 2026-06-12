package events

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newTestBus returns an event bus with a no-op logger.
func newTestBus() *EventBus {
	return NewEventBus(&observability.Logger{Logger: zap.NewNop()})
}

// waitForEvent receives an event from the channel or fails the test after a
// generous timeout.
func waitForEvent(t *testing.T, ch <-chan Event) Event {
	t.Helper()
	select {
	case event := <-ch:
		return event
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
		return Event{}
	}
}

func TestNewEventBus(t *testing.T) {
	bus := newTestBus()
	require.NotNil(t, bus)
	assert.NotNil(t, bus.subscribers)
}

func TestPublish_Validation(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		payload  []byte
		expected error
	}{
		{"empty topic", "", []byte("data"), ErrTopicEmpty},
		{"nil payload", "topic", nil, ErrPayloadEmpty},
		{"empty payload", "topic", []byte{}, ErrPayloadEmpty},
	}

	bus := newTestBus()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bus.Publish(context.Background(), tt.topic, tt.payload, nil)
			assert.ErrorIs(t, err, tt.expected)
		})
	}
}

func TestPublish_NoSubscribers(t *testing.T) {
	bus := newTestBus()
	err := bus.Publish(context.Background(), "lonely.topic", []byte(`{"a":1}`), nil)
	assert.NoError(t, err)
}

func TestSubscribePublish(t *testing.T) {
	bus := newTestBus()
	received := make(chan Event, 1)

	err := bus.Subscribe(context.Background(), []string{"user.created"}, func(ctx context.Context, event Event) error {
		received <- event
		return nil
	})
	require.NoError(t, err)

	headers := map[string]string{"source": "test"}
	require.NoError(t, bus.Publish(context.Background(), "user.created", []byte(`{"id":"42"}`), headers))

	event := waitForEvent(t, received)
	assert.Equal(t, "user.created", event.Topic)
	assert.JSONEq(t, `{"id":"42"}`, string(event.Payload))
	assert.Equal(t, headers, event.Headers)
	assert.NotEmpty(t, event.ID)
	assert.False(t, event.Timestamp.IsZero())
}

func TestSubscribe_MultipleHandlersReceiveEvent(t *testing.T) {
	bus := newTestBus()
	first := make(chan Event, 1)
	second := make(chan Event, 1)

	require.NoError(t, bus.Subscribe(context.Background(), []string{"order.placed"}, func(ctx context.Context, event Event) error {
		first <- event
		return nil
	}))
	require.NoError(t, bus.Subscribe(context.Background(), []string{"order.placed"}, func(ctx context.Context, event Event) error {
		second <- event
		return nil
	}))

	require.NoError(t, bus.Publish(context.Background(), "order.placed", []byte(`{"order":"1"}`), nil))

	assert.Equal(t, "order.placed", waitForEvent(t, first).Topic)
	assert.Equal(t, "order.placed", waitForEvent(t, second).Topic)
}

func TestSubscribe_MultipleTopics(t *testing.T) {
	bus := newTestBus()
	received := make(chan Event, 2)

	require.NoError(t, bus.Subscribe(context.Background(), []string{"topic.a", "topic.b"}, func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}))

	require.NoError(t, bus.Publish(context.Background(), "topic.a", []byte(`{"n":1}`), nil))
	require.NoError(t, bus.Publish(context.Background(), "topic.b", []byte(`{"n":2}`), nil))

	topics := map[string]bool{}
	topics[waitForEvent(t, received).Topic] = true
	topics[waitForEvent(t, received).Topic] = true
	assert.True(t, topics["topic.a"])
	assert.True(t, topics["topic.b"])
}

func TestSubscribe_TopicIsolation(t *testing.T) {
	bus := newTestBus()
	var count int32

	require.NoError(t, bus.Subscribe(context.Background(), []string{"topic.a"}, func(ctx context.Context, event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}))

	// Publish to a different topic; the handler must not fire.
	require.NoError(t, bus.Publish(context.Background(), "topic.b", []byte(`{"n":1}`), nil))

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(&count))
}

func TestPublish_HandlerErrorDoesNotFailPublish(t *testing.T) {
	bus := newTestBus()
	invoked := make(chan struct{}, 1)

	require.NoError(t, bus.Subscribe(context.Background(), []string{"failing.topic"}, func(ctx context.Context, event Event) error {
		invoked <- struct{}{}
		return errors.New("handler failure")
	}))

	err := bus.Publish(context.Background(), "failing.topic", []byte(`{"n":1}`), nil)
	assert.NoError(t, err, "publish must not propagate handler errors")

	select {
	case <-invoked:
	case <-time.After(5 * time.Second):
		t.Fatal("handler was never invoked")
	}
}

func TestClose_RemovesSubscribers(t *testing.T) {
	bus := newTestBus()
	var count int32

	require.NoError(t, bus.Subscribe(context.Background(), []string{"topic.x"}, func(ctx context.Context, event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}))

	require.NoError(t, bus.Close())

	require.NoError(t, bus.Publish(context.Background(), "topic.x", []byte(`{"n":1}`), nil))
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(&count), "handlers must not fire after Close")
}

func TestPublish_Concurrent(t *testing.T) {
	bus := newTestBus()
	var count int32

	require.NoError(t, bus.Subscribe(context.Background(), []string{"concurrent.topic"}, func(ctx context.Context, event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}))

	const publishers = 8
	const eventsPerPublisher = 25

	for p := 0; p < publishers; p++ {
		go func() {
			for i := 0; i < eventsPerPublisher; i++ {
				assert.NoError(t, bus.Publish(context.Background(), "concurrent.topic", []byte(`{"n":1}`), nil))
			}
		}()
	}

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&count) == publishers*eventsPerPublisher
	}, 10*time.Second, 10*time.Millisecond, "all concurrently published events must be handled")
}

func TestConcurrentSubscribeAndPublish(t *testing.T) {
	bus := newTestBus()
	done := make(chan struct{})

	// Subscribe concurrently with publishes to exercise the lock paths.
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			assert.NoError(t, bus.Subscribe(context.Background(), []string{"mixed.topic"}, func(ctx context.Context, event Event) error {
				return nil
			}))
		}
	}()

	for i := 0; i < 50; i++ {
		require.NoError(t, bus.Publish(context.Background(), "mixed.topic", []byte(`{"n":1}`), nil))
	}

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent subscribe did not finish")
	}
}
