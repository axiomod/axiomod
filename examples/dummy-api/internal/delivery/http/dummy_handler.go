package http

import (
	"net/http"

	"github.com/axiomod/axiomod/platform/observability"
	"github.com/gofiber/fiber/v2"
)

type DummyHandler struct {
	logger *observability.Logger
}

func NewDummyHandler(logger *observability.Logger) *DummyHandler {
	return &DummyHandler{
		logger: logger,
	}
}

func (h *DummyHandler) RegisterRoutes(app *fiber.App) {
	h.logger.Info("Registering dummy API routes")

	app.Get("/info", h.GetInfo)
	app.Get("/users", h.GetUsers)
	app.Get("/products", h.GetProducts)
}

func (h *DummyHandler) GetInfo(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Axiomod Dummy API",
		"version": "1.0.0",
		"status":  "up",
	})
}

func (h *DummyHandler) GetUsers(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"users": []fiber.Map{
			{"id": 1, "name": "User 1", "email": "user1@example.com"},
			{"id": 2, "name": "User 2", "email": "user2@example.com"},
		},
	})
}

func (h *DummyHandler) GetProducts(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"products": []fiber.Map{
			{"id": "p1", "name": "Dummy Product A", "price": 99.99},
			{"id": "p2", "name": "Dummy Product B", "price": 49.50},
		},
	})
}
