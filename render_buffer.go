package fasttui

import (
	"strconv"
	"strings"
)

// RenderBuffer builds ANSI escape sequences efficiently
type RenderBuffer struct {
	builder strings.Builder
}

func NewRenderBuffer() *RenderBuffer {
	return &RenderBuffer{}
}

func (rb *RenderBuffer) BeginSync() {
	rb.builder.WriteString("\x1b[?2026h")
}

func (rb *RenderBuffer) EndSync() {
	rb.builder.WriteString("\x1b[?2026l")
}

func (rb *RenderBuffer) ClearAll() {
	rb.builder.WriteString("\x1b[3J\x1b[2J\x1b[H")
}

func (rb *RenderBuffer) ClearLine() {
	rb.builder.WriteString("\x1b[2K")
}

func (rb *RenderBuffer) CarriageReturn() {
	rb.builder.WriteString("\r")
}

func (rb *RenderBuffer) NewLine() {
	rb.builder.WriteString("\r\n")
}

func (rb *RenderBuffer) Write(text string) {
	rb.builder.WriteString(text)
}

func (rb *RenderBuffer) MoveUp(lines int) {
	if lines > 0 {
		rb.builder.WriteString("\x1b[")
		rb.builder.WriteString(strconv.Itoa(lines))
		rb.builder.WriteString("A")
	}
}

func (rb *RenderBuffer) MoveDown(lines int) {
	if lines > 0 {
		rb.builder.WriteString("\x1b[")
		rb.builder.WriteString(strconv.Itoa(lines))
		rb.builder.WriteString("B")
	}
}

func (rb *RenderBuffer) Scroll(lines int) {
	rb.builder.WriteString(strings.Repeat("\r\n", lines))
}

func (rb *RenderBuffer) MoveCursor(currentRow, targetRow, currentViewportTop, targetViewportTop int) {
	lineDiff := (targetRow - targetViewportTop) - (currentRow - currentViewportTop)
	if lineDiff > 0 {
		rb.MoveDown(lineDiff)
	} else if lineDiff < 0 {
		rb.MoveUp(-lineDiff)
	}
}

func (rb *RenderBuffer) ClearExtraLines(count int) {
	if count > 0 {
		rb.builder.WriteString("\x1b[1B")
	}
	for i := range count {
		rb.CarriageReturn()
		rb.ClearLine()
		if i < count-1 {
			rb.builder.WriteString("\x1b[1B")
		}
	}
	if count > 0 {
		rb.MoveUp(count)
	}
}

func (rb *RenderBuffer) String() string {
	return rb.builder.String()
}
