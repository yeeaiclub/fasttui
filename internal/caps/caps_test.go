package caps

import (
	"bytes"
	"testing"
)

func TestDetectorQueryEndsWithPrimaryDA(t *testing.T) {
	d := New()
	q := d.Query()
	if !bytes.Contains(q, []byte(queryKittyKeyboard)) {
		t.Fatalf("Query() missing Kitty keyboard query: %q", q)
	}
	if !bytes.Contains(q, []byte(queryDECRQMSync)) {
		t.Fatalf("Query() missing sync DECRQM: %q", q)
	}
	if !bytes.Contains(q, []byte(queryDECRQMUnicode)) {
		t.Fatalf("Query() missing unicode DECRQM: %q", q)
	}
	if !bytes.Contains(q, []byte(queryDECRQMBP)) {
		t.Fatalf("Query() missing bracketed paste DECRQM: %q", q)
	}
	if !bytes.Contains(q, []byte(queryKittyGraphics)) {
		t.Fatalf("Query() missing kitty graphics query: %q", q)
	}
	if !bytes.Contains(q, []byte(queryXTGetTCapRGB)) {
		t.Fatalf("Query() missing XTGETTCAP RGB: %q", q)
	}
	if !bytes.HasSuffix(q, []byte(queryPrimaryDA)) {
		t.Fatalf("Query() must end with Primary DA barrier: %q", q)
	}
}

func TestDetectorKittyThenDA1(t *testing.T) {
	d := New()
	// Avoid COLORTERM influencing Enable checks in this unit test.
	d.caps.TrueColor = false

	if got := d.Feed("\x1b[?1u"); !got {
		t.Fatal("Feed(Kitty reply) = false, want true")
	}
	if d.Done() {
		t.Fatal("Done() = true before Primary DA")
	}
	if !d.Caps().KittyKeyboard {
		t.Fatal("Caps.KittyKeyboard = false, want true")
	}

	if got := d.Feed("\x1b[?1;2c"); !got {
		t.Fatal("Feed(Primary DA) = false, want true")
	}
	if !d.Done() {
		t.Fatal("Done() = false after Primary DA")
	}

	enable := d.Enable()
	if !bytes.Equal(enable, []byte(enableKittyKeyboard+enableBracketedPaste)) {
		t.Fatalf("Enable() = %q, want kitty+bp", enable)
	}
	disable := d.Disable()
	if !bytes.Equal(disable, []byte(disableBracketedPaste+disableKittyKeyboard)) {
		t.Fatalf("Disable() = %q, want bp+kitty", disable)
	}

	// After Done, further sequences are not consumed.
	if d.Feed("\x1b[?1u") {
		t.Fatal("Feed after Done returned true, want false")
	}
}

func TestDetectorFullProbe(t *testing.T) {
	d := New()
	d.caps.TrueColor = false

	replies := []string{
		"\x1b[?2026;2$y",                   // sync reset = supported
		"\x1b[?2027;1$y",                   // unicode set = supported
		"\x1b[?2004;1$y",                   // bracketed paste set
		"\x1b[?7u",                         // kitty keyboard
		"\x1b_Gi=1;OK\x1b\\",               // kitty graphics
		"\x1bP1+r524742=524742\x1b\\",      // XTGETTCAP RGB
		"\x1b[?61;6;22c",                   // Primary DA barrier
	}
	for _, seq := range replies {
		if !d.Feed(seq) {
			t.Fatalf("Feed(%q) = false, want true", seq)
		}
	}
	if !d.Done() {
		t.Fatal("Done() = false after Primary DA")
	}
	c := d.Caps()
	if !c.SyncOutput || !c.UnicodeCore || !c.BracketedPaste ||
		!c.KittyKeyboard || !c.KittyGraphics || !c.TrueColor {
		t.Fatalf("Caps incomplete: %+v", c)
	}

	enable := d.Enable()
	wantEnable := enableKittyKeyboard + enableUnicodeCore + enableBracketedPaste
	if string(enable) != wantEnable {
		t.Fatalf("Enable() = %q, want %q", enable, wantEnable)
	}
	disable := d.Disable()
	wantDisable := disableBracketedPaste + disableUnicodeCore + disableKittyKeyboard
	if string(disable) != wantDisable {
		t.Fatalf("Disable() = %q, want %q", disable, wantDisable)
	}
}

func TestDetectorDECRPMUnsupported(t *testing.T) {
	d := New()
	d.caps.TrueColor = false

	if !d.Feed("\x1b[?2026;0$y") {
		t.Fatal("Feed(DECRPM 0) = false, want true")
	}
	if !d.Feed("\x1b[?2027;4$y") {
		t.Fatal("Feed(DECRPM 4) = false, want true")
	}
	if d.Caps().SyncOutput || d.Caps().UnicodeCore {
		t.Fatalf("unsupported DECRPM should leave caps false: %+v", d.Caps())
	}
}

func TestDetectorXTGetTCapFailure(t *testing.T) {
	d := New()
	d.caps.TrueColor = false

	if !d.Feed("\x1bP0+r524742\x1b\\") {
		t.Fatal("Feed(XTGETTCAP fail) = false, want true")
	}
	if d.Caps().TrueColor {
		t.Fatal("TrueColor set on failed XTGETTCAP")
	}
}

func TestDetectorTimeoutWithoutKitty(t *testing.T) {
	d := New()
	d.caps.TrueColor = false
	d.Complete()
	if !d.Done() {
		t.Fatal("Done() = false after Complete")
	}
	if d.Caps().KittyKeyboard {
		t.Fatal("Caps.KittyKeyboard = true, want false")
	}
	// Bracketed paste is always enabled even with an empty probe.
	if got := d.Enable(); string(got) != enableBracketedPaste {
		t.Fatalf("Enable() = %q, want %q", got, enableBracketedPaste)
	}
	if got := d.Disable(); string(got) != disableBracketedPaste {
		t.Fatalf("Disable() = %q, want %q", got, disableBracketedPaste)
	}
}

func TestNewCOLORTERM(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	d := New()
	if !d.Caps().TrueColor {
		t.Fatal("COLORTERM=truecolor should set TrueColor")
	}

	t.Setenv("COLORTERM", "")
	d2 := New()
	if d2.Caps().TrueColor {
		t.Fatal("empty COLORTERM should leave TrueColor false")
	}
}

func TestEnableVHSRecordDisablesKitty(t *testing.T) {
	t.Setenv("VHS_RECORD", "1")
	d := New()
	d.caps.TrueColor = false
	d.Feed("\x1b[?7u")
	d.Feed("\x1b[?1;2c")
	if !d.Caps().KittyKeyboard {
		t.Fatal("KittyKeyboard should be true before Enable")
	}
	enable := d.Enable()
	if bytes.Contains(enable, []byte(enableKittyKeyboard)) {
		t.Fatalf("Enable under VHS_RECORD must not push Kitty: %q", enable)
	}
	if d.Caps().KittyKeyboard {
		t.Fatal("Caps.KittyKeyboard should be false after Enable under VHS_RECORD")
	}
	if !bytes.Equal(enable, []byte(enableBracketedPaste)) {
		t.Fatalf("Enable() = %q, want bracketed paste only", enable)
	}
}

func TestFeed(t *testing.T) {
	tests := []struct {
		name      string
		seq       string
		wantCons  bool
		wantKitty bool
		wantDone  bool
	}{
		{
			name:      "kitty flags",
			seq:       "\x1b[?7u",
			wantCons:  true,
			wantKitty: true,
		},
		{
			name:     "primary da",
			seq:      "\x1b[?61;6;22c",
			wantCons: true,
			wantDone: true,
		},
		{
			name:     "empty primary da",
			seq:      "\x1b[?c",
			wantCons: true,
			wantDone: true,
		},
		{
			name:     "decrpm sync",
			seq:      "\x1b[?2026;2$y",
			wantCons: true,
		},
		{
			name:     "kitty graphics",
			seq:      "\x1b_Gi=1;OK\x1b\\",
			wantCons: true,
		},
		{
			name: "plain key",
			seq:  "a",
		},
		{
			name: "csi u key event not query reply",
			seq:  "\x1b[13;2u",
		},
		{
			name: "incomplete kitty",
			seq:  "\x1b[?u",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New()
			d.caps.TrueColor = false
			got := d.Feed(tt.seq)
			if got != tt.wantCons {
				t.Fatalf("Feed(%q) = %v, want %v", tt.seq, got, tt.wantCons)
			}
			if d.Caps().KittyKeyboard != tt.wantKitty {
				t.Fatalf("KittyKeyboard = %v, want %v", d.Caps().KittyKeyboard, tt.wantKitty)
			}
			if d.Done() != tt.wantDone {
				t.Fatalf("Done() = %v, want %v", d.Done(), tt.wantDone)
			}
		})
	}
}
