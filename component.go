package fasttui

type Component interface {
	Render(width int) []string

	HandleInput(data string)

	WantsKeyRelease() bool

	Invalidate()
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
	options   any
	preFocus  Component
	hidden    bool
}
