package fasttui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceByColumn(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		startCol int
		length   int
		strict   bool
		expected string
	}{
		// Basic ASCII tests
		{"Empty string", "", 0, 5, false, ""},
		{"Simple ASCII slice", "Hello", 0, 3, false, "Hel"},
		{"ASCII with offset", "Hello", 1, 3, false, "ell"},
		{"ASCII full string", "Hello", 0, 10, false, "Hello"},
		{"ASCII beyond length", "Hello", 10, 5, false, ""},

		// Chinese characters (width 2)
		{"Chinese slice", "你好世界", 0, 4, false, "你好"},
		{"Chinese with offset", "你好世界", 2, 4, false, "好世"},
		{"Chinese partial", "你好世界", 0, 2, false, "你"},
		{"Chinese middle", "你好世界", 1, 3, false, "好"},

		// Mixed ASCII and Chinese
		{"Mixed content", "Hello世界", 0, 7, false, "Hello世"},
		{"Mixed with offset", "Hello世界", 5, 4, false, "世界"},

		// With ANSI codes
		{"ANSI red text", "\x1b[31mHello\x1b[0m", 0, 3, false, "\x1b[31mHel"},
		{"ANSI with offset", "\x1b[31mHello\x1b[0m", 1, 3, false, "\x1b[31mell"},
		{"ANSI before range", "\x1b[31mHello\x1b[0m", 0, 5, false, "\x1b[31mHello"},

		// Zero length
		{"Zero length", "Hello", 0, 0, false, ""},
		{"Negative length", "Hello", 0, -1, false, ""},

		// Strict mode
		{"Strict mode fits", "Hello", 0, 3, true, "Hel"},
		{"Strict mode Chinese", "你好", 0, 2, true, "你"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceByColumn(tt.line, tt.startCol, tt.length, tt.strict)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSliceWithWidth(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		startCol      int
		length        int
		strict        bool
		expectedText  string
		expectedWidth int
	}{
		// Basic tests
		{"Empty", "", 0, 5, false, "", 0},
		{"Simple ASCII", "Hello", 0, 3, false, "Hel", 3},
		{"Full ASCII", "Hello", 0, 10, false, "Hello", 5},

		// Chinese characters
		{"Chinese full", "你好", 0, 4, false, "你好", 4},
		{"Chinese partial", "你好", 0, 2, false, "你", 2},
		{"Chinese offset", "你好世界", 2, 4, false, "好世", 4},

		// With ANSI
		{"ANSI included", "\x1b[31mRed\x1b[0m", 0, 3, false, "\x1b[31mRed", 3},
		{"ANSI pending", "\x1b[31mHello", 0, 5, false, "\x1b[31mHello", 5},

		// Strict mode
		{"Strict Chinese", "你好", 0, 3, true, "你", 2},
		{"Strict fits", "Hello", 0, 5, true, "Hello", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceWithWidth(tt.line, tt.startCol, tt.length, tt.strict)
			assert.Equal(t, tt.expectedText, result.text)
			assert.Equal(t, tt.expectedWidth, result.width)
		})
	}
}

func TestExtractSegments(t *testing.T) {
	tests := []struct {
		name            string
		line            string
		beforeEnd       int
		afterStart      int
		afterLen        int
		strictAfter     bool
		expectedBefore  string
		expectedBeforeW int
		expectedAfter   string
		expectedAfterW  int
	}{
		// Basic split
		{"Simple split", "HelloWorld", 5, 5, 5, false, "Hello", 5, "World", 5},
		{"ASCII only", "ABCDEF", 3, 3, 3, false, "ABC", 3, "DEF", 3},

		// Chinese characters
		{"Chinese split", "你好世界", 4, 4, 4, false, "你好", 4, "世界", 4},
		{"Chinese partial", "你好世界", 2, 2, 2, false, "你", 2, "好", 2},

		// Overlapping ranges
		{"Overlapping", "HelloWorld", 3, 2, 6, false, "Hel", 3, "loWor", 5},

		// Zero after length
		{"No after", "HelloWorld", 5, 5, 0, false, "Hello", 5, "", 0},

		// With ANSI codes
		{"ANSI before", "\x1b[31mHello\x1b[0mWorld", 5, 5, 5, false, "\x1b[31mHello", 5, "World", 5},
		{"ANSI after", "Hello\x1b[31mWorld\x1b[0m", 5, 5, 5, false, "Hello", 5, "\x1b[31mWorld", 5},

		// Strict after
		{"Strict after", "HelloWorld", 5, 5, 3, true, "Hello", 5, "Wor", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, beforeW, after, afterW := ExtractSegments(
				tt.line, tt.beforeEnd, tt.afterStart, tt.afterLen, tt.strictAfter,
			)
			assert.Equal(t, tt.expectedBefore, before)
			assert.Equal(t, tt.expectedBeforeW, beforeW)
			assert.Equal(t, tt.expectedAfter, after)
			assert.Equal(t, tt.expectedAfterW, afterW)
		})
	}
}

func TestSliceByColumn_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		startCol int
		length   int
		strict   bool
		expected string
	}{
		{"Emoji slice", "😀😁😂", 0, 2, false, "😀"},
		{"Emoji offset", "😀😁😂", 2, 2, false, "😁"},
		{"Tab character", "a\tb", 0, 2, false, "a\tb"},
		{"Newline in text", "a\nb", 0, 2, false, "a\nb"},
		{"Combining marks", "a\u0301b", 0, 2, false, "a\u0301b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceByColumn(tt.line, tt.startCol, tt.length, tt.strict)
			assert.Equal(t, tt.expected, result)
		})
	}
}
