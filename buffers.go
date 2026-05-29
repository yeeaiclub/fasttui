package fasttui

import (
	"strings"
	"sync"
)

const maxPooledBuilderCap = 64 * 1024

var builderPool = sync.Pool{
	New: func() any {
		b := new(strings.Builder)
		b.Grow(256)
		return b
	},
}

func acquireBuilder() *strings.Builder {
	b := builderPool.Get().(*strings.Builder)
	b.Reset()
	return b
}

// AcquireBuilder returns a pooled strings.Builder. Call ReleaseBuilder when done.
func AcquireBuilder() *strings.Builder {
	return acquireBuilder()
}

func releaseBuilder(b *strings.Builder) {
	if b.Cap() > maxPooledBuilderCap {
		return
	}
	builderPool.Put(b)
}

// ReleaseBuilder returns a builder from AcquireBuilder to the pool.
func ReleaseBuilder(b *strings.Builder) {
	releaseBuilder(b)
}

const maxPooledSegmentCap = 4096

var textSegmentPool = sync.Pool{
	New: func() any {
		s := make([]textSegment, 0, 32)
		return &s
	},
}

func acquireTextSegments() []textSegment {
	p := textSegmentPool.Get().(*[]textSegment)
	return (*p)[:0]
}

func releaseTextSegments(segs []textSegment) {
	if cap(segs) > maxPooledSegmentCap {
		return
	}
	textSegmentPool.Put(&segs)
}
