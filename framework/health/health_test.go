package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	logger, _ := observability.NewLogger(&config.Config{})
	h := New(logger)

	t.Run("Initially UP", func(t *testing.T) {
		assert.Equal(t, StatusUp, h.GetStatus())
	})

	t.Run("Register and Run Checks", func(t *testing.T) {
		h.RegisterCheck("db", func() error { return nil })
		h.RegisterCheck("redis", func() error { return errors.New("redis down") })

		h.RunChecks()

		assert.Equal(t, StatusDown, h.GetStatus())
		resp := h.GetResponse()
		assert.Equal(t, StatusDown, resp.Status)
		assert.Equal(t, StatusUp, resp.Components["db"].Status)
		assert.Equal(t, StatusDown, resp.Components["redis"].Status)
		assert.Equal(t, "redis down", resp.Components["redis"].Error)
	})

	t.Run("HTTP Handler", func(t *testing.T) {
		handler := h.Handler()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "DOWN")
	})

	t.Run("Background Checks", func(t *testing.T) {
		h.RegisterCheck("bg", func() error { return nil })
		// Just verify it doesn't panic and can be stopped
		ctx, cancel := context.WithCancel(context.Background())
		go h.StartBackgroundChecks(ctx, 10*time.Millisecond)
		time.Sleep(25 * time.Millisecond)
		cancel()
		time.Sleep(10 * time.Millisecond)
	})
}

func newTestHealth(t *testing.T) *Health {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)
	return New(logger)
}

func TestHealthAggregation(t *testing.T) {
	h := newTestHealth(t)

	h.RegisterCheck("ok", func() error { return nil })
	h.RegisterCheck("broken", func() error { return errors.New("dependency down") })

	// Before any run, components are UNKNOWN but overall status is UP.
	assert.Equal(t, StatusUp, h.GetStatus())

	h.RunChecks()
	assert.Equal(t, StatusDown, h.GetStatus(), "one DOWN component must fail the aggregate")

	resp := h.GetResponse()
	assert.Equal(t, StatusDown, resp.Status)
	assert.Equal(t, StatusUp, resp.Components["ok"].Status)
	assert.Equal(t, StatusDown, resp.Components["broken"].Status)
	assert.Equal(t, "dependency down", resp.Components["broken"].Error)
	assert.False(t, resp.Timestamp.IsZero())
}

func TestHealthHandler(t *testing.T) {
	t.Run("healthy returns 200", func(t *testing.T) {
		h := newTestHealth(t)
		h.RegisterCheck("ok", func() error { return nil })

		rec := httptest.NewRecorder()
		h.Handler()(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp Response
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, StatusUp, resp.Status)
	})

	t.Run("unhealthy returns 503", func(t *testing.T) {
		h := newTestHealth(t)
		h.RegisterCheck("broken", func() error { return errors.New("boom") })

		rec := httptest.NewRecorder()
		h.Handler()(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})
}

func TestHealthBackgroundChecks(t *testing.T) {
	h := newTestHealth(t)

	var calls atomic.Int32
	h.RegisterCheck("counted", func() error {
		calls.Add(1)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.StartBackgroundChecks(ctx, 20*time.Millisecond)
	}()

	// The loop primes immediately and then ticks; wait for a few runs.
	assert.Eventually(t, func() bool { return calls.Load() >= 2 }, 2*time.Second, 10*time.Millisecond)

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("background checks did not stop on context cancellation")
	}
	assert.Equal(t, StatusUp, h.GetStatus())
}
