package components

import (
	"strconv"
	"strings"

	"github.com/yeeaiclub/fasttui"
)

type Editor struct {
	isInPaste    bool
	pasteBuffer  string
	historyIndex int
	lastAction   string
	undoStack    []EditorState
	state        EditorState

	paddingX    int
	layoutWidth int

	term         fasttui.Terminal
	scrollOffset int

	focused bool
}

type EditorState struct {
	lines      []string
	cursorLine int
	cursorCol  int
}

func NewEditor() *Editor {
	return &Editor{
		undoStack: make([]EditorState, 0),
	}
}

func (e *Editor) HandleInput(data string) {
	if strings.Contains(data, "\x1b[200~") {
		e.isInPaste = true
		e.pasteBuffer = ""
		data = strings.ReplaceAll(data, "", "\x1b[200~")
	}

	if e.isInPaste {
		e.handlePaste(data)
		return
	}
}

func (e *Editor) Render(width int) []string {
	maxPadding := max(0, (width-1)/2)
	paddingX := min(e.paddingX, maxPadding)
	contentWidth := max(1, width-paddingX*2)

	leftPadding := strings.Repeat(" ", paddingX)
	rightPadding := leftPadding

	layoutWidth := max(1, contentWidth)
	e.layoutWidth = layoutWidth

	_, height := e.term.GetSize()
	maxVisibleLines := max(5, int(float64(height)*0.3))
	layoutLines := e.layoutText(width)
	index := 0
	for i, line := range layoutLines {
		if line.HasCursor {
			index = i
			break
		}
	}

	if index < e.scrollOffset {
		e.scrollOffset = index
	} else if index >= e.scrollOffset+maxVisibleLines {
		e.scrollOffset = index - maxVisibleLines + 1
	}
	maxScrollOffset := max(0, len(layoutLines)-maxVisibleLines)
	e.scrollOffset = max(0, min(e.scrollOffset, maxScrollOffset))

	visibleLines := layoutLines[e.scrollOffset : e.scrollOffset+maxVisibleLines]
	var result []string

	// Render top border (with scroll indicator if scrolled down)
	horizontal := "─"
	if e.scrollOffset > 0 {
		indicator := "─── ↑ " + strconv.Itoa(e.scrollOffset) + " more "
		indicatorWidth := fasttui.VisibleWidth(indicator)
		remaining := width - indicatorWidth
		borderLine := indicator + strings.Repeat(horizontal, max(0, remaining))
		result = append(result, borderLine)
	} else {
		result = append(result, strings.Repeat(horizontal, width))
	}

	// for _, line := range visibleLines {
	// 	displayText := line.Text
	// 	lineVisibleWidth := fasttui.VisibleWidth(displayText)

	// 	if line.HasCursor {
	// 		before := displayText[:line.CursorPos]
	// 		after := displayText[line.CursorPos:]
	// 	}
	// }
	return result
}

func (e *Editor) IsFocused() bool {
	return e.focused
}

func (e *Editor) SetFocused(focused bool) {
	e.focused = focused
}

func (e *Editor) WantsKeyRelease() bool {
	return true
}

func (e *Editor) Invalidate() {}
func (e *Editor) handlePaste(data string) {
	e.pasteBuffer += data
	index := strings.Index(data, "\x1b[201~")
	if index == -1 {
		return
	}

	content := e.pasteBuffer[:index]
	if len(content) > 0 {
		e.processPastedContent(content)
	}
	e.isInPaste = false
	e.pasteBuffer = ""
	remaining := e.pasteBuffer[index+6:]
	if remaining != "" {
		e.HandleInput(remaining)
	}
}

func (e *Editor) processPastedContent(content string) {
}

func (e *Editor) addNewLine() {
	e.historyIndex = -1
	e.lastAction = ""
	e.pushUndoSnapshot()
}

func (e *Editor) pushUndoSnapshot() {
	state := EditorState{
		cursorLine: e.state.cursorLine,
		cursorCol:  e.state.cursorCol,
		lines:      append([]string{}, e.state.lines...),
	}
	e.undoStack = append(e.undoStack, state)
}

type LayoutLine struct {
	Text      string
	HasCursor bool
	CursorPos int
}

func (e *Editor) layoutText(width int) []LayoutLine {
	var layoutLines []LayoutLine
	return layoutLines
}
