package components

type Markdown struct {
	text     string
	paddingX int
	paddingY int
}

func NewMarkdown() *Markdown {
	return &Markdown{}
}
