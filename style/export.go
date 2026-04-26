package style

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DetectTerminalBackground returns "dark" or "light" using COLORFGBG when present (simple heuristic).
func DetectTerminalBackground() string {
	colorfgbg := os.Getenv("COLORFGBG")
	if colorfgbg != "" {
		parts := strings.Split(colorfgbg, ";")
		if len(parts) >= 2 {
			bg, err := strconv.Atoi(parts[1])
			if err == nil {
				if bg < 8 {
					return "dark"
				}
				return "light"
			}
		}
	}
	return "dark"
}

// DefaultThemeName picks "light" or "dark" from [DetectTerminalBackground].
func DefaultThemeName() string {
	if DetectTerminalBackground() == "light" {
		return "light"
	}
	return "dark"
}

// ResolvedThemeColors returns theme colors as #RRGGBB (or fallback for empty / index colors) for CSS.
func ResolvedThemeColors(themeName string) (map[string]string, error) {
	if themeName == "" {
		themeName = DefaultThemeName()
	}
	tf, err := LoadThemeFile(themeName)
	if err != nil {
		return nil, err
	}
	resolved, err := resolveThemeColors(tf.Colors, tf.Vars)
	if err != nil {
		return nil, err
	}
	isLight := themeName == "light"
	defaultText := "#e5e5e7"
	if isLight {
		defaultText = "#000000"
	}
	out := make(map[string]string, len(resolved))
	for k, rc := range resolved {
		out[k] = resolvedToCSS(rc, defaultText)
	}
	return out, nil
}

// IsLightTheme reports whether userMessageBg resolves to a light background (luminance > 0.5).
func IsLightTheme(themeName string) bool {
	if themeName == "" {
		themeName = "dark"
	}
	var tf *ThemeFile
	var err error
	if data, ok := BuiltinThemeJSON(themeName); ok {
		tf, err = ParseThemeJSON(data)
	} else {
		path := filepath.Join(ThemesDir(), themeName+".json")
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return false
		}
		tf, err = ParseThemeJSON(raw)
	}
	if err != nil || tf == nil {
		return false
	}
	v, ok := tf.Colors["userMessageBg"]
	if !ok {
		return false
	}
	visited := make(map[string]bool)
	rc, err := resolveColorValueDeep(v, tf.Vars, visited)
	if err != nil {
		return false
	}
	hex := resolvedToCSS(rc, "#000000")
	if !strings.HasPrefix(hex, "#") || len(hex) != 7 {
		return false
	}
	r, err := strconv.ParseInt(hex[1:3], 16, 64)
	if err != nil {
		return false
	}
	g, err := strconv.ParseInt(hex[3:5], 16, 64)
	if err != nil {
		return false
	}
	b, err := strconv.ParseInt(hex[5:7], 16, 64)
	if err != nil {
		return false
	}
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255
	lum := 0.2126*rf + 0.7152*gf + 0.0722*bf
	return lum > 0.5
}

// ThemeExportColors holds optional HTML export palette entries.
type ThemeExportColors struct {
	PageBg *string
	CardBg *string
	InfoBg *string
}

// exportColorToCSS matches theme.ts getThemeExportColors: index → hex, no fallback for "".
func exportColorToCSS(rc ResolvedColor) string {
	if rc.IsIdx {
		return Ansi256ToHex(rc.Index)
	}
	if rc.Hex == "" {
		return ""
	}
	return rc.Hex
}

// exportResolveValue mirrors getThemeExportColors' resolve() in theme.ts:
// - optional $varName for vars lookup, unknown refs return the original string
// - "" and "#..." pass through; numbers become hex from the 256 palette.
func exportResolveValue(v any, vars map[string]any) (*string, error) {
	if v == nil {
		return nil, nil
	}
	if vars == nil {
		vars = map[string]any{}
	}
	if s, ok := v.(string); ok {
		if s == "" || strings.HasPrefix(s, "#") {
			return &s, nil
		}
		varName := strings.TrimPrefix(s, "$")
		if _, has := vars[varName]; has {
			visited := make(map[string]bool)
			rc, err := resolveVarRefs(varName, vars, visited)
			if err != nil {
				return nil, err
			}
			out := exportColorToCSS(rc)
			return &out, nil
		}
		return &s, nil
	}
	visited := make(map[string]bool)
	rc, err := resolveColorValueDeep(v, vars, visited)
	if err != nil {
		return nil, err
	}
	out := exportColorToCSS(rc)
	return &out, nil
}

// ExportColors loads explicit export.* colors from theme JSON, resolving vars when needed.
func ExportColors(themeName string) (*ThemeExportColors, error) {
	if themeName == "" {
		themeName = DefaultThemeName()
	}
	tf, err := LoadThemeFile(themeName)
	if err != nil {
		return nil, err
	}
	if tf.Export == nil {
		return &ThemeExportColors{}, nil
	}
	vars := tf.Vars
	if vars == nil {
		vars = map[string]any{}
	}
	out := &ThemeExportColors{}
	if tf.Export.PageBg != nil {
		s, err := exportResolveValue(tf.Export.PageBg, vars)
		if err != nil {
			return nil, fmt.Errorf("export.pageBg: %w", err)
		}
		out.PageBg = s
	}
	if tf.Export.CardBg != nil {
		s, err := exportResolveValue(tf.Export.CardBg, vars)
		if err != nil {
			return nil, fmt.Errorf("export.cardBg: %w", err)
		}
		out.CardBg = s
	}
	if tf.Export.InfoBg != nil {
		s, err := exportResolveValue(tf.Export.InfoBg, vars)
		if err != nil {
			return nil, fmt.Errorf("export.infoBg: %w", err)
		}
		out.InfoBg = s
	}
	return out, nil
}
