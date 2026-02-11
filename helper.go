package fasttui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

// parseSizeValue parses a size value which can be either an absolute pixel value (int)
// or a percentage string (e.g., "50%"). When a percentage is provided, it calculates
// the size based on the total available space. Returns 0 for invalid or negative values.
func parseSizeValue(value any, total int) int {
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

// parseMargin parses margin values which can be either:
// - An int: applies the same margin to all sides
// - A map[string]int: specifies individual margins for top, right, bottom, left
// Returns 0 for any negative margin values or unsupported types.
func parseMargin(margin any) (marginTop, marginRight, marginBottom, marginLeft int) {
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

// extractCursorPosition searches for a cursor marker in the given lines and returns
// its row and column position. The cursor marker is removed from the line after extraction.
// Returns (-1, -1) if no cursor marker is found. The search starts from the bottom
// of the visible viewport and goes upward.
func extractCursorPosition(lines []string, height int) (int, int) {
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

// containsImage checks if a line contains image data using either:
// - Kitty graphics protocol (\x1b_G)
// - iTerm2 inline image protocol (\x1b]1337;File=)
func containsImage(line string) bool {
	return strings.Contains(line, "\x1b_G") || strings.Contains(line, "\x1b]1337;File=")
}

// applyLineRests applies segment reset codes to each line. Lines containing images
// are left unchanged, while other lines get a reset code appended to ensure proper
// terminal rendering.
func applyLineRests(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		if containsImage(line) {
			result[i] = line
		} else {
			result[i] = line + SEGMENT_RESET
		}
	}
	return result
}
