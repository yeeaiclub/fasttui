package components

import (
	"strings"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/keys"
)

type Input struct {
	value        string
	cursor       int
	onSubmit     func(string)
	onEscape     func()
	focused      bool
	pastedBuffer string
	isInPaste    bool
}

func NewInput() *Input {
	return &Input{
		value:        "",
		cursor:       0,
		onSubmit:     nil,
		onEscape:     nil,
		focused:      false,
		pastedBuffer: "",
		isInPaste:    false,
	}
}

func (i *Input) Render(width int) []string {
	// Calculate visible window
	prompt := "> "
	availableWidth := width - len(prompt)

	if availableWidth <= 0 {
		return []string{prompt}
	}

	visibleText := ""
	cursorDisplay := i.cursor

	if len(i.value) < availableWidth {
		// Everything fits (leave room for cursor at end)
		visibleText = i.value
	} else {
		// Need horizontal scrolling
		// Reserve one character for cursor if it's at the end
		scrollWidth := availableWidth
		if i.cursor == len(i.value) {
			scrollWidth = availableWidth - 1
		}
		halfWidth := scrollWidth / 2

		if i.cursor < halfWidth {
			// Cursor near start
			visibleText = i.value[:scrollWidth]
			cursorDisplay = i.cursor
		} else if i.cursor > len(i.value)-halfWidth {
			// Cursor near end
			visibleText = i.value[len(i.value)-scrollWidth:]
			cursorDisplay = scrollWidth - (len(i.value) - i.cursor)
		} else {
			// Cursor in middle
			start := i.cursor - halfWidth
			visibleText = i.value[start : start+scrollWidth]
			cursorDisplay = halfWidth
		}
	}

	// Build line with fake cursor
	// Insert cursor character at cursor position
	beforeCursor := ""
	if cursorDisplay > 0 {
		beforeCursor = visibleText[:cursorDisplay]
	}

	atCursor := " " // Character at cursor, or space if at end
	if cursorDisplay < len(visibleText) {
		atCursor = string(visibleText[cursorDisplay])
	}

	afterCursor := ""
	if cursorDisplay+1 < len(visibleText) {
		afterCursor = visibleText[cursorDisplay+1:]
	}

	// Use inverse video to show cursor
	cursorChar := "\x1b[7m" + atCursor + "\x1b[27m" // ESC[7m = reverse video, ESC[27m = normal
	textWithCursor := beforeCursor + cursorChar + afterCursor

	// Calculate visual width
	visualLength := fasttui.VisibleWidth(textWithCursor)
	paddingCount := availableWidth - visualLength
	if paddingCount < 0 {
		paddingCount = 0
	}
	padding := strings.Repeat(" ", paddingCount)
	line := prompt + textWithCursor + padding

	return []string{line}
}

func (i *Input) HandleInput(data string) {
	// Handle bracketed paste mode
	// Start of paste: \x1b[200~
	// End of paste: \x1b[201~

	// Check if we're starting a bracketed paste
	if strings.Contains(data, "\x1b[200~") {
		i.isInPaste = true
		i.pastedBuffer = ""
		data = strings.ReplaceAll(data, "\x1b[200~", "")
	}

	// If we're in a paste, buffer the data
	if i.isInPaste {
		// Check if this chunk contains the end marker
		i.pastedBuffer += data

		endIndex := strings.Index(i.pastedBuffer, "\x1b[201~")
		if endIndex != -1 {
			// Extract the pasted content
			pasteContent := i.pastedBuffer[:endIndex]

			// Process the complete paste
			i.handlePaste(pasteContent)

			// Reset paste state
			i.isInPaste = false

			// Handle any remaining input after the paste marker
			remaining := i.pastedBuffer[endIndex+6:] // 6 = length of \x1b[201~
			i.pastedBuffer = ""
			if remaining != "" {
				i.HandleInput(remaining)
			}
		}
		return
	}

	kb := keys.GetEditorKeybindings()

	// Escape/Cancel
	if kb.Matches(data, keys.EditorActionSelectCancel) {
		if i.onEscape != nil {
			i.onEscape()
		}
		return
	}

	// Submit
	if kb.Matches(data, keys.EditorActionSubmit) || data == "\n" {
		if i.onSubmit != nil {
			i.onSubmit(i.value)
		}
		return
	}

	// Deletion
	if kb.Matches(data, keys.EditorActionDeleteCharBackward) {
		if i.cursor > 0 {
			beforeCursor := i.value[:i.cursor]
			graphemeLength := 1
			if len(beforeCursor) > 0 {
				// Simple implementation: delete one rune
				runes := []rune(beforeCursor)
				if len(runes) > 0 {
					graphemeLength = len(string(runes[len(runes)-1]))
				}
			}
			i.value = i.value[:i.cursor-graphemeLength] + i.value[i.cursor:]
			i.cursor -= graphemeLength
		}
		return
	}

	if kb.Matches(data, keys.EditorActionDeleteCharForward) {
		if i.cursor < len(i.value) {
			afterCursor := i.value[i.cursor:]
			graphemeLength := 1
			if len(afterCursor) > 0 {
				// Simple implementation: delete one rune
				runes := []rune(afterCursor)
				if len(runes) > 0 {
					graphemeLength = len(string(runes[0]))
				}
			}
			i.value = i.value[:i.cursor] + i.value[i.cursor+graphemeLength:]
		}
		return
	}

	if kb.Matches(data, keys.EditorActionDeleteWordBackward) {
		i.deleteWordBackwards()
		return
	}

	if kb.Matches(data, keys.EditorActionDeleteToLineStart) {
		i.value = i.value[i.cursor:]
		i.cursor = 0
		return
	}

	if kb.Matches(data, keys.EditorActionDeleteToLineEnd) {
		i.value = i.value[:i.cursor]
		return
	}

	// Cursor movement
	if kb.Matches(data, keys.EditorActionCursorLeft) {
		if i.cursor > 0 {
			beforeCursor := i.value[:i.cursor]
			graphemeLength := 1
			if len(beforeCursor) > 0 {
				runes := []rune(beforeCursor)
				if len(runes) > 0 {
					graphemeLength = len(string(runes[len(runes)-1]))
				}
			}
			i.cursor -= graphemeLength
		}
		return
	}

	if kb.Matches(data, keys.EditorActionCursorRight) {
		if i.cursor < len(i.value) {
			afterCursor := i.value[i.cursor:]
			graphemeLength := 1
			if len(afterCursor) > 0 {
				runes := []rune(afterCursor)
				if len(runes) > 0 {
					graphemeLength = len(string(runes[0]))
				}
			}
			i.cursor += graphemeLength
		}
		return
	}

	if kb.Matches(data, keys.EditorActionCursorLineStart) {
		i.cursor = 0
		return
	}

	if kb.Matches(data, keys.EditorActionCursorLineEnd) {
		i.cursor = len(i.value)
		return
	}

	if kb.Matches(data, keys.EditorActionCursorWordLeft) {
		i.moveWordBackwards()
		return
	}

	if kb.Matches(data, keys.EditorActionCursorWordRight) {
		i.moveWordForwards()
		return
	}

	// Regular character input - accept printable characters including Unicode,
	// but reject control characters (C0: 0x00-0x1F, DEL: 0x7F, C1: 0x80-0x9F)
	hasControlChars := false
	for _, ch := range data {
		code := int(ch)
		if code < 32 || code == 0x7f || (code >= 0x80 && code <= 0x9f) {
			hasControlChars = true
			break
		}
	}
	if !hasControlChars {
		i.value = i.value[:i.cursor] + data + i.value[i.cursor:]
		i.cursor += len(data)
	}
}

func (i *Input) WantsKeyRelease() bool {
	return false
}

func (i *Input) Invalidate() {
}

func (i *Input) GetValue() string {
	return i.value
}

func (i *Input) SetFocused(focused bool) {
	i.focused = focused
}

func (i *Input) IsFocused() bool {
	return i.focused
}

func (i *Input) SetOnSubmit(onSubmit func(string)) {
	i.onSubmit = onSubmit
}

func (i *Input) SetOnEscape(onEscape func()) {
	i.onEscape = onEscape
}

func (i *Input) handlePaste(content string) {
	// Clean the pasted text - remove newlines and carriage returns
	cleanText := strings.ReplaceAll(content, "\r\n", "")
	cleanText = strings.ReplaceAll(cleanText, "\r", "")
	cleanText = strings.ReplaceAll(cleanText, "\n", "")

	// Insert at cursor position
	i.value = i.value[:i.cursor] + cleanText + i.value[i.cursor:]
	i.cursor += len(cleanText)
}

func (i *Input) deleteWordBackwards() {
	if i.cursor == 0 {
		return
	}

	oldCursor := i.cursor
	i.moveWordBackwards()
	deleteFrom := i.cursor
	i.cursor = oldCursor

	i.value = i.value[:deleteFrom] + i.value[i.cursor:]
	i.cursor = deleteFrom
}

func (i *Input) moveWordBackwards() {
	if i.cursor == 0 {
		return
	}

	textBeforeCursor := i.value[:i.cursor]
	runes := []rune(textBeforeCursor)

	// Skip trailing whitespace
	for len(runes) > 0 && fasttui.IsWhitespaceChar(string(runes[len(runes)-1])) {
		i.cursor -= len(string(runes[len(runes)-1]))
		runes = runes[:len(runes)-1]
	}

	if len(runes) > 0 {
		lastRune := string(runes[len(runes)-1])
		if fasttui.IsPunctuationChar(lastRune) {
			// Skip punctuation run
			for len(runes) > 0 && fasttui.IsPunctuationChar(string(runes[len(runes)-1])) {
				i.cursor -= len(string(runes[len(runes)-1]))
				runes = runes[:len(runes)-1]
			}
		} else {
			// Skip word run
			for len(runes) > 0 &&
				!fasttui.IsWhitespaceChar(string(runes[len(runes)-1])) &&
				!fasttui.IsPunctuationChar(string(runes[len(runes)-1])) {
				i.cursor -= len(string(runes[len(runes)-1]))
				runes = runes[:len(runes)-1]
			}
		}
	}
}

func (i *Input) moveWordForwards() {
	if i.cursor >= len(i.value) {
		return
	}

	textAfterCursor := i.value[i.cursor:]
	runes := []rune(textAfterCursor)
	idx := 0

	// Skip leading whitespace
	for idx < len(runes) && fasttui.IsWhitespaceChar(string(runes[idx])) {
		i.cursor += len(string(runes[idx]))
		idx++
	}

	if idx < len(runes) {
		firstRune := string(runes[idx])
		if fasttui.IsPunctuationChar(firstRune) {
			// Skip punctuation run
			for idx < len(runes) && fasttui.IsPunctuationChar(string(runes[idx])) {
				i.cursor += len(string(runes[idx]))
				idx++
			}
		} else {
			// Skip word run
			for idx < len(runes) &&
				!fasttui.IsWhitespaceChar(string(runes[idx])) &&
				!fasttui.IsPunctuationChar(string(runes[idx])) {
				i.cursor += len(string(runes[idx]))
				idx++
			}
		}
	}
}
