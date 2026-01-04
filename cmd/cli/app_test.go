package cli

import (
	"hotaisle-cli/internal/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, os.MkdirAll(configDir, 0o700))

	app := &App{
		Config: cfg,
	}

	return app, tmpDir
}
