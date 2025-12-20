package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/axiomod/axiomod/framework/circuitbreaker"
)

// HTTPClient is a resilient HTTP client with circuit breaker, retries, and timeouts
type HTTPClient struct {
	client         *http.Client
	circuitBreaker *circuitbreaker.CircuitBreaker
	maxRetries     int
	retryDelay     time.Duration
}

// Options contains options for creating a new HTTPClient
type Options struct {
	// Timeout is the timeout for HTTP requests
	Timeout time.Duration
	// CircuitBreakerOptions contains options for the circuit breaker
	CircuitBreakerOptions circuitbreaker.Options
	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// DefaultOptions returns the default options for an HTTP client
func DefaultOptions() Options {
	return Options{
		Timeout:               30 * time.Second,
		CircuitBreakerOptions: circuitbreaker.DefaultOptions(),
		MaxRetries:            3,
		RetryDelay:            100 * time.Millisecond,
	}
}

// New creates a new HTTPClient with the given options
func New(options Options) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: options.Timeout,
		},
		circuitBreaker: circuitbreaker.New(options.CircuitBreakerOptions),
		maxRetries:     options.MaxRetries,
		retryDelay:     options.RetryDelay,
	}
}

// Get performs a GET request with circuit breaker and retry logic
func (c *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return c.doWithRetry(req)
}

// GetJSON performs a GET request and unmarshals the response into the given value
func (c *HTTPClient) GetJSON(ctx context.Context, url string, headers map[string]string, v interface{}) error {
	resp, err := c.Get(ctx, url, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// Post performs a POST request with circuit breaker and retry logic
func (c *HTTPClient) Post(ctx context.Context, url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return c.doWithRetry(req)
}

// PostJSON performs a POST request with a JSON body and unmarshals the response into the given value
func (c *HTTPClient) PostJSON(ctx context.Context, url string, headers map[string]string, body interface{}, v interface{}) error {
	// Marshal the body to JSON
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	// Add content type header
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"

	resp, err := c.Post(ctx, url, headers, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// doWithRetry performs an HTTP request with circuit breaker and retry logic
func (c *HTTPClient) doWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Execute with circuit breaker
	err = c.circuitBreaker.Execute(func() error {
		// Retry logic
		for i := 0; i <= c.maxRetries; i++ {
			// Clone the request for each retry to ensure it can be reused
			reqClone := req.Clone(req.Context())

			resp, err = c.client.Do(reqClone)

			// If successful or context canceled, return immediately
			if err == nil || req.Context().Err() != nil {
				return err
			}

			// If this was the last retry, return the error
			if i == c.maxRetries {
				return err
			}

			// Wait before retrying
			select {
			case <-time.After(c.retryDelay):
				// Continue to next retry
			case <-req.Context().Done():
				// Context canceled, return immediately
				return req.Context().Err()
			}
		}

		return err
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
