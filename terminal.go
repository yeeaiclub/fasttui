package fasttui

import (
	"os"

	"golang.org/x/term"
)

// Terminal 终端接口定义
type Terminal interface {
	// Start 启动终端并设置输入和调整大小的处理器
	// @param onInput - 输入处理函数
	// @param onResize - 调整大小处理函数
	Start(onInput func(data string), onResize func())

	// Stop 停止终端并恢复状态
	Stop()

	// Write 向终端写入输出
	// @param data - 要写入的数据
	Write(data string)

	// GetColumns 获取终端列数
	GetColumns() int

	// GetRows 获取终端行数
	GetRows() int

	// IsKittyProtocolActive 检查 Kitty 键盘协议是否激活
	IsKittyProtocolActive() bool

	// MoveBy 相对当前位置移动光标
	// @param lines - 向上（负数）或向下（正数）移动的行数
	MoveBy(lines int)

	// HideCursor 隐藏光标
	HideCursor()

	// ShowCursor 显示光标
	ShowCursor()

	// ClearLine 清除当前行
	ClearLine()

	// ClearFromCursor 清除从光标到屏幕末尾的内容
	ClearFromCursor()

	// ClearScreen 清除整个屏幕并将光标移动到 (0,0)
	ClearScreen()

	// SetTitle 设置终端窗口标题
	// @param title - 要设置的标题
	SetTitle(title string)
}

type ProcessTerminal struct {
	wasRaw        bool
	writeLogPath  string
	inputHandler  func(data string)
	resizeHandler func()
	stdinFd       int
}

func NewProcessTerminal() (*ProcessTerminal, error) {
	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		return &ProcessTerminal{
			stdinFd: fd,
		}, nil
	}
	return &ProcessTerminal{wasRaw: false}, nil
}

func (p *ProcessTerminal) Start(onInput func(data string), onResize func()) {
	p.inputHandler = onInput
	p.resizeHandler = onResize
}

func (p *ProcessTerminal) Stop() {
	os.Stdout.WriteString(`\x1b[?2004l`)
	p.inputHandler = nil
	p.resizeHandler = nil
}

func (p *ProcessTerminal) setupStdinBuffer() {
}

func (p *ProcessTerminal) queryAndEnableKittyProtocol() {
}

func (p *ProcessTerminal) Write(data string) {
	_, err := os.Stdout.WriteString(data)
	if err != nil {
		return
	}
}

func (p *ProcessTerminal) MoveBy(lines int) {
	if lines > 0 {
		os.Stdout.WriteString(`\x1b[${lines}B`)
	} else if lines < 0 {
		// Move up
		os.Stdout.WriteString(`\x1b[${-lines}A`)
	}
	// lines === 0: no movement
}

func (p *ProcessTerminal) HideCursor() {
	os.Stdout.WriteString(`\x1b[?25l`)
}

func (p *ProcessTerminal) ShowCursor() {
	os.Stdout.WriteString(`\x1b[?25h`)
}

func (p *ProcessTerminal) ClearLine() {
	os.Stdout.WriteString(`\x1b[2K`)
}

func (p *ProcessTerminal) ClearFromCursor() {
	os.Stdout.WriteString(`\x1b[J`)
}

func (p *ProcessTerminal) ClearScreen() {
	os.Stdout.WriteString(`\x1b[2J\x1b[H`) // Clear screen and move to home (1,1)
}

func (p *ProcessTerminal) SetTitle(title string) {
	// OSC 0;title BEL - set terminal window title
	os.Stdout.WriteString(`\x1b]0;` + title + `\x07`)
}
