package components

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock UI renderer for testing
type mockUIRenderer struct {
	renderCount int
}

func (m *mockUIRenderer) RequestRender(force bool) {
	m.renderCount++
}

func TestNewLoader(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Testing...")

	require.NotNil(t, loader, "NewLoader should not return nil")
	assert.Equal(t, "Testing...", loader.message, "message should be set")
	assert.Equal(t, 10, len(loader.frames), "should have 10 frames")
	assert.True(t, loader.running, "loader should be running after creation")
}

func TestNewLoaderDefaultMessage(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "")

	assert.Equal(t, "Loading...", loader.message, "should use default message")
}

func TestLoaderStart(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Testing...")

	// Loader starts automatically in NewLoader
	assert.True(t, loader.running, "loader should be running")

	// Wait for a few frames
	time.Sleep(250 * time.Millisecond)

	// Should have rendered multiple times
	assert.Greater(t, ui.renderCount, 2, "should have rendered multiple times")

	loader.Stop()
}

func TestLoaderStop(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Testing...")

	assert.True(t, loader.running, "loader should be running initially")

	loader.Stop()
	assert.False(t, loader.running, "loader should be stopped")

	// Multiple stops should be safe
	loader.Stop()
	assert.False(t, loader.running, "loader should still be stopped")
}

func TestLoaderSetMessage(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Initial")

	loader.SetMessage("Updated")
	assert.Equal(t, "Updated", loader.message, "message should be updated")

	loader.Stop()
}

func TestLoaderRender(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Testing...")

	result := loader.Render(50)

	// Should have at least one line (empty line at top)
	require.NotEmpty(t, result, "result should not be empty")

	// First line should be empty
	assert.Empty(t, strings.TrimSpace(result[0]), "first line should be empty")

	loader.Stop()
}

func TestLoaderWithColorFunctions(t *testing.T) {
	ui := &mockUIRenderer{}

	spinnerColor := func(s string) string {
		return "\x1b[32m" + s + "\x1b[0m" // Green
	}

	messageColor := func(s string) string {
		return "\x1b[33m" + s + "\x1b[0m" // Yellow
	}

	loader := NewLoader(ui, spinnerColor, messageColor, "Colored")

	result := loader.Render(50)
	require.NotEmpty(t, result, "result should not be empty")

	// Should contain ANSI color codes
	fullText := strings.Join(result, "")
	assert.Contains(t, fullText, "\x1b[32m", "should contain green color code")
	assert.Contains(t, fullText, "\x1b[33m", "should contain yellow color code")

	loader.Stop()
}

func TestLoaderFrameAnimation(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Animating...")

	initialFrame := loader.currentFrame

	// Wait for at least 2 frame updates (80ms each)
	time.Sleep(200 * time.Millisecond)

	currentFrame := loader.currentFrame
	assert.NotEqual(t, initialFrame, currentFrame, "frame should have changed")

	loader.Stop()
}

func TestLoaderFrameWraparound(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Testing...")

	loader.currentFrame = len(loader.frames) - 1

	// Wait for one frame update
	time.Sleep(100 * time.Millisecond)

	currentFrame := loader.currentFrame

	// Should wrap around to 0
	assert.Equal(t, 0, currentFrame, "frame should wrap around to 0")

	loader.Stop()
}

func TestLoaderConcurrentAccess(t *testing.T) {
	ui := &mockUIRenderer{}
	loader := NewLoader(ui, nil, nil, "Concurrent")

	// Concurrent message updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			loader.SetMessage("Message")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should have a valid message
	assert.NotEmpty(t, loader.message, "message should not be empty")

	loader.Stop()
}
