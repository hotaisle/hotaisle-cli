package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hotaisle-cli/internal/config"

	"github.com/urfave/cli/v3"
)

func setupTestApp(t *testing.T) (*App, string) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &config.Config{
		ApiToken:    "",
		LogLevel:    "info",
		DefaultTeam: "",
	}

	app := &App{
		Config: cfg,
	}

	return app, tmpDir
}

func TestMakeApp(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	app, err := makeApp()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	assert.NotNil(t, app.Config)
	assert.GreaterOrEqual(t, len(app.Config.LogLevel), 0)

	assert.NotNil(t, app.Client)
	assert.NotNil(t, app.Client.Api)

	assert.NotNil(t, app.AppCli)
	assert.Equal(t, "Manage Hot Aisle resources from your terminal.", app.AppCli.Usage)
	assert.True(t, app.AppCli.EnableShellCompletion)
}

func TestMakeAppWithConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	restoreVersion := Version
	restoreCommit := Commit
	restoreBranch := Branch
	defer func() {
		Version = restoreVersion
		Commit = restoreCommit
		Branch = restoreBranch
	}()

	Version = "1.2.3"
	Commit = "abc123"
	Branch = "main"

	app, err := makeApp()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	versionInfo := app.AppCli.Version
	assert.Contains(t, versionInfo, "1.2.3")
	assert.Contains(t, versionInfo, "abc123")
	assert.Contains(t, versionInfo, "main")
	assert.Contains(t, versionInfo, "Go version")
}

func TestMakeAppCommands(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	app, err := makeApp()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	assert.NotNil(t, app.AppCli.Commands)
	assert.Len(t, app.AppCli.Commands, 5)

	expectedCommands := []string{"config", "user", "team", "bm", "vm"}
	commandNames := []string{}
	for _, cmd := range app.AppCli.Commands {
		commandNames = append(commandNames, cmd.Name)
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, commandNames, expected)
	}
}

func TestMakeAppConfigFileFlag(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	app, err := makeApp()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	assert.NotNil(t, app.AppCli.Flags)
	assert.Len(t, app.AppCli.Flags, 1)

	flag := app.AppCli.Flags[0]
	stringFlag, ok := flag.(*cli.StringFlag)
	assert.True(t, ok)
	assert.Equal(t, "config-file", stringFlag.Name)
	assert.Contains(t, stringFlag.Value, config.Pretty)
}

func TestMakeAppWithLogLevel(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	app, err := makeApp()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	assert.Equal(t, "info", app.Config.LogLevel)
}

func TestMakeAppWithInvalidConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configDir := filepath.Join(tmp, ".hotaisle")
	err := os.MkdirAll(configDir, 0o700)
	require.NoError(t, err)
	configPath := filepath.Join(configDir, "config.json")
	err = os.WriteFile(configPath, []byte("{invalid"), 0o600)
	require.NoError(t, err)

	app, err := makeApp()
	assert.Error(t, err)
	assert.Nil(t, app)
}

func TestMakeCommands(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, _ := config.Load(nil)
	app := &App{
		Config: cfg,
	}

	commands := makeCommands(app)
	assert.NotNil(t, commands)
	assert.Len(t, commands, 5)

	expectedCommands := []string{"config", "user", "team", "bm", "vm"}
	commandNames := []string{}
	for _, cmd := range commands {
		commandNames = append(commandNames, cmd.Name)
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, commandNames, expected)
	}
}

func TestMakeAppVersionFormat(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	restoreVersion := Version
	restoreCommit := Commit
	restoreBranch := Branch
	defer func() {
		Version = restoreVersion
		Commit = restoreCommit
		Branch = restoreBranch
	}()

	Version = "2.0.0"
	Commit = "def456"
	Branch = "release"

	app, err := makeApp()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	versionInfo := app.AppCli.Version
	versions := strings.Split(versionInfo, "\n")
	assert.GreaterOrEqual(t, len(versions), 3)

	assert.Contains(t, versionInfo, "2.0.0")
	assert.Contains(t, versionInfo, "def456")
	assert.Contains(t, versionInfo, "release")
}

func TestMakeAppDefaultValues(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	app, err := makeApp()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	assert.Equal(t, "help", app.AppCli.DefaultCommand)
	assert.True(t, app.AppCli.EnableShellCompletion)
}
