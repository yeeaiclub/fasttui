package fasttui

import "strings"

// OverlayManager handles overlay stack and composition
type OverlayManager struct {
	overlayStacks []Overlay
	terminal      Terminal
}

func newOverlayManager(terminal Terminal) *OverlayManager {
	return &OverlayManager{
		overlayStacks: make([]Overlay, 0),
		terminal:      terminal,
	}
}

func (om *OverlayManager) Show(component Component, options OverlayOption, preFocus Component, onFocusChange func(Component)) (func(), func(bool), func() bool) {
	entryIndex := len(om.overlayStacks)
	entry := Overlay{
		component: component,
		options:   options,
		preFocus:  preFocus,
		hidden:    false,
	}
	om.overlayStacks = append(om.overlayStacks, entry)

	if om.isVisible(&om.overlayStacks[entryIndex]) {
		onFocusChange(component)
	}
	om.terminal.HideCursor()

	hide := func() {
		index := om.findOverlayIndex(component)
		if index != -1 {
			preFocus := om.overlayStacks[index].preFocus
			om.overlayStacks = append(om.overlayStacks[:index], om.overlayStacks[index+1:]...)

			topVisible := om.getTopmostVisible()
			if topVisible != nil {
				onFocusChange(topVisible.component)
			} else {
				onFocusChange(preFocus)
			}

			if len(om.overlayStacks) == 0 {
				om.terminal.HideCursor()
			}
		}
	}

	setHidden := func(hidden bool) {
		index := om.findOverlayIndex(component)
		if index == -1 || om.overlayStacks[index].hidden == hidden {
			return
		}

		om.overlayStacks[index].hidden = hidden
		if hidden {
			topVisible := om.getTopmostVisible()
			if topVisible != nil {
				onFocusChange(topVisible.component)
			} else {
				onFocusChange(om.overlayStacks[index].preFocus)
			}
		} else {
			if om.isVisible(&om.overlayStacks[index]) {
				onFocusChange(component)
			}
		}
	}

	isHidden := func() bool {
		index := om.findOverlayIndex(component)
		if index == -1 {
			return true
		}
		return om.overlayStacks[index].hidden
	}

	return hide, setHidden, isHidden
}

func (om *OverlayManager) Hide(onFocusChange func(Component)) {
	if len(om.overlayStacks) == 0 {
		return
	}

	overlay := om.overlayStacks[len(om.overlayStacks)-1]
	om.overlayStacks = om.overlayStacks[:len(om.overlayStacks)-1]

	topVisible := om.getTopmostVisible()
	if topVisible != nil {
		onFocusChange(topVisible.component)
	} else {
		onFocusChange(overlay.preFocus)
	}

	if len(om.overlayStacks) == 0 {
		om.terminal.HideCursor()
	}
}

func (om *OverlayManager) HasVisible() bool {
	for i := range om.overlayStacks {
		if om.isVisible(&om.overlayStacks[i]) {
			return true
		}
	}
	return false
}

func (om *OverlayManager) Composite(baseLines []string, width, height, maxLinesRendered int, layoutResolver OverlayLayoutResolver) []string {
	if len(om.overlayStacks) == 0 {
		return baseLines
	}

	result := make([]string, len(baseLines))
	copy(result, baseLines)

	type renderedOverlay struct {
		overlayLines []string
		row          int
		col          int
		w            int
	}

	rendered := make([]renderedOverlay, 0)
	minLinesNeeded := len(result)

	for i := range om.overlayStacks {
		entry := &om.overlayStacks[i]
		if !om.isVisible(entry) {
			continue
		}

		layout := layoutResolver.Resolve(entry.options, 0, width, height)
		overlayLines := entry.component.Render(layout.width)

		if layout.maxHeight != nil && len(overlayLines) > *layout.maxHeight {
			overlayLines = overlayLines[:*layout.maxHeight]
		}

		finalLayout := layoutResolver.Resolve(entry.options, len(overlayLines), width, height)

		rendered = append(rendered, renderedOverlay{
			overlayLines: overlayLines,
			row:          finalLayout.row,
			col:          finalLayout.col,
			w:            finalLayout.width,
		})

		if finalLayout.row+len(overlayLines) > minLinesNeeded {
			minLinesNeeded = finalLayout.row + len(overlayLines)
		}
	}

	workingHeight := max(maxLinesRendered, minLinesNeeded)
	for len(result) < workingHeight {
		result = append(result, "")
	}

	viewportStart := max(0, workingHeight-height)
	modifiedLines := make(map[int]bool)

	for _, ro := range rendered {
		for i := range ro.overlayLines {
			idx := viewportStart + ro.row + i
			if idx >= 0 && idx < len(result) {
				truncatedOverlayLine := ro.overlayLines[i]
				if VisibleWidth(ro.overlayLines[i]) > ro.w {
					truncatedOverlayLine = SliceByColumn(ro.overlayLines[i], 0, ro.w, true)
				}
				result[idx] = compositeLineAt(result[idx], truncatedOverlayLine, ro.col, ro.w, width)
				modifiedLines[idx] = true
			}
		}
	}

	for idx := range modifiedLines {
		if VisibleWidth(result[idx]) > width {
			result[idx] = SliceByColumn(result[idx], 0, width, true)
		}
	}

	return result
}

func (om *OverlayManager) isVisible(entry *Overlay) bool {
	if entry.hidden {
		return false
	}
	if entry.options.Visible != nil {
		width, height := om.terminal.GetSize()
		return entry.options.Visible(width, height)
	}
	return true
}

func (om *OverlayManager) getTopmostVisible() *Overlay {
	for i := len(om.overlayStacks) - 1; i >= 0; i-- {
		entry := om.overlayStacks[i]
		if om.isVisible(&entry) {
			return &entry
		}
	}
	return nil
}

func (om *OverlayManager) findOverlayIndex(component Component) int {
	for i := range om.overlayStacks {
		if om.overlayStacks[i].component == component {
			return i
		}
	}
	return -1
}

func (om *OverlayManager) GetTopmostVisibleComponent() Component {
	overlay := om.getTopmostVisible()
	if overlay != nil {
		return overlay.component
	}
	return nil
}

func compositeLineAt(baseLine string, overlayLine string, col int, overlayWidth int, termWidth int) string {
	if containsImage(baseLine) {
		return baseLine
	}

	afterStart := col + overlayWidth
	before, beforeWidth, after, afterWidth := ExtractSegments(baseLine, col, afterStart, termWidth-afterStart, true)

	overlay := SliceWithWidth(overlayLine, 0, overlayWidth, true)

	beforePad := max(0, col-beforeWidth)
	overlayPad := max(0, overlayWidth-overlay.width)
	actualBeforeWidth := max(col, beforeWidth)
	actualOverlayWidth := max(overlayWidth, overlay.width)
	afterTarget := max(0, termWidth-actualBeforeWidth-actualOverlayWidth)
	afterPad := max(0, afterTarget-afterWidth)

	var result strings.Builder
	result.WriteString(before)
	result.WriteString(strings.Repeat(" ", beforePad))
	result.WriteString(SEGMENT_RESET)
	result.WriteString(overlay.text)
	result.WriteString(strings.Repeat(" ", overlayPad))
	result.WriteString(SEGMENT_RESET)
	result.WriteString(after)
	result.WriteString(strings.Repeat(" ", afterPad))

	resultStr := result.String()
	resultWidth := VisibleWidth(resultStr)
	if resultWidth <= termWidth {
		return resultStr
	}
	return SliceByColumn(resultStr, 0, termWidth, true)
}
