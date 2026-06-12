package kafka

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(t *testing.T) *observability.Logger {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)
	return logger
}

func TestPlugin_Name(t *testing.T) {
	p := &Plugin{}
	assert.Equal(t, "kafka", p.Name())
}

func TestPlugin_InitializeSettings(t *testing.T) {
	tests := []struct {
		name            string
		settings        map[string]interface{}
		wantErr         bool
		expectedBrokers []string
		expectedTimeout time.Duration
	}{
		{
			name:            "defaults",
			settings:        map[string]interface{}{},
			expectedBrokers: []string{"localhost:9092"},
			expectedTimeout: 10 * time.Second,
		},
		{
			name:            "string slice brokers",
			settings:        map[string]interface{}{"brokers": []string{"a:9092", "b:9092"}},
			expectedBrokers: []string{"a:9092", "b:9092"},
			expectedTimeout: 10 * time.Second,
		},
		{
			name:            "interface slice brokers (YAML decoding)",
			settings:        map[string]interface{}{"brokers": []interface{}{"a:9092"}},
			expectedBrokers: []string{"a:9092"},
			expectedTimeout: 10 * time.Second,
		},
		{
			name:            "custom timeout",
			settings:        map[string]interface{}{"timeout": "3s"},
			expectedBrokers: []string{"localhost:9092"},
			expectedTimeout: 3 * time.Second,
		},
		{
			name:     "invalid timeout",
			settings: map[string]interface{}{"timeout": "soon"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{}
			err := p.Initialize(tt.settings, newTestLogger(t), nil, nil, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBrokers, p.producerConfig.Brokers)
			assert.Equal(t, tt.expectedTimeout, p.producerConfig.Timeout)
		})
	}
}

func TestPlugin_RegistersHealthCheck(t *testing.T) {
	logger := newTestLogger(t)
	h := health.New(logger)

	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{}, logger, nil, nil, h))

	h.RunChecks()
	resp := h.GetResponse()
	component, ok := resp.Components["kafka"]
	require.True(t, ok, "kafka health check must be registered")
	assert.Equal(t, health.StatusDown, component.Status)
}

func TestPlugin_PublishBeforeStart(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{}, newTestLogger(t), nil, nil, nil))

	err := p.Publish(context.Background(), "topic", "key", []byte("value"))
	assert.ErrorContains(t, err, "not started")
	assert.NoError(t, p.Stop())
}

func TestToStringSlice(t *testing.T) {
	assert.Equal(t, []string{"a"}, toStringSlice([]string{"a"}))
	assert.Equal(t, []string{"a", "b"}, toStringSlice([]interface{}{"a", "b"}))
	assert.Empty(t, toStringSlice([]interface{}{1, 2}))
	assert.Nil(t, toStringSlice("not-a-slice"))
	assert.Nil(t, toStringSlice(nil))
}

func TestPlugin_Integration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test; set RUN_INTEGRATION_TESTS=true")
	}

	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{}, newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())
	defer func() { assert.NoError(t, p.Stop()) }()

	assert.NotNil(t, p.Producer())
	assert.NoError(t, p.Publish(context.Background(), "axiomod-test", "k", []byte("v")))
}
