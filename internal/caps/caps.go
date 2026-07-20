// Package caps detects and enables terminal capabilities via VT queries.
//
// Detection and enabling are separate steps: callers write Query, feed stdin
// sequences through Feed until Done, then write Enable. Unsupported terminals
// typically ignore unknown queries and leave Caps fields false.
package caps

import "strings"

// Caps is a snapshot of features discovered for the current terminal.
// Zero value means nothing was detected; callers must treat false as
// "do not use the feature".
type Caps struct {
	KittyKeyboard  bool // CSI u / Kitty keyboard protocol
	KittyGraphics  bool // Kitty graphics protocol (not probed yet)
	TrueColor      bool // 24-bit color (not probed yet)
	SyncOutput     bool // Mode 2026 (not probed yet)
	UnicodeCore    bool // Mode 2027 (not probed yet)
	BracketedPaste bool // Mode 2004 (not probed yet)
}

const (
	queryKittyKeyboard = "\x1b[?u"
	queryPrimaryDA     = "\x1b[c"

	// enableKittyKeyboard pushes Kitty keyboard flags:
	// 1 = disambiguate escape codes
	// 2 = report event types (press/repeat/release)
	// 4 = report alternate keys (shifted / base layout)
	enableKittyKeyboard  = "\x1b[>7u"
	disableKittyKeyboard = "\x1b[<u"
)

// Detector runs a progressive capability probe.
//
// Phase 1 probes Kitty keyboard support and uses Primary DA (CSI c) as a
// barrier so callers know when prior query responses have arrived (or been
// ignored). A Detector is not safe for concurrent use.
type Detector struct {
	caps Caps
	done bool
}

// New returns a Detector with all capabilities unset.
func New() *Detector {
	return &Detector{}
}

// Query returns the byte sequence to write to the terminal to start probing.
// The Primary DA request is last and acts as an end-of-batch barrier.
func (d *Detector) Query() []byte {
	return []byte(queryKittyKeyboard + queryPrimaryDA)
}

// Feed consumes one parsed input sequence during the probe window.
// It returns true when the sequence was a probe response and must not be
// delivered to the application as input.
//
// After Done reports true, Feed is a no-op and always returns false.
func (d *Detector) Feed(seq string) bool {
	if d.done {
		return false
	}
	switch {
	case isKittyKeyboardReply(seq):
		d.caps.KittyKeyboard = true
		return true
	case isPrimaryDAReply(seq):
		d.done = true
		return true
	default:
		return false
	}
}

// Complete marks probing finished without waiting for Primary DA.
// Call this on timeout so Enable may still run with whatever Caps were set.
func (d *Detector) Complete() {
	d.done = true
}

// Done reports whether probing has finished (Primary DA received or Complete).
func (d *Detector) Done() bool {
	return d.done
}

// Caps returns a copy of the capabilities discovered so far.
func (d *Detector) Caps() Caps {
	return d.caps
}

// Enable returns sequences that turn on features present in Caps.
// Call only after Done is true. Empty means nothing to enable.
func (d *Detector) Enable() []byte {
	if !d.caps.KittyKeyboard {
		return nil
	}
	return []byte(enableKittyKeyboard)
}

// Disable returns sequences that reverse Enable for features that were on.
// Safe to call even if Enable was never written.
func (d *Detector) Disable() []byte {
	if !d.caps.KittyKeyboard {
		return nil
	}
	return []byte(disableKittyKeyboard)
}

// isKittyKeyboardReply reports whether seq is CSI ? <flags> u.
func isKittyKeyboardReply(seq string) bool {
	const prefix = "\x1b[?"
	if !strings.HasPrefix(seq, prefix) || !strings.HasSuffix(seq, "u") {
		return false
	}
	flags := seq[len(prefix) : len(seq)-1]
	if flags == "" {
		return false
	}
	return isDigits(flags)
}

// isPrimaryDAReply reports whether seq is a Primary DA response (CSI ? … c).
func isPrimaryDAReply(seq string) bool {
	const prefix = "\x1b[?"
	if len(seq) < len(prefix)+1 || !strings.HasPrefix(seq, prefix) || seq[len(seq)-1] != 'c' {
		return false
	}
	body := seq[len(prefix) : len(seq)-1]
	if body == "" {
		return true // ESC [ ? c
	}
	for _, r := range body {
		if r != ';' && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
