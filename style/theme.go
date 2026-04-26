package style

import (
	"fmt"
	"maps"
	"strings"
)

// CreateThemeOptions configures how a [ThemeFile] is compiled to terminal ANSI.
type CreateThemeOptions struct {
	Mode           ColorMode
	SymbolPreset   *SymbolPreset // if non-nil, overrides theme JSON symbols.preset
	ColorBlindMode bool
}

// Theme holds resolved ANSI sequences and the active symbol table.
type Theme struct {
	fg      map[string]string
	bg      map[string]string
	mode    ColorMode
	preset  SymbolPreset
	symbols map[string]string
}

// NewTheme builds a [Theme] from validated theme data.
func NewTheme(tf *ThemeFile, opt CreateThemeOptions) (*Theme, error) {
	mode := opt.Mode
	if mode == "" {
		mode = DetectColorMode()
	}
	resolved, err := resolveThemeColors(tf.Colors, tf.Vars)
	if err != nil {
		return nil, err
	}
	if opt.ColorBlindMode {
		if rc, ok := resolved["toolDiffAdded"]; ok && !rc.IsIdx && strings.HasPrefix(rc.Hex, "#") {
			h, err := applyColorblindAdjustment(rc.Hex)
			if err == nil {
				resolved["toolDiffAdded"] = ResolvedColor{Hex: h}
			}
		}
	}
	fg := make(map[string]string)
	bg := make(map[string]string)
	for key, rc := range resolved {
		if _, isBg := themeBgKeySet[key]; isBg {
			s, err := colorToBgANSI(rc, mode)
			if err != nil {
				return nil, fmt.Errorf("bg %s: %w", key, err)
			}
			bg[key] = s
			continue
		}
		if _, isFg := themeFgKeySet[key]; isFg {
			s, err := colorToFgANSI(rc, mode)
			if err != nil {
				return nil, fmt.Errorf("fg %s: %w", key, err)
			}
			fg[key] = s
		}
	}
	preset := SymbolPresetUnicode
	if tf.Symbols != nil && tf.Symbols.Preset != nil {
		preset = *tf.Symbols.Preset
	}
	if opt.SymbolPreset != nil {
		preset = *opt.SymbolPreset
	}
	base, ok := symbolPresetMaps[preset]
	if !ok {
		return nil, fmt.Errorf("unknown symbol preset %q", preset)
	}
	symbols := maps.Clone(base)
	if tf.Symbols != nil {
		for k, v := range tf.Symbols.Overrides {
			if _, exists := symbols[k]; exists {
				symbols[k] = v
			}
		}
	}
	return &Theme{
		fg: fg, bg: bg, mode: mode, preset: preset, symbols: symbols,
	}, nil
}

// Fg paints text with a foreground theme token and resets only the foreground color.
func (t *Theme) Fg(color ThemeColor, text string) string {
	ansi, ok := t.fg[string(color)]
	if !ok {
		panic(fmt.Sprintf("unknown theme color: %s", color))
	}
	return ansi + text + "\x1b[39m"
}

// Bg paints text with a background theme token and resets only the background color.
func (t *Theme) Bg(bg ThemeBg, text string) string {
	ansi, ok := t.bg[string(bg)]
	if !ok {
		panic(fmt.Sprintf("unknown theme background: %s", bg))
	}
	return ansi + text + "\x1b[49m"
}

// Bold wraps text with ANSI bold (resets with 22).
func (t *Theme) Bold(text string) string {
	return "\x1b[1m" + text + "\x1b[22m"
}

// Italic wraps text with ANSI italic (resets with 23).
func (t *Theme) Italic(text string) string {
	return "\x1b[3m" + text + "\x1b[23m"
}

// Underline wraps text with ANSI underline.
func (t *Theme) Underline(text string) string {
	return "\x1b[4m" + text + "\x1b[24m"
}

// Strikethrough wraps text with ANSI strikethrough.
func (t *Theme) Strikethrough(text string) string {
	return "\x1b[9m" + text + "\x1b[29m"
}

// Inverse wraps text with ANSI reverse video.
func (t *Theme) Inverse(text string) string {
	return "\x1b[7m" + text + "\x1b[27m"
}

// FgANSI returns the raw ANSI sequence for a foreground token.
func (t *Theme) FgANSI(color ThemeColor) string {
	ansi, ok := t.fg[string(color)]
	if !ok {
		panic(fmt.Sprintf("unknown theme color: %s", color))
	}
	return ansi
}

// BgANSI returns the raw ANSI sequence for a background token.
func (t *Theme) BgANSI(bg ThemeBg) string {
	ansi, ok := t.bg[string(bg)]
	if !ok {
		panic(fmt.Sprintf("unknown theme background: %s", bg))
	}
	return ansi
}

// ColorMode returns the color mode used to build this theme.
func (t *Theme) ColorMode() ColorMode { return t.mode }

// Symbol returns a glyph for the given logical key (e.g. "status.success").
func (t *Theme) Symbol(key string) string {
	return t.symbols[key]
}

// StyledSymbol returns a symbol wrapped in a foreground color.
func (t *Theme) StyledSymbol(key string, color ThemeColor) string {
	return t.Fg(color, t.Symbol(key))
}

// SymbolPreset returns the active symbol preset.
func (t *Theme) SymbolPreset() SymbolPreset { return t.preset }

// SpinnerFrames returns spinner animation frames for the given spinner kind.
func (t *Theme) SpinnerFrames(kind SpinnerType) []string {
	if m, ok := spinnerFrames[t.preset]; ok {
		if frames, ok := m[kind]; ok {
			out := make([]string, len(frames))
			copy(out, frames)
			return out
		}
	}
	return nil
}

// ThinkingBorderColor returns a function that colors text for a thinking level (off, minimal, low, ...).
func (t *Theme) ThinkingBorderColor(level string) func(string) string {
	switch strings.ToLower(level) {
	case "off":
		return func(s string) string { return t.Fg(ColorThinkingOff, s) }
	case "minimal":
		return func(s string) string { return t.Fg(ColorThinkingMinimal, s) }
	case "low":
		return func(s string) string { return t.Fg(ColorThinkingLow, s) }
	case "medium":
		return func(s string) string { return t.Fg(ColorThinkingMedium, s) }
	case "high":
		return func(s string) string { return t.Fg(ColorThinkingHigh, s) }
	case "xhigh":
		return func(s string) string { return t.Fg(ColorThinkingXhigh, s) }
	default:
		return func(s string) string { return t.Fg(ColorThinkingOff, s) }
	}
}

// BashModeBorderColor colors text with the bash-mode border token.
func (t *Theme) BashModeBorderColor() func(string) string {
	return func(s string) string { return t.Fg(ColorBashMode, s) }
}

// PythonModeBorderColor colors text with the python-mode border token.
func (t *Theme) PythonModeBorderColor() func(string) string {
	return func(s string) string { return t.Fg(ColorPythonMode, s) }
}

// LangIcon returns the icon string for a language id or file extension token.
func (t *Theme) LangIcon(lang string) string {
	if lang == "" {
		return t.Symbol("lang.default")
	}
	key, ok := langIconKeys[strings.ToLower(lang)]
	if !ok {
		return t.Symbol("lang.default")
	}
	return t.Symbol(key)
}

// InputCursor returns the cursor glyph used in text inputs (matches TS ascii vs unicode rule).
func (t *Theme) InputCursor() string {
	if t.preset == SymbolPresetAscii {
		return "|"
	}
	return "▏"
}
