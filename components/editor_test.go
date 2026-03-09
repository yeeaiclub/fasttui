package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapLine(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		contentWidth int
		cursorCol    int
		hasCursor    bool
		want         []LayoutLine
	}{
		{
			name:         "short line no wrap",
			line:         "hello",
			contentWidth: 10,
			cursorCol:    0,
			hasCursor:    false,
			want: []LayoutLine{
				{Text: "hello", HasCursor: false, CursorPos: 0},
			},
		},
		{
			name:         "short line with cursor",
			line:         "hello",
			contentWidth: 10,
			cursorCol:    3,
			hasCursor:    true,
			want: []LayoutLine{
				{Text: "hello", HasCursor: true, CursorPos: 3},
			},
		},
		{
			name:         "long line wraps once",
			line:         "hello world",
			contentWidth: 6,
			cursorCol:    0,
			hasCursor:    false,
			want: []LayoutLine{
				{Text: "hello ", HasCursor: false, CursorPos: 0},
				{Text: "world", HasCursor: false, CursorPos: 0},
			},
		},
		{
			name:         "long line wraps with cursor in first chunk",
			line:         "hello world",
			contentWidth: 6,
			cursorCol:    3,
			hasCursor:    true,
			want: []LayoutLine{
				{Text: "hello ", HasCursor: true, CursorPos: 3},
				{Text: "world", HasCursor: false, CursorPos: 0},
			},
		},
		{
			name:         "long line wraps with cursor in second chunk",
			line:         "hello world",
			contentWidth: 6,
			cursorCol:    8,
			hasCursor:    true,
			want: []LayoutLine{
				{Text: "hello ", HasCursor: false, CursorPos: 0},
				{Text: "world", HasCursor: true, CursorPos: 2},
			},
		},
		{
			name:         "long line wraps multiple times",
			line:         "abcdefghijklmnop",
			contentWidth: 5,
			cursorCol:    0,
			hasCursor:    false,
			want: []LayoutLine{
				{Text: "abcde", HasCursor: false, CursorPos: 0},
				{Text: "fghij", HasCursor: false, CursorPos: 0},
				{Text: "klmno", HasCursor: false, CursorPos: 0},
				{Text: "p", HasCursor: false, CursorPos: 0},
			},
		},
		{
			name:         "cursor at end of line",
			line:         "hello",
			contentWidth: 10,
			cursorCol:    5,
			hasCursor:    true,
			want: []LayoutLine{
				{Text: "hello", HasCursor: true, CursorPos: 5},
			},
		},
		{
			name:         "empty line",
			line:         "",
			contentWidth: 10,
			cursorCol:    0,
			hasCursor:    false,
			want:         nil,
		},
		{
			name:         "empty line with cursor",
			line:         "",
			contentWidth: 10,
			cursorCol:    0,
			hasCursor:    true,
			want:         nil,
		},
		{
			name:         "unicode characters",
			line:         "你好世界",
			contentWidth: 6,
			cursorCol:    0,
			hasCursor:    false,
			want: []LayoutLine{
				{Text: "你好世", HasCursor: false, CursorPos: 0},
				{Text: "界", HasCursor: false, CursorPos: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapLine(tt.line, tt.contentWidth, tt.cursorCol, tt.hasCursor)
			assert.Equal(t, tt.want, got)
		})
	}
}
