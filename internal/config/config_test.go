package config

import (
	"os"
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

func TestParseLibraryFolders(t *testing.T) {
	vdf := filepath.Join(t.TempDir(), "libraryfolders.vdf")
	content := `"libraryfolders"
{
	"0"
	{
		"path"		"/home/user/.local/share/Steam"
	}
	"1"
	{
		"path"		"/mnt/m2/SteamLibrary"
	}
}`
	if err := writeFile(vdf, content); err != nil {
		t.Fatal(err)
	}
	libs := parseLibraryFolders(vdf)
	if len(libs) != 2 {
		t.Fatalf("expected 2 libs, got %d: %v", len(libs), libs)
	}
	if libs[0] != "/home/user/.local/share/Steam" {
		t.Fatalf("unexpected lib[0]: %s", libs[0])
	}
	if libs[1] != "/mnt/m2/SteamLibrary" {
		t.Fatalf("unexpected lib[1]: %s", libs[1])
	}
}

func TestParseLibraryFoldersMissing(t *testing.T) {
	libs := parseLibraryFolders(filepath.Join(t.TempDir(), "nonexistent.vdf"))
	if len(libs) != 0 {
		t.Fatalf("expected 0 libs, got %d", len(libs))
	}
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
