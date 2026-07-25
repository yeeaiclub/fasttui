// Package debug provides file-based diagnostic logging for fasttui.
//
// Logging is opt-in via environment variables and never writes to stdout
// (which would corrupt the TUI). All writes are synchronous.
//
// Environment variables:
//
//	FASTTUI_DEBUG_REDRAW=1  — append full-redraw reasons to {logDir}/fasttui-debug.log
//	FASTTUI_DEBUG=1         — dump each differential render snapshot under /tmp/tui/
//	FASTTUI_WRITE_LOG=<path|dir> — capture raw ANSI written to the terminal
//	FASTTUI_LOG_DIR=<dir>   — override log directory (default: ~/.fasttui)
package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	envDebugRedraw = "FASTTUI_DEBUG_REDRAW"
	envDebug       = "FASTTUI_DEBUG"
	envWriteLog    = "FASTTUI_WRITE_LOG"
	envLogDir      = "FASTTUI_LOG_DIR"

	debugLogName = "fasttui-debug.log"
	crashLogName = "fasttui-crash.log"
	renderDumpDir = "/tmp/tui"
)

var (
	logDirMu   sync.RWMutex
	logDirOverride string

	writeLogPathOnce sync.Once
	writeLogPath     string
)

// SetLogDir overrides the log directory used for redraw and crash logs.
// An empty string clears the override and falls back to env / default.
func SetLogDir(dir string) {
	logDirMu.Lock()
	defer logDirMu.Unlock()
	logDirOverride = dir
}

// LogDir returns the directory for debug and crash log files.
// Priority: SetLogDir > FASTTUI_LOG_DIR > ~/.fasttui
func LogDir() string {
	logDirMu.RLock()
	override := logDirOverride
	logDirMu.RUnlock()
	if override != "" {
		return override
	}
	if dir := os.Getenv(envLogDir); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".fasttui")
}

// RedrawEnabled reports whether full-redraw reason logging is on.
func RedrawEnabled() bool {
	return os.Getenv(envDebugRedraw) == "1"
}

// RenderDumpEnabled reports whether differential render snapshots are written.
func RenderDumpEnabled() bool {
	return os.Getenv(envDebug) == "1"
}

// DebugLogPath returns the path of the append-only debug log.
func DebugLogPath() string {
	return filepath.Join(LogDir(), debugLogName)
}

// CrashLogPath returns the path of the crash dump file.
func CrashLogPath() string {
	return filepath.Join(LogDir(), crashLogName)
}

// Logf appends a timestamped line to the debug log when redraw logging is enabled.
func Logf(format string, args ...any) {
	if !RedrawEnabled() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	appendDebugLog(msg)
}

// LogRedraw records why a full render was triggered.
func LogRedraw(reason string, prevLines, newLines, height int) {
	if !RedrawEnabled() {
		return
	}
	msg := fmt.Sprintf("fullRender: %s (prev=%d, new=%d, height=%d)", reason, prevLines, newLines, height)
	appendDebugLog(msg)
}

func appendDebugLog(msg string) {
	path := DebugLogPath()
	line := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), msg)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}

// RenderDump holds state captured for a differential render snapshot.
type RenderDump struct {
	FirstChanged      int
	ViewportTop       int
	CursorRow         int
	Height            int
	HardwareCursorRow int
	RenderEnd         int
	FinalCursorRow    int
	CursorPosRow      int
	CursorPosCol      int
	NewLines          []string
	PreviousLines     []string
	Buffer            string
}

// WriteRenderDump writes a detailed differential-render snapshot under
// /tmp/tui/render-<ms>.log (same convention as pi-tui). Callers should gate
// on RenderDumpEnabled when attaching to hot paths.
func WriteRenderDump(d RenderDump) {
	_ = os.MkdirAll(renderDumpDir, 0o755)

	ts := time.Now().UnixNano() / int64(time.Millisecond)
	path := filepath.Join(renderDumpDir, "render-"+strconv.FormatInt(ts, 10)+".log")

	var b strings.Builder
	b.WriteString("firstChanged: ")
	b.WriteString(strconv.Itoa(d.FirstChanged))
	b.WriteString("\nviewportTop: ")
	b.WriteString(strconv.Itoa(d.ViewportTop))
	b.WriteString("\ncursorRow: ")
	b.WriteString(strconv.Itoa(d.CursorRow))
	b.WriteString("\nheight: ")
	b.WriteString(strconv.Itoa(d.Height))
	b.WriteString("\nhardwareCursorRow: ")
	b.WriteString(strconv.Itoa(d.HardwareCursorRow))
	b.WriteString("\nrenderEnd: ")
	b.WriteString(strconv.Itoa(d.RenderEnd))
	b.WriteString("\nfinalCursorRow: ")
	b.WriteString(strconv.Itoa(d.FinalCursorRow))
	b.WriteString("\ncursorPos: row=")
	b.WriteString(strconv.Itoa(d.CursorPosRow))
	b.WriteString(" col=")
	b.WriteString(strconv.Itoa(d.CursorPosCol))
	b.WriteString("\nnewLines.length: ")
	b.WriteString(strconv.Itoa(len(d.NewLines)))
	b.WriteString("\npreviousLines.length: ")
	b.WriteString(strconv.Itoa(len(d.PreviousLines)))
	b.WriteString("\n\n=== newLines ===\n")
	for i, line := range d.NewLines {
		b.WriteString("[")
		b.WriteString(strconv.Itoa(i))
		b.WriteString("] ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n=== previousLines ===\n")
	for i, line := range d.PreviousLines {
		b.WriteString("[")
		b.WriteString(strconv.Itoa(i))
		b.WriteString("] ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	if d.Buffer != "" {
		b.WriteString("\n=== buffer ===\n")
		b.WriteString(strconv.Quote(d.Buffer))
		b.WriteString("\n")
	}

	_ = os.WriteFile(path, []byte(b.String()), 0o644)
}

// WriteCrashLog overwrites the crash log with the given content.
func WriteCrashLog(data string) {
	path := CrashLogPath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(data), 0o644)
}

// ResolveWriteLogPath returns the file path for ANSI write capture, or "" if disabled.
// FASTTUI_WRITE_LOG may be an existing directory (then a timestamped file is created)
// or a concrete file path.
func ResolveWriteLogPath() string {
	writeLogPathOnce.Do(func() {
		writeLogPath = resolveWriteLogPath(os.Getenv(envWriteLog))
	})
	return writeLogPath
}

// ResetWriteLogPathForTest clears the cached write-log path (tests only).
func ResetWriteLogPathForTest() {
	writeLogPathOnce = sync.Once{}
	writeLogPath = ""
}

func resolveWriteLogPath(env string) string {
	if env == "" {
		return ""
	}
	if info, err := os.Stat(env); err == nil && info.IsDir() {
		now := time.Now()
		ts := fmt.Sprintf("%04d-%02d-%02d_%02d-%02d-%02d",
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second())
		return filepath.Join(env, fmt.Sprintf("tui-%s-%d.log", ts, os.Getpid()))
	}
	return env
}

// AppendWrite appends raw terminal output to the write-log file when configured.
// Failures are ignored so logging never breaks the TUI.
func AppendWrite(data string) {
	path := ResolveWriteLogPath()
	if path == "" || data == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(data)
	_ = f.Close()
}
