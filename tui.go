package fasttui

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func (t *TUI) RequestRender(force bool) {
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

// requestRender is an internal alias for backward compatibility
func (t *TUI) requestRender(force bool) {
	t.RequestRender(force)
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

	width, height := t.terminal.GetSize()
	viewportTop := max(0, t.maxLinesRendered-height)
	prevViewportTop := t.previousViewportTop
	hardwareCursorRow := t.hardwareCursorRow

	computeLineDiff := func(targetRow int) int {
		cs := hardwareCursorRow - prevViewportTop
		ct := targetRow - viewportTop
		return ct - cs
	}
	// render all components in container
	newLines := t.Render(width)
	if len(t.overlayStacks) > 0 {
		newLines = t.compositeOverlays(newLines, width, height)
	}

	row, col := t.extractCursorPosition(newLines, height)
	newLines = t.applyLineRests(newLines)
	widthChanged := t.previousWidth != 0 && t.previousWidth != width

	fullRender := t.fullRender(newLines, height, row, col, width)

	if len(t.previousLines) == 0 && !widthChanged {
		fullRender(false)
		return
	}

	if widthChanged {
		fullRender(true)
		return
	}

	// Find first and last changed lines
	firstChanged := -1
	lastChanged := -1
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

	appendedLines := len(newLines) > len(t.previousLines)
	if appendedLines {
		if firstChanged == -1 {
			firstChanged = len(t.previousLines)
		}
		lastChanged = len(newLines) - 1
	}

	appendStart := appendedLines && firstChanged == len(t.previousLines) && firstChanged > 0

	// No changes - but still need to update hardware cursor position if it moved
	if firstChanged == -1 {
		t.positionHardwareCursor(row, col, len(newLines))
		t.previousViewportTop = max(0, t.maxLinesRendered-height)
		return
	}

	// All changes are in deleted lines (nothing to render, just clear)
	if firstChanged >= len(newLines) {
		if len(t.previousLines) > len(newLines) {
			var buffer strings.Builder
			buffer.WriteString("\x1b[?2026h")

			// Move to end of new content (clamp to 0 for empty content)
			targetRow := max(0, len(newLines)-1)
			lineDiff := computeLineDiff(targetRow)
			if lineDiff > 0 {
				buffer.WriteString("\x1b[")
				buffer.WriteString(strconv.Itoa(lineDiff))
				buffer.WriteString("B")
			} else if lineDiff < 0 {
				buffer.WriteString("\x1b[")
				buffer.WriteString(strconv.Itoa(-lineDiff))
				buffer.WriteString("A")
			}
			buffer.WriteString("\r")

			// Clear extra lines without scrolling
			extraLines := len(t.previousLines) - len(newLines)
			if extraLines > height {
				fullRender(true)
				return
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

			buffer.WriteString("\x1b[?2026l")
			t.terminal.Write(buffer.String())
			t.cursorRow = targetRow
			t.hardwareCursorRow = targetRow
		}

		t.positionHardwareCursor(row, col, len(newLines))
		t.previousLines = newLines
		t.previousWidth = width
		t.previousViewportTop = max(0, t.maxLinesRendered-height)
		return
	}

	// Check if firstChanged is outside the viewport
	// Viewport is based on max lines ever rendered (terminal's working area)
	if firstChanged < viewportTop {
		// First change is above viewport - need full re-render
		fullRender(true)
		return
	}

	// Render from first changed line to end
	// Build buffer with all updates wrapped in synchronized output
	var buffer strings.Builder
	buffer.WriteString("\x1b[?2026h") // Begin synchronized output

	prevViewportBottom := prevViewportTop + height - 1
	moveTargetRow := firstChanged
	if appendStart {
		moveTargetRow = firstChanged - 1
	}

	if moveTargetRow > prevViewportBottom {
		currentScreenRow := max(0, min(height-1, hardwareCursorRow-prevViewportTop))
		moveToBottom := height - 1 - currentScreenRow
		if moveToBottom > 0 {
			buffer.WriteString("\x1b[")
			buffer.WriteString(strconv.Itoa(moveToBottom))
			buffer.WriteString("B")
		}

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
	renderEnd := min(lastChanged, len(newLines)-1)
	for i := firstChanged; i <= renderEnd; i++ {
		if i > firstChanged {
			buffer.WriteString("\r\n")
		}
		buffer.WriteString("\x1b[2K") // Clear current line

		line := newLines[i]
		isImageLine := t.containsImage(line)
		if !isImageLine && VisibleWidth(line) > width {
			// Log all lines to crash file for debugging
			crashLogPath := t.getCrashLogPath()
			var crashData strings.Builder
			crashData.WriteString("Crash at ")
			crashData.WriteString(time.Now().Format(time.RFC3339))
			crashData.WriteString("\n")
			crashData.WriteString("Terminal width: ")
			crashData.WriteString(strconv.Itoa(width))
			crashData.WriteString("\n")
			crashData.WriteString("Line ")
			crashData.WriteString(strconv.Itoa(i))
			crashData.WriteString(" visible width: ")
			crashData.WriteString(strconv.Itoa(VisibleWidth(line)))
			crashData.WriteString("\n\n")
			crashData.WriteString("=== All rendered lines ===\n")
			for idx, l := range newLines {
				crashData.WriteString("[")
				crashData.WriteString(strconv.Itoa(idx))
				crashData.WriteString("] (w=")
				crashData.WriteString(strconv.Itoa(VisibleWidth(l)))
				crashData.WriteString(") ")
				crashData.WriteString(l)
				crashData.WriteString("\n")
			}

			t.writeCrashLog(crashLogPath, crashData.String())

			// Clean up terminal state before panicking
			t.Stop()

			var errorMsg strings.Builder
			errorMsg.WriteString("Rendered line ")
			errorMsg.WriteString(strconv.Itoa(i))
			errorMsg.WriteString(" exceeds terminal width (")
			errorMsg.WriteString(strconv.Itoa(VisibleWidth(line)))
			errorMsg.WriteString(" > ")
			errorMsg.WriteString(strconv.Itoa(width))
			errorMsg.WriteString(").\n\n")
			errorMsg.WriteString("This is likely caused by a custom TUI component not truncating its output.\n")
			errorMsg.WriteString("Use VisibleWidth() to measure and truncate lines.\n\n")
			errorMsg.WriteString("Debug log written to: ")
			errorMsg.WriteString(crashLogPath)

			panic(errorMsg.String())
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

	buffer.WriteString("\x1b[?2026l") // End synchronized output

	// Debug logging if enabled
	if os.Getenv("PI_TUI_DEBUG") == "1" {
		t.writeDebugLog(firstChanged, viewportTop, finalCursorRow, hardwareCursorRow,
			renderEnd, row, col, height, newLines)
	}

	// Write entire buffer at once
	t.terminal.Write(buffer.String())

	// Track cursor position for next render
	// cursorRow tracks end of content (for viewport calculation)
	// hardwareCursorRow tracks actual terminal cursor position (for movement)
	t.cursorRow = max(0, len(newLines)-1)
	t.hardwareCursorRow = finalCursorRow

	// Track terminal's working area (grows but doesn't shrink unless cleared)
	t.maxLinesRendered = max(t.maxLinesRendered, len(newLines))
	t.previousViewportTop = max(0, t.maxLinesRendered-height)

	// Position hardware cursor for IME
	t.positionHardwareCursor(row, col, len(newLines))
	t.previousLines = newLines
	t.previousWidth = width
}

func (t *TUI) fullRender(newLines []string, height int, row int, col int, width int) func(clear bool) {
	fullRender := func(clear bool) {
		t.fullRedrawCount++
		var buffer strings.Builder
		buffer.WriteString("\x1b[?2026h") // Begin synchronized output
		if clear {
			buffer.WriteString("\x1b[3J\x1b[2J\x1b[H") // Clear scrollback, screen, and home
		}
		for i := 0; i < len(newLines); i++ {
			if i > 0 {
				buffer.WriteString("\r\n")
			}
			buffer.WriteString(newLines[i])
		}
		buffer.WriteString("\x1b[?2026l") // End synchronized output
		t.terminal.Write(buffer.String())
		t.cursorRow = max(0, len(newLines)-1)
		t.hardwareCursorRow = t.cursorRow
		// Reset max lines when clearing, otherwise track growth
		if clear {
			t.maxLinesRendered = len(newLines)
		} else {
			t.maxLinesRendered = max(t.maxLinesRendered, len(newLines))
		}
		t.previousViewportTop = max(0, t.maxLinesRendered-height)
		t.positionHardwareCursor(row, col, len(newLines))
		t.previousLines = newLines
		t.previousWidth = width
	}
	return fullRender
}

func (t *TUI) writeDebugLog(firstChanged, viewportTop, finalCursorRow, hardwareCursorRow,
	renderEnd, cursorRow, cursorCol, height int, newLines []string) {
	debugDir := "/tmp/tui"
	os.MkdirAll(debugDir, 0755)

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	debugPath := filepath.Join(debugDir, "render-"+strconv.FormatInt(timestamp, 10)+".log")

	var debugData strings.Builder
	debugData.WriteString("firstChanged: ")
	debugData.WriteString(strconv.Itoa(firstChanged))
	debugData.WriteString("\nviewportTop: ")
	debugData.WriteString(strconv.Itoa(viewportTop))
	debugData.WriteString("\ncursorRow: ")
	debugData.WriteString(strconv.Itoa(t.cursorRow))
	debugData.WriteString("\nheight: ")
	debugData.WriteString(strconv.Itoa(height))
	debugData.WriteString("\nhardwareCursorRow: ")
	debugData.WriteString(strconv.Itoa(hardwareCursorRow))
	debugData.WriteString("\nrenderEnd: ")
	debugData.WriteString(strconv.Itoa(renderEnd))
	debugData.WriteString("\nfinalCursorRow: ")
	debugData.WriteString(strconv.Itoa(finalCursorRow))
	debugData.WriteString("\ncursorPos: row=")
	debugData.WriteString(strconv.Itoa(cursorRow))
	debugData.WriteString(" col=")
	debugData.WriteString(strconv.Itoa(cursorCol))
	debugData.WriteString("\nnewLines.length: ")
	debugData.WriteString(strconv.Itoa(len(newLines)))
	debugData.WriteString("\npreviousLines.length: ")
	debugData.WriteString(strconv.Itoa(len(t.previousLines)))
	debugData.WriteString("\n\n=== newLines ===\n")
	for i, line := range newLines {
		debugData.WriteString("[")
		debugData.WriteString(strconv.Itoa(i))
		debugData.WriteString("] ")
		debugData.WriteString(line)
		debugData.WriteString("\n")
	}
	debugData.WriteString("\n=== previousLines ===\n")
	for i, line := range t.previousLines {
		debugData.WriteString("[")
		debugData.WriteString(strconv.Itoa(i))
		debugData.WriteString("] ")
		debugData.WriteString(line)
		debugData.WriteString("\n")
	}

	os.WriteFile(debugPath, []byte(debugData.String()), 0644)
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
	if t.containsImage(baseLine) {
		return baseLine
	}

	// Single pass through baseLine extracts both before and after segments
	afterStart := col + overlayWidth
	before, beforeWidth, after, afterWidth := ExtractSegments(baseLine, col, afterStart, termWidth-afterStart, true)

	// Extract overlay with width tracking (strict=true to exclude wide chars at boundary)
	overlay := SliceWithWidth(overlayLine, 0, overlayWidth, true)

	// Pad segments to target widths
	beforePad := max(0, col-beforeWidth)
	overlayPad := max(0, overlayWidth-overlay.width)
	actualBeforeWidth := max(col, beforeWidth)
	actualOverlayWidth := max(overlayWidth, overlay.width)
	afterTarget := max(0, termWidth-actualBeforeWidth-actualOverlayWidth)
	afterPad := max(0, afterTarget-afterWidth)

	// Compose result
	var result strings.Builder
	result.WriteString(before)
	result.WriteString(strings.Repeat(" ", beforePad))
	result.WriteString(SEGMENT_RESET)
	result.WriteString(overlay.text)
	result.WriteString(strings.Repeat(" ", overlayPad))
	result.WriteString(SEGMENT_RESET)
	result.WriteString(after)
	result.WriteString(strings.Repeat(" ", afterPad))

	// CRITICAL: Always verify and truncate to terminal width.
	// This is the final safeguard against width overflow which would crash the TUI.
	resultStr := result.String()
	resultWidth := VisibleWidth(resultStr)
	if resultWidth <= termWidth {
		return resultStr
	}
	// Truncate with strict=true to ensure we don't exceed termWidth
	return SliceByColumn(resultStr, 0, termWidth, true)
}

var CURSOR_MARKER = "\x1b_pi:c\x07"

var SEGMENT_RESET = "\x1b[0m\x1b]8;;\x07"

func (t *TUI) containsImage(line string) bool {
	return strings.Contains(line, "\x1b_G") || strings.Contains(line, "\x1b]1337;File=")
}

func (t *TUI) getCrashLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".pi", "agent", "pi-crash.log")
}

func (t *TUI) writeCrashLog(path string, data string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(path, []byte(data), 0644)
}

func (t *TUI) applyLineRests(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		if t.containsImage(line) {
			result[i] = line
		} else {
			result[i] = line + SEGMENT_RESET
		}
	}
	return result
}
