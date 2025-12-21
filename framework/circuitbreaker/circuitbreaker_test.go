package circuitbreaker

import (
	"errors"
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
