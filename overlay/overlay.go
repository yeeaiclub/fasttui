package overlay

import "github.com/yeeaiclub/fasttui"

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
	Col       int
	OffsetX   int
	OffsetY   int
	Margin    any
	Visible   func(width int, height int) bool
}

type OverlayLayout struct {
	width     int
	row       int
	col       int
	maxHeight *int
}

func NewOverlayLayout(width int, row int, col int, maxHeight *int) *OverlayLayout {
	return &OverlayLayout{
		width:     width,
		row:       row,
		col:       col,
		maxHeight: maxHeight,
	}
}

type Overlay struct {
	component fasttui.Component
	options   OverlayOption
	preFocus  fasttui.Component
	hidden    bool
}
