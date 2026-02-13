package fasttui

// TUI is the main coordinator that delegates to specialized managers
type TUIRefactored struct {
	Container
	stopped  bool
	terminal Terminal

	renderRequested bool
	renderChan      chan struct{}

	// Specialized managers
	renderState    *RenderState
	cursorManager  *CursorManager
	overlayManager *OverlayManager
	focusManager   *FocusManager
	inputHandler   *InputHandler
	layoutResolver OverlayLayoutResolver

	clearOnShrink bool
}

func NewTUIRefactored(terminal Terminal, showHardwareCursor bool) *TUIRefactored {
	t := &TUIRefactored{
		renderChan:     make(chan struct{}, 1),
		terminal:       terminal,
		renderState:    newRenderState(),
		cursorManager:  newCursorManager(terminal, showHardwareCursor),
		overlayManager: newOverlayManager(terminal),
		focusManager:   newFocusManager(),
		layoutResolver: &DefaultLayoutResolver{},
	}

	t.inputHandler = newInputHandler(terminal, func() {
		t.Invalidate()
		t.RequestRender(false)
	})

	return t
}

func (t *TUIRefactored) Start() {
	t.start()
	go t.renderLoop()
	t.doRender()
}

func (t *TUIRefactored) Stop() {
	t.stopped = true
	t.terminal.Stop()
}

func (t *TUIRefactored) renderLoop() {
	for range t.renderChan {
		t.renderRequested = false
		t.doRender()
	}
}

func (t *TUIRefactored) RequestRender(force bool) {
	if force {
		t.renderState.Reset()
		t.cursorManager.Reset()
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

func (t *TUIRefactored) doRender() {
	if t.stopped {
		return
	}

	width, height := t.terminal.GetSize()

	// Render base content
	newLines := t.Render(width)

	// Composite overlays if any
	if t.overlayManager.HasVisible() {
		newLines = t.overlayManager.Composite(
			newLines,
			width,
			height,
			t.renderState.maxLinesRendered,
			t.layoutResolver,
		)
	}

	// Extract and position cursor
	row, col := extractCursorPosition(newLines, height)
	newLines = applyLineRests(newLines)

	// Check if width changed
	widthChanged := t.renderState.previousWidth != 0 && t.renderState.previousWidth != width

	// Determine render strategy
	if len(t.renderState.previousLines) == 0 || widthChanged {
		t.fullRender(newLines, width, height, row, col, widthChanged)
		return
	}

	// Incremental render
	t.incrementalRender(newLines, width, height, row, col)
}

func (t *TUIRefactored) fullRender(newLines []string, width, height, row, col int, clear bool) {
	t.renderState.fullRedrawCount++

	buffer := NewRenderBuffer()
	buffer.BeginSync()

	if clear {
		buffer.ClearAll()
	}

	for i := range newLines {
		if i > 0 {
			buffer.NewLine()
		}
		buffer.Write(newLines[i])
	}

	buffer.EndSync()
	t.terminal.Write(buffer.String())

	t.cursorManager.UpdatePosition(max(0, len(newLines)-1))

	if clear {
		t.renderState.UpdateAfterClear(newLines, height)
	} else {
		t.renderState.UpdateAfterRender(newLines, width, height)
	}

	t.cursorManager.Position(row, col, len(newLines))
	t.renderState.previousWidth = width
}

func (t *TUIRefactored) incrementalRender(newLines []string, width, height, row, col int) {
	viewportTop := t.renderState.GetViewportTop(height)
	firstChanged, lastChanged := t.renderState.FindChangedLineRange(newLines)

	// Handle line appends
	appendedLines := len(newLines) > len(t.renderState.previousLines)
	if appendedLines {
		if firstChanged == -1 {
			firstChanged = len(t.renderState.previousLines)
		}
		lastChanged = len(newLines) - 1
	}

	// No changes - just update cursor
	if firstChanged == -1 {
		t.cursorManager.Position(row, col, len(newLines))
		t.renderState.previousViewportTop = viewportTop
		return
	}

	// All changes in deleted lines
	if firstChanged >= len(newLines) {
		t.handleDeletedLines(newLines, height, row, col, viewportTop)
		return
	}

	// Changes outside viewport - full render
	if firstChanged < viewportTop {
		t.fullRender(newLines, width, height, row, col, true)
		return
	}

	// Render changed lines
	t.renderChangedLines(newLines, width, height, firstChanged, lastChanged, appendedLines, viewportTop, row, col)
}

func (t *TUIRefactored) handleDeletedLines(newLines []string, height, row, col, viewportTop int) {
	if len(t.renderState.previousLines) <= len(newLines) {
		t.cursorManager.Position(row, col, len(newLines))
		t.renderState.UpdateAfterRender(newLines, t.renderState.previousWidth, height)
		return
	}

	buffer := NewRenderBuffer()
	buffer.BeginSync()

	// Move to end of new content
	targetRow := max(0, len(newLines)-1)
	buffer.MoveCursor(t.cursorManager.hardwareCursorRow, targetRow, viewportTop, viewportTop)
	buffer.CarriageReturn()

	// Clear extra lines
	extraLines := len(t.renderState.previousLines) - len(newLines)
	if extraLines > height {
		t.fullRender(newLines, t.renderState.previousWidth, height, row, col, true)
		return
	}

	buffer.ClearExtraLines(extraLines)
	buffer.EndSync()

	t.terminal.Write(buffer.String())
	t.cursorManager.UpdatePosition(targetRow)
	t.cursorManager.Position(row, col, len(newLines))
	t.renderState.UpdateAfterRender(newLines, t.renderState.previousWidth, height)
}

func (t *TUIRefactored) renderChangedLines(newLines []string, width, height, firstChanged, lastChanged int, appendStart bool, viewportTop, row, col int) {
	buffer := NewRenderBuffer()
	buffer.BeginSync()

	// Move to first changed line
	moveTargetRow := firstChanged
	if appendStart {
		moveTargetRow = firstChanged - 1
	}

	prevViewportTop := t.renderState.previousViewportTop
	prevViewportBottom := prevViewportTop + height - 1

	if moveTargetRow > prevViewportBottom {
		currentScreenRow := max(0, min(height-1, t.cursorManager.hardwareCursorRow-prevViewportTop))
		moveToBottom := height - 1 - currentScreenRow
		if moveToBottom > 0 {
			buffer.MoveDown(moveToBottom)
		}

		scroll := moveTargetRow - prevViewportBottom
		buffer.Scroll(scroll)
		prevViewportTop += scroll
		viewportTop += scroll
		t.cursorManager.SetHardwareRow(moveTargetRow)
	}

	buffer.MoveCursor(t.cursorManager.hardwareCursorRow, moveTargetRow, prevViewportTop, viewportTop)

	if appendStart {
		buffer.NewLine()
	} else {
		buffer.CarriageReturn()
	}

	// Render changed lines with width validation
	renderEnd := min(lastChanged, len(newLines)-1)
	for i := firstChanged; i <= renderEnd; i++ {
		if i > firstChanged {
			buffer.NewLine()
		}
		buffer.ClearLine()

		line := newLines[i]
		if !containsImage(line) && VisibleWidth(line) > width {
			t.handleWidthOverflow(line, i, newLines, width)
		}
		buffer.Write(line)
	}

	finalCursorRow := renderEnd

	// Clear extra lines if content shrunk
	if len(t.renderState.previousLines) > len(newLines) {
		if renderEnd < len(newLines)-1 {
			moveDown := len(newLines) - 1 - renderEnd
			buffer.MoveDown(moveDown)
			finalCursorRow = len(newLines) - 1
		}

		extraLines := len(t.renderState.previousLines) - len(newLines)
		for i := len(newLines); i < len(t.renderState.previousLines); i++ {
			buffer.NewLine()
			buffer.ClearLine()
		}
		buffer.MoveUp(extraLines)
	}

	buffer.EndSync()
	t.terminal.Write(buffer.String())

	t.cursorManager.cursorRow = max(0, len(newLines)-1)
	t.cursorManager.hardwareCursorRow = finalCursorRow
	t.renderState.UpdateAfterRender(newLines, width, height)
	t.cursorManager.Position(row, col, len(newLines))
}

func (t *TUIRefactored) start() error {
	t.stopped = false
	return t.terminal.Start(
		func(data string) {
			t.HandleInput(data)
		},
		func() {
			t.RequestRender(false)
		},
	)
}

func (t *TUIRefactored) SetFocus(component Component) {
	t.focusManager.SetFocus(component)
}

func (t *TUIRefactored) HandleInput(data string) {
	focusedComponent := t.focusManager.GetFocused()

	// Check if focused overlay is still visible
	if focusedComponent != nil {
		topVisible := t.overlayManager.GetTopmostVisibleComponent()
		if topVisible != nil && topVisible != focusedComponent {
			t.SetFocus(topVisible)
			focusedComponent = topVisible
		}
	}

	data = t.inputHandler.ProcessInput(data, focusedComponent)
	if data == "" {
		return
	}

	if focusedComponent != nil {
		focusedComponent.HandleInput(data)
		t.RequestRender(false)
	}
}

func (t *TUIRefactored) ShowOverlay(component Component, options OverlayOption) (func(), func(bool), func() bool) {
	return t.overlayManager.Show(component, options, t.focusManager.GetFocused(), func(c Component) {
		t.SetFocus(c)
		t.RequestRender(false)
	})
}

func (t *TUIRefactored) HideOverlay() {
	t.overlayManager.Hide(func(c Component) {
		t.SetFocus(c)
		t.RequestRender(false)
	})
}

func (t *TUIRefactored) HasOverlay() bool {
	return t.overlayManager.HasVisible()
}

func (t *TUIRefactored) SetShowHardwareCursor(enabled bool) {
	t.cursorManager.SetShowHardwareCursor(enabled)
	t.RequestRender(false)
}

func (t *TUIRefactored) SetClearOnShrink(enabled bool) {
	t.clearOnShrink = enabled
}

func (t *TUIRefactored) QueryCellSize() {
	t.inputHandler.QueryCellSize()
}

func (t *TUIRefactored) GetFullRedraws() int {
	return t.renderState.fullRedrawCount
}

func (t *TUIRefactored) GetShowHardwareCursor() bool {
	return t.cursorManager.showHardwareCursor
}

func (t *TUIRefactored) handleWidthOverflow(line string, lineIndex int, allLines []string, width int) {
	crashLogPath := t.getCrashLogPath()
	crashData := formatCrashLog(line, lineIndex, allLines, width)
	t.writeCrashLog(crashLogPath, crashData)
	t.Stop()
	panic(formatWidthError(lineIndex, VisibleWidth(line), width, crashLogPath))
}
