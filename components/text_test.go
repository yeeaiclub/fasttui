package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewText(t *testing.T) {
	text := NewText("Hello", 1, 1, nil)
	require.NotNil(t, text, "NewText should not return nil")
	assert.Equal(t, "Hello", text.text, "text should be 'Hello'")
	assert.Equal(t, 1, text.paddingX, "paddingX should be 1")
	assert.Equal(t, 1, text.paddingY, "paddingY should be 1")
}

func TestTextSetText(t *testing.T) {
	text := NewText("Hello", 1, 1, nil)
	text.SetText("World")
	assert.Equal(t, "World", text.text, "text should be updated to 'World'")
	assert.False(t, text.cacheValid, "cache should be invalidated after SetText")
}

func TestTextRenderEmpty(t *testing.T) {
	text := NewText("", 1, 1, nil)
	result := text.Render(10)
	assert.Empty(t, result, "result should be empty for empty text")
}

func TestTextRenderSimple(t *testing.T) {
	text := NewText("Hello", 1, 0, nil)
	result := text.Render(10)

	require.NotEmpty(t, result, "result should not be empty")

	// Should have: " Hello    " (1 space padding left, text, padding right to width)
	expected := " Hello    "
	assert.Equal(t, expected, result[0], "rendered line should match expected format")
}

func TestTextRenderWithPaddingY(t *testing.T) {
	text := NewText("Hello", 1, 1, nil)
	result := text.Render(10)

	// Should have: 1 empty line (top padding) + 1 content line + 1 empty line (bottom padding)
	assert.Len(t, result, 3, "should have 3 lines with paddingY=1")

	// First and last lines should be empty (spaces)
	assert.Empty(t, strings.TrimSpace(result[0]), "first line should be empty (padding)")
	assert.Empty(t, strings.TrimSpace(result[2]), "last line should be empty (padding)")
}

func TestTextRenderWithWrapping(t *testing.T) {
	text := NewText("Hello World", 1, 0, nil)
	result := text.Render(8)

	// With width 8 and paddingX 1, content width is 6
	// "Hello World" should wrap
	assert.GreaterOrEqual(t, len(result), 2, "text should wrap into at least 2 lines")
}

func TestTextRenderWithTabs(t *testing.T) {
	text := NewText("Hello\tWorld", 0, 0, nil)
	result := text.Render(20)

	require.NotEmpty(t, result, "result should not be empty")

	// Tabs should be replaced with 3 spaces
	assert.Contains(t, result[0], "Hello   World", "tabs should be replaced with 3 spaces")
}

func TestTextRenderCache(t *testing.T) {
	text := NewText("Hello", 1, 1, nil)

	// First render
	result1 := text.Render(10)
	assert.True(t, text.cacheValid, "cache should be valid after first render")

	// Second render with same width should use cache
	result2 := text.Render(10)
	assert.Equal(t, len(result1), len(result2), "cached result should match first result")

	// Render with different width should update cache
	text.Render(20)
	assert.Equal(t, 20, text.cachedWidth, "cache should be updated with new width")
}

func TestTextInvalidate(t *testing.T) {
	text := NewText("Hello", 1, 1, nil)
	text.Render(10)

	assert.True(t, text.cacheValid, "cache should be valid after render")

	text.Invalidate()
	assert.False(t, text.cacheValid, "cache should be invalid after Invalidate()")
}

func TestTextWithCustomBackground(t *testing.T) {
	bgFn := func(s string) string {
		return "\x1b[44m" + s + "\x1b[0m" // Blue background
	}

	text := NewText("Hello", 1, 0, bgFn)
	result := text.Render(10)

	require.NotEmpty(t, result, "result should not be empty")

	// Should contain ANSI codes
	assert.Contains(t, result[0], "\x1b[44m", "result should contain background ANSI code")
}

func TestTextSetCustomBgFn(t *testing.T) {
	text := NewText("Hello", 1, 1, nil)
	text.Render(10)

	bgFn := func(s string) string {
		return "\x1b[44m" + s + "\x1b[0m"
	}

	text.SetCustomBgFn(bgFn)
	assert.False(t, text.cacheValid, "cache should be invalidated after SetCustomBgFn")
}
