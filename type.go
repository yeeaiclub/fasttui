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

type RenderRequest struct {
	force bool
}

type InputRequest struct {
	data string
}

type FocusRequest struct {
	component Component
}

type QueryRequest struct {
	action   string // "getShowHardwareCursor", "getFullRedraws"
	response chan any
}
