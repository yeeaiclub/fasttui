package fasttui

// FocusManager handles component focus state
type FocusManager struct {
	focusedComponent Component
}

func newFocusManager() *FocusManager {
	return &FocusManager{}
}

func (fm *FocusManager) SetFocus(component Component) {
	if fm.focusedComponent != nil {
		if f, ok := fm.focusedComponent.(Focusable); ok {
			f.SetFocused(false)
		}
	}

	fm.focusedComponent = component

	if component != nil {
		if f, ok := component.(Focusable); ok {
			f.SetFocused(true)
		}
	}
}

func (fm *FocusManager) GetFocused() Component {
	return fm.focusedComponent
}
