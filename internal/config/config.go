package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// DefaultClientPaths returns platform-specific candidate paths for Client.txt.
// Paths are deduplicated (after tilde expansion) and returned with ~ preserved for display.
func DefaultClientPaths() []string {
	return deduplicatePaths(defaultClientPaths())
}

func deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		expanded := expandTilde(p)
		if !seen[expanded] {
			seen[expanded] = true
			out = append(out, p)
		}
	}
	return out
}

// DefaultPathDisplay returns DefaultPath with ~ replacing the home directory for help text.
func DefaultPathDisplay() string {
	p := DefaultPath()
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

// FindClientTxt returns the first existing Client.txt from DefaultClientPaths, or "".
func FindClientTxt() string {
	for _, p := range DefaultClientPaths() {
		expanded := expandTilde(p)
		if _, err := os.Stat(expanded); err == nil {
			return expanded
		}
	}
	return ""
}

func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	return filepath.Join(home, path[2:])
}
