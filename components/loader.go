package components

type Loader struct {
}

func (l *Loader) Render(width int) []string {
	panic("implement me")
}

func (l *Loader) HandleInput(data string) {
	panic("implement me")
}

func (l *Loader) WantsKeyRelease() bool {
	panic("implement me")
}

func (l *Loader) Invalidate() {
	panic("implement me")
}
