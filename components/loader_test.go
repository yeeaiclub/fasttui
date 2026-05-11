package components

import (
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeeaiclub/fasttui"
)

func TestLoader_buildSpinnerLine(t *testing.T) {
	l := &Loader{
		Text:    NewText("", 1, 0),
		frames:  []string{"A", "B"},
		message: "ignored",
	}
	assert.Equal(t, "A hello", l.buildSpinnerLine(0, "hello"))
	assert.Equal(t, "B world", l.buildSpinnerLine(1, "world"))

	l.spinnerColorFn = func(s string) string { return "<" + s + ">" }
	assert.Equal(t, "<B> z", l.buildSpinnerLine(1, "z"))

	l.spinnerColorFn = nil
	l.messageColorFn = func(s string) string { return "[" + s + "]" }
	assert.Equal(t, "A [m]", l.buildSpinnerLine(0, "m"))

	l.spinnerColorFn = func(s string) string { return "(" + s + ")" }
	assert.Equal(t, "(A) [x]", l.buildSpinnerLine(0, "x"))
}

func TestLoader_NewLoader_defaultMessage(t *testing.T) {
	l := NewLoader(nil, "")
	defer l.Stop()
	waitLoaderVisibleContent(t, l)
	body := strings.Join(l.Render(72), "\n")
	require.Contains(t, body, "Loading...")
}

func TestLoader_nonPositiveTickInterval_usesDefault(t *testing.T) {
	l := NewLoader(nil, "ok", WithLoaderTickInterval(0), WithLoaderTickInterval(-time.Millisecond))
	defer l.Stop()
	waitLoaderSubstring(t, l, "ok")
}

func TestLoader_Render_layoutAndMessage(t *testing.T) {
	l := NewLoader(nil, "PING")
	defer l.Stop()
	waitLoaderVisibleContent(t, l)

	lines := l.Render(80)
	require.GreaterOrEqual(t, len(lines), 2, "leading blank line + content")
	assert.Equal(t, "", lines[0])
	joined := strings.Join(lines[1:], "")
	require.Contains(t, joined, "PING")
	require.True(t, strings.ContainsAny(joined, "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"), "spinner frame present")
}

func TestLoader_SetMessage_updatesDisplay(t *testing.T) {
	l := NewLoader(nil, "alpha")
	defer l.Stop()
	waitLoaderSubstring(t, l, "alpha")

	l.SetMessage("beta")
	require.Eventually(t, func() bool {
		return strings.Contains(strings.Join(l.Render(80), ""), "beta")
	}, 500*time.Millisecond, 15*time.Millisecond, "rendered line should show new message")
}

func TestLoader_Stop_idempotent(t *testing.T) {
	l := NewLoader(nil, "x")
	waitLoaderVisibleContent(t, l)
	l.Stop()
	l.Stop()
}

func TestLoader_Start_secondCallNoop(t *testing.T) {
	l := &Loader{
		Text:    NewText("", 1, 0),
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: "m",
	}
	l.Start()
	waitLoaderVisibleContent(t, l)
	l.Start()
	l.Stop()
}

func TestLoader_StartStop_cycle(t *testing.T) {
	l := &Loader{
		Text:    NewText("", 1, 0),
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: "once",
	}
	l.Start()
	waitLoaderSubstring(t, l, "once")
	l.Stop()

	l.message = "twice"
	l.Start()
	waitLoaderSubstring(t, l, "twice")
	l.Stop()
}

func TestLoader_concurrentRender(t *testing.T) {
	l := NewLoader(nil, "race")
	defer l.Stop()
	waitLoaderVisibleContent(t, l)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 40; j++ {
				_ = l.Render(64)
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()
}

func waitLoaderVisibleContent(t *testing.T, l *Loader) {
	t.Helper()
	require.Eventually(t, func() bool {
		lines := l.Render(80)
		if len(lines) < 2 {
			return false
		}
		rest := strings.Join(lines[1:], "")
		return fasttui.VisibleWidth(strings.TrimSpace(rest)) > 0
	}, 500*time.Millisecond, 10*time.Millisecond, "expected loader to paint non-empty content")
}

func waitLoaderSubstring(t *testing.T, l *Loader, sub string) {
	t.Helper()
	require.Eventually(t, func() bool {
		return strings.Contains(strings.Join(l.Render(80), ""), sub)
	}, 500*time.Millisecond, 10*time.Millisecond, "expected substring %q in render", sub)
}
