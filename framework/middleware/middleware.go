package middleware

import (
	"context"
	"time"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the fx options for the middleware module
var Module = fx.Options(
	fx.Provide(NewLoggingMiddleware),
	fx.Provide(NewAuthMiddleware),
	fx.Provide(NewRoleMiddleware),
	fx.Provide(NewTimeoutMiddleware),
	fx.Provide(NewRecoveryMiddleware),
	fx.Provide(NewMetricsMiddleware),
	fx.Provide(NewTracingMiddleware),
)

// LoggingMiddleware logs HTTP requests
type LoggingMiddleware struct {
	logger *observability.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *observability.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Handle returns a Fiber middleware handler
func (m *LoggingMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Get request details
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		// Process request
		err := c.Next()

		// Get response details
		status := c.Response().StatusCode()
		latency := time.Since(start)

		// Log request
		m.logger.Info("HTTP request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", ip),
			zap.String("user_agent", userAgent),
		)

		return err
	}
}

// AuthMiddleware authenticates HTTP requests
type AuthMiddleware struct {
	jwtService *auth.JWTService
	logger     *observability.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService *auth.JWTService, logger *observability.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

// Handle returns a Fiber middleware handler
func (m *AuthMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header
		token := c.Get("Authorization")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Validate token
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			m.logger.Warn("Invalid token", zap.Error(err))
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}

// RoleMiddleware checks if the user has the required role
type RoleMiddleware struct {
	logger *observability.Logger
}

// NewRoleMiddleware creates a new role middleware
func NewRoleMiddleware(logger *observability.Logger) *RoleMiddleware {
	return &RoleMiddleware{
		logger: logger,
	}
}

// RequireRole returns a Fiber middleware handler that requires a specific role
func (m *RoleMiddleware) RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get roles from context
		roles, ok := c.Locals("roles").([]string)
		if !ok {
			m.logger.Warn("No roles found in context")
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}

		// Check if user has the required role
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}

		m.logger.Warn("User does not have the required role",
			zap.String("required_role", role),
			zap.Strings("user_roles", roles),
		)
		return fiber.NewError(fiber.StatusForbidden, "access denied")
	}
}

// TimeoutMiddleware adds a timeout to HTTP requests
type TimeoutMiddleware struct {
	timeout time.Duration
	logger  *observability.Logger
}

// NewTimeoutMiddleware creates a new timeout middleware
func NewTimeoutMiddleware(timeout time.Duration, logger *observability.Logger) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		timeout: timeout,
		logger:  logger,
	}
}

// Handle returns a Fiber middleware handler
func (m *TimeoutMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Context(), m.timeout)
		defer cancel()

		// Store the context in Fiber context
		c.SetUserContext(ctx)

		// Process request
		err := c.Next()

		// Check if context timed out
		if ctx.Err() == context.DeadlineExceeded {
			m.logger.Warn("Request timed out",
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Duration("timeout", m.timeout),
			)
			return fiber.NewError(fiber.StatusRequestTimeout, "request timed out")
		}

		return err
	}
}

// RecoveryMiddleware recovers from panics
type RecoveryMiddleware struct {
	logger *observability.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger *observability.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// Handle returns a Fiber middleware handler
func (m *RecoveryMiddleware) Handle() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				m.logger.Error("Recovered from panic",
					zap.Any("panic", r),
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
				)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "internal server error",
				})
			}
		}()

		return c.Next()
	}
}
