package fasttui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yeeaiclub/fasttui/internal/debug"
)

// GetCrashLogPath returns the path to the crash log file.
func GetCrashLogPath() string {
	return debug.CrashLogPath()
}

// WriteCrashLog writes crash data to the specified path.
func WriteCrashLog(path string, data string) {
	if path == "" || path == debug.CrashLogPath() {
		debug.WriteCrashLog(data)
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(data), 0o644)
}

// LogCrashInfo logs detailed crash information including terminal width,
// line index, and all rendered lines.
func LogCrashInfo(width int, lineIndex int, line string, newLines []string) {
	var crashData strings.Builder
	crashData.WriteString("Crash at ")
	crashData.WriteString(time.Now().Format(time.RFC3339))
	crashData.WriteString("\n")
	crashData.WriteString("Terminal width: ")
	crashData.WriteString(strconv.Itoa(width))
	crashData.WriteString("\n")
	crashData.WriteString("Line ")
	crashData.WriteString(strconv.Itoa(lineIndex))
	crashData.WriteString(" visible width: ")
	crashData.WriteString(strconv.Itoa(VisibleWidth(line)))
	crashData.WriteString("\n\n")
	crashData.WriteString("=== All rendered lines ===\n")
	for idx, l := range newLines {
		crashData.WriteString("[")
		crashData.WriteString(strconv.Itoa(idx))
		crashData.WriteString("] (w=")
		crashData.WriteString(strconv.Itoa(VisibleWidth(l)))
		crashData.WriteString(") ")
		crashData.WriteString(l)
		crashData.WriteString("\n")
	}
	debug.WriteCrashLog(crashData.String())
}

// BuildWidthExceedErrorMsg builds an error message for when a rendered line
// exceeds the terminal width.
func BuildWidthExceedErrorMsg(lineIndex int, lineWidth int, termWidth int, crashLogPath string) string {
	var errorMsg strings.Builder
	errorMsg.WriteString("Rendered line ")
	errorMsg.WriteString(strconv.Itoa(lineIndex))
	errorMsg.WriteString(" exceeds terminal width (")
	errorMsg.WriteString(strconv.Itoa(lineWidth))
	errorMsg.WriteString(" > ")
	errorMsg.WriteString(strconv.Itoa(termWidth))
	errorMsg.WriteString(").\n\n")
	errorMsg.WriteString("This is likely caused by a custom TUI component not truncating its output.\n")
	errorMsg.WriteString("Use VisibleWidth() to measure and truncate lines.\n\n")
	errorMsg.WriteString("Debug log written to: ")
	errorMsg.WriteString(crashLogPath)
	return errorMsg.String()
}

// WriteDebugLog writes detailed debug information about the rendering process.
// With FASTTUI_DEBUG=1 the render loop dumps automatically; this helper is for
// explicit dumps.
func WriteDebugLog(firstChanged, viewportTop, finalCursorRow, hardwareCursorRow,
	renderEnd, cursorRow, cursorCol, height int, tCursorRow int, newLines, previousLines []string) {
	debug.WriteRenderDump(debug.RenderDump{
		FirstChanged:      firstChanged,
		ViewportTop:       viewportTop,
		CursorRow:         tCursorRow,
		Height:            height,
		HardwareCursorRow: hardwareCursorRow,
		RenderEnd:         renderEnd,
		FinalCursorRow:    finalCursorRow,
		CursorPosRow:      cursorRow,
		CursorPosCol:      cursorCol,
		NewLines:          newLines,
		PreviousLines:     previousLines,
	})
}
