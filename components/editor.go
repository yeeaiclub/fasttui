package components

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/yeeaiclub/fasttui/keys"

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

	// Callbacks
	OnSubmit func(text string)
	OnChange func(text string)

	// History for up/down navigation
	history []string

	// Kill ring for Emacs-style operations
	killRing    []string
	borderColor func(string) string
}

type EditorState struct {
	lines      []string
	cursorLine int
	cursorCol  int
}

func NewEditor() *Editor {
	return &Editor{
		undoStack:    make([]EditorState, 0),
		state:        EditorState{lines: []string{""}},
		historyIndex: -1,
		history:      make([]string, 0),
		killRing:     make([]string, 0),
	}
}

func (e *Editor) HandleInput(data string) {
	// Handle bracketed paste mode
	if strings.Contains(data, "\x1b[200~") {
		e.isInPaste = true
		e.pasteBuffer = ""
		data = strings.ReplaceAll(data, "\x1b[200~", "")
	}

	if e.isInPaste {
		e.pasteBuffer += data
		endIndex := strings.Index(e.pasteBuffer, "\x1b[201~")
		if endIndex != -1 {
			pasteContent := e.pasteBuffer[:endIndex]
			if len(pasteContent) > 0 {
				e.handlePaste(pasteContent)
			}
			e.isInPaste = false
			remaining := e.pasteBuffer[endIndex+6:]
			e.pasteBuffer = ""
			if len(remaining) > 0 {
				e.HandleInput(remaining)
			}
		}
		return
	}
	kb := keys.GetEditorKeybindings()
	if kb.Matches(data, keys.EditorActionCopy) {
		return
	}

	if kb.Matches(data, keys.EditorActionUndo) {
		e.undo()
		return
	}
}

func (e *Editor) Render(width int) []string {
	maxPadding := max(0, (width-1)/2)
	paddingX := min(e.paddingX, maxPadding)
	contentWidth := max(1, width-paddingX*2)

	layoutWidth := 1
	if paddingX > 0 {
		layoutWidth = max(layoutWidth, contentWidth-paddingX)
	} else {
		layoutWidth = max(layoutWidth, contentWidth-1)
	}
	e.layoutWidth = layoutWidth

	horizontal := e.borderColor("─")

	//layout the text
	layoutLines := e.layoutText(layoutWidth)

	_, height := e.term.GetSize()
	maxVisibleLines := max(5, int(float64(height)*0.3))

	// Find cursor line index
	cursorLineIndex := 0
	for i, line := range layoutLines {
		if line.HasCursor {
			cursorLineIndex = i
			break
		}
	}

	// Adjust scroll offset to keep cursor visible
	if cursorLineIndex < e.scrollOffset {
		e.scrollOffset = cursorLineIndex
	} else if cursorLineIndex >= e.scrollOffset+maxVisibleLines {
		e.scrollOffset = cursorLineIndex - maxVisibleLines + 1
	}

	maxScrollOffset := max(0, len(layoutLines)-maxVisibleLines)
	e.scrollOffset = max(0, min(e.scrollOffset, maxScrollOffset))

	// Get visible lines slice
	visibleLines := layoutLines[e.scrollOffset : e.scrollOffset+maxVisibleLines]

	var result []string
	leftPadding := strings.Repeat(" ", paddingX)
	rightPadding := leftPadding

	if e.scrollOffset > 0 {
		result = append(result, e.renderTopBorder(width, e.scrollOffset))
	} else {
		result = append(result, strings.Repeat(horizontal, width))
	}

	for _, line := range visibleLines {
		displayText := line.Text
		lineVisibleWith := fasttui.VisibleWidth(line.Text)
		cursorInpadding := false
		if line.HasCursor {
			before := displayText[:line.CursorPos]
			after := displayText[line.CursorPos:]
			marker := ""
			if e.focused {
				marker = CURSOR_MARKER
			}
			if len(after) > 0 {
				// Get the first grapheme (rune) from 'after'
				afterRunes := []rune(after)
				var firstGrapheme string
				var restAfter string
				if len(afterRunes) > 0 {
					firstGrapheme = string(afterRunes[0])
					restAfter = string(afterRunes[1:])
				} else {
					firstGrapheme = ""
					restAfter = ""
				}
				cursor := "\x1b[7m" + firstGrapheme + "\x1b[0m"
				displayText = before + marker + cursor + restAfter
				// lineVisibleWith stays the same - we're replacing, not adding
			} else {
				cursor := "\x1b[7m \x1b[0m"
				displayText = before + marker + cursor
				lineVisibleWith = lineVisibleWith + 1
				if lineVisibleWith > contentWidth && paddingX > 0 {
					cursorInpadding = true
				}
			}
		}
		padding := strings.Repeat(" ", max(0, contentWidth-lineVisibleWith))
		var lineRenderPadding string
		if cursorInpadding {
			lineRenderPadding = string(rightPadding[1])
		} else {
			lineRenderPadding = rightPadding
		}
		lineRender := leftPadding + displayText + padding + lineRenderPadding
		result = append(result, lineRender)
	}

	linesBelow := len(layoutLines) - (e.scrollOffset + len(visibleLines))
	if linesBelow > 0 {
		//indicator := fmt.Sprintf("%s")`─── ↓ ${linesBelow} more `
		// remaining := width - fasttui.VisibleWidth(indicator)
		// result.push(this.borderColor(indicator + "─".repeat(Math.max(0, remaining))))
		result = append(result, strings.Repeat(horizontal, width))
	} else {
		result = append(result, strings.Repeat(horizontal, width))
	}
	return result
}

func (e *Editor) renderTopBorder(width int, scrollOffset int) string {
	indicator := fmt.Sprintf("─── ↑ %d more ", scrollOffset)
	remaining := max(width-fasttui.VisibleWidth(indicator), 0)
	return indicator + strings.Repeat("─", remaining)
}

const CURSOR_MARKER = "\x1b_pi:c\x07" // Not used - we render visible cursor instead

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

// SetTerminal sets the terminal for the editor
func (e *Editor) SetTerminal(term fasttui.Terminal) {
	e.term = term
}

// SetPaddingX sets the horizontal padding
func (e *Editor) SetPaddingX(padding int) {
	e.paddingX = padding
}

// SetText sets the editor content
func (e *Editor) SetText(lines []string) {
	e.state.lines = lines
}

// GetText returns the editor content
func (e *Editor) GetText() []string {
	return e.state.lines
}

// SetCursor sets the cursor position
func (e *Editor) SetCursor(line, col int) {
	// Ensure lines exist
	if len(e.state.lines) == 0 {
		e.state.lines = []string{""}
	}

	// Clamp line to valid range
	if line < 0 {
		line = 0
	}
	if line >= len(e.state.lines) {
		line = len(e.state.lines) - 1
	}

	// Clamp column to valid range for the line
	if col < 0 {
		col = 0
	}
	lineLen := len(e.state.lines[line])
	if col > lineLen {
		col = lineLen
	}

	e.state.cursorLine = line
	e.state.cursorCol = col
}

// GetCursor returns the cursor position
func (e *Editor) GetCursor() (line, col int) {
	return e.state.cursorLine, e.state.cursorCol
}

func (e *Editor) handlePaste(content string) {
	e.historyIndex = -1
	e.lastAction = ""
	e.pushUndoSnapshot()

	// Clean the pasted text
	cleanText := strings.ReplaceAll(content, "\r\n", "\n")
	cleanText = strings.ReplaceAll(cleanText, "\r", "\n")

	// Convert tabs to spaces
	cleanText = strings.ReplaceAll(cleanText, "\t", "    ")

	// Filter non-printable characters except newlines
	var filtered strings.Builder
	for _, r := range cleanText {
		if r == '\n' || r >= 32 {
			filtered.WriteRune(r)
		}
	}
	filteredText := filtered.String()

	e.insertTextAtCursorInternal(filteredText)
}

func (e *Editor) addNewLine() {
	e.historyIndex = -1
	e.lastAction = ""
	e.pushUndoSnapshot()

	if len(e.state.lines) == 0 {
		e.state.lines = []string{"", ""}
		e.state.cursorLine = 1
		e.state.cursorCol = 0
		if e.OnChange != nil {
			e.OnChange(e.GetTextString())
		}
		return
	}

	currentLine := ""
	if e.state.cursorLine < len(e.state.lines) {
		currentLine = e.state.lines[e.state.cursorLine]
	}

	before := ""
	after := ""
	if e.state.cursorCol <= len(currentLine) {
		before = currentLine[:e.state.cursorCol]
		after = currentLine[e.state.cursorCol:]
	} else {
		before = currentLine
	}

	// Split current line
	e.state.lines[e.state.cursorLine] = before
	// Insert new line after current
	newLines := make([]string, 0, len(e.state.lines)+1)
	newLines = append(newLines, e.state.lines[:e.state.cursorLine+1]...)
	newLines = append(newLines, after)
	newLines = append(newLines, e.state.lines[e.state.cursorLine+1:]...)
	e.state.lines = newLines

	// Move cursor to start of new line
	e.state.cursorLine++
	e.state.cursorCol = 0

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
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

func (e *Editor) layoutText(contentWidth int) []LayoutLine {
	var layoutLines []LayoutLine

	// If no lines or empty editor, return empty layout with cursor
	if len(e.state.lines) == 0 || (len(e.state.lines) == 1 && e.state.lines[0] == "") {
		layoutLines = append(layoutLines, LayoutLine{
			Text:      "",
			HasCursor: true,
			CursorPos: 0,
		})
		return layoutLines
	}

	// Process each logical line
	for i, line := range e.state.lines {
		isCurrentLine := i == e.state.cursorLine
		lineVisibleWidth := fasttui.VisibleWidth(line)

		if lineVisibleWidth <= contentWidth {
			// Line fits in one layout line
			cursorPos := 0
			if isCurrentLine {
				cursorPos = min(e.state.cursorCol, len(line))
			}
			layoutLines = append(layoutLines, LayoutLine{
				Text:      line,
				HasCursor: isCurrentLine,
				CursorPos: cursorPos,
			})
		} else {
			// Line needs wrapping - character-based wrapping
			runes := []rune(line)
			start := 0

			for start < len(runes) {
				end := start
				width := 0

				// Find how many runes fit in contentWidth
				for end < len(runes) {
					r := runes[end]
					rw := runeWidth(r)
					if width+rw > contentWidth {
						break
					}
					width += rw
					end++
				}

				// Ensure we make progress
				if end == start && start < len(runes) {
					end = start + 1
				}

				chunk := string(runes[start:end])

				// Determine if cursor is in this chunk
				hasCursor := false
				cursorPos := 0
				if isCurrentLine {
					// Cursor position in runes
					cursorRunePos := len([]rune(line[:min(e.state.cursorCol, len(line))]))

					if cursorRunePos >= start && cursorRunePos <= end {
						hasCursor = true
						// Calculate byte position within chunk
						cursorPos = len(string(runes[start:min(cursorRunePos, end)]))
					}
				}

				layoutLines = append(layoutLines, LayoutLine{
					Text:      chunk,
					HasCursor: hasCursor,
					CursorPos: cursorPos,
				})

				start = end
			}
		}
	}

	return layoutLines
}

func runeWidth(r rune) int {
	// Simple width calculation - can be improved with proper Unicode width library
	if r < 32 {
		return 0
	}
	if r >= 0x1100 && (r <= 0x115F || r >= 0x2E80 && r <= 0xA4CF || r >= 0xAC00 && r <= 0xD7A3 || r >= 0xF900 && r <= 0xFAFF || r >= 0xFE10 && r <= 0xFE19 || r >= 0xFE30 && r <= 0xFE6F || r >= 0xFF00 && r <= 0xFF60 || r >= 0xFFE0 && r <= 0xFFE6 || r >= 0x20000 && r <= 0x2FFFD || r >= 0x30000 && r <= 0x3FFFD) {
		return 2
	}
	return 1
}

// insertTextAtCursorInternal inserts text at cursor position
func (e *Editor) insertTextAtCursorInternal(text string) {
	if text == "" {
		return
	}

	// Normalize line endings
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	insertedLines := strings.Split(normalized, "\n")

	if len(e.state.lines) == 0 {
		e.state.lines = []string{""}
	}

	currentLine := ""
	if e.state.cursorLine < len(e.state.lines) {
		currentLine = e.state.lines[e.state.cursorLine]
	}

	beforeCursor := ""
	afterCursor := ""
	if e.state.cursorCol <= len(currentLine) {
		beforeCursor = currentLine[:e.state.cursorCol]
		afterCursor = currentLine[e.state.cursorCol:]
	} else {
		beforeCursor = currentLine
	}

	if len(insertedLines) == 1 {
		// Single line - insert at cursor position
		e.state.lines[e.state.cursorLine] = beforeCursor + normalized + afterCursor
		e.state.cursorCol += len(normalized)
	} else {
		// Multi-line insertion
		newLines := make([]string, 0)
		// All lines before current line
		newLines = append(newLines, e.state.lines[:e.state.cursorLine]...)
		// First inserted line merged with text before cursor
		newLines = append(newLines, beforeCursor+insertedLines[0])
		// All middle inserted lines
		newLines = append(newLines, insertedLines[1:len(insertedLines)-1]...)
		// Last inserted line with text after cursor
		newLines = append(newLines, insertedLines[len(insertedLines)-1]+afterCursor)
		// All lines after current line
		newLines = append(newLines, e.state.lines[e.state.cursorLine+1:]...)

		e.state.lines = newLines
		e.state.cursorLine += len(insertedLines) - 1
		e.state.cursorCol = len(insertedLines[len(insertedLines)-1])
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// insertCharacter inserts a single character at cursor
func (e *Editor) insertCharacter(char string) {
	e.historyIndex = -1
	e.pushUndoSnapshot()
	e.lastAction = "type-word"

	if len(e.state.lines) == 0 {
		e.state.lines = []string{""}
		e.state.cursorLine = 0
		e.state.cursorCol = 0
	}

	// Ensure cursor line is within bounds
	if e.state.cursorLine >= len(e.state.lines) {
		e.state.cursorLine = len(e.state.lines) - 1
	}
	if e.state.cursorLine < 0 {
		e.state.cursorLine = 0
	}

	line := e.state.lines[e.state.cursorLine]

	// Ensure cursor column is within bounds
	if e.state.cursorCol > len(line) {
		e.state.cursorCol = len(line)
	}
	if e.state.cursorCol < 0 {
		e.state.cursorCol = 0
	}

	before := line[:e.state.cursorCol]
	after := line[e.state.cursorCol:]

	e.state.lines[e.state.cursorLine] = before + char + after
	e.state.cursorCol += len(char)

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// handleBackspace handles backspace key
func (e *Editor) handleBackspace() {
	e.historyIndex = -1
	e.lastAction = ""

	if len(e.state.lines) == 0 {
		return
	}

	if e.state.cursorCol > 0 {
		e.pushUndoSnapshot()

		line := e.state.lines[e.state.cursorLine]
		runes := []rune(line[:e.state.cursorCol])
		if len(runes) > 0 {
			before := string(runes[:len(runes)-1])
			after := line[e.state.cursorCol:]
			e.state.lines[e.state.cursorLine] = before + after
			e.state.cursorCol = len(before)
		}
	} else if e.state.cursorLine > 0 {
		e.pushUndoSnapshot()

		currentLine := e.state.lines[e.state.cursorLine]
		previousLine := e.state.lines[e.state.cursorLine-1]

		e.state.lines[e.state.cursorLine-1] = previousLine + currentLine
		e.state.lines = append(e.state.lines[:e.state.cursorLine], e.state.lines[e.state.cursorLine+1:]...)

		e.state.cursorLine--
		e.state.cursorCol = len(previousLine)
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// handleForwardDelete handles delete key
func (e *Editor) handleForwardDelete() {
	e.historyIndex = -1
	e.lastAction = ""

	if len(e.state.lines) == 0 {
		return
	}

	currentLine := e.state.lines[e.state.cursorLine]

	if e.state.cursorCol < len(currentLine) {
		e.pushUndoSnapshot()

		runes := []rune(currentLine)
		if e.state.cursorCol < len(runes) {
			before := string(runes[:e.state.cursorCol])
			after := string(runes[e.state.cursorCol+1:])
			e.state.lines[e.state.cursorLine] = before + after
		}
	} else if e.state.cursorLine < len(e.state.lines)-1 {
		e.pushUndoSnapshot()

		nextLine := e.state.lines[e.state.cursorLine+1]
		e.state.lines[e.state.cursorLine] = currentLine + nextLine
		e.state.lines = append(e.state.lines[:e.state.cursorLine+1], e.state.lines[e.state.cursorLine+2:]...)
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// moveToLineStart moves cursor to start of line
func (e *Editor) moveToLineStart() {
	e.lastAction = ""
	e.state.cursorCol = 0
}

// moveToLineEnd moves cursor to end of line
func (e *Editor) moveToLineEnd() {
	e.lastAction = ""
	if e.state.cursorLine < len(e.state.lines) {
		e.state.cursorCol = len(e.state.lines[e.state.cursorLine])
	}
}

// deleteToStartOfLine deletes from cursor to start of line
func (e *Editor) deleteToStartOfLine() {
	e.historyIndex = -1

	if len(e.state.lines) == 0 {
		return
	}

	currentLine := e.state.lines[e.state.cursorLine]

	if e.state.cursorCol > 0 {
		e.pushUndoSnapshot()

		deletedText := currentLine[:e.state.cursorCol]
		e.addToKillRing(deletedText, true)
		e.lastAction = "kill"

		e.state.lines[e.state.cursorLine] = currentLine[e.state.cursorCol:]
		e.state.cursorCol = 0
	} else if e.state.cursorLine > 0 {
		e.pushUndoSnapshot()

		e.addToKillRing("\n", true)
		e.lastAction = "kill"

		previousLine := e.state.lines[e.state.cursorLine-1]
		e.state.lines[e.state.cursorLine-1] = previousLine + currentLine
		e.state.lines = append(e.state.lines[:e.state.cursorLine], e.state.lines[e.state.cursorLine+1:]...)
		e.state.cursorLine--
		e.state.cursorCol = len(previousLine)
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// deleteToEndOfLine deletes from cursor to end of line
func (e *Editor) deleteToEndOfLine() {
	e.historyIndex = -1

	if len(e.state.lines) == 0 {
		return
	}

	currentLine := e.state.lines[e.state.cursorLine]

	if e.state.cursorCol < len(currentLine) {
		e.pushUndoSnapshot()

		deletedText := currentLine[e.state.cursorCol:]
		e.addToKillRing(deletedText, false)
		e.lastAction = "kill"

		e.state.lines[e.state.cursorLine] = currentLine[:e.state.cursorCol]
	} else if e.state.cursorLine < len(e.state.lines)-1 {
		e.pushUndoSnapshot()

		e.addToKillRing("\n", false)
		e.lastAction = "kill"

		nextLine := e.state.lines[e.state.cursorLine+1]
		e.state.lines[e.state.cursorLine] = currentLine + nextLine
		e.state.lines = append(e.state.lines[:e.state.cursorLine+1], e.state.lines[e.state.cursorLine+2:]...)
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// deleteWordBackwards deletes word backwards
func (e *Editor) deleteWordBackwards() {
	e.historyIndex = -1

	if len(e.state.lines) == 0 {
		return
	}

	currentLine := e.state.lines[e.state.cursorLine]

	if e.state.cursorCol == 0 {
		if e.state.cursorLine > 0 {
			e.pushUndoSnapshot()

			e.addToKillRing("\n", true)
			e.lastAction = "kill"

			previousLine := e.state.lines[e.state.cursorLine-1]
			e.state.lines[e.state.cursorLine-1] = previousLine + currentLine
			e.state.lines = append(e.state.lines[:e.state.cursorLine], e.state.lines[e.state.cursorLine+1:]...)
			e.state.cursorLine--
			e.state.cursorCol = len(previousLine)
		}
	} else {
		e.pushUndoSnapshot()

		oldCursorCol := e.state.cursorCol
		e.moveWordBackwards()
		deleteFrom := e.state.cursorCol

		deletedText := currentLine[deleteFrom:oldCursorCol]
		e.addToKillRing(deletedText, true)
		e.lastAction = "kill"

		e.state.lines[e.state.cursorLine] = currentLine[:deleteFrom] + currentLine[oldCursorCol:]
		e.state.cursorCol = deleteFrom
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// deleteWordForward deletes word forward
func (e *Editor) deleteWordForward() {
	e.historyIndex = -1

	if len(e.state.lines) == 0 {
		return
	}

	currentLine := e.state.lines[e.state.cursorLine]

	if e.state.cursorCol >= len(currentLine) {
		if e.state.cursorLine < len(e.state.lines)-1 {
			e.pushUndoSnapshot()

			e.addToKillRing("\n", false)
			e.lastAction = "kill"

			nextLine := e.state.lines[e.state.cursorLine+1]
			e.state.lines[e.state.cursorLine] = currentLine + nextLine
			e.state.lines = append(e.state.lines[:e.state.cursorLine+1], e.state.lines[e.state.cursorLine+2:]...)
		}
	} else {
		e.pushUndoSnapshot()

		oldCursorCol := e.state.cursorCol
		e.moveWordForwards()
		deleteTo := e.state.cursorCol

		deletedText := currentLine[oldCursorCol:deleteTo]
		e.addToKillRing(deletedText, false)
		e.lastAction = "kill"

		e.state.lines[e.state.cursorLine] = currentLine[:oldCursorCol] + currentLine[deleteTo:]
		e.state.cursorCol = oldCursorCol
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// moveCursor moves cursor by delta
func (e *Editor) moveCursor(deltaLine, deltaCol int) {
	e.lastAction = ""

	if len(e.state.lines) == 0 {
		return
	}

	if deltaLine != 0 {
		newLine := e.state.cursorLine + deltaLine
		if newLine >= 0 && newLine < len(e.state.lines) {
			e.state.cursorLine = newLine
			if e.state.cursorCol > len(e.state.lines[e.state.cursorLine]) {
				e.state.cursorCol = len(e.state.lines[e.state.cursorLine])
			}
		}
	}

	if deltaCol != 0 {
		currentLine := e.state.lines[e.state.cursorLine]

		if deltaCol > 0 {
			if e.state.cursorCol < len(currentLine) {
				runes := []rune(currentLine[e.state.cursorCol:])
				if len(runes) > 0 {
					e.state.cursorCol += len(string(runes[0]))
				}
			} else if e.state.cursorLine < len(e.state.lines)-1 {
				e.state.cursorLine++
				e.state.cursorCol = 0
			}
		} else {
			if e.state.cursorCol > 0 {
				runes := []rune(currentLine[:e.state.cursorCol])
				if len(runes) > 0 {
					e.state.cursorCol -= len(string(runes[len(runes)-1]))
				}
			} else if e.state.cursorLine > 0 {
				e.state.cursorLine--
				e.state.cursorCol = len(e.state.lines[e.state.cursorLine])
			}
		}
	}
}

// moveWordBackwards moves cursor backwards by one word
func (e *Editor) moveWordBackwards() {
	e.lastAction = ""

	if len(e.state.lines) == 0 {
		return
	}

	currentLine := e.state.lines[e.state.cursorLine]

	if e.state.cursorCol == 0 {
		if e.state.cursorLine > 0 {
			e.state.cursorLine--
			e.state.cursorCol = len(e.state.lines[e.state.cursorLine])
		}
		return
	}

	runes := []rune(currentLine[:e.state.cursorCol])
	pos := len(runes) - 1

	// Skip trailing whitespace
	for pos >= 0 && unicode.IsSpace(runes[pos]) {
		pos--
	}

	if pos >= 0 {
		if unicode.IsPunct(runes[pos]) {
			// Skip punctuation run
			for pos >= 0 && unicode.IsPunct(runes[pos]) {
				pos--
			}
		} else {
			// Skip word run
			for pos >= 0 && !unicode.IsSpace(runes[pos]) && !unicode.IsPunct(runes[pos]) {
				pos--
			}
		}
	}

	e.state.cursorCol = len(string(runes[:pos+1]))
}

// moveWordForwards moves cursor forwards by one word
func (e *Editor) moveWordForwards() {
	e.lastAction = ""

	if len(e.state.lines) == 0 {
		return
	}

	currentLine := e.state.lines[e.state.cursorLine]

	if e.state.cursorCol >= len(currentLine) {
		if e.state.cursorLine < len(e.state.lines)-1 {
			e.state.cursorLine++
			e.state.cursorCol = 0
		}
		return
	}

	runes := []rune(currentLine[e.state.cursorCol:])
	pos := 0

	// Skip leading whitespace
	for pos < len(runes) && unicode.IsSpace(runes[pos]) {
		pos++
	}

	if pos < len(runes) {
		if unicode.IsPunct(runes[pos]) {
			// Skip punctuation run
			for pos < len(runes) && unicode.IsPunct(runes[pos]) {
				pos++
			}
		} else {
			// Skip word run
			for pos < len(runes) && !unicode.IsSpace(runes[pos]) && !unicode.IsPunct(runes[pos]) {
				pos++
			}
		}
	}

	e.state.cursorCol += len(string(runes[:pos]))
}

// yank pastes from kill ring
func (e *Editor) yank() {
	if len(e.killRing) == 0 {
		return
	}

	e.pushUndoSnapshot()

	text := e.killRing[len(e.killRing)-1]
	e.insertYankedText(text)

	e.lastAction = "yank"
}

// yankPop cycles through kill ring
func (e *Editor) yankPop() {
	if e.lastAction != "yank" || len(e.killRing) <= 1 {
		return
	}

	e.pushUndoSnapshot()

	e.deleteYankedText()

	// Rotate ring
	lastEntry := e.killRing[len(e.killRing)-1]
	e.killRing = append([]string{lastEntry}, e.killRing[:len(e.killRing)-1]...)

	text := e.killRing[len(e.killRing)-1]
	e.insertYankedText(text)

	e.lastAction = "yank"
}

// insertYankedText inserts text at cursor
func (e *Editor) insertYankedText(text string) {
	e.historyIndex = -1
	lines := strings.Split(text, "\n")

	if len(e.state.lines) == 0 {
		e.state.lines = []string{""}
	}

	if len(lines) == 1 {
		currentLine := e.state.lines[e.state.cursorLine]
		before := currentLine[:e.state.cursorCol]
		after := currentLine[e.state.cursorCol:]
		e.state.lines[e.state.cursorLine] = before + text + after
		e.state.cursorCol += len(text)
	} else {
		currentLine := e.state.lines[e.state.cursorLine]
		before := currentLine[:e.state.cursorCol]
		after := currentLine[e.state.cursorCol:]

		newLines := make([]string, 0)
		newLines = append(newLines, e.state.lines[:e.state.cursorLine]...)
		newLines = append(newLines, before+lines[0])
		newLines = append(newLines, lines[1:len(lines)-1]...)
		newLines = append(newLines, lines[len(lines)-1]+after)
		newLines = append(newLines, e.state.lines[e.state.cursorLine+1:]...)

		e.state.lines = newLines
		e.state.cursorLine += len(lines) - 1
		e.state.cursorCol = len(lines[len(lines)-1])
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// deleteYankedText deletes previously yanked text
func (e *Editor) deleteYankedText() {
	if len(e.killRing) == 0 {
		return
	}

	yankedText := e.killRing[len(e.killRing)-1]
	yankLines := strings.Split(yankedText, "\n")

	if len(yankLines) == 1 {
		currentLine := e.state.lines[e.state.cursorLine]
		deleteLen := len(yankedText)
		before := currentLine[:e.state.cursorCol-deleteLen]
		after := currentLine[e.state.cursorCol:]
		e.state.lines[e.state.cursorLine] = before + after
		e.state.cursorCol -= deleteLen
	} else {
		startLine := e.state.cursorLine - (len(yankLines) - 1)
		startCol := len(e.state.lines[startLine]) - len(yankLines[0])

		afterCursor := e.state.lines[e.state.cursorLine][e.state.cursorCol:]
		beforeYank := e.state.lines[startLine][:startCol]

		newLines := make([]string, 0)
		newLines = append(newLines, e.state.lines[:startLine]...)
		newLines = append(newLines, beforeYank+afterCursor)
		newLines = append(newLines, e.state.lines[e.state.cursorLine+1:]...)

		e.state.lines = newLines
		e.state.cursorLine = startLine
		e.state.cursorCol = startCol
	}

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// addToKillRing adds text to kill ring
func (e *Editor) addToKillRing(text string, prepend bool) {
	if text == "" {
		return
	}

	if e.lastAction == "kill" && len(e.killRing) > 0 {
		lastEntry := e.killRing[len(e.killRing)-1]
		e.killRing = e.killRing[:len(e.killRing)-1]
		if prepend {
			e.killRing = append(e.killRing, text+lastEntry)
		} else {
			e.killRing = append(e.killRing, lastEntry+text)
		}
	} else {
		e.killRing = append(e.killRing, text)
	}
}

// undo undoes last action
func (e *Editor) undo() {
	e.historyIndex = -1
	if len(e.undoStack) == 0 {
		return
	}

	snapshot := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.state = EditorState{
		lines:      append([]string{}, snapshot.lines...),
		cursorLine: snapshot.cursorLine,
		cursorCol:  snapshot.cursorCol,
	}
	e.lastAction = ""

	if e.OnChange != nil {
		e.OnChange(e.GetTextString())
	}
}

// handleSubmit handles submit action
func (e *Editor) handleSubmit() {
	result := strings.TrimSpace(strings.Join(e.state.lines, "\n"))

	e.state = EditorState{lines: []string{""}, cursorLine: 0, cursorCol: 0}
	e.historyIndex = -1
	e.scrollOffset = 0
	e.undoStack = make([]EditorState, 0)
	e.lastAction = ""

	if e.OnChange != nil {
		e.OnChange("")
	}
	if e.OnSubmit != nil {
		e.OnSubmit(result)
	}
}

// GetTextString returns text as a single string
func (e *Editor) GetTextString() string {
	return strings.Join(e.state.lines, "\n")
}
