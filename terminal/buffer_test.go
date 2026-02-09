package terminal

import (
	"fmt"
	"testing"
	"time"
)

func TestStdinBuffer_Process(t *testing.T) {
	buf := NewStdinBuffer()

	buf.OnData = func(seq string) {
		fmt.Printf("Received data: %q\n", seq)
	}

	buf.OnPaste = func(paste string) {
		fmt.Printf("Received paste: %q\n", paste)
	}

	testCases := []struct {
		name string
		data string
	}{
		{"Regular characters", "hello"},
		{"Enter key", "\r"},
		{"Up arrow", "\x1b[A"},
		{"Down arrow", "\x1b[B"},
		{"Left arrow", "\x1b[D"},
		{"Right arrow", "\x1b[C"},
		{"Paste content", "\x1b[200~hello world\x1b[201~"},
		{"Tab key", "\t"},
		{"Delete key", "\x1b[3~"},
		{"Home key", "\x1b[H"},
		{"End key", "\x1b[F"},
	}

	for _, tc := range testCases {
		buf.Process(tc.data)
		time.Sleep(50 * time.Millisecond)
	}
}
