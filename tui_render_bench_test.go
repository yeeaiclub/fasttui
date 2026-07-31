//go:build dirty

package fasttui

import (
	"fmt"
	"strings"
	"testing"
)

type benchTerminal struct {
	w, h    int
	written int
}

func (m *benchTerminal) Start(func(string), func()) error { return nil }
func (m *benchTerminal) Stop()                            {}
func (m *benchTerminal) Write(data string)                { m.written += len(data) }
func (m *benchTerminal) WriteBytes(data []byte)           { m.written += len(data) }
func (m *benchTerminal) GetSize() (int, int)              { return m.w, m.h }
func (m *benchTerminal) IsKittyProtocolActive() bool      { return false }
func (m *benchTerminal) MoveBy(int)                       {}
func (m *benchTerminal) HideCursor()                      {}
func (m *benchTerminal) ShowCursor()                      {}
func (m *benchTerminal) ClearLine()                       {}
func (m *benchTerminal) ClearFromCursor()                 {}
func (m *benchTerminal) ClearScreen()                     {}
func (m *benchTerminal) SetTitle(string)                  {}

type benchCachedComponent struct {
	lines       []string
	cacheValid  bool
	cachedWidth int
	cachedLines []string
	renders     int
}

func (c *benchCachedComponent) Render(width int) []string {
	if c.cacheValid && c.cachedWidth == width && c.cachedLines != nil {
		return c.cachedLines
	}
	c.renders++
	out := make([]string, len(c.lines))
	copy(out, c.lines)
	c.cacheValid = true
	c.cachedWidth = width
	c.cachedLines = out
	return out
}
func (c *benchCachedComponent) HandleInput(string)    {}
func (c *benchCachedComponent) WantsKeyRelease() bool { return false }
func (c *benchCachedComponent) Invalidate() {
	c.cacheValid = false
	c.cachedLines = nil
}
func (c *benchCachedComponent) SetLines(lines []string) {
	c.lines = lines
	c.Invalidate()
}

func buildStaticLines(n int, width int) []string {
	lines := make([]string, n)
	pad := strings.Repeat("x", max(1, width-10))
	for i := range n {
		lines[i] = fmt.Sprintf("line-%04d-%s", i, pad)
		if len(lines[i]) > width {
			lines[i] = lines[i][:width]
		}
	}
	return lines
}

// BenchmarkDoRender_Idle measures the short-circuit path when nothing is dirty.
func BenchmarkDoRender_Idle(b *testing.B) {
	term := &benchTerminal{w: 80, h: 40}
	tui := NewTUI(term, false)
	tui.AddChild(&concurrencyLineComponent{lines: buildStaticLines(200, 80)})
	tui.doRender() // first paint
	if tui.dirty.Load() {
		b.Fatal("expected dirty cleared after first paint")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		tui.doRender()
	}
}

// BenchmarkDoRender_Spinner marks dirty each frame while a large cached subtree
// stays unchanged (approximates loader animation over a chat history).
func BenchmarkDoRender_Spinner(b *testing.B) {
	term := &benchTerminal{w: 80, h: 40}
	tui := NewTUI(term, false)

	static := &benchCachedComponent{lines: buildStaticLines(150, 80)}
	tui.AddChild(static)

	editor := &benchCachedComponent{lines: []string{"hello editor cache", "second line"}}
	tui.AddChild(editor)

	spinner := &benchCachedComponent{lines: []string{"⠋ Thinking..."}}
	tui.AddChild(spinner)
	tui.doRender()

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		spinner.SetLines([]string{frames[i%len(frames)] + " Thinking..."})
		tui.MarkDirty()
		select {
		case <-tui.eventChan:
		default:
		}
		tui.doRender()
	}
}

// BenchmarkDoRender_EditorKey measures a single-line content edit against a
// large static history above the "editor".
func BenchmarkDoRender_EditorKey(b *testing.B) {
	term := &benchTerminal{w: 80, h: 40}
	tui := NewTUI(term, false)

	tui.AddChild(&benchCachedComponent{lines: buildStaticLines(120, 80)})

	editor := &benchCachedComponent{lines: []string{""}}
	tui.AddChild(editor)
	tui.dirty.Store(true)
	tui.doRender()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		ch := string(rune('a' + (i % 26)))
		editor.SetLines([]string{ch})
		tui.dirty.Store(true)
		tui.doRender()
	}
}
