package fasttui

import "strings"

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
