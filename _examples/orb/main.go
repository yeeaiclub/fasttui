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
	tui := fasttui.NewTUI(term, false)

	orb := components.NewOrb(
		tui,
		components.WithOrbSize(28),
		components.WithOrbRush(true),
	)
	tui.AddChild(orb)
	tui.AddChild(components.NewText("  ASCII orb — Ctrl+C to quit", 0, 1))

	done := make(chan struct{})
	tui.SetOnInterrupt(func() {
		select {
		case <-done:
		default:
			close(done)
		}
	})

	tui.Start()
	defer orb.Stop()
	defer tui.Stop()

	// External kill still delivers signals; Ctrl+C is handled via SetOnInterrupt
	// because raw mode disables ISIG.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
	case <-done:
	}
}
