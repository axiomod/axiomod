package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/circuitbreaker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTransient = errors.New("transient failure")

// testOptions returns resilience options with fast timings suitable for tests.
func testOptions() *ResilienceOptions {
	cbOpts := circuitbreaker.DefaultOptions()
	cbOpts.MaxFailures = 1000 // Effectively disable the circuit breaker by default.
	return &ResilienceOptions{
		Retry: &RetryOptions{
			MaxRetries:      3,
			RetryDelay:      5 * time.Millisecond,
			BackoffFactor:   2.0,
			MaxDelay:        50 * time.Millisecond,
			RetryableErrors: []error{},
		},
		Timeout:        &TimeoutOptions{Timeout: 10 * time.Second},
		CircuitBreaker: &cbOpts,
		Fallback:       DefaultFallbackOptions(),
	}
}

func TestDefaultOptions(t *testing.T) {
	t.Run("retry defaults", func(t *testing.T) {
		opts := DefaultRetryOptions()
		assert.Equal(t, 3, opts.MaxRetries)
		assert.Equal(t, 100*time.Millisecond, opts.RetryDelay)
		assert.Equal(t, 2.0, opts.BackoffFactor)
		assert.Equal(t, 30*time.Second, opts.MaxDelay)
		assert.Empty(t, opts.RetryableErrors)
	})

	t.Run("timeout defaults", func(t *testing.T) {
		opts := DefaultTimeoutOptions()
		assert.Equal(t, 30*time.Second, opts.Timeout)
	})

	t.Run("fallback defaults", func(t *testing.T) {
		opts := DefaultFallbackOptions()
		assert.Nil(t, opts.FallbackFunc)
	})

	t.Run("resilience defaults", func(t *testing.T) {
		opts := DefaultResilienceOptions()
		require.NotNil(t, opts.Retry)
		require.NotNil(t, opts.Timeout)
		require.NotNil(t, opts.CircuitBreaker)
		require.NotNil(t, opts.Fallback)
	})
}

func TestNew(t *testing.T) {
	t.Run("nil options uses defaults", func(t *testing.T) {
		r := New(nil)
		require.NotNil(t, r)
		assert.NotNil(t, r.GetOptions())
		assert.NotNil(t, r.GetCircuitBreaker())
		assert.Equal(t, 3, r.GetOptions().Retry.MaxRetries)
	})

	t.Run("custom options are kept", func(t *testing.T) {
		opts := testOptions()
		r := New(opts)
		assert.Same(t, opts, r.GetOptions())
	})
}

func TestExecute_SuccessFirstAttempt(t *testing.T) {
	r := New(testOptions())
	var attempts int32

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		atomic.AddInt32(&attempts, 1)
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestExecute_RetryWithBackoffThenSuccess(t *testing.T) {
	r := New(testOptions())
	var attempts int32

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		if atomic.AddInt32(&attempts, 1) < 3 {
			return nil, errTransient
		}
		return "recovered", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "recovered", result)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestExecute_MaxRetriesExhausted(t *testing.T) {
	opts := testOptions()
	opts.Retry.MaxRetries = 2
	r := New(opts)
	var attempts int32

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		atomic.AddInt32(&attempts, 1)
		return nil, errTransient
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRetryFailed)
	assert.ErrorIs(t, err, errTransient)
	// Initial attempt + 2 retries.
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestExecute_NonRetryableErrorStopsImmediately(t *testing.T) {
	errPermanent := errors.New("permanent failure")

	opts := testOptions()
	opts.Retry.RetryableErrors = []error{errTransient}
	r := New(opts)
	var attempts int32

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		atomic.AddInt32(&attempts, 1)
		return nil, errPermanent
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, errPermanent)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts), "non-retryable error must not be retried")
}

func TestExecute_RetryableErrorListIsRetried(t *testing.T) {
	opts := testOptions()
	opts.Retry.RetryableErrors = []error{errTransient}
	r := New(opts)
	var attempts int32

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		if atomic.AddInt32(&attempts, 1) < 2 {
			return nil, errTransient
		}
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestExecute_TimeoutDuringFunction(t *testing.T) {
	opts := testOptions()
	opts.Timeout = &TimeoutOptions{Timeout: 100 * time.Millisecond}
	r := New(opts)

	start := time.Now()
	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTimeout)
	assert.Less(t, time.Since(start), 5*time.Second)
}

// TestExecute_TimeoutWhileWaitingBetweenRetries covers a previously fixed bug:
// when the context deadline expires while Execute is sleeping between retries,
// it must return ErrTimeout and stop retrying instead of running the full
// retry schedule.
func TestExecute_TimeoutWhileWaitingBetweenRetries(t *testing.T) {
	opts := testOptions()
	opts.Timeout = &TimeoutOptions{Timeout: 200 * time.Millisecond}
	opts.Retry = &RetryOptions{
		MaxRetries:      20,
		RetryDelay:      2 * time.Second, // Much longer than the timeout.
		BackoffFactor:   1.0,
		MaxDelay:        2 * time.Second,
		RetryableErrors: []error{},
	}
	r := New(opts)
	var attempts int32

	start := time.Now()
	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		atomic.AddInt32(&attempts, 1)
		return nil, errTransient
	})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTimeout, "deadline during retry wait must surface ErrTimeout")
	// It must give up as soon as the deadline expires, not run all 20 retries
	// (which would take ~40s).
	assert.LessOrEqual(t, atomic.LoadInt32(&attempts), int32(2), "must stop retrying once the deadline expires")
	assert.Less(t, elapsed, 5*time.Second, "must return promptly after the deadline expires")
}

func TestExecute_ParentContextDeadline(t *testing.T) {
	opts := testOptions()
	opts.Timeout = nil // No per-call timeout; only the parent context deadline applies.
	r := New(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := r.Execute(ctx, func(c context.Context) (interface{}, error) {
		<-c.Done()
		return nil, c.Err()
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTimeout)
}

func TestExecute_FallbackOnFailure(t *testing.T) {
	opts := testOptions()
	opts.Retry.MaxRetries = 1
	opts.Fallback = &FallbackOptions{
		FallbackFunc: func(ctx context.Context, err error) (interface{}, error) {
			assert.ErrorIs(t, err, errTransient)
			return "fallback-value", nil
		},
	}
	r := New(opts)

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		return nil, errTransient
	})

	require.NoError(t, err)
	assert.Equal(t, "fallback-value", result)
}

func TestExecute_FallbackFailure(t *testing.T) {
	errFallback := errors.New("fallback exploded")

	opts := testOptions()
	opts.Retry.MaxRetries = 0
	opts.Fallback = &FallbackOptions{
		FallbackFunc: func(ctx context.Context, err error) (interface{}, error) {
			return nil, errFallback
		},
	}
	r := New(opts)

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		return nil, errTransient
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrFallbackFail)
	assert.ErrorIs(t, err, errFallback)
}

func TestExecute_FallbackNotCalledOnSuccess(t *testing.T) {
	var fallbackCalled int32

	opts := testOptions()
	opts.Fallback = &FallbackOptions{
		FallbackFunc: func(ctx context.Context, err error) (interface{}, error) {
			atomic.AddInt32(&fallbackCalled, 1)
			return nil, nil
		},
	}
	r := New(opts)

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		return 42, nil
	})

	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, int32(0), atomic.LoadInt32(&fallbackCalled))
}

func TestExecute_CircuitBreakerOpen(t *testing.T) {
	cbOpts := circuitbreaker.Options{
		Name:          "test",
		MaxFailures:   1,
		ResetTimeout:  time.Hour, // Stays open for the duration of the test.
		HalfOpenLimit: 1,
	}
	opts := &ResilienceOptions{
		Retry:          nil, // No retries so a single failure trips the breaker.
		Timeout:        nil,
		CircuitBreaker: &cbOpts,
		Fallback:       nil,
	}
	r := New(opts)

	// First call fails and trips the breaker.
	_, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		return nil, errTransient
	})
	require.ErrorIs(t, err, errTransient)
	assert.Equal(t, circuitbreaker.StateOpen, r.GetCircuitBreaker().State())

	// Second call is rejected without invoking the function.
	var attempts int32
	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		atomic.AddInt32(&attempts, 1)
		return "should not run", nil
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCircuitOpen)
	assert.Nil(t, result)
	assert.Equal(t, int32(0), atomic.LoadInt32(&attempts))
}

func TestExecute_CircuitBreakerOpenUsesFallback(t *testing.T) {
	cbOpts := circuitbreaker.Options{
		Name:          "test",
		MaxFailures:   1,
		ResetTimeout:  time.Hour,
		HalfOpenLimit: 1,
	}
	var fallbackErrors []error
	opts := &ResilienceOptions{
		Retry:          nil,
		Timeout:        nil,
		CircuitBreaker: &cbOpts,
		Fallback: &FallbackOptions{
			FallbackFunc: func(ctx context.Context, err error) (interface{}, error) {
				fallbackErrors = append(fallbackErrors, err)
				return "degraded", nil
			},
		},
	}
	r := New(opts)

	// First call fails, trips the breaker, and is handled by the fallback.
	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		return nil, errTransient
	})
	require.NoError(t, err)
	assert.Equal(t, "degraded", result)
	require.Equal(t, circuitbreaker.StateOpen, r.GetCircuitBreaker().State())

	// Second call is rejected by the open circuit and falls back without
	// invoking the function.
	var attempts int32
	result, err = r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		atomic.AddInt32(&attempts, 1)
		return "should not run", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "degraded", result)
	assert.Equal(t, int32(0), atomic.LoadInt32(&attempts))

	require.Len(t, fallbackErrors, 2)
	assert.ErrorIs(t, fallbackErrors[0], errTransient)
	assert.ErrorIs(t, fallbackErrors[1], ErrCircuitOpen)
}

func TestReset(t *testing.T) {
	cbOpts := circuitbreaker.Options{
		Name:          "test",
		MaxFailures:   1,
		ResetTimeout:  time.Hour,
		HalfOpenLimit: 1,
	}
	opts := &ResilienceOptions{
		Retry:          nil,
		Timeout:        nil,
		CircuitBreaker: &cbOpts,
	}
	r := New(opts)

	_, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		return nil, errTransient
	})
	require.Error(t, err)
	require.Equal(t, circuitbreaker.StateOpen, r.GetCircuitBreaker().State())

	r.Reset()
	assert.Equal(t, circuitbreaker.StateClosed, r.GetCircuitBreaker().State())

	result, err := r.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}

func TestIsRetryableError(t *testing.T) {
	wrapped := errors.Join(errors.New("outer"), errTransient)

	tests := []struct {
		name            string
		err             error
		retryableErrors []error
		expected        bool
	}{
		{"empty list retries everything", errTransient, nil, true},
		{"error in list", errTransient, []error{errTransient}, true},
		{"wrapped error in list", wrapped, []error{errTransient}, true},
		{"error not in list", errors.New("other"), []error{errTransient}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isRetryableError(tt.err, tt.retryableErrors))
		})
	}
}
