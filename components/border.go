package components

type DynamicBorder struct {
	color func(string) string
}

func NewDynamicBorder(color func(string) string) *DynamicBorder {
	if color == nil {
		color = func(s string) string { return s }
	}
	return &DynamicBorder{
		color: color,
	}
}

func (d *DynamicBorder) Render(width int) []string {
	line := ""
	for range max(0, width) {
		line += "─"
	}
	return []string{d.color(line)}
}

func (d *DynamicBorder) HandleInput(data string) {
}

func (d *DynamicBorder) WantsKeyRelease() bool {
	return false
}

func (d *DynamicBorder) Invalidate() {
}
