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

func LoadDefault() (*Config, error) {
	return Load("")
}

func Load(path string) (*Config, error) {
	if len(path) == 0 {
		defaultPath, err := defaultConfigPath()
		if err != nil {
			return nil, err
		}
		path = defaultPath
	}

	config := NewConfig()

	slog.Debug("Loading config", "path", path)
	b, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(b, &config); err != nil {
		return config, err
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
