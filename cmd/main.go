package main

import (
	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

func main() {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)
	tui.AddChild(components.NewDynamicBorder(func(s string) string { return s }))

	input := components.NewInput()
	input.SetOnSubmit(func(value string) {
		// fmt.Printf("Submitted: %q\n", value)
	})
	input.SetOnEscape(func() {
		// fmt.Println("Escape pressed")
	})
	tui.AddChild(input)
	tui.SetFocus(input)
	tui.AddChild(components.NewDynamicBorder(func(s string) string { return s }))

	tui.Start()
	for true {
	}
	defer func() {
		tui.Stop()
	}()
}
