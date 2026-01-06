package cli

import (
	"github.com/stretchr/testify/assert"

	"hotaisle-cli/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func setupTestApp(t *testing.T) (*App, string) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &config.Config{
		ApiToken:    "",
		LogLevel:    "info",
		DefaultTeam: "",
	}

	// Create config directory
	configDir := filepath.Join(tmpDir, ".hotaisle")
	assert.NoError(t, os.MkdirAll(configDir, 0o700))

	app := &App{
		Config: cfg,
	}

	return app, tmpDir
}
