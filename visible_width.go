package fasttui

import (
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	widthCacheSize = 512
)

var (
	widthCache      = make(map[string]int)
	widthCacheMutex sync.RWMutex
	ansiCSIPattern   = regexp.MustCompile(`\x1b\[[0-9;]*[mGKHJ]`)
	ansiOSCPattern   = regexp.MustCompile(`\x1b\]8;;[^\x07]*\x07`)
	ansiAPCPattern   = regexp.MustCompile(`\x1b_[^\x07\x1b]*(\x07|\x1b\\)`)
)

func VisibleWidth(s string) int {
	if len(s) == 0 {
		return 0
	}

	widthCacheMutex.RLock()
	if cached, ok := widthCache[s]; ok {
		widthCacheMutex.RUnlock()
		return cached
	}
	widthCacheMutex.RUnlock()

	clean := s
	if strings.Contains(clean, "\t") {
		clean = strings.ReplaceAll(clean, "\t", "   ")
	}
	if strings.Contains(clean, "\x1b") {
		clean = ansiCSIPattern.ReplaceAllString(clean, "")
		clean = ansiOSCPattern.ReplaceAllString(clean, "")
		clean = ansiAPCPattern.ReplaceAllString(clean, "")
	}

	width := utf8.RuneCountInString(clean)

	widthCacheMutex.Lock()
	if len(widthCache) >= widthCacheSize {
		for key := range widthCache {
			delete(widthCache, key)
			break
		}
	}
	widthCache[s] = width
	widthCacheMutex.Unlock()

	return width
}
