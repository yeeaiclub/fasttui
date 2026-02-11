package fasttui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)



func (t *TUI) ShowOverlay(component Component, options OverlayOption) (func(), func(bool), func() bool) {
	entryIndex := len(t.overlayStacks)
	entry := Overlay{
		component: component,
		options:   options,
		preFocus:  t.focusedComponent,
		hidden:    false,
	}
	t.overlayStacks = append(t.overlayStacks, entry)

	if t.isOverlayVisible(&t.overlayStacks[entryIndex]) {
		t.SetFocus(component)
	}
	t.terminal.HideCursor()
	t.RequestRender(false)

	hide := func() {
		index := -1
		for i := range t.overlayStacks {
			if t.overlayStacks[i].component == component {
				index = i
				break
			}
		}
		if index != -1 {
			preFocus := t.overlayStacks[index].preFocus
			t.overlayStacks = append(t.overlayStacks[:index], t.overlayStacks[index+1:]...)
			if t.focusedComponent == component {
				topVisible := t.getTopmostVisibleOverlay()
				if topVisible != nil {
					t.SetFocus(topVisible.component)
				} else {
					t.SetFocus(preFocus)
				}
			}
			if len(t.overlayStacks) == 0 {
				t.terminal.HideCursor()
			}
			t.RequestRender(false)
		}
	}

	setHidden := func(hidden bool) {
		// Find the entry in the slice
		index := -1
		for i := range t.overlayStacks {
			if t.overlayStacks[i].component == component {
				index = i
				break
			}
		}
		if index == -1 {
			return
		}

		if t.overlayStacks[index].hidden == hidden {
			return
		}
		t.overlayStacks[index].hidden = hidden
		if hidden {
			if t.focusedComponent == component {
				topVisible := t.getTopmostVisibleOverlay()
				if topVisible != nil {
					t.SetFocus(topVisible.component)
				} else {
					t.SetFocus(t.overlayStacks[index].preFocus)
				}
			}
		} else {
			if t.isOverlayVisible(&t.overlayStacks[index]) {
				t.SetFocus(component)
			}
		}
		t.RequestRender(false)
	}

	isHidden := func() bool {
		for i := range t.overlayStacks {
			if t.overlayStacks[i].component == component {
				return t.overlayStacks[i].hidden
			}
		}
		return true
	}

	return hide, setHidden, isHidden
}

func (t *TUI) HideOverlay() {
	if len(t.overlayStacks) == 0 {
		return
	}
	overlay := t.overlayStacks[len(t.overlayStacks)-1]
	t.overlayStacks = t.overlayStacks[:len(t.overlayStacks)-1]

	// Find topmost visible overlay, or fall back to preFocus
	topVisible := t.getTopmostVisibleOverlay()
	if topVisible != nil {
		t.SetFocus(topVisible.component)
	} else {
		t.SetFocus(overlay.preFocus)
	}

	if len(t.overlayStacks) == 0 {
		t.terminal.HideCursor()
	}
	t.RequestRender(false)
}

func (t *TUI) HasOverlay() bool {
	for _, entry := range t.overlayStacks {
		if t.isOverlayVisible(&entry) {
			return true
		}
	}
	return false
}

// isOverlayVisible check if an overlay entry is currently visible.
func (t *TUI) isOverlayVisible(entry *Overlay) bool {
	if entry.hidden {
		return false
	}
	if entry.options.Visible != nil {
		width, height := t.terminal.GetSize()
		return entry.options.Visible(width, height)
	}
	return true
}

// GetTopmostVisibleOverlay returns the topmost visible overlay, or nil if none.
func (t *TUI) getTopmostVisibleOverlay() *Overlay {
	for i := len(t.overlayStacks) - 1; i >= 0; i-- {
		entry := t.overlayStacks[i]
		if t.isOverlayVisible(&entry) {
			return &entry
		}
	}
	return nil
}

func (t *TUI) SetShowHardwareCursor(enabled bool) {
	if t.showHardwareCursor == enabled {
		return
	}
	t.showHardwareCursor = enabled
	if !enabled {
		t.terminal.HideCursor()
	}
	t.RequestRender(false)
}

func (t *TUI) SetClearOnShrink(enabled bool) {
	t.clearOnShrink = enabled
}

func (t *TUI) QueryCellSize() {
	if !t.terminal.IsKittyProtocolActive() {
		return
	}
	t.cellSizeQueryPending = true
	t.terminal.Write("\x1b[16t")
}

func (t *TUI) resolveAnchorRow(anchor OverlayAnchor, height int, availHeight int, marginTop int) int {
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

func (t *TUI) resolveAnchorCol(anchor OverlayAnchor, width int, availWidth int, marginLeft int) int {
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

func (t *TUI) GetFullRedraws() int {
	return t.fullRedrawCount
}

func (t *TUI) GetShowHardwareCursor() bool {
	return t.showHardwareCursor
}

func (t *TUI) positionHardwareCursor(row int, col int, totalLines int) {
	// Check if no cursor position was found (row == -1, col == -1)
	if (row < 0 || col < 0) || totalLines <= 0 {
		t.terminal.HideCursor()
		return
	}

	targetRow := max(0, min(row, totalLines-1))
	targetCol := max(0, col)

	rowDelta := targetRow - t.hardwareCursorRow
	var builder strings.Builder

	if rowDelta > 0 {
		// move down
		builder.WriteString(fmt.Sprintf("\x1b[%dB", rowDelta))
	} else if rowDelta < 0 {
		// move up
		builder.WriteString(fmt.Sprintf("\x1b[%dA", -rowDelta))
	}

	builder.WriteString(fmt.Sprintf("\x1b[%dG", targetCol+1))
	if builder.Len() > 0 {
		t.terminal.Write(builder.String())
	}

	t.hardwareCursorRow = targetRow

	if t.showHardwareCursor {
		t.terminal.ShowCursor()
	} else {
		t.terminal.HideCursor()
	}
}

func (t *TUI) parseSizeValue(value any, total int) int {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		if v <= 0 {
			return 0
		}
		return v
	case string:
		match := regexp.MustCompile(`^(\d+(?:\.\d+)?)%$`).FindStringSubmatch(v)
		if match != nil {
			percentage, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				result := int((float64(total) * percentage) / 100)
				if result <= 0 {
					return 0
				}
				return result
			}
		}
	}

	return 0
}

func (t *TUI) parseMargin(margin any) (marginTop, marginRight, marginBottom, marginLeft int) {
	if margin == nil {
		return 0, 0, 0, 0
	}

	switch v := margin.(type) {
	case int:
		return max(0, v), max(0, v), max(0, v), max(0, v)
	case map[string]int:
		return max(0, v["top"]), max(0, v["right"]), max(0, v["bottom"]), max(0, v["left"])
	default:
		return 0, 0, 0, 0
	}
}

func (t *TUI) extractCursorPosition(lines []string, height int) (int, int) {
	viewportTop := max(0, len(lines)-height)
	for row := len(lines) - 1; row >= viewportTop; row-- {
		line := lines[row]
		index := strings.Index(line, CursorMarker)
		if index != -1 {
			beforeMarker := line[:index]
			col := VisibleWidth(beforeMarker)
			lines[row] = line[:index] + line[index+len(CursorMarker):]
			return row, col
		}
	}
	return -1, -1 // Return -1, -1 to indicate no cursor found
}
