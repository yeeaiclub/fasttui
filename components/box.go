package components

import (
	"hash"
	"hash/fnv"
	"strings"

	"github.com/yeeaiclub/fasttui"
)

var _ fasttui.Component = (*Box)(nil)

// Box is a container component that applies padding and optional background to all children.
type Box struct {
	children []fasttui.Component
	paddingX  int
	paddingY  int
	bgFn      func(string) string

	cacheKey  uint64
	cacheResult []string
}

// BoxOption configures optional theme/background for Box.
type BoxOption func(*Box)

// WithBoxBackground sets the background function used when rendering.
func WithBoxBackground(bgFn func(string) string) BoxOption {
	return func(b *Box) {
		b.bgFn = bgFn
	}
}

// NewBox creates a Box with the given padding and optional background option.
func NewBox(paddingX, paddingY int, opts ...BoxOption) *Box {
	if paddingX < 0 {
		paddingX = 0
	}
	if paddingY < 0 {
		paddingY = 0
	}
	b := &Box{
		children:    nil,
		paddingX:    paddingX,
		paddingY:    paddingY,
		cacheKey:    0,
		cacheResult: nil,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}
	return b
}

// AddChild adds a child component and invalidates the cache.
func (b *Box) AddChild(component fasttui.Component) {
	b.children = append(b.children, component)
	b.invalidateCache()
}

// RemoveChild removes the first occurrence of the given component and invalidates the cache.
func (b *Box) RemoveChild(component fasttui.Component) {
	for i, c := range b.children {
		if c == component {
			b.children = append(b.children[:i], b.children[i+1:]...)
			b.invalidateCache()
			return
		}
	}
}

// Clear removes all children and invalidates the cache.
func (b *Box) Clear() {
	b.children = nil
	b.invalidateCache()
}

// SetBgFn sets the background function used when rendering. Does not invalidate cache;
// cache key includes a sample of bgFn output so changes are detected on next render.
func (b *Box) SetBgFn(bgFn func(string) string) {
	b.bgFn = bgFn
}

// GetChildren returns a copy of the current children slice.
func (b *Box) GetChildren() []fasttui.Component {
	if len(b.children) == 0 {
		return nil
	}
	out := make([]fasttui.Component, len(b.children))
	copy(out, b.children)
	return out
}

func (b *Box) invalidateCache() {
	b.cacheKey = 0
	b.cacheResult = nil
}

func (b *Box) computeCacheKey(width int, childLines []string, bgSample string) uint64 {
	h := fnv.New64a()
	writeUint64(h, uint64(width))
	writeUint64(h, uint64(len(childLines)))
	for _, line := range childLines {
		writeUint64(h, uint64(len(line)))
		h.Write([]byte(line))
	}
	h.Write([]byte(bgSample))
	return h.Sum64()
}

func writeUint64(h hash.Hash, u uint64) {
	var buf [8]byte
	buf[0] = byte(u)
	buf[1] = byte(u >> 8)
	buf[2] = byte(u >> 16)
	buf[3] = byte(u >> 24)
	buf[4] = byte(u >> 32)
	buf[5] = byte(u >> 40)
	buf[6] = byte(u >> 48)
	buf[7] = byte(u >> 56)
	h.Write(buf[:])
}

func (b *Box) Invalidate() {
	b.invalidateCache()
	for _, child := range b.children {
		child.Invalidate()
	}
}

func (b *Box) Render(width int) []string {
	if len(b.children) == 0 {
		return nil
	}

	contentWidth := max(1, width-b.paddingX*2)
	leftPad := strings.Repeat(" ", b.paddingX)

	var childLines []string
	for _, child := range b.children {
		lines := child.Render(contentWidth)
		for _, line := range lines {
			childLines = append(childLines, leftPad+line)
		}
	}

	if len(childLines) == 0 {
		return nil
	}

	bgSample := ""
	if b.bgFn != nil {
		bgSample = b.bgFn("test")
	}
	cacheKey := b.computeCacheKey(width, childLines, bgSample)
	if b.cacheResult != nil && b.cacheKey == cacheKey {
		return b.cacheResult
	}

	result := make([]string, 0, b.paddingY*2+len(childLines))

	for i := 0; i < b.paddingY; i++ {
		result = append(result, b.applyBg("", width))
	}
	for _, line := range childLines {
		result = append(result, b.applyBg(line, width))
	}
	for i := 0; i < b.paddingY; i++ {
		result = append(result, b.applyBg("", width))
	}

	b.cacheKey = cacheKey
	b.cacheResult = result
	return result
}

func (b *Box) applyBg(line string, width int) string {
	visLen := fasttui.VisibleWidth(line)
	padNeeded := max(0, width-visLen)
	padded := line + strings.Repeat(" ", padNeeded)
	if b.bgFn != nil {
		return fasttui.ApplyBackgroundToLine(padded, width, b.bgFn)
	}
	return padded
}

func (b *Box) HandleInput(data string) {}

func (b *Box) WantsKeyRelease() bool {
	return false
}
