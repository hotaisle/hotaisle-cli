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

func TestVMListCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockVMs := []client.VirtualMachineDetails{
		{
			VirtualMachine: client.VirtualMachine{
				Name: "vm-1",
			},
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/", http.MethodGet, 200, mockVMs)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{"team": "test-team"}
	cmd, err := getCommand(app, virtualMachineCommands, "list", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result []client.VirtualMachineDetails
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "vm-1", result[0].Name)
}

func TestVMGetCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockVM := &client.VirtualMachineDetails{
		VirtualMachine: client.VirtualMachine{
			Name: "vm-1",
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/", http.MethodGet, 200, mockVM)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "get", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.VirtualMachineDetails
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, "vm-1", result.Name)
}

func TestVMProvisionCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockVM := &client.VirtualMachineDetails{
		VirtualMachine: client.VirtualMachine{
			Name: "vm-1",
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/", http.MethodPost, 200, mockVM)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":          "test-team",
		"cpu-cores":     "2",
		"ram-gb":        "4",
		"disk-gb":       "20",
		"user-data-url": "http://example.com/user-data",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "provision", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.VirtualMachineDetails
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, "vm-1", result.Name)
}

func TestVMUpdateCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/", http.MethodPatch, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":        "test-team",
		"vm":          "vm-1",
		"description": "new-desc",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "update", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM updated successfully")
}

func TestVMDeleteCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/", http.MethodDelete, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "delete", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM deleted successfully")
}

func TestVMAvailableCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockAvailable := []client.AvailableVirtualMachineTypes{
		{
			Quantity: 10,
		},
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/available/", http.MethodGet, 200, mockAvailable)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{"team": "test-team"}
	cmd, err := getCommand(app, virtualMachineCommands, "available", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result []client.AvailableVirtualMachineTypes
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(10), result[0].Quantity)
}

func TestVMStateCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockState := &client.VirtualMachineState{
		State: "running",
	}

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/state/", http.MethodGet, 200, mockState)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "state", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)

	var result client.VirtualMachineState
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, "running", result.State)
}

func TestVMStartCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/start/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "start", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM start command sent")
}

func TestVMStopCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/stop/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "stop", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM stop command sent")
}

func TestVMShutdownCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/shutdown/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "shutdown", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM shutdown command sent")
}

func TestVMRebootCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/reboot/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "reboot", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM reboot command sent")
}

func TestVMHardResetCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/hard-reset/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team": "test-team",
		"vm":   "vm-1",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "hard-reset", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM hard-reset command sent")
}

func TestVMRebuildCommand_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	mockClient := test.NewMockHTTPClientWithAssertions(t, "/api/teams/test-team/virtual_machines/vm-1/rebuild/", http.MethodPost, 200, nil)
	app.Client = api.NewClient("test-token", "1.0.0", client.WithHTTPClient(mockClient))

	flags := map[string]string{
		"team":          "test-team",
		"vm":            "vm-1",
		"user-data-url": "http://example.com/new-user-data",
	}
	cmd, err := getCommand(app, virtualMachineCommands, "rebuild", flags)
	assert.NoError(t, err)

	output := executeCommand(t, cmd)
	assert.Contains(t, output, "VM rebuild command sent")
}
