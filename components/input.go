package components

type Input struct {
	value    string
	cursor   int
	onSubmit func(value string)
	onEscape func()
	focused  bool
}

func (i *Input) Render(width int) []string {
	const prompt = "> "
	return []string{prompt + i.value}
}

func (i *Input) HandleInput(data string) {
	panic("implement me")
}

func (i *Input) WantsKeyRelease() bool {
	panic("implement me")
}

func (i *Input) Invalidate() {
	panic("implement me")
}
