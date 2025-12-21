package middleware

import (
	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RBACMiddleware checks if the user has permission to access a resource
type RBACMiddleware struct {
	rbacService *auth.RBACService
	logger      *observability.Logger
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(rbacService *auth.RBACService, logger *observability.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		rbacService: rbacService,
		logger:      logger,
	}
}

// Handle returns a Fiber middleware handler that enforces RBAC
func (m *RBACMiddleware) Handle(obj, act string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get subject from context (stored by AuthMiddleware)
		sub, ok := c.Locals("username").(string)
		if !ok || sub == "" {
			// fallback to user_id if username is not set
			sub, ok = c.Locals("user_id").(string)
			if !ok || sub == "" {
				m.logger.Warn("No subject found in context")
				return fiber.NewError(fiber.StatusForbidden, "access denied")
			}
		}

		// Enforce policy
		allowed, err := m.rbacService.Enforce(sub, obj, act)
		if err != nil {
			m.logger.Error("RBAC enforcement error", zap.Error(err))
			return fiber.NewError(fiber.StatusInternalServerError, "authorization error")
		}

		if !allowed {
			m.logger.Warn("User not authorized",
				zap.String("subject", sub),
				zap.String("object", obj),
				zap.String("action", act),
			)
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}

		return c.Next()
	}
}
