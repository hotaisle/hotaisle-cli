package config

import (
	"hotaisle-cli/internal/log"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFirstTime(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// There should be an error since the config file doesn't exist in the temp directory
	cfg, err := Load("")
	assert.NotNil(t, err)
	assert.NotNil(t, cfg)

	// Config file should not exist after loading fresh
	_, err = os.Stat(filepath.Join(tmp, ".hotaisle", "config.json"))
	assert.NotNil(t, err)
}

func TestLoadFirstTimeAndSave(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	slog.SetDefault(log.New(log.LevelDebug))

	// There should be an error since the config file doesn't exist in the temp directory
	cfg, err := Load("")
	assert.NotNil(t, err)
	assert.NotNil(t, cfg)

	apiToken := "abc123"
	cfg.ApiToken = apiToken
	logLevel := "warn"
	cfg.LogLevel = logLevel
	defaultTeam := "devs"
	cfg.DefaultTeam = defaultTeam

	err = Save(cfg)
	assert.Nil(t, err)

	cfg, err = Load("")
	assert.NotNil(t, cfg)
	assert.Nil(t, err)

	assert.Equal(t, apiToken, cfg.ApiToken)
	assert.Equal(t, logLevel, cfg.LogLevel)
	assert.Equal(t, defaultTeam, cfg.DefaultTeam)
}

// TestLoadInvalidJSON verifies that if the config file contains invalid JSON,
// Load() should return an error instead of silently succeeding.
func TestLoadInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Write invalid JSON to the config file
	dir := filepath.Join(tmp, ".hotaisle")
	_ = os.MkdirAll(dir, 0o700)
	path := filepath.Join(dir, "config.json")
	err := os.WriteFile(path, []byte("{invalid"), 0o600)
	assert.Nil(t, err)

	_, err = Load("")
	assert.NotNil(t, err)
}
