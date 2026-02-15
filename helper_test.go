package fasttui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSizeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		total    int
		expected int
	}{
		{
			name:     "nil value returns 0",
			value:    nil,
			total:    100,
			expected: 0,
		},
		{
			name:     "positive int value",
			value:    50,
			total:    100,
			expected: 50,
		},
		{
			name:     "zero int value returns 0",
			value:    0,
			total:    100,
			expected: 0,
		},
		{
			name:     "negative int value returns 0",
			value:    -10,
			total:    100,
			expected: 0,
		},
		{
			name:     "valid percentage string",
			value:    "50%",
			total:    100,
			expected: 50,
		},
		{
			name:     "percentage with decimal",
			value:    "25.5%",
			total:    100,
			expected: 25,
		},
		{
			name:     "100% of total",
			value:    "100%",
			total:    200,
			expected: 200,
		},
		{
			name:     "0% returns 0",
			value:    "0%",
			total:    100,
			expected: 0,
		},
		{
			name:     "invalid percentage string",
			value:    "abc",
			total:    100,
			expected: 0,
		},
		{
			name:     "invalid type returns 0",
			value:    3.14,
			total:    100,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSizeValue(tt.value, tt.total)
			assert.Equal(t, tt.expected, result, "parseSizeValue(%v, %d)", tt.value, tt.total)
		})
	}
}

func TestParseMargin(t *testing.T) {
	tests := []struct {
		name       string
		margin     any
		wantTop    int
		wantRight  int
		wantBottom int
		wantLeft   int
	}{
		{
			name:       "nil margin returns all zeros",
			margin:     nil,
			wantTop:    0,
			wantRight:  0,
			wantBottom: 0,
			wantLeft:   0,
		},
		{
			name:       "positive int applies to all sides",
			margin:     10,
			wantTop:    10,
			wantRight:  10,
			wantBottom: 10,
			wantLeft:   10,
		},
		{
			name:       "zero int returns all zeros",
			margin:     0,
			wantTop:    0,
			wantRight:  0,
			wantBottom: 0,
			wantLeft:   0,
		},
		{
			name:       "negative int returns all zeros",
			margin:     -5,
			wantTop:    0,
			wantRight:  0,
			wantBottom: 0,
			wantLeft:   0,
		},
		{
			name:       "map with all values",
			margin:     map[string]int{"top": 1, "right": 2, "bottom": 3, "left": 4},
			wantTop:    1,
			wantRight:  2,
			wantBottom: 3,
			wantLeft:   4,
		},
		{
			name:       "map with partial values",
			margin:     map[string]int{"top": 5},
			wantTop:    5,
			wantRight:  0,
			wantBottom: 0,
			wantLeft:   0,
		},
		{
			name:       "map with negative values",
			margin:     map[string]int{"top": -1, "right": 2},
			wantTop:    0,
			wantRight:  2,
			wantBottom: 0,
			wantLeft:   0,
		},
		{
			name:       "unsupported type returns all zeros",
			margin:     "invalid",
			wantTop:    0,
			wantRight:  0,
			wantBottom: 0,
			wantLeft:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			top, right, bottom, left := parseMargin(tt.margin)
			assert.Equal(t, tt.wantTop, top, "top margin")
			assert.Equal(t, tt.wantRight, right, "right margin")
			assert.Equal(t, tt.wantBottom, bottom, "bottom margin")
			assert.Equal(t, tt.wantLeft, left, "left margin")
		})
	}
}

func TestExtractCursorPosition(t *testing.T) {
	tests := []struct {
		name         string
		lines        []string
		height       int
		wantRow      int
		wantCol      int
		modifiedLine string
	}{
		{
			name:         "cursor at end of single line",
			lines:        []string{"hello" + CursorMarker},
			height:       10,
			wantRow:      0,
			wantCol:      5,
			modifiedLine: "hello",
		},
		{
			name:         "cursor in middle of line",
			lines:        []string{"he" + CursorMarker + "llo"},
			height:       10,
			wantRow:      0,
			wantCol:      2,
			modifiedLine: "hello",
		},
		{
			name:         "cursor at beginning",
			lines:        []string{CursorMarker + "hello"},
			height:       10,
			wantRow:      0,
			wantCol:      0,
			modifiedLine: "hello",
		},
		{
			name:         "no cursor marker returns -1, -1",
			lines:        []string{"hello", "world"},
			height:       10,
			wantRow:      -1,
			wantCol:      -1,
			modifiedLine: "",
		},
		{
			name:         "cursor in second line",
			lines:        []string{"hello", "world" + CursorMarker},
			height:       10,
			wantRow:      1,
			wantCol:      5,
			modifiedLine: "world",
		},
		{
			name:         "multiple lines with cursor in viewport",
			lines:        []string{"line1", "line2", "line3" + CursorMarker},
			height:       2,
			wantRow:      2,
			wantCol:      5,
			modifiedLine: "line3",
		},
		{
			name:         "empty line with cursor",
			lines:        []string{CursorMarker},
			height:       10,
			wantRow:      0,
			wantCol:      0,
			modifiedLine: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of lines to avoid modifying the test case
			lines := make([]string, len(tt.lines))
			copy(lines, tt.lines)

			row, col := extractCursorPosition(lines, tt.height)

			assert.Equal(t, tt.wantRow, row, "row")
			assert.Equal(t, tt.wantCol, col, "col")

			if tt.modifiedLine != "" {
				assert.Equal(t, tt.modifiedLine, lines[tt.wantRow], "line after extraction")
			}
		})
	}
}

func TestContainsImage(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "kitty graphics protocol",
			line:     "\x1b_Gf=32,s=10,v=10;data\x1b\\",
			expected: true,
		},
		{
			name:     "iterm2 inline image protocol",
			line:     "\x1b]1337;File=name=test.png;inline=1:data\x07",
			expected: true,
		},
		{
			name:     "plain text without image",
			line:     "hello world",
			expected: false,
		},
		{
			name:     "empty string",
			line:     "",
			expected: false,
		},
		{
			name:     "text with kitty marker in middle",
			line:     "before\x1b_Gafter",
			expected: true,
		},
		{
			name:     "text with iterm2 marker in middle",
			line:     "before\x1b]1337;File=after",
			expected: true,
		},
		{
			name:     "both protocols present",
			line:     "\x1b_Gkitty\x1b]1337;File=iterm2",
			expected: true,
		},
		{
			name:     "similar but not exact match",
			line:     "_G]1337;File=",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsImage(tt.line)
			assert.Equal(t, tt.expected, result, "containsImage(%q)", tt.line)
		})
	}
}

func TestApplyLineRests(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected []string
	}{
		{
			name:     "plain text lines get reset",
			lines:    []string{"hello", "world"},
			expected: []string{"hello" + SEGMENT_RESET, "world" + SEGMENT_RESET},
		},
		{
			name:     "lines with kitty image are unchanged",
			lines:    []string{"\x1b_Gimage"},
			expected: []string{"\x1b_Gimage"},
		},
		{
			name:     "lines with iterm2 image are unchanged",
			lines:    []string{"\x1b]1337;File=image"},
			expected: []string{"\x1b]1337;File=image"},
		},
		{
			name:     "mixed lines",
			lines:    []string{"text", "\x1b_Gimage", "more text"},
			expected: []string{"text" + SEGMENT_RESET, "\x1b_Gimage", "more text" + SEGMENT_RESET},
		},
		{
			name:     "empty line gets reset",
			lines:    []string{""},
			expected: []string{"" + SEGMENT_RESET},
		},
		{
			name:     "empty input",
			lines:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyLineRests(tt.lines)
			assert.Equal(t, tt.expected, result, "applyLineRests()")
		})
	}
}

func TestFindChangedLineRange(t *testing.T) {
	tests := []struct {
		name          string
		oldLines      []string
		newLines      []string
		wantFirst     int
		wantLast      int
	}{
		{
			name:          "identical lines return -1, -1",
			oldLines:      []string{"hello", "world"},
			newLines:      []string{"hello", "world"},
			wantFirst:     -1,
			wantLast:      -1,
		},
		{
			name:          "single line changed",
			oldLines:      []string{"hello", "world"},
			newLines:      []string{"hello", "changed"},
			wantFirst:     1,
			wantLast:      1,
		},
		{
			name:          "multiple lines changed",
			oldLines:      []string{"line1", "line2", "line3", "line4"},
			newLines:      []string{"line1", "changed2", "changed3", "line4"},
			wantFirst:     1,
			wantLast:      2,
		},
		{
			name:          "first line changed",
			oldLines:      []string{"old", "line2", "line3"},
			newLines:      []string{"new", "line2", "line3"},
			wantFirst:     0,
			wantLast:      0,
		},
		{
			name:          "last line changed",
			oldLines:      []string{"line1", "line2", "old"},
			newLines:      []string{"line1", "line2", "new"},
			wantFirst:     2,
			wantLast:      2,
		},
		{
			name:          "new lines appended",
			oldLines:      []string{"line1", "line2"},
			newLines:      []string{"line1", "line2", "line3", "line4"},
			wantFirst:     2,
			wantLast:      3,
		},
		{
			name:          "lines removed",
			oldLines:      []string{"line1", "line2", "line3", "line4"},
			newLines:      []string{"line1", "line2"},
			wantFirst:     2,
			wantLast:      3,
		},
		{
			name:          "empty old lines",
			oldLines:      []string{},
			newLines:      []string{"line1", "line2"},
			wantFirst:     0,
			wantLast:      1,
		},
		{
			name:          "empty new lines",
			oldLines:      []string{"line1", "line2"},
			newLines:      []string{},
			wantFirst:     0,
			wantLast:      1,
		},
		{
			name:          "both empty",
			oldLines:      []string{},
			newLines:      []string{},
			wantFirst:     -1,
			wantLast:      -1,
		},
		{
			name:          "all lines changed",
			oldLines:      []string{"a", "b", "c"},
			newLines:      []string{"x", "y", "z"},
			wantFirst:     0,
			wantLast:      2,
		},
		{
			name:          "single line unchanged",
			oldLines:      []string{"same"},
			newLines:      []string{"same"},
			wantFirst:     -1,
			wantLast:      -1,
		},
		{
			name:          "single line changed",
			oldLines:      []string{"old"},
			newLines:      []string{"new"},
			wantFirst:     0,
			wantLast:      0,
		},
		{
			name:          "different lengths with common prefix",
			oldLines:      []string{"prefix", "old"},
			newLines:      []string{"prefix", "new", "extra"},
			wantFirst:     1,
			wantLast:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, last := findChangedLineRange(tt.oldLines, tt.newLines)
			assert.Equal(t, tt.wantFirst, first, "first changed index")
			assert.Equal(t, tt.wantLast, last, "last changed index")
		})
	}
}
