# Error Handling

## Framework Errors (framework/errors)

Always use `framework/errors` for error creation and wrapping. **Never** use raw `fmt.Errorf` in framework or application code.

### Construction

```go
errors.New("something went wrong")                    // New error with stack trace
errors.Wrap(err, "additional context")                 // Wrap existing error
errors.WithCode(err, errors.CodeNotFound)              // Attach error code
errors.WithMetadata(err, "user_id", userID)            // Attach metadata
```

### Shorthand Constructors

```go
errors.NewNotFound(err, "user not found")
errors.NewInvalidInput(err, "email is required")
errors.NewUnauthorized(err, "invalid token")
errors.NewForbidden(err, "insufficient permissions")
errors.NewConflict(err, "resource already exists")
errors.NewInternal(err, "unexpected failure")
```

### Error Codes

`CodeNotFound`, `CodeInvalidInput`, `CodeUnauthorized`, `CodeForbidden`, `CodeInternal`, `CodeUnavailable`, `CodeTimeout`, `CodeAlreadyExists`, `CodeConflict`, `CodeNotImplemented`, `CodeValidation`, `CodeDeadlineExceeded`, `CodeCanceled`

### Protocol Mapping

- `errors.ToHTTPCode(err)` -- Maps to HTTP status codes (404, 400, 401, 403, 409, 408, 503, 500)
- `errors.ToGRPCCode(err)` -- Maps to gRPC status codes

### Introspection

```go
errors.GetCode(err)       // Extract code
errors.GetMetadata(err)   // Extract metadata
errors.GetStack(err)      // Extract stack trace
errors.Is(err, target)    // Standard unwrap check
errors.As(err, &target)   // Standard type assertion
```

## Domain Errors (entity-level)

Entities define their own domain errors using `DomainError`:

```go
type DomainError struct {
    Code    string
    Message string
}

var (
    ErrEmptyName = NewDomainError("example.empty_name", "Name cannot be empty")
)
```

Repository errors reuse the entity's `NewDomainError`:

```go
var (
    ErrExampleNotFound = entity.NewDomainError("repository.example_not_found", "Example not found")
)
```

## Rules

1. Framework/platform code: use `framework/errors` wrapping
2. Entity code: use `DomainError` for business rule violations
3. Repository code: use `DomainError` for data-layer errors
4. Infrastructure code: `fmt.Errorf("...: %w", err)` is acceptable for low-level wrapping
5. **Never** swallow errors silently -- always log or return them
6. **Never** expose internal error details to API consumers
