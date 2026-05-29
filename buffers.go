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

func releaseBuilder(b *strings.Builder) {
	if b.Cap() > maxPooledBuilderCap {
		return
	}
	builderPool.Put(b)
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
