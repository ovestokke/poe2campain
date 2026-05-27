//go:build !windows

package config

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// defaultClientPaths returns Linux candidate paths for Client.txt.
// Includes static paths and dynamic paths discovered from Steam's libraryfolders.vdf.
func defaultClientPaths() []string {
	paths := []string{
		// Steam (native Linux)
		"~/.steam/steam/steamapps/common/Path of Exile 2/logs/Client.txt",
		// Steam (Steamdeck / Flatpak alternative)
		"~/.local/share/Steam/steamapps/common/Path of Exile 2/logs/Client.txt",
		// Standalone installer
		"~/Path of Exile 2/logs/Client.txt",
	}
	// Append paths discovered from Steam's libraryfolders.vdf
	paths = append(paths, steamLibPaths()...)
	return paths
}

var vdfPathRe = regexp.MustCompile(`"path"\s+"([^"]+)"`)

// steamLibPaths parses Steam's libraryfolders.vdf to find all library folders
// and returns candidate Client.txt paths for Path of Exile 2.
func steamLibPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}

	// Steam stores libraryfolders.vdf in two possible locations
	candidates := []string{
		filepath.Join(home, ".steam", "steam", "config", "libraryfolders.vdf"),
		filepath.Join(home, ".steam", "steam", "steamapps", "libraryfolders.vdf"),
		filepath.Join(home, ".local", "share", "Steam", "config", "libraryfolders.vdf"),
		filepath.Join(home, ".local", "share", "Steam", "steamapps", "libraryfolders.vdf"),
	}

	var paths []string
	seen := make(map[string]bool)
	for _, vdf := range candidates {
		libs := parseLibraryFolders(vdf)
		for _, lib := range libs {
			p := filepath.Join(lib, "steamapps", "common", "Path of Exile 2", "logs", "Client.txt")
			if !seen[p] {
				seen[p] = true
				paths = append(paths, p)
			}
		}
	}
	return paths
}

// parseLibraryFolders extracts "path" values from a Steam libraryfolders.vdf file.
func parseLibraryFolders(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var libs []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if m := vdfPathRe.FindStringSubmatch(line); m != nil {
			libs = append(libs, m[1])
		}
	}
	return libs
}
