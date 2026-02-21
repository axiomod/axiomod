# Resilience Patterns

## Circuit Breaker

`framework/circuitbreaker/circuitbreaker.go`:

```go
cb := circuitbreaker.New(circuitbreaker.Options{
    MaxFailures:  5,
    ResetTimeout: 30 * time.Second,
    HalfOpenLimit: 1,
})

err := cb.Execute(func() error {
    return callExternalService()
})
```

States: `StateClosed` -> `StateOpen` -> `StateHalfOpen` -> `StateClosed`

Thread-safe with `sync.RWMutex`. Manual controls: `AllowRequest()`, `RecordResult(err)`, `Reset()`.

## Resilience Wrapper

`framework/resilience/resilience.go` combines multiple patterns:

```go
r := resilience.New(resilience.Options{
    CircuitBreaker: cb,
    MaxRetries:     3,
    RetryDelay:     100 * time.Millisecond,
    BackoffFactor:  2.0,
    MaxDelay:       5 * time.Second,
    Timeout:        10 * time.Second,
    Fallback:       func() (interface{}, error) { return cachedResult, nil },
    IsRetryable:    func(err error) bool { return !errors.Is(err, ErrPermanent) },
})

result, err := r.Execute(ctx, func() (interface{}, error) {
    return callExternalAPI()
})
```

Orchestrates: circuit breaker + retry + timeout + fallback.

## HTTP Client

`framework/client/http_client.go` with built-in resilience:

```go
client := client.New(client.Options{
    CircuitBreaker: cb,
    MaxRetries:     3,
})

resp, err := client.GetJSON(ctx, url, &result)
```

## Rules

1. Wrap all external service calls with circuit breaker
2. Use exponential backoff for retries
3. Always set timeouts on external calls
4. Provide fallback values for degraded operation
5. Filter retryable vs permanent errors
