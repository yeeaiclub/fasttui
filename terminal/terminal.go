package terminal

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/yeeaiclub/fasttui/internal/caps"
	"github.com/yeeaiclub/fasttui/internal/debug"
	"github.com/yeeaiclub/fasttui/keys"
)

// Default probe window for VT capability queries (matches libvaxis).
const probeTimeout = time.Second

type ProcessTerminal struct {
	buffer           *StdinBuffer
	rawState         *rawModeState
	stdinFD          int
	stdoutFD         int
	detector         *caps.Detector
	inputHandler     func(data string)
	stdinDataBuffer  func(data []byte)
	wasRaw           bool
	resizeHandler    func()
	resizeSignalChan chan os.Signal
	stopChan         chan struct{}
	stopOnce         sync.Once
	stopResizeSignal func()

	probeDoneOnce sync.Once
	probeDoneCh   chan struct{}
}

func NewProcessTerminal() *ProcessTerminal {
	buffer := NewStdinBuffer()
	return &ProcessTerminal{
		buffer:           buffer,
		stdinFD:          int(os.Stdin.Fd()),
		stdoutFD:         int(os.Stdout.Fd()),
		resizeSignalChan: make(chan os.Signal, 1),
		stopChan:         make(chan struct{}),
	}
}

func (p *ProcessTerminal) GetSize() (int, int) {
	w, h, err := getTerminalSize(p.stdoutFD)
	if err != nil {
		return 80, 24
	}
	return w, h
}

// Caps returns features discovered during Start's capability probe.
// Before Start (or if probing never ran), returns the zero value.
func (p *ProcessTerminal) Caps() caps.Caps {
	if p.detector == nil {
		return caps.Caps{}
	}
	return p.detector.Caps()
}

// IsKittyProtocolActive reports whether Kitty keyboard was detected and enabled.
func (p *ProcessTerminal) IsKittyProtocolActive() bool {
	return p.Caps().KittyKeyboard
}

// SupportsSyncOutput reports whether mode 2026 was confirmed via DECRQM.
func (p *ProcessTerminal) SupportsSyncOutput() bool {
	return p.Caps().SyncOutput
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

	p.setupStdinBuffer()
	go p.readInputLoop()

	p.probeAndEnable()

	// Set up resize signal handling
	p.stopResizeSignal = registerResizeSignal(p)
	go p.handleResizeSignal()

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

// setupStdinBuffer splits batched input into individual sequences and feeds
// capability-probe replies into detector until probing finishes.
func (p *ProcessTerminal) setupStdinBuffer() {
	p.buffer = NewStdinBuffer()
	p.detector = caps.New()
	p.probeDoneCh = make(chan struct{})
	p.probeDoneOnce = sync.Once{}

	p.buffer.OnData = func(seq string) {
		if p.detector != nil && !p.detector.Done() {
			if p.detector.Feed(seq) {
				if p.detector.Done() {
					p.signalProbeDone()
				}
				return // probe response — do not forward to the app
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

	p.stdinDataBuffer = func(data []byte) {
		if p.buffer != nil {
			p.buffer.Process(data)
		}
	}
}

// probeAndEnable writes VT queries, waits for Primary DA (or timeout), then
// enables detected features. Reader must already be running.
func (p *ProcessTerminal) probeAndEnable() {
	p.print(string(p.detector.Query()))

	select {
	case <-p.probeDoneCh:
	case <-time.After(probeTimeout):
		p.detector.Complete()
		p.signalProbeDone()
	}

	p.print(string(p.detector.Enable()))
	keys.SetKittyProtocolActive(p.detector.Caps().KittyKeyboard)
}

func (p *ProcessTerminal) signalProbeDone() {
	p.probeDoneOnce.Do(func() {
		close(p.probeDoneCh)
	})
}

func (p *ProcessTerminal) Stop() {
	p.stopOnce.Do(func() {
		// Signal goroutines to stop
		close(p.stopChan)

		if p.detector != nil {
			p.print(string(p.detector.Disable()))
		} else {
			p.print("\x1b[?2004l")
		}
		keys.SetKittyProtocolActive(false)

		// Show cursor (ensure it's visible after exit)
		p.print("\x1b[?25h")

		// Reset SGR (colors, etc)
		p.print("\x1b[0m")

		// Move cursor down 4 lines to avoid overwriting last rendered content
		p.print("\r\n\r\n\r\n\r\n")

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
	debug.AppendWrite(s)
}
