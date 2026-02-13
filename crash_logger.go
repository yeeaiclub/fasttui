package fasttui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (t *TUIRefactored) getCrashLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".panda", "panda-crash.log")
}

func (t *TUIRefactored) writeCrashLog(path string, data string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(path, []byte(data), 0644)
}

func formatCrashLog(line string, lineIndex int, allLines []string, width int) string {
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
	for idx, l := range allLines {
		crashData.WriteString("[")
		crashData.WriteString(strconv.Itoa(idx))
		crashData.WriteString("] (w=")
		crashData.WriteString(strconv.Itoa(VisibleWidth(l)))
		crashData.WriteString(") ")
		crashData.WriteString(l)
		crashData.WriteString("\n")
	}
	return crashData.String()
}

func formatWidthError(lineIndex, lineWidth, termWidth int, crashLogPath string) string {
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
