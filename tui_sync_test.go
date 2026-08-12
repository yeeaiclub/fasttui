package fasttui

import (
	"strings"
	"testing"
)

type syncCapTerminal struct {
	recordingTerminal
	sync bool
}

func (s *syncCapTerminal) SupportsSyncOutput() bool { return s.sync }

func TestWriteSyncGatedByCapability(t *testing.T) {
	t.Run("unsupported", func(t *testing.T) {
		term := &syncCapTerminal{}
		tui := NewTUI(term, false)
		var b strings.Builder
		tui.writeSyncBegin(&b)
		tui.writeSyncEnd(&b)
		if b.String() != "" {
			t.Fatalf("got %q, want empty when SyncOutput unsupported", b.String())
		}
	})

	t.Run("supported", func(t *testing.T) {
		term := &syncCapTerminal{sync: true}
		tui := NewTUI(term, false)
		var b strings.Builder
		tui.writeSyncBegin(&b)
		tui.writeSyncEnd(&b)
		want := SyncOutputBegin + SyncOutputEnd
		if b.String() != want {
			t.Fatalf("got %q, want %q", b.String(), want)
		}
	})
}
