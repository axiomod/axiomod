package elk

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bulkRecorder struct {
	mu    sync.Mutex
	lines []string
}

func (r *bulkRecorder) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/_bulk" {
			scanner := bufio.NewScanner(req.Body)
			r.mu.Lock()
			for scanner.Scan() {
				r.lines = append(r.lines, scanner.Text())
			}
			r.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errors":false}`))
			return
		}
		// Root ping
		_, _ = w.Write([]byte(`{"cluster_name":"test"}`))
	}
}

func (r *bulkRecorder) lineCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.lines)
}

func newTestLogger(t *testing.T) *observability.Logger {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)
	return logger
}

func TestPlugin_Lifecycle(t *testing.T) {
	rec := &bulkRecorder{}
	server := httptest.NewServer(rec.handler())
	defer server.Close()

	p := &Plugin{}
	assert.Equal(t, "elk", p.Name())

	settings := map[string]interface{}{"elasticsearchUrl": server.URL}
	require.NoError(t, p.Initialize(settings, newTestLogger(t), nil, nil, nil))
	assert.NoError(t, p.Start())
	assert.NoError(t, p.Stop())
}

func TestPlugin_InitializeValidation(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{"missing url", map[string]interface{}{}, true},
		{"valid url", map[string]interface{}{"elasticsearchUrl": "http://localhost:9200"}, false},
		{"invalid flush interval", map[string]interface{}{
			"elasticsearchUrl": "http://localhost:9200", "flushInterval": "not-a-duration"}, true},
		{"valid custom settings", map[string]interface{}{
			"elasticsearchUrl": "http://localhost:9200", "index": "app-logs",
			"flushInterval": "1s", "bufferSize": 10}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{}
			err := p.Initialize(tt.settings, newTestLogger(t), nil, nil, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestPlugin_EnqueueAndFlush(t *testing.T) {
	rec := &bulkRecorder{}
	server := httptest.NewServer(rec.handler())
	defer server.Close()

	p := &Plugin{}
	settings := map[string]interface{}{
		"elasticsearchUrl": server.URL,
		"index":            "test-logs",
		"flushInterval":    "1h", // ensure flushes in this test are explicit
	}
	require.NoError(t, p.Initialize(settings, newTestLogger(t), nil, nil, nil))

	require.NoError(t, p.Enqueue(map[string]interface{}{"msg": "one"}))
	require.NoError(t, p.Enqueue(map[string]interface{}{"msg": "two"}))
	require.NoError(t, p.Flush(context.Background()))

	// Two documents = two action lines + two source lines
	assert.Equal(t, 4, rec.lineCount())

	// Flushing an empty buffer is a no-op
	require.NoError(t, p.Flush(context.Background()))
	assert.Equal(t, 4, rec.lineCount())
}

func TestPlugin_EnqueueFlushesWhenBufferFull(t *testing.T) {
	rec := &bulkRecorder{}
	server := httptest.NewServer(rec.handler())
	defer server.Close()

	p := &Plugin{}
	settings := map[string]interface{}{
		"elasticsearchUrl": server.URL,
		"bufferSize":       2,
		"flushInterval":    "1h",
	}
	require.NoError(t, p.Initialize(settings, newTestLogger(t), nil, nil, nil))

	require.NoError(t, p.Enqueue(map[string]interface{}{"msg": "one"}))
	assert.Equal(t, 0, rec.lineCount(), "buffer below threshold must not flush")

	require.NoError(t, p.Enqueue(map[string]interface{}{"msg": "two"}))
	assert.Equal(t, 4, rec.lineCount(), "reaching bufferSize must flush inline")
}

func TestPlugin_StopFlushesRemaining(t *testing.T) {
	rec := &bulkRecorder{}
	server := httptest.NewServer(rec.handler())
	defer server.Close()

	p := &Plugin{}
	settings := map[string]interface{}{
		"elasticsearchUrl": server.URL,
		"flushInterval":    "1h",
	}
	require.NoError(t, p.Initialize(settings, newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())

	require.NoError(t, p.Enqueue(map[string]interface{}{"msg": "pending"}))
	require.NoError(t, p.Stop())

	assert.Equal(t, 2, rec.lineCount(), "Stop must flush buffered documents")
}

func TestPlugin_HealthCheck(t *testing.T) {
	rec := &bulkRecorder{}
	server := httptest.NewServer(rec.handler())
	defer server.Close()

	logger := newTestLogger(t)
	h := health.New(logger)

	p := &Plugin{}
	settings := map[string]interface{}{"elasticsearchUrl": server.URL}
	require.NoError(t, p.Initialize(settings, logger, nil, nil, h))

	assert.NoError(t, p.ping())

	server.Close()
	assert.Error(t, p.ping())
}
