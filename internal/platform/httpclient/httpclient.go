package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ClientConfig holds configuration for the HTTP client wrapper.
type ClientConfig struct {
	BaseURL        string
	Timeout        time.Duration
	MaxRetries     int
	RetryWaitMin   time.Duration
	RetryWaitMax   time.Duration
	DefaultHeaders map[string]string
	Transport      http.RoundTripper // optional custom transport
}

// RequestOptions carries per-request options.
type RequestOptions struct {
	Headers map[string]string
	Query   url.Values
	// If Body is non-nil and BodyReader is nil, Body will be JSON-encoded.
	Body        any
	BodyReader  io.Reader
	ContentType string // when using BodyReader
}

// Response contains a simplified response snapshot.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// HTTPError represents a non-2xx response.
type HTTPError struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

func (e *HTTPError) Error() string {
	snippet := string(e.Body)
	if len(snippet) > 256 {
		snippet = snippet[:256] + "…"
	}
	return fmt.Sprintf("http %d: %s", e.StatusCode, snippet)
}

// Client is a thin wrapper over http.Client with retries, JSON helpers, and base URL support.
type Client struct {
	raw  *http.Client
	cfg  ClientConfig
	base *url.URL
}

// New creates a new HTTP client with sensible defaults.
func New(cfg ClientConfig) (*Client, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	if cfg.RetryWaitMin <= 0 {
		cfg.RetryWaitMin = 200 * time.Millisecond
	}
	if cfg.RetryWaitMax <= 0 {
		cfg.RetryWaitMax = 2 * time.Second
	}
	// default: 2 retries (total attempts <= 3)
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	transport := cfg.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	var base *url.URL
	var err error
	if strings.TrimSpace(cfg.BaseURL) != "" {
		base, err = url.Parse(cfg.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("parse base url: %w", err)
		}
	}
	return &Client{
		raw:  &http.Client{Timeout: cfg.Timeout, Transport: transport},
		cfg:  cfg,
		base: base,
	}, nil
}

// Do executes an HTTP request with retries (for idempotent methods) and optional JSON marshal/unmarshal.
// If out is non-nil and Content-Type is application/json, the body is unmarshaled into out.
func (c *Client) Do(ctx context.Context, method, pathOrURL string, opts *RequestOptions, out any) (*Response, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}

	// Build URL
	fullURL, err := c.resolveURL(pathOrURL, opts.Query)
	if err != nil {
		return nil, err
	}

	// Prepare body
	var bodyReader io.Reader
	var contentType string
	if opts.BodyReader != nil {
		bodyReader = opts.BodyReader
		contentType = opts.ContentType
	} else if opts.Body != nil {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(opts.Body); err != nil {
			return nil, fmt.Errorf("encode json body: %w", err)
		}
		bodyReader = buf
		contentType = "application/json"
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	// Headers: default + per-request
	for k, v := range c.cfg.DefaultHeaders {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	for k, v := range opts.Headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json, */*;q=0.1")
	}

	// Execute with retries for idempotent methods (GET, HEAD, OPTIONS, PUT, DELETE)
	attempts := 0
	for {
		attempts++
		resp, err := c.raw.Do(req)
		if err != nil {
			// Retry on transient network errors
			if shouldRetryError(method, err) && attempts <= c.cfg.MaxRetries+1 {
				if err := backoffWait(ctx, attempts-1, c.cfg.RetryWaitMin, c.cfg.RetryWaitMax); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		// Read body
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}

		r := &Response{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: body}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Decode JSON when requested
			if out != nil && isJSON(resp.Header.Get("Content-Type")) && len(body) > 0 {
				if err := json.Unmarshal(body, out); err != nil {
					return r, fmt.Errorf("decode json: %w", err)
				}
			}
			return r, nil
		}

		// Retry on 429 or 5xx for idempotent methods
		if shouldRetryStatus(method, resp.StatusCode) && attempts <= c.cfg.MaxRetries+1 {
			if err := backoffWait(ctx, attempts-1, c.cfg.RetryWaitMin, c.cfg.RetryWaitMax); err != nil {
				return nil, err
			}
			continue
		}

		return r, &HTTPError{StatusCode: resp.StatusCode, Body: body, Header: resp.Header.Clone()}
	}
}

// Convenience helpers
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) (*Response, error) {
	return c.Do(ctx, http.MethodGet, path, &RequestOptions{Query: query}, out)
}

func (c *Client) PostJSON(ctx context.Context, path string, body any, out any) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, &RequestOptions{Body: body}, out)
}

func (c *Client) PutJSON(ctx context.Context, path string, body any, out any) (*Response, error) {
	return c.Do(ctx, http.MethodPut, path, &RequestOptions{Body: body}, out)
}

func (c *Client) PatchJSON(ctx context.Context, path string, body any, out any) (*Response, error) {
	return c.Do(ctx, http.MethodPatch, path, &RequestOptions{Body: body}, out)
}

func (c *Client) Delete(ctx context.Context, path string, query url.Values) (*Response, error) {
	return c.Do(ctx, http.MethodDelete, path, &RequestOptions{Query: query}, nil)
}

// Helpers
func (c *Client) resolveURL(pathOrURL string, q url.Values) (string, error) {
	if c.base == nil {
		// absolute or raw provided
		u, err := url.Parse(pathOrURL)
		if err != nil {
			return "", fmt.Errorf("parse url: %w", err)
		}
		if q != nil {
			if u.RawQuery == "" {
				u.RawQuery = q.Encode()
			} else {
				vals, _ := url.ParseQuery(u.RawQuery)
				for k, vs := range q {
					for _, v := range vs {
						vals.Add(k, v)
					}
				}
				u.RawQuery = vals.Encode()
			}
		}
		return u.String(), nil
	}
	// join base + path
	rel, err := url.Parse(pathOrURL)
	if err != nil {
		return "", fmt.Errorf("parse path: %w", err)
	}
	u := c.base.ResolveReference(rel)
	if q != nil {
		vals := u.Query()
		for k, vs := range q {
			for _, v := range vs {
				vals.Add(k, v)
			}
		}
		u.RawQuery = vals.Encode()
	}
	return u.String(), nil
}

func isJSON(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "application/json") || strings.Contains(ct, "+json")
}

func shouldRetryStatus(method string, status int) bool {
	if status == http.StatusTooManyRequests { // 429
		return isIdempotent(method)
	}
	if status >= 500 && status != http.StatusNotImplemented { // 5xx
		return isIdempotent(method)
	}
	return false
}

func shouldRetryError(method string, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	// For network errors, http.Client usually returns *url.Error; allow retry for idempotent methods.
	return isIdempotent(method)
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func backoffWait(ctx context.Context, attempt int, min, max time.Duration) error {
	// Exponential with jitter.
	pow := math.Pow(2, float64(attempt))
	wait := time.Duration(float64(min) * pow)
	if wait > max {
		wait = max
	}
	// Full jitter
	if wait > 0 {
		jitter := time.Duration(rand.Int63n(int64(wait)))
		wait = jitter
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
