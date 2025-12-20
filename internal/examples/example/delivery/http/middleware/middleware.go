package middleware

import (
	"axiomod/internal/platform/observability"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// AuthMiddleware handles authentication for HTTP requests
type AuthMiddleware struct {
	logger *observability.Logger
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(logger *observability.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		logger: logger,
	}
}

// Handle returns a middleware function that handles authentication
func (m *AuthMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		token := c.Get("Authorization")
		if token == "" {
			m.logger.Info("Missing authorization token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization token",
			})
		}

		// In a real implementation, we would validate the token here
		// For example, using JWT validation or calling an auth service
		// For now, we'll just check if the token starts with "Bearer "
		if len(token) < 7 || token[:7] != "Bearer " {
			m.logger.Info("Invalid authorization token format")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization token format",
			})
		}

		// Set user ID in context for downstream handlers
		// In a real implementation, we would extract this from the token
		c.Locals("userID", "example-user-id")

		// Continue to the next middleware or handler
		return c.Next()
	}
}

// LoggingMiddleware handles logging for HTTP requests
type LoggingMiddleware struct {
	logger *observability.Logger
}

// NewLoggingMiddleware creates a new LoggingMiddleware
func NewLoggingMiddleware(logger *observability.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Handle returns a middleware function that handles logging
func (m *LoggingMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Log request details before processing
		m.logger.Info("Incoming request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.String("user-agent", c.Get("User-Agent")),
		)

		// Process request
		err := c.Next()

		// Log response details after processing
		m.logger.Info("Outgoing response",
			zap.Int("status", c.Response().StatusCode()),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("response-time", string(c.Response().Header.Peek("X-Response-Time"))),
		)

		return err
	}
}
