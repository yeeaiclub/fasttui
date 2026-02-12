package components

import (
	"strings"

	"github.com/yeeaiclub/fasttui"
)

type LayoutLine struct {
	Text      string
	HasCursor bool
	CursorPos int
}

type EditorState struct {
	lines      []string
	cursorCol  int
	cursorLine int
}

type Editor struct {
	history     []string
	state       EditorState
	paddingX    int
	lastWidth   int
	borderColor func(str string) string
}

func NewEditor() *Editor {
	return &Editor{
		history: make([]string, 0),
	}
}

func (e *Editor) AddToHistory(text string) {
	if len(e.history) > 0 && e.history[0] == text {
		return
	}
	e.history = append([]string{text}, e.history...)
	if len(e.history) > 100 {
		e.history = e.history[:100]
	}
}

func (e *Editor) Render(width int) []string {
	maxPadding := max(0, (width-1)/2)
	paddingX := min(e.paddingX, maxPadding)
	contentWidth := max(1, width-paddingX*2)

	// Layout width: with padding the cursor can overflow into it,
	// without padding we reserve 1 column for the cursor.
	var layoutWidth int
	if paddingX > 0 {
		layoutWidth = max(1, contentWidth)
	} else {
		layoutWidth = max(1, contentWidth-1)
	}

	// Store for cursor navigation (must match wrapping width)
	e.lastWidth = layoutWidth

	return nil
}

func (e *Editor) HandleInput(data string) {
	if strings.Contains(data, "\x1b[200~") {

	}
}

func (e *Editor) LayoutText(width int) []LayoutLine {
	layoutLines := make([]LayoutLine, 0)
	if len(e.state.lines) == 0 {
		layoutLines = append(layoutLines, LayoutLine{
			Text:      "",
			HasCursor: true,
		})
	}

	for i := 0; i < len(e.state.lines); i++ {
		line := e.state.lines[i]
		isCurrentLine := i == e.state.cursorLine
		lineVisibleWidth := fasttui.VisibleWidth(line)
		if lineVisibleWidth <= width {
			if isCurrentLine {
				layoutLines = append(layoutLines, LayoutLine{
					Text:      line,
					HasCursor: true,
					CursorPos: e.state.cursorCol,
				})
			} else {
				layoutLines = append(layoutLines, LayoutLine{
					Text:      line,
					HasCursor: false,
				})
			}
		}
	}
	return nil
}

type TextChunk struct {
}

func wordWrapLine(line string, maxWidth int) []TextChunk {
	return nil
}
