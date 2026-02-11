package fasttui

type Container struct {
	children []Component
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) AddChild(component Component) {
	c.children = append(c.children, component)
}

func (c *Container) RemoveChild(component Component) {
	for i, child := range c.children {
		if child == component {
			c.children = append(c.children[:i], c.children[i+1:]...)
			break
		}
	}
}

func (c *Container) Clear() {
	c.children = []Component{}
}

func (c *Container) Invalidate() {
	for _, child := range c.children {
		child.Invalidate()
	}
}

func (c *Container) Render(width int) []string {
	var lines []string
	for _, child := range c.children {
		lines = append(lines, child.Render(width)...)
	}
	return lines
}

func (c *Container) HandleInput(data string) {
	// 容器本身不处理输入，可由具体实现覆盖
}

func (c *Container) WantsKeyRelease() bool {
	return false
}
