package resilience

import (
	"context"
	"errors"
	"time"

	"axiomod/internal/framework/circuitbreaker"
)

// Common errors
var (
	ErrTimeout      = errors.New("operation timed out")
	ErrCircuitOpen  = errors.New("circuit breaker is open")
	ErrRetryFailed  = errors.New("all retries failed")
	ErrFallbackFail = errors.New("fallback failed")
)

// RetryOptions contains options for retry
type RetryOptions struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	// BackoffFactor is the factor by which to increase the delay after each retry
	BackoffFactor float64
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// RetryableErrors is a list of errors that should trigger a retry
	RetryableErrors []error
}

// DefaultRetryOptions returns the default retry options
func DefaultRetryOptions() *RetryOptions {
	return &RetryOptions{
		MaxRetries:      3,
		RetryDelay:      100 * time.Millisecond,
		BackoffFactor:   2.0,
		MaxDelay:        30 * time.Second,
		RetryableErrors: []error{},
	}
}

// TimeoutOptions contains options for timeout
type TimeoutOptions struct {
	// Timeout is the timeout duration
	Timeout time.Duration
}

// DefaultTimeoutOptions returns the default timeout options
func DefaultTimeoutOptions() *TimeoutOptions {
	return &TimeoutOptions{
		Timeout: 30 * time.Second,
	}
}

// FallbackOptions contains options for fallback
type FallbackOptions struct {
	// FallbackFunc is the function to call when the operation fails
	FallbackFunc func(ctx context.Context, err error) (interface{}, error)
}

// DefaultFallbackOptions returns the default fallback options
func DefaultFallbackOptions() *FallbackOptions {
	return &FallbackOptions{
		FallbackFunc: nil,
	}
}

// ResilienceOptions contains options for resilience
type ResilienceOptions struct {
	// Retry contains retry options
	Retry *RetryOptions
	// Timeout contains timeout options
	Timeout *TimeoutOptions
	// CircuitBreaker contains circuit breaker options
	CircuitBreaker *circuitbreaker.Options
	// Fallback contains fallback options
	Fallback *FallbackOptions
}

// DefaultResilienceOptions returns the default resilience options
func DefaultResilienceOptions() *ResilienceOptions {
	cbOpts := circuitbreaker.DefaultOptions()
	return &ResilienceOptions{
		Retry:          DefaultRetryOptions(),
		Timeout:        DefaultTimeoutOptions(),
		CircuitBreaker: &cbOpts,
		Fallback:       DefaultFallbackOptions(),
	}
}

// Resilience provides resilience patterns
type Resilience struct {
	options        *ResilienceOptions
	circuitBreaker *circuitbreaker.CircuitBreaker
}

// New creates a new Resilience instance
func New(options *ResilienceOptions) *Resilience {
	if options == nil {
		options = DefaultResilienceOptions()
	}

	return &Resilience{
		options:        options,
		circuitBreaker: circuitbreaker.New(*options.CircuitBreaker),
	}
}

// Execute executes a function with resilience patterns
func (r *Resilience) Execute(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// Apply circuit breaker
	if !r.circuitBreaker.AllowRequest() {
		if r.options.Fallback != nil && r.options.Fallback.FallbackFunc != nil {
			return r.options.Fallback.FallbackFunc(ctx, ErrCircuitOpen)
		}
		return nil, ErrCircuitOpen
	}

	// Apply timeout
	var timeoutCtx context.Context
	var cancel context.CancelFunc
	if r.options.Timeout != nil && r.options.Timeout.Timeout > 0 {
		timeoutCtx, cancel = context.WithTimeout(ctx, r.options.Timeout.Timeout)
		defer cancel()
	} else {
		timeoutCtx = ctx
	}

	// Apply retry
	var result interface{}
	var err error
	var retryCount int
	var delay time.Duration

	if r.options.Retry != nil {
		delay = r.options.Retry.RetryDelay
	}

	for {
		// Execute function
		result, err = fn(timeoutCtx)

		// Record result in circuit breaker
		r.circuitBreaker.RecordResult(err)

		// Check if successful
		if err == nil {
			return result, nil
		}

		// Check if context is done
		if timeoutCtx.Err() != nil {
			if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
				err = ErrTimeout
			}
			break
		}

		// Check if retry is enabled
		if r.options.Retry == nil || retryCount >= r.options.Retry.MaxRetries {
			break
		}

		// Check if error is retryable
		if !isRetryableError(err, r.options.Retry.RetryableErrors) {
			break
		}

		// Wait before retrying
		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-timeoutCtx.Done():
			if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
				err = ErrTimeout
			}
			break
		}

		// Increase retry count
		retryCount++

		// Increase delay for next retry
		if r.options.Retry.BackoffFactor > 1.0 {
			delay = time.Duration(float64(delay) * r.options.Retry.BackoffFactor)
			if delay > r.options.Retry.MaxDelay {
				delay = r.options.Retry.MaxDelay
			}
		}
	}

	// Apply fallback
	if r.options.Fallback != nil && r.options.Fallback.FallbackFunc != nil {
		result, err = r.options.Fallback.FallbackFunc(ctx, err)
		if err != nil {
			return nil, errors.Join(ErrFallbackFail, err)
		}
		return result, nil
	}

	// Return error
	if r.options.Retry != nil && r.options.Retry.MaxRetries > 0 {
		return nil, errors.Join(ErrRetryFailed, err)
	}
	return nil, err
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error, retryableErrors []error) bool {
	// If no retryable errors are specified, all errors are retryable
	if len(retryableErrors) == 0 {
		return true
	}

	// Check if the error is in the list of retryable errors
	for _, retryableErr := range retryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}

// Reset resets the circuit breaker
func (r *Resilience) Reset() {
	r.circuitBreaker.Reset()
}

// GetCircuitBreaker returns the circuit breaker
func (r *Resilience) GetCircuitBreaker() *circuitbreaker.CircuitBreaker {
	return r.circuitBreaker
}

// GetOptions returns the resilience options
func (r *Resilience) GetOptions() *ResilienceOptions {
	return r.options
}
