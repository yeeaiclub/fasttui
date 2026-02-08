package main

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
}
