package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/terminal"
)

func main() {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)
	logger := NewKeyLogger(tui)

	tui.AddChild(logger)
	tui.SetFocus(logger)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		tui.Stop()
		os.Exit(0)
	}()

	tui.Start()
	select {}
}