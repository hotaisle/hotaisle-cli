package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"hotaisle-cli/client"
	"hotaisle-cli/internal/api"
	"hotaisle-cli/test"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	// Create a mock user response
	mockUser := &client.GetUserResponse{
		User: client.User{
			Name:    "Test User",
			Email:   "test@example.com",
			Created: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Teams: []client.UserTeam{{
			Team: client.Team{
				Handle:                  "test-team",
				Name:                    "Test Team",
				Description:             "",
				MaximumVirtualMachines:  0,
				MaximumBareMetalServers: 0,
			},
			Roles:          []string{"owner"},
			EffectiveRoles: []string{"operator"},
			Invitation:     true,
		}},
	}

	// Create a mock HTTP client
	mockClient := test.NewMockClient(test.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "/api/user/", req.URL.Path)
		assert.Equal(t, http.MethodGet, req.Method)

		body, err := json.Marshal(mockUser)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd := newCommandUser(app)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})
	slog.Info(output)

	// Verify the output is valid JSON and matches the expected structure
	var result client.GetUserResponse
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "output should be valid JSON")

	assert.Equal(t, "Test User", result.User.Name)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Len(t, result.Teams, 1)
	assert.Equal(t, "test-team", result.Teams[0].Handle)
}

func TestUserCommand_APIError(t *testing.T) {
	app, _ := setupTestApp(t)

	// Create mock HTTP client that returns an error
	mockClient := test.NewMockClient(test.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 401,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"unauthorized"}`))),
			Header:     make(http.Header),
		}, nil
	}))

	app.Client = api.NewClient("invalid-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd := newCommandUser(app)

	ctx := context.Background()
	err := cmd.Action(ctx, nil)
	slog.Info(err.Error())
	assert.Error(t, err, "should return error for failed API call")
}

func TestUserCommand_EmptyTeams(t *testing.T) {
	app, _ := setupTestApp(t)

	// Create mock user response with no teams
	mockUser := &client.GetUserResponse{
		User: client.User{
			Name:    "Solo User",
			Email:   "solo@example.com",
			Created: time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC),
		},
		Teams: []client.UserTeam{},
	}

	mockClient := test.NewMockClient(test.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, err := json.Marshal(mockUser)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd := newCommandUser(app)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	var result client.GetUserResponse
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "Solo User", result.User.Name)
	assert.Empty(t, result.Teams)
}

func TestUserCommand_Structure(t *testing.T) {
	app, _ := setupTestApp(t)
	cmd := newCommandUser(app)

	assert.Equal(t, "user", cmd.Name)
	assert.Equal(t, "Gets the current user.", cmd.Usage)
	assert.NotNil(t, cmd.Action)
}

func TestUserCommand_MultipleTeams(t *testing.T) {
	app, _ := setupTestApp(t)

	mockUser := &client.GetUserResponse{
		User: client.User{
			Name:    "Multi Team User",
			Email:   "multi@example.com",
			Created: time.Date(2024, 3, 10, 8, 15, 0, 0, time.UTC),
		},
		Teams: []client.UserTeam{
			{
				Team: client.Team{
					Handle:                  "team-one",
					Name:                    "Test Team",
					Description:             "",
					MaximumVirtualMachines:  0,
					MaximumBareMetalServers: 0,
				},
				Roles:          []string{"owner"},
				EffectiveRoles: []string{"operator"},
				Invitation:     true,
			},
			{
				Team: client.Team{
					Handle:                  "team-two",
					Name:                    "Test Team",
					Description:             "",
					MaximumVirtualMachines:  0,
					MaximumBareMetalServers: 0,
				},
				Roles:          []string{"owner"},
				EffectiveRoles: []string{"operator"},
				Invitation:     false,
			},
			{
				Team: client.Team{
					Handle:                  "team-three",
					Name:                    "Test Team",
					Description:             "",
					MaximumVirtualMachines:  0,
					MaximumBareMetalServers: 0,
				},
				Roles:          []string{"owner"},
				EffectiveRoles: []string{"operator"},
				Invitation:     true,
			},
		},
	}

	mockClient := test.NewMockClient(test.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, err := json.Marshal(mockUser)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd := newCommandUser(app)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	var result client.GetUserResponse
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "Multi Team User", result.User.Name)
	assert.Len(t, result.Teams, 3)
	assert.Equal(t, "team-one", result.Teams[0].Handle)
	assert.Equal(t, "team-two", result.Teams[1].Handle)
	assert.Equal(t, "team-three", result.Teams[2].Handle)
	assert.True(t, result.Teams[0].Invitation)
	assert.False(t, result.Teams[1].Invitation)
	assert.True(t, result.Teams[2].Invitation)
}
