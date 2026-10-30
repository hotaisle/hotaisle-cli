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
	"reflect"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default API base URL
	DefaultBaseURL = "https://admin.hotaisle.app/api"
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout       = 30 * time.Second
	maxResponseBodyBytes = 16 << 20
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
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
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
func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	response, err := c.request(ctx, method, path, body)
	if err != nil {
		return err
	}
	return response.decode(result)
}

type response struct {
	status int
	body   []byte
}

func (c *Client) request(ctx context.Context, method, path string, body any) (response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return response{}, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// hardcode a long timeout just to be safe, nothing should really take this long
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return response{}, fmt.Errorf("failed to create request: %w", err)
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
		return response{}, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes+1))
	if err != nil {
		return response{}, fmt.Errorf("failed to read response body: %w", err)
	}
	if len(respBody) > maxResponseBodyBytes {
		return response{}, &ResponseError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("response body exceeds %d bytes", maxResponseBodyBytes)}
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		message := strings.TrimSpace(string(respBody))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return response{}, &APIError{
			StatusCode: resp.StatusCode,
			Message:    message,
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return response{}, &ResponseError{StatusCode: resp.StatusCode, Message: "unexpected HTTP status"}
	}
	return response{status: resp.StatusCode, body: bytes.TrimSpace(respBody)}, nil
}

func (r response) decode(result any) error {
	// Mutations documented without a response body still receive an explicit
	// CLI acknowledgement. Data-returning operations must supply actual data.
	if result == nil {
		return nil
	}
	body := r.body
	resourceType := reflect.TypeOf(result).Elem()
	isList := resourceType.Kind() == reflect.Slice
	if isList {
		resourceType = resourceType.Elem()
	}
	if len(body) == 0 || bytes.Equal(body, []byte("null")) {
		if !isList {
			return &ResponseError{StatusCode: r.status, Message: "expected a resource object, received no data"}
		}
		body = []byte("[]")
	}
	// Check the wire shape before decoding: null and {} otherwise silently
	// become zero-valued structs, including null items inside a collection.
	var objects []map[string]json.RawMessage
	if isList {
		if err := json.Unmarshal(body, &objects); err != nil {
			return &ResponseError{StatusCode: r.status, Message: "expected a JSON array of resource objects"}
		}
	} else {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(body, &object); err != nil {
			return &ResponseError{StatusCode: r.status, Message: "expected a JSON resource object"}
		}
		objects = append(objects, object)
	}
	for _, object := range objects {
		if len(object) == 0 {
			return &ResponseError{StatusCode: r.status, Message: "received an empty resource object"}
		}
		if !hasResponseField(resourceType, object) {
			return &ResponseError{StatusCode: r.status, Message: "resource object is missing expected fields"}
		}
	}
	if err := json.Unmarshal(body, result); err != nil {
		return &ResponseError{StatusCode: r.status, Message: "resource data does not match the expected response format"}
	}
	return nil
}

func hasResponseField(resourceType reflect.Type, object map[string]json.RawMessage) bool {
	for field := range resourceType.Fields() {
		if field.Anonymous {
			if hasResponseField(field.Type, object) {
				return true
			}
			continue
		}
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		for key, value := range object {
			// Preserve encoding/json's existing case-insensitive field matching.
			if strings.EqualFold(key, name) && !bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return true
			}
		}
	}
	return false
}

// ResponseError reports an API response whose status or data cannot be used.
// It deliberately excludes the response body, which may contain credentials.
type ResponseError struct {
	StatusCode int
	Message    string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("invalid API response (status %d): %s", e.StatusCode, e.Message)
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
	if len(params) == 0 {
		return template
	}
	replacements := make([]string, 0, len(params)*2)
	for key, value := range params {
		replacements = append(replacements, "{"+key+"}", url.PathEscape(value))
	}
	return strings.NewReplacer(replacements...).Replace(template)
}
