package fasttui

// OverlayLayoutResolver resolves overlay positioning and sizing
type OverlayLayoutResolver interface {
	Resolve(options OverlayOption, overlayHeight, termWidth, termHeight int) OverlayLayout
}

type DefaultLayoutResolver struct{}

func (r *DefaultLayoutResolver) Resolve(options OverlayOption, overlayHeight, termWidth, termHeight int) OverlayLayout {
	marginTop, marginRight, marginBottom, marginLeft := parseMargin(options.Margin)

	availWidth := max(1, termWidth-marginLeft-marginRight)
	availHeight := max(1, termHeight-marginTop-marginBottom)

	width := parseSizeValue(options.Width, termWidth)
	if width == 0 {
		width = min(80, availWidth)
	}
	if options.MiniWidth > 0 {
		width = max(width, options.MiniWidth)
	}
	width = max(1, min(width, availWidth))

	var maxHeight *int
	if options.MaxHeight > 0 {
		maxHeightVal := parseSizeValue(options.MaxHeight, termHeight)
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

	if options.Row != 0 {
		row = options.Row
	} else {
		anchor := options.Anchor
		if anchor == "" {
			anchor = AnchorCenter
		}
		row = resolveAnchorRow(anchor, effectiveHeight, availHeight, marginTop)
	}

	if options.Col != 0 {
		col = options.Col
	} else {
		anchor := options.Anchor
		if anchor == "" {
			anchor = AnchorCenter
		}
		col = resolveAnchorCol(anchor, width, availWidth, marginLeft)
	}

	if options.OffsetY != 0 {
		row += options.OffsetY
	}
	if options.OffsetX != 0 {
		col += options.OffsetX
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

func resolveAnchorRow(anchor OverlayAnchor, height int, availHeight int, marginTop int) int {
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

func resolveAnchorCol(anchor OverlayAnchor, width int, availWidth int, marginLeft int) int {
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
