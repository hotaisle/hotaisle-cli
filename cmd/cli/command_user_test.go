package cli

import (
	"context"
	"encoding/json"
	"hotaisle-cli/client"
	"hotaisle-cli/internal/api"
	"hotaisle-cli/test"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestUserGetCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

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

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/", http.MethodGet, 200, mockUser)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "get", nil)
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})
	slog.Info(output)

	var result client.GetUserResponse
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "output should be valid JSON")

	assert.Equal(t, "Test User", result.User.Name)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Len(t, result.Teams, 1)
	assert.Equal(t, "test-team", result.Teams[0].Handle)
}

func TestUserGetCommand_APIError(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "", "", 401, nil)
	app.Client = api.NewClient("invalid-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "get", nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = cmd.Action(ctx, nil)
	slog.Info(err.Error())
	assert.Error(t, err, "should return error for failed API call")
}

func TestUserGetCommand_EmptyTeams(t *testing.T) {
	app, _ := setupTestApp(t)

	mockUser := &client.GetUserResponse{
		User: client.User{
			Name:    "Solo User",
			Email:   "solo@example.com",
			Created: time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC),
		},
		Teams: []client.UserTeam{},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "", "", 200, mockUser)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "get", nil)
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	var result client.GetUserResponse
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "Solo User", result.User.Name)
	assert.Empty(t, result.Teams)
}

func TestUserCommand_Structure(t *testing.T) {
	app, _ := setupTestApp(t)
	cmd := newCommandUser(app)

	assert.Equal(t, "user", cmd.Name)
	assert.Equal(t, "Manage user account.", cmd.Usage)
	assert.NotNil(t, cmd.Commands)
	assert.Len(t, cmd.Commands, 4)
}

func TestUserGetCommand_MultipleTeams(t *testing.T) {
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

	mockClient := test.NewMockHTTPClientWithAssertions(t, "", "", 200, mockUser)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "get", nil)
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	var result client.GetUserResponse
	err = json.Unmarshal([]byte(output), &result)
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

func TestUserUpdateCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockUser := &client.User{
		Name:    "Updated Name",
		Email:   "test@example.com",
		Created: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/", http.MethodPatch, 200, mockUser)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "update", map[string]string{"name": "Updated Name"})
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, cmd)
	})

	var result client.User
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", result.Name)
}

func TestUserSSHKeysListCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockKeys := []client.SSHKey{
		{
			Type:        "ssh-rsa",
			PublicKey:   "AAAAB3NzaC1yc2EAAA...",
			Fingerprint: "SHA256:abc123",
			Comment:     "user@example.com",
		},
		{
			Type:        "ssh-ed25519",
			PublicKey:   "AAAAC3NzaC1lZDI1NTE5...",
			Fingerprint: "SHA256:def456",
			Comment:     "user@work.com",
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/ssh_keys/", http.MethodGet, 200, mockKeys)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "ssh-keys.list", nil)
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	var result []client.SSHKey
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "SHA256:abc123", result[0].Fingerprint)
	assert.Equal(t, "SHA256:def456", result[1].Fingerprint)
}

func TestUserSSHKeysAddCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockKey := &client.SSHKey{
		Type:        "ssh-rsa",
		PublicKey:   "AAAAB3NzaC1yc2EAAA...",
		Fingerprint: "SHA256:abc123",
		Comment:     "user@example.com",
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/ssh_keys/", http.MethodPost, 200, mockKey)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "ssh-keys.add", map[string]string{
		"key": "ssh-rsa AAAAB3NzaC1yc2EAAA... user@example.com",
	})
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, cmd)
	})

	var result client.SSHKey
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "SHA256:abc123", result.Fingerprint)
}

func TestUserSSHKeysDeleteCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/ssh_keys/SHA256:abc123/", http.MethodDelete, 204, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "ssh-keys.delete", map[string]string{"fingerprint": "SHA256:abc123"})
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, cmd)
	})

	assert.Contains(t, output, "SSH key deleted successfully")
}

func TestUserAPIKeysListCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockKeys := []client.UserAPIKey{
		{
			Prefix:   "abc123",
			Label:    "Development key",
			UserRole: "owner",
			Teams:    []client.APIKeyTeam{},
		},
		{
			Prefix:   "def456",
			Label:    "Production key",
			UserRole: "user",
			Teams:    []client.APIKeyTeam{},
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/api_keys/", http.MethodGet, 200, mockKeys)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "api-keys.list", nil)
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	var result []client.UserAPIKey
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "abc123", result[0].Prefix)
	assert.Equal(t, "def456", result[1].Prefix)
}

func TestUserAPIKeysGetCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockKey := &client.UserAPIKey{
		Prefix:   "abc123",
		Label:    "Development key",
		UserRole: "owner",
		Teams:    []client.APIKeyTeam{},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/api_keys/abc123/", http.MethodGet, 200, mockKey)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "api-keys.get", map[string]string{"prefix": "abc123"})
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, cmd)
	})

	var result client.UserAPIKey
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "abc123", result.Prefix)
	assert.Equal(t, "Development key", result.Label)
}

func TestUserAPIKeysCreateCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockKey := &client.UserAPIKeyWithToken{
		UserAPIKey: client.UserAPIKey{
			Prefix:   "abc123",
			Label:    "New key",
			UserRole: "user",
			Teams:    []client.APIKeyTeam{},
		},
		Token: "abc123.full-token-here",
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/api_keys/", http.MethodPost, 200, mockKey)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "api-keys.create", map[string]string{
		"label":     "New key",
		"user-role": "user",
	})
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, cmd)
	})

	var result client.UserAPIKeyWithToken
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "abc123", result.Prefix)
	assert.Equal(t, "abc123.full-token-here", result.Token)
}

func TestUserAPIKeysUpdateCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockKey := &client.UserAPIKey{
		Prefix:   "abc123",
		Label:    "Updated key",
		UserRole: "owner",
		Teams:    []client.APIKeyTeam{},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/api_keys/abc123/", http.MethodPatch, 200, mockKey)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "api-keys.update", map[string]string{
		"prefix":    "abc123",
		"label":     "Updated key",
		"user-role": "owner",
	})
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, cmd)
	})

	var result client.UserAPIKey
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "abc123", result.Prefix)
	assert.Equal(t, "Updated key", result.Label)
}

func TestUserAPIKeysDeleteCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/user/api_keys/abc123/", http.MethodDelete, 204, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, userCommands, "api-keys.delete", map[string]string{"prefix": "abc123"})
	require.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, cmd)
	})

	assert.Contains(t, output, "API key deleted successfully")
}
