package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/stretchr/testify/assert"
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
