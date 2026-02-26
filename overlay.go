package fasttui

type Overlay struct {
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

func (anchor OverlayAnchor) getCol(width int, availWidth int, marginLeft int) int {
	switch anchor {
	case AnchorTopLeft, AnchorBottomLeft, AnchorLeftCenter:
		return marginLeft
	case AnchorTopRight, AnchorBottomRight, AnchorRightCenter:
		return marginLeft + max(0, availWidth-width)
	case AnchorTopCenter, AnchorBottomCenter, AnchorCenter:
		return marginLeft + max(0, availWidth-width)/2
	default:
		return marginLeft
	}
}

func (anchor OverlayAnchor) getRow(height int, availHeight int, marginTop int) int {
	switch anchor {
	case AnchorTopLeft, AnchorTopCenter, AnchorTopRight:
		return marginTop
	case AnchorBottomLeft, AnchorBottomCenter, AnchorBottomRight:
		return marginTop + max(0, availHeight-height)
	case AnchorLeftCenter, AnchorRightCenter, AnchorCenter:
		return marginTop + max(0, availHeight-height)/2
	default:
		return marginTop
	}
}

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

func (o OverlayOption) ResolveLayout(overlayHeight int, termWidth int, termHeight int) OverlayLayout {
	marginTop, marginRight, marginBottom, marginLeft := parseMargin(o.Margin)

	availWidth := max(1, termWidth-marginLeft-marginRight)
	availHeight := max(1, termHeight-marginTop-marginBottom)

	width := parseSizeValue(o.Width, termWidth)
	if width == 0 {
		width = min(80, availWidth)
	}
	if o.MiniWidth > 0 {
		width = max(width, o.MiniWidth)
	}
	width = max(1, min(width, availWidth))

	var maxHeight *int
	if o.MaxHeight > 0 {
		maxHeightVal := parseSizeValue(o.MaxHeight, termHeight)
		if maxHeightVal > 0 {
			maxHeightVal = max(1, min(maxHeightVal, availHeight))
			maxHeight = &maxHeightVal
		}
	}

	effectiveHeight := overlayHeight
	if maxHeight != nil {
		effectiveHeight = min(overlayHeight, *maxHeight)
	}

	var row, col int

	if o.Row != 0 {
		row = o.Row
	} else {
		anchor := o.Anchor
		if anchor == "" {
			anchor = AnchorCenter
		}
		row = anchor.getRow(effectiveHeight, availHeight, marginTop)
	}

	if o.Col != 0 {
		col = o.Col
	} else {
		anchor := o.Anchor
		if anchor == "" {
			anchor = AnchorCenter
		}
		col = anchor.getCol(width, availWidth, marginLeft)
	}

	if o.OffsetY != 0 {
		row += o.OffsetY
	}
	if o.OffsetX != 0 {
		col += o.OffsetX
	}

	row = max(marginTop, min(row, termHeight-marginBottom-effectiveHeight))
	col = max(marginLeft, min(col, termWidth-marginRight-width))

	return OverlayLayout{
		width:     width,
		row:       row,
		col:       col,
		maxHeight: maxHeight,
	}
}
