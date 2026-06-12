// Package elk provides a plugin that ships log documents to Elasticsearch
// using the bulk API, with buffering and periodic background flushing.
package elk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/plugins"

	"go.uber.org/zap"
)

const (
	defaultIndex         = "axiomod-logs"
	defaultFlushInterval = 5 * time.Second
	defaultBufferSize    = 256
	defaultHTTPTimeout   = 10 * time.Second
)

// Plugin buffers documents and ships them to Elasticsearch in bulk.
type Plugin struct {
	logger *observability.Logger

	url           string
	index         string
	flushInterval time.Duration
	bufferSize    int
	client        *http.Client

	mu     sync.Mutex
	buffer []map[string]interface{}

	cancel context.CancelFunc
	done   chan struct{}
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "elk"
}

// Initialize configures the plugin from its settings.
//
// Settings:
//   - elasticsearchUrl (string): base URL of the Elasticsearch cluster (required)
//   - index (string): target index, default "axiomod-logs"
//   - flushInterval (string): Go duration between background flushes, default "5s"
//   - bufferSize (int): documents buffered before a forced flush, default 256
func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, h *health.Health) error {
	settings = plugins.NormalizeSettings(settings)
	p.logger = logger
	p.index = defaultIndex
	p.flushInterval = defaultFlushInterval
	p.bufferSize = defaultBufferSize
	p.client = &http.Client{Timeout: defaultHTTPTimeout}

	url, _ := settings["elasticsearchurl"].(string)
	if url == "" {
		return fmt.Errorf("elk plugin: elasticsearchUrl setting is required")
	}
	p.url = url

	if index, ok := settings["index"].(string); ok && index != "" {
		p.index = index
	}
	if interval, ok := settings["flushinterval"].(string); ok && interval != "" {
		parsed, err := time.ParseDuration(interval)
		if err != nil {
			return fmt.Errorf("elk plugin: invalid flushInterval %q: %w", interval, err)
		}
		p.flushInterval = parsed
	}
	if size, ok := settings["buffersize"].(int); ok && size > 0 {
		p.bufferSize = size
	}

	if h != nil {
		h.RegisterCheck(p.Name(), p.ping)
	}

	return nil
}

// Start launches the background flush loop.
func (p *Plugin) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.done = make(chan struct{})

	go func() {
		defer close(p.done)
		ticker := time.NewTicker(p.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := p.Flush(ctx); err != nil && p.logger != nil {
					p.logger.Error("ELK flush failed", zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	if p.logger != nil {
		p.logger.Info("ELK plugin started",
			zap.String("url", p.url),
			zap.String("index", p.index),
			zap.Duration("flush_interval", p.flushInterval),
		)
	}
	return nil
}

// Stop cancels the background loop and flushes any remaining documents.
func (p *Plugin) Stop() error {
	if p.cancel != nil {
		p.cancel()
		<-p.done
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
	defer cancel()
	return p.Flush(ctx)
}

// Enqueue adds a document to the buffer. The buffer is flushed inline when
// it reaches the configured size.
func (p *Plugin) Enqueue(doc map[string]interface{}) error {
	p.mu.Lock()
	p.buffer = append(p.buffer, doc)
	full := len(p.buffer) >= p.bufferSize
	p.mu.Unlock()

	if full {
		ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
		defer cancel()
		return p.Flush(ctx)
	}
	return nil
}

// Flush ships all buffered documents to Elasticsearch via the bulk API.
func (p *Plugin) Flush(ctx context.Context) error {
	p.mu.Lock()
	docs := p.buffer
	p.buffer = nil
	p.mu.Unlock()

	if len(docs) == 0 {
		return nil
	}

	var body bytes.Buffer
	action := fmt.Sprintf(`{"index":{"_index":%q}}`, p.index)
	for _, doc := range docs {
		line, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("elk plugin: failed to marshal document: %w", err)
		}
		body.WriteString(action)
		body.WriteByte('\n')
		body.Write(line)
		body.WriteByte('\n')
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url+"/_bulk", &body)
	if err != nil {
		return fmt.Errorf("elk plugin: failed to create bulk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("elk plugin: bulk request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("elk plugin: bulk request returned status %s", resp.Status)
	}
	return nil
}

// ping checks that the Elasticsearch cluster is reachable.
func (p *Plugin) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("elasticsearch returned status %s", resp.Status)
	}
	return nil
}
