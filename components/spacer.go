package components

import (
	"strings"

	"github.com/yeeaiclub/fasttui"
)

var _ fasttui.Component = (*Spacer)(nil)

type Spacer struct {
	lines int
}

func NewSpacer(lines int) *Spacer {
	return &Spacer{lines: lines}
}

func (s Spacer) Render(width int) []string {
	result := make([]string, s.lines)
	for i := range s.lines {
		result[i] = strings.Repeat(" ", width)
	}
	return result
}

func (s Spacer) HandleInput(data string) {}

func (s Spacer) WantsKeyRelease() bool {
	return false
}

func (s Spacer) Invalidate() {
}
