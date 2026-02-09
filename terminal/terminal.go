package terminal

import (
	"log"
	"os"
	"strconv"

	"golang.org/x/term"
)

type ProcessTerminal struct {
	buffer *StdinBuffer
	stdout *os.File
	fd     int
	saved  *term.State
}

func NewProcessTerminal() *ProcessTerminal {
	buffer := NewStdinBuffer()
	return &ProcessTerminal{buffer: buffer}
}

func (p *ProcessTerminal) GetSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return w, h
}

func (p *ProcessTerminal) IsKittyProtocolActive() bool {
	return false
}

func (p *ProcessTerminal) Start(onInput func(data string), onResize func()) error {
	p.stdout = os.Stdout
	p.fd = int(os.Stdout.Fd())
	saved, err := term.MakeRaw(p.fd)
	if err != nil {
		return err
	}
	p.saved = saved
	return nil
}

func (p *ProcessTerminal) DrainInput(maxMs int, idleMs int) error {
	return nil
}

func (p *ProcessTerminal) queryAndEnableKittyProtocol() {
	p.stdout.WriteString("\x1b[?u")
}

func (p *ProcessTerminal) Stop() {
}

func (p *ProcessTerminal) Write(data string) {
	_, err := p.stdout.WriteString(data)
	if err != nil {
		log.Printf("Error writing to stdout: %v", err)
		return
	}
}

func (p *ProcessTerminal) MoveBy(lines int) {
	if lines > 0 {
		p.stdout.WriteString("\x1b[" + strconv.Itoa(lines) + "B")
	} else if lines < 0 {
		// Move up
		p.stdout.WriteString("\x1b[" + strconv.Itoa(-lines) + "A")
	}
	// lines === 0: no movement
}

func (p *ProcessTerminal) HideCursor() {
	p.stdout.WriteString("\x1b[?25l")
}

func (p *ProcessTerminal) ShowCursor() {
	p.stdout.WriteString("\x1b[?25h")
}

func (p *ProcessTerminal) ClearLine() {
	p.stdout.WriteString("\x1b[2K")
}

func (p *ProcessTerminal) ClearFromCursor() {
	p.stdout.WriteString("\x1b[J")
}

func (p *ProcessTerminal) ClearScreen() {
	p.stdout.WriteString("\x1b[2J\x1b[H") // Clear screen and move to home (1,1)
}

func (p *ProcessTerminal) SetTitle(title string) {
	// OSC 0;title BEL - set terminal window title
	p.stdout.WriteString("\x1b]0;" + title + "\x07")
}
