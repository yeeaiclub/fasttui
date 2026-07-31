package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

func main() {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)
	tui.AddChild(components.NewDynamicBorder(components.WithBorderColor(func(s string) string { return s })))

	done := make(chan struct{})
	quit := func() {
		select {
		case <-done:
		default:
			close(done)
		}
	}

	input := components.NewInput()
	input.SetOnSubmit(func(value string) {
		// fmt.Printf("Submitted: %q\n", value)
	})
	// Input maps Escape and Ctrl+C to this callback (raw mode: no SIGINT).
	input.SetOnEscape(quit)
	tui.AddChild(input)
	tui.SetFocus(input)
	tui.AddChild(components.NewDynamicBorder(components.WithBorderColor(func(s string) string { return s })))

	tui.Start()
	defer tui.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
	case <-done:
	}
}
