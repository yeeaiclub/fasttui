package fasttui

const (
	CursorMarker = "\x1b_pi:c\x07"
	SegmentReset = "\x1b[0m\x1b]8;;\x07"
)

func GetCursorMarker() string {
	return CursorMarker
}

func GetSegmentReset() string {
	return SegmentReset
}

// ExtractAnsiCode extracts an ANSI escape sequence starting at the given position.
// It supports three types of sequences:
//   - CSI (Control Sequence Introducer): ESC [ ... terminator (e.g., ESC[31m for red text)
//   - OSC (Operating System Command): ESC ] ... BEL or ESC \ (e.g., ESC]8;;url for hyperlinks)
//   - APC (Application Program Command): ESC _ ... BEL or ESC \ (e.g., ESC_pi:c for cursor marker)
//
// Returns the complete escape sequence, its length, and whether extraction succeeded.
func ExtractAnsiCode(s string, pos int) (code string, length int, ok bool) {
	if pos >= len(s) || s[pos] != '\x1b' {
		return "", 0, false
	}

	if pos+1 >= len(s) {
		return "", 0, false
	}

	next := s[pos+1]

	switch next {
	case '[': // CSI sequence
		j := pos + 2
		for j < len(s) && !isTerminator(s[j]) {
			j++
		}
		if j < len(s) {
			return s[pos : j+1], j + 1 - pos, true
		}
		return "", 0, false

	case ']', '_': // OSC or APC sequence
		return extractOscOrApc(s, pos)
	}

	return "", 0, false
}

func extractOscOrApc(s string, pos int) (code string, length int, ok bool) {
	j := pos + 2
	for j < len(s) {
		if s[j] == '\x07' {
			return s[pos : j+1], j + 1 - pos, true
		}
		if j+1 < len(s) && s[j] == '\x1b' && s[j+1] == '\\' {
			return s[pos : j+2], j + 2 - pos, true
		}
		j++
	}
	return "", 0, false
}

func isTerminator(b byte) bool {
	return b == 'm' || b == 'G' || b == 'K' || b == 'H' || b == 'J'
}

func updateTrackerFromText(text string, tracker *AnsiCodeTracker) {
	i := 0
	for i < len(text) {
		code, length, ok := ExtractAnsiCode(text, i)
		if ok {
			tracker.Process(code)
			i += length
		} else {
			i++
		}
	}
}
