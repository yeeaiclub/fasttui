package fasttui

import (
	"strings"
	"testing"
)

func TestVisibleWidthASCIIWithAnsi(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"\x1b[31mhello\x1b[0m", 5},
		{"\x1b[31mhello\x1b[0m world", 11},
		{"task_name", 9},
		{"\x1b[1mtask_name\x1b[0m", 9},
		{"a\tb", 5},
	}

	for _, tt := range tests {
		got, ok := visibleWidthASCIIWithAnsi(tt.input)
		if !ok {
			t.Fatalf("visibleWidthASCIIWithAnsi(%q) = not ok", tt.input)
		}
		if got != tt.want {
			t.Fatalf("visibleWidthASCIIWithAnsi(%q) = %d, want %d", tt.input, got, tt.want)
		}
		if VisibleWidth(tt.input) != tt.want {
			t.Fatalf("VisibleWidth(%q) = %d, want %d", tt.input, VisibleWidth(tt.input), tt.want)
		}
	}
}

func TestScanTextSegments_PreservesContent(t *testing.T) {
	input := "\x1b[31mhello\x1b[0m世界"
	segments := scanTextSegments(input)

	var b strings.Builder
	for _, seg := range segments {
		b.WriteString(seg.value)
	}
	if b.String() != input {
		t.Fatalf("rejoined segments = %q, want %q", b.String(), input)
	}
}

func TestSplitASCIITokensWithAnsi(t *testing.T) {
	input := "\x1b[31mhello\x1b[0m world"
	tokens := splitASCIITokensWithAnsi(input)
	if len(tokens) != 3 {
		t.Fatalf("got %d tokens: %#v", len(tokens), tokens)
	}
	if tokens[0] != "\x1b[31mhello\x1b[0m" {
		t.Fatalf("token[0] = %q", tokens[0])
	}
	if tokens[1] != " " {
		t.Fatalf("token[1] = %q", tokens[1])
	}
	if tokens[2] != "world" {
		t.Fatalf("token[2] = %q", tokens[2])
	}
}
