package terminal

import (
	"bufio"
	"golang.org/x/term"
	"log"
	"os"
	"regexp"
	"strconv"
)

var kittyPattern = regexp.MustCompile(`^\x1b\[\?(\d+)u$`)

type ProcessTerminal struct {
	buffer *StdinBuffer
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
}

func (p *ProcessTerminal) Start(onInput func(data string), onResize func()) error {
}

func (p *ProcessTerminal) readStdin() {
	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			return
		}
		p.buffer.Process(string(buf[:n]))
	}
}

func (p *ProcessTerminal) DrainInput(maxMs int, idleMs int) error {
}

func (p *ProcessTerminal) queryAndEnableKittyProtocol() {
	os.Stdout.WriteString("\x1b[?u")
}

func (p *ProcessTerminal) Stop() {
}

func (p *ProcessTerminal) Write(data string) {
	_, err := os.Stdout.WriteString(data)
	if err != nil {
		log.Printf("Error writing to stdout: %v", err)
		return
	}
}

func (p *ProcessTerminal) MoveBy(lines int) {
	if lines > 0 {
		os.Stdout.WriteString("\x1b[" + strconv.Itoa(lines) + "B")
	} else if lines < 0 {
		// Move up
		os.Stdout.WriteString("\x1b[" + strconv.Itoa(-lines) + "A")
	}
	// lines === 0: no movement
}

func (p *ProcessTerminal) HideCursor() {
	os.Stdout.WriteString("\x1b[?25l")
}

func (p *ProcessTerminal) ShowCursor() {
	os.Stdout.WriteString("\x1b[?25h")
}

func (p *ProcessTerminal) ClearLine() {
	os.Stdout.WriteString("\x1b[2K")
}

func (p *ProcessTerminal) ClearFromCursor() {
	os.Stdout.WriteString("\x1b[J")
}

func (p *ProcessTerminal) ClearScreen() {
	os.Stdout.WriteString("\x1b[2J\x1b[H") // Clear screen and move to home (1,1)
}

func (p *ProcessTerminal) SetTitle(title string) {
	// OSC 0;title BEL - set terminal window title
	os.Stdout.WriteString("\x1b]0;" + title + "\x07")
}
