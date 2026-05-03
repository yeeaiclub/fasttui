//go:build linux || solaris || aix

package terminal

import "golang.org/x/sys/unix"

// rawModeState stores the original terminal state for restoration.
type rawModeState struct {
	fd      int
	termios unix.Termios
}

// enableRawMode puts the terminal into raw mode and returns the previous state.
func enableRawMode(fd int) (*rawModeState, error) {
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}

	state := &rawModeState{fd: fd, termios: *termios}

	termios.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Iflag &^= unix.IXON | unix.ICRNL | unix.BRKINT | unix.INPCK | unix.ISTRIP
	termios.Oflag &^= unix.OPOST
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, termios); err != nil {
		return nil, err
	}

	return state, nil
}

// disableRawMode restores the terminal to its previous state.
func disableRawMode(state *rawModeState) error {
	if state == nil {
		return nil
	}
	return unix.IoctlSetTermios(state.fd, unix.TCSETS, &state.termios)
}

// getTerminalSize returns the terminal dimensions.
func getTerminalSize(fd int) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(ws.Col), int(ws.Row), nil
}
