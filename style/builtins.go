package style

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// BuiltinThemeJSON returns raw JSON bytes for an embedded theme name (e.g. "dark", "alabaster").
func BuiltinThemeJSON(name string) ([]byte, bool) {
	b, ok := builtinThemeData[name]
	return b, ok
}

// BuiltinThemeNames returns sorted embedded theme base names (filename without .json).
func BuiltinThemeNames() []string {
	out := make([]string, 0, len(builtinThemeData))
	for k := range builtinThemeData {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

var builtinThemeData map[string][]byte

func init() {
	builtinThemeData = make(map[string][]byte)
	_ = fs.WalkDir(themeEmbedded, "theme", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		base := filepath.Base(path)
		if base == "theme-schema.json" {
			return nil
		}
		data, err := themeEmbedded.ReadFile(path)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(base, ".json")
		builtinThemeData[name] = data
		return nil
	})
}
