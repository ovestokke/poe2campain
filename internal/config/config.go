package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const AppName = "poe2campain"

// Config stores user-specific settings. Keep it small and hand-editable.
type Config struct {
	ClientPath string `json:"client_txt"`
}

func DefaultPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName, "config.json")
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, AppName, "config.json")
	}
	return filepath.Join(".", "config.json")
}

func DefaultStatePath() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName, "state.json")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "state", AppName, "state.json")
	}
	return filepath.Join(".", "state.json")
}

func Load(path string) (Config, bool, error) {
	if path == "" {
		path = DefaultPath()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, false, nil
		}
		return Config{}, false, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, true, fmt.Errorf("decode config %s: %w", path, err)
	}
	return cfg, true, nil
}

func Save(path string, cfg Config) error {
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0644)
}

func Exists(path string) bool {
	if path == "" {
		path = DefaultPath()
	}
	_, err := os.Stat(path)
	return err == nil
}
