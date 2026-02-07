package fasttui

import (
	"bufio"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/term"
)

type Terminal interface {
	Start(onInput func(data string), onResize func())
	Stop()
	Write(data string)
	GetSize() (int, int)
	IsKittyProtocolActive() bool
	MoveBy(lines int)
	HideCursor()
	ShowCursor()
	ClearLine()
	ClearFromCursor()
	ClearScreen()
	SetTitle(title string)
}

var kittyPattern = regexp.MustCompile(`^\x1b\[\?(\d+)u$`)

type ProcessTerminal struct {
	wasRaw              bool
	writeLogPath        string
	inputHandler        func(data string)
	resizeHandler       func()
	stdinFd             int
	oldState            *term.State
	buffer              *StdinBuffer
	kittyProtocolActive bool
}

func NewProcessTerminal() (*ProcessTerminal, error) {
	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		return &ProcessTerminal{
			stdinFd:             fd,
			kittyProtocolActive: false,
		}, nil
	}
	return &ProcessTerminal{wasRaw: false, kittyProtocolActive: false}, nil
}

func (p *ProcessTerminal) GetSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return w, h
}

func (p *ProcessTerminal) IsKittyProtocolActive() bool {
	return p.kittyProtocolActive
}

func (p *ProcessTerminal) Start(onInput func(data string), onResize func()) {
	p.inputHandler = onInput
	p.resizeHandler = onResize

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	p.oldState = oldState
	p.wasRaw = true

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		for range sigChan {
			if p.resizeHandler != nil {
				p.resizeHandler()
			}
		}
	}()
}

func (p *ProcessTerminal) setupStdinBuffer() {
	p.buffer = NewStdinBuffer()
	p.buffer.OnData = func(seq string) {
		if !p.kittyProtocolActive {
			if matches := kittyPattern.FindStringSubmatch(seq); matches != nil {
				p.kittyProtocolActive = true
				os.Stdout.WriteString("\x1b[>7u")
				return
			}
		}
		if p.inputHandler != nil {
			p.inputHandler(seq)
		}
	}
	go p.readStdin()
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
	if p.kittyProtocolActive {
		os.Stdout.WriteString("\x1b[<u")
		p.kittyProtocolActive = false
	}

	previousHandler := p.inputHandler
	p.inputHandler = nil

	lastDataTime := time.Now()
	var oldOnData func(seq string)
	if p.buffer != nil {
		oldOnData = p.buffer.OnData
		p.buffer.OnData = func(seq string) {
			lastDataTime = time.Now()
			if oldOnData != nil {
				oldOnData(seq)
			}
		}
	}

	endTime := time.Now().Add(time.Duration(maxMs) * time.Millisecond)
	idleDuration := time.Duration(idleMs) * time.Millisecond

	for {
		now := time.Now()
		if now.After(endTime) {
			break
		}
		if now.Sub(lastDataTime) >= idleDuration {
			break
		}
		sleepDuration := idleDuration
		if endTime.Sub(now) < sleepDuration {
			sleepDuration = endTime.Sub(now)
		}
		time.Sleep(sleepDuration)
	}

	if p.buffer != nil {
		p.buffer.OnData = oldOnData
	}
	p.inputHandler = previousHandler
	return nil
}

func (p *ProcessTerminal) queryAndEnableKittyProtocol() {
	p.setupStdinBuffer()
	os.Stdout.WriteString("\x1b[?u")
}

func (p *ProcessTerminal) Stop() {
	os.Stdout.WriteString("\x1b[?2004l")
	p.inputHandler = nil
	p.resizeHandler = nil
}

func (p *ProcessTerminal) Write(data string) {
	_, err := os.Stdout.WriteString(data)
	if err != nil {
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
