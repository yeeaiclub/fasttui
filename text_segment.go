package fasttui

import (
	"regexp"
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

var (
	// Combining marks (should not be counted separately, but with base character)
	combiningMarkRegex = regexp.MustCompile(`[\x{0300}-\x{036F}\x{1AB0}-\x{1AFF}\x{1DC0}-\x{1DFF}\x{20D0}-\x{20FF}\x{FE20}-\x{FE2F}]`)

	// Pure zero-width characters (not including combining marks)
	pureZeroWidthRegex = regexp.MustCompile(`^[\x{200B}-\x{200D}\x{FEFF}]+$`)
)

func GetSegmenter() any {
	return nil
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

	// Check for pure zero-width characters
	if pureZeroWidthRegex.MatchString(s) {
		return 0
	}

	// Get the first rune
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return 0
	}

	// Check if the first character is a combining mark (shouldn't happen in well-formed text)
	if combiningMarkRegex.MatchString(string(r)) {
		return 0
	}

	// Get the base character's width using go-runewidth with our custom condition
	// This library handles emoji, East Asian width, and other special cases
	width := runewidthCondition.RuneWidth(r)

	// Handle trailing characters
	if len(s) > size {
		for _, char := range s[size:] {
			// Skip combining marks - they don't add width
			if combiningMarkRegex.MatchString(string(char)) {
				continue
			}

			cp := int(char)
			// Halfwidth and Fullwidth Forms block
			if cp >= 0xFF00 && cp <= 0xFFEF {
				width += runewidthCondition.RuneWidth(char)
			}
		}
	}

	return width
}
