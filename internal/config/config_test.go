package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"hotaisle-cli/internal/log"

	"github.com/stretchr/testify/assert"
)

func TestLoadFirstTime(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := Load(nil)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "info", cfg.LogLevel)

	// Config file should exist after loading fresh
	_, err = os.Stat(filepath.Join(tmp, ".hotaisle", "config.json"))
	assert.Nil(t, err)
}

func TestLoadFirstTimeAndSave(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	slog.SetDefault(log.New(log.LevelDebug))

	cfg, err := Load(nil)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	apiToken := "abc123"
	cfg.ApiToken = apiToken
	logLevel := "warn"
	cfg.LogLevel = logLevel
	defaultTeam := "devs"
	cfg.DefaultTeam = defaultTeam

	err = Save(cfg)
	assert.Nil(t, err)

	cfg, err = Load(nil)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

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

	cfg, err := Load(nil)
	assert.Nil(t, cfg)
	assert.NotNil(t, err)
}

func TestLoadCustomPath(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "custom-config.json")
	customContent := `{
		"log_level": "debug",
		"api_token": "custom-token",
		"default_team": "custom-team"
	}`

	err := os.WriteFile(customPath, []byte(customContent), 0o600)
	assert.Nil(t, err)

	cfg, err := Load(&customPath)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "custom-token", cfg.ApiToken)
	assert.Equal(t, "custom-team", cfg.DefaultTeam)
}

func TestLoadPartialConfig(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "partial-config.json")

	partialContent := `{"api_token": "partial-token"}`
	err := os.WriteFile(customPath, []byte(partialContent), 0o600)
	assert.Nil(t, err)

	cfg, err := Load(&customPath)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "partial-token", cfg.ApiToken)
	assert.Empty(t, cfg.DefaultTeam)
}

func TestLoadEmptyConfig(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "empty-config.json")

	err := os.WriteFile(customPath, []byte("{}"), 0o600)
	assert.Nil(t, err)

	cfg, err := Load(&customPath)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Empty(t, cfg.ApiToken)
	assert.Empty(t, cfg.DefaultTeam)
}

func TestLoadUnreadableFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".hotaisle")
	_ = os.MkdirAll(dir, 0o700)
	path := filepath.Join(dir, "config.json")

	err := os.WriteFile(path, []byte(`{"log_level": "debug"}`), 0o000)
	assert.Nil(t, err)

	cfg, err := Load(nil)
	assert.Nil(t, cfg)
	assert.NotNil(t, err)
	assert.True(t, os.IsPermission(err))
}

func TestLoadMalformedJSONFields(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "malformed-config.json")

	malformedContent := `{"log_level": 12345}`
	err := os.WriteFile(customPath, []byte(malformedContent), 0o600)
	assert.Nil(t, err)

	cfg, err := Load(&customPath)
	assert.Nil(t, cfg)
	assert.NotNil(t, err)
}

func TestLoadComplexConfig(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "complex-config.json")

	complexContent := `{
		"log_level": "error",
		"api_token": "very-long-api-token-with-special-chars:!@#$%^&*()",
		"default_team": "team_123/subteam"
	}`
	err := os.WriteFile(customPath, []byte(complexContent), 0o600)
	assert.Nil(t, err)

	cfg, err := Load(&customPath)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "error", cfg.LogLevel)
	assert.Equal(t, "very-long-api-token-with-special-chars:!@#$%^&*()", cfg.ApiToken)
	assert.Equal(t, "team_123/subteam", cfg.DefaultTeam)
}
