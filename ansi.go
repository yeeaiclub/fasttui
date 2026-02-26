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

func ExtractAnsiCode(s string, pos int) (code string, length int, ok bool) {
	if pos >= len(s) || s[pos] != '\x1b' {
		return "", 0, false
	}

	if pos+1 >= len(s) {
		return "", 0, false
	}

	next := s[pos+1]

	if next == '[' {
		j := pos + 2
		for j < len(s) && !isTerminator(s[j]) {
			j++
		}
		if j < len(s) {
			return s[pos : j+1], j + 1 - pos, true
		}
		return "", 0, false
	}

	if next == ']' {
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

	if next == '_' {
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
