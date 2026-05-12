package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpacer(t *testing.T) {
	spacer := NewSpacer(3)
	require.NotNil(t, spacer, "NewSpacer should not return nil")
}

func TestSpacerRender(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		spacer := NewSpacer(1)
		result := spacer.Render(10)
		assert.Len(t, result, 1, "should return 1 line")
		assert.Equal(t, "          ", result[0], "line should be 10 spaces")
	})

	t.Run("multiple lines", func(t *testing.T) {
		spacer := NewSpacer(3)
		result := spacer.Render(8)
		assert.Len(t, result, 3, "should return 3 lines")
		for i, line := range result {
			assert.Equal(t, "        ", line, "line %d should be 8 spaces", i)
		}
	})

	t.Run("zero width", func(t *testing.T) {
		spacer := NewSpacer(2)
		result := spacer.Render(0)
		assert.Len(t, result, 2, "should return 2 lines")
		for i, line := range result {
			assert.Equal(t, "", line, "line %d should be empty string", i)
		}
	})

	t.Run("zero lines", func(t *testing.T) {
		spacer := NewSpacer(0)
		result := spacer.Render(10)
		assert.Empty(t, result, "should return empty slice for 0 lines")
	})

	t.Run("width one", func(t *testing.T) {
		spacer := NewSpacer(2)
		result := spacer.Render(1)
		assert.Len(t, result, 2, "should return 2 lines")
		assert.Equal(t, " ", result[0], "each line should be a single space")
		assert.Equal(t, " ", result[1], "each line should be a single space")
	})

	t.Run("large width", func(t *testing.T) {
		spacer := NewSpacer(1)
		result := spacer.Render(100)
		assert.Len(t, result, 1, "should return 1 line")
		assert.Equal(t, 100, len(result[0]), "line should be 100 characters long")
		for _, ch := range result[0] {
			assert.Equal(t, ' ', ch, "each character should be a space")
		}
	})
}

func TestSpacerHandleInput(t *testing.T) {
	spacer := NewSpacer(1)
	spacer.HandleInput("test")
	// Should not panic
}

func TestSpacerWantsKeyRelease(t *testing.T) {
	spacer := NewSpacer(1)
	assert.False(t, spacer.WantsKeyRelease(), "Spacer should not want key release")
}

func TestSpacerInvalidate(t *testing.T) {
	spacer := NewSpacer(1)
	spacer.Invalidate()
	// Should not panic
}
