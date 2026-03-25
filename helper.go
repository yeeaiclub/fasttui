package fasttui

import (
	"strings"
)

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

// appendSegmentResetCodes appends segment reset codes to each line. Lines containing images
// are left unchanged, while other lines get a reset code appended to ensure proper
// terminal rendering.
func appendSegmentResetCodes(lines []string) []string {
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
