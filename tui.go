package fasttui

import (
	"fmt"
	"strings"
)

type Component interface {
	// Render 根据给定的视口宽度将组件渲染为行。
	// width 是当前视口宽度，返回字符串数组，每个字符串代表一行。
	Render(width int) []string

	// HandleInput 处理组件获得焦点时的键盘输入。
	HandleInput(data string)

	// WantsKeyRelease 返回组件是否接收按键释放事件（Kitty 协议）。
	// 返回 true 表示组件将接收按键释放事件，false 表示释放事件会被过滤掉。
	WantsKeyRelease() bool

	// Invalidate 使任何缓存的渲染状态失效。
	// 在主题更改或组件需要从头重新渲染时调用。
	Invalidate()
}

type Focusable interface {
	Component
	SetFocused(bool)
	IsFocused() bool
}

// OverlayStack 弹窗提示
type OverlayStack struct {
	component Component
}

type TUI struct {
	Container
	stopped             bool
	terminal            Terminal
	previousLines       []string
	previousWidth       int
	focusedComponent    Component
	showHardWareCursor  bool
	cursorRow           int
	hardwareCursorRow   int
	maxLinesRendered    int
	previousViewportTop int
	renderRequested     bool
	fullRedrawCount     int
	renderChan          chan struct{}
	showHardwareCursor  bool
}

func NewTUI(terminal Terminal, showHardwareCursor bool) *TUI {
	t := &TUI{
		renderChan:         make(chan struct{}, 1),
		terminal:           terminal,
		showHardWareCursor: showHardwareCursor,
	}
	return t
}

func (t *TUI) SetShowHardwareCursor(enabled bool) {
	if t.showHardWareCursor == enabled {
		return
	}
	t.showHardWareCursor = enabled
	if !enabled {
		t.terminal.HideCursor()
	}
	t.requestRender(false)
}

func (t *TUI) requestRender(force bool) {
	if force {
		t.previousLines = nil
		t.previousWidth = -1
		t.cursorRow = 0
		t.hardwareCursorRow = 0
		t.maxLinesRendered = 0
		t.previousViewportTop = 0
	}
	if t.renderRequested {
		return
	}
}

func (t *TUI) renderLoop() {
	for range t.renderChan {
		<-t.renderChan
	}
}

func (t *TUI) SetFocus(component Component) {
	if t.focusedComponent != nil {
		if f, ok := t.focusedComponent.(Focusable); ok {
			f.SetFocused(false)
		}
	}

	t.focusedComponent = component

	if component != nil {
		if f, ok := t.focusedComponent.(Focusable); ok {
			f.SetFocused(true)
		}
	}
}

func (t *TUI) start() {
	t.stopped = false
	t.terminal.Start(
		func(data string) {

		},
		func() {

		},
	)
}

func (t *TUI) handleInput() {

}

func (t *TUI) parseCellSizeResponse() string {
	return ""
}

func (t *TUI) doRender() {
	if t.stopped {
		return
	}
	width, height := t.terminal.GetSize()
	viewportTop := max(0, t.maxLinesRendered-height)
	prevViewportTop := t.previousViewportTop
	hardwareCursorRaw := t.hardwareCursorRow
	computeLineDiff := func(targetRow int) int {
		cs := hardwareCursorRaw - prevViewportTop
		ts := targetRow - viewportTop
		return ts - cs
	}

	newLines := t.Render(width)
	row, col := t.extractCursorPosition(newLines, height)
	widthChanged := t.previousWidth != 0 && t.previousWidth != width

	if t.previousLines == nil {
		t.previousLines = []string{}
	}

	if len(t.previousLines) == 0 && !widthChanged {
		t.fullRender(false, newLines, width, height)
		return
	}

	if widthChanged {
		t.fullRender(true, newLines, width, height)
		return
	}

	firstChanged, lastChanged := t.getScope(newLines)
	appendedLines := len(newLines) > len(t.previousLines)
	if appendedLines {
		if firstChanged == -1 {
			firstChanged = len(t.previousLines)
		}
		lastChanged = len(newLines) - 1
	}
	appendStart := appendedLines && firstChanged == len(t.previousLines) && firstChanged > 0

	if firstChanged == -1 {
		t.positionHardwareCursor(row, col, len(newLines))
		t.previousViewportTop = max(0, t.maxLinesRendered-height)
		return
	}

	if firstChanged >= len(newLines) {
		if len(t.previousLines) > len(newLines) {
			var builder strings.Builder
			builder.WriteString("\x1b[?2026h")
			targetRow := max(0, len(newLines)-1)
			lineDiff := computeLineDiff(targetRow)
			if lineDiff > 0 {
				builder.WriteString(fmt.Sprintf("\x1b[%dB", lineDiff))
			} else if lineDiff < 0 {
				builder.WriteString(fmt.Sprintf("\x1b[%dA", -lineDiff))
			}
			builder.WriteString("\r")
			extraLines := len(t.previousLines) - len(newLines)
			if extraLines > height {
				t.fullRender(true, newLines, width, height)
				return
			}
			if extraLines > 0 {
				builder.WriteString("\x1b[1B")
			}
			for i := 0; i < extraLines; i++ {
				builder.WriteString("\r\x1b[2K")
				if i < extraLines-1 {
					builder.WriteString("\x1b[1B")
				}
			}
			if extraLines > 0 {
				builder.WriteString(fmt.Sprintf("\x1b[%dA", extraLines))
			}
			builder.WriteString("\x1b[?2026l")
			t.terminal.Write(builder.String())
			t.cursorRow = targetRow
			t.hardwareCursorRow = targetRow
		}
		t.positionHardwareCursor(row, col, len(newLines))
		t.previousLines = newLines
		t.previousWidth = width
		t.previousViewportTop = max(0, t.maxLinesRendered-height)
		return
	}

	previousContentViewportTop := max(0, len(t.previousLines)-height)
	if firstChanged < previousContentViewportTop {
		t.fullRender(true, newLines, width, height)
		return
	}

	var builder strings.Builder
	builder.WriteString("\x1b[?2026h")
	prevViewportBottom := prevViewportTop + height - 1
	moveTargetRow := firstChanged
	if appendStart {
		moveTargetRow = firstChanged - 1
	}
	if moveTargetRow > prevViewportBottom {
		currentScreenRow := max(0, min(height-1, hardwareCursorRaw-prevViewportTop))
		moveToBottom := height - 1 - currentScreenRow
		if moveToBottom > 0 {
			builder.WriteString(fmt.Sprintf("\x1b[%dB", moveToBottom))
		}
		scroll := moveTargetRow - prevViewportBottom
		for i := 0; i < scroll; i++ {
			builder.WriteString("\r\n")
		}
		prevViewportTop += scroll
		viewportTop += scroll
		hardwareCursorRaw = moveTargetRow
	}

	lineDiff := computeLineDiff(moveTargetRow)
	if lineDiff > 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dB", lineDiff))
	} else if lineDiff < 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dA", -lineDiff))
	}

	if appendStart {
		builder.WriteString("\r\n")
	} else {
		builder.WriteString("\r")
	}

	renderEnd := min(lastChanged, len(newLines)-1)
	for i := firstChanged; i <= renderEnd; i++ {
		if i > firstChanged {
			builder.WriteString("\r\n")
		}
		builder.WriteString("\x1b[2K")
		line := newLines[i]
		if VisibleWidth(line) > width {
			panic(fmt.Sprintf("Rendered line %d exceeds terminal width (%d > %d)", i, VisibleWidth(line), width))
		}
		builder.WriteString(line)
	}

	finalCursorRow := renderEnd
	if len(t.previousLines) > len(newLines) {
		if renderEnd < len(newLines)-1 {
			moveDown := len(newLines) - 1 - renderEnd
			builder.WriteString(fmt.Sprintf("\x1b[%dB", moveDown))
			finalCursorRow = len(newLines) - 1
		}
		extraLines := len(t.previousLines) - len(newLines)
		for i := len(newLines); i < len(t.previousLines); i++ {
			builder.WriteString("\r\n\x1b[2K")
		}
		builder.WriteString(fmt.Sprintf("\x1b[%dA", extraLines))
	}

	builder.WriteString("\x1b[?2026l")
	t.terminal.Write(builder.String())

	t.cursorRow = max(0, len(newLines)-1)
	t.hardwareCursorRow = finalCursorRow
	t.maxLinesRendered = max(t.maxLinesRendered, len(newLines))
	t.previousViewportTop = max(0, t.maxLinesRendered-height)
	t.positionHardwareCursor(row, col, len(newLines))
	t.previousLines = newLines
	t.previousWidth = width
}

func (t *TUI) getScope(newLines []string) (int, int) {
	var (
		firstChanged = -1
		lastChanged  = -1
	)
	maxLines := max(len(newLines), len(t.previousLines))
	for i := 0; i < maxLines; i++ {
		oldLine := ""
		newLine := ""
		if i < len(t.previousLines) {
			oldLine = t.previousLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			if firstChanged == -1 {
				firstChanged = i
			}
			lastChanged = i
		}
	}
	appendLines := len(newLines) > len(t.previousLines)
	if appendLines {
		if firstChanged == -1 {
			firstChanged = len(t.previousLines)
		}
		lastChanged = len(newLines) - 1
	}

	//appendStart := appendLines && firstChanged == len(t.previousLines) && firstChanged > 0
	return firstChanged, lastChanged
}

// buildFullRenderBuffer 将所有的字符输出到终端上面去
func (t *TUI) buildFullRenderBuffer(clear bool, newLines []string) string {
	var builder strings.Builder
	builder.WriteString("\x1b[?2026h")
	if clear {
		builder.WriteString("\x1b[3J\x1b[2J\x1b[H")
	}
	for i, line := range newLines {
		if i > 0 {
			builder.WriteString("\r\n")
		}
		builder.WriteString(line)
	}
	builder.WriteString("\x1b[?2026l")
	return builder.String()
}

func (t *TUI) fullRender(clear bool, newLines []string, width int, height int) {
	t.fullRedrawCount++
	buffer := t.buildFullRenderBuffer(clear, newLines)
	t.terminal.Write(buffer)
	t.updateRenderState(clear, len(newLines), width, height)
}

func (t *TUI) buildIncrRenderBuffer() {
	var builder strings.Builder
	builder.WriteString("\x1b[?2026h]")
}

func (t *TUI) updateRenderState(clear bool, newLinesLen, width, height int) {
	t.cursorRow = max(0, newLinesLen-1)
	t.hardwareCursorRow = t.cursorRow
	if clear {
		t.maxLinesRendered = newLinesLen
	} else {
		t.maxLinesRendered = max(t.maxLinesRendered, newLinesLen)
	}
	t.previousViewportTop = max(0, t.maxLinesRendered-height)
	t.positionHardwareCursor(0, 0, newLinesLen)
	t.previousWidth = width
}

func (t *TUI) fullReader(clear bool, newLines []string) {
	t.fullRedrawCount++
	t.previousLines = newLines
	buffer := t.buildFullRenderBuffer(clear, newLines)
	t.terminal.Write(buffer)
}

func (t *TUI) showOverlay(component Component) {

}

func (t *TUI) hideOverlay() {

}

var CURSOR_MARKER = "\x1b_pi:c\x07"

func (t *TUI) extractCursorPosition(lines []string, height int) (int, int) {
	viewportTop := max(0, len(lines)-height)
	for row := len(lines) - 1; row >= viewportTop; row-- {
		line := lines[row]
		index := strings.Index(line, CursorMarker)
		if index != -1 {
			beforeMarker := line[:index]
			// todo: 需要处理 unicode
			col := VisibleWidth(beforeMarker)
			lines[row] = line[:index] + line[index+len(CURSOR_MARKER):]
			return row, col
		}
	}
	return 0, 0
}

var SEGMENT_RESET = "\x1b[0m\x1b]8;;\x07"

func (t *TUI) applyLineRests(lines []string) []string {
	//todo: 需要处理图片line
	//todo: 暂时不实现
	return lines
}

func (t *TUI) positionHardwareCursor(row int, col int, totalLines int) {
	if (row == 0 && col == 0) || totalLines <= 0 {
		t.terminal.HideCursor()
		return
	}

	targetRow := max(0, min(row, totalLines-1))
	targetCol := max(0, col)

	rowDelta := targetRow - t.hardwareCursorRow
	var builder strings.Builder

	if rowDelta > 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dB", rowDelta))
	} else if rowDelta < 0 {
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
