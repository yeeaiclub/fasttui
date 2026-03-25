package fasttui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yeeaiclub/fasttui/keys"
)

var (
	SyncOutputBegin = "\x1b[?2026h"
	SyncOutputEnd   = "\x1b[?2026l"

	cellSizeResponsePattern = regexp.MustCompile(`\x1b\[6;(\d+);(\d+)t`)
	cellSizePartialPattern  = regexp.MustCompile(`\x1b(\[6?;?[\d;]*)?$`)
)

type TUI struct {
	Container
	stopped  bool
	terminal Terminal

	fullRedrawCount int
	renderChan      chan RenderRequest
	inputChan       chan InputRequest
	focusChan       chan FocusRequest
	queryChan       chan QueryRequest
	stopChan        chan struct{}

	previousLines       []string // previous rendered line contents
	previousWidth       int      // previous terminal width
	previousViewportTop int      // previous viewport top row number
	maxLinesRendered    int      // max lines ever rendered (for clearing screen)

	cursorRow          int
	hardwareCursorRow  int
	showHardwareCursor bool

	focusedComponent Component

	cellSizeQueryPending bool
	inputBuffer          strings.Builder

	clearOnShrink bool
}

func NewTUI(terminal Terminal, showHardwareCursor bool) *TUI {
	t := &TUI{
		renderChan:         make(chan RenderRequest, 10),
		inputChan:          make(chan InputRequest, 100),
		focusChan:          make(chan FocusRequest, 10),
		queryChan:          make(chan QueryRequest, 10),
		stopChan:           make(chan struct{}),
		terminal:           terminal,
		showHardwareCursor: showHardwareCursor,
		previousLines:      nil,
	}
	return t
}

func (t *TUI) Start() {
	t.start()
	go t.eventLoop()
}

func (t *TUI) Stop() {
	if !t.stopped {
		t.stopped = true
		close(t.stopChan)
		t.terminal.Stop()
	}
}

func (t *TUI) TriggerRender() {
	select {
	case t.renderChan <- RenderRequest{force: false}:
	case <-t.stopChan:
	default:
	}
}

func (t *TUI) ForceRender() {
	select {
	case t.renderChan <- RenderRequest{force: true}:
	case <-t.stopChan:
	default:
	}
}

func (t *TUI) eventLoop() {
	pendingRender := false
	forceRender := false

	t.doRender()

	for {
		select {
		case <-t.stopChan:
			return

		case req := <-t.renderChan:
			if req.force {
				forceRender = true
			}
			pendingRender = true

		case input := <-t.inputChan:
			t.handleInput(input.data)
			pendingRender = true

		case focus := <-t.focusChan:
			t.setFocus(focus.component)
			pendingRender = true

		case query := <-t.queryChan:
			t.handleQueryRequest(query)
		}

		if pendingRender {
			if forceRender {
				t.forceRender()
				forceRender = false
			} else {
				t.doRender()
			}
			pendingRender = false
		}
	}
}

func (t *TUI) forceRender() {
	t.previousLines = nil
	t.previousWidth = -1
	t.cursorRow = 0
	t.hardwareCursorRow = 0
	t.maxLinesRendered = 0
	t.previousViewportTop = 0
	t.doRender()
}

func (t *TUI) doRender() {
	if t.stopped {
		return
	}

	width, height := t.terminal.GetSize()

	viewportTop := max(0, t.maxLinesRendered-height)
	prevViewportTop := t.previousViewportTop
	hardwareCursorRow := t.hardwareCursorRow

	computeLineDiff := func(targetRow int) int {
		cs := hardwareCursorRow - prevViewportTop
		ct := targetRow - viewportTop
		return ct - cs
	}

	newLines := t.renderComponent(width)
	row, col := extractCursorPosition(newLines, height)

	newLines = appendSegmentResetCodes(newLines)
	fullRender := t.getFullRender(newLines, height, row, col, width)

	widthChanged := t.previousWidth != 0 && t.previousWidth != width
	if t.previousLines == nil && !widthChanged {
		fullRender.Render(false)
		return
	}

	if widthChanged {
		fullRender.Render(true)
		return
	}

	// Find first and last changed lines
	firstChangedIdx, lastChangedIdx := findChangedLineRange(t.previousLines, newLines)

	renderLinesLength := len(newLines)
	isAppendedLines := renderLinesLength > len(t.previousLines)
	if isAppendedLines {
		if firstChangedIdx == -1 {
			firstChangedIdx = len(t.previousLines)
		}
		lastChangedIdx = renderLinesLength - 1
	}

	// No changes - but still need to update hardware cursor position if it moved
	if firstChangedIdx == -1 {
		t.moveHardwareCursorTo(row, col, renderLinesLength)
		t.previousViewportTop = max(0, t.maxLinesRendered-height)
		return
	}

	// deleted lines (nothing to render, just clear)
	if firstChangedIdx >= renderLinesLength {
		if len(t.previousLines) > renderLinesLength {
			targetRow := max(0, renderLinesLength-1)
			cursorOffset := computeLineDiff(targetRow)
			extra := renderLinesLength - len(t.previousLines)
			if t.clearTrailingLines(cursorOffset, extra, height, fullRender) {
				return
			}
			t.cursorRow = targetRow
			t.hardwareCursorRow = targetRow
		}

		t.moveHardwareCursorTo(row, col, renderLinesLength)
		t.previousLines = newLines
		t.previousWidth = width
		t.previousViewportTop = max(0, t.maxLinesRendered-height)
		return
	}

	// If first change is above what was visible last frame, incremental draw cannot reach it.
	// Use len(previousLines) (not maxLinesRendered) so content shrink does not inflate viewportTop
	// and force false full re-renders.
	previousContentViewportTop := max(0, len(t.previousLines)-height)
	if firstChangedIdx < previousContentViewportTop {
		fullRender.Render(true)
		return
	}

	// appendStart is true when:
	// - new lines were appended (appendedLines)
	// - the first changed line is exactly at the end of previous content (firstChanged == len(t.previousLines))
	// - there was previous content (firstChanged > 0)
	// This means we should start rendering from the line after the previous content ends
	appendStart := isAppendedLines && firstChangedIdx == len(t.previousLines) && firstChangedIdx > 0
	finalCursorRow := t.renderChangedLines(width, height, firstChangedIdx, lastChangedIdx, newLines, appendStart)

	// Track cursor position for next render
	// cursorRow tracks end of content (for viewport calculation)
	// hardwareCursorRow tracks actual terminal cursor position (for movement)
	t.cursorRow = max(0, len(newLines)-1)
	t.hardwareCursorRow = finalCursorRow

	// Track terminal's working area (grows but doesn't shrink unless cleared)
	t.maxLinesRendered = max(t.maxLinesRendered, renderLinesLength)
	t.previousViewportTop = max(0, t.maxLinesRendered-height)

	// Position hardware cursor for IME
	t.moveHardwareCursorTo(row, col, renderLinesLength)
	t.previousLines = newLines
	t.previousWidth = width
}

func (t *TUI) renderChangedLines(width, height, firstChangedIdx, lastChangedIdx int, newLines []string, appendStart bool) int {
	viewportTop := max(0, t.maxLinesRendered-height)
	hardwareCursorRow := t.hardwareCursorRow
	prevViewportTop := t.previousViewportTop

	computeLineDiff := func(targetRow int) int {
		cs := hardwareCursorRow - prevViewportTop
		ct := targetRow - viewportTop
		return ct - cs
	}

	// Render from first changed line to end
	// Build buffer with all updates wrapped in synchronized output
	var buffer strings.Builder
	buffer.WriteString(SyncOutputBegin) // Begin synchronized output

	// Calculate the bottom row index of the previous viewport
	// Used to determine if scrolling is needed when moving to a target row
	prevViewportBottom := prevViewportTop + height - 1

	moveTargetRow := firstChangedIdx
	if appendStart {
		moveTargetRow = firstChangedIdx - 1
	}

	// If target row is below the visible area, scroll down
	if moveTargetRow > prevViewportBottom {
		// Get current cursor position on screen
		currentScreenRow := max(0, min(height-1, hardwareCursorRow-prevViewportTop))
		// Move cursor to bottom of screen
		moveToBottom := height - 1 - currentScreenRow
		if moveToBottom > 0 {
			buffer.WriteString("\x1b[")
			buffer.WriteString(strconv.Itoa(moveToBottom))
			buffer.WriteString("B")
		}
		// Scroll down to show the target row
		scroll := moveTargetRow - prevViewportBottom
		buffer.WriteString(strings.Repeat("\r\n", scroll))
		prevViewportTop += scroll
		viewportTop += scroll
		hardwareCursorRow = moveTargetRow
	}

	// Move cursor to first changed line (use hardwareCursorRow for actual position)
	lineDiff := computeLineDiff(moveTargetRow)
	if lineDiff > 0 {
		buffer.WriteString("\x1b[")
		buffer.WriteString(strconv.Itoa(lineDiff))
		buffer.WriteString("B") // Move down
	} else if lineDiff < 0 {
		buffer.WriteString("\x1b[")
		buffer.WriteString(strconv.Itoa(-lineDiff))
		buffer.WriteString("A") // Move up
	}

	if appendStart {
		buffer.WriteString("\r\n") // Move to column 0
	} else {
		buffer.WriteString("\r")
	}

	// Only render changed lines (firstChanged to lastChanged), not all lines to end
	// This reduces flicker when only a single line changes (e.g., spinner animation)
	renderEnd := min(lastChangedIdx, len(newLines)-1)
	for i := firstChangedIdx; i <= renderEnd; i++ {
		if i > firstChangedIdx {
			buffer.WriteString("\r\n")
		}
		buffer.WriteString("\x1b[2K") // Clear current line

		line := newLines[i]
		if !containsImage(line) && VisibleWidth(line) > width {
			LogCrashInfo(width, i, line, newLines)
			crashLogPath := GetCrashLogPath()

			t.Stop()
			panic(BuildWidthExceedErrorMsg(i, VisibleWidth(line), width, crashLogPath))
		}
		buffer.WriteString(line)
	}

	// Track where cursor ended up after rendering
	finalCursorRow := renderEnd

	// If we had more lines before, clear them and move cursor back
	if len(t.previousLines) > len(newLines) {
		// Move to end of new content first if we stopped before it
		if renderEnd < len(newLines)-1 {
			moveDown := len(newLines) - 1 - renderEnd
			buffer.WriteString("\x1b[")
			buffer.WriteString(strconv.Itoa(moveDown))
			buffer.WriteString("B")
			finalCursorRow = len(newLines) - 1
		}

		extraLines := len(t.previousLines) - len(newLines)
		for i := len(newLines); i < len(t.previousLines); i++ {
			buffer.WriteString("\r\n\x1b[2K")
		}

		// Move cursor back to end of new content
		buffer.WriteString("\x1b[")
		buffer.WriteString(strconv.Itoa(extraLines))
		buffer.WriteString("A")
	}

	buffer.WriteString(SyncOutputEnd) // End synchronized output

	// Write entire buffer at once
	t.terminal.Write(buffer.String())
	return finalCursorRow
}

// clearTrailingLines clears extra lines in terminal to prevent content scrolling
func (t *TUI) clearTrailingLines(cursorOffset int, extraLines int, height int, fullRender FullRenderer) bool {
	var buffer strings.Builder
	buffer.WriteString(SyncOutputBegin)

	// Move to end of new content (clamp to 0 for empty content)
	if cursorOffset > 0 {
		buffer.WriteString("\x1b[")
		buffer.WriteString(strconv.Itoa(cursorOffset))
		buffer.WriteString("B")
	} else if cursorOffset < 0 {
		buffer.WriteString("\x1b[")
		buffer.WriteString(strconv.Itoa(-cursorOffset))
		buffer.WriteString("A")
	}
	buffer.WriteString("\r")

	// Clear extra lines without scrolling
	if extraLines > height {
		fullRender.Render(true)
		return true
	}

	if extraLines > 0 {
		buffer.WriteString("\x1b[1B")
	}
	for i := range extraLines {
		buffer.WriteString("\r\x1b[2K")
		if i < extraLines-1 {
			buffer.WriteString("\x1b[1B")
		}
	}
	if extraLines > 0 {
		buffer.WriteString("\x1b[")
		buffer.WriteString(strconv.Itoa(extraLines))
		buffer.WriteString("A")
	}

	buffer.WriteString(SyncOutputEnd)
	t.terminal.Write(buffer.String())
	return false
}

func (t *TUI) renderComponent(width int) []string {
	newLines := t.Render(width)
	return newLines
}

type FullRenderer struct {
	newLines []string
	height   int
	width    int
	row      int
	col      int
	tui      *TUI
}

func (f FullRenderer) Render(clear bool) {
	var buffer strings.Builder
	buffer.WriteString(SyncOutputBegin) // Begin synchronized output
	if clear {
		buffer.WriteString("\x1b[3J\x1b[2J\x1b[H") // Clear scrollback, screen, and home
	}
	buffer.WriteString(strings.Join(f.newLines, "\r\n"))
	buffer.WriteString(SyncOutputEnd) // End synchronized output
	f.tui.terminal.Write(buffer.String())

	f.tui.cursorRow = max(0, len(f.newLines)-1)
	f.tui.hardwareCursorRow = f.tui.cursorRow

	if clear {
		f.tui.maxLinesRendered = len(f.newLines)
	} else {
		f.tui.maxLinesRendered = max(f.tui.maxLinesRendered, len(f.newLines))
	}
	f.tui.previousViewportTop = max(0, f.tui.maxLinesRendered-f.height)
	f.tui.moveHardwareCursorTo(f.row, f.col, f.height)
	f.tui.previousLines = f.newLines
	f.tui.previousWidth = f.width
}

func (t *TUI) getFullRender(newLines []string, height int, row int, col int, width int) FullRenderer {
	return FullRenderer{
		newLines: newLines,
		height:   height,
		width:    width,
		row:      row,
		col:      col,
		tui:      t,
	}
}

func (t *TUI) start() error {
	t.stopped = false
	return t.terminal.Start(
		func(data string) {
			t.HandleInput(data)
		},
		func() {
			t.TriggerRender()
		},
	)
}

func (t *TUI) handleQueryRequest(query QueryRequest) {
	switch {
	case query.action == "getShowHardwareCursor":
		query.response <- t.showHardwareCursor
	case query.action == "getFullRedraws":
		query.response <- t.fullRedrawCount
	case query.action == "queryCellSize":
		t.cellSizeQueryPending = true
		t.terminal.Write("\x1b[16t")
		close(query.response)
	case strings.HasPrefix(query.action, "setShowHardwareCursor_"):
		enabled := strings.HasSuffix(query.action, "true")
		t.setShowHardwareCursor(enabled)
		close(query.response)
	case strings.HasPrefix(query.action, "setClearOnShrink_"):
		enabled := strings.HasSuffix(query.action, "true")
		t.clearOnShrink = enabled
		close(query.response)
	}
}

// SetFocus sets the component that currently receives keyboard input.
// In a TUI, only one interactive component (editor, selector, list, etc.) can receive input at a time.
// This method switches the "input focus": first unfocus the old component, then focus the new one.
func (t *TUI) SetFocus(component Component) {
	select {
	case t.focusChan <- FocusRequest{component: component}:
	case <-t.stopChan:
	}
}

func (t *TUI) setFocus(component Component) {
	// Unfocus the previously focused component
	if t.focusedComponent != nil {
		if f, ok := t.focusedComponent.(Focusable); ok {
			f.SetFocused(false)
		}
	}

	// Switch to the new component
	t.focusedComponent = component

	// Activate focus on the new component
	if component != nil {
		if f, ok := t.focusedComponent.(Focusable); ok {
			f.SetFocused(true)
		}
	}
}

func (t *TUI) HandleInput(data string) {
	select {
	case t.inputChan <- InputRequest{data: data}:
	case <-t.stopChan:
	}
}

func (t *TUI) handleInput(data string) {
	if t.cellSizeQueryPending {
		t.inputBuffer.WriteString(data)
		filtered := t.parseCellSizeResponse()
		if filtered == "" {
			return
		}
		data = filtered
	}

	if t.focusedComponent != nil {
		if keys.IsKeyRelease(data) && !t.focusedComponent.WantsKeyRelease() {
			return
		}
		t.focusedComponent.HandleInput(data)
	}
}

func (t *TUI) parseCellSizeResponse() string {
	data := t.inputBuffer.String()

	matches := cellSizeResponsePattern.FindStringSubmatch(data)

	if len(matches) == 3 {
		heightPx, err1 := strconv.Atoi(matches[1])
		widthPx, err2 := strconv.Atoi(matches[2])

		if err1 == nil && err2 == nil && heightPx > 0 && widthPx > 0 {
			t.Invalidate()
			t.TriggerRender()
			t.inputBuffer.Reset()
			t.cellSizeQueryPending = false
			return ""
		}
	}

	if cellSizePartialPattern.MatchString(data) {
		if len(data) > 0 {
			lastChar := data[len(data)-1]
			if !((lastChar >= 'a' && lastChar <= 'z') || (lastChar >= 'A' && lastChar <= 'Z') || lastChar == '~') {
				return ""
			}
		}
	}

	result := t.inputBuffer.String()
	t.inputBuffer.Reset()
	t.cellSizeQueryPending = false
	return result
}

var SEGMENT_RESET = "\x1b[0m\x1b]8;;\x07"

func (t *TUI) SetShowHardwareCursor(enabled bool) {
	select {
	case t.queryChan <- QueryRequest{action: "setShowHardwareCursor_" + fmt.Sprintf("%v", enabled), response: make(chan interface{}, 1)}:
	case <-t.stopChan:
	}
}

func (t *TUI) setShowHardwareCursor(enabled bool) {
	if t.showHardwareCursor == enabled {
		return
	}
	t.showHardwareCursor = enabled
	if !enabled {
		t.terminal.HideCursor()
	}
}

func (t *TUI) SetClearOnShrink(enabled bool) {
	select {
	case t.queryChan <- QueryRequest{action: "setClearOnShrink_" + fmt.Sprintf("%v", enabled), response: make(chan any, 1)}:
	case <-t.stopChan:
	}
}

func (t *TUI) QueryCellSize() {
	if !t.terminal.IsKittyProtocolActive() {
		return
	}
	select {
	case t.queryChan <- QueryRequest{action: "queryCellSize", response: make(chan any, 1)}:
	case <-t.stopChan:
	}
}

func (t *TUI) GetFullRedraws() int {
	respChan := make(chan any, 1)
	select {
	case t.queryChan <- QueryRequest{action: "getFullRedraws", response: respChan}:
		result := <-respChan
		return result.(int)
	case <-t.stopChan:
		return 0
	}
}

func (t *TUI) GetShowHardwareCursor() bool {
	respChan := make(chan any, 1)
	select {
	case t.queryChan <- QueryRequest{action: "getShowHardwareCursor", response: respChan}:
		result := <-respChan
		return result.(bool)
	case <-t.stopChan:
		return false
	}
}

// moveHardwareCursorTo moves the hardware cursor to the specified position.
// If row or col is negative, the cursor will be hidden.
//
// Parameters:
//   - row: target row index (0-based), -1 means no valid position
//   - col: target column index (0-based), -1 means no valid position
//   - totalLines: total number of lines in the viewport (for bounds checking)
func (t *TUI) moveHardwareCursorTo(row int, col int, totalLines int) {
	// Check if no cursor position was found (row == -1, col == -1)
	if (row < 0 || col < 0) || totalLines <= 0 {
		t.terminal.HideCursor()
		return
	}

	targetRow := max(0, min(row, totalLines-1))
	targetCol := max(0, col)

	rowDelta := targetRow - t.hardwareCursorRow
	var builder strings.Builder

	if rowDelta > 0 {
		// move down
		builder.WriteString(fmt.Sprintf("\x1b[%dB", rowDelta))
	} else if rowDelta < 0 {
		// move up
		builder.WriteString(fmt.Sprintf("\x1b[%dA", -rowDelta))
	}

	builder.WriteString(fmt.Sprintf("\x1b[%dG", targetCol+1))
	if builder.Len() > 0 {
		t.terminal.Write(builder.String())
	}

	t.hardwareCursorRow = targetRow

	if t.showHardwareCursor {
		t.terminal.ShowCursor()
	} else {
		t.terminal.HideCursor()
	}
}
