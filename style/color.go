package style

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// ColorMode selects truecolor vs 256-color ANSI output.
type ColorMode string

const (
	ColorModeTruecolor ColorMode = "truecolor"
	ColorMode256       ColorMode = "256color"
)

// DetectColorMode mirrors the TypeScript terminal heuristic.
func DetectColorMode() ColorMode {
	ct := os.Getenv("COLORTERM")
	if ct == "truecolor" || ct == "24bit" {
		return ColorModeTruecolor
	}
	if os.Getenv("WT_SESSION") != "" {
		return ColorModeTruecolor
	}
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" || term == "linux" {
		return ColorMode256
	}
	return ColorModeTruecolor
}

// ResolvedColor is either a hex/default string or a 256-color index.
type ResolvedColor struct {
	Hex   string
	Index int
	IsIdx bool
}

func colorToFgANSI(c ResolvedColor, mode ColorMode) (string, error) {
	if c.IsIdx {
		return fmt.Sprintf("\x1b[38;5;%dm", c.Index), nil
	}
	if c.Hex == "" {
		return "\x1b[39m", nil
	}
	if strings.HasPrefix(c.Hex, "#") {
		if mode == ColorMode256 {
			idx := hexTo256Approx(c.Hex)
			return fmt.Sprintf("\x1b[38;5;%dm", idx), nil
		}
		r, g, b, err := parseHexRGB(c.Hex)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b), nil
	}
	return "", fmt.Errorf("unsupported resolved color %q", c.Hex)
}

func colorToBgANSI(c ResolvedColor, mode ColorMode) (string, error) {
	if c.IsIdx {
		return fmt.Sprintf("\x1b[48;5;%dm", c.Index), nil
	}
	if c.Hex == "" {
		return "\x1b[49m", nil
	}
	if strings.HasPrefix(c.Hex, "#") {
		if mode == ColorMode256 {
			idx := hexTo256Approx(c.Hex)
			return fmt.Sprintf("\x1b[48;5;%dm", idx), nil
		}
		r, g, b, err := parseHexRGB(c.Hex)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b), nil
	}
	return "", fmt.Errorf("unsupported resolved color %q", c.Hex)
}

func parseHexRGB(s string) (r, g, b int, err error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("expected #RRGGBB")
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	return int(v >> 16), int((v >> 8) & 0xff), int(v & 0xff), nil
}

// hexTo256Approx maps #RRGGBB to the xterm 256-color cube (rough match).
func hexTo256Approx(hex string) int {
	r, g, b, err := parseHexRGB(hex)
	if err != nil {
		return 7
	}
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 238 {
			return 231
		}
		return 232 + (r-8)/10
	}
	r6 := quantize6(r)
	g6 := quantize6(g)
	b6 := quantize6(b)
	return 16 + 36*r6 + 6*g6 + b6
}

func quantize6(v int) int {
	if v < 48 {
		return 0
	}
	if v < 115 {
		return 1
	}
	return (v-35)/40 + 1
}

func resolveVarRefs(name string, vars map[string]any, visited map[string]bool) (ResolvedColor, error) {
	if visited[name] {
		return ResolvedColor{}, fmt.Errorf("circular variable reference: %s", name)
	}
	visited[name] = true
	defer delete(visited, name)
	v, ok := vars[name]
	if !ok {
		return ResolvedColor{}, fmt.Errorf("unknown variable %q", name)
	}
	return resolveColorValueDeep(v, vars, visited)
}

func resolveColorValueDeep(v any, vars map[string]any, visited map[string]bool) (ResolvedColor, error) {
	switch x := v.(type) {
	case string:
		if x == "" || strings.HasPrefix(x, "#") {
			return ResolvedColor{Hex: x}, nil
		}
		return resolveVarRefs(x, vars, visited)
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return ResolvedColor{}, err
		}
		if n < 0 || n > 255 {
			return ResolvedColor{}, fmt.Errorf("color index out of range: %d", n)
		}
		return ResolvedColor{Index: int(n), IsIdx: true}, nil
	case float64:
		if x != math.Trunc(x) {
			return ResolvedColor{}, fmt.Errorf("invalid color number %v", x)
		}
		n := int(x)
		if n < 0 || n > 255 {
			return ResolvedColor{}, fmt.Errorf("color index out of range: %d", n)
		}
		return ResolvedColor{Index: n, IsIdx: true}, nil
	default:
		return ResolvedColor{}, fmt.Errorf("unsupported color type %T", v)
	}
}

func resolveThemeColors(colors map[string]any, vars map[string]any) (map[string]ResolvedColor, error) {
	out := make(map[string]ResolvedColor, len(colors))
	visited := make(map[string]bool)
	for k, v := range colors {
		rc, err := resolveColorValueDeep(v, vars, visited)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
		out[k] = rc
	}
	return out, nil
}

// colorblindHueShift shifts green-ish diff additions toward blue (HSV space).
const colorblindHueShift = 60.0
const colorblindSatMul = 0.71

func applyColorblindAdjustment(hex string) (string, error) {
	if !strings.HasPrefix(hex, "#") {
		return hex, nil
	}
	r, g, b, err := parseHexRGB(hex)
	if err != nil {
		return hex, err
	}
	h, s, v := rgbToHSV(float64(r)/255, float64(g)/255, float64(b)/255)
	h = math.Mod(h+colorblindHueShift, 360)
	s *= colorblindSatMul
	if s > 1 {
		s = 1
	}
	nr, ng, nb := hsvToRGB(h, s, v)
	return fmt.Sprintf("#%02x%02x%02x", clamp255(nr), clamp255(ng), clamp255(nb)), nil
}

func clamp255(x float64) int {
	if x < 0 {
		return 0
	}
	if x > 255 {
		return 255
	}
	return int(math.Round(x))
}

func rgbToHSV(r, g, b float64) (h, s, v float64) {
	maxv := math.Max(r, math.Max(g, b))
	minv := math.Min(r, math.Min(g, b))
	d := maxv - minv
	v = maxv
	if maxv == 0 {
		return 0, 0, v
	}
	s = d / maxv
	if d == 0 {
		return 0, s, v
	}
	var hh float64
	switch maxv {
	case r:
		hh = math.Mod((g-b)/d, 6)
	case g:
		hh = (b-r)/d + 2
	default:
		hh = (r-g)/d + 4
	}
	h = hh * 60
	if h < 0 {
		h += 360
	}
	return h, s, v
}

func hsvToRGB(h, s, v float64) (r, g, b float64) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r1, g1, b1 float64
	switch {
	case h < 60:
		r1, g1, b1 = c, x, 0
	case h < 120:
		r1, g1, b1 = x, c, 0
	case h < 180:
		r1, g1, b1 = 0, c, x
	case h < 240:
		r1, g1, b1 = 0, x, c
	case h < 300:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}
	return (r1 + m) * 255, (g1 + m) * 255, (b1 + m) * 255
}

// Ansi256ToHex converts a 256-color index to #RRGGBB for HTML/CSS.
func Ansi256ToHex(index int) string {
	basic := []string{
		"#000000", "#800000", "#008000", "#808000", "#000080", "#800080", "#008080", "#c0c0c0",
		"#808080", "#ff0000", "#00ff00", "#ffff00", "#0000ff", "#ff00ff", "#00ffff", "#ffffff",
	}
	if index < 16 {
		if index < 0 {
			index = 0
		}
		return basic[index]
	}
	if index < 232 {
		cube := index - 16
		r := cube / 36
		g := (cube % 36) / 6
		b := cube % 6
		toHex := func(n int) int {
			if n == 0 {
				return 0
			}
			return 55 + n*40
		}
		return fmt.Sprintf("#%02x%02x%02x", toHex(r), toHex(g), toHex(b))
	}
	gray := 8 + (index-232)*10
	return fmt.Sprintf("#%02x%02x%02x", gray, gray, gray)
}

func resolvedToCSS(rc ResolvedColor, emptyFallback string) string {
	if rc.IsIdx {
		return Ansi256ToHex(rc.Index)
	}
	if rc.Hex == "" {
		return emptyFallback
	}
	return rc.Hex
}
