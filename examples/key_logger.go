package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/keys"
	"github.com/yeeaiclub/fasttui/terminal"
)

type KeyLogger struct {
	log      []string
	maxLines int
	tui      *fasttui.TUI
}

func NewKeyLogger(tui *fasttui.TUI) *KeyLogger {
	return &KeyLogger{
		log:      make([]string, 0),
		maxLines: 20,
		tui:      tui,
	}
}

func (k *KeyLogger) HandleInput(data string) {
	if keys.MatchesKey(data, "ctrl+c") {
		k.tui.Stop()
		fmt.Println("\nExiting...")
		return
	}

	hex := k.toHex(data)
	charCodes := k.toCharCodes(data)
	repr := k.toRepr(data)

	logLine := fmt.Sprintf("Hex: %-20s | Chars: [%-15s] | Repr: \"%s\"", hex, charCodes, repr)

	k.log = append(k.log, logLine)

	if len(k.log) > k.maxLines {
		k.log = k.log[1:]
	}
}

func (k *KeyLogger) WantsKeyRelease() bool {
	return false
}

func (k *KeyLogger) Invalidate() {}

func (k *KeyLogger) Render(width int) []string {
	lines := make([]string, 0)

	lines = append(lines, strings.Repeat("=", width))
	title := "Key Code Tester - Press keys to see their codes (Ctrl+C to exit)"
	lines = append(lines, k.padRight(title, width))
	lines = append(lines, strings.Repeat("=", width))
	lines = append(lines, "")

	for _, entry := range k.log {
		lines = append(lines, k.padRight(entry, width))
	}

	remaining := max(0, 25-len(lines))
	for i := 0; i < remaining; i++ {
		lines = append(lines, strings.Repeat(" ", width))
	}

	lines = append(lines, strings.Repeat("=", width))
	lines = append(lines, k.padRight("Test these:", width))
	lines = append(lines, k.padRight("  - Shift + Enter (should show: \\x1b[13;2u with Kitty protocol)", width))
	lines = append(lines, k.padRight("  - Alt/Option + Enter", width))
	lines = append(lines, k.padRight("  - Option/Alt + Backspace", width))
	lines = append(lines, k.padRight("  - Cmd/Ctrl + Backspace", width))
	lines = append(lines, k.padRight("  - Regular Backspace", width))
	lines = append(lines, strings.Repeat("=", width))

	return lines
}

func (k *KeyLogger) toHex(data string) string {
	result := ""
	for _, c := range data {
		result += fmt.Sprintf("%02x", c)
	}
	return result
}

func (k *KeyLogger) toCharCodes(data string) string {
	codes := make([]string, len(data))
	for i, c := range data {
		codes[i] = fmt.Sprintf("%d", c)
	}
	return strings.Join(codes, ", ")
}

func (k *KeyLogger) toRepr(data string) string {
	result := data
	result = strings.ReplaceAll(result, "\x1b", "\\x1b")
	result = strings.ReplaceAll(result, "\r", "\\r")
	result = strings.ReplaceAll(result, "\n", "\\n")
	result = strings.ReplaceAll(result, "\t", "\\t")
	result = strings.ReplaceAll(result, "\x7f", "\\x7f")
	return result
}

func (k *KeyLogger) padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func main() {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)
	logger := NewKeyLogger(tui)

	tui.AddChild(logger)
	tui.SetFocus(logger)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		tui.Stop()
		fmt.Println("\nExiting...")
		os.Exit(0)
	}()

	tui.Start()

	done := make(chan struct{})
	<-done
}
