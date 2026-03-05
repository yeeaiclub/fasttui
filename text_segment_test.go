package fasttui

import (
	"testing"
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
