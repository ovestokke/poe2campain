package config

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingConfig(t *testing.T) {
	cfg, found, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expected missing config")
	}
	if cfg.ClientPath != "" {
		t.Fatalf("expected empty config, got %+v", cfg)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "poe2campain", "config.json")
	want := Config{ClientPath: "/tmp/Client.txt"}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, found, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected config")
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
