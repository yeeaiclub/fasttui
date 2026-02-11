package components

type Spacer struct {
	lines int
}

func NewSpacer(lines int) *Spacer {
	return &Spacer{lines: lines}
}
