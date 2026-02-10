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

type OverlayAnchor string

const (
	AnchorCenter       OverlayAnchor = "center"
	AnchorTopLeft      OverlayAnchor = "top-left"
	AnchorTopRight     OverlayAnchor = "top-right"
	AnchorBottomLeft   OverlayAnchor = "bottom-left"
	AnchorBottomRight  OverlayAnchor = "bottom-right"
	AnchorTopCenter    OverlayAnchor = "top-center"
	AnchorBottomCenter OverlayAnchor = "bottom-center"
	AnchorLeftCenter   OverlayAnchor = "left-center"
	AnchorRightCenter  OverlayAnchor = "right-center"
)

type OverlayOption struct {
	Width     int
	MiniWidth int
	MaxHeight int
	Anchor    OverlayAnchor
	Offset    OverlayAnchor
	Row       int
	Margin    any
	Visible   func(width int, height int) bool
}
