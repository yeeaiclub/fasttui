package fasttui

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	CursorMarker = "\x1b_pi:c\x07"
	SegmentReset = "\x1b[0m\x1b]8;;\x07"
)

type segmentType int

const (
	segmentTypeAnsi segmentType = iota
	segmentTypeGrapheme
)

type textSegment struct {
	segType segmentType
	value   string
}

type SliceResult struct {
	text  string
	width int
}

func GetSegmenter() any {
	return nil
}

func GetCursorMarker() string {
	return CursorMarker
}

func GetSegmentReset() string {
	return SegmentReset
}

func GraphemeWidth(s string) int {
	if len(s) == 0 {
		return 0
	}
	return len(s)
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

func splitIntoTokensWithAnsi(text string) []string {
	var tokens []string
	var current string
	var pendingAnsi string
	inWhitespace := false
	i := 0

	for i < len(text) {
		code, length, ok := ExtractAnsiCode(text, i)
		if ok {
			pendingAnsi += code
			i += length
			continue
		}

		char := string(text[i])
		charIsSpace := char == " "

		if charIsSpace != inWhitespace && current != "" {
			tokens = append(tokens, current)
			current = ""
		}

		if pendingAnsi != "" {
			current += pendingAnsi
			pendingAnsi = ""
		}

		inWhitespace = charIsSpace
		current += char
		i++
	}

	if pendingAnsi != "" {
		current += pendingAnsi
	}

	if current != "" {
		tokens = append(tokens, current)
	}

	return tokens
}

func WrapTextWithAnsi(text string, width int) []string {
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

	if VisibleWidth(line) <= width {
		return []string{line}
	}

	var wrapped []string
	tracker := NewAnsiCodeTracker()
	tokens := splitIntoTokensWithAnsi(line)

	currentLine := ""
	currentVisibleLength := 0

	for _, token := range tokens {
		tokenVisibleLength := VisibleWidth(token)
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
				currentVisibleLength = VisibleWidth(currentLine)
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

	var segments []textSegment
	i := 0

	for i < len(word) {
		code, length, ok := ExtractAnsiCode(word, i)
		if ok {
			segments = append(segments, textSegment{segType: segmentTypeAnsi, value: code})
			i += length
		} else {
			end := i
			for end < len(word) {
				if _, _, ok := ExtractAnsiCode(word, end); ok {
					break
				}
				end++
			}
			textPortion := word[i:end]
			for _, c := range textPortion {
				segments = append(segments, textSegment{segType: segmentTypeGrapheme, value: string(c)})
			}
			i = end
		}
	}

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

		graphemeW := 1

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

func TruncateToWidth(text string, maxWidth int, ellipsis string, pad bool) string {
	if ellipsis == "" {
		ellipsis = "..."
	}

	textVisibleWidth := VisibleWidth(text)

	if textVisibleWidth <= maxWidth {
		if pad {
			return text + repeatSpaces(maxWidth-textVisibleWidth)
		}
		return text
	}

	ellipsisWidth := len(ellipsis)
	targetWidth := maxWidth - ellipsisWidth

	if targetWidth <= 0 {
		if maxWidth <= len(ellipsis) {
			return ellipsis[:maxWidth]
		}
		return ellipsis
	}

	var segments []textSegment
	i := 0

	for i < len(text) {
		code, length, ok := ExtractAnsiCode(text, i)
		if ok {
			segments = append(segments, textSegment{segType: segmentTypeAnsi, value: code})
			i += length
		} else {
			end := i
			for end < len(text) {
				if _, _, ok := ExtractAnsiCode(text, end); ok {
					break
				}
				end++
			}
			textPortion := text[i:end]
			for _, c := range textPortion {
				segments = append(segments, textSegment{segType: segmentTypeGrapheme, value: string(c)})
			}
			i = end
		}
	}

	result := ""
	currentWidth := 0

	for _, seg := range segments {
		if seg.segType == segmentTypeAnsi {
			result += seg.value
			continue
		}

		grapheme := seg.value
		if grapheme == "" {
			continue
		}

		graphemeW := 1

		if currentWidth+graphemeW > targetWidth {
			break
		}

		result += grapheme
		currentWidth += graphemeW
	}

	truncated := result + "\x1b[0m" + ellipsis
	if pad {
		truncatedWidth := len(truncated)
		return truncated + repeatSpaces(maxWidth-truncatedWidth)
	}
	return truncated
}

func SliceByColumn(line string, startCol int, length int, strict bool) string {
	result := SliceWithWidth(line, startCol, length, strict)
	return result.text
}

func SliceWithWidth(line string, startCol int, length int, strict bool) SliceResult {
	if length <= 0 {
		return SliceResult{text: "", width: 0}
	}

	endCol := startCol + length
	result := ""
	resultWidth := 0
	currentCol := 0
	i := 0
	pendingAnsi := ""

	for i < len(line) {
		code, codeLen, ok := ExtractAnsiCode(line, i)
		if ok {
			if currentCol >= startCol && currentCol < endCol {
				result += code
			} else if currentCol < startCol {
				pendingAnsi += code
			}
			i += codeLen
			continue
		}

		textEnd := i
		for textEnd < len(line) {
			if _, _, ok := ExtractAnsiCode(line, textEnd); ok {
				break
			}
			textEnd++
		}

		textPortion := line[i:textEnd]
		for _, c := range textPortion {
			w := 1
			inRange := currentCol >= startCol && currentCol < endCol
			fits := !strict || currentCol+w <= endCol
			if inRange && fits {
				if pendingAnsi != "" {
					result += pendingAnsi
					pendingAnsi = ""
				}
				result += string(c)
				resultWidth += w
			}
			currentCol += w
			if currentCol >= endCol {
				break
			}
		}
		i = textEnd
		if currentCol >= endCol {
			break
		}
	}

	return SliceResult{text: result, width: resultWidth}
}

func ExtractSegments(line string, beforeEnd int, afterStart int, afterLen int, strictAfter bool) (string, int, string, int) {
	before := ""
	beforeWidth := 0
	after := ""
	afterWidth := 0
	currentCol := 0
	i := 0
	pendingAnsiBefore := ""
	afterStarted := false
	afterEnd := afterStart + afterLen

	pooledStyleTracker := NewAnsiCodeTracker()
	pooledStyleTracker.Clear()

	for i < len(line) {
		code, codeLen, ok := ExtractAnsiCode(line, i)
		if ok {
			pooledStyleTracker.Process(code)
			if currentCol < beforeEnd {
				pendingAnsiBefore += code
			} else if currentCol >= afterStart && currentCol < afterEnd && afterStarted {
				after += code
			}
			i += codeLen
			continue
		}

		textEnd := i
		for textEnd < len(line) {
			if _, _, ok := ExtractAnsiCode(line, textEnd); ok {
				break
			}
			textEnd++
		}

		textPortion := line[i:textEnd]
		for _, c := range textPortion {
			w := 1

			if currentCol < beforeEnd {
				if pendingAnsiBefore != "" {
					before += pendingAnsiBefore
					pendingAnsiBefore = ""
				}
				before += string(c)
				beforeWidth += w
			} else if currentCol >= afterStart && currentCol < afterEnd {
				fits := !strictAfter || currentCol+w <= afterEnd
				if fits {
					if !afterStarted {
						after += pooledStyleTracker.GetActiveCodes()
						afterStarted = true
					}
					after += string(c)
					afterWidth += w
				}
			}

			currentCol += w
			if afterLen <= 0 {
				if currentCol >= beforeEnd {
					break
				}
			} else {
				if currentCol >= afterEnd {
					break
				}
			}
		}
		i = textEnd
		if afterLen <= 0 {
			if currentCol >= beforeEnd {
				break
			}
		} else {
			if currentCol >= afterEnd {
				break
			}
		}
	}

	return before, beforeWidth, after, afterWidth
}

func ApplyBackgroundToLine(line string, width int, bgFn func(string) string) string {
	visibleLen := VisibleWidth(line)
	paddingNeeded := max(width-visibleLen, 0)
	padding := repeatSpaces(paddingNeeded)

	withPadding := line + padding
	return bgFn(withPadding)
}

func IsWhitespaceChar(char string) bool {
	if len(char) == 0 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(char)
	return unicode.IsSpace(r)
}

func IsPunctuationChar(char string) bool {
	if len(char) == 0 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(char)
	punctuation := "(){}[]<>.,;:'\"!?+-*/\\|&%^$#@~`"
	for _, p := range punctuation {
		if r == p {
			return true
		}
	}
	return false
}

func trimRight(s string) string {
	return strings.TrimRight(s, " ")
}

func repeatSpaces(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(" ", count)
}
