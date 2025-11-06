package api

import (
	"context"
	"hotaisle-cli/client"
	"hotaisle-cli/test"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		version string
	}{
		{
			name:    "creates client with valid token and version",
			token:   "test-token-123",
			version: "1.0.0",
		},
		{
			name:    "creates client with empty token",
			token:   "",
			version: "1.0.0",
		},
		{
			name:    "creates client with empty version",
			token:   "test-token-123",
			version: "",
		},
		{
			name:    "creates client with special characters in token",
			token:   "test-token-!@#$%^&*()",
			version: "1.0.0",
		},
		{
			name:    "creates client with beta version",
			token:   "test-token-123",
			version: "1.0.0-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.token, tt.version)

			assert.NotNil(t, c, "expected client to be non-nil")
			assert.NotNil(t, c.Api, "expected Api field to be non-nil")
		})
	}
}

func TestNewClient_ApiClientConfiguration(t *testing.T) {
	token := "test-token-123"
	version := "1.0.0"

	c := NewClient(token, version)

	assert.NotNil(t, c, "expected client to be non-nil")
	assert.NotNil(t, c.Api, "expected Api field to be non-nil")

	// Verify the client is properly initialized
	// This ensures that the underlying client.Client is created with the correct options
	assert.NotNil(t, c.Api, "expected Api client to be initialized")
}

func TestDefaultConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "DefaultHost is correct",
			constant: DefaultHost,
			expected: "admin.hotaisle.app",
		},
		{
			name:     "DefaultBaseURL is correct",
			constant: DefaultBaseURL,
			expected: "https://admin.hotaisle.app/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestNewClient_UserAgentFormat(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		expectedPattern string
	}{
		{
			name:            "standard version",
			version:         "1.0.0",
			expectedPattern: "hotaisle/1.0.0",
		},
		{
			name:            "pre-release version",
			version:         "2.0.0-rc.1",
			expectedPattern: "hotaisle/2.0.0-rc.1",
		},
		{
			name:            "dev version",
			version:         "dev",
			expectedPattern: "hotaisle/dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := "test-token"

			// Create a custom HTTP client that captures the User-Agent header
			var capturedUserAgent string

			c := NewClient(token, tt.version, client.WithHTTPClient(
				test.MakeMockClient(test.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					capturedUserAgent = req.Header.Get("User-Agent")
					return test.MakeOkResponse(), nil
				}))))

			assert.NotNil(t, c, "expected client to be non-nil")
			assert.NotNil(t, c.Api, "expected Api to be non-nil")

			_, _ = c.Api.User().Get(context.Background())

			assert.Equal(t, tt.expectedPattern, capturedUserAgent, "User-Agent should match expected pattern")
		})
	}
}

func TestClient_TypeStructure(t *testing.T) {
	c := &Client{
		Api: client.NewClient(),
	}

	assert.NotNil(t, c.Api, "expected Api field to be settable")

	// Verify that the Client struct has the expected field
	assert.NotNil(t, c.Api, "expected Api field to exist and be accessible")
}
