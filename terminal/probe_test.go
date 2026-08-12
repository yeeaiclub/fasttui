package terminal

import (
	"testing"
	"time"

	"github.com/yeeaiclub/fasttui/keys"
)

func TestProbeConsumesRepliesAndEnables(t *testing.T) {
	p := NewProcessTerminal()
	var forwarded []string
	p.inputHandler = func(data string) {
		forwarded = append(forwarded, data)
	}
	p.setupStdinBuffer()

	go func() {
		time.Sleep(20 * time.Millisecond)
		p.buffer.Process([]byte(
			"\x1b[?2026;2$y" +
				"\x1b[?7u" +
				"\x1b[?1;2c",
		))
	}()

	select {
	case <-p.probeDoneCh:
	case <-time.After(probeTimeout):
		t.Fatal("probe did not complete after DA1")
	}

	c := p.Caps()
	if !c.KittyKeyboard {
		t.Fatalf("KittyKeyboard = false, want true; caps=%+v", c)
	}
	if !c.SyncOutput {
		t.Fatalf("SyncOutput = false, want true; caps=%+v", c)
	}
	if len(forwarded) != 0 {
		t.Fatalf("probe replies forwarded to app: %q", forwarded)
	}

	// After Done, normal input should forward (StdinBuffer delivers async).
	got := make(chan string, 1)
	p.inputHandler = func(data string) { got <- data }
	p.buffer.Process([]byte("a"))
	select {
	case s := <-got:
		if s != "a" {
			t.Fatalf("forwarded = %q, want a", s)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for forwarded input")
	}
}

func TestProbeTimeoutLeavesKittyOff(t *testing.T) {
	p := NewProcessTerminal()
	p.setupStdinBuffer()

	p.detector.Complete()
	p.signalProbeDone()

	c := p.Caps()
	if c.KittyKeyboard {
		t.Fatal("KittyKeyboard true after timeout with no replies")
	}
	keys.SetKittyProtocolActive(false)
}
