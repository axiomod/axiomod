# Testing Guide

This guide describes the testing patterns, best practices, and tools used in the Axiomod framework to maintain a high level of code quality and coverage.

## 1. Testing Philosophy

Axiomod prioritizes **automated verification** at every level of the stack.

- **Unit Tests**: Focus on individual functions and components.
- **Integration Tests**: Verify that multiple components (e.g., framework modules, plugins) work together correctly.
- **Contract Tests**: Ensure that HTTP and gRPC interfaces adhere to their specifications.

## 2. Table-Driven Tests

We use table-driven tests extensively for clarity and to cover multiple edge cases efficiently.

```go
func TestValidator(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"Valid email", "test@example.com", false},
        {"Invalid email", "not-an-email", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validate(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 3. Mocking and Dependencies

- Use `github.com/stretchr/testify/mock` or manual interfaces to mock external dependencies.
- For database tests, use `sqlmock` (if integrated) or a temporary test database.
- For OIDC/external APIs, use `httptest.NewServer` to provide a mock response.

## 4. Code Coverage

We target **>80% code coverage** for all core framework modules.

### Running Coverage Reports

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

## 5. Continuous Integration (CI)

The CI pipeline automatically runs all tests on every push and merge request.

- **Go Version Matrix**: Tests are executed against multiple supported Go versions (1.24, 1.25).
- **Coverage Gates**: PRs are blocked if coverage drops below the 80% threshold.
- **Static Analysis**: `go vet`, `staticcheck`, and `gosec` are run as part of the validation suite.

## 6. Project Integration Tests

Located in `tests/integration/`, these tests bootstrap the entire framework using `fx.App` to ensure that all modules are provided and dependencies are correctly resolved.

```go
func TestFullBootstrap(t *testing.T) {
    app := fx.New(
        // ... modules ...
        fx.NopLogger,
    )
    // app.Start/Stop
}
```
