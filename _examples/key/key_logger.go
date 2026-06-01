package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/keys"
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
		os.Exit(0)
	}

	hex := k.toHex(data)
	charCodes := k.toCharCodes(data)
	repr := k.toRepr(data)

	logLine := fmt.Sprintf("Hex: %s | Chars: [%s] | Repr: %q", hex, charCodes, repr)

	k.log = append(k.log, logLine)

	if len(k.log) > k.maxLines {
		k.log = k.log[1:]
	}

	k.tui.ForceRender()
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
	// Convert string to bytes to get correct hex representation
	bytes := []byte(data)
	result := ""
	for _, b := range bytes {
		result += fmt.Sprintf("%02x", b)
	}
	return result
}

func (k *KeyLogger) toCharCodes(data string) string {
	// Convert string to bytes to get correct char codes
	bytes := []byte(data)
	codes := make([]string, len(bytes))
	for i, b := range bytes {
		codes[i] = fmt.Sprintf("%d", b)
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
	if width <= 0 {
		return ""
	}

	// Use VisibleWidth to handle ANSI codes and wide characters correctly
	visibleLen := fasttui.VisibleWidth(s)

	if visibleLen > width {
		// Truncate to fit
		truncated := fasttui.SliceByColumn(s, 0, width, true)
		// Double-check the result doesn't exceed width
		if fasttui.VisibleWidth(truncated) > width {
			// Fallback: use simple string slicing if SliceByColumn fails
			// This shouldn't happen, but it's a safety net
			if len(s) > width {
				return s[:width]
			}
			return s
		}
		return truncated
	} else if visibleLen < width {
		// Pad to width
		return s + strings.Repeat(" ", width-visibleLen)
	}
	// Exactly the right width
	return s
}

