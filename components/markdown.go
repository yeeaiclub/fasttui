package components

type Markdown struct {
	text     string
	paddingX int
	paddingY int
}

func NewMarkdonw() *Markdown {
	return &Markdown{}
}
