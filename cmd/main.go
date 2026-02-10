package main

import (
	"fmt"
	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)
	tui.AddChild(components.NewDynamicBorder(func(s string) string { return s }))

	input := components.NewInput()
	input.SetOnSubmit(func(value string) {
		fmt.Printf("Submitted: %q\n", value)
	})
	input.SetOnEscape(func() {
		fmt.Println("Escape pressed")
	})
	tui.AddChild(input)
	tui.SetFocus(input)

	tui.AddChild(components.NewDynamicBorder(func(s string) string { return s }))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		tui.Stop()
		fmt.Println("\nExiting...")
		os.Exit(0)
	}()

	tui.Start()

	done := make(chan struct{})
	<-done

}
