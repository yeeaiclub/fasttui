package fasttui

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
	terminal            Terminal
	previousLines       []string
	previousWidth       int
	focusedComponent    Component
	stopped             bool
	showHardWareCursor  bool
	cursorRow           int
	hardwareCursorRaw   int
	maxLinesRendered    int
	previousViewportTop int
	renderRequested     bool
	renderChan          chan struct{}
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
		t.hardwareCursorRaw = 0
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
	width, height := t.terminal.GetSize()
	viewportTop := max(0, t.maxLinesRendered-height)
	prevViewportTop := t.previousViewportTop
	hardwareCursorRaw := t.hardwareCursorRaw
	computeLineDiff := func(targetRow int) int {
		cs := hardwareCursorRaw - prevViewportTop
		ts := targetRow - viewportTop
		return ts - cs
	}
	// 表示渲染了多少行
	newLines := t.Render(width)
	cursorPos, _ := t.extractCursorPosition(newLines, height)

}
func (t *TUI) detectChangeType() {

}

func (t *TUI) applyLineRests(newLines []string) {

}

func (t *TUI) showOverlay(component Component) {

}

func (t *TUI) hideOverlay() {

}

const CURSOR_MARKER = "\x1b_pi:c\x07"

func (t *TUI) extractCursorPosition(lines []string, height int) (int, int) {
	//viewportTop := len(lines) - height
	return 0, 0
}
