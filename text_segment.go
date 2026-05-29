package fasttui

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Configure runewidth to treat ambiguous-width characters as narrow (width 1)
// This matches the behavior of most modern terminals
var runewidthCondition = runewidth.NewCondition()

func init() {
	// Set EastAsianWidth to false to treat ambiguous characters (like box drawing
	// characters ─│┌┐└┘ and arrows ↑↓←→) as width 1 instead of width 2.
	// This matches how most terminals actually render these characters.
	runewidthCondition.EastAsianWidth = false
}

type segmentType int

const (
	segmentTypeAnsi segmentType = iota
	segmentTypeGrapheme
)

type textSegment struct {
	segType segmentType
	value   string
}

type SliceResult struct {
	text  string
	width int
}

func GetSegmenter() any {
	return nil
}

func isPureZeroWidthRune(r rune) bool {
	return r >= 0x200B && r <= 0x200D || r == 0xFEFF
}

func isPureZeroWidthString(s string) bool {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !isPureZeroWidthRune(r) {
			return false
		}
		i += size
	}
	return true
}

func isCombiningMark(r rune) bool {
	return r >= 0x0300 && r <= 0x036F ||
		r >= 0x1AB0 && r <= 0x1AFF ||
		r >= 0x1DC0 && r <= 0x1DFF ||
		r >= 0x20D0 && r <= 0x20FF ||
		r >= 0xFE20 && r <= 0xFE2F
}

// GraphemeWidth calculates the display width of a grapheme cluster
// It handles:
// - Zero-width characters (zero-width joiners, zero-width spaces)
// - Combining marks (counted with their base character)
// - Emoji (typically width 2)
// - East Asian characters (width 2 for fullwidth, 1 for halfwidth)
// - Regular ASCII (width 1)
func GraphemeWidth(s string) int {
	if len(s) == 0 {
		return 0
	}

	if isPureZeroWidthString(s) {
		return 0
	}

	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return 0
	}

	if isCombiningMark(r) {
		return 0
	}

	width := runewidthCondition.RuneWidth(r)

	for i := size; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if isCombiningMark(r) {
			i += n
			continue
		}
		if r >= 0xFF00 && r <= 0xFFEF {
			width += runewidthCondition.RuneWidth(r)
		}
		i += n
	}

	return width
}
