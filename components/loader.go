package components

import (
	"sync"
	"time"

	"github.com/yeeaiclub/fasttui"
)

var _ fasttui.Component = (*Loader)(nil)

// Loader component that updates every 80ms with spinning animation.
// A single background goroutine owns frame/message updates and talks to the TUI via channels;
// only Text cross-thread access uses textMu (Render vs worker).
type Loader struct {
	*Text
	textMu sync.RWMutex
	life   sync.Mutex

	running bool
	stopCh  chan struct{}
	msgCh   chan string
	doneCh  chan struct{}

	frames         []string
	ui             *fasttui.TUI
	spinnerColorFn func(string) string
	messageColorFn func(string) string
	message        string
}

// LoaderOption configures optional theming and behavior of Loader.
type LoaderOption func(*Loader)

// WithLoaderSpinnerColor sets the color function used for the spinner.
func WithLoaderSpinnerColor(fn func(string) string) LoaderOption {
	return func(l *Loader) {
		l.spinnerColorFn = fn
	}
}

// WithLoaderMessageColor sets the color function used for the message.
func WithLoaderMessageColor(fn func(string) string) LoaderOption {
	return func(l *Loader) {
		l.messageColorFn = fn
	}
}

func NewLoader(
	ui *fasttui.TUI,
	message string,
	opts ...LoaderOption,
) *Loader {
	if message == "" {
		message = "Loading..."
	}

	loader := &Loader{
		Text:    NewText("", 1, 0),
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		ui:      ui,
		message: message,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(loader)
		}
	}

	loader.Start()
	return loader
}

func (l *Loader) Render(width int) []string {
	l.textMu.RLock()
	defer l.textMu.RUnlock()
	lines := l.Text.Render(width)
	result := make([]string, 0, len(lines)+1)
	result = append(result, "")
	result = append(result, lines...)
	return result
}

func (l *Loader) Start() {
	l.life.Lock()
	defer l.life.Unlock()
	if l.running {
		return
	}
	l.running = true
	l.stopCh = make(chan struct{})
	l.msgCh = make(chan string, 8)
	l.doneCh = make(chan struct{})
	go l.loop()
}

func (l *Loader) loop() {
	stop := l.stopCh
	msgCh := l.msgCh
	done := l.doneCh
	ticker := time.NewTicker(80 * time.Millisecond)
	defer func() {
		ticker.Stop()
		close(done)
	}()

	msg := l.message
	l.paint(0, msg)
	frame := 0

	for {
		select {
		case <-stop:
			return
		case m := <-msgCh:
			msg = m
			l.paint(frame, msg)
		case <-ticker.C:
			frame = (frame + 1) % len(l.frames)
			l.paint(frame, msg)
		}
	}
}

func (l *Loader) paint(frame int, msg string) {
	line := l.buildSpinnerLine(frame, msg)
	l.textMu.Lock()
	l.Text.SetText(line)
	l.textMu.Unlock()

	if l.ui != nil {
		l.ui.TriggerRender()
	}
}

func (l *Loader) Stop() {
	l.life.Lock()
	if !l.running {
		l.life.Unlock()
		return
	}
	close(l.stopCh)
	done := l.doneCh
	l.running = false
	l.stopCh = nil
	l.msgCh = nil
	l.doneCh = nil
	l.life.Unlock()

	<-done
}

func (l *Loader) SetMessage(message string) {
	l.life.Lock()
	ch := l.msgCh
	running := l.running
	l.life.Unlock()
	if !running || ch == nil {
		return
	}
	ch <- message
}

func (l *Loader) buildSpinnerLine(frame int, message string) string {
	f := l.frames[frame]
	if l.spinnerColorFn != nil && l.messageColorFn != nil {
		return l.spinnerColorFn(f) + " " + l.messageColorFn(message)
	}
	if l.spinnerColorFn != nil {
		return l.spinnerColorFn(f) + " " + message
	}
	if l.messageColorFn != nil {
		return f + " " + l.messageColorFn(message)
	}
	return f + " " + message
}
