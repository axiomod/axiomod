# Testing Conventions

## Libraries

- `github.com/stretchr/testify/assert` -- standard assertions
- `github.com/stretchr/testify/require` -- fatal assertions (stop test on failure)
- `go.uber.org/fx/fxtest` -- fx DI integration tests
- `net/http/httptest` -- HTTP testing

## Test File Location

Test files are co-located with source code (`*_test.go` next to the `.go` file).

## Table-Driven Tests (preferred pattern)

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
        wantErr  bool
    }{
        {"valid input", "hello", 5, false},
        {"empty input", "", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := DoSomething(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Mock Pattern (manual interface mocks)

```go
type mockPlugin struct {
    name        string
    initialized bool
    started     bool
}

func (m *mockPlugin) Name() string             { return m.name }
func (m *mockPlugin) Initialize(...) error     { m.initialized = true; return nil }
func (m *mockPlugin) Start() error             { m.started = true; return nil }
func (m *mockPlugin) Stop() error              { return nil }
```

No code-gen mocking tools (gomock, mockery). Use manual interface implementations.

## fx Integration Tests

```go
testApp := fxtest.New(t,
    fx.Provide(func() *config.Config { return &config.Config{...} }),
    fx.Provide(observability.NewLogger),
    fx.Invoke(func(logger *observability.Logger) {
        assert.NotNil(t, logger)
    }),
    fx.StartTimeout(5*time.Second),
)
err := testApp.Start(ctx)
assert.NoError(t, err)
defer testApp.Stop(ctx)
```

## HTTP Handler Tests (Fiber)

```go
app := fiber.New()
app.Use(middleware.Handle())
app.Get("/path", handler)

req := httptest.NewRequest(http.MethodGet, "/path", nil)
resp, _ := app.Test(req)
assert.Equal(t, http.StatusOK, resp.StatusCode)
```

## Test Helpers

```go
func createTestConfigFile(t *testing.T, content string) string {
    t.Helper()
    dir := t.TempDir()
    path := filepath.Join(dir, "test_config.yaml")
    err := os.WriteFile(path, []byte(content), 0644)
    assert.NoError(t, err)
    return path
}
```

## Integration Tests

Gate behind environment variable:

```go
if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
    t.Skip("Skipping integration test; set RUN_INTEGRATION_TESTS=true")
}
```

## Benchmarks

```go
func BenchmarkNew(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = New("test error")
    }
}
```

## Coverage Target

- **>80%** for core framework modules
- Race detector: always use `-race` in CI (`go test -v -race ./...`)

## Commands

```
make test                              # go test -v ./...
go test -v ./framework/errors/...      # Single package
go test -race ./...                    # With race detector
go test -bench=. ./framework/errors/   # Benchmarks
```
