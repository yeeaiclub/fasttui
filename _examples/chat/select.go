package exselect

type ExSelectorComponent struct {
	options     []string
	selectIndex int
	baseTitle   string
	container   Container
}

func NewExSelectorComponent(
	title string,
	options []string,
) *ExSelectorComponent {
	e := &ExSelectorComponent{}
	e.options = options
	e.baseTitle = title
	return e
}

func (e *ExSelectorComponent) HandleInput() {
}

func (e *ExSelectorComponent) updateList() {

}
