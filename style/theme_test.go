package style

import (
	"slices"
	"strings"
	"testing"
)

func TestParseDarkBuiltin(t *testing.T) {
	data, ok := BuiltinThemeJSON("dark")
	if !ok {
		t.Fatal("missing embedded dark theme")
	}
	tf, err := ParseThemeJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if tf.Name != "dark" {
		t.Fatalf("name: got %q", tf.Name)
	}
	th, err := NewTheme(tf, WithColorMode(ColorModeTruecolor))
	if err != nil {
		t.Fatal(err)
	}
	s := th.Fg(ColorAccent, "x")
	if !strings.HasPrefix(s, "\x1b[38;2;") {
		t.Fatalf("expected truecolor fg, got %q", s)
	}
}

func TestListThemeNamesIncludesDark(t *testing.T) {
	names, err := ListThemeNames()
	if err != nil {
		t.Fatal(err)
	}
	found := slices.Contains(names, "dark")
	if !found {
		t.Fatalf("expected dark in %v", names)
	}
}

func TestExportColorsLight(t *testing.T) {
	ec, err := ExportColors("light")
	if err != nil {
		t.Fatal(err)
	}
	if ec.PageBg == nil || *ec.PageBg != "#f8f8f8" {
		t.Fatalf("pageBg: %v", ec.PageBg)
	}
	if ec.CardBg == nil || *ec.CardBg != "#ffffff" {
		t.Fatalf("cardBg: %v", ec.CardBg)
	}
	if ec.InfoBg == nil || *ec.InfoBg != "#fffae6" {
		t.Fatalf("infoBg: %v", ec.InfoBg)
	}
}
