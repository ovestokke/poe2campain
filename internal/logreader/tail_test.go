package logreader

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFollowEmitsAppendedAreaEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Client.txt")
	if err := os.WriteFile(path, []byte("old line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan AreaEvent, 1)
	errs := make(chan error, 1)
	go func() {
		errs <- Follow(ctx, path, false, 10*time.Millisecond, func(event AreaEvent) {
			events <- event
		})
	}()

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.WriteString(`2026/05/26 00:00:00 Generating level 13 area "G1_13_2"` + "\n")
	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-events:
		if event.Level != 13 || event.AreaID != "G1_13_2" {
			t.Fatalf("unexpected event: %+v", event)
		}
	case err := <-errs:
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("follow stopped before event")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}

	cancel()
	select {
	case err := <-errs:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for follow to stop")
	}
}
