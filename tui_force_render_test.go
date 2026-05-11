package fasttui

import (
	"strings"
	"sync"
	"testing"
	"time"
)

type recordingTerminal struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (r *recordingTerminal) Start(onInput func(data string), onResize func()) error {
	return nil
}

func (r *recordingTerminal) Stop() {}

func (r *recordingTerminal) Write(data string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf.WriteString(data)
}

func (r *recordingTerminal) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}

func (r *recordingTerminal) GetSize() (int, int) {
	return 80, 24
}

func (r *recordingTerminal) IsKittyProtocolActive() bool {
	return false
}

func (r *recordingTerminal) MoveBy(lines int)      {}
func (r *recordingTerminal) HideCursor()         {}
func (r *recordingTerminal) ShowCursor()         {}
func (r *recordingTerminal) ClearLine()          {}
func (r *recordingTerminal) ClearFromCursor()    {}
func (r *recordingTerminal) ClearScreen()        {}
func (r *recordingTerminal) SetTitle(title string) {}

// TestForceRenderClearsScreenAfterStateReset verifies that ForceRender issues a
// full clear + home so repaints stay aligned when external processes have moved
// the terminal cursor.
func TestForceRenderClearsScreenAfterStateReset(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	comp := &MockComponent{}
	tui.AddChild(comp)
	tui.Start()
	defer tui.Stop()

	time.Sleep(15 * time.Millisecond)
	tui.TriggerRender()
	time.Sleep(15 * time.Millisecond)

	out1 := term.String()
	if strings.Contains(out1, "\x1b[2J\x1b[H") {
		t.Fatalf("first paint should not clear screen; got clear in output")
	}

	tui.ForceRender()
	time.Sleep(30 * time.Millisecond)

	out2 := term.String()
	if !strings.Contains(out2, "\x1b[2J\x1b[H") {
		t.Fatalf("ForceRender should emit clear+home; output tail: %q", tailQuoted(out2, 200))
	}
}

func tailQuoted(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}
