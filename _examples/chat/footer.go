package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/yeeaiclub/fasttui"
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

	// Calculate padding to right-align model using visible width
	minPadding := 2
	statsLeftWidth := fasttui.VisibleWidth(statsLeft)
	modelStrWidth := fasttui.VisibleWidth(modelStr)
	availableSpace := width - statsLeftWidth - modelStrWidth - minPadding

	var statsLine string
	if availableSpace >= 0 {
		padding := strings.Repeat(" ", availableSpace+minPadding)
		statsLine = statsLeft + padding + modelStr
	} else {
		// Not enough space, truncate model name
		maxModelLen := width - statsLeftWidth - minPadding
		if maxModelLen > 3 {
			// Use fasttui.TruncateToWidth for proper Unicode handling
			modelStr = fasttui.TruncateToWidth(modelStr, maxModelLen, "...", false)
			padding := strings.Repeat(" ", minPadding)
			statsLine = statsLeft + padding + modelStr
		} else {
			statsLine = statsLeft
		}
	}

	// CRITICAL: Always truncate to exact width to prevent overflow
	statsLineWidth := fasttui.VisibleWidth(statsLine)
	if statsLineWidth > width {
		statsLine = fasttui.SliceByColumn(statsLine, 0, width, true)
	} else if statsLineWidth < width {
		// Pad to exact width if needed
		statsLine = statsLine + strings.Repeat(" ", width-statsLineWidth)
	}

	// Truncate pwd if needed using fasttui functions
	pwdWidth := fasttui.VisibleWidth(pwd)
	if pwdWidth > width {
		// Use TruncateToWidth with middle ellipsis pattern
		half := width/2 - 2
		if half > 0 {
			leftPart := fasttui.SliceByColumn(pwd, 0, half, true)
			// Get the last 'half' columns from the end
			rightStart := pwdWidth - half
			if rightStart > 0 {
				rightPart := fasttui.SliceByColumn(pwd, rightStart, half, true)
				pwd = leftPart + "..." + rightPart
			} else {
				pwd = fasttui.TruncateToWidth(pwd, width, "...", false)
			}
		} else {
			pwd = fasttui.TruncateToWidth(pwd, width, "...", false)
		}
	}

	// CRITICAL: Ensure pwd also doesn't exceed width
	pwdFinalWidth := fasttui.VisibleWidth(pwd)
	if pwdFinalWidth > width {
		pwd = fasttui.SliceByColumn(pwd, 0, width, true)
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
