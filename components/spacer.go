package components

type Spacer struct {
	lines int
}

func NewSpacer() *Spacer {
	return &Spacer{}
}

func (s Spacer) Render(width int) []string {
	line := make([]string, width)
	for i := range width {
		line[i] = " "
	}
	return line
}

func (s Spacer) HandleInput(data string) {}

func (s Spacer) WantsKeyRelease() bool {
	return false
}

func (s Spacer) Invalidate() {
}
