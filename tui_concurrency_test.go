package fasttui

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// MockTerminal for testing
type MockTerminal struct {
	onInput  func(data string)
	onResize func()
	stopped  bool
}

func (m *MockTerminal) Start(onInput func(data string), onResize func()) error {
	m.onInput = onInput
	m.onResize = onResize
	return nil
}

func (m *MockTerminal) Stop() {
	m.stopped = true
}

func (m *MockTerminal) Write(data string) {}

func (m *MockTerminal) GetSize() (int, int) {
	return 80, 24
}

func (m *MockTerminal) IsKittyProtocolActive() bool {
	return false
}

func (m *MockTerminal) MoveBy(lines int)      {}
func (m *MockTerminal) HideCursor()           {}
func (m *MockTerminal) ShowCursor()           {}
func (m *MockTerminal) ClearLine()            {}
func (m *MockTerminal) ClearFromCursor()      {}
func (m *MockTerminal) ClearScreen()          {}
func (m *MockTerminal) SetTitle(title string) {}

// MockComponent for testing
type MockComponent struct {
	renderCount int
	inputCount  int
	focused     bool
}

func (m *MockComponent) Render(width int) []string {
	m.renderCount++
	return []string{"test"}
}

func (m *MockComponent) HandleInput(data string) {
	m.inputCount++
}

func (m *MockComponent) WantsKeyRelease() bool {
	return false
}

func (m *MockComponent) Invalidate() {}

func (m *MockComponent) SetFocused(focused bool) {
	m.focused = focused
}

func (m *MockComponent) IsFocused() bool {
	return m.focused
}

// TestConcurrentRenderRequests tests that multiple concurrent render requests don't cause race conditions
func TestConcurrentRenderRequests(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	// Give event loop time to start
	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Trigger many concurrent renders
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.TriggerRender()
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond) // Let renders complete

	// If we get here without panic, the test passes
}

// TestConcurrentInputAndRender tests concurrent input handling and rendering
func TestConcurrentInputAndRender(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)

	comp := &MockComponent{}
	tui.AddChild(comp)
	tui.SetFocus(comp)

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	numOps := 50

	// Concurrent input handling
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.HandleInput("a")
		}()
	}

	// Concurrent render requests
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.TriggerRender()
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Verify component received inputs
	if comp.inputCount == 0 {
		t.Error("Component should have received input")
	}
}

// TestConcurrentFocusChanges tests concurrent focus changes
func TestConcurrentFocusChanges(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)

	comp1 := &MockComponent{}
	comp2 := &MockComponent{}
	tui.AddChild(comp1)
	tui.AddChild(comp2)

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	numOps := 100

	// Rapidly switch focus between components
	for i := 0; i < numOps; i++ {
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

	// One of them should be focused
	if !comp1.focused && !comp2.focused {
		t.Error("One component should be focused")
	}
}

// TestStopWhileProcessing tests stopping TUI while operations are in progress
func TestStopWhileProcessing(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)
	tui.Start()

	time.Sleep(10 * time.Millisecond)

	// Start many operations
	for range 100 {
		go tui.TriggerRender()
		go tui.HandleInput("x")
	}

	// Stop immediately
	time.Sleep(5 * time.Millisecond)
	tui.Stop()

	// Should not panic or deadlock
	time.Sleep(50 * time.Millisecond)
}

// TestForceRenderConcurrent tests concurrent force renders
func TestForceRenderConcurrent(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tui.ForceRender()
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)
}

// TestConcurrentContainerOperations tests concurrent container modifications
func TestConcurrentContainerOperations(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)
	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	numOps := 50

	// Concurrent AddChild operations
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comp := &MockComponent{}
			tui.AddChild(comp)
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	// Should have added all components
	if len(tui.GetChildren()) != numOps {
		t.Errorf("Expected %d children, got %d", numOps, len(tui.GetChildren()))
	}
}

// TestConcurrentContainerMixedOps tests mixed container operations
func TestConcurrentContainerMixedOps(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)

	// Add some initial components before starting
	for i := 0; i < 10; i++ {
		tui.AddChild(&MockComponent{})
	}

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	numOps := 30

	// Concurrent mixed operations
	for i := range numOps {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			switch idx % 3 {
			case 0:
				// Add
				tui.AddChild(&MockComponent{})
			case 1:
				// Remove at index if exists
				children := tui.GetChildren()
				if len(children) > 0 {
					tui.RemoveChildAt(0)
				}
			case 2:
				// Insert
				tui.InsertChildAt(0, &MockComponent{})
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	// Should not panic and have some children
	children := tui.GetChildren()
	if children == nil {
		t.Error("Children should not be nil")
	}
}

// TestContainerOpsBeforeStart tests that container operations work before Start()
func TestContainerOpsBeforeStart(t *testing.T) {
	term := &MockTerminal{}
	tui := NewTUI(term, false)

	// Add children before starting
	comp1 := &MockComponent{}
	comp2 := &MockComponent{}
	tui.AddChild(comp1)
	tui.AddChild(comp2)

	if len(tui.GetChildren()) != 2 {
		t.Errorf("Expected 2 children before start, got %d", len(tui.GetChildren()))
	}

	tui.Start()
	defer tui.Stop()

	time.Sleep(10 * time.Millisecond)

	// Add more after starting
	comp3 := &MockComponent{}
	tui.AddChild(comp3)

	time.Sleep(10 * time.Millisecond)

	if len(tui.GetChildren()) != 3 {
		t.Errorf("Expected 3 children after start, got %d", len(tui.GetChildren()))
	}
}

type stopRaceTerminal struct {
	writeStarted chan struct{}
	allowWrite   chan struct{}
	startedOnce  sync.Once
	stopped      atomic.Bool
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
func (m *stopRaceTerminal) GetSize() (int, int)        { return 80, 24 }
func (m *stopRaceTerminal) IsKittyProtocolActive() bool { return false }
func (m *stopRaceTerminal) MoveBy(lines int)            {}
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

// TestStopWaitsForRenderCompletion ensures terminal.Stop is called after in-flight render exits.
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

	// Release blocked write so render can complete and event loop can exit.
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
