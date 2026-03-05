package fasttui

import (
	"regexp"
	"strings"
	"sync"

	"github.com/clipperhouse/uax29/v2/graphemes"
)

const (
	widthCacheSize = 512
)

var (
	widthCache      = make(map[string]int)
	widthCacheMutex sync.RWMutex
	ansiCSIPattern  = regexp.MustCompile(`\x1b\[[0-9;]*[mGKHJ]`)
	ansiOSCPattern  = regexp.MustCompile(`\x1b\]8;;[^\x07]*\x07`)
	ansiAPCPattern  = regexp.MustCompile(`\x1b_[^\x07\x1b]*(\x07|\x1b\\)`)
)

// VisibleWidth calculates the display width of a string, handling:
// - ANSI escape codes (stripped)
// - Tabs (converted to 3 spaces)
// - East Asian characters (width 2)
// - Emoji (typically width 2)
// - Combining marks (width 0)
// - Regular ASCII (width 1)
func VisibleWidth(s string) int {
	if len(s) == 0 {
		return 0
	}

	// Fast path: pure ASCII printable characters
	isPureAscii := true
	for i := 0; i < len(s); i++ {
		code := s[i]
		if code < 0x20 || code > 0x7e {
			isPureAscii = false
			break
		}
	}
	if isPureAscii {
		return len(s)
	}

	// Check cache
	widthCacheMutex.RLock()
	if cached, ok := widthCache[s]; ok {
		widthCacheMutex.RUnlock()
		return cached
	}
	widthCacheMutex.RUnlock()

	// Normalize: tabs to 3 spaces, strip ANSI escape codes
	clean := s
	if strings.Contains(clean, "\t") {
		clean = strings.ReplaceAll(clean, "\t", "   ")
	}
	if strings.Contains(clean, "\x1b") {
		clean = ansiCSIPattern.ReplaceAllString(clean, "")
		clean = ansiOSCPattern.ReplaceAllString(clean, "")
		clean = ansiAPCPattern.ReplaceAllString(clean, "")
	}

	// Calculate width using grapheme segmentation
	width := 0
	g := graphemes.FromString(clean)
	for g.Next() {
		grapheme := g.Value()
		width += GraphemeWidth(grapheme)
	}

	// Cache result
	widthCacheMutex.Lock()
	if len(widthCache) >= widthCacheSize {
		// Remove first entry (simple eviction strategy)
		for key := range widthCache {
			delete(widthCache, key)
			break
		}
	}
	widthCache[s] = width
	widthCacheMutex.Unlock()

	return width
}
