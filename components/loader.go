package components

import (
	"time"
)

// UIRenderer interface for components that need to trigger UI updates
type UIRenderer interface {
	RequestRender(force bool)
}

// Loader component that updates every 80ms with spinning animation
type Loader struct {
	*Text
	frames         []string
	currentFrame   int
	ticker         *time.Ticker
	stopChan       chan struct{}
	ui             UIRenderer
	spinnerColorFn func(string) string
	messageColorFn func(string) string
	message        string
	running        bool
}

func NewLoader(
	ui UIRenderer,
	spinnerColorFn func(string) string,
	messageColorFn func(string) string,
	message string,
) *Loader {
	if message == "" {
		message = "Loading..."
	}

	loader := &Loader{
		Text:           NewText("", 1, 0, nil),
		frames:         []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		currentFrame:   0,
		ui:             ui,
		spinnerColorFn: spinnerColorFn,
		messageColorFn: messageColorFn,
		message:        message,
		stopChan:       make(chan struct{}),
	}

	loader.Start()
	return loader
}

func (l *Loader) Render(width int) []string {
	lines := l.Text.Render(width)
	result := make([]string, 0, len(lines)+1)
	result = append(result, "")
	result = append(result, lines...)
	return result
}

func (l *Loader) Start() {
	if l.running {
		return
	}

	l.running = true
	l.updateDisplay()

	l.ticker = time.NewTicker(80 * time.Millisecond)
	ticker := l.ticker
	stopChan := l.stopChan

	go func() {
		for {
			select {
			case <-ticker.C:
				if !l.running {
					return
				}
				l.currentFrame = (l.currentFrame + 1) % len(l.frames)
				l.updateDisplay()
			case <-stopChan:
				return
			}
		}
	}()
}

func (l *Loader) Stop() {
	if !l.running {
		return
	}

	l.running = false
	if l.ticker != nil {
		l.ticker.Stop()
		l.ticker = nil
	}
	close(l.stopChan)
	l.stopChan = make(chan struct{})
}

func (l *Loader) SetMessage(message string) {
	l.message = message
	l.updateDisplay()
}

func (l *Loader) updateDisplay() {
	frame := l.frames[l.currentFrame]
	var text string

	if l.spinnerColorFn != nil && l.messageColorFn != nil {
		text = l.spinnerColorFn(frame) + " " + l.messageColorFn(l.message)
	} else if l.spinnerColorFn != nil {
		text = l.spinnerColorFn(frame) + " " + l.message
	} else if l.messageColorFn != nil {
		text = frame + " " + l.messageColorFn(l.message)
	} else {
		text = frame + " " + l.message
	}

	l.Text.SetText(text)

	if l.ui != nil {
		l.ui.RequestRender(false)
	}
}
