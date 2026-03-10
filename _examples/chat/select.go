package main

import (
	"strconv"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/keys"
)

type ExtensionSelectorOptions struct {
	TUI     fasttui.Terminal
	Timeout int
}

type ExtensionSelectorComponent struct {
	*fasttui.Container
	options          []string
	selectedIndex    int
	listContainer    *fasttui.Container
	onSelectCallback func(option string)
	onCancelCallback func()
	titleText        *components.Text
	baseTitle        string
	countdown        *CountdownTimer
	keybindings      *keys.EditorKeybindingsManager
}

func NewExtensionSelectorComponent(
	title string,
	options []string,
	onSelect func(option string),
	onCancel func(),
	opts *ExtensionSelectorOptions,
) *ExtensionSelectorComponent {
	e := &ExtensionSelectorComponent{
		Container:        fasttui.NewContainer(),
		options:          options,
		onSelectCallback: onSelect,
		onCancelCallback: onCancel,
		baseTitle:        title,
		keybindings:      keys.GetEditorKeybindings(),
	}

	e.AddChild(components.NewDynamicBorder(nil))
	e.AddChild(components.NewSpacer(1))

	e.titleText = components.NewText(ThemeFg("accent", title), 1, 0, nil)
	e.AddChild(e.titleText)
	e.AddChild(components.NewSpacer(1))

	if opts != nil && opts.Timeout > 0 && opts.TUI != nil {
		e.countdown = NewCountdownTimer(
			opts.Timeout,
			opts.TUI,
			func(s int) {
				titleWithTimer := e.baseTitle + " (" + strconv.Itoa(s) + "s)"
				e.titleText.SetText(ThemeFg("accent", titleWithTimer))
			},
			onCancel,
		)
	}

	e.listContainer = fasttui.NewContainer()
	e.AddChild(e.listContainer)
	e.AddChild(components.NewSpacer(1))
	e.AddChild(components.NewDynamicBorder(nil))
	e.updateList()
	return e
}

func (e *ExtensionSelectorComponent) updateList() {
	e.listContainer.Clear()
	for i, option := range e.options {
		isSelected := i == e.selectedIndex
		var text string
		if isSelected {
			text = ThemeFg("accent", "→ ") + ThemeFg("accent", option)
		} else {
			text = "  " + ThemeFg("text", option)
		}
		e.listContainer.AddChild(components.NewText(text, 1, 0, nil))
	}
}

func (e *ExtensionSelectorComponent) HandleInput(keyData string) {
	if e.keybindings.Matches(keyData, keys.EditorActionSelectUp) || keyData == "k" {
		if e.selectedIndex > 0 {
			e.selectedIndex--
		}
		e.updateList()
	} else if e.keybindings.Matches(keyData, keys.EditorActionSelectDown) || keyData == "j" {
		if e.selectedIndex < len(e.options)-1 {
			e.selectedIndex++
		}
		e.updateList()
	} else if e.keybindings.Matches(keyData, keys.EditorActionSelectConfirm) || keyData == "\n" {
		if e.selectedIndex >= 0 && e.selectedIndex < len(e.options) {
			e.onSelectCallback(e.options[e.selectedIndex])
		}
	} else if e.keybindings.Matches(keyData, keys.EditorActionSelectCancel) {
		e.onCancelCallback()
	}
}

func (e *ExtensionSelectorComponent) Dispose() {
	if e.countdown != nil {
		e.countdown.Dispose()
	}
}
