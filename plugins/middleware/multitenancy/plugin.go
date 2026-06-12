// Package multitenancy provides a plugin that resolves the tenant for each
// HTTP request from a configurable header and exposes it to handlers.
package multitenancy

import (
	"fmt"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// DefaultTenantHeader is the header used when none is configured.
const DefaultTenantHeader = "X-Tenant-ID"

// LocalsKey is the Fiber locals key under which the tenant ID is stored.
const LocalsKey = "tenant_id"

// Plugin resolves the tenant for incoming requests from a request header.
type Plugin struct {
	logger   *observability.Logger
	header   string
	required bool
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "multitenancy"
}

// Initialize configures the plugin from its settings.
//
// Settings:
//   - header (string): header carrying the tenant ID, default "X-Tenant-ID"
//   - required (bool): reject requests without a tenant header, default false
func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.logger = logger
	p.header = DefaultTenantHeader

	if header, ok := settings["header"].(string); ok && header != "" {
		p.header = header
	}
	if required, ok := settings["required"].(bool); ok {
		p.required = required
	}

	return nil
}

// Start starts the plugin.
func (p *Plugin) Start() error {
	if p.header == "" {
		p.header = DefaultTenantHeader
	}
	if p.logger != nil {
		p.logger.Info("Multitenancy plugin started",
			zap.String("header", p.header),
			zap.Bool("required", p.required),
		)
	}
	return nil
}

// Stop stops the plugin.
func (p *Plugin) Stop() error {
	return nil
}

// Header returns the configured tenant header name.
func (p *Plugin) Header() string {
	return p.header
}

// Middleware returns a Fiber handler that extracts the tenant ID from the
// configured header and stores it in the request locals under LocalsKey.
// When the plugin is configured as required, requests without a tenant
// header are rejected with 400.
func (p *Plugin) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Get(p.header)
		if tenantID == "" && p.required {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("missing required tenant header %s", p.header),
			})
		}

		if tenantID != "" {
			c.Locals(LocalsKey, tenantID)
		}

		return c.Next()
	}
}

// TenantID returns the tenant ID resolved for the request, or an empty
// string when the request carries no tenant.
func TenantID(c *fiber.Ctx) string {
	if tenantID, ok := c.Locals(LocalsKey).(string); ok {
		return tenantID
	}
	return ""
}
