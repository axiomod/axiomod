package multitenancy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/observability"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPlugin(t *testing.T, settings map[string]interface{}) *Plugin {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)

	p := &Plugin{}
	require.NoError(t, p.Initialize(settings, logger, nil, nil, nil))
	return p
}

func TestPlugin_Lifecycle(t *testing.T) {
	p := newTestPlugin(t, map[string]interface{}{})

	assert.Equal(t, "multitenancy", p.Name())
	assert.NoError(t, p.Start())
	assert.NoError(t, p.Stop())
	assert.Equal(t, DefaultTenantHeader, p.Header())
}

func TestPlugin_InitializeSettings(t *testing.T) {
	tests := []struct {
		name           string
		settings       map[string]interface{}
		expectedHeader string
		required       bool
	}{
		{"defaults", map[string]interface{}{}, DefaultTenantHeader, false},
		{"custom header", map[string]interface{}{"header": "X-Org-ID"}, "X-Org-ID", false},
		{"required", map[string]interface{}{"required": true}, DefaultTenantHeader, true},
		{"empty header falls back", map[string]interface{}{"header": ""}, DefaultTenantHeader, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestPlugin(t, tt.settings)
			assert.Equal(t, tt.expectedHeader, p.Header())
			assert.Equal(t, tt.required, p.required)
		})
	}
}

func TestPlugin_Middleware(t *testing.T) {
	tests := []struct {
		name           string
		settings       map[string]interface{}
		header         map[string]string
		expectedStatus int
		expectedTenant string
	}{
		{
			name:           "tenant resolved from default header",
			settings:       map[string]interface{}{},
			header:         map[string]string{DefaultTenantHeader: "acme"},
			expectedStatus: http.StatusOK,
			expectedTenant: "acme",
		},
		{
			name:           "missing tenant allowed when not required",
			settings:       map[string]interface{}{},
			header:         nil,
			expectedStatus: http.StatusOK,
			expectedTenant: "",
		},
		{
			name:           "missing tenant rejected when required",
			settings:       map[string]interface{}{"required": true},
			header:         nil,
			expectedStatus: http.StatusBadRequest,
			expectedTenant: "",
		},
		{
			name:           "custom header",
			settings:       map[string]interface{}{"header": "X-Org-ID"},
			header:         map[string]string{"X-Org-ID": "globex"},
			expectedStatus: http.StatusOK,
			expectedTenant: "globex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestPlugin(t, tt.settings)

			var gotTenant string
			app := fiber.New()
			app.Use(p.Middleware())
			app.Get("/", func(c *fiber.Ctx) error {
				gotTenant = TenantID(c)
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tt.header {
				req.Header.Set(k, v)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedTenant, gotTenant)
			}
		})
	}
}
