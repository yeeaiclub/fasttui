package components

import (
	"strings"
)

// DefaultTextStyle defines the default styling for markdown content
type DefaultTextStyle struct {
	Color         func(string) string
	BgColor       func(string) string
	Bold          bool
	Italic        bool
	Strikethrough bool
	Underline     bool
}

// MarkdownTheme defines theme functions for markdown elements
type MarkdownTheme struct {
	Heading         func(string) string
	Link            func(string) string
	LinkURL         func(string) string
	Code            func(string) string
	CodeBlock       func(string) string
	CodeBlockBorder func(string) string
	Quote           func(string) string
	QuoteBorder     func(string) string
	HR              func(string) string
	ListBullet      func(string) string
	Bold            func(string) string
	Italic          func(string) string
	Strikethrough   func(string) string
	Underline       func(string) string
	HighlightCode   func(code string, lang string) []string
	CodeBlockIndent string
}

type Markdown struct {
	text             string
	paddingX         int
	paddingY         int
	defaultTextStyle *DefaultTextStyle
	theme            *MarkdownTheme

	cachedText         *string
	cachedWidth        *int
	cachedLines        []string
	defaultStylePrefix *string
}

func NewMarkdown(text string, paddingX, paddingY int, theme *MarkdownTheme, defaultTextStyle *DefaultTextStyle) *Markdown {
	return &Markdown{
		text:             text,
		paddingX:         paddingX,
		paddingY:         paddingY,
		theme:            theme,
		defaultTextStyle: defaultTextStyle,
	}
}

func (m *Markdown) SetText(text string) {
	m.text = text
	m.Invalidate()
}

func (m *Markdown) Invalidate() {
	m.cachedText = nil
	m.cachedWidth = nil
	m.cachedLines = nil
}

func (m *Markdown) HandleInput(data string) {}

func (m *Markdown) WantsKeyRelease() bool {
	return false
}

func (m *Markdown) Render(width int) []string {
	// Check cache
	if m.cachedLines != nil && m.cachedText != nil && *m.cachedText == m.text &&
		m.cachedWidth != nil && *m.cachedWidth == width {
		return m.cachedLines
	}

	// Calculate available width for content
	contentWidth := max(1, width-m.paddingX*2)

	// Don't render anything if there's no actual text
	if strings.TrimSpace(m.text) == "" {
		result := []string{}
		m.cachedText = &m.text
		m.cachedWidth = &width
		m.cachedLines = result
		return result
	}

	// Replace tabs with 3 spaces
	normalizedText := strings.ReplaceAll(m.text, "\t", "   ")

	// Parse and render markdown
	renderedLines := m.renderMarkdown(normalizedText, contentWidth)

	// Add margins and background
	result := m.applyPaddingAndBackground(renderedLines, width)

	// Update cache
	m.cachedText = &m.text
	m.cachedWidth = &width
	m.cachedLines = result

	if len(result) == 0 {
		return []string{""}
	}
	return result
}

func (m *Markdown) renderMarkdown(text string, width int) []string {
	lines := strings.Split(text, "\n")
	result := []string{}
	inCodeBlock := false
	codeBlockLang := ""

	for i := range lines {
		line := lines[i]

		// Code blocks
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "```"); ok {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockLang = strings.TrimSpace(after)
				if m.theme != nil && m.theme.CodeBlockBorder != nil {
					result = append(result, m.theme.CodeBlockBorder("```"+codeBlockLang))
				} else {
					result = append(result, "```"+codeBlockLang)
				}
			} else {
				inCodeBlock = false
				if m.theme != nil && m.theme.CodeBlockBorder != nil {
					result = append(result, m.theme.CodeBlockBorder("```"))
				} else {
					result = append(result, "```")
				}
				result = append(result, "")
			}
			continue
		}

		if inCodeBlock {
			indent := "  "
			if m.theme != nil && m.theme.CodeBlockIndent != "" {
				indent = m.theme.CodeBlockIndent
			}
			styledLine := line
			if m.theme != nil && m.theme.CodeBlock != nil {
				styledLine = m.theme.CodeBlock(line)
			}
			result = append(result, indent+styledLine)
			continue
		}

		// Headings
		if strings.HasPrefix(line, "#") {
			level := 0
			for _, ch := range line {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			if level > 0 && level <= 6 && len(line) > level && line[level] == ' ' {
				headingText := strings.TrimSpace(line[level:])
				styledHeading := headingText
				if m.theme != nil {
					if level == 1 && m.theme.Heading != nil && m.theme.Bold != nil && m.theme.Underline != nil {
						styledHeading = m.theme.Heading(m.theme.Bold(m.theme.Underline(headingText)))
					} else if level == 2 && m.theme.Heading != nil && m.theme.Bold != nil {
						styledHeading = m.theme.Heading(m.theme.Bold(headingText))
					} else if m.theme.Heading != nil && m.theme.Bold != nil {
						prefix := strings.Repeat("#", level) + " "
						styledHeading = m.theme.Heading(m.theme.Bold(prefix + headingText))
					}
				}
				result = append(result, styledHeading)
				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
					result = append(result, "")
				}
				continue
			}
		}

		// Horizontal rule
		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "***" || strings.TrimSpace(line) == "___" {
			hrLine := strings.Repeat("─", min(width, 80))
			if m.theme != nil && m.theme.HR != nil {
				hrLine = m.theme.HR(hrLine)
			}
			result = append(result, hrLine)
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
				result = append(result, "")
			}
			continue
		}

		// Blockquote
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			quoteText := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ">"))
			quoteLine := quoteText
			if m.theme != nil {
				border := "│ "
				if m.theme.QuoteBorder != nil {
					border = m.theme.QuoteBorder(border)
				}
				if m.theme.Quote != nil && m.theme.Italic != nil {
					quoteLine = border + m.theme.Quote(m.theme.Italic(quoteText))
				} else {
					quoteLine = border + quoteText
				}
			}
			result = append(result, quoteLine)
			continue
		}

		// Lists
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			listText := strings.TrimSpace(trimmed[2:])
			bullet := "- "
			if m.theme != nil && m.theme.ListBullet != nil {
				bullet = m.theme.ListBullet(bullet)
			}
			result = append(result, bullet+m.renderInline(listText))
			continue
		}

		// Numbered lists
		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' {
			dotIdx := strings.Index(trimmed, ". ")
			if dotIdx > 0 && dotIdx < 4 {
				listText := strings.TrimSpace(trimmed[dotIdx+2:])
				bullet := trimmed[:dotIdx+2]
				if m.theme != nil && m.theme.ListBullet != nil {
					bullet = m.theme.ListBullet(bullet)
				}
				result = append(result, bullet+m.renderInline(listText))
				continue
			}
		}

		// Empty lines
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
			continue
		}

		// Regular paragraph
		result = append(result, m.renderInline(line))
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if nextLine != "" && !strings.HasPrefix(nextLine, "-") && !strings.HasPrefix(nextLine, "*") {
				result = append(result, "")
			}
		}
	}

	return result
}

func (m *Markdown) renderInline(text string) string {
	result := text

	// Bold **text** or __text__
	result = m.replaceInlineStyle(result, "**", m.theme.Bold)
	result = m.replaceInlineStyle(result, "__", m.theme.Bold)

	// Italic *text* or _text_
	result = m.replaceInlineStyle(result, "*", m.theme.Italic)
	result = m.replaceInlineStyle(result, "_", m.theme.Italic)

	// Strikethrough ~~text~~
	result = m.replaceInlineStyle(result, "~~", m.theme.Strikethrough)

	// Inline code `code`
	result = m.replaceInlineCode(result)

	// Links [text](url)
	result = m.replaceLinks(result)

	// Apply default style
	return m.applyDefaultStyle(result)
}

func (m *Markdown) replaceInlineStyle(text string, marker string, styleFn func(string) string) string {
	if styleFn == nil {
		return text
	}

	result := ""
	remaining := text

	for {
		start := strings.Index(remaining, marker)
		if start == -1 {
			result += remaining
			break
		}

		end := strings.Index(remaining[start+len(marker):], marker)
		if end == -1 {
			result += remaining
			break
		}

		end += start + len(marker)
		result += remaining[:start]
		content := remaining[start+len(marker) : end]
		result += styleFn(content)
		remaining = remaining[end+len(marker):]
	}

	return result
}

func (m *Markdown) replaceInlineCode(text string) string {
	if m.theme == nil || m.theme.Code == nil {
		return text
	}

	result := ""
	remaining := text

	for {
		start := strings.Index(remaining, "`")
		if start == -1 {
			result += remaining
			break
		}

		end := strings.Index(remaining[start+1:], "`")
		if end == -1 {
			result += remaining
			break
		}

		end += start + 1
		result += remaining[:start]
		content := remaining[start+1 : end]
		result += m.theme.Code(content)
		remaining = remaining[end+1:]
	}

	return result
}

func (m *Markdown) replaceLinks(text string) string {
	if m.theme == nil || m.theme.Link == nil {
		return text
	}

	result := ""
	remaining := text

	for {
		start := strings.Index(remaining, "[")
		if start == -1 {
			result += remaining
			break
		}

		textEnd := strings.Index(remaining[start:], "]")
		if textEnd == -1 {
			result += remaining
			break
		}
		textEnd += start

		if textEnd+1 >= len(remaining) || remaining[textEnd+1] != '(' {
			result += remaining[:textEnd+1]
			remaining = remaining[textEnd+1:]
			continue
		}

		urlEnd := strings.Index(remaining[textEnd+2:], ")")
		if urlEnd == -1 {
			result += remaining
			break
		}
		urlEnd += textEnd + 2

		result += remaining[:start]
		linkText := remaining[start+1 : textEnd]
		linkURL := remaining[textEnd+2 : urlEnd]

		// Check if link text matches URL
		if linkText == linkURL || (strings.HasPrefix(linkURL, "mailto:") && linkText == linkURL[7:]) {
			if m.theme.Underline != nil {
				result += m.theme.Link(m.theme.Underline(linkText))
			} else {
				result += m.theme.Link(linkText)
			}
		} else {
			if m.theme.Underline != nil {
				result += m.theme.Link(m.theme.Underline(linkText))
			} else {
				result += m.theme.Link(linkText)
			}
			if m.theme.LinkURL != nil {
				result += m.theme.LinkURL(" (" + linkURL + ")")
			} else {
				result += " (" + linkURL + ")"
			}
		}

		remaining = remaining[urlEnd+1:]
	}

	return result
}

func (m *Markdown) applyDefaultStyle(text string) string {
	if m.defaultTextStyle == nil {
		return text
	}

	styled := text

	if m.defaultTextStyle.Color != nil {
		styled = m.defaultTextStyle.Color(styled)
	}

	if m.defaultTextStyle.Bold && m.theme != nil {
		styled = m.theme.Bold(styled)
	}
	if m.defaultTextStyle.Italic && m.theme != nil {
		styled = m.theme.Italic(styled)
	}
	if m.defaultTextStyle.Strikethrough && m.theme != nil {
		styled = m.theme.Strikethrough(styled)
	}
	if m.defaultTextStyle.Underline && m.theme != nil {
		styled = m.theme.Underline(styled)
	}

	return styled
}

func (m *Markdown) applyPaddingAndBackground(lines []string, width int) []string {
	leftMargin := strings.Repeat(" ", m.paddingX)
	rightMargin := strings.Repeat(" ", m.paddingX)

	var bgFn func(string) string
	if m.defaultTextStyle != nil {
		bgFn = m.defaultTextStyle.BgColor
	}

	contentLines := []string{}

	for _, line := range lines {
		lineWithMargins := leftMargin + line + rightMargin
		contentLines = append(contentLines, lineWithMargins)
	}

	// Add top/bottom padding
	emptyLine := strings.Repeat(" ", width)
	emptyLines := []string{}
	for i := 0; i < m.paddingY; i++ {
		line := emptyLine
		if bgFn != nil {
			line = bgFn(line)
		}
		emptyLines = append(emptyLines, line)
	}

	result := append(emptyLines, contentLines...)
	result = append(result, emptyLines...)

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
