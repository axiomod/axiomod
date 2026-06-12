package http

import (
	"errors"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
	"github.com/axiomod/axiomod/examples/example/usecase"
	"github.com/axiomod/axiomod/framework/observability"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ExampleHandler handles HTTP requests for the Example entity
type ExampleHandler struct {
	createUseCase *usecase.CreateExampleUseCase
	getUseCase    *usecase.GetExampleUseCase
	updateUseCase *usecase.UpdateExampleUseCase
	deleteUseCase *usecase.DeleteExampleUseCase
	listUseCase   *usecase.ListExamplesUseCase
	logger        *observability.Logger
}

// NewExampleHandler creates a new ExampleHandler
func NewExampleHandler(
	createUseCase *usecase.CreateExampleUseCase,
	getUseCase *usecase.GetExampleUseCase,
	updateUseCase *usecase.UpdateExampleUseCase,
	deleteUseCase *usecase.DeleteExampleUseCase,
	listUseCase *usecase.ListExamplesUseCase,
	logger *observability.Logger,
) *ExampleHandler {
	return &ExampleHandler{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
		listUseCase:   listUseCase,
		logger:        logger,
	}
}

// RegisterRoutes registers the routes for the ExampleHandler
func (h *ExampleHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/examples")
	group.Post("/", h.Create)
	group.Get("/", h.List)
	group.Get("/:id", h.Get)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

// Create handles the creation of a new Example
func (h *ExampleHandler) Create(c *fiber.Ctx) error {
	var input usecase.CreateExampleInput
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	output, err := h.createUseCase.Execute(c.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create example", zap.Error(err))
		return h.errorResponse(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(output)
}

// Get handles the retrieval of an Example by ID
func (h *ExampleHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	output, err := h.getUseCase.Execute(c.Context(), usecase.GetExampleInput{ID: id})
	if err != nil {
		h.logger.Error("Failed to get example", zap.Error(err), zap.String("id", id))
		return h.errorResponse(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(output)
}

// Update handles updating an existing Example
func (h *ExampleHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var input usecase.UpdateExampleInput
	if err := c.BodyParser(&input); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	input.ID = id

	output, err := h.updateUseCase.Execute(c.Context(), input)
	if err != nil {
		h.logger.Error("Failed to update example", zap.Error(err), zap.String("id", id))
		return h.errorResponse(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(output)
}

// Delete handles deleting an Example by ID
func (h *ExampleHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	output, err := h.deleteUseCase.Execute(c.Context(), usecase.DeleteExampleInput{ID: id})
	if err != nil {
		h.logger.Error("Failed to delete example", zap.Error(err), zap.String("id", id))
		return h.errorResponse(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(output)
}

// List handles listing Examples with optional filters
func (h *ExampleHandler) List(c *fiber.Ctx) error {
	input := usecase.ListExamplesInput{
		Name:      c.Query("name"),
		ValueType: c.Query("valueType"),
		Tag:       c.Query("tag"),
		Limit:     c.QueryInt("limit", 0),
		Offset:    c.QueryInt("offset", 0),
	}

	output, err := h.listUseCase.Execute(c.Context(), input)
	if err != nil {
		h.logger.Error("Failed to list examples", zap.Error(err))
		return h.errorResponse(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(output)
}

// errorResponse maps domain errors onto HTTP status codes: not-found to 404,
// validation failures to 400, everything else to 500.
func (h *ExampleHandler) errorResponse(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError

	var domainErr entity.DomainError
	switch {
	case errors.Is(err, repository.ErrExampleNotFound):
		status = fiber.StatusNotFound
	case errors.As(err, &domainErr):
		status = fiber.StatusBadRequest
	}

	return c.Status(status).JSON(fiber.Map{
		"error": err.Error(),
	})
}
