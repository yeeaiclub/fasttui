package fasttui

import (
	"fmt"
	"strings"

	"github.com/yeeaiclub/fasttui/keys"
)

type TUI struct {
	Container
	stopped  bool
	terminal Terminal

	renderRequested bool
	fullRedrawCount int
	renderChan      chan struct{}

	previousLines       []string
	previousWidth       int
	previousViewportTop int
	maxLinesRendered    int

	cursorRow          int
	hardwareCursorRow  int
	showHardwareCursor bool

	focusedComponent Component

	cellSizeQueryPending bool
	inoutBuffer          strings.Builder

	overlayStacks []OverlayStack
}

func NewTUI(terminal Terminal, showHardwareCursor bool) *TUI {
	t := &TUI{
		renderChan:         make(chan struct{}, 1),
		overlayStacks:      make([]OverlayStack, 0),
		terminal:           terminal,
		showHardwareCursor: showHardwareCursor,
		previousLines:      nil,
	}
	return t
}

func (t *TUI) GetFullRedraws() int {
	return t.fullRedrawCount
}

func (t *TUI) GetShowHardwareCursor() bool {
	return t.showHardwareCursor
}

func (t *TUI) SetShowHardwareCursor(enabled bool) {
	if t.showHardwareCursor == enabled {
		return
	}
	t.showHardwareCursor = enabled
	if !enabled {
		t.terminal.HideCursor()
	}
	t.requestRender(false)
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

func (t *TUI) ShowOverlay(component Component, options OverlayOption) (func(), func(bool), func() bool) {
	entry := OverlayStack{
		component: component,
		options:   options,
		preFocus:  t.focusedComponent,
	}
	t.overlayStacks = append(t.overlayStacks, entry)

	if t.isOverlayVisible(&entry) {
		t.SetFocus(entry.component)
	}
	t.terminal.HideCursor()
	t.requestRender(false)

	hide := func() {
		index := -1
		for i, e := range t.overlayStacks {
			if e.component == component {
				index = i
				break
			}
		}
		if index != -1 {
			t.overlayStacks = append(t.overlayStacks[:index], t.overlayStacks[index+1:]...)
			if t.focusedComponent == component {
				topVisible := t.getTopmostVisibleOverlay()
				if topVisible != nil {
					t.SetFocus(topVisible.component)
				} else {
					t.SetFocus(entry.preFocus)
				}
			}
			if len(t.overlayStacks) == 0 {
				t.terminal.HideCursor()
			}
			t.requestRender(false)
		}
	}

	setHidden := func(hidden bool) {
		if entry.hidden == hidden {
			return
		}
		entry.hidden = hidden
		if hidden {
			if t.focusedComponent == component {
				topVisible := t.getTopmostVisibleOverlay()
				if topVisible != nil {
					t.SetFocus(topVisible.component)
				} else {
					t.SetFocus(entry.preFocus)
				}
			}
		} else {
			if t.isOverlayVisible(&entry) {
				t.SetFocus(component)
			}
		}
		t.requestRender(false)
	}

	return hide, setHidden, func() bool { return entry.hidden }
}

func (t *TUI) isOverlayVisible(entry *OverlayStack) bool {
	if entry.hidden {
		return false
	}
	if entry.options.Visible != nil {
		width, height := t.terminal.GetSize()
		return entry.options.Visible(width, height)
	}
	return true
}

func (t *TUI) HideOverlay() {
	// POP last overlay
	if len(t.overlayStacks) > 0 {
		entry := t.overlayStacks[len(t.overlayStacks)-1]
		t.overlayStacks = t.overlayStacks[:len(t.overlayStacks)-1]
		if !entry.hidden {
			t.SetFocus(entry.preFocus)
		}
	}
}

// GetTopmostVisibleOverlay returns the topmost visible overlay, or nil if none.
func (t *TUI) getTopmostVisibleOverlay() *OverlayStack {
	for i := len(t.overlayStacks) - 1; i >= 0; i-- {
		entry := t.overlayStacks[i]
		if t.isOverlayVisible(&entry) {
			return &entry
		}
	}
	return nil
}

func (t *TUI) HasOverlay() bool {
	for _, entry := range t.overlayStacks {
		if t.isOverlayVisible(&entry) {
			return true
		}
	}
	return false
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
	t.renderRequested = true
	select {
	case t.renderChan <- struct{}{}:
	default:
	}
}

func (t *TUI) HandleInput(data string) {
	if t.cellSizeQueryPending {
		t.inoutBuffer.WriteString(data)
		filtered := t.parseCellSizeResponse()
		if filtered == "" {
			return
		}
		data = filtered
	}

	var focusedOverlay *OverlayStack
	for i := range t.overlayStacks {
		if t.overlayStacks[i].component == t.focusedComponent {
			focusedOverlay = &t.overlayStacks[i]
			break
		}
	}
	if focusedOverlay != nil && !t.isOverlayVisible(focusedOverlay) {
		topVisible := t.getTopmostVisibleOverlay()
		if topVisible != nil {
			t.SetFocus(topVisible.component)
		} else {
			t.SetFocus(focusedOverlay.preFocus)
		}
	}

	if t.focusedComponent != nil {
		if keys.IsKeyRelease(data) && !t.focusedComponent.WantsKeyRelease() {
			return
		}
		t.focusedComponent.HandleInput(data)
		t.requestRender(false)
	}
}

func (t *TUI) QueryCellSize() {
	if !t.terminal.IsKittyProtocolActive() {
		return
	}
	t.cellSizeQueryPending = true
	t.terminal.Write("\x1b[16t")
}

func (t *TUI) ResolveOverlayLayout(options OverlayOption, overlayHeight int, termWidth int, termHeight int) OverlayLayout {
	marginTop, marginRight, marginBottom, marginLeft := t.parseMargin(options.Margin)

	availWidth := max(1, termWidth-marginLeft-marginRight)
	availHeight := max(1, termHeight-marginTop-marginBottom)

	width := t.parseSizeValue(options.Width, termWidth)
	if width == 0 {
		width = min(80, availWidth)
	}
	if options.MiniWidth > 0 {
		width = max(width, options.MiniWidth)
	}
	width = max(1, min(width, availWidth))

	var maxHeight *int
	if options.MaxHeight > 0 {
		maxHeightVal := t.parseSizeValue(options.MaxHeight, termHeight)
		if maxHeightVal > 0 {
			maxHeightVal = max(1, min(maxHeightVal, availHeight))
			maxHeight = &maxHeightVal
		}
	}

	effectiveHeight := overlayHeight
	if maxHeight != nil {
		effectiveHeight = min(overlayHeight, *maxHeight)
	}

	var row, col int

	if options.Row != 0 {
		row = options.Row
	} else {
		anchor := options.Anchor
		if anchor == "" {
			anchor = AnchorCenter
		}
		row = t.resolveAnchorRow(anchor, effectiveHeight, availHeight, marginTop)
	}

	if options.Col != 0 {
		col = options.Col
	} else {
		anchor := options.Anchor
		if anchor == "" {
			anchor = AnchorCenter
		}
		col = t.resolveAnchorCol(anchor, width, availWidth, marginLeft)
	}

	if options.OffsetY != 0 {
		row += options.OffsetY
	}
	if options.OffsetX != 0 {
		col += options.OffsetX
	}

	row = max(marginTop, min(row, termHeight-marginBottom-effectiveHeight))
	col = max(marginLeft, min(col, termWidth-marginRight-width))

	return OverlayLayout{
		width:     width,
		row:       row,
		col:       col,
		maxHeight: maxHeight,
	}
}

func (t *TUI) parseMargin(margin any) (marginTop, marginRight, marginBottom, marginLeft int) {
	if margin == nil {
		return 0, 0, 0, 0
	}

	switch v := margin.(type) {
	case int:
		return max(0, v), max(0, v), max(0, v), max(0, v)
	case map[string]int:
		return max(0, v["top"]), max(0, v["right"]), max(0, v["bottom"]), max(0, v["left"])
	default:
		return 0, 0, 0, 0
	}
}

func (t *TUI) parseSizeValue(value int, total int) int {
	if value <= 0 {
		return 0
	}
	return value
}

func (t *TUI) resolveAnchorRow(anchor OverlayAnchor, height int, availHeight int, marginTop int) int {
	switch anchor {
	case AnchorTopLeft, AnchorTopCenter, AnchorTopRight:
		return marginTop
	case AnchorBottomLeft, AnchorBottomCenter, AnchorBottomRight:
		return marginTop + max(0, availHeight-height)
	case AnchorLeftCenter, AnchorRightCenter, AnchorCenter:
		return marginTop + max(0, availHeight-height)/2
	default:
		return marginTop
	}
}

func (t *TUI) resolveAnchorCol(anchor OverlayAnchor, width int, availWidth int, marginLeft int) int {
	switch anchor {
	case AnchorTopLeft, AnchorBottomLeft, AnchorLeftCenter:
		return marginLeft
	case AnchorTopRight, AnchorBottomRight, AnchorRightCenter:
		return marginLeft + max(0, availWidth-width)
	case AnchorTopCenter, AnchorBottomCenter, AnchorCenter:
		return marginLeft + max(0, availWidth-width)/2
	default:
		return marginLeft
	}
}

func (t *TUI) renderLoop() {
	for range t.renderChan {
		t.renderRequested = false
		t.doRender()
	}
}

func (t *TUI) start() error {
	t.stopped = false
	return t.terminal.Start(
		func(data string) {
			t.HandleInput(data)
		},
		func() {
			t.requestRender(false)
		},
	)
}

func (t *TUI) Start() {
	t.start()
	go t.renderLoop()
	t.doRender()
}

func (t *TUI) Stop() {
	t.stopped = true
	t.terminal.Stop()
}

func (t *TUI) handleInput(data string) {
	if t.cellSizeQueryPending {
		t.inoutBuffer.WriteString(data)
	}
}

func (t *TUI) parseCellSizeResponse() string {
	return ""
}

type renderContext struct {
	width             int
	height            int
	viewportTop       int
	prevViewportTop   int
	hardwareCursorRaw int
}

func (t *TUI) newRenderContext() renderContext {
	width, height := t.terminal.GetSize()
	return renderContext{
		width:             width,
		height:            height,
		viewportTop:       max(0, t.maxLinesRendered-height),
		prevViewportTop:   t.previousViewportTop,
		hardwareCursorRaw: t.hardwareCursorRow,
	}
}

func (ctx *renderContext) computeLineDiff(targetRow int) int {
	cs := ctx.hardwareCursorRaw - ctx.prevViewportTop
	ts := targetRow - ctx.viewportTop
	return ts - cs
}

func (t *TUI) doRender() {
	if t.stopped {
		return
	}

	ctx := t.newRenderContext()
	newLines := t.Render(ctx.width)
	cursorRow, cursorCol := t.extractCursorPosition(newLines, ctx.height)
	widthChanged := t.previousWidth != 0 && t.previousWidth != ctx.width

	if t.shouldFullRenderInit(widthChanged) {
		t.fullRender(widthChanged, newLines, ctx.width, ctx.height)
		return
	}

	firstChanged, lastChanged := t.prepareRenderScope(newLines)

	if firstChanged == -1 {
		t.updateViewportOnly(ctx, cursorRow, cursorCol, len(newLines))
		return
	}

	if firstChanged >= len(newLines) {
		t.handleLineDeletion(ctx, newLines, cursorRow, cursorCol)
		return
	}

	if t.shouldFullRender(ctx, firstChanged) {
		t.fullRender(true, newLines, ctx.width, ctx.height)
		return
	}

	t.incrementalRender(ctx, newLines, firstChanged, lastChanged, cursorRow, cursorCol)
}

func (t *TUI) shouldFullRenderInit(widthChanged bool) bool {
	if t.previousLines == nil {
		t.previousLines = []string{}
	}
	return len(t.previousLines) == 0 || widthChanged
}

func (t *TUI) prepareRenderScope(newLines []string) (firstChanged, lastChanged int) {
	firstChanged, lastChanged = t.getScope(newLines)
	appendedLines := len(newLines) > len(t.previousLines)
	if appendedLines {
		if firstChanged == -1 {
			firstChanged = len(t.previousLines)
		}
		lastChanged = len(newLines) - 1
	}
	return
}

func (t *TUI) updateViewportOnly(ctx renderContext, row, col, totalLines int) {
	t.positionHardwareCursor(row, col, totalLines)
	t.previousViewportTop = max(0, t.maxLinesRendered-ctx.height)
}

func (t *TUI) handleLineDeletion(ctx renderContext, newLines []string, row int, col int) {
	if len(t.previousLines) <= len(newLines) {
		t.updateRenderStateAfter(ctx, newLines, row, col)
		return
	}

	extraLines := len(t.previousLines) - len(newLines)
	if extraLines > ctx.height {
		t.fullRender(true, newLines, ctx.width, ctx.height)
		return
	}

	t.deleteExtraLines(ctx, newLines, extraLines, row, col)
}

func (t *TUI) deleteExtraLines(ctx renderContext, newLines []string, extraLines int, row int, col int) {
	var builder strings.Builder
	builder.WriteString("\x1b[?2026h")

	targetRow := max(0, len(newLines)-1)
	t.moveCursorToRow(ctx, &builder, targetRow)
	builder.WriteString("\r")

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

	t.updateRenderStateAfter(ctx, newLines, row, col)
}

func (t *TUI) moveCursorToRow(ctx renderContext, builder *strings.Builder, targetRow int) {
	lineDiff := ctx.computeLineDiff(targetRow)
	if lineDiff > 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dB", lineDiff))
	} else if lineDiff < 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dA", -lineDiff))
	}
}

func (t *TUI) updateRenderStateAfter(ctx renderContext, newLines []string, row int, col int) {
	t.positionHardwareCursor(row, col, len(newLines))
	t.previousLines = newLines
	t.previousWidth = ctx.width
	t.previousViewportTop = max(0, t.maxLinesRendered-ctx.height)
}

func (t *TUI) shouldFullRender(ctx renderContext, firstChanged int) bool {
	previousContentViewportTop := max(0, len(t.previousLines)-ctx.height)
	return firstChanged < previousContentViewportTop
}

func (t *TUI) incrementalRender(ctx renderContext, newLines []string, firstChanged int, lastChanged int, row int, col int) {
	appendedLines := len(newLines) > len(t.previousLines)
	appendStart := appendedLines && firstChanged == len(t.previousLines) && firstChanged > 0

	var builder strings.Builder
	builder.WriteString("\x1b[?2026h")

	t.prepareCursorForRender(ctx, &builder, firstChanged, appendStart)
	t.renderChangedLines(ctx, &builder, newLines, firstChanged, lastChanged)
	finalCursorRow := t.handleTrailingLines(&builder, newLines, lastChanged)

	builder.WriteString("\x1b[?2026l")
	t.terminal.Write(builder.String())

	t.updateRenderStateAfterIncremental(ctx, newLines, finalCursorRow, row, col)
}

func (t *TUI) prepareCursorForRender(ctx renderContext, builder *strings.Builder, firstChanged int, appendStart bool) {
	prevViewportBottom := ctx.prevViewportTop + ctx.height - 1
	moveTargetRow := firstChanged
	if appendStart {
		moveTargetRow = firstChanged - 1
	}

	if moveTargetRow > prevViewportBottom {
		t.scrollViewport(ctx, builder, moveTargetRow, prevViewportBottom)
	}

	t.moveCursorToRow(ctx, builder, moveTargetRow)

	if appendStart {
		builder.WriteString("\r\n")
	} else {
		builder.WriteString("\r")
	}
}

func (t *TUI) scrollViewport(ctx renderContext, builder *strings.Builder, moveTargetRow, prevViewportBottom int) {
	currentScreenRow := max(0, min(ctx.height-1, ctx.hardwareCursorRaw-ctx.prevViewportTop))
	moveToBottom := ctx.height - 1 - currentScreenRow
	if moveToBottom > 0 {
		builder.WriteString(fmt.Sprintf("\x1b[%dB", moveToBottom))
	}
	scroll := moveTargetRow - prevViewportBottom
	for range scroll {
		builder.WriteString("\r\n")
	}
	ctx.prevViewportTop += scroll
	ctx.viewportTop += scroll
	ctx.hardwareCursorRaw = moveTargetRow
}

func (t *TUI) renderChangedLines(ctx renderContext, builder *strings.Builder, newLines []string, firstChanged, lastChanged int) {
	renderEnd := min(lastChanged, len(newLines)-1)
	for i := firstChanged; i <= renderEnd; i++ {
		if i > firstChanged {
			builder.WriteString("\r\n")
		}
		builder.WriteString("\x1b[2K")
		line := newLines[i]
		if VisibleWidth(line) > ctx.width {
			panic(fmt.Sprintf("Rendered line %d exceeds terminal width (%d > %d)", i, VisibleWidth(line), ctx.width))
		}
		builder.WriteString(line)
	}
}

func (t *TUI) handleTrailingLines(builder *strings.Builder, newLines []string, lastChanged int) int {
	finalCursorRow := lastChanged
	if len(t.previousLines) > len(newLines) {
		renderEnd := min(lastChanged, len(newLines)-1)
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
	return finalCursorRow
}

func (t *TUI) updateRenderStateAfterIncremental(ctx renderContext, newLines []string, finalCursorRow, row, col int) {
	t.cursorRow = max(0, len(newLines)-1)
	t.hardwareCursorRow = finalCursorRow
	t.maxLinesRendered = max(t.maxLinesRendered, len(newLines))
	t.previousViewportTop = max(0, t.maxLinesRendered-ctx.height)
	t.positionHardwareCursor(row, col, len(newLines))
	t.previousLines = newLines
	t.previousWidth = ctx.width
}

func (t *TUI) getScope(newLines []string) (int, int) {
	var (
		firstChanged = -1
		lastChanged  = -1
	)
	maxLines := max(len(newLines), len(t.previousLines))
	for i := range maxLines {
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

func (t *TUI) fullRender(clear bool, newLines []string, width int, height int) {
	t.fullRedrawCount++
	buffer := t.buildFullRenderBuffer(clear, newLines)
	t.terminal.Write(buffer)
	t.updateRenderState(clear, len(newLines), width, height)
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
