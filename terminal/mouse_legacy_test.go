package terminal

import (
	"sync"
	"testing"
	"time"
)

func TestStdinBuffer_LegacyMouseESCM(t *testing.T) {
	buf := NewStdinBuffer()
	defer buf.Close()

	var mu sync.Mutex
	received := make([]string, 0)
	buf.OnData = func(seq string) {
		mu.Lock()
		received = append(received, seq)
		mu.Unlock()
	}

	// Legacy mouse format: ESC [ M + 3 bytes (button, x+32, y+32)
	// Feed partial chunks to ensure parser waits for all 6 bytes.
	buf.Process([]byte("\x1b[M"))
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	if len(received) != 0 {
		t.Fatalf("expected no event for partial ESC[M, got %v", received)
	}
	mu.Unlock()

	buf.Process([]byte{0x20, 0x21, 0x22})
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 legacy mouse event, got %d (%v)", len(received), received)
	}
	if received[0] != "\x1b[M !\"" {
		t.Fatalf("unexpected event: %q", received[0])
	}
}
