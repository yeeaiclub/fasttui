package components

type Editor struct {
	history []string
}

type LayoutLine struct {
	Text      string
	HasCursor bool
	CursorPos int
}

func NewEditor() *Editor {
	return &Editor{
		history: make([]string, 0),
	}
}

func (e *Editor) AddToHistory(text string) {
	if len(e.history) > 0 && e.history[0] == text {
		return
	}
	e.history = append([]string{text}, e.history...)
	if len(e.history) > 100 {
		e.history = e.history[:100]
	}
}

func (e *Editor) Render(width int) []string {
	return nil
}

func (e *Editor) HandleInput(data string) {

}

func (e *Editor) LayoutText(width int) {

}
