package components

type Spacer struct {
	lines int
}

func NewSpacer(lines int) *Spacer {
	return &Spacer{lines: lines}
}

func (s Spacer) Render(width int) []string {
	line := make([]string, s.lines)
	for i := range s.lines {
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
