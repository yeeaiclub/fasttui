package fasttui

import (
	"fmt"
	"strings"
)

// CursorManager handles hardware cursor positioning and visibility
type CursorManager struct {
	cursorRow          int
	hardwareCursorRow  int
	showHardwareCursor bool
	terminal           Terminal
}

func newCursorManager(terminal Terminal, showHardwareCursor bool) *CursorManager {
	return &CursorManager{
		cursorRow:          0,
		hardwareCursorRow:  0,
		showHardwareCursor: showHardwareCursor,
		terminal:           terminal,
	}
}

func (cm *CursorManager) Reset() {
	cm.cursorRow = 0
	cm.hardwareCursorRow = 0
}

func (cm *CursorManager) SetShowHardwareCursor(enabled bool) {
	if cm.showHardwareCursor == enabled {
		return
	}
	cm.showHardwareCursor = enabled
	if !enabled {
		cm.terminal.HideCursor()
	}
}

func (cm *CursorManager) Position(row, col, totalLines int) {
	if (row < 0 || col < 0) || totalLines <= 0 {
		cm.terminal.HideCursor()
		return
	}

	targetRow := max(0, min(row, totalLines-1))
	targetCol := max(0, col)

	rowDelta := targetRow - cm.hardwareCursorRow
	var builder strings.Builder

	if rowDelta > 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dB", rowDelta))
	} else if rowDelta < 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dA", -rowDelta))
	}

	builder.WriteString(fmt.Sprintf("\x1b[%dG", targetCol+1))
	if builder.Len() > 0 {
		cm.terminal.Write(builder.String())
	}

	cm.hardwareCursorRow = targetRow

	if cm.showHardwareCursor {
		cm.terminal.ShowCursor()
	} else {
		cm.terminal.HideCursor()
	}
}

func (cm *CursorManager) UpdatePosition(row int) {
	cm.cursorRow = row
	cm.hardwareCursorRow = row
}

func (cm *CursorManager) SetHardwareRow(row int) {
	cm.hardwareCursorRow = row
}
