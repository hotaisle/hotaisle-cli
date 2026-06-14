package cli

import (
	"encoding/json"
	"hotaisle-cli/client"
	"hotaisle-cli/internal/api"
	"hotaisle-cli/test"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeamListCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockTeams := []client.UserTeam{
		{
			Team: client.Team{
				Handle: "team-1",
				Name:   "Team 1",
			},
			Roles: []string{"owner"},
		},
		{
			Team: client.Team{
				Handle: "team-2",
				Name:   "Team 2",
			},
			Roles: []string{"member"},
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/", http.MethodGet, 200, mockTeams)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	cmd, err := getCommand(app, teamCommands, "list", nil)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result []client.UserTeam
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "team-1", result[0].Handle)
}

func TestTeamCreateCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockTeam := &client.UserTeamWithMembers{
		UserTeam: client.UserTeam{
			Team: client.Team{
				Handle:      "new-team",
				Name:        "New Team",
				Description: "A new team",
			},
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/", http.MethodPost, 200, mockTeam)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"handle": "new-team",
		"name":   "New Team",
	}
	cmd, err := getCommand(app, teamCommands, "create", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.UserTeamWithMembers
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, "new-team", result.Handle)
}

func TestTeamGetCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockTeam := &client.UserTeamDetails{
		UserTeamWithMembers: client.UserTeamWithMembers{
			UserTeam: client.UserTeam{
				Team: client.Team{
					Handle: "test-team",
					Name:   "Test Team",
				},
			},
			Members: []client.TeamMember{
				{Name: "User 1", Email: "user1@example.com"},
			},
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/", http.MethodGet, 200, mockTeam)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"handle": "test-team",
	}
	cmd, err := getCommand(app, teamCommands, "get", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.UserTeamDetails
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, "test-team", result.Handle)
	assert.Len(t, result.Members, 1)
}

func TestTeamBalanceCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockBalance := &client.BalanceInfo{
		AvailableBalance:     1000,
		HourlyRate:           10,
		BareMetalServerCount: 2,
		VirtualMachineCount:  3,
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/balance/", http.MethodGet, 200, mockBalance)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"handle": "test-team",
	}
	cmd, err := getCommand(app, teamCommands, "balance", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.BalanceInfo
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), result.AvailableBalance)
	assert.Equal(t, int64(2), result.BareMetalServerCount)
	assert.Equal(t, int64(3), result.VirtualMachineCount)
}

func TestTeamMembersInviteCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/members/invitations/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"handle": "test-team",
		"email":  "new@example.com",
		"name":   "New User",
		"role":   "member",
	}
	cmd, err := getCommand(app, teamCommands, "members.invite", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	assert.Contains(t, output, "Invitation sent successfully")
}

func TestTeamMembersRemoveCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/members/user@example.com/", http.MethodDelete, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"handle": "test-team",
		"email":  "user@example.com",
	}
	cmd, err := getCommand(app, teamCommands, "members.remove", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	assert.Contains(t, output, "Member removed successfully")
}
