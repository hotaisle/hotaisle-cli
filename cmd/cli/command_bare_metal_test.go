package cli

import (
	"encoding/json"
	"hotaisle-cli/client"
	"hotaisle-cli/internal/api"
	"hotaisle-cli/test"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBareMetalListCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockServers := []client.BareMetalServerDetails{
		{
			BareMetalServer: client.BareMetalServer{
				Name: "server-1",
			},
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/", http.MethodGet, 200, mockServers)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{"team": "test-team"}
	cmd, err := getCommand(app, bareMetalCommands, "list", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)

	var result []client.BareMetalServerDetails
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "server-1", result[0].Name)
}

func TestBareMetalGetCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockServer := &client.BareMetalServerDetails{
		BareMetalServer: client.BareMetalServer{
			Name: "server-1",
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/", http.MethodGet, 200, mockServer)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "get", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.BareMetalServerDetails
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Equal(t, "server-1", result.Name)
}

func TestBareMetalReserveCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockResp := &client.BareMetalServerReservationResponse{
		BareMetalServer: client.BareMetalServer{
			Name: "server-1",
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/", http.MethodPost, 200, mockResp)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":        "test-team",
		"description": "test-desc",
		"cpu-cores":   "8",
		"ram-gb":      "16",
		"disk-gb":     "100",
	}
	cmd, err := getCommand(app, bareMetalCommands, "reserve", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.BareMetalServerReservationResponse
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Equal(t, "server-1", result.Name)
}

func TestBareMetalUpdateCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/", http.MethodPatch, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":        "test-team",
		"server":      "server-1",
		"description": "new-desc",
	}
	cmd, err := getCommand(app, bareMetalCommands, "update", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Server updated successfully")
}

func TestBareMetalDeleteCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/", http.MethodDelete, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "delete", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Server deleted successfully")
}

func TestBareMetalAvailableCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockAvailable := []client.AvailableBareMetalTypes{
		{
			Quantity: 5,
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/available/", http.MethodGet, 200, mockAvailable)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{"team": "test-team"}
	cmd, err := getCommand(app, bareMetalCommands, "available", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)

	var result []client.AvailableBareMetalTypes
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(5), result[0].Quantity)
}

func TestBareMetalPowerStatusCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockState := &client.BareMetalServerPowerState{
		State: "on",
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/power/", http.MethodGet, 200, mockState)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "power.status", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.BareMetalServerPowerState
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Equal(t, "on", result.State)
}

func TestBareMetalPowerOnCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/power/power_on/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "power.on", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Power on command sent")
}

func TestBareMetalPowerShutdownCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/power/graceful_shutdown/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "power.shutdown", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Graceful shutdown command sent")
}

func TestBareMetalPowerForceShutdownCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/power/force_shutdown/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "power.force-shutdown", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Force shutdown command sent")
}

func TestBareMetalPowerRebootCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/power/warm_reboot/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "power.reboot", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Warm reboot command sent")
}

func TestBareMetalPowerColdRebootCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/power/cold_reboot/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "power.cold-reboot", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Cold reboot command sent")
}

func TestBareMetalPowerACResetCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/power/ac_reset/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "power.ac-reset", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "AC reset command sent")
}

func TestBareMetalReinstallCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockServer := &client.BareMetalServerDetails{
		BareMetalServer: client.BareMetalServer{
			Name: "server-1",
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/reinstall/", http.MethodPost, 200, mockServer)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "reinstall", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.BareMetalServerDetails
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Equal(t, "server-1", result.Name)
}

func TestBareMetalConsoleCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockConsole := &client.BareMetalServerConsoleURL{
		URL: "http://console.example.com",
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/console/", http.MethodPost, 200, mockConsole)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "console", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.BareMetalServerConsoleURL
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Equal(t, "http://console.example.com", result.URL)
}

func TestBareMetalSupportAccessEnableCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/support_access_enable/", http.MethodPut, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "support-access.enable", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Support access enabled")
}

func TestBareMetalSupportAccessDisableCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/bare_metal/server-1/support_access_enable/", http.MethodDelete, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":   "test-team",
		"server": "server-1",
	}
	cmd, err := getCommand(app, bareMetalCommands, "support-access.disable", flags)
	require.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "Support access disabled")
}
