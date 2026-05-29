package fasttui

import (
	"sync"
	"testing"
)

type containerLineComponent struct {
	lines []string
}

func (c *containerLineComponent) Render(width int) []string { return c.lines }
func (c *containerLineComponent) HandleInput(string)        {}
func (c *containerLineComponent) WantsKeyRelease() bool     { return false }
func (c *containerLineComponent) Invalidate()               {}

func TestContainer_RenderSnapshot(t *testing.T) {
	c := NewContainer()
	c.AddChild(&containerLineComponent{lines: []string{"a"}})
	c.AddChild(&containerLineComponent{lines: []string{"b", "c"}})

	got := c.Render(80)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("Render() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestContainer_GetChildrenReturnsCopy(t *testing.T) {
	c := NewContainer()
	child := &containerLineComponent{lines: []string{"x"}}
	c.AddChild(child)

	children := c.GetChildren()
	children[0] = nil

	if c.GetChildren()[0] != child {
		t.Fatal("GetChildren should return a copy, not the internal slice")
	}
}

func TestContainer_RenderDoesNotHoldLockDuringChildRender(t *testing.T) {
	c := NewContainer()
	block := make(chan struct{})
	child := &blockingRenderComponent{block: block}
	c.AddChild(child)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = c.Render(80)
	}()

	// Render is blocked inside child; container mutations must still proceed.
	c.AddChild(&containerLineComponent{lines: []string{"added"}})
	close(block)
	wg.Wait()

	if len(c.GetChildren()) != 2 {
		t.Fatalf("expected 2 children after concurrent AddChild, got %d", len(c.GetChildren()))
	}
}

type blockingRenderComponent struct {
	block chan struct{}
}

func (b *blockingRenderComponent) Render(width int) []string {
	<-b.block
	return []string{"done"}
}
func (b *blockingRenderComponent) HandleInput(string)    {}
func (b *blockingRenderComponent) WantsKeyRelease() bool { return false }
func (b *blockingRenderComponent) Invalidate()           {}
