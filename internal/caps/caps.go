// Package caps detects and enables terminal capabilities via VT queries.
//
// Detection and enabling are separate steps: callers write Query, feed stdin
// sequences through Feed until Done, then write Enable. Unsupported terminals
// typically ignore unknown queries and leave Caps fields false.
//
// Probe set mirrors libvaxis (DECRQM, Kitty keyboard/graphics, Primary DA
// barrier), plus XTGETTCAP "RGB" / COLORTERM for truecolor and DECRQM for
// synchronized output and bracketed paste.
package caps

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Caps is a snapshot of features discovered for the current terminal.
// Zero value means nothing was detected; callers must treat false as
// "do not use the feature".
type Caps struct {
	KittyKeyboard  bool // CSI u / Kitty keyboard protocol
	KittyGraphics  bool // Kitty graphics protocol
	TrueColor      bool // 24-bit color (COLORTERM or XTGETTCAP RGB)
	SyncOutput     bool // Mode 2026
	UnicodeCore    bool // Mode 2027
	BracketedPaste bool // Mode 2004
}

const (
	queryDECRQMSync    = "\x1b[?2026$p"
	queryDECRQMUnicode = "\x1b[?2027$p"
	queryDECRQMBP      = "\x1b[?2004$p"
	queryKittyKeyboard = "\x1b[?u"
	queryKittyGraphics = "\x1b_Gi=1,a=q\x1b\\"
	queryXTGetTCapRGB  = "\x1bP+q524742\x1b\\" // DCS + q "RGB"
	queryPrimaryDA     = "\x1b[c"

	// enableKittyKeyboard pushes Kitty keyboard flags:
	// 1 = disambiguate escape codes
	// 2 = report event types (press/repeat/release)
	// 4 = report alternate keys (shifted / base layout)
	enableKittyKeyboard  = "\x1b[>7u"
	disableKittyKeyboard = "\x1b[<u"

	enableUnicodeCore  = "\x1b[?2027h"
	disableUnicodeCore = "\x1b[?2027l"

	enableBracketedPaste  = "\x1b[?2004h"
	disableBracketedPaste = "\x1b[?2004l"
)

// Detector runs a progressive capability probe.
//
// Queries are flushed in one batch ending with Primary DA (CSI c) as a
// barrier so callers know when prior responses have arrived (or been
// ignored). Feed may run concurrently with Done / Caps / Complete
// (typical when stdin is read on a background goroutine).
type Detector struct {
	mu   sync.Mutex
	caps Caps
	done bool
}

// New returns a Detector with all capabilities unset, except TrueColor when
// COLORTERM already advertises 24-bit color.
func New() *Detector {
	d := &Detector{}
	if ct := os.Getenv("COLORTERM"); ct == "truecolor" || ct == "24bit" {
		d.caps.TrueColor = true
	}
	return d
}

// Query returns the byte sequence to write to the terminal to start probing.
// The Primary DA request is last and acts as an end-of-batch barrier.
func (d *Detector) Query() []byte {
	return []byte(
		queryDECRQMSync +
			queryDECRQMUnicode +
			queryDECRQMBP +
			queryKittyKeyboard +
			queryKittyGraphics +
			queryXTGetTCapRGB +
			queryPrimaryDA,
	)
}

// Feed consumes one parsed input sequence during the probe window.
// It returns true when the sequence was a probe response and must not be
// delivered to the application as input.
//
// After Done reports true, Feed is a no-op and always returns false.
func (d *Detector) Feed(seq string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.done {
		return false
	}
	switch {
	case isKittyKeyboardReply(seq):
		d.caps.KittyKeyboard = true
		return true
	case isKittyGraphicsReply(seq):
		d.caps.KittyGraphics = true
		return true
	case isXTGetTCapReply(seq):
		if isXTGetTCapRGBSuccess(seq) {
			d.caps.TrueColor = true
		}
		return true
	case feedDECRPM(&d.caps, seq):
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
	d.mu.Lock()
	defer d.mu.Unlock()
	d.done = true
}

// Done reports whether probing has finished (Primary DA received or Complete).
func (d *Detector) Done() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.done
}

// Caps returns a copy of the capabilities discovered so far.
func (d *Detector) Caps() Caps {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.caps
}

// Enable returns sequences that turn on features present in Caps.
// Call only after Done is true. Empty means nothing to enable — except
// bracketed paste, which is always requested: many hosts support mode 2004
// but do not answer DECRQM, and unsupported terminals ignore the mode.
//
// SyncOutput and TrueColor / KittyGraphics are capability flags only —
// synchronized output is bracketed per frame by the renderer, and color /
// graphics need no mode switch here.
//
// Before building the sequence, Enable applies environment overrides
// (mirroring libvaxis): VHS_RECORD disables Kitty keyboard.
func (d *Detector) Enable() []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.applyEnvOverridesLocked()
	var b strings.Builder
	if d.caps.KittyKeyboard {
		b.WriteString(enableKittyKeyboard)
	}
	if d.caps.UnicodeCore {
		b.WriteString(enableUnicodeCore)
	}
	// Always enable; Caps.BracketedPaste still reflects DECRQM confirmation.
	b.WriteString(enableBracketedPaste)
	return []byte(b.String())
}

// applyEnvOverridesLocked mutates Caps for known broken or recording hosts.
// Caller must hold d.mu.
func (d *Detector) applyEnvOverridesLocked() {
	if os.Getenv("VHS_RECORD") != "" {
		d.caps.KittyKeyboard = false
	}
}

// Disable returns sequences that reverse Enable for features that were on.
// Safe to call even if Enable was never written.
func (d *Detector) Disable() []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	var b strings.Builder
	b.WriteString(disableBracketedPaste)
	if d.caps.UnicodeCore {
		b.WriteString(disableUnicodeCore)
	}
	if d.caps.KittyKeyboard {
		b.WriteString(disableKittyKeyboard)
	}
	return []byte(b.String())
}

// feedDECRPM handles CSI ? Ps ; Pm $ y (DECRPM). Returns whether seq was consumed.
// Caller must hold d.mu.
func feedDECRPM(c *Caps, seq string) bool {
	mode, status, ok := parseDECRPM(seq)
	if !ok {
		return false
	}
	supported := status != 0 && status != 4
	switch mode {
	case 2026:
		c.SyncOutput = supported
	case 2027:
		c.UnicodeCore = supported
	case 2004:
		c.BracketedPaste = supported
	default:
		// Unknown mode reply — still consume so it is not treated as input.
	}
	return true
}

// parseDECRPM parses CSI ? <mode> ; <status> $ y.
func parseDECRPM(seq string) (mode, status int, ok bool) {
	const prefix = "\x1b[?"
	const suffix = "$y"
	if !strings.HasPrefix(seq, prefix) || !strings.HasSuffix(seq, suffix) {
		return 0, 0, false
	}
	body := seq[len(prefix) : len(seq)-len(suffix)]
	before, after, ok := strings.Cut(body, ";")
	if !ok {
		return 0, 0, false
	}
	mode, err1 := strconv.Atoi(before)
	status, err2 := strconv.Atoi(after)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return mode, status, true
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

// isKittyGraphicsReply reports whether seq is an APC Kitty graphics response.
func isKittyGraphicsReply(seq string) bool {
	return strings.HasPrefix(seq, "\x1b_G") && strings.HasSuffix(seq, "\x1b\\")
}

// isXTGetTCapReply reports any XTGETTCAP DCS reply (success or failure).
func isXTGetTCapReply(seq string) bool {
	if !strings.HasPrefix(seq, "\x1bP") || !strings.HasSuffix(seq, "\x1b\\") {
		return false
	}
	// DCS Pt where Pt starts with 0+r / 1+r (XTGETTCAP).
	return strings.HasPrefix(seq, "\x1bP0+r") || strings.HasPrefix(seq, "\x1bP1+r")
}

// isXTGetTCapRGBSuccess reports a successful XTGETTCAP reply for key "RGB".
// Success form: DCS 1 + r 524742=... ST
func isXTGetTCapRGBSuccess(seq string) bool {
	return strings.HasPrefix(seq, "\x1bP1+r") && strings.Contains(seq, "524742")
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
