//go:build windows

package config

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var vdfPathRe = regexp.MustCompile(`"path"\s+"([^"]+)"`)

// defaultClientPaths returns Windows candidate paths for Client.txt.
func defaultClientPaths() []string {
	paths := []string{
		// Steam on Windows (common case)
		`C:\Program Files (x86)\Steam\steamapps\common\Path of Exile 2\logs\Client.txt`,
		// Standalone installer
		`C:\Program Files\Grinding Gear Games\Path of Exile 2\logs\Client.txt`,
	}
	// Append paths discovered from Steam's libraryfolders.vdf
	paths = append(paths, steamLibPaths()...)
	return paths
}

// steamLibPaths parses Steam's libraryfolders.vdf to find all library folders
// and returns candidate Client.txt paths for Path of Exile 2.
func steamLibPaths() []string {
	// Check common Steam locations on Windows
	steamDirs := []string{
		`C:\Program Files (x86)\Steam`,
	}
	if p := os.Getenv("STEAM_PATH"); p != "" {
		steamDirs = append(steamDirs, p)
	}

	var paths []string
	seen := make(map[string]bool)
	for _, dir := range steamDirs {
		for _, rel := range []string{`config\libraryfolders.vdf`, `steamapps\libraryfolders.vdf`} {
			vdf := filepath.Join(dir, rel)
			for _, lib := range parseLibraryFolders(vdf) {
				p := filepath.Join(lib, "steamapps", "common", "Path of Exile 2", "logs", "Client.txt")
				if !seen[p] {
					seen[p] = true
					paths = append(paths, p)
				}
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
