package components

type Input struct {
	value    string
	cursor   int
	onSubmit func(string)
	onEscape func()
	focused  bool
}

func (i *Input) Render(width int) []string {
	return []string{"input"}
}
func (i *Input) HandleInput(data string) {
}
func (i *Input) WantsKeyRelease() bool {
	return false
}
func (i *Input) Invalidate() {
}
