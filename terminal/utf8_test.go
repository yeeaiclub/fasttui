package terminal

import (
	"testing"
	"time"
)

func TestStdinBuffer_UTF8Characters(t *testing.T) {
	buf := NewStdinBuffer()

	received := make([]string, 0)
	buf.OnData = func(seq string) {
		received = append(received, seq)
	}

	// Test Chinese characters
	testInput := []byte("你好世界")
	buf.Process(testInput)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Should receive 4 complete Chinese characters
	if len(received) != 4 {
		t.Errorf("Expected 4 characters, got %d", len(received))
	}

	expected := []string{"你", "好", "世", "界"}
	for i, exp := range expected {
		if i >= len(received) {
			t.Errorf("Missing character at index %d", i)
			continue
		}
		if received[i] != exp {
			t.Errorf("Character %d: expected %q, got %q", i, exp, received[i])
		}
	}

	buf.Close()
}

func TestStdinBuffer_MixedContent(t *testing.T) {
	buf := NewStdinBuffer()

	received := make([]string, 0)
	buf.OnData = func(seq string) {
		received = append(received, seq)
	}

	// Test mixed English and Chinese
	testInput := []byte("Hello世界")
	buf.Process(testInput)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Should receive: H, e, l, l, o, 世, 界
	expected := []string{"H", "e", "l", "l", "o", "世", "界"}
	if len(received) != len(expected) {
		t.Errorf("Expected %d characters, got %d", len(expected), len(received))
	}

	for i, exp := range expected {
		if i >= len(received) {
			t.Errorf("Missing character at index %d", i)
			continue
		}
		if received[i] != exp {
			t.Errorf("Character %d: expected %q, got %q", i, exp, received[i])
		}
	}

	buf.Close()
}
