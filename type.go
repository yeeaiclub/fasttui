package fasttui

// Component: render + keyboard input.
type Component interface {
	// Render returns terminal lines for the given width.
	Render(width int) []string
	// HandleInput handles raw input when focused.
	HandleInput(data string)
	// WantsKeyRelease: receive key-up events.
	WantsKeyRelease() bool
	// Invalidate clears cached render state.
	Invalidate()
}

type Terminal interface {
	Start(onInput func(data string), onResize func()) error
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

type Focusable interface {
	Component
	SetFocused(bool)
	IsFocused() bool
}

type eventKind uint8

const (
	eventRender eventKind = iota
	eventForceRender
	eventInput
	eventFocus
	eventQuery
)

type tuiEvent struct {
	kind      eventKind
	data      string
	component Component
	response  chan any
}
