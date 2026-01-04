package cli

import (
	"context"
	"hotaisle-cli/test"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestConfigSetToken_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	// Set the environment variable
	testToken := "test-token-12345"
	t.Setenv("HOTAISLE_API_TOKEN", testToken)

	cmd := newCommandConfig(app)

	// Find the "set token" command
	setCmd := cmd.Commands[0]      // "set"
	tokenCmd := setCmd.Commands[0] // "token"

	ctx := context.Background()
	err := tokenCmd.Action(ctx, nil)

	assert.NoError(t, err)
	assert.Equal(t, testToken, app.Config.ApiToken)
}

func TestConfigSetToken_MissingEnvVar(t *testing.T) {
	app, _ := setupTestApp(t)

	// Make sure the environment variable is not set
	err := os.Unsetenv("HOTAISLE_API_TOKEN")
	if err != nil {
		assert.FailNow(t, "failed to unset environment variable")
	}

	cmd := newCommandConfig(app)

	// Find the "set token" command
	setCmd := cmd.Commands[0]      // "set"
	tokenCmd := setCmd.Commands[0] // "token"

	ctx := context.Background()
	err = tokenCmd.Action(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token, set HOTAISLE_API_TOKEN")
	assert.Empty(t, app.Config.ApiToken)
}

func TestConfigSetToken_EmptyEnvVar(t *testing.T) {
	app, _ := setupTestApp(t)

	// Set an empty environment variable
	t.Setenv("HOTAISLE_API_TOKEN", "")

	cmd := newCommandConfig(app)

	// Find the "set token" command
	setCmd := cmd.Commands[0]      // "set"
	tokenCmd := setCmd.Commands[0] // "token"

	ctx := context.Background()
	err := tokenCmd.Action(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token, set HOTAISLE_API_TOKEN")
	assert.Empty(t, app.Config.ApiToken)
}

func TestConfigSetToken_WhitespaceEnvVar(t *testing.T) {
	app, _ := setupTestApp(t)

	// Set whitespace-only environment variable
	t.Setenv("HOTAISLE_API_TOKEN", "   \t\n  ")

	cmd := newCommandConfig(app)

	// Find the "set token" command
	setCmd := cmd.Commands[0]      // "set"
	tokenCmd := setCmd.Commands[0] // "token"

	ctx := context.Background()
	err := tokenCmd.Action(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token, set HOTAISLE_API_TOKEN")
	assert.Empty(t, app.Config.ApiToken)
}

func TestConfigSetLogLevel_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	cmd := newCommandConfig(app)

	// Verify the command structure
	setCmd := cmd.Commands[0]         // "set"
	logLevelCmd := setCmd.Commands[1] // "log-level"
	assert.Equal(t, "log-level", logLevelCmd.Name)

	// Execute the full command with arguments
	app.AppCli = &cli.Command{
		Commands: []*cli.Command{cmd},
	}

	ctx := context.Background()
	err := app.AppCli.Run(ctx, []string{"app", "config", "set", "log-level", "some random log level"})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "some random log level", app.Config.LogLevel)
}

func TestConfigSetDefaultTeam_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	cmd := newCommandConfig(app)

	// Find the "set default-team" command
	setCmd := cmd.Commands[0]            // "set"
	defaultTeamCmd := setCmd.Commands[2] // "default-team"
	assert.Equal(t, "default-team", defaultTeamCmd.Name)

	// Execute the full command with arguments
	app.AppCli = &cli.Command{
		Commands: []*cli.Command{cmd},
	}

	ctx := context.Background()
	err := app.AppCli.Run(ctx, []string{"app", "config", "set", "default-team", "some default team"})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "some default team", app.Config.DefaultTeam)
}

func TestConfigGetToken(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.ApiToken = "test-token-get"

	cmd := newCommandConfig(app)

	// Find the "get token" command
	getCmd := cmd.Commands[1]      // "get"
	tokenCmd := getCmd.Commands[0] // "token"

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return tokenCmd.Action(ctx, nil)
	})

	assert.Equal(t, "test-token-get", output)
}

func TestConfigGetLogLevel(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.LogLevel = "warn"

	cmd := newCommandConfig(app)

	// Find the "get log-level" command
	getCmd := cmd.Commands[1]         // "get"
	logLevelCmd := getCmd.Commands[1] // "log-level"

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return logLevelCmd.Action(ctx, nil)
	})

	assert.Equal(t, "warn", output)
}

func TestConfigGetDefaultTeam(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.DefaultTeam = "test-team"

	cmd := newCommandConfig(app)

	// Find the "get default-team" command
	getCmd := cmd.Commands[1]            // "get"
	defaultTeamCmd := getCmd.Commands[2] // "default-team"

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return defaultTeamCmd.Action(ctx, nil)
	})

	assert.Equal(t, "test-team", output)
}

func TestConfigCommandStructure(t *testing.T) {
	app, _ := setupTestApp(t)
	cmd := newCommandConfig(app)

	// Test top-level command
	assert.Equal(t, "config", cmd.Name)
	assert.Equal(t, "Config File Management", cmd.Usage)
	assert.Len(t, cmd.Commands, 2) // "set" and "get"

	// Test "set" command
	setCmd := cmd.Commands[0]
	assert.Equal(t, "set", setCmd.Name)
	assert.Len(t, setCmd.Commands, 3) // "token", "log-level", "default-team"

	// Test "get" command
	getCmd := cmd.Commands[1]
	assert.Equal(t, "get", getCmd.Name)
	assert.Len(t, getCmd.Commands, 3) // "token", "log-level", "default-team"
}
