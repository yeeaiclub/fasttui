//go:build windows

package terminal

import "os"

func registerResizeSignal(ch chan os.Signal) func() {
	return func() {}
}
