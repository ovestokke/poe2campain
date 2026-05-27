package logreader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// Follow watches Client.txt for newly appended area-generation lines.
// It only emits verified "Generating level ... area ..." events.
func Follow(ctx context.Context, path string, startAtEnd bool, pollInterval time.Duration, emit func(AreaEvent)) error {
	if pollInterval <= 0 {
		pollInterval = 500 * time.Millisecond
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if startAtEnd {
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			return err
		}
	}

	reader := bufio.NewReader(file)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if event, ok := ParseAreaEvent(line); ok {
				emit(event)
			}
		}
		if err == nil {
			continue
		}
		if err != io.EOF {
			return fmt.Errorf("follow client log: %w", err)
		}

		pos, seekErr := file.Seek(0, io.SeekCurrent)
		info, statErr := file.Stat()
		if seekErr == nil && statErr == nil && info.Size() < pos {
			// Log was truncated/recreated. Start again at the beginning of this file.
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return err
			}
			reader.Reset(file)
			continue
		}

		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil
		case <-timer.C:
		}
	}
}
