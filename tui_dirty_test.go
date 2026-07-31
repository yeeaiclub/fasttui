//go:build dirty

package fasttui

import (
	"strings"
	"testing"
	"time"
)

func TestDoRenderIdleShortCircuit(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	tui.AddChild(&concurrencyLineComponent{lines: []string{"hello", "world"}})

	tui.doRender()
	afterFirst := term.String()
	if afterFirst == "" {
		t.Fatal("first render should write output")
	}
	if tui.dirty.Load() {
		t.Fatal("dirty should be clear after paint")
	}

	before := term.String()
	tui.doRender() // idle: should no-op
	if term.String() != before {
		t.Fatalf("idle doRender wrote unexpected output")
	}
}

func TestTriggerRenderMarksDirtyAndPaints(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	comp := &concurrencyLineComponent{lines: []string{"v1"}}
	tui.AddChild(comp)
	tui.Start()
	defer tui.Stop()

	time.Sleep(20 * time.Millisecond)
	baseline := term.String()

	comp.lines = []string{"v2-via-trigger"}
	tui.TriggerRender()
	time.Sleep(30 * time.Millisecond)
	out := term.String()
	if out == baseline {
		t.Fatal("TriggerRender should cause a repaint")
	}
	if !strings.Contains(out, "v2-via-trigger") {
		t.Fatalf("repaint missing updated content; tail=%q", tailQuoted(out, 120))
	}
}

func TestIdleDoRenderSkipsWhenNotDirty(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	tui.AddChild(&concurrencyLineComponent{lines: []string{"static"}})
	tui.doRender()
	before := term.String()
	tui.doRender() // dirty already cleared — idle short-circuit
	if term.String() != before {
		t.Fatalf("idle doRender should not write")
	}
}

func TestMarkDirtySchedulesPaint(t *testing.T) {
	term := &recordingTerminal{}
	tui := NewTUI(term, false)
	comp := &concurrencyLineComponent{lines: []string{"v1"}}
	tui.AddChild(comp)
	tui.Start()
	defer tui.Stop()

	time.Sleep(20 * time.Millisecond)
	baseline := term.String()

	comp.lines = []string{"v2-changed"}
	tui.MarkDirty()
	time.Sleep(30 * time.Millisecond)
	out := term.String()
	if out == baseline {
		t.Fatal("MarkDirty should cause a repaint")
	}
	if !strings.Contains(out, "v2-changed") {
		t.Fatalf("repaint missing updated content; tail=%q", tailQuoted(out, 120))
	}
}
