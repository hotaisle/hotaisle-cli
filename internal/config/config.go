package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil { // #nosec G302 -- directories require owner execute permission
		return "", err
	}
	return dir, nil
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

	slog.Debug("Loading config", "path", *path)
	configData, err := os.ReadFile(*path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("Config file does not exist, saving defaults to file", "path", *path)
			err := Save(config)
			if err != nil {
				return nil, err
			}
			return config, nil
		}
		slog.Debug("Failed to read config file", "path", *path, "error", err)
		return nil, err
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		slog.Debug("Failed to parse config", "path", *path, "error", err)
		return nil, err
	}
	return config, nil
}

func Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("nil config")
	}
	dir, err := defaultConfigDir()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeConfig(dir, b)
}

func writeConfig(dir string, data []byte) (err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := root.Close(); err == nil {
			err = closeErr
		}
	}()
	file, err := root.OpenFile(File, os.O_WRONLY|os.O_CREATE, 0o600)
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
