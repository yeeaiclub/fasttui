//go:build !windows

package terminal

import (
	"os"
	"os/signal"
	"syscall"
)

func registerResizeSignal(ch chan os.Signal) func() {
	signal.Notify(ch, syscall.SIGWINCH)
	return func() {
		signal.Stop(ch)
	}
}
