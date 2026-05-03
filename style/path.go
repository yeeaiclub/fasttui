package style

import (
	"path/filepath"
	"strings"
)

// LanguageFromPath returns the highlight grammar id for a file path (from extension), or "" if unknown.
func LanguageFromPath(path string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if ext == "" {
		return ""
	}
	if hl, _, ok := LangByExtension(ext); ok {
		return hl
	}
	return ""
}
