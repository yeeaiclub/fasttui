package fasttui

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	CursorMarker = "\x1b_pi:c\x07"
	SegmentReset = "\x1b[0m\x1b]8;;\x07"
)

const (
	widthCacheSize = 512
)

var (
	widthCache = make(map[string]int)
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

type AnsiCodeTracker struct {
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	blink         bool
	inverse       bool
	hidden        bool
	strikethrough bool
	fgColor       string
	bgColor       string
}

func GetSegmenter() interface{} {
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

func VisibleWidth(s string) int {
	if len(s) == 0 {
		return 0
	}

	if cached, ok := widthCache[s]; ok {
		return cached
	}

	clean := s
	if containsString(clean, "\t") {
		clean = replaceAllString(clean, "\t", "   ")
	}
	if containsString(clean, "\x1b") {
		clean = replaceAllString(clean, "\x1b\\[[0-9;]*[mGKHJ]", "")
		clean = replaceAllString(clean, "\x1b\\]8;;[^\x07]*\x07", "")
		clean = replaceAllString(clean, "\x1b_[^\x07\x1b]*(?:\x07|\x17\\\\)", "")
	}

	width := len(clean)

	if len(widthCache) >= widthCacheSize {
		for key := range widthCache {
			delete(widthCache, key)
			break
		}
	}
	widthCache[s] = width

	return width
}

func NewAnsiCodeTracker() *AnsiCodeTracker {
	return &AnsiCodeTracker{}
}

func (t *AnsiCodeTracker) Process(ansiCode string) {
	if len(ansiCode) == 0 || ansiCode[len(ansiCode)-1] != 'm' {
		return
	}

	params := ansiCode[2 : len(ansiCode)-1]
	if params == "" || params == "0" {
		t.Reset()
		return
	}

	parts := splitSemicolon(params)
	i := 0
	for i < len(parts) {
		code := parseCode(parts[i])

		if code == 38 || code == 48 {
			if i+2 < len(parts) && parts[i+1] == "5" {
				colorCode := parts[i] + ";" + parts[i+1] + ";" + parts[i+2]
				if code == 38 {
					t.fgColor = colorCode
				} else {
					t.bgColor = colorCode
				}
				i += 3
				continue
			} else if i+4 < len(parts) && parts[i+1] == "2" {
				colorCode := parts[i] + ";" + parts[i+1] + ";" + parts[i+2] + ";" + parts[i+3] + ";" + parts[i+4]
				if code == 38 {
					t.fgColor = colorCode
				} else {
					t.bgColor = colorCode
				}
				i += 5
				continue
			}
		}

		switch code {
		case 0:
			t.Reset()
		case 1:
			t.bold = true
		case 2:
			t.dim = true
		case 3:
			t.italic = true
		case 4:
			t.underline = true
		case 5:
			t.blink = true
		case 7:
			t.inverse = true
		case 8:
			t.hidden = true
		case 9:
			t.strikethrough = true
		case 21:
			t.bold = false
		case 22:
			t.bold = false
			t.dim = false
		case 23:
			t.italic = false
		case 24:
			t.underline = false
		case 25:
			t.blink = false
		case 27:
			t.inverse = false
		case 28:
			t.hidden = false
		case 29:
			t.strikethrough = false
		case 39:
			t.fgColor = ""
		case 49:
			t.bgColor = ""
		default:
			if code >= 30 && code <= 37 {
				t.fgColor = parts[i]
			} else if code >= 40 && code <= 47 {
				t.bgColor = parts[i]
			} else if code >= 90 && code <= 97 {
				t.fgColor = parts[i]
			} else if code >= 100 && code <= 107 {
				t.bgColor = parts[i]
			}
		}
		i++
	}
}

func (t *AnsiCodeTracker) Reset() {
	t.bold = false
	t.dim = false
	t.italic = false
	t.underline = false
	t.blink = false
	t.inverse = false
	t.hidden = false
	t.strikethrough = false
	t.fgColor = ""
	t.bgColor = ""
}

func (t *AnsiCodeTracker) Clear() {
	t.Reset()
}

func (t *AnsiCodeTracker) GetActiveCodes() string {
	if !t.HasActiveCodes() {
		return ""
	}

	var codes []string
	if t.bold {
		codes = append(codes, "1")
	}
	if t.dim {
		codes = append(codes, "2")
	}
	if t.italic {
		codes = append(codes, "3")
	}
	if t.underline {
		codes = append(codes, "4")
	}
	if t.blink {
		codes = append(codes, "5")
	}
	if t.inverse {
		codes = append(codes, "7")
	}
	if t.hidden {
		codes = append(codes, "8")
	}
	if t.strikethrough {
		codes = append(codes, "9")
	}
	if t.fgColor != "" {
		codes = append(codes, t.fgColor)
	}
	if t.bgColor != "" {
		codes = append(codes, t.bgColor)
	}

	if len(codes) == 0 {
		return ""
	}
	return "\x1b[" + joinStrings(codes, ";") + "m"
}

func (t *AnsiCodeTracker) HasActiveCodes() bool {
	return t.bold || t.dim || t.italic || t.underline ||
		t.blink || t.inverse || t.hidden || t.strikethrough ||
		t.fgColor != "" || t.bgColor != ""
}

func (t *AnsiCodeTracker) GetLineEndReset() string {
	if t.underline {
		return "\x1b[24m"
	}
	return ""
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

func parseCode(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return result
}

func splitSemicolon(s string) []string {
	return strings.Split(s, ";")
}

func joinStrings(parts []string, sep string) string {
	return strings.Join(parts, sep)
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

	inputLines := splitLines(text)
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
		isWhitespace := trimSpace(token) == ""

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
	paddingNeeded := width - visibleLen
	if paddingNeeded < 0 {
		paddingNeeded = 0
	}
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

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
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

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func findSubstring(s, substr string) int {
	return strings.Index(s, substr)
}

func replaceAllString(s, pattern, replacement string) string {
	return strings.ReplaceAll(s, pattern, replacement)
}
