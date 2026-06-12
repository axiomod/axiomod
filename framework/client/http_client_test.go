package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/circuitbreaker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testOptions returns client options with fast timings and an effectively
// disabled circuit breaker, suitable for tests.
func testOptions() Options {
	return Options{
		Timeout: 5 * time.Second,
		CircuitBreakerOptions: circuitbreaker.Options{
			Name:          "test",
			MaxFailures:   1000,
			ResetTimeout:  time.Hour,
			HalfOpenLimit: 1,
		},
		MaxRetries: 3,
		RetryDelay: 10 * time.Millisecond,
	}
}

// newFlakyServer returns a server that abruptly closes the connection for the
// first `failures` requests (producing a transport error on the client side)
// and then serves a JSON payload. It also returns a request counter.
func newFlakyServer(failures int32) (*httptest.Server, *int32) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&count, 1)
		if n <= failures {
			hj, ok := w.(http.Hijacker)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				return
			}
			_ = conn.Close()
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	return srv, &count
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, 30*time.Second, opts.Timeout)
	assert.Equal(t, 3, opts.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, opts.RetryDelay)
	assert.Equal(t, 5, opts.CircuitBreakerOptions.MaxFailures)
}

func TestNew(t *testing.T) {
	c := New(DefaultOptions())
	require.NotNil(t, c)
	assert.NotNil(t, c.client)
	assert.NotNil(t, c.circuitBreaker)
}

func TestGet_Success(t *testing.T) {
	var gotHeader atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader.Store(r.Header.Get("X-Custom"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := New(testOptions())
	resp, err := c.Get(context.Background(), srv.URL, map[string]string{"X-Custom": "value"})
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "value", gotHeader.Load())
}

func TestGet_RetriesOnTransportError(t *testing.T) {
	srv, count := newFlakyServer(2)
	defer srv.Close()

	c := New(testOptions())
	resp, err := c.Get(context.Background(), srv.URL, nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(count), "two failed attempts plus one success")
}

func TestGet_RetriesExhausted(t *testing.T) {
	srv, count := newFlakyServer(1000)
	defer srv.Close()

	opts := testOptions()
	opts.MaxRetries = 2
	c := New(opts)

	resp, err := c.Get(context.Background(), srv.URL, nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, int32(3), atomic.LoadInt32(count), "initial attempt plus two retries")
}

func TestGet_ServerErrorStatusIsNotRetried(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(testOptions())
	resp, err := c.Get(context.Background(), srv.URL, nil)
	require.NoError(t, err, "a 5xx response is not a transport error")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestGet_ContextCancellationStopsRetries(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		// Block until the client gives up.
		<-r.Context().Done()
	}))
	defer srv.Close()

	opts := testOptions()
	opts.MaxRetries = 5
	c := New(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	resp, err := c.Get(ctx, srv.URL, nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count), "a canceled context must stop retries")
	assert.Less(t, time.Since(start), 3*time.Second)
}

func TestGet_ClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	opts := testOptions()
	opts.Timeout = 100 * time.Millisecond
	opts.MaxRetries = 0
	c := New(opts)

	resp, err := c.Get(context.Background(), srv.URL, nil)
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestGet_CircuitBreakerOpen(t *testing.T) {
	srv, count := newFlakyServer(1000)
	defer srv.Close()

	opts := testOptions()
	opts.MaxRetries = 0
	opts.CircuitBreakerOptions = circuitbreaker.Options{
		Name:          "test",
		MaxFailures:   1,
		ResetTimeout:  time.Hour,
		HalfOpenLimit: 1,
	}
	c := New(opts)

	// First request fails and trips the breaker.
	resp, err := c.Get(context.Background(), srv.URL, nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	require.Equal(t, int32(1), atomic.LoadInt32(count))

	// Second request is rejected by the open circuit without hitting the server.
	resp, err = c.Get(context.Background(), srv.URL, nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "circuit breaker is open")
	assert.Equal(t, int32(1), atomic.LoadInt32(count))
}

func TestGet_InvalidURL(t *testing.T) {
	c := New(testOptions())
	resp, err := c.Get(context.Background(), "://missing-scheme", nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create request")
}

func TestGetJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		errSubstr  string
	}{
		{"success", http.StatusOK, `{"message":"hello","count":2}`, false, ""},
		{"non-200 status", http.StatusNotFound, `{"error":"not found"}`, true, "unexpected status code: 404"},
		{"malformed json", http.StatusOK, `{invalid`, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := New(testOptions())
			var result struct {
				Message string `json:"message"`
				Count   int    `json:"count"`
			}
			err := c.GetJSON(context.Background(), srv.URL, nil, &result)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "hello", result.Message)
			assert.Equal(t, 2, result.Count)
		})
	}
}

func TestPost_Success(t *testing.T) {
	var gotBody atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody.Store(string(buf))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(testOptions())
	resp, err := c.Post(context.Background(), srv.URL, nil, strings.NewReader("payload"))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "payload", gotBody.Load())
}

func TestPostJSON(t *testing.T) {
	type request struct {
		Name string `json:"name"`
	}
	type response struct {
		ID string `json:"id"`
	}

	tests := []struct {
		name       string
		statusCode int
		body       interface{}
		target     bool
		wantErr    bool
		errSubstr  string
	}{
		{"created with response", http.StatusCreated, request{Name: "axiomod"}, true, false, ""},
		{"ok without response target", http.StatusOK, request{Name: "axiomod"}, false, false, ""},
		{"unexpected status", http.StatusBadRequest, request{Name: "axiomod"}, true, true, "unexpected status code: 400"},
		{"unmarshalable body", http.StatusOK, make(chan int), true, true, "failed to marshal body"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotContentType atomic.Value
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotContentType.Store(r.Header.Get("Content-Type"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"id":"123"}`))
			}))
			defer srv.Close()

			c := New(testOptions())
			var result *response
			if tt.target {
				result = &response{}
			}

			var err error
			if tt.target {
				err = c.PostJSON(context.Background(), srv.URL, nil, tt.body, result)
			} else {
				err = c.PostJSON(context.Background(), srv.URL, nil, tt.body, nil)
			}

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "application/json", gotContentType.Load())
			if tt.target {
				assert.Equal(t, "123", result.ID)
			}
		})
	}
}
