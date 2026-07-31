package components

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeeaiclub/fasttui"
)

func TestOrb_DrawFillsEllipse(t *testing.T) {
	orb := NewOrb(nil, WithOrbSize(20))
	orb.SetTime(0.5)

	lines := orb.Render(20)
	require.Len(t, lines, 20)

	nonEmpty := 0
	for _, line := range lines {
		plain := fasttui.StripAnsi(line)
		require.Equal(t, 20, len([]rune(plain)), "each line should be width 20")
		for _, r := range plain {
			if r != ' ' {
				nonEmpty++
			}
		}
	}
	assert.GreaterOrEqual(t, nonEmpty, 40, "expected orb cells")
}

func TestOrb_RenderClampsToWidth(t *testing.T) {
	orb := NewOrb(nil, WithOrbSize(40))
	orb.SetTime(1)
	lines := orb.Render(12)
	require.NotEmpty(t, lines)
	assert.LessOrEqual(t, len(lines), 24, "height should shrink with narrow width")
	for _, line := range lines {
		assert.Equal(t, 12, len([]rune(fasttui.StripAnsi(line))))
	}
}

func TestOrb_RushChangesOutput(t *testing.T) {
	a := NewOrb(nil, WithOrbSize(16), WithOrbRush(false))
	b := NewOrb(nil, WithOrbSize(16), WithOrbRush(true))
	a.SetTime(2)
	b.SetTime(2)

	la := strings.Join(a.Render(16), "\n")
	lb := strings.Join(b.Render(16), "\n")
	assert.NotEqual(t, la, lb, "rush speed should change noise sampling")
}

func TestOrb_CustomPalette(t *testing.T) {
	orb := NewOrb(nil,
		WithOrbSize(10),
		WithOrbPrimary(255, 0, 0),
		WithOrbSecondary(255, 255, 0),
	)
	orb.SetTime(0.5)
	joined := strings.Join(orb.Render(10), "")
	assert.Contains(t, joined, "\x1b[38;2;")
}

func TestOrb_StartStop(t *testing.T) {
	orb := NewOrb(nil, WithOrbSize(8), WithOrbTickInterval(10*time.Millisecond))
	orb.Start()
	defer orb.Stop()

	require.Eventually(t, func() bool {
		return orb.Time() > 0
	}, 500*time.Millisecond, 10*time.Millisecond)

	orb.Stop()
	orb.Stop() // idempotent
}

func TestOrb_InvalidateClearsCache(t *testing.T) {
	orb := NewOrb(nil, WithOrbSize(8))
	orb.SetTime(0.1)
	_ = orb.Render(8)
	orb.Invalidate()
	orb.SetTime(0.2)
	lines := orb.Render(8)
	require.Len(t, lines, 8)
}

func TestGlowNoise_sampleRange(t *testing.T) {
	n := newGlowNoise(2654435761)
	for x := 0; x < 20; x++ {
		for y := 0; y < 20; y++ {
			v := n.sample(float64(x), float64(y), 0.5, 1)
			assert.GreaterOrEqual(t, v, 0.0)
			assert.LessOrEqual(t, v, 1.0)
		}
	}
}
