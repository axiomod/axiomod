package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/axiomod/axiomod/platform/observability"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the fx options for the health module
var Module = fx.Options(
	fx.Provide(New),
)

// Status represents the health status of a component
type Status string

const (
	// StatusUp indicates the component is healthy
	StatusUp Status = "UP"
	// StatusDown indicates the component is unhealthy
	StatusDown Status = "DOWN"
	// StatusUnknown indicates the component's health is unknown
	StatusUnknown Status = "UNKNOWN"
)

// CheckFunc is a function that checks the health of a component
type CheckFunc func() error

// Component represents a component with health status
type Component struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Health provides health checking functionality
type Health struct {
	mu       sync.RWMutex
	checks   map[string]CheckFunc
	statuses map[string]Component
	logger   *observability.Logger
}

// Response represents the health check response
type Response struct {
	Status     Status               `json:"status"`
	Components map[string]Component `json:"components,omitempty"`
	Timestamp  time.Time            `json:"timestamp"`
}

// New creates a new Health instance
func New(logger *observability.Logger) *Health {
	return &Health{
		checks:   make(map[string]CheckFunc),
		statuses: make(map[string]Component),
		logger:   logger,
	}
}

// RegisterCheck registers a health check for a component
func (h *Health) RegisterCheck(name string, check CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checks[name] = check
	h.statuses[name] = Component{
		Name:   name,
		Status: StatusUnknown,
	}

	h.logger.Debug("Registered health check", zap.String("component", name))
}

// RunChecks runs all registered health checks
func (h *Health) RunChecks() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, check := range h.checks {
		component := Component{
			Name:   name,
			Status: StatusUp,
		}

		if err := check(); err != nil {
			component.Status = StatusDown
			component.Error = err.Error()
			h.logger.Warn("Health check failed", zap.String("component", name), zap.Error(err))
		} else {
			h.logger.Debug("Health check passed", zap.String("component", name))
		}

		h.statuses[name] = component
	}
}

// GetStatus returns the overall health status
func (h *Health) GetStatus() Status {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, component := range h.statuses {
		if component.Status == StatusDown {
			return StatusDown
		}
	}

	return StatusUp
}

// GetResponse returns the health check response
func (h *Health) GetResponse() Response {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := h.GetStatus()
	components := make(map[string]Component)

	for name, component := range h.statuses {
		components[name] = component
	}

	return Response{
		Status:     status,
		Components: components,
		Timestamp:  time.Now(),
	}
}

// Handler returns an HTTP handler for health checks
func (h *Health) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Run health checks
		h.RunChecks()

		// Get response
		response := h.GetResponse()

		// Set status code
		if response.Status == StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		// Set content type
		w.Header().Set("Content-Type", "application/json")

		// Write response
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Error("Failed to encode health check response", zap.Error(err))
		}
	}
}

// StartBackgroundChecks starts running health checks in the background
func (h *Health) StartBackgroundChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	h.logger.Info("Starting background health checks", zap.Duration("interval", interval))

	for {
		select {
		case <-ticker.C:
			h.RunChecks()
		case <-ctx.Done():
			h.logger.Info("Stopping background health checks")
			return
		}
	}
}
