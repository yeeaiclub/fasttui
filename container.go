package fasttui

import (
	"slices"
	"sync"
)

const defaultContainerChildrenCap = 8

type Container struct {
	mu       sync.RWMutex
	children []Component
}

func NewContainer() *Container {
	return &Container{
		children: make([]Component, 0, defaultContainerChildrenCap),
	}
}

func (c *Container) AddChild(component Component) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.children = append(c.children, component)
}

func (c *Container) RemoveChild(component Component) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, child := range c.children {
		if child == component {
			c.children = slices.Delete(c.children, i, i+1)
			return
		}
	}
}

func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.children = nil
}

func (c *Container) Invalidate() {
	for _, child := range c.childrenSnapshot() {
		child.Invalidate()
	}
}

func (c *Container) Render(width int) []string {
	snapshot := c.childrenSnapshot()
	if len(snapshot) == 0 {
		return nil
	}

	lines := make([]string, 0, len(snapshot)*2)
	for _, child := range snapshot {
		lines = append(lines, child.Render(width)...)
	}
	return lines
}

func (c *Container) HandleInput(data string) {}

func (c *Container) WantsKeyRelease() bool {
	return false
}

func (c *Container) GetChildren() []Component {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.children) == 0 {
		return nil
	}
	out := make([]Component, len(c.children))
	copy(out, c.children)
	return out
}

func (c *Container) RemoveChildAt(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index >= 0 && index < len(c.children) {
		c.children = slices.Delete(c.children, index, index+1)
	}
}

func (c *Container) InsertChildAt(index int, component Component) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 {
		index = 0
	}
	if index >= len(c.children) {
		c.children = append(c.children, component)
		return
	}
	c.children = slices.Insert(c.children, index, component)
}

func (c *Container) childrenSnapshot() []Component {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.children) == 0 {
		return nil
	}
	out := make([]Component, len(c.children))
	copy(out, c.children)
	return out
}
