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

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, Directory)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, File)
	return path, nil
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
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
