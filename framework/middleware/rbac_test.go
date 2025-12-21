package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware(t *testing.T) {
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "model.conf")
	policyPath := filepath.Join(tempDir, "policy.csv")

	modelContent := `
[request_definition]
r = sub, obj, act
[policy_definition]
p = sub, obj, act
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	policyContent := "p, alice, data1, read\n"
	_ = os.WriteFile(modelPath, []byte(modelContent), 0644)
	_ = os.WriteFile(policyPath, []byte(policyContent), 0644)

	rbacService, _ := auth.NewRBACService(auth.CasbinConfig{
		ModelPath:  modelPath,
		PolicyPath: policyPath,
	})
	logger, _ := observability.NewLogger(&config.Config{})
	middleware := NewRBACMiddleware(rbacService, logger)

	app := fiber.New()
	app.Get("/data1", func(c *fiber.Ctx) error {
		c.Locals("username", "alice")
		return c.Next()
	}, middleware.Handle("data1", "read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Get("/data2", func(c *fiber.Ctx) error {
		c.Locals("username", "alice")
		return c.Next()
	}, middleware.Handle("data2", "read"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	t.Run("Allowed access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/data1", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Denied access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/data2", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
