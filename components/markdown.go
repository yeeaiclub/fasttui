package components

type Markdown struct {
}

func (m Markdown) Render(width int) []string {
	panic("implement me")
}

func (m Markdown) HandleInput(data string) {
	panic("implement me")
}

func (m Markdown) WantsKeyRelease() bool {
	panic("implement me")
}

func (m Markdown) Invalidate() {
	panic("implement me")
}
