package fasttui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRedrawDebugLogOnFirstRender(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FASTTUI_DEBUG_REDRAW", "1")

	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	tui.SetLogDir(dir)
	tui.AddChild(&concurrencyLineComponent{lines: []string{"hello"}})
	tui.Start()
	defer tui.Stop()

	time.Sleep(30 * time.Millisecond)

	data, err := os.ReadFile(filepath.Join(dir, "fasttui-debug.log"))
	if err != nil {
		t.Fatalf("read debug log: %v", err)
	}
	if !strings.Contains(string(data), "fullRender: first render") {
		t.Fatalf("expected first render log, got %q", data)
	}
	if tui.GetFullRedraws() < 1 {
		t.Fatalf("expected fullRedraws >= 1, got %d", tui.GetFullRedraws())
	}
}
