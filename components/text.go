package components

type Text struct {
	text               string
	paddingX, paddingY string
	customBgFn         func(text string) string
}

func (t *Text) Render(width int) []string {
	return nil
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
