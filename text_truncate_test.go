package fasttui

import (
	"testing"
)

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		ellipsis string
		pad      bool
		expected string
	}{
		// Basic ASCII tests
		{
			name:     "text fits exactly",
			text:     "hello",
			maxWidth: 5,
			ellipsis: "...",
			pad:      false,
			expected: "hello",
		},
		{
			name:     "text fits with padding",
			text:     "hello",
			maxWidth: 10,
			ellipsis: "...",
			pad:      true,
			expected: "hello     ",
		},
		{
			name:     "simple truncation",
			text:     "hello world",
			maxWidth: 8,
			ellipsis: "...",
			pad:      false,
			expected: "hello\x1b[0m...",
		},
		{
			name:     "truncation with padding",
			text:     "hello world",
			maxWidth: 10,
			ellipsis: "...",
			pad:      true,
			expected: "hello w\x1b[0m...",
		},
		{
			name:     "maxWidth too small for ellipsis",
			text:     "hello world",
			maxWidth: 2,
			ellipsis: "...",
			pad:      false,
			expected: "..",
		},
		{
			name:     "maxWidth equals ellipsis",
			text:     "hello world",
			maxWidth: 3,
			ellipsis: "...",
			pad:      false,
			expected: "...",
		},
		{
			name:     "empty text",
			text:     "",
			maxWidth: 5,
			ellipsis: "...",
			pad:      false,
			expected: "",
		},
		{
			name:     "empty text with padding",
			text:     "",
			maxWidth: 5,
			ellipsis: "...",
			pad:      true,
			expected: "     ",
		},

		// Chinese character tests
		{
			name:     "Chinese fits",
			text:     "你好",
			maxWidth: 4,
			ellipsis: "...",
			pad:      false,
			expected: "你好",
		},
		{
			name:     "Chinese truncation",
			text:     "你好世界",
			maxWidth: 5,
			ellipsis: "...",
			pad:      false,
			expected: "你\x1b[0m...",
		},
		{
			name:     "Chinese truncation exact",
			text:     "你好世界",
			maxWidth: 7,
			ellipsis: "...",
			pad:      false,
			expected: "你好\x1b[0m...",
		},
		{
			name:     "mixed English Chinese",
			text:     "Hello你好World",
			maxWidth: 10,
			ellipsis: "...",
			pad:      false,
			expected: "Hello你\x1b[0m...",
		},

		// ANSI code tests
		{
			name:     "ANSI codes preserved",
			text:     "\x1b[31mred text here\x1b[0m",
			maxWidth: 8,
			ellipsis: "...",
			pad:      false,
			expected: "\x1b[31mred t\x1b[0m...",
		},
		{
			name:     "ANSI codes with Chinese",
			text:     "\x1b[31m你好世界\x1b[0m",
			maxWidth: 7,
			ellipsis: "...",
			pad:      false,
			expected: "\x1b[31m你好\x1b[0m...",
		},
		{
			name:     "multiple ANSI codes",
			text:     "\x1b[1m\x1b[31mbold red text\x1b[0m",
			maxWidth: 9,
			ellipsis: "...",
			pad:      false,
			expected: "\x1b[1m\x1b[31mbold r\x1b[0m...",
		},

		// Custom ellipsis tests
		{
			name:     "custom ellipsis ASCII",
			text:     "hello world",
			maxWidth: 8,
			ellipsis: "..",
			pad:      false,
			expected: "hello \x1b[0m..",
		},
		{
			name:     "Chinese ellipsis",
			text:     "你好世界",
			maxWidth: 6,
			ellipsis: "。。",
			pad:      false,
			expected: "你\x1b[0m。。",
		},

		// Edge cases
		{
			name:     "maxWidth zero",
			text:     "hello",
			maxWidth: 0,
			ellipsis: "...",
			pad:      false,
			expected: "",
		},
		{
			name:     "maxWidth one",
			text:     "hello",
			maxWidth: 1,
			ellipsis: "...",
			pad:      false,
			expected: ".",
		},
		{
			name:     "emoji truncation",
			text:     "Hello 😀 World",
			maxWidth: 11,
			ellipsis: "...",
			pad:      false,
			expected: "Hello 😀\x1b[0m...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateToWidth(tt.text, tt.maxWidth, tt.ellipsis, tt.pad)
			if result != tt.expected {
				t.Errorf("TruncateToWidth(%q, %d, %q, %v)\ngot:  %q\nwant: %q",
					tt.text, tt.maxWidth, tt.ellipsis, tt.pad, result, tt.expected)
			}

			// Verify width constraint (when not padding)
			if !tt.pad {
				resultWidth := VisibleWidth(result)
				if resultWidth > tt.maxWidth {
					t.Errorf("Result width %d exceeds maxWidth %d for input %q",
						resultWidth, tt.maxWidth, tt.text)
				}
			}

			// Verify exact width when padding
			if tt.pad && tt.text != "" {
				resultWidth := VisibleWidth(result)
				if resultWidth != tt.maxWidth {
					t.Errorf("Padded result width %d != maxWidth %d for input %q",
						resultWidth, tt.maxWidth, tt.text)
				}
			}
		})
	}
}

func TestTruncateToWidth_DefaultEllipsis(t *testing.T) {
	// Test that empty ellipsis defaults to "..."
	result := TruncateToWidth("hello world", 8, "", false)
	expected := "hello\x1b[0m..."
	if result != expected {
		t.Errorf("TruncateToWidth with empty ellipsis\ngot:  %q\nwant: %q", result, expected)
	}
}

func TestSegmentText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []textSegment
	}{
		{
			name: "plain text",
			text: "hello",
			expected: []textSegment{
				{segType: segmentTypeGrapheme, value: "h"},
				{segType: segmentTypeGrapheme, value: "e"},
				{segType: segmentTypeGrapheme, value: "l"},
				{segType: segmentTypeGrapheme, value: "l"},
				{segType: segmentTypeGrapheme, value: "o"},
			},
		},
		{
			name: "text with ANSI",
			text: "\x1b[31mred\x1b[0m",
			expected: []textSegment{
				{segType: segmentTypeAnsi, value: "\x1b[31m"},
				{segType: segmentTypeGrapheme, value: "r"},
				{segType: segmentTypeGrapheme, value: "e"},
				{segType: segmentTypeGrapheme, value: "d"},
				{segType: segmentTypeAnsi, value: "\x1b[0m"},
			},
		},
		{
			name: "Chinese characters",
			text: "你好",
			expected: []textSegment{
				{segType: segmentTypeGrapheme, value: "你"},
				{segType: segmentTypeGrapheme, value: "好"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := segmentText(tt.text)
			if len(result) != len(tt.expected) {
				t.Errorf("segmentText(%q) returned %d segments, expected %d",
					tt.text, len(result), len(tt.expected))
				return
			}

			for i, seg := range result {
				if seg.segType != tt.expected[i].segType || seg.value != tt.expected[i].value {
					t.Errorf("segmentText(%q) segment %d:\ngot:  {type: %d, value: %q}\nwant: {type: %d, value: %q}",
						tt.text, i, seg.segType, seg.value, tt.expected[i].segType, tt.expected[i].value)
				}
			}
		})
	}
}

func BenchmarkTruncateToWidth_ASCII(b *testing.B) {
	text := "Hello, World! This is a test string that needs truncation."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TruncateToWidth(text, 30, "...", false)
	}
}

func BenchmarkTruncateToWidth_Chinese(b *testing.B) {
	text := "你好世界，这是一个需要截断的测试字符串。"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TruncateToWidth(text, 20, "...", false)
	}
}

func BenchmarkTruncateToWidth_WithANSI(b *testing.B) {
	text := "\x1b[31mHello, World!\x1b[0m This is a \x1b[1mtest\x1b[0m string."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TruncateToWidth(text, 30, "...", false)
	}
}
