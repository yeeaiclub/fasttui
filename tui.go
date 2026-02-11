package fasttui

import (
	"regexp"
	"strconv"
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
	inputBuffer          strings.Builder

	overlayStacks []Overlay
	clearOnShrink bool
}

func NewTUI(terminal Terminal, showHardwareCursor bool) *TUI {
	t := &TUI{
		renderChan:         make(chan struct{}, 1),
		overlayStacks:      make([]Overlay, 0),
		terminal:           terminal,
		showHardwareCursor: showHardwareCursor,
		previousLines:      nil,
	}
	return t
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
		t.inputBuffer.WriteString(data)
		filtered := t.parseCellSizeResponse()
		if filtered == "" {
			return
		}
		data = filtered
	}

	var focusedOverlay *Overlay
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

func (t *TUI) parseCellSizeResponse() string {
	data := t.inputBuffer.String()

	responsePattern := `\x1b\[6;(\d+);(\d+)t`
	re := regexp.MustCompile(responsePattern)
	matches := re.FindStringSubmatch(data)

	if len(matches) == 3 {
		heightPx, err1 := strconv.Atoi(matches[1])
		widthPx, err2 := strconv.Atoi(matches[2])

		if err1 == nil && err2 == nil && heightPx > 0 && widthPx > 0 {
			t.Invalidate()
			t.requestRender(false)

			t.inputBuffer.Reset()
			t.cellSizeQueryPending = false
			return ""
		}
	}

	partialPattern := `\x1b(\[6?;?[\d;]*)?$`
	rePartial := regexp.MustCompile(partialPattern)
	if rePartial.MatchString(data) {
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

func (t *TUI) doRender() {
	if t.stopped {
		return
	}

	// width, height := t.terminal.GetSize()
	// viewportTop := max(0, t.maxLinesRendered-height)
	// prevViewportTop := t.previousViewportTop
	// hardwareCursorRow := t.hardwareCursorRow

	// computeLineDiff := func(targetRow int) int {
	// 	cs := hardwareCursorRow - prevViewportTop
	// 	ct := targetRow - viewportTop
	// 	return ct - cs
	// }
	// // render all components in container
	// newLines := t.Render(width)
	// if len(t.overlayStacks) > 0 {
	// 	newLines = t.compositeOverlays(newLines, width, height)
	// }
}

func (t *TUI) compositeOverlays(newLines []string, width, height int) []string {
	if len(t.overlayStacks) == 0 {
		return newLines
	}

	result := make([]string, len(newLines))
	copy(result, newLines)

	type renderedOverlay struct {
		overlayLines []string
		row          int
		col          int
		w            int
	}

	rendered := make([]renderedOverlay, 0)
	minLinesNeeded := len(result)

	for i := range t.overlayStacks {
		entry := &t.overlayStacks[i]
		if !t.isOverlayVisible(entry) {
			continue
		}

		options := entry.options

		layout := t.ResolveOverlayLayout(options, 0, width, height)
		overlayWidth := layout.width
		component := entry.component
		overlayLines := component.Render(overlayWidth)

		if layout.maxHeight != nil && len(overlayLines) > *layout.maxHeight {
			overlayLines = overlayLines[:*layout.maxHeight]
		}

		finalLayout := t.ResolveOverlayLayout(options, len(overlayLines), width, height)

		rendered = append(rendered, renderedOverlay{
			overlayLines: overlayLines,
			row:          finalLayout.row,
			col:          finalLayout.col,
			w:            finalLayout.width,
		})

		if finalLayout.row+len(overlayLines) > minLinesNeeded {
			minLinesNeeded = finalLayout.row + len(overlayLines)
		}
	}

	workingHeight := max(t.maxLinesRendered, minLinesNeeded)

	for len(result) < workingHeight {
		result = append(result, "")
	}

	viewportStart := max(0, workingHeight-height)

	modifiedLines := make(map[int]bool)

	for _, ro := range rendered {
		for i := range ro.overlayLines {
			idx := viewportStart + ro.row + i
			if idx >= 0 && idx < len(result) {
				truncatedOverlayLine := ro.overlayLines[i]
				if VisibleWidth(ro.overlayLines[i]) > ro.w {
					truncatedOverlayLine = SliceByColumn(ro.overlayLines[i], 0, ro.w, true)
				}
				result[idx] = t.compositeLineAt(result[idx], truncatedOverlayLine, ro.col, ro.w, width)
				modifiedLines[idx] = true
			}
		}
	}

	for idx := range modifiedLines {
		if VisibleWidth(result[idx]) > width {
			result[idx] = SliceByColumn(result[idx], 0, width, true)
		}
	}

	return result
}

func (t *TUI) compositeLineAt(baseLine string, overlayLine string, col int, overlayWidth int, termWidth int) string {
	if col >= termWidth {
		return baseLine
	}

	baseWidth := VisibleWidth(baseLine)
	if baseWidth < col {
		padding := strings.Repeat(" ", col-baseWidth)
		return baseLine + padding + overlayLine
	}

	before, _, after, _ := ExtractSegments(baseLine, col, col+overlayWidth, overlayWidth, false)
	return before + overlayLine + after
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
