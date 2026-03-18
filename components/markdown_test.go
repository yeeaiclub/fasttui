package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeeaiclub/fasttui"
)

func TestApplyPaddingAndBackground_ShortLine(t *testing.T) {
	m := NewMarkdown("", 1, 0)
	lines := []string{"Hi"}
	width := 10
	result := m.applyPaddingAndBackground(lines, width)

	require.Len(t, result, 1)
	// contentWidth=8, "Hi" fits; lineWithMargins = " Hi " then pad to 10
	assert.Equal(t, 10, fasttui.VisibleWidth(result[0]), "line must not exceed width")
	assert.Equal(t, " Hi       ", result[0])
}

func TestApplyPaddingAndBackground_LongLineWraps(t *testing.T) {
	m := NewMarkdown("", 1, 0)
	lines := []string{"Hello world here"}
	width := 10
	result := m.applyPaddingAndBackground(lines, width)

	require.GreaterOrEqual(t, len(result), 2, "long line should wrap to multiple lines")
	for i, line := range result {
		assert.LessOrEqual(t, fasttui.VisibleWidth(line), width,
			"line %d visible width must not exceed width", i)
	}
}

func TestApplyPaddingAndBackground_ExactContentWidth(t *testing.T) {
	m := NewMarkdown("", 1, 0)
	width := 10
	lines := []string{"12345678"}
	result := m.applyPaddingAndBackground(lines, width)

	require.Len(t, result, 1)
	assert.Equal(t, width, fasttui.VisibleWidth(result[0]))
	assert.Equal(t, " 12345678 ", result[0])
}

func TestApplyPaddingAndBackground_TopBottomPadding(t *testing.T) {
	m := NewMarkdown("", 0, 1)
	lines := []string{"x"}
	width := 5
	result := m.applyPaddingAndBackground(lines, width)

	require.Len(t, result, 3, "1 top padding + 1 content + 1 bottom padding")
	assert.Equal(t, "     ", result[0], "top padding: 5 spaces")
	assert.Equal(t, "x    ", result[1], "content padded to width")
	assert.Equal(t, "     ", result[2], "bottom padding: 5 spaces")
}

func TestApplyPaddingAndBackground_EmptyLines(t *testing.T) {
	m := NewMarkdown("", 0, 1)
	lines := []string{}
	width := 6
	result := m.applyPaddingAndBackground(lines, width)

	require.Len(t, result, 2, "only top and bottom padding when no content")
	assert.Equal(t, "      ", result[0])
	assert.Equal(t, "      ", result[1])
}

func TestApplyPaddingAndBackground_WithBgFn(t *testing.T) {
	wrapped := ""
	bgFn := func(s string) string {
		wrapped += "[" + s + "]"
		return "[" + s + "]"
	}
	m := NewMarkdown("", 0, 1, WithMarkdownDefaultTextStyle(&DefaultTextStyle{BgColor: bgFn}))
	lines := []string{"a"}
	width := 3
	result := m.applyPaddingAndBackground(lines, width)

	require.Len(t, result, 3)
	assert.Contains(t, wrapped, "[   ]", "bgFn should be applied to empty padding lines")
	assert.Equal(t, "a  ", result[1], "content line unchanged by bgFn in this impl")
}

func TestApplyPaddingAndBackground_NoExceedWidth(t *testing.T) {
	m := NewMarkdown("", 1, 0)
	lines := []string{"这是一段很长的中文内容需要被正确换行显示"}
	width := 12
	result := m.applyPaddingAndBackground(lines, width)

	for i, line := range result {
		assert.LessOrEqual(t, fasttui.VisibleWidth(line), width,
			"line %d must not exceed terminal width (prevents TUI panic)", i)
	}
}
