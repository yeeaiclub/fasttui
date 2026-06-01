package terminal

import (
	"sync"
	"testing"
	"time"
)

func TestStdinBuffer_UTF8Characters(t *testing.T) {
	buf := NewStdinBuffer()

	var mu sync.Mutex
	received := make([]string, 0)
	buf.OnData = func(seq string) {
		mu.Lock()
		received = append(received, seq)
		mu.Unlock()
	}

	testInput := []byte("你好世界")
	buf.Process(testInput)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	got := append([]string(nil), received...)
	mu.Unlock()

	if len(got) != 4 {
		t.Errorf("Expected 4 characters, got %d", len(got))
	}

	expected := []string{"你", "好", "世", "界"}
	for i, exp := range expected {
		if i >= len(got) {
			t.Errorf("Missing character at index %d", i)
			continue
		}
		if got[i] != exp {
			t.Errorf("Character %d: expected %q, got %q", i, exp, got[i])
		}
	}

	buf.Close()
}

func TestStdinBuffer_MixedContent(t *testing.T) {
	buf := NewStdinBuffer()

	var mu sync.Mutex
	received := make([]string, 0)
	buf.OnData = func(seq string) {
		mu.Lock()
		received = append(received, seq)
		mu.Unlock()
	}

	testInput := []byte("Hello世界")
	buf.Process(testInput)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	got := append([]string(nil), received...)
	mu.Unlock()

	expected := []string{"H", "e", "l", "l", "o", "世", "界"}
	if len(got) != len(expected) {
		t.Errorf("Expected %d characters, got %d", len(expected), len(got))
	}

	for i, exp := range expected {
		if i >= len(got) {
			t.Errorf("Missing character at index %d", i)
			continue
		}
		if got[i] != exp {
			t.Errorf("Character %d: expected %q, got %q", i, exp, got[i])
		}
	}

	buf.Close()
}
