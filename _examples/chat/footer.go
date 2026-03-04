package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FooterComponent displays status information at the bottom of the chat
type FooterComponent struct {
	tokenCount int
	model      string
	contextPct float64
}

func NewFooterComponent(tokenCount int, model string) *FooterComponent {
	return &FooterComponent{
		tokenCount: tokenCount,
		model:      model,
		contextPct: 0,
	}
}

func (f *FooterComponent) SetTokenCount(count int) {
	f.tokenCount = count
}

func (f *FooterComponent) SetModel(model string) {
	f.model = model
}

func (f *FooterComponent) SetContextPercent(pct float64) {
	f.contextPct = pct
}

func (f *FooterComponent) Render(width int) []string {
	// Format token count (similar to fotter.ts)
	tokenStr := formatTokens(f.tokenCount)

	// Build left side stats
	statsLeft := "↑" + tokenStr

	// Build right side (model name)
	modelStr := f.model
	if modelStr == "" {
		modelStr = "no-model"
	}

	// Get current directory
	pwd, _ := os.Getwd()
	pwd = shortenPath(pwd)

	// Calculate padding to right-align model
	minPadding := 2
	availableSpace := width - len(statsLeft) - len(modelStr) - minPadding

	var statsLine string
	if availableSpace >= 0 {
		padding := strings.Repeat(" ", availableSpace+minPadding)
		statsLine = statsLeft + padding + modelStr
	} else {
		// Not enough space, truncate model name
		maxModelLen := width - len(statsLeft) - minPadding
		if maxModelLen > 3 {
			modelStr = modelStr[:maxModelLen-3] + "..."
			padding := strings.Repeat(" ", minPadding)
			statsLine = statsLeft + padding + modelStr
		} else {
			statsLine = statsLeft
		}
	}

	// Truncate if still too long
	if len(statsLine) > width {
		statsLine = statsLine[:width]
	}

	// Truncate pwd if needed
	if len(pwd) > width {
		half := width/2 - 2
		if half > 0 {
			pwd = pwd[:half] + "..." + pwd[len(pwd)-half:]
		} else {
			pwd = pwd[:width]
		}
	}

	return []string{pwd, statsLine}
}

func (f *FooterComponent) HandleInput(data string) {}

func (f *FooterComponent) WantsKeyRelease() bool {
	return false
}

func (f *FooterComponent) Invalidate() {}

// formatTokens formats token counts like "1.2k", "15k", "1.5M"
func formatTokens(count int) string {
	if count < 1000 {
		return strconv.Itoa(count)
	}
	if count < 10000 {
		return strconv.FormatFloat(float64(count)/1000, 'f', 1, 64) + "k"
	}
	if count < 1000000 {
		return strconv.Itoa(count/1000) + "k"
	}
	if count < 10000000 {
		return strconv.FormatFloat(float64(count)/1000000, 'f', 1, 64) + "M"
	}
	return strconv.Itoa(count/1000000) + "M"
}

// shortenPath replaces home directory with ~
func shortenPath(pwd string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(pwd, home) {
		return "~" + strings.TrimPrefix(pwd, home)
	}
	return pwd
}

// truncateToWidth truncates text to fit within width, adding ellipsis if truncated
func truncateToWidth(text string, width int, ellipsis string) string {
	if len(text) <= width {
		return text
	}
	if width <= len(ellipsis) {
		return text[:width]
	}
	return text[:width-len(ellipsis)] + ellipsis
}

// visibleWidth calculates visible width (ignoring ANSI codes)
func visibleWidth(s string) int {
	// Simple implementation - strip ANSI escape sequences
	result := ""
	inEscape := false
	for _, ch := range s {
		if inEscape {
			if ch == 'm' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
			continue
		}
		if ch == '\x1b' {
			inEscape = true
			continue
		}
		result += string(ch)
	}
	return len(result)
}

// filepath helper
func getDirName(path string) string {
	base := filepath.Base(path)
	if base == "." || base == "/" {
		return path
	}
	return base
}
