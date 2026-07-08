package components

import (
	"strings"
	"testing"

	"github.com/yeeaiclub/fasttui"
)

type mockEditorTerm struct{ w, h int }

func (m *mockEditorTerm) GetSize() (int, int)              { return m.w, m.h }
func (m *mockEditorTerm) Start(func(string), func()) error { return nil }
func (m *mockEditorTerm) Stop()                            {}
func (m *mockEditorTerm) Write(string)                     {}
func (m *mockEditorTerm) IsKittyProtocolActive() bool      { return false }
func (m *mockEditorTerm) MoveBy(int)                       {}
func (m *mockEditorTerm) HideCursor()                      {}
func (m *mockEditorTerm) ShowCursor()                      {}
func (m *mockEditorTerm) ClearLine()                       {}
func (m *mockEditorTerm) ClearFromCursor()                 {}
func (m *mockEditorTerm) ClearScreen()                     {}
func (m *mockEditorTerm) SetTitle(string)                  {}

func renderEditorLinesForTest(t *testing.T, e *Editor, width, height int) []string {
	t.Helper()
	lines := e.Render(width)
	removeCursorMarkersForTest(lines, height)
	return lines
}

func removeCursorMarkersForTest(lines []string, height int) {
	marker := fasttui.GetCursorMarker()
	viewportTop := max(0, len(lines)-height)
	for row := len(lines) - 1; row >= viewportTop; row-- {
		line := lines[row]
		for {
			index := strings.Index(line, marker)
			if index == -1 {
				break
			}
			line = line[:index] + line[index+len(marker):]
		}
		lines[row] = line
	}
}

func assertNoCursorMarkerLeak(t *testing.T, lines []string) {
	t.Helper()
	marker := fasttui.GetCursorMarker()
	for li, line := range lines {
		if strings.Contains(line, marker) {
			t.Fatalf("line %d still contains cursor marker: %q", li, line)
		}
		if strings.Contains(line, "i:c") {
			t.Fatalf("line %d contains leaked marker fragment i:c: %q", li, line)
		}
	}
}

func countCursorMarkers(lines []string) int {
	marker := fasttui.GetCursorMarker()
	count := 0
	for _, line := range lines {
		count += strings.Count(line, marker)
	}
	return count
}
