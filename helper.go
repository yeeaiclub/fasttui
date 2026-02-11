package fasttui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func (t *TUI) SetFocus(component Component) {
	if t.focusedComponent != nil {
		if f, ok := t.focusedComponent.(Focusable); ok {
			f.SetFocused(false)
		}
	}

	t.focusedComponent = component

	if component != nil {
		if f, ok := t.focusedComponent.(Focusable); ok {
			f.SetFocused(true)
		}
	}
}

func (t *TUI) ShowOverlay(component Component, options OverlayOption) (func(), func(bool), func() bool) {
	entry := Overlay{
		component: component,
		options:   options,
		preFocus:  t.focusedComponent,
		hidden:    false,
	}
	t.overlayStacks = append(t.overlayStacks, entry)

	if t.isOverlayVisible(&entry) {
		t.SetFocus(entry.component)
	}
	t.terminal.HideCursor()
	t.requestRender(false)

	hide := func() {
		index := -1
		for i, e := range t.overlayStacks {
			if e.component == component {
				index = i
				break
			}
		}
		if index != -1 {
			t.overlayStacks = append(t.overlayStacks[:index], t.overlayStacks[index+1:]...)
			if t.focusedComponent == component {
				topVisible := t.getTopmostVisibleOverlay()
				if topVisible != nil {
					t.SetFocus(topVisible.component)
				} else {
					t.SetFocus(entry.preFocus)
				}
			}
			if len(t.overlayStacks) == 0 {
				t.terminal.HideCursor()
			}
			t.requestRender(false)
		}
	}

	setHidden := func(hidden bool) {
		if entry.hidden == hidden {
			return
		}
		entry.hidden = hidden
		if hidden {
			if t.focusedComponent == component {
				topVisible := t.getTopmostVisibleOverlay()
				if topVisible != nil {
					t.SetFocus(topVisible.component)
				} else {
					t.SetFocus(entry.preFocus)
				}
			}
		} else {
			if t.isOverlayVisible(&entry) {
				t.SetFocus(component)
			}
		}
		t.requestRender(false)
	}

	return hide, setHidden, func() bool { return entry.hidden }
}

func (t *TUI) HideOverlay() {
	// POP last overlay
	if len(t.overlayStacks) > 0 {
		entry := t.overlayStacks[len(t.overlayStacks)-1]
		t.overlayStacks = t.overlayStacks[:len(t.overlayStacks)-1]
		if !entry.hidden {
			t.SetFocus(entry.preFocus)
		}
	}
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
	t.requestRender(false)
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
	if (row == 0 && col == 0) || totalLines <= 0 {
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
