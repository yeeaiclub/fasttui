package style

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ThemesDir returns the directory for user-defined themes (override with FASTTUI_THEMES_DIR).
func ThemesDir() string {
	if d := os.Getenv("FASTTUI_THEMES_DIR"); d != "" {
		return d
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "fasttui", "themes")
}

// LoadThemeFile loads a theme by name from embedded builtins first, then ThemesDir.
func LoadThemeFile(name string) (*ThemeFile, error) {
	if data, ok := BuiltinThemeJSON(name); ok {
		return ParseThemeJSON(data)
	}
	path := filepath.Join(ThemesDir(), name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("theme not found: %s", name)
		}
		return nil, err
	}
	return ParseThemeJSON(data)
}

// LoadTheme is shorthand for [LoadThemeFile] + [NewTheme] with options.
func LoadTheme(name string, opt CreateThemeOptions) (*Theme, error) {
	tf, err := LoadThemeFile(name)
	if err != nil {
		return nil, err
	}
	return NewTheme(tf, opt)
}

// ListThemeNames returns builtin names plus any *.json in [ThemesDir] (sorted, unique).
func ListThemeNames() ([]string, error) {
	seen := make(map[string]struct{})
	for _, n := range BuiltinThemeNames() {
		seen[n] = struct{}{}
	}
	entries, err := os.ReadDir(ThemesDir())
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if filepath.Ext(name) != ".json" {
				continue
			}
			seen[strings.TrimSuffix(name, ".json")] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}
