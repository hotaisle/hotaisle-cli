// Package client_gen provides a generated HTTPS client for the HotAisle API
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default API base URL
	DefaultBaseURL = "https://admin.hotaisle.app/api"
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client is an HTTPS client for the HotAisle API
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	userAgent  string
}

// Option is a function that configures a Client
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client (primarily for testing)
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithToken sets the authentication token
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

// WithBaseURL sets a custom base URL
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithUserAgent sets a custom user agent
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// NewClient creates a new HotAisle API client
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		userAgent: "hotaisle/1.0",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SetToken updates the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// doRequest executes an HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	// Handle 204 No Content
	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return nil
	}

	// Unmarshal response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// APIError represents an API error response
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// buildPath constructs a URL path with path parameters
func buildPath(template string, params map[string]string) string {
	path := template
	for key, value := range params {
		placeholder := "{" + key + "}"
		path = strings.ReplaceAll(path, placeholder, url.PathEscape(value))
	}
	return path
}

// buildQuery constructs a URL query string
//func buildQuery(params map[string]string) string {
//	if len(params) == 0 {
//		return ""
//	}
//	query := url.Values{}
//	for key, value = range params {
//		query.Set(key, value)
//	}
//	return "?" + query.Encode()
//}
