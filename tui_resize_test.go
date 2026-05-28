package fasttui

import (
	"strings"
	"testing"
)

// TestMoveHardwareCursorUsesContentLineCount verifies that cursor movement is
// clamped to the rendered content height, not the terminal viewport height.
// Using terminal height caused wrong cursor/border alignment after resize when
// chat history pushed total lines above the visible area.
func TestMoveHardwareCursorUsesContentLineCount(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, true)

	// Simulate cursor at end of a 50-line full redraw.
	tui.hardwareCursorRow = 49

	tui.moveHardwareCursorTo(45, 10, 50)
	got := term.String()
	if !strings.Contains(got, "\x1b[4A") {
		t.Fatalf("expected move up 4 lines for 50-line content, got %q", got)
	}
	if !strings.Contains(got, "\x1b[11G") {
		t.Fatalf("expected column 11, got %q", got)
	}

	term.buf.Reset()
	tui.hardwareCursorRow = 49

	// Wrong bound (terminal height 24) clamps row to 23 and over-corrects.
	tui.moveHardwareCursorTo(45, 10, 24)
	gotWrong := term.String()
	if strings.Contains(gotWrong, "\x1b[4A") {
		t.Fatalf("terminal-height clamp should not produce 4-line move, got %q", gotWrong)
	}
	if !strings.Contains(gotWrong, "\x1b[26A") {
		t.Fatalf("expected large upward correction when wrongly clamped, got %q", gotWrong)
	}
}

func TestFullRendererCursorUsesContentLines(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, true)

	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	lines[45] = "cursor" + CursorMarker + "here"

	row, col := extractCursorPosition(lines, 24)
	if row != 45 || col != 6 {
		t.Fatalf("extractCursorPosition: got row=%d col=%d", row, col)
	}

	fr := tui.getFullRender(lines, 24, row, col, 80)
	tui.hardwareCursorRow = 0
	fr.Render(false)

	got := term.String()
	if !strings.Contains(got, "\x1b[4A") {
		t.Fatalf("full render should move cursor relative to content lines, got %q", got)
	}
}
