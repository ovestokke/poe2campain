package logreader

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var areaRe = regexp.MustCompile(`Generating level (\d+) area "([^"]+)"`)

type AreaEvent struct {
	Level  int
	AreaID string
}

func ParseAreaEvent(line string) (AreaEvent, bool) {
	match := areaRe.FindStringSubmatch(line)
	if match == nil {
		return AreaEvent{}, false
	}
	level, err := strconv.Atoi(match[1])
	if err != nil {
		return AreaEvent{}, false
	}
	areaID := strings.TrimSpace(match[2])
	if areaID == "" {
		return AreaEvent{}, false
	}
	return AreaEvent{Level: level, AreaID: areaID}, true
}

func ScanLatest(path string) (AreaEvent, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return AreaEvent{}, false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Client.txt lines should be short, but allow more than Scanner's 64 KiB default
	// without ever reading the whole log into memory.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var latest AreaEvent
	found := false
	for scanner.Scan() {
		event, ok := ParseAreaEvent(scanner.Text())
		if !ok {
			continue
		}
		latest = event
		found = true
	}
	if err := scanner.Err(); err != nil {
		return AreaEvent{}, false, fmt.Errorf("scan client log: %w", err)
	}
	return latest, found, nil
}
