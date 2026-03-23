package fasttui

import (
	"testing"
)

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "plain string",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "string with tabs",
			input:    "hello\tworld",
			expected: 13, // "hello" (5) + "   " (3) + "world" (5) = 13
		},
		{
			name:     "string with ANSI color codes",
			input:    "\x1b[31mred text\x1b[0m",
			expected: 8, // "red text"
		},
		{
			name:     "string with ANSI CSI sequences",
			input:    "\x1b[1;32mbold green\x1b[0m",
			expected: 10, // "bold green"
		},
		{
			name:     "string with OSC sequences (hyperlink)",
			input:    "\x1b]8;;https://example.com\x07link\x1b]8;;\x07",
			expected: 4, // "link"
		},
		{
			name: "OSC 133 shell integration (zero visible width)",
			input: "\x1b]133;A\x07" + "hello" + "\x1b]133;B\x07\x1b]133;C\x07",
			expected: 5,
		},
		{
			name:     "string with APC sequences",
			input:    "\x1b_test\x07content",
			expected: 7, // "content"
		},
		{
			name:     "mixed tabs and ANSI codes",
			input:    "\x1b[31mred\t\x1b[32mgreen\x1b[0m",
			expected: 11, // "red" (3) + "   " (3) + "green" (5)
		},
		{
			name:     "multiple tabs",
			input:    "a\tb\tc",
			expected: 9, // "a" (1) + "   " (3) + "b" (1) + "   " (3) + "c" (1)
		},
		{
			name:     "complex ANSI sequences",
			input:    "\x1b[38;5;214mOrange\x1b[0m",
			expected: 6, // "Orange"
		},
		// Chinese characters
		{
			name:     "single Chinese character",
			input:    "你",
			expected: 2,
		},
		{
			name:     "Chinese phrase",
			input:    "你好",
			expected: 4,
		},
		{
			name:     "Chinese sentence",
			input:    "你好世界",
			expected: 8,
		},
		{
			name:     "mixed English and Chinese",
			input:    "Hello世界",
			expected: 9, // "Hello" (5) + "世界" (4)
		},
		{
			name:     "Chinese with ANSI codes",
			input:    "\x1b[31m你好\x1b[0m",
			expected: 4,
		},
		// Japanese characters
		{
			name:     "Hiragana",
			input:    "こんにちは",
			expected: 10, // 5 chars * 2
		},
		{
			name:     "Katakana",
			input:    "カタカナ",
			expected: 8, // 4 chars * 2
		},
		// Korean characters
		{
			name:     "Hangul",
			input:    "안녕하세요",
			expected: 10, // 5 chars * 2
		},
		// Emoji
		{
			name:     "simple emoji",
			input:    "😀",
			expected: 2,
		},
		{
			name:     "text with emoji",
			input:    "Hello 😀",
			expected: 8, // "Hello " (6) + emoji (2)
		},
		// Combining marks
		{
			name:     "ASCII with combining mark",
			input:    "a\u0301", // á
			expected: 1,
		},
		// Fullwidth forms
		{
			name:     "fullwidth A",
			input:    "Ａ",
			expected: 2,
		},
		{
			name:     "fullwidth numbers",
			input:    "０１２",
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VisibleWidth(tt.input)
			if result != tt.expected {
				t.Errorf("VisibleWidth(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestVisibleWidthCache(t *testing.T) {
	// Clear cache
	widthCache = make(map[string]int)

	// Use non-ASCII string to bypass fast path
	input := "你好世界"

	// First call should calculate and cache
	result1 := VisibleWidth(input)
	if result1 != 8 {
		t.Errorf("First call: VisibleWidth(%q) = %d, expected 8", input, result1)
	}

	// Check if cached
	if _, ok := widthCache[input]; !ok {
		t.Error("Expected string to be cached after first call")
	}

	// Second call should read from cache
	result2 := VisibleWidth(input)
	if result2 != result1 {
		t.Errorf("Second call: VisibleWidth(%q) = %d, expected %d", input, result2, result1)
	}
}

func TestVisibleWidthCacheEviction(t *testing.T) {
	// Clear cache
	widthCache = make(map[string]int)

	// Fill cache to limit using non-ASCII strings (to bypass fast path)
	for i := 0; i < widthCacheSize; i++ {
		// Use Chinese characters to ensure caching
		VisibleWidth("你" + string(rune(i)))
	}

	initialSize := len(widthCache)
	if initialSize != widthCacheSize {
		t.Errorf("Cache size = %d, expected %d", initialSize, widthCacheSize)
	}

	// Add new item, should trigger cache eviction
	VisibleWidth("新项目")

	// Cache should have evicted one item and added the new one
	if len(widthCache) > widthCacheSize {
		t.Errorf("Cache size exceeded limit: %d > %d", len(widthCache), widthCacheSize)
	}
}

func BenchmarkVisibleWidth(b *testing.B) {
	testCases := []string{
		"simple text",
		"\x1b[31mcolored text\x1b[0m",
		"text\twith\ttabs",
		"\x1b[1;32mbold green\x1b[0m with \ttabs",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				VisibleWidth(tc)
			}
		})
	}
}
