package debug_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeeaiclub/fasttui/internal/debug"
)

func TestLogRedrawWritesWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	debug.SetLogDir(dir)
	t.Cleanup(func() { debug.SetLogDir("") })

	t.Setenv("FASTTUI_DEBUG_REDRAW", "1")

	debug.LogRedraw("first render", 0, 3, 24)

	data, err := os.ReadFile(filepath.Join(dir, "fasttui-debug.log"))
	if err != nil {
		t.Fatalf("read debug log: %v", err)
	}
	if !strings.Contains(string(data), "fullRender: first render (prev=0, new=3, height=24)") {
		t.Fatalf("unexpected log content: %q", data)
	}
}

func TestLogRedrawSilentWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	debug.SetLogDir(dir)
	t.Cleanup(func() { debug.SetLogDir("") })

	t.Setenv("FASTTUI_DEBUG_REDRAW", "")

	debug.LogRedraw("should not appear", 1, 2, 10)

	if _, err := os.Stat(filepath.Join(dir, "fasttui-debug.log")); !os.IsNotExist(err) {
		t.Fatalf("expected no debug log file, err=%v", err)
	}
}

func TestWriteRenderDump(t *testing.T) {
	// Avoid polluting /tmp/tui in unit tests: write via public dump and check shape.
	dump := debug.RenderDump{
		FirstChanged: 1,
		ViewportTop:  0,
		Height:       24,
		NewLines:     []string{"a", "b"},
		PreviousLines: []string{"a"},
	}
	debug.WriteRenderDump(dump)

	entries, err := os.ReadDir("/tmp/tui")
	if err != nil {
		t.Fatalf("read /tmp/tui: %v", err)
	}
	var found string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "render-") && strings.HasSuffix(e.Name(), ".log") {
			found = filepath.Join("/tmp/tui", e.Name())
		}
	}
	if found == "" {
		t.Fatal("expected a render-*.log file under /tmp/tui")
	}
	t.Cleanup(func() { _ = os.Remove(found) })

	data, err := os.ReadFile(found)
	if err != nil {
		t.Fatalf("read dump: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "firstChanged: 1") {
		t.Fatalf("missing firstChanged: %q", content)
	}
	if !strings.Contains(content, "=== newLines ===") {
		t.Fatalf("missing newLines section: %q", content)
	}
}

func TestResolveWriteLogPathDirectory(t *testing.T) {
	debug.ResetWriteLogPathForTest()
	t.Cleanup(debug.ResetWriteLogPathForTest)

	dir := t.TempDir()
	t.Setenv("FASTTUI_WRITE_LOG", dir)

	path := debug.ResolveWriteLogPath()
	if path == "" {
		t.Fatal("expected write log path")
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("expected path under %s, got %s", dir, path)
	}
	if !strings.Contains(filepath.Base(path), "tui-") {
		t.Fatalf("unexpected filename: %s", path)
	}

	debug.AppendWrite("hello")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read write log: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("got %q", data)
	}
}

func TestCrashLogPathUsesLogDir(t *testing.T) {
	dir := t.TempDir()
	debug.SetLogDir(dir)
	t.Cleanup(func() { debug.SetLogDir("") })

	if got := debug.CrashLogPath(); got != filepath.Join(dir, "fasttui-crash.log") {
		t.Fatalf("CrashLogPath = %q", got)
	}

	debug.WriteCrashLog("boom")
	data, err := os.ReadFile(debug.CrashLogPath())
	if err != nil {
		t.Fatalf("read crash log: %v", err)
	}
	if string(data) != "boom" {
		t.Fatalf("got %q", data)
	}
}
