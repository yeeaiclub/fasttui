//go:build windows

package terminal

import (
	"time"
)

// registerResizeSignal polls console dimensions: Windows has no SIGWINCH for
// ConPTY/Win32 console the way Unix does. Node's process.stdout "resize" uses
// a similar poll; this keeps TUI in sync while dragging the window.
func registerResizeSignal(p *ProcessTerminal) func() {
	done := make(chan struct{})
	go func() {
		lastW, lastH, err := getTerminalSize(p.stdoutFD)
		if err != nil {
			lastW, lastH = 80, 24
		}
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				w, h, err := getTerminalSize(p.stdoutFD)
				if err != nil {
					continue
				}
				if w == lastW && h == lastH {
					continue
				}
				lastW, lastH = w, h
				if p.resizeHandler != nil {
					p.resizeHandler()
				}
			}
		}
	}()
	return func() { close(done) }
}
