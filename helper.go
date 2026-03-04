package fasttui

import (
	"regexp"
	"strconv"
	"strings"
)

// Layout Parsing Functions

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

// Cursor Handling Functions

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
	return -1, -1
}

// Line Processing Functions

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

// Diff Detection Functions

// findChangedLineRange finds the range of lines that have changed between two sets of lines.
// Returns (firstChanged, lastChanged) indices, or (-1, -1) if no changes detected.
func findChangedLineRange(oldLines, newLines []string) (int, int) {
	firstChanged := -1
	lastChanged := -1
	maxLines := max(len(newLines), len(oldLines))
	for i := range maxLines {
		oldLine := ""
		newLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			if firstChanged == -1 {
				firstChanged = i
			}
			lastChanged = i
		}
	}
	return firstChanged, lastChanged
}