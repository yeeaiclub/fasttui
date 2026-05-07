package terminal

import (
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var kittyResponsePattern = regexp.MustCompile(`^\x1b\[\?(\d+)u$`)

type ProcessTerminal struct {
	buffer                *StdinBuffer
	rawState              *rawModeState
	stdinFD               int
	stdoutFD              int
	isKittyProtocolActive bool
	inputHandler          func(data string)
	stdinDataBuffer       func(data []byte)
	wasRaw                bool
	resizeHandler         func()
	resizeSignalChan      chan os.Signal
	stopChan              chan struct{}
	stopOnce              sync.Once
	stopResizeSignal      func()
}

func NewProcessTerminal() *ProcessTerminal {
	buffer := NewStdinBuffer()
	return &ProcessTerminal{
		buffer:                buffer,
		stdinFD:               int(os.Stdin.Fd()),
		stdoutFD:              int(os.Stdout.Fd()),
		isKittyProtocolActive: false,
		wasRaw:                false,
		resizeSignalChan:      make(chan os.Signal, 1),
		stopChan:              make(chan struct{}),
	}
}

func (p *ProcessTerminal) GetSize() (int, int) {
	w, h, err := getTerminalSize(p.stdoutFD)
	if err != nil {
		return 80, 24
	}
	return w, h
}

func (p *ProcessTerminal) IsKittyProtocolActive() bool {
	return p.isKittyProtocolActive
}

func (p *ProcessTerminal) Start(onInput func(data string), onResize func()) error {
	p.inputHandler = onInput
	p.resizeHandler = onResize

	// Save previous state and enable raw mode on STDIN (not stdout!)
	rawState, err := enableRawMode(p.stdinFD)
	if err != nil {
		return err
	}
	p.rawState = rawState

	// Enable bracketed paste mode - terminal will wrap pastes in \x1b[200~ ... \x1b[201~
	p.print("\x1b[?2004h")

	// Alternate screen: keeps scrollback intact and gives a clean grid for incremental draws.
	// Clear + home so hardware row 0 matches the TUI's first full paint (previousLines == nil).
	// Note: processes that inherit stdout can still interleave on this buffer; redirect their
	// stdout/stderr if you spawn them while the TUI is running.
	p.print("\x1b[?1049h\x1b[2J\x1b[H")

	// Query and enable Kitty keyboard protocol
	p.queryAndEnableKittyProtocol()

	// Set up resize signal handling
	p.stopResizeSignal = registerResizeSignal(p.resizeSignalChan)
	go p.handleResizeSignal()

	// Start reading input in background
	go p.readInputLoop()
	return nil
}

func (p *ProcessTerminal) readInputLoop() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-p.stopChan:
			return
		default:
			// Set a read deadline to allow checking stopChan periodically
			os.Stdin.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := os.Stdin.Read(buf)
			if err != nil {
				// Check if it's a timeout error (expected)
				if os.IsTimeout(err) {
					continue
				}
				// Other errors mean we should stop
				return
			}
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				if p.stdinDataBuffer != nil {
					p.stdinDataBuffer(data)
				}
			}
		}
	}
}

func (p *ProcessTerminal) handleResizeSignal() {
	for {
		select {
		case <-p.resizeSignalChan:
			if p.resizeHandler != nil {
				p.resizeHandler()
			}
		case <-p.stopChan:
			return
		}
	}
}

// setupStdinBuffer sets up StdinBuffer to split batched input into individual sequences.
// This ensures components receive single events, making matchesKey/isKeyRelease work correctly.
//
// Also watches for Kitty protocol response and enables it when detected.
func (p *ProcessTerminal) setupStdinBuffer() {
	p.buffer = NewStdinBuffer()

	// Forward individual sequences to the input handler
	p.buffer.OnData = func(seq string) {
		// Check for Kitty protocol response (only if not already enabled)
		if !p.isKittyProtocolActive {
			match := kittyResponsePattern.FindStringSubmatch(seq)
			if len(match) > 1 {
				p.isKittyProtocolActive = true

				// Enable Kitty keyboard protocol (push flags)
				// Flag 1 = disambiguate escape codes
				// Flag 2 = report event types (press/repeat/release)
				// Flag 4 = report alternate keys (shifted key, base layout key)
				// Base layout key enables shortcuts to work with non-Latin keyboard layouts
				p.print("\x1b[>7u")
				return // Don't forward protocol response to TUI
			}
		}
		if p.inputHandler != nil {
			p.inputHandler(seq)
		}
	}

	// Re-wrap paste content with bracketed paste markers for existing editor handling
	p.buffer.OnPaste = func(paste string) {
		if p.inputHandler != nil {
			p.inputHandler("\x1b[200~" + paste + "\x1b[201~")
		}
	}

	// Handler that pipes stdin data through the buffer
	p.stdinDataBuffer = func(data []byte) {
		if p.buffer != nil {
			p.buffer.Process(data)
		}
	}
}

// queryAndEnableKittyProtocol queries terminal for Kitty keyboard protocol support and enables if available.
//
// Sends CSI ? u to query current flags. If terminal responds with CSI ? <flags> u,
// it supports the protocol and we enable it with CSI > 7 u.
//
// The response is detected in setupStdinBuffer's data handler.
func (p *ProcessTerminal) queryAndEnableKittyProtocol() {
	p.setupStdinBuffer()
	p.print("\x1b[?u")
}

func (p *ProcessTerminal) DrainInput(maxMs int, idleMs int) error {
	if p.isKittyProtocolActive {
		p.print("\x1b[<u")
		p.isKittyProtocolActive = false
	}

	previousHandler := p.inputHandler
	p.inputHandler = nil

	lastDataTime := time.Now()
	endTime := time.Now().Add(time.Duration(maxMs) * time.Millisecond)
	idleTimeout := time.Duration(idleMs) * time.Millisecond

	buf := make([]byte, 1024)

	for time.Now().Before(endTime) {
		os.Stdin.SetReadDeadline(time.Now().Add(idleTimeout))
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			lastDataTime = time.Now()
		}

		if time.Since(lastDataTime) > idleTimeout {
			break
		}
	}

	os.Stdin.SetReadDeadline(time.Time{})
	p.inputHandler = previousHandler
	return nil
}

func (p *ProcessTerminal) Stop() {
	p.stopOnce.Do(func() {
		// Signal goroutines to stop
		close(p.stopChan)

		// Disable bracketed paste mode (while still on the alternate screen)
		p.print("\x1b[?2004l")

		// Disable Kitty keyboard protocol (pop the flags we pushed) - only if we enabled it
		if p.isKittyProtocolActive {
			p.print("\x1b[<u")
			p.isKittyProtocolActive = false
		}

		// Reset SGR before leaving alternate screen
		p.print("\x1b[0m")

		// Leave alternate screen (restores primary buffer from before Start)
		p.print("\x1b[?1049l")

		// Show cursor for the shell on the primary screen
		p.print("\x1b[?25h")

		// One fresh line after restore (no need for multiple blank lines on primary)
		p.print("\r\n")

		// Clean up StdinBuffer
		if p.buffer != nil {
			p.buffer.Close()
			p.buffer = nil
		}

		// Remove event handlers
		p.stdinDataBuffer = nil
		p.inputHandler = nil
		p.resizeHandler = nil

		// Stop signal notifications
		if p.stopResizeSignal != nil {
			p.stopResizeSignal()
			p.stopResizeSignal = nil
		}

		// Wait a moment for goroutines to exit
		time.Sleep(50 * time.Millisecond)

		// Restore terminal state
		if p.rawState != nil {
			// Clear any read deadline before restoring
			os.Stdin.SetReadDeadline(time.Time{})
			_ = disableRawMode(p.rawState)
			p.rawState = nil
		}
	})
}

func (p *ProcessTerminal) Write(data string) {
	p.print(data)
}

func (p *ProcessTerminal) MoveBy(lines int) {
	if lines > 0 {
		p.print("\x1b[" + strconv.Itoa(lines) + "B")
	} else if lines < 0 {
		// Move up
		p.print("\x1b[" + strconv.Itoa(-lines) + "A")
	}
	// lines === 0: no movement
}

func (p *ProcessTerminal) HideCursor() {
	p.print("\x1b[?25l")
}

func (p *ProcessTerminal) ShowCursor() {
	p.print("\x1b[?25h")
}

func (p *ProcessTerminal) ClearLine() {
	p.print("\x1b[K")
}

func (p *ProcessTerminal) ClearFromCursor() {
	p.print("\x1b[J")
}

func (p *ProcessTerminal) ClearScreen() {
	p.print("\x1b[2J\x1b[H") // Clear screen and move to home (1,1)
}

func (p *ProcessTerminal) SetTitle(title string) {
	// OSC 0;title BEL - set terminal window title
	p.print("\x1b]0;" + title + "\x07")
}

func (p *ProcessTerminal) print(s string) {
	_, _ = os.Stdout.WriteString(s)
}
