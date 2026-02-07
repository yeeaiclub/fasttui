package components

type Spacer struct {
	lines int
}

func NewSpacer(lines int) *Spacer {
	if lines < 1 {
		lines = 1
	}
	return &Spacer{lines: lines}
}

func (s *Spacer) Render(width int) []string {
	result := make([]string, width)
	for i := range width {
		result[i] = ""
	}
	return result
}

func (s Spacer) HandleInput(data string) {}

func (s Spacer) WantsKeyRelease() bool {
	return false
}

func (s Spacer) Invalidate() {
	// no cache
}
