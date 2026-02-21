# HTTP Delivery (Fiber)

## Handler Structure

```go
type ExampleHandler struct {
    createUseCase *usecase.CreateExampleUseCase
    getUseCase    *usecase.GetExampleUseCase
    logger        *observability.Logger
}

func NewExampleHandler(
    createUseCase *usecase.CreateExampleUseCase,
    getUseCase    *usecase.GetExampleUseCase,
    logger        *observability.Logger,
) *ExampleHandler {
    return &ExampleHandler{
        createUseCase: createUseCase,
        getUseCase:    getUseCase,
        logger:        logger,
    }
}
```

## Route Registration

Handlers expose `RegisterRoutes(router fiber.Router)`:

```go
func (h *ExampleHandler) RegisterRoutes(router fiber.Router) {
    group := router.Group("/examples")
    group.Post("/", h.Create)
    group.Get("/:id", h.Get)
}
```

Registered via `fx.Invoke`:

```go
func registerHTTPRoutes(app *fiber.App, handler *http.ExampleHandler,
    loggingMw *middleware.LoggingMiddleware, authMw *middleware.AuthMiddleware) {
    api := app.Group("/api/v1")
    api.Use(loggingMw.Handle())
    api.Use(authMw.Handle())
    handler.RegisterRoutes(api)
}
```

## Handler Method Pattern

```go
func (h *ExampleHandler) Create(c *fiber.Ctx) error {
    // 1. Parse request
    var input usecase.CreateExampleInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    // 2. Delegate to use case (NO business logic here)
    output, err := h.createUseCase.Execute(c.Context(), input)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    // 3. Return response
    return c.Status(fiber.StatusCreated).JSON(output)
}
```

## Error Response Format

```go
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
    "error": "Invalid request body",
})
```

## Health/Infra Endpoints (built-in)

- `/live` -- Liveness probe
- `/ready` -- Readiness probe
- `/health` -- Legacy health check
- `/metrics` -- Prometheus metrics

## Rules

1. **No business logic in handlers** -- delegate to use cases
2. Handlers parse request, call use case, format response
3. Use `c.Context()` to propagate context to use cases
4. Route groups under `/api/v1` (or versioned prefix)
5. Apply middleware at the group level, not per-route
