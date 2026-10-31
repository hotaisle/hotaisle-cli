package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	Directory string = ".hotaisle"
	File      string = "config.json"
	Path      string = Directory + "/" + File
	Pretty    string = "~/" + Path
)

type Config struct {
	LogLevel    string `json:"log_level,omitempty" default:"info"`
	ApiToken    string `json:"api_token"`
	DefaultTeam string `json:"default_team"`
	path        string
}

func NewConfig() *Config {
	return &Config{
		LogLevel: "info",
	}
}

func defaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, Directory)
	return dir, nil
}

// ResolvePath expands the default path and a leading ~/ without writing files.
func ResolvePath(path string) (string, error) {
	if path == "" || path == Pretty {
		return defaultConfigPath()
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return filepath.Clean(path), nil
}

func defaultConfigPath() (string, error) {
	dir, err := defaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, File), nil
}

func Load(path *string) (*Config, error) {
	config := NewConfig()

	if path == nil {
		defaultPath, err := defaultConfigPath()
		if err != nil {
			slog.Debug("Failed to get config path, using defaults", "error", err)
			return nil, err
		}
		path = &defaultPath
	}
	resolved, err := ResolvePath(*path)
	if err != nil {
		return nil, err
	}
	config.path = resolved
	path = &resolved

	slog.Debug("Loading config", "path", *path)
	configData, err := os.ReadFile(*path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		slog.Debug("Failed to read config file", "path", *path, "error", err)
		return nil, err
	}
	if err := json.Unmarshal(configData, config); err != nil {
		slog.Debug("Failed to parse config", "path", *path, "error", err)
		return nil, err
	}
	return config, nil
}

func Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("nil config")
	}
	path := cfg.path
	var err error
	if path == "" {
		path, err = defaultConfigPath()
		if err != nil {
			return err
		}
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	defaultPath, err := defaultConfigPath()
	if err == nil && path == defaultPath {
		if err := os.Chmod(dir, 0o700); err != nil { // #nosec G302 -- directories require owner execute permission
			return err
		}
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeConfig(dir, filepath.Base(path), b)
}

func writeConfig(dir, name string, data []byte) (err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := root.Close(); err == nil {
			err = closeErr
		}
	}()
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	if err := file.Truncate(0); err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}
