package cli

import (
	"context"
	"errors"
	"testing"

	"hotaisle-cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

// getCommand looks up a command by path and sets flags for testing
func getCommand(app *App, def commandDef, path string, flags map[string]string) (*cli.Command, error) {
	cmdDef := def.findCommand(path)
	if cmdDef == nil {
		return nil, errors.New("command not found: " + path)
	}

	cmd := buildCommand(app, *cmdDef)

	// Set flags if provided
	for name, value := range flags {
		if err := cmd.Set(name, value); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

func executeCommand(t *testing.T, cmd *cli.Command) string {
	return test.CaptureStdout(t, func() error {
		return cmd.Action(context.Background(), cmd)
	})
}

// TestBuildCommandWithDefaultTeam tests that buildCommand modifies the team flag when DefaultTeam is set
func TestBuildCommandWithDefaultTeam(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.DefaultTeam = "default-team"

	// Create a commandDef with a required team flag
	def := commandDef{
		Name: "test",
		Flags: []flagDef{
			{Name: "team", Usage: "Team handle", Required: true},
			{Name: "vm", Usage: "VM name", Required: true},
		},
		Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}

	cmd := buildCommand(app, def)

	// Verify the team flag is no longer required
	teamFlag, ok := cmd.Flags[0].(*cli.StringFlag)
	assert.True(t, ok, "team flag should be a StringFlag")

	assert.False(t, teamFlag.Required, "team flag should not be required when DefaultTeam is set")
	assert.Equal(t, "default-team", teamFlag.Value, "team flag should use DefaultTeam value")
	assert.Contains(t, teamFlag.Usage, "uses default_team from config", "usage should indicate default_team usage")

	// Verify other flags are not affected
	vmFlag, ok := cmd.Flags[1].(*cli.StringFlag)
	assert.True(t, ok, "vm flag should be a StringFlag")
	assert.True(t, vmFlag.Required, "vm flag should still be required")
}

// TestBuildCommandWithoutDefaultTeam tests that a team flag remains required when DefaultTeam is not set
func TestBuildCommandWithoutDefaultTeam(t *testing.T) {
	app, _ := setupTestApp(t)
	// DefaultTeam is empty by default

	def := commandDef{
		Name: "test",
		Flags: []flagDef{
			{Name: "team", Usage: "Team handle", Required: true},
		},
		Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}

	cmd := buildCommand(app, def)

	teamFlag, ok := cmd.Flags[0].(*cli.StringFlag)
	assert.True(t, ok, "team flag should be a StringFlag")

	assert.True(t, teamFlag.Required, "team flag should be required when DefaultTeam is not set")
}

// TestBuildCommandFlagUsageUpdate tests that the usage text is correctly updated
func TestBuildCommandFlagUsageUpdate(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.DefaultTeam = "my-team"

	def := commandDef{
		Name: "test",
		Flags: []flagDef{
			{Name: "team", Usage: "Team handle", Required: true},
		},
		Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}

	cmd := buildCommand(app, def)

	teamFlag, ok := cmd.Flags[0].(*cli.StringFlag)
	assert.True(t, ok, "team flag should be a StringFlag")

	expectedUsage := "Team handle (uses default_team from config)"
	assert.Equal(t, expectedUsage, teamFlag.Usage, "usage should indicate default_team usage")
}

// TestBuildCommandMultipleFlags tests that all flags are correctly processed
func TestBuildCommandMultipleFlags(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.DefaultTeam = "default-team"

	def := commandDef{
		Name: "test",
		Flags: []flagDef{
			{Name: "team", Usage: "Team handle", Required: true},
			{Name: "vm", Usage: "VM name", Required: true},
			{Name: "user-data-url", Usage: "User data URL", Required: false, Value: ""},
		},
		Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}

	cmd := buildCommand(app, def)

	assert.Len(t, cmd.Flags, 3, "should have 3 flags")

	// Team flag
	teamFlag, ok := cmd.Flags[0].(*cli.StringFlag)
	assert.True(t, ok)
	assert.False(t, teamFlag.Required)
	assert.Equal(t, "default-team", teamFlag.Value)

	// VM flag
	vmFlag, ok := cmd.Flags[1].(*cli.StringFlag)
	assert.True(t, ok)
	assert.True(t, vmFlag.Required)

	// User data URL flag
	userDataFlag, ok := cmd.Flags[2].(*cli.StringFlag)
	assert.True(t, ok)
	assert.False(t, userDataFlag.Required)
	assert.Equal(t, "", userDataFlag.Value)
}

// TestBuildCommandNilConfig tests behavior when Config is nil
func TestBuildCommandNilConfig(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config = nil

	def := commandDef{
		Name: "test",
		Flags: []flagDef{
			{Name: "team", Usage: "Team handle", Required: true},
		},
		Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}

	cmd := buildCommand(app, def)

	teamFlag, ok := cmd.Flags[0].(*cli.StringFlag)
	assert.True(t, ok, "team flag should be a StringFlag")

	// When Config is nil, the flag should still be required
	assert.True(t, teamFlag.Required, "team flag should be required when Config is nil")
}

// TestVMCommandUsesDefaultTeam tests that VM commands respect DefaultTeam
func TestVMCommandUsesDefaultTeam(t *testing.T) {
	app, _ := setupTestApp(t)
	app.Config.DefaultTeam = "default-team"

	cmd, err := getCommand(app, virtualMachineCommands, "list", nil)
	assert.NoError(t, err)

	teamFlag, ok := cmd.Flags[0].(*cli.StringFlag)
	assert.True(t, ok, "team flag should be a StringFlag")

	assert.False(t, teamFlag.Required, "team flag should not be required when DefaultTeam is set")
	assert.Equal(t, "default-team", teamFlag.Value)
	assert.Contains(t, teamFlag.Usage, "uses default_team from config")
}

// TestVMCommandRequiresTeamWithoutDefault tests that VM commands require team when no default
func TestVMCommandRequiresTeamWithoutDefault(t *testing.T) {
	app, _ := setupTestApp(t)
	// DefaultTeam is empty by default

	cmd, err := getCommand(app, virtualMachineCommands, "list", nil)
	assert.NoError(t, err)

	teamFlag, ok := cmd.Flags[0].(*cli.StringFlag)
	assert.True(t, ok, "team flag should be a StringFlag")

	assert.True(t, teamFlag.Required, "team flag should be required when DefaultTeam is not set")
}
