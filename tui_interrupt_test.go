package fasttui

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSetOnInterrupt_noFocus(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	var hit atomic.Bool
	tui.SetOnInterrupt(func() { hit.Store(true) })

	tui.HandleInput("\x03") // Ctrl+C under raw mode
	deadline := time.Now().Add(500 * time.Millisecond)
	for !hit.Load() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !hit.Load() {
		t.Fatal("expected SetOnInterrupt to fire for Ctrl+C when unfocused")
	}
}

func TestSetOnInterrupt_focusedConsumes(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	var interruptHit atomic.Bool
	var componentHit atomic.Bool
	tui.SetOnInterrupt(func() { interruptHit.Store(true) })
	comp := &interruptProbe{hit: &componentHit}
	tui.AddChild(comp)
	tui.SetFocus(comp)

	tui.HandleInput("\x03")
	time.Sleep(50 * time.Millisecond)

	if !componentHit.Load() {
		t.Fatal("focused component should receive Ctrl+C")
	}
	if interruptHit.Load() {
		t.Fatal("SetOnInterrupt must not fire when a component is focused")
	}
}

func TestStopFromInputHandler_noDeadlock(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	tui.Start()

	done := make(chan struct{})
	comp := &stopOnCtrlC{tui: tui, done: done}
	tui.AddChild(comp)
	tui.SetFocus(comp)

	tui.HandleInput("\x03")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop from input handler deadlocked")
	}

	// Second Stop from outside should also complete.
	finished := make(chan struct{})
	go func() {
		tui.Stop()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("subsequent Stop hung")
	}
}

type stopOnCtrlC struct {
	tui  *TUI
	done chan struct{}
}

func (s *stopOnCtrlC) Render(width int) []string { return nil }
func (s *stopOnCtrlC) HandleInput(data string) {
	if data == "\x03" {
		s.tui.Stop()
		select {
		case <-s.done:
		default:
			close(s.done)
		}
	}
}
func (s *stopOnCtrlC) WantsKeyRelease() bool { return false }
func (s *stopOnCtrlC) Invalidate()           {}
func (s *stopOnCtrlC) SetFocused(bool)       {}
func (s *stopOnCtrlC) IsFocused() bool       { return true }

type interruptProbe struct {
	hit *atomic.Bool
}

func (p *interruptProbe) Render(width int) []string { return nil }
func (p *interruptProbe) HandleInput(data string) {
	if data == "\x03" {
		p.hit.Store(true)
	}
}
func (p *interruptProbe) WantsKeyRelease() bool { return false }
func (p *interruptProbe) Invalidate()           {}
func (p *interruptProbe) SetFocused(bool)       {}
func (p *interruptProbe) IsFocused() bool       { return true }

