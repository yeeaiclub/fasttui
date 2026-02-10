package terminal

import (
	"os"
	"regexp"
	"strconv"

	"golang.org/x/term"
)

var kittyResponsePattern = regexp.MustCompile(`^\x1b\[\?(\d+)u$`)

type ProcessTerminal struct {
	buffer *StdinBuffer
	stdout *os.File
	fd     int
	// saved                 *term.State
	isKittyProtocolActive bool
	inputHandler          func(data string)
	stdinDataBuffer       func(data string)
}

func NewProcessTerminal() *ProcessTerminal {
	buffer := NewStdinBuffer()
	return &ProcessTerminal{buffer: buffer, isKittyProtocolActive: false}
}

func (p *ProcessTerminal) GetSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return w, h
}

func (p *ProcessTerminal) IsKittyProtocolActive() bool {
	return p.isKittyProtocolActive
}

func (p *ProcessTerminal) Start(onInput func(data string), onResize func()) error {
	p.stdout = os.Stdout
	p.fd = int(os.Stdout.Fd())
	p.inputHandler = onInput

	p.print("\x1b[?2004h")

	p.queryAndEnableKittyProtocol()
	go p.readInputLoop()
	return nil
}

func (p *ProcessTerminal) readInputLoop() {
	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		if n > 0 {
			data := string(buf[:n])
			if p.stdinDataBuffer != nil {
				p.stdinDataBuffer(data)
			}
		}
	}
}

func (p *ProcessTerminal) setupStdinBuffer() {
	p.buffer = NewStdinBuffer()
	p.buffer.OnData = func(seq string) {
		if !p.isKittyProtocolActive {
			match := kittyResponsePattern.FindStringSubmatch(seq)
			if len(match) > 0 {
				p.isKittyProtocolActive = true
				p.print("\x1b[>7u")
				return
			}
		}
		if p.inputHandler != nil {
			p.inputHandler(seq)
		}
	}
	p.buffer.OnPaste = func(paste string) {
		if p.inputHandler != nil {
			p.inputHandler("\x1b[200~" + paste + "\x1b[201~")
		}
	}
	p.stdinDataBuffer = func(data string) {
		p.buffer.Process(data)
	}
}

func (p *ProcessTerminal) queryAndEnableKittyProtocol() {
	p.setupStdinBuffer()
	p.print("\x1b[?u")
}

func (p *ProcessTerminal) DrainInput(maxMs int, idleMs int) error {
	return nil
}

func (p *ProcessTerminal) Stop() {
	p.print("\x1b[?2004l")
	if p.IsKittyProtocolActive() {
		p.print("\x1b[>7l")
	}
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
	p.print("\x1b[2K")
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
	_, _ = p.stdout.WriteString(s)
}
