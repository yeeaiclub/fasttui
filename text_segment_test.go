package fasttui

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/graphemes"
)

func TestGraphemeWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// ASCII characters
		{"Empty string", "", 0},
		{"Single ASCII", "a", 1},
		{"ASCII letter", "A", 1},
		{"ASCII digit", "5", 1},

		// Chinese characters (fullwidth)
		{"Single Chinese", "你", 2},
		{"Chinese character", "好", 2},
		{"Chinese character 2", "世", 2},
		{"Chinese character 3", "界", 2},

		// Japanese characters
		{"Hiragana", "あ", 2},
		{"Katakana", "ア", 2},
		{"Kanji", "漢", 2},

		// Korean characters
		{"Hangul", "한", 2},
		{"Hangul 2", "글", 2},

		// Emoji (typically width 2)
		{"Simple emoji", "😀", 2},
		{"Heart emoji", "❤", 1}, // Some emoji are width 1
		{"Thumbs up", "👍", 2},

		// Halfwidth and Fullwidth Forms
		{"Fullwidth A", "Ａ", 2},
		{"Fullwidth 0", "０", 2},

		// Zero-width characters
		{"Zero-width joiner", "\u200D", 0},
		{"Zero-width space", "\u200B", 0},

		// Special cases
		{"Tab", "\t", 0},     // Tab is a control character, width 0 in runewidth
		{"Newline", "\n", 0}, // Newline is a control character, width 0
		{"Space", " ", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GraphemeWidth(tt.input)
			if result != tt.expected {
				t.Errorf("GraphemeWidth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGraphemeWidth_CombiningMarks(t *testing.T) {
	// GraphemeWidth is designed for single grapheme clusters
	// Combining marks should be handled with their base character
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"ASCII with combining acute", "a\u0301", 1}, // a + combining acute = á
		{"e with combining acute", "e\u0301", 1},     // e + combining acute = é
		{"Chinese with combining", "你\u0301", 2},     // Chinese + combining mark
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GraphemeWidth(tt.input)
			if result != tt.expected {
				t.Errorf("GraphemeWidth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ---- Naive regex implementation for benchmark comparison ----

var (
	benchCombiningMarkRegex = regexp.MustCompile(`[\x{0300}-\x{036F}\x{1AB0}-\x{1AFF}\x{1DC0}-\x{1DFF}\x{20D0}-\x{20FF}\x{FE20}-\x{FE2F}]`)
	benchPureZeroWidthRegex = regexp.MustCompile(`^[\x{200B}-\x{200D}\x{FEFF}]+$`)
)

func graphemeWidthRegex(s string) int {
	if len(s) == 0 {
		return 0
	}

	if benchPureZeroWidthRegex.MatchString(s) {
		return 0
	}

	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return 0
	}

	if benchCombiningMarkRegex.MatchString(string(r)) {
		return 0
	}

	width := runewidthCondition.RuneWidth(r)

	if len(s) > size {
		for _, char := range s[size:] {
			if benchCombiningMarkRegex.MatchString(string(char)) {
				continue
			}

			cp := int(char)
			if cp >= 0xFF00 && cp <= 0xFFEF {
				width += runewidthCondition.RuneWidth(char)
			}
		}
	}

	return width
}

func TestGraphemeWidth_regexParity(t *testing.T) {
	inputs := []string{
		"", "a", "A", "你", "好", "😀", "👍", "❤", "Ａ",
		"\u200B", "\u200D", "\u200B\u200D", "a\u0301", "你\u0301",
		"e\u0301", "\t", "\n", " ",
	}
	for _, input := range inputs {
		got := GraphemeWidth(input)
		want := graphemeWidthRegex(input)
		if got != want {
			t.Errorf("GraphemeWidth(%q) = %d, regex impl = %d", input, got, want)
		}
	}
}

func collectGraphemes(s string) []string {
	g := graphemes.FromString(s)
	var out []string
	for g.Next() {
		out = append(out, g.Value())
	}
	return out
}

func benchmarkGraphemeWidthSingle(b *testing.B, s string, fn func(string) int) {
	b.SetBytes(1)
	for b.Loop() {
		if fn(s) < 0 {
			b.Fatal("unexpected width")
		}
	}
}

func benchmarkGraphemeWidthLoop(b *testing.B, text string, fn func(string) int) {
	graphemeList := collectGraphemes(text)
	b.SetBytes(int64(len(text)))
	for b.Loop() {
		width := 0
		for _, g := range graphemeList {
			width += fn(g)
		}
		if width <= 0 {
			b.Fatal("unexpected width")
		}
	}
}

func BenchmarkGraphemeWidth_regex_ascii(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "hello", graphemeWidthRegex)
}

func BenchmarkGraphemeWidth_optimized_ascii(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "hello", GraphemeWidth)
}

func BenchmarkGraphemeWidth_regex_cjk(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "你", graphemeWidthRegex)
}

func BenchmarkGraphemeWidth_optimized_cjk(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "你", GraphemeWidth)
}

func BenchmarkGraphemeWidth_regex_combining(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "e\u0301", graphemeWidthRegex)
}

func BenchmarkGraphemeWidth_optimized_combining(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "e\u0301", GraphemeWidth)
}

func BenchmarkGraphemeWidth_regex_emoji(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "👍", graphemeWidthRegex)
}

func BenchmarkGraphemeWidth_optimized_emoji(b *testing.B) {
	benchmarkGraphemeWidthSingle(b, "👍", GraphemeWidth)
}

var benchMixedGraphemeText = strings.Repeat("Hello世界 test你好 ", 200) + "😀👍"

func BenchmarkGraphemeWidth_regex_mixedLoop(b *testing.B) {
	benchmarkGraphemeWidthLoop(b, benchMixedGraphemeText, graphemeWidthRegex)
}

func BenchmarkGraphemeWidth_optimized_mixedLoop(b *testing.B) {
	benchmarkGraphemeWidthLoop(b, benchMixedGraphemeText, GraphemeWidth)
}

var benchCJKGraphemeText = strings.Repeat("你好世界", 500)

func BenchmarkGraphemeWidth_regex_cjkLoop(b *testing.B) {
	benchmarkGraphemeWidthLoop(b, benchCJKGraphemeText, graphemeWidthRegex)
}

func BenchmarkGraphemeWidth_optimized_cjkLoop(b *testing.B) {
	benchmarkGraphemeWidthLoop(b, benchCJKGraphemeText, GraphemeWidth)
}
