package fasttui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWrapAnsiText_CJKRespectsWidthAndValidUTF8(t *testing.T) {
	const col = 12
	s := "WRAP_MIX " + strings.Repeat("混合Mixed中文English数字123标点。", 4)
	lines := WrapAnsiText(s, col)
	for i, line := range lines {
		vw := VisibleWidth(line)
		if vw > col {
			t.Fatalf("line %d exceeds width: vw=%d want<=%d line=%q", i, vw, col, line)
		}
		if !utf8.ValidString(line) {
			t.Fatalf("line %d invalid UTF-8: %q", i, line)
		}
	}
	rejoined := strings.Join(lines, "")
	if !strings.Contains(rejoined, "WRAP_MIX") || !strings.Contains(rejoined, "混合") {
		t.Fatalf("unexpected content after wrap: %q", rejoined)
	}
}

func TestSplitIntoTokensWithAnsi_PreservesUTF8(t *testing.T) {
	s := "a中b"
	toks := splitIntoTokensWithAnsi(s)
	if len(toks) != 1 {
		t.Fatalf("want single token, got %d: %#v", len(toks), toks)
	}
	if !utf8.ValidString(toks[0]) || toks[0] != s {
		t.Fatalf("token = %q want %q", toks[0], s)
	}
}
