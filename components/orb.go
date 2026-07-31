package components

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/yeeaiclub/fasttui"
)

var _ fasttui.Component = (*Orb)(nil)

// Classic orb charset.
const orbCharset = " .:-=+*#%@"

const (
	orbDefaultSize         = 40
	orbDefaultTickInterval = time.Second / 30
	orbPaletteSize         = 64
)

type orbRGB struct{ r, g, b uint8 }

func lerpOrbRGB(a, b orbRGB, t float64) orbRGB {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return orbRGB{
		r: uint8(math.Round(float64(a.r) + float64(b.r-a.r)*t)),
		g: uint8(math.Round(float64(a.g) + float64(b.g-a.g)*t)),
		b: uint8(math.Round(float64(a.b) + float64(b.b-a.b)*t)),
	}
}

// Orb is an animated ASCII glowing sphere.
// Ported from openamp's Orb: ellipse crop + dual-octave value noise + green→cyan palette.
type Orb struct {
	mu   sync.RWMutex
	life sync.Mutex

	width  int
	height int
	time   float64
	rush   bool

	primary   orbRGB
	secondary orbRGB
	noise     glowNoise
	palette   []string // ANSI truecolor fg prefixes
	chars     []rune

	ui           *fasttui.TUI
	tickInterval time.Duration
	startedAt    time.Time

	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}

	cachedLines []string
	cachedWidth int
	cachedTime  float64
	cacheValid  bool
}

// OrbOption configures optional Orb behavior.
type OrbOption func(*Orb)

// WithOrbSize sets the sphere width and height in cells (default 40).
func WithOrbSize(size int) OrbOption {
	return func(o *Orb) {
		if size > 0 {
			o.width = size
			o.height = size
		}
	}
}

// WithOrbDimensions sets width and height independently.
func WithOrbDimensions(width, height int) OrbOption {
	return func(o *Orb) {
		if width > 0 {
			o.width = width
		}
		if height > 0 {
			o.height = height
		}
	}
}

// WithOrbRush enables faster noise scroll (speed 2.5 vs 1.0).
func WithOrbRush(rush bool) OrbOption {
	return func(o *Orb) {
		o.rush = rush
	}
}

// WithOrbTickInterval sets the animation refresh interval.
// Non-positive values are replaced with the default when Start runs.
func WithOrbTickInterval(d time.Duration) OrbOption {
	return func(o *Orb) {
		o.tickInterval = d
	}
}

// WithOrbPrimary sets the dark end of the color palette (default #003700).
func WithOrbPrimary(r, g, b uint8) OrbOption {
	return func(o *Orb) {
		o.primary = orbRGB{r, g, b}
	}
}

// WithOrbSecondary sets the bright end of the color palette (default #00FF88).
func WithOrbSecondary(r, g, b uint8) OrbOption {
	return func(o *Orb) {
		o.secondary = orbRGB{r, g, b}
	}
}

// NewOrb creates an animated Orb and starts its ticker when ui is non-nil.
func NewOrb(ui *fasttui.TUI, opts ...OrbOption) *Orb {
	o := &Orb{
		width:        orbDefaultSize,
		height:       orbDefaultSize,
		ui:           ui,
		tickInterval: orbDefaultTickInterval,
		primary:      orbRGB{0, 55, 0},
		secondary:    orbRGB{0, 255, 136},
		noise:        newGlowNoise(2654435761),
		chars:        []rune(orbCharset),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	o.buildPalette()
	if ui != nil {
		o.Start()
	}
	return o
}

func (o *Orb) buildPalette() {
	o.palette = make([]string, orbPaletteSize)
	for i := 0; i < orbPaletteSize; i++ {
		c := lerpOrbRGB(o.primary, o.secondary, float64(i)/float64(orbPaletteSize-1))
		o.palette[i] = fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.r, c.g, c.b)
	}
}

// SetTime sets the animation time in seconds (for tests / external drivers).
func (o *Orb) SetTime(t float64) {
	o.mu.Lock()
	o.time = t
	o.cacheValid = false
	o.mu.Unlock()
}

// Time returns the current animation time in seconds.
func (o *Orb) Time() float64 {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.time
}

// SetSize updates the orb dimensions.
func (o *Orb) SetSize(width, height int) {
	o.mu.Lock()
	if width > 0 {
		o.width = width
	}
	if height > 0 {
		o.height = height
	}
	o.cacheValid = false
	o.mu.Unlock()
}

func (o *Orb) Invalidate() {
	o.mu.Lock()
	o.cacheValid = false
	o.mu.Unlock()
}

func (o *Orb) HandleInput(_ string) {}

func (o *Orb) WantsKeyRelease() bool { return false }

func (o *Orb) Start() {
	o.life.Lock()
	defer o.life.Unlock()
	if o.running {
		return
	}
	if o.tickInterval <= 0 {
		o.tickInterval = orbDefaultTickInterval
	}
	o.running = true
	o.startedAt = time.Now()
	o.stopCh = make(chan struct{})
	o.doneCh = make(chan struct{})
	go o.loop()
}

func (o *Orb) loop() {
	stop := o.stopCh
	done := o.doneCh
	ticker := time.NewTicker(o.tickInterval)
	defer func() {
		ticker.Stop()
		close(done)
	}()

	o.advance()
	if o.ui != nil {
		o.ui.TriggerRender()
	}

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			o.advance()
			if o.ui != nil {
				o.ui.TriggerRender()
			}
		}
	}
}

func (o *Orb) advance() {
	o.mu.Lock()
	o.time = time.Since(o.startedAt).Seconds()
	o.cacheValid = false
	o.mu.Unlock()
}

func (o *Orb) Stop() {
	o.life.Lock()
	if !o.running {
		o.life.Unlock()
		return
	}
	close(o.stopCh)
	done := o.doneCh
	o.running = false
	o.stopCh = nil
	o.doneCh = nil
	o.life.Unlock()
	<-done
}

func (o *Orb) Render(width int) []string {
	o.mu.Lock()
	defer o.mu.Unlock()

	w, h := o.width, o.height
	if width > 0 && w > width {
		w = width
	}
	if w < 3 {
		w = 3
	}
	if h < 3 {
		h = 3
	}
	// Keep a roughly square sphere when terminal width is the bottleneck.
	if h > w*2 {
		h = w * 2
	}

	if o.cacheValid && o.cachedWidth == w && o.cachedTime == o.time {
		return o.cachedLines
	}

	lines := o.draw(w, h)
	o.cachedLines = lines
	o.cachedWidth = w
	o.cachedTime = o.time
	o.cacheValid = true
	return lines
}

func (o *Orb) draw(w, h int) []string {
	// Aspect: terminal cells are ~2× taller than wide (kx = 0.5).
	const kx = 0.5
	cx := float64(w) / 2
	cy := float64(h) / 2
	rx := math.Max(1, float64(w)/2-1)
	ry := math.Max(1, float64(h)/(2*kx)-1)
	radius := math.Min(rx, ry)
	r2 := radius * radius
	invR2 := 1 / r2
	invKx := 1 / kx

	speed := 1.0
	if o.rush {
		speed = 2.5
	}
	nChars := len(o.chars)
	nPal := len(o.palette)
	t := o.time

	lines := make([]string, h)
	var b strings.Builder
	b.Grow(w * 24) // rough: ANSI prefix + char + reset per cell

	for row := 0; row < h; row++ {
		b.Reset()
		py := (float64(row) - cy) * invKx
		py2 := py * py
		if py2 >= r2 {
			lines[row] = strings.Repeat(" ", w)
			continue
		}
		half := math.Sqrt(r2 - py2)
		x0 := int(math.Max(0, math.Floor(cx-half)))
		x1 := int(math.Min(float64(w-1), math.Ceil(cx+half)))

		if x0 > 0 {
			b.WriteString(strings.Repeat(" ", x0))
		}

		lastPi := -1
		for col := x0; col <= x1; col++ {
			px := float64(col) - cx
			d2 := px*px + py2
			if d2 >= r2 {
				if lastPi >= 0 {
					b.WriteString("\x1b[39m")
					lastPi = -1
				}
				b.WriteByte(' ')
				continue
			}
			falloff := 1 - d2*invR2
			g := o.noise.sample(float64(col), float64(row), t, speed) * falloff
			if g <= 0 {
				if lastPi >= 0 {
					b.WriteString("\x1b[39m")
					lastPi = -1
				}
				b.WriteByte(' ')
				continue
			}
			gi := int(math.Min(float64(nChars-1), math.Floor(g*float64(nChars))))
			pi := int(math.Min(float64(nPal-1), math.Floor(g*float64(nPal))))
			if pi != lastPi {
				b.WriteString(o.palette[pi])
				lastPi = pi
			}
			b.WriteRune(o.chars[gi])
		}
		if lastPi >= 0 {
			b.WriteString("\x1b[39m")
		}
		if x1 < w-1 {
			b.WriteString(strings.Repeat(" ", w-1-x1))
		}
		lines[row] = b.String()
	}
	return lines
}
