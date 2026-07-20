package caps

import (
	"bytes"
	"testing"
)

func TestDetectorQueryEndsWithPrimaryDA(t *testing.T) {
	d := New()
	q := d.Query()
	if !bytes.HasPrefix(q, []byte(queryKittyKeyboard)) {
		t.Fatalf("Query() missing Kitty query: %q", q)
	}
	if !bytes.HasSuffix(q, []byte(queryPrimaryDA)) {
		t.Fatalf("Query() must end with Primary DA barrier: %q", q)
	}
}

func TestDetectorKittyThenDA1(t *testing.T) {
	d := New()

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
	if !bytes.Equal(enable, []byte(enableKittyKeyboard)) {
		t.Fatalf("Enable() = %q, want %q", enable, enableKittyKeyboard)
	}
	disable := d.Disable()
	if !bytes.Equal(disable, []byte(disableKittyKeyboard)) {
		t.Fatalf("Disable() = %q, want %q", disable, disableKittyKeyboard)
	}

	// After Done, further sequences are not consumed.
	if d.Feed("\x1b[?1u") {
		t.Fatal("Feed after Done returned true, want false")
	}
}

func TestDetectorTimeoutWithoutKitty(t *testing.T) {
	d := New()
	d.Complete()
	if !d.Done() {
		t.Fatal("Done() = false after Complete")
	}
	if d.Caps().KittyKeyboard {
		t.Fatal("Caps.KittyKeyboard = true, want false")
	}
	if got := d.Enable(); got != nil {
		t.Fatalf("Enable() = %q, want nil", got)
	}
	if got := d.Disable(); got != nil {
		t.Fatalf("Disable() = %q, want nil", got)
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
