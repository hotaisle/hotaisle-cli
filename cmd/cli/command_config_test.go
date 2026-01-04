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

	testToken := "test-token-12345"
	t.Setenv("HOTAISLE_API_TOKEN", testToken)

	cmd, err := getCommand(app, configCommands, "set.token", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	err = cmd.Action(ctx, nil)

	assert.NoError(t, err)
	assert.Equal(t, testToken, app.Config.ApiToken)
}

func TestConfigSetToken_MissingEnvVar(t *testing.T) {
	app, _ := setupTestApp(t)

	err := os.Unsetenv("HOTAISLE_API_TOKEN")
	if err != nil {
		assert.FailNow(t, "failed to unset environment variable")
	}

	cmd, err := getCommand(app, configCommands, "set.token", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	err = cmd.Action(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token, set HOTAISLE_API_TOKEN")
	assert.Empty(t, app.Config.ApiToken)
}

func TestConfigSetToken_EmptyEnvVar(t *testing.T) {
	app, _ := setupTestApp(t)

	t.Setenv("HOTAISLE_API_TOKEN", "")

	cmd, err := getCommand(app, configCommands, "set.token", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	err = cmd.Action(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token, set HOTAISLE_API_TOKEN")
	assert.Empty(t, app.Config.ApiToken)
}

func TestConfigSetToken_WhitespaceEnvVar(t *testing.T) {
	app, _ := setupTestApp(t)

	t.Setenv("HOTAISLE_API_TOKEN", "   \t\n  ")

	cmd, err := getCommand(app, configCommands, "set.token", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	err = cmd.Action(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token, set HOTAISLE_API_TOKEN")
	assert.Empty(t, app.Config.ApiToken)
}

func TestConfigSetLogLevel_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	cmd := newCommandConfig(app)

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

	cmd, err := getCommand(app, configCommands, "get.token", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	assert.Equal(t, "test-token-get", output)
}

func TestConfigGetLogLevel(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.LogLevel = "warn"

	cmd, err := getCommand(app, configCommands, "get.log-level", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	assert.Equal(t, "warn", output)
}

func TestConfigGetDefaultTeam(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.DefaultTeam = "test-team"

	cmd, err := getCommand(app, configCommands, "get.default-team", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	output := test.CaptureStdout(t, func() error {
		return cmd.Action(ctx, nil)
	})

	assert.Equal(t, "test-team", output)
}

func TestConfigCommandStructure(t *testing.T) {
	app, _ := setupTestApp(t)
	cmd := newCommandConfig(app)

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
