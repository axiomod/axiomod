package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	logger, _ := observability.NewLogger(&config.Config{})
	m := NewLoggingMiddleware(logger)

	app := fiber.New()
	app.Use(m.Handle())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret"
	jwtService := auth.NewJWTService(secret, time.Hour)
	logger, _ := observability.NewLogger(&config.Config{})
	m := NewAuthMiddleware(jwtService, logger)

	app := fiber.New()
	app.Use(m.Handle())
	app.Get("/me", func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("username").(string))
	})

	t.Run("Valid token", func(t *testing.T) {
		token, _ := jwtService.GenerateToken("123", "alice", "alice@example.com", nil)
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestTimeoutMiddleware(t *testing.T) {
	logger, _ := observability.NewLogger(&config.Config{})
	m := NewTimeoutMiddleware(10*time.Millisecond, logger)

	app := fiber.New()
	app.Get("/slow", m.Handle(), func(c *fiber.Ctx) error {
		select {
		case <-c.UserContext().Done():
			return nil
		case <-time.After(50 * time.Millisecond):
			return c.SendString("too slow")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	resp, _ := app.Test(req, 100)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
}
