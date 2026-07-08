//go:build integration

package integration_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
)

type sizedRecordingTerminal struct {
	width, height int
	written       strings.Builder
}

func (t *sizedRecordingTerminal) Start(func(string), func()) error { return nil }
func (t *sizedRecordingTerminal) Stop()                              {}
func (t *sizedRecordingTerminal) Write(data string)                { t.written.WriteString(data) }
func (t *sizedRecordingTerminal) GetSize() (int, int)              { return t.width, t.height }
func (t *sizedRecordingTerminal) IsKittyProtocolActive() bool      { return false }
func (t *sizedRecordingTerminal) MoveBy(int)                       {}
func (t *sizedRecordingTerminal) HideCursor()                      {}
func (t *sizedRecordingTerminal) ShowCursor()                      {}
func (t *sizedRecordingTerminal) ClearLine()                       {}
func (t *sizedRecordingTerminal) ClearFromCursor()                 {}
func (t *sizedRecordingTerminal) ClearScreen()                     {}
func (t *sizedRecordingTerminal) SetTitle(string)                  {}

func (t *sizedRecordingTerminal) reset() string {
	prev := t.written.String()
	t.written.Reset()
	return prev
}

func waitForRender() {
	time.Sleep(25 * time.Millisecond)
}

func TestEditorBackspaceNoCursorMarkerLeakViaTUI(t *testing.T) {
	term := &sizedRecordingTerminal{width: 80, height: 24}
	tui := fasttui.NewTUI(term, true)

	editor := components.NewEditor(term, nil)
	tui.AddChild(editor)
	tui.SetFocus(editor)

	tui.Start()
	defer tui.Stop()
	waitForRender()

	line := `!git commit -m "   feat(prisma): add Prisma node, DTOs, and img2pdf dependency` +
		strings.Repeat(" ", 200) + `"`
	editor.SetText([]string{line})
	editor.SetCursor(0, 79)

	tui.TriggerRender()
	waitForRender()

	marker := fasttui.GetCursorMarker()
	for i := 0; i < 30; i++ {
		out := term.reset()
		tui.TriggerRender()
		waitForRender()
		out += term.written.String()

		if strings.Contains(out, "i:c") {
			t.Fatalf("step %d: terminal leaked cursor marker fragment i:c\noutput tail: %q", i, tail(out, 300))
		}
		if strings.Contains(out, marker) {
			t.Fatalf("step %d: terminal leaked raw cursor marker\noutput tail: %q", i, tail(out, 300))
		}

		tui.HandleInput("\x7f") // backspace
		waitForRender()
	}
}

func TestEditorPasteAndBackspaceViaTUI(t *testing.T) {
	term := &sizedRecordingTerminal{width: 80, height: 24}
	tui := fasttui.NewTUI(term, true)

	editor := components.NewEditor(term, nil)
	tui.AddChild(editor)
	tui.SetFocus(editor)

	tui.Start()
	defer tui.Stop()
	waitForRender()

	paste := "!git commit -m \"   feat(prisma): add Prisma node, DTOs, and img2pdf dependency" +
		strings.Repeat(" ", 100) + "\"\n\n```\n   feat(prisma): add Prisma node\n```"

	tui.HandleInput("\x1b[200~" + paste + "\x1b[201~")
	waitForRender()

	marker := fasttui.GetCursorMarker()
	for i := 0; i < 40; i++ {
		out := term.reset()
		tui.TriggerRender()
		waitForRender()
		out += term.written.String()

		if strings.Contains(out, "i:c") || strings.Contains(out, marker) {
			t.Fatalf("step %d: cursor marker leaked after paste/backspace\noutput tail: %q", i, tail(out, 300))
		}

		tui.HandleInput("\x7f")
		waitForRender()
	}
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
