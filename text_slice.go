package fasttui

import (
	"strings"

	"github.com/clipperhouse/uax29/v2/graphemes"
)

func SliceByColumn(line string, startCol int, length int, strict bool) string {
	result := SliceWithWidth(line, startCol, length, strict)
	return result.text
}

func SliceWithWidth(line string, startCol int, length int, strict bool) SliceResult {
	if length <= 0 {
		return SliceResult{text: "", width: 0}
	}

	endCol := startCol + length
	var result strings.Builder
	resultWidth := 0
	currentCol := 0
	i := 0
	pendingAnsi := ""

	for i < len(line) {
		code, codeLen, ok := ExtractAnsiCode(line, i)
		if ok {
			if currentCol >= startCol && currentCol < endCol {
				result.WriteString(code)
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
		g := graphemes.FromString(textPortion)
		for g.Next() {
			grapheme := g.Value()
			w := GraphemeWidth(grapheme)
			inRange := currentCol >= startCol && currentCol < endCol
			fits := !strict || currentCol+w <= endCol
			if inRange && fits {
				if pendingAnsi != "" {
					result.WriteString(pendingAnsi)
					pendingAnsi = ""
				}
				result.WriteString(grapheme)
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

	return SliceResult{text: result.String(), width: resultWidth}
}

func ExtractSegments(line string, beforeEnd int, afterStart int, afterLen int, strictAfter bool) (string, int, string, int) {
	var before, after strings.Builder
	beforeWidth := 0
	afterWidth := 0
	currentCol := 0
	i := 0
	pendingAnsiBefore := ""
	afterStarted := false
	afterEnd := afterStart + afterLen

	tracker := NewAnsiCodeTracker()

	for i < len(line) {
		code, codeLen, ok := ExtractAnsiCode(line, i)
		if ok {
			tracker.Process(code)
			if currentCol < beforeEnd {
				pendingAnsiBefore += code
			} else if currentCol >= afterStart && currentCol < afterEnd && afterStarted {
				after.WriteString(code)
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
		g := graphemes.FromString(textPortion)
		for g.Next() {
			grapheme := g.Value()
			w := GraphemeWidth(grapheme)

			if currentCol < beforeEnd {
				if pendingAnsiBefore != "" {
					before.WriteString(pendingAnsiBefore)
					pendingAnsiBefore = ""
				}
				before.WriteString(grapheme)
				beforeWidth += w
			} else if currentCol >= afterStart && currentCol < afterEnd {
				fits := !strictAfter || currentCol+w <= afterEnd
				if fits {
					if !afterStarted {
						after.WriteString(tracker.GetActiveCodes())
						afterStarted = true
					}
					after.WriteString(grapheme)
					afterWidth += w
				}
			}

			currentCol += w

			// Determine break point and check once
			breakPoint := beforeEnd
			if afterLen > 0 {
				breakPoint = afterEnd
			}
			if currentCol >= breakPoint {
				break
			}
		}

		i = textEnd

		// Same consolidated break condition for outer loop
		breakPoint := beforeEnd
		if afterLen > 0 {
			breakPoint = afterEnd
		}
		if currentCol >= breakPoint {
			break
		}
	}

	return before.String(), beforeWidth, after.String(), afterWidth
}