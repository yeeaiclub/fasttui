package fasttui

type Component interface {
	Render(width int) []string
	HandleInput(data string)
	WantsKeyRelease() bool
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

type OverlayHandle interface {
	Hide()
	SetHidden(hidden bool)
	isHidden() bool
}

type OverlayStack struct {
	component Component
	options   OverlayOption
	preFocus  Component
	hidden    bool
}

type OverlayOption struct {
	Width int
}

func isFocused(component Component) bool {
	if focusable, ok := component.(Focusable); ok {
		return focusable.IsFocused()
	}
	return false
}
