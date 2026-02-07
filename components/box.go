package components

import "github.com/yeeaiclub/fasttui"

type Box struct {
	children []fasttui.Component
}

type RenderCache struct {
	childLines []string
	width      int
	bgSample   string
	lines      string
}

func NewBox() *Box {
	return &Box{}
}

func (b *Box) Render(width int) []string {
	panic("implement me")
}

func (b *Box) HandleInput(data string) {
	panic("implement me")
}

func (b *Box) WantsKeyRelease() bool {
	panic("implement me")
}

func (b *Box) Invalidate() {
	panic("implement me")
}
