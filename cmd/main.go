package main

import (
	"time"

	"github.com/yeeaiclub/fasttui"
)

type SimpleText struct {
	content string
}

func (s *SimpleText) Render(width int) []string {
	return []string{s.content}
}

func (s *SimpleText) HandleInput(data string) {}

func (s *SimpleText) WantsKeyRelease() bool {
	return false
}

func (s *SimpleText) Invalidate() {}

func NewSimpleText(content string) *SimpleText {
	return &SimpleText{content: content}
}

func main() {
	term, _ := fasttui.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)

	tui.AddChild(NewSimpleText("Hello, FastTUI!"))
	tui.AddChild(NewSimpleText("This is a simple demo."))
	tui.AddChild(NewSimpleText("Line 3: Testing rendering."))
	tui.AddChild(NewSimpleText("Line 4: Incremental updates."))

	tui.Start()

	time.Sleep(3 * time.Second)
}
