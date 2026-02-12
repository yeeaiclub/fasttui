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

	input := "test string"

	// First call should calculate and cache
	result1 := VisibleWidth(input)
	if result1 != 11 {
		t.Errorf("First call: VisibleWidth(%q) = %d, expected 11", input, result1)
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

	// Fill cache to limit
	for i := 0; i < widthCacheSize; i++ {
		VisibleWidth(string(rune(i)))
	}

	initialSize := len(widthCache)
	if initialSize != widthCacheSize {
		t.Errorf("Cache size = %d, expected %d", initialSize, widthCacheSize)
	}

	// Add new item, should trigger cache eviction
	VisibleWidth("new item")

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
