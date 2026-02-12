package components

import (
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

	term fasttui.Terminal
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

func (e *Editor) Render(width int) {
	maxPadding := max(0, (width-1)/2)
	paddingX := min(e.paddingX, maxPadding)
	contentWidth := max(1, width-paddingX*2)

	layoutWidth := max(1, contentWidth)
	e.layoutWidth = layoutWidth

	layoutLines := e.layoutText(width)
	_, height := e.term.GetSize()
	maxVisibleLines = max(5, int(float64(height)*0.3))
	return
}

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
	Hascursor string
	CursorPos int
}

func (e *Editor) layoutText(width int) []LayoutLine {
	var layoutLines []LayoutLine
	return layoutLines
}
