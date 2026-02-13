package main

import (
	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

func main() {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)
	editor := components.NewEditor(term, func(v string) {
		components.NewMarkdown(v, 1, 1)
		tui.RequestRender(false)
	})
	tui.AddChild(editor)
	tui.SetFocus(editor)

	// Start TUI
	tui.Start()

	// Keep running
	select {}
}
