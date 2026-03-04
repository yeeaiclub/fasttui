package fasttui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Logging Functions

// GetCrashLogPath returns the path to the crash log file.
func GetCrashLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".panda", "panda-crash.log")
}

// WriteCrashLog writes crash data to the specified path.
func WriteCrashLog(path string, data string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(path, []byte(data), 0644)
}

// LogCrashInfo logs detailed crash information including terminal width,
// line index, and all rendered lines.
func LogCrashInfo(width int, lineIndex int, line string, newLines []string) {
	crashLogPath := GetCrashLogPath()
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
	WriteCrashLog(crashLogPath, crashData.String())
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
func WriteDebugLog(firstChanged, viewportTop, finalCursorRow, hardwareCursorRow,
	renderEnd, cursorRow, cursorCol, height int, tCursorRow int, newLines, previousLines []string) {
	debugDir := "/tmp/tui"
	os.MkdirAll(debugDir, 0755)

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	debugPath := filepath.Join(debugDir, "render-"+strconv.FormatInt(timestamp, 10)+".log")

	var debugData strings.Builder
	debugData.WriteString("firstChanged: ")
	debugData.WriteString(strconv.Itoa(firstChanged))
	debugData.WriteString("\nviewportTop: ")
	debugData.WriteString(strconv.Itoa(viewportTop))
	debugData.WriteString("\ncursorRow: ")
	debugData.WriteString(strconv.Itoa(tCursorRow))
	debugData.WriteString("\nheight: ")
	debugData.WriteString(strconv.Itoa(height))
	debugData.WriteString("\nhardwareCursorRow: ")
	debugData.WriteString(strconv.Itoa(hardwareCursorRow))
	debugData.WriteString("\nrenderEnd: ")
	debugData.WriteString(strconv.Itoa(renderEnd))
	debugData.WriteString("\nfinalCursorRow: ")
	debugData.WriteString(strconv.Itoa(finalCursorRow))
	debugData.WriteString("\ncursorPos: row=")
	debugData.WriteString(strconv.Itoa(cursorRow))
	debugData.WriteString(" col=")
	debugData.WriteString(strconv.Itoa(cursorCol))
	debugData.WriteString("\nnewLines.length: ")
	debugData.WriteString(strconv.Itoa(len(newLines)))
	debugData.WriteString("\npreviousLines.length: ")
	debugData.WriteString(strconv.Itoa(len(previousLines)))
	debugData.WriteString("\n\n=== newLines ===\n")
	for i, line := range newLines {
		debugData.WriteString("[")
		debugData.WriteString(strconv.Itoa(i))
		debugData.WriteString("] ")
		debugData.WriteString(line)
		debugData.WriteString("\n")
	}
	debugData.WriteString("\n=== previousLines ===\n")
	for i, line := range previousLines {
		debugData.WriteString("[")
		debugData.WriteString(strconv.Itoa(i))
		debugData.WriteString("] ")
		debugData.WriteString(line)
		debugData.WriteString("\n")
	}

	os.WriteFile(debugPath, []byte(debugData.String()), 0644)
}
