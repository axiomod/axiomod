// Package audit provides a plugin that records structured audit events
// describing who did what to which resource, with an optional append-only
// JSON-lines file sink in addition to the structured log.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/observability"

	"go.uber.org/zap"
)

// Event is a single audit record.
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Plugin records audit events.
type Plugin struct {
	logger *observability.Logger

	mu       sync.Mutex
	filePath string
	file     *os.File
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "auditing"
}

// Initialize configures the plugin from its settings.
//
// Settings:
//   - filePath (string): optional path of an append-only JSON-lines audit
//     log. When empty, events are only written to the structured logger.
func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.logger = logger

	if filePath, ok := settings["filePath"].(string); ok {
		p.filePath = filePath
	}

	return nil
}

// Start opens the file sink when one is configured.
func (p *Plugin) Start() error {
	if p.filePath != "" {
		file, err := os.OpenFile(p.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("audit plugin: failed to open audit log %s: %w", p.filePath, err)
		}
		p.mu.Lock()
		p.file = file
		p.mu.Unlock()
	}

	if p.logger != nil {
		p.logger.Info("Audit plugin started", zap.String("file_path", p.filePath))
	}
	return nil
}

// Stop closes the file sink.
func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.file != nil {
		err := p.file.Close()
		p.file = nil
		return err
	}
	return nil
}

// Record writes an audit event to the structured log and, when configured,
// to the JSON-lines file sink. A zero timestamp is filled in with the
// current time.
func (p *Plugin) Record(event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if p.logger != nil {
		p.logger.Info("audit",
			zap.Time("timestamp", event.Timestamp),
			zap.String("actor", event.Actor),
			zap.String("action", event.Action),
			zap.String("resource", event.Resource),
			zap.String("result", event.Result),
			zap.String("tenant_id", event.TenantID),
			zap.Any("metadata", event.Metadata),
		)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.file == nil {
		return nil
	}

	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("audit plugin: failed to marshal event: %w", err)
	}
	if _, err := p.file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("audit plugin: failed to write event: %w", err)
	}
	return nil
}
