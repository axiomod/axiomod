package redis

import (
	"net"
	"os"
	"testing"

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
	assert.Equal(t, "redis", p.Name())
}

func TestPlugin_InitializeSettings(t *testing.T) {
	tests := []struct {
		name         string
		settings     map[string]interface{}
		expectedAddr string
		expectedDB   int
	}{
		{"defaults", map[string]interface{}{}, defaultAddr, 0},
		{"custom addr", map[string]interface{}{"addr": "redis.internal:6380"}, "redis.internal:6380", 0},
		{"custom db", map[string]interface{}{"db": 3}, defaultAddr, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{}
			require.NoError(t, p.Initialize(tt.settings, newTestLogger(t), nil, nil, nil))
			assert.Equal(t, tt.expectedAddr, p.addr)
			assert.Equal(t, tt.expectedDB, p.db)
		})
	}
}

func TestPlugin_RegistersHealthCheck(t *testing.T) {
	logger := newTestLogger(t)
	h := health.New(logger)

	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{}, logger, nil, nil, h))

	// The check is registered but the plugin is not started, so it reports
	// an error rather than panicking.
	h.RunChecks()
	resp := h.GetResponse()
	component, ok := resp.Components["redis"]
	require.True(t, ok, "redis health check must be registered")
	assert.Equal(t, health.StatusDown, component.Status)
}

func TestPlugin_StartFailsWhenServerUnreachable(t *testing.T) {
	// Reserve a port and close it so nothing is listening there.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{"addr": addr}, newTestLogger(t), nil, nil, nil))

	err = p.Start()
	assert.Error(t, err)
	assert.NoError(t, p.Stop())
}

func TestPlugin_Integration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test; set RUN_INTEGRATION_TESTS=true")
	}

	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{}, newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())
	defer func() { assert.NoError(t, p.Stop()) }()

	assert.NotNil(t, p.Client())
	assert.NoError(t, p.ping())
}
