package components

import "github.com/yeeaiclub/fasttui"

var _ fasttui.Component = (*DynamicBorder)(nil)

type DynamicBorder struct {
	color func(string) string
}

// DynamicBorderOption configures optional theming for DynamicBorder.
type DynamicBorderOption func(*DynamicBorder)

// WithBorderColor sets the color function used to render the border.
func WithBorderColor(color func(string) string) DynamicBorderOption {
	return func(d *DynamicBorder) {
		d.color = color
	}
}

func NewDynamicBorder(opts ...DynamicBorderOption) *DynamicBorder {
	d := &DynamicBorder{
		color: func(s string) string { return s },
	}

	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}

	return d
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
