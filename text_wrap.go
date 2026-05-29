package fasttui

import (
	"strings"
	"unicode/utf8"
)

func splitIntoTokensWithAnsi(text string) []string {
	if isPrintableASCII(text) {
		return splitPureASCII(text)
	}
	if IsASCII(text) {
		return splitASCIITokensWithAnsi(text)
	}

	b := acquireBuilder()
	defer releaseBuilder(b)

	var tokens []string
	inWhitespace := false
	i := 0

	for i < len(text) {
		code, length, ok := ExtractAnsiCode(text, i)
		if ok {
			b.WriteString(code)
			i += length
			continue
		}

		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			size = 1
		}
		char := text[i : i+size]
		charIsSpace := r == ' '

		if charIsSpace != inWhitespace && b.Len() > 0 {
			tokens = append(tokens, b.String())
			b.Reset()
		}

		inWhitespace = charIsSpace
		b.WriteString(char)
		i += size
	}

	if b.Len() > 0 {
		tokens = append(tokens, b.String())
	}

	return tokens
}

func splitASCIITokensWithAnsi(text string) []string {
	b := acquireBuilder()
	defer releaseBuilder(b)

	var tokens []string
	inWhitespace := false
	i := 0

	for i < len(text) {
		code, length, ok := ExtractAnsiCode(text, i)
		if ok {
			b.WriteString(code)
			i += length
			continue
		}

		charIsSpace := text[i] == ' '
		if charIsSpace != inWhitespace && b.Len() > 0 {
			tokens = append(tokens, b.String())
			b.Reset()
		}

		inWhitespace = charIsSpace
		b.WriteByte(text[i])
		i++
	}

	if b.Len() > 0 {
		tokens = append(tokens, b.String())
	}

	return tokens
}

func splitPureASCII(s string) []string {
	var tokens []string
	start := 0
	inSpace := s[0] == ' '

	for i := 0; i < len(s); i++ {
		isSpace := s[i] == ' '
		if isSpace != inSpace {
			tokens = append(tokens, s[start:i])
			start = i
			inSpace = isSpace
		}
	}

	if start < len(s) {
		tokens = append(tokens, s[start:])
	}

	if len(tokens) == 0 {
		return []string{s}
	}
	return tokens
}

// WrapAnsiText wraps text to the given width,
//  preserving ANSI escape codes (colors, bold, etc.).
func WrapAnsiText(text string, width int) []string {
	if text == "" {
		return []string{""}
	}

	inputLines := strings.Split(text, "\n")
	var result []string
	tracker := NewAnsiCodeTracker()

	for _, inputLine := range inputLines {
		var prefix string
		if len(result) > 0 {
			prefix = tracker.GetActiveCodes()
		}
		wrapped := wrapSingleLine(prefix+inputLine, width)
		result = append(result, wrapped...)
		updateTrackerFromText(inputLine, tracker)
	}

	if len(result) == 0 {
		return []string{""}
	}
	return result
}

func wrapSingleLine(line string, width int) []string {
	if line == "" {
		return []string{""}
	}

	if visibleWidthFast(line) <= width {
		return []string{line}
	}

	var wrapped []string
	tracker := NewAnsiCodeTracker()
	tokens := splitIntoTokensWithAnsi(line)

	currentLine := ""
	currentVisibleLength := 0

	for _, token := range tokens {
		tokenVisibleLength := visibleWidthFast(token)
		isWhitespace := strings.TrimSpace(token) == ""

		if tokenVisibleLength > width && !isWhitespace {
			if currentLine != "" {
				lineEndReset := tracker.GetLineEndReset()
				if lineEndReset != "" {
					currentLine += lineEndReset
				}
				wrapped = append(wrapped, currentLine)
				currentLine = ""
				currentVisibleLength = 0
			}

			broken := breakLongWord(token, width, tracker)
			if len(broken) > 0 {
				wrapped = append(wrapped, broken[:len(broken)-1]...)
				currentLine = broken[len(broken)-1]
				currentVisibleLength = visibleWidthFast(currentLine)
			}
			continue
		}

		totalNeeded := currentVisibleLength + tokenVisibleLength

		if totalNeeded > width && currentVisibleLength > 0 {
			lineToWrap := trimRight(currentLine)
			lineEndReset := tracker.GetLineEndReset()
			if lineEndReset != "" {
				lineToWrap += lineEndReset
			}
			wrapped = append(wrapped, lineToWrap)
			if isWhitespace {
				currentLine = tracker.GetActiveCodes()
				currentVisibleLength = 0
			} else {
				currentLine = tracker.GetActiveCodes() + token
				currentVisibleLength = tokenVisibleLength
			}
		} else {
			currentLine += token
			currentVisibleLength += tokenVisibleLength
		}

		updateTrackerFromText(token, tracker)
	}

	if currentLine != "" {
		wrapped = append(wrapped, currentLine)
	}

	if len(wrapped) == 0 {
		return []string{""}
	}
	for i := range wrapped {
		wrapped[i] = trimRight(wrapped[i])
	}
	return wrapped
}

func breakLongWord(word string, width int, tracker *AnsiCodeTracker) []string {
	var lines []string
	currentLine := tracker.GetActiveCodes()
	currentWidth := 0

	segments := acquireTextSegments()
	segments, _ = appendTextSegments(segments, word, 0)
	defer releaseTextSegments(segments)

	for _, seg := range segments {
		if seg.segType == segmentTypeAnsi {
			currentLine += seg.value
			tracker.Process(seg.value)
			continue
		}

		grapheme := seg.value
		if grapheme == "" {
			continue
		}

		graphemeW := GraphemeWidth(grapheme)

		if currentWidth+graphemeW > width {
			lineEndReset := tracker.GetLineEndReset()
			if lineEndReset != "" {
				currentLine += lineEndReset
			}
			lines = append(lines, currentLine)
			currentLine = tracker.GetActiveCodes()
			currentWidth = 0
		}

		currentLine += grapheme
		currentWidth += graphemeW
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}
