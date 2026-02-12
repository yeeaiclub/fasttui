package components

import (
	"strings"

	"github.com/yeeaiclub/fasttui"
)

// Text component - displays multi-line text with word wrapping
type Text struct {
	text       string
	paddingX   int // Left/right padding
	paddingY   int // Top/bottom padding
	customBgFn func(string) string

	// Cache for rendered output
	cachedText  string
	cachedWidth int
	cachedLines []string
	cacheValid  bool
}

func NewText(text string, paddingX int, paddingY int, customBgFn func(string) string) *Text {
	return &Text{
		text:       text,
		paddingX:   paddingX,
		paddingY:   paddingY,
		customBgFn: customBgFn,
	}
}

func (t *Text) SetText(text string) {
	t.text = text
	t.cacheValid = false
}

func (t *Text) SetCustomBgFn(customBgFn func(string) string) {
	t.customBgFn = customBgFn
	t.cacheValid = false
}

func (t *Text) Invalidate() {
	t.cacheValid = false
}

func (t *Text) Render(width int) []string {
	// Check cache
	if t.cacheValid && t.cachedText == t.text && t.cachedWidth == width {
		return t.cachedLines
	}

	// Don't render anything if there's no actual text
	if t.text == "" || strings.TrimSpace(t.text) == "" {
		result := []string{}
		t.cachedText = t.text
		t.cachedWidth = width
		t.cachedLines = result
		t.cacheValid = true
		return result
	}

	// Replace tabs with 3 spaces
	normalizedText := strings.ReplaceAll(t.text, "\t", "   ")

	// Calculate content width (subtract left/right margins)
	contentWidth := max(1, width-t.paddingX*2)

	// Wrap text (this preserves ANSI codes but does NOT pad)
	wrappedLines := fasttui.WrapTextWithAnsi(normalizedText, contentWidth)

	// Add margins and background to each line
	leftMargin := strings.Repeat(" ", t.paddingX)
	rightMargin := strings.Repeat(" ", t.paddingX)
	contentLines := make([]string, 0, len(wrappedLines))

	for _, line := range wrappedLines {
		// Add margins
		lineWithMargins := leftMargin + line + rightMargin

		// Apply background if specified (this also pads to full width)
		if t.customBgFn != nil {
			contentLines = append(contentLines, fasttui.ApplyBackgroundToLine(lineWithMargins, width, t.customBgFn))
		} else {
			// No background - just pad to width with spaces
			visibleLen := fasttui.VisibleWidth(lineWithMargins)
			paddingNeeded := max(0, width-visibleLen)
			contentLines = append(contentLines, lineWithMargins+strings.Repeat(" ", paddingNeeded))
		}
	}

	// Add top/bottom padding (empty lines)
	emptyLine := strings.Repeat(" ", width)
	emptyLines := make([]string, 0, t.paddingY)
	for i := 0; i < t.paddingY; i++ {
		var line string
		if t.customBgFn != nil {
			line = fasttui.ApplyBackgroundToLine(emptyLine, width, t.customBgFn)
		} else {
			line = emptyLine
		}
		emptyLines = append(emptyLines, line)
	}

	result := make([]string, 0, len(emptyLines)*2+len(contentLines))
	result = append(result, emptyLines...)
	result = append(result, contentLines...)
	result = append(result, emptyLines...)

	// Update cache
	t.cachedText = t.text
	t.cachedWidth = width
	t.cachedLines = result
	t.cacheValid = true

	if len(result) > 0 {
		return result
	}
	return []string{""}
}
