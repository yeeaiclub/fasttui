package components

type Text struct {
}

func (t *Text) Render(width int) []string {
	panic("implement me")
}

func (t *Text) HandleInput(data string) {
	panic("implement me")
}

func (t *Text) WantsKeyRelease() bool {
	panic("implement me")
}

func (t *Text) Invalidate() {
	panic("implement me")
}
