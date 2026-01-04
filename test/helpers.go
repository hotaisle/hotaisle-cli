package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

func NewMockClient(rtf http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: rtf,
	}
}

// NewMockHTTPClient creates a mock HTTP client with a simple handler function
func NewMockHTTPClient(handler func(req *http.Request) (*http.Response, error)) *http.Client {
	return NewMockClient(RoundTripFunc(handler))
}

// NewMockHTTPClientWithAssertions creates a mock HTTP client that validates the request and returns a response
// If data is nil, an empty response is returned. Otherwise, a JSON response is returned.
func NewMockHTTPClientWithAssertions(t testing.TB, expectedPath, expectedMethod string, statusCode int, data interface{}) *http.Client {
	return NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if expectedPath != "" && req.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, req.URL.Path)
		}
		if expectedMethod != "" && req.Method != expectedMethod {
			t.Errorf("Expected method %s, got %s", expectedMethod, req.Method)
		}

		if data == nil {
			return NewEmptyResponse(statusCode), nil
		}
		return NewJSONResponse(t, statusCode, data), nil
	})
}

func NewOkResponse() *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
}

// NewJSONResponse creates an HTTP response with a JSON body
func NewJSONResponse(t testing.TB, statusCode int, data interface{}) *http.Response {
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

// NewEmptyResponse creates an HTTP response with an empty body
func NewEmptyResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
		Header:     make(http.Header),
	}
}

// RoundTripFunc is a helper type for creating mock HTTP transports
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func CaptureStdout(t *testing.T, fn func() error) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = fn()
	if err != nil {
		t.Fatal(err)
	}

	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatal(err)
	}

	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}

	return buf.String()
}
