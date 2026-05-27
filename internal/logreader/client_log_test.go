package logreader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAreaEvent(t *testing.T) {
	line := `2026/05/26 19:10:40 442387806 2caa22d2 [DEBUG Client 308] Generating level 13 area "G1_13_2" with seed 2087688264`
	event, ok := ParseAreaEvent(line)
	if !ok {
		t.Fatal("expected area event")
	}
	if event.Level != 13 || event.AreaID != "G1_13_2" {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestParseAreaEventIgnoresOtherLines(t *testing.T) {
	if _, ok := ParseAreaEvent("chat or unrelated log line"); ok {
		t.Fatal("unexpected match")
	}
}

func TestScanLatest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Client.txt")
	content := "unrelated\n" +
		`2026/05/26 [DEBUG Client 1] Generating level 2 area "G1_2" with seed 1` + "\n" +
		`2026/05/26 [DEBUG Client 1] Generating level 7 area "G1_7" with seed 2` + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	event, found, err := ScanLatest(path)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected match")
	}
	if event.Level != 7 || event.AreaID != "G1_7" {
		t.Fatalf("unexpected latest event: %+v", event)
	}
}
