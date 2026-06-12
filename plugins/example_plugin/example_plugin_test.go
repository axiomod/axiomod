package example_plugin

import (
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/plugins"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// The reference plugin must always satisfy the real Plugin interface —
// the plugin-development guide is generated from this code.
var _ plugins.Plugin = (*ExamplePlugin)(nil)

func TestExamplePluginLifecycle(t *testing.T) {
	p := &ExamplePlugin{}
	logger := &observability.Logger{Logger: zap.NewNop()}
	h := health.New(logger)

	settings := map[string]interface{}{"greeting": "Hi"}
	require.NoError(t, p.Initialize(settings, logger, nil, &config.Config{}, h))
	assert.Equal(t, "example", p.Name())
	assert.False(t, p.IsActive())

	// The health check reports DOWN before Start and UP after.
	require.NoError(t, p.Start())
	assert.True(t, p.IsActive())

	require.NoError(t, p.Stop())
	assert.False(t, p.IsActive())
}
