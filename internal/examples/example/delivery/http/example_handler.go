package http

import (
	"axiomod/internal/examples/example/usecase"
	"axiomod/internal/platform/observability"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ExampleHandler handles HTTP requests for the Example entity
type ExampleHandler struct {
	createUseCase *usecase.CreateExampleUseCase
	getUseCase    *usecase.GetExampleUseCase
	logger        *observability.Logger
}

// NewExampleHandler creates a new ExampleHandler
func NewExampleHandler(
	createUseCase *usecase.CreateExampleUseCase,
	getUseCase *usecase.GetExampleUseCase,
	logger *observability.Logger,
) *ExampleHandler {
	return &ExampleHandler{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		logger:        logger,
	}
}

// RegisterRoutes registers the routes for the ExampleHandler
func (h *ExampleHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/examples")
	group.Post("/", h.Create)
	group.Get("/:id", h.Get)
}

// Create handles the creation of a new Example
func (h *ExampleHandler) Create(c *fiber.Ctx) error {
	// Parse request body
	var input usecase.CreateExampleInput
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Execute use case
	output, err := h.createUseCase.Execute(c.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create example", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(output)
}

// Get handles the retrieval of an Example by ID
func (h *ExampleHandler) Get(c *fiber.Ctx) error {
	// Get ID from path parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	// Execute use case
	output, err := h.getUseCase.Execute(c.Context(), usecase.GetExampleInput{ID: id})
	if err != nil {
		h.logger.Error("Failed to get example", zap.Error(err), zap.String("id", id))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(output)
}
