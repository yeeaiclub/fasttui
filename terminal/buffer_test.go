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
		data []byte
	}{
		{"Regular characters", []byte("hello")},
		{"Enter key", []byte("\r")},
		{"Up arrow", []byte("\x1b[A")},
		{"Down arrow", []byte("\x1b[B")},
		{"Left arrow", []byte("\x1b[D")},
		{"Right arrow", []byte("\x1b[C")},
		{"Paste content", []byte("\x1b[200~hello world\x1b[201~")},
		{"Tab key", []byte("\t")},
		{"Delete key", []byte("\x1b[3~")},
		{"Home key", []byte("\x1b[H")},
		{"End key", []byte("\x1b[F")},
		{"Chinese characters", []byte("你好世界")},
		{"Mixed content", []byte("Hello 世界")},
	}

	for _, tc := range testCases {
		buf.Process(tc.data)
		time.Sleep(50 * time.Millisecond)
	}
}
