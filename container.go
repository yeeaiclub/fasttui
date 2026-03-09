package fasttui

import "sync"

type Container struct {
	mu       sync.RWMutex
	children []Component
}

func NewContainer() *Container {
	return &Container{}
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
			c.children = append(c.children[:i], c.children[i+1:]...)
			break
		}
	}
}

func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.children = []Component{}
}

func (c *Container) Invalidate() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, child := range c.children {
		child.Invalidate()
	}
}

func (c *Container) Render(width int) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var lines []string
	for _, child := range c.children {
		lines = append(lines, child.Render(width)...)
	}
	return lines
}

func (c *Container) HandleInput(data string) {
}

func (c *Container) WantsKeyRelease() bool {
	return false
}

func (c *Container) GetChildren() []Component {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.children
}

func (c *Container) RemoveChildAt(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index >= 0 && index < len(c.children) {
		c.children = append(c.children[:index], c.children[index+1:]...)
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
	c.children = append(c.children[:index+1], c.children[index:]...)
	c.children[index] = component
}
