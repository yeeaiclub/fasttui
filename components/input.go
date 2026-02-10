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
	const prompt = "> "

	availableWidth := width - len(prompt)

	if availableWidth <= 0 {
		return []string{prompt}
	}

	var visibleText string
	cursorDisplay := i.cursor

	if len(i.value) < availableWidth {
		visibleText = i.value
	} else {
		scrollWidth := availableWidth
		if i.cursor == len(i.value) {
			scrollWidth = availableWidth - 1
		}
		scrollWidth = max(1, scrollWidth)
		halfWidth := scrollWidth / 2

		if i.cursor < halfWidth {
			visibleText = i.value[:scrollWidth]
			cursorDisplay = i.cursor
		} else if i.cursor > len(i.value)-halfWidth {
			visibleText = i.value[len(i.value)-scrollWidth:]
			cursorDisplay = scrollWidth - (len(i.value) - i.cursor)
		} else {
			start := i.cursor - halfWidth
			visibleText = i.value[start : start+scrollWidth]
			cursorDisplay = halfWidth
		}
	}

	if len(visibleText) == 0 {
		return []string{prompt}
	}

	cursorDisplay = max(0, min(cursorDisplay, len(visibleText)))

	beforeCursor := visibleText[:cursorDisplay]
	atCursor := " "
	if cursorDisplay < len(visibleText) {
		atCursor = string(visibleText[cursorDisplay])
	}
	afterCursor := ""
	if cursorDisplay+1 < len(visibleText) {
		afterCursor = visibleText[cursorDisplay+1:]
	}

	marker := ""
	if i.focused {
		marker = fasttui.CURSOR_MARKER
	}

	cursorChar := "\x1b[7m" + atCursor + "\x1b[27m"
	textWithCursor := beforeCursor + marker + cursorChar + afterCursor

	visualLength := fasttui.VisibleWidth(textWithCursor)
	padding := strings.Repeat(" ", max(0, availableWidth-visualLength))
	line := prompt + textWithCursor + padding

	return []string{line}
}
func (i *Input) HandleInput(data string) {
	if strings.Contains(data, "\x1b[200~") {
		i.isInPaste = true
		i.pastedBuffer = ""
		data = strings.ReplaceAll(data, "\x1b[200~", "")
	}

	if i.isInPaste {
		i.pastedBuffer += data

		endIndex := strings.Index(i.pastedBuffer, "\x1b[201~")
		if endIndex != -1 {
			pasteContent := i.pastedBuffer[:endIndex]

			i.handlePaste(pasteContent)

			i.isInPaste = false

			remaining := i.pastedBuffer[endIndex+6:]
			i.pastedBuffer = ""
			if remaining != "" {
				i.HandleInput(remaining)
			}
		}
		return
	}

	kb := keys.GetEditorKeybindings()
	if kb.Matches(data, keys.EditorActionSelectCancel) {
		if i.onEscape != nil {
			i.onEscape()
		}
		return
	}

	if kb.Matches(data, keys.EditorActionSubmit) || data == "\n" {
		if i.onSubmit != nil {
			i.onSubmit(i.value)
		}
		return
	}

	hasControlChars := false
	for _, ch := range data {
		code := int(ch)
		if code < 32 || code == 0x7f || (code >= 0x80 && code <= 0x9f) {
			hasControlChars = true
			break
		}
	}

	if !hasControlChars {
		before := i.value[:i.cursor]
		after := i.value[i.cursor:]
		i.value = before + data + after
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
	if content == "" {
		return
	}

	cleanText := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", ""), "\r", ""), "\n", "")

	before := i.value[:i.cursor]
	after := i.value[i.cursor:]
	i.value = before + cleanText + after
	i.cursor += len(cleanText)
}
