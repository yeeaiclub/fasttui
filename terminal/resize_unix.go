//go:build !windows

package terminal

import (
	"os/signal"
	"syscall"
)

func registerResizeSignal(p *ProcessTerminal) func() {
	signal.Notify(p.resizeSignalChan, syscall.SIGWINCH)
	return func() {
		signal.Stop(p.resizeSignalChan)
	}
}
