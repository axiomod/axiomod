package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxFailures = 2
	opts.ResetTimeout = 10 * time.Millisecond
	cb := New(opts)

	t.Run("Initially Closed", func(t *testing.T) {
		assert.Equal(t, StateClosed, cb.State())
		assert.True(t, cb.AllowRequest())
	})

	t.Run("Trip to Open", func(t *testing.T) {
		err := errors.New("fail")
		cb.RecordResult(err) // 1st failure
		assert.Equal(t, StateClosed, cb.State())

		cb.RecordResult(err) // 2nd failure -> Open
		assert.Equal(t, StateOpen, cb.State())
		assert.False(t, cb.AllowRequest())
	})

	t.Run("Transition to Half-Open", func(t *testing.T) {
		time.Sleep(15 * time.Millisecond)
		assert.True(t, cb.AllowRequest()) // Should transition to Half-Open and return true
		assert.Equal(t, StateHalfOpen, cb.State())
	})

	t.Run("Transition back to Closed", func(t *testing.T) {
		cb.RecordResult(nil) // Success in Half-Open -> Closed
		assert.Equal(t, StateClosed, cb.State())
		assert.True(t, cb.AllowRequest())
	})

	t.Run("Half-Open to Open on failure", func(t *testing.T) {
		// Trip again
		cb.RecordResult(errors.New("fail"))
		cb.RecordResult(errors.New("fail"))
		assert.Equal(t, StateOpen, cb.State())

		time.Sleep(15 * time.Millisecond)
		assert.True(t, cb.AllowRequest()) // Half-Open

		cb.RecordResult(errors.New("fail")) // Failure in Half-Open -> Open
		assert.Equal(t, StateOpen, cb.State())
	})

	t.Run("Execute protection", func(t *testing.T) {
		cb.Reset()
		assert.Equal(t, StateClosed, cb.State())

		err := cb.Execute(func() error { return nil })
		assert.NoError(t, err)
		assert.Equal(t, StateClosed, cb.State())

		err = cb.Execute(func() error { return errors.New("fail") })
		assert.Error(t, err)

		err = cb.Execute(func() error { return errors.New("fail") })
		assert.Error(t, err)
		assert.Equal(t, StateOpen, cb.State())

		err = cb.Execute(func() error { return nil })
		assert.Error(t, err)
		assert.Equal(t, "circuit breaker is open", err.Error())
	})
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, "default", opts.Name)
	assert.Equal(t, 5, opts.MaxFailures)
	assert.Equal(t, 10*time.Second, opts.ResetTimeout)
	assert.Equal(t, 1, opts.HalfOpenLimit)
}

func TestCircuitBreakerFullStateWalk(t *testing.T) {
	cb := New(Options{
		Name:          "walk",
		MaxFailures:   2,
		ResetTimeout:  50 * time.Millisecond,
		HalfOpenLimit: 2,
	})
	boom := errors.New("boom")

	// Closed -> Open after MaxFailures consecutive failures.
	for i := 0; i < 2; i++ {
		assert.ErrorIs(t, cb.Execute(func() error { return boom }), boom)
	}
	assert.Equal(t, StateOpen, cb.State())

	// While open, requests are rejected without invoking the function.
	invoked := false
	err := cb.Execute(func() error { invoked = true; return nil })
	assert.Error(t, err)
	assert.False(t, invoked, "open circuit must short-circuit calls")

	// After the reset timeout the breaker admits trial requests (half-open).
	time.Sleep(80 * time.Millisecond)
	assert.True(t, cb.AllowRequest())
	assert.Equal(t, StateHalfOpen, cb.State())

	// HalfOpenLimit=2: one success keeps it half-open, the second closes it.
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, StateHalfOpen, cb.State())
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerHalfOpenFailureReopens(t *testing.T) {
	cb := New(Options{Name: "reopen", MaxFailures: 1, ResetTimeout: 30 * time.Millisecond, HalfOpenLimit: 1})
	boom := errors.New("boom")

	assert.ErrorIs(t, cb.Execute(func() error { return boom }), boom)
	assert.Equal(t, StateOpen, cb.State())

	time.Sleep(60 * time.Millisecond)
	cb.AllowRequest()
	assert.Equal(t, StateHalfOpen, cb.State())

	assert.ErrorIs(t, cb.Execute(func() error { return boom }), boom)
	assert.Equal(t, StateOpen, cb.State(), "failure in half-open must reopen the circuit")
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := New(Options{Name: "reset", MaxFailures: 1, ResetTimeout: time.Hour, HalfOpenLimit: 1})

	assert.Error(t, cb.Execute(func() error { return errors.New("boom") }))
	assert.Equal(t, StateOpen, cb.State())

	cb.Reset()
	assert.Equal(t, StateClosed, cb.State())
	assert.NoError(t, cb.Execute(func() error { return nil }))
}

func TestCircuitBreakerConcurrentExecute(t *testing.T) {
	cb := New(Options{Name: "concurrent", MaxFailures: 1000, ResetTimeout: time.Second, HalfOpenLimit: 1})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = cb.Execute(func() error {
				if n%2 == 0 {
					return errors.New("even failure")
				}
				return nil
			})
		}(i)
	}
	wg.Wait()
	assert.Equal(t, StateClosed, cb.State(), "failures below threshold keep the circuit closed")
}
