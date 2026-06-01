package fasttui

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testTerminal struct {
	onInput  func(data string)
	onResize func()
	stopped  atomic.Bool
}

func (m *testTerminal) Start(onInput func(data string), onResize func()) error {
	m.onInput = onInput
	m.onResize = onResize
	return nil
}

func (m *testTerminal) Stop() {
	m.stopped.Store(true)
}

func (m *testTerminal) Write(data string) {}

func (m *testTerminal) GetSize() (int, int) {
	return 80, 24
}

func (m *testTerminal) IsKittyProtocolActive() bool {
	return false
}

func (m *testTerminal) MoveBy(lines int)      {}
func (m *testTerminal) HideCursor()           {}
func (m *testTerminal) ShowCursor()           {}
func (m *testTerminal) ClearLine()            {}
func (m *testTerminal) ClearFromCursor()      {}
func (m *testTerminal) ClearScreen()          {}
func (m *testTerminal) SetTitle(title string) {}

type concurrencyLineComponent struct {
	lines []string
}

func (c *concurrencyLineComponent) Render(width int) []string { return c.lines }
func (c *concurrencyLineComponent) HandleInput(string)        {}
func (c *concurrencyLineComponent) WantsKeyRelease() bool     { return false }
func (c *concurrencyLineComponent) Invalidate()               {}

type concurrencyFocusComponent struct {
	lines      []string
	inputCount atomic.Int32
	focused    atomic.Bool
}

func (c *concurrencyFocusComponent) Render(width int) []string { return c.lines }
func (c *concurrencyFocusComponent) HandleInput(string) {
	c.inputCount.Add(1)
}
func (c *concurrencyFocusComponent) WantsKeyRelease() bool { return false }
func (c *concurrencyFocusComponent) Invalidate()           {}
func (c *concurrencyFocusComponent) SetFocused(focused bool) {
	c.focused.Store(focused)
}
func (c *concurrencyFocusComponent) IsFocused() bool {
	return c.focused.Load()
}

func newLineComponent(lines ...string) *concurrencyLineComponent {
	return &concurrencyLineComponent{lines: lines}
}

func newFocusComponent(lines ...string) *concurrencyFocusComponent {
	return &concurrencyFocusComponent{lines: lines}
}

func TestConcurrentRenderRequests(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.TriggerRender()
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)
}

func TestConcurrentInputAndRender(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)

	comp := newFocusComponent("test")
	tui.AddChild(comp)
	tui.SetFocus(comp)

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	const numOps = 50

	for range numOps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.HandleInput("a")
		}()
	}

	for range numOps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.TriggerRender()
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	if comp.inputCount.Load() == 0 {
		t.Error("Component should have received input")
	}
}

func TestConcurrentFocusChanges(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)

	comp1 := newFocusComponent("a")
	comp2 := newFocusComponent("b")
	tui.AddChild(comp1)
	tui.AddChild(comp2)

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				tui.SetFocus(comp1)
			} else {
				tui.SetFocus(comp2)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	if !comp1.IsFocused() && !comp2.IsFocused() {
		t.Error("One component should be focused")
	}
}

func TestStopWhileProcessing(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)
	tui.Start()

	time.Sleep(10 * time.Millisecond)

	for range 100 {
		go tui.TriggerRender()
		go tui.HandleInput("x")
	}

	time.Sleep(5 * time.Millisecond)
	tui.Stop()
	time.Sleep(50 * time.Millisecond)
}

func TestForceRenderConcurrent(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.ForceRender()
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)
}

func TestConcurrentContainerOperations(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	const numOps = 50

	for range numOps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.AddChild(newLineComponent("test"))
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	if len(tui.GetChildren()) != numOps {
		t.Errorf("Expected %d children, got %d", numOps, len(tui.GetChildren()))
	}
}

func TestConcurrentContainerMixedOps(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)

	for range 10 {
		tui.AddChild(newLineComponent("init"))
	}

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	const numOps = 30

	for i := range numOps {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			switch idx % 3 {
			case 0:
				tui.AddChild(newLineComponent("added"))
			case 1:
				children := tui.GetChildren()
				if len(children) > 0 {
					tui.RemoveChildAt(0)
				}
			case 2:
				tui.InsertChildAt(0, newLineComponent("inserted"))
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	if children := tui.GetChildren(); children == nil {
		t.Error("Children should not be nil")
	}
}

func TestContainerOpsBeforeStart(t *testing.T) {
	term := &testTerminal{}
	tui := NewTUI(term, false)

	comp1 := newLineComponent("a")
	comp2 := newLineComponent("b")
	tui.AddChild(comp1)
	tui.AddChild(comp2)

	if len(tui.GetChildren()) != 2 {
		t.Fatalf("Expected 2 children before start, got %d", len(tui.GetChildren()))
	}

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	tui.AddChild(newLineComponent("c"))
	time.Sleep(10 * time.Millisecond)

	if len(tui.GetChildren()) != 3 {
		t.Errorf("Expected 3 children after start, got %d", len(tui.GetChildren()))
	}
}

type stopRaceTerminal struct {
	writeStarted  chan struct{}
	allowWrite    chan struct{}
	startedOnce   sync.Once
	stopped       atomic.Bool
	hideAfterStop atomic.Int32
}

func newStopRaceTerminal() *stopRaceTerminal {
	return &stopRaceTerminal{
		writeStarted: make(chan struct{}),
		allowWrite:   make(chan struct{}),
	}
}

func (m *stopRaceTerminal) Start(onInput func(data string), onResize func()) error { return nil }
func (m *stopRaceTerminal) Stop()                                                  { m.stopped.Store(true) }
func (m *stopRaceTerminal) Write(data string) {
	if strings.Contains(data, SyncOutputBegin) {
		m.startedOnce.Do(func() { close(m.writeStarted) })
		<-m.allowWrite
	}
}
func (m *stopRaceTerminal) GetSize() (int, int)         { return 80, 24 }
func (m *stopRaceTerminal) IsKittyProtocolActive() bool { return false }
func (m *stopRaceTerminal) MoveBy(lines int)              {}
func (m *stopRaceTerminal) HideCursor() {
	if m.stopped.Load() {
		m.hideAfterStop.Add(1)
	}
}
func (m *stopRaceTerminal) ShowCursor()           {}
func (m *stopRaceTerminal) ClearLine()            {}
func (m *stopRaceTerminal) ClearFromCursor()      {}
func (m *stopRaceTerminal) ClearScreen()          {}
func (m *stopRaceTerminal) SetTitle(title string) {}

func TestStopWaitsForRenderCompletion(t *testing.T) {
	term := newStopRaceTerminal()
	tui := NewTUI(term, false)
	tui.Start()

	select {
	case <-term.writeStarted:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for render write to start")
	}

	done := make(chan struct{})
	go func() {
		tui.Stop()
		close(done)
	}()

	close(term.allowWrite)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop did not return in time")
	}

	if got := term.hideAfterStop.Load(); got != 0 {
		t.Fatalf("expected no HideCursor after terminal stop, got %d", got)
	}
}
