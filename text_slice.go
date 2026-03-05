package fasttui

import "github.com/clipperhouse/uax29/v2/graphemes"

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
		// Use grapheme segmentation for proper Unicode handling
		g := graphemes.FromString(textPortion)
		for g.Next() {
			grapheme := g.Value()
			w := GraphemeWidth(grapheme)
			inRange := currentCol >= startCol && currentCol < endCol
			fits := !strict || currentCol+w <= endCol
			if inRange && fits {
				if pendingAnsi != "" {
					result += pendingAnsi
					pendingAnsi = ""
				}
				result += grapheme
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
		// Use grapheme segmentation for proper Unicode handling
		g := graphemes.FromString(textPortion)
		for g.Next() {
			grapheme := g.Value()
			w := GraphemeWidth(grapheme)

			if currentCol < beforeEnd {
				if pendingAnsiBefore != "" {
					before += pendingAnsiBefore
					pendingAnsiBefore = ""
				}
				before += grapheme
				beforeWidth += w
			} else if currentCol >= afterStart && currentCol < afterEnd {
				fits := !strictAfter || currentCol+w <= afterEnd
				if fits {
					if !afterStarted {
						after += pooledStyleTracker.GetActiveCodes()
						afterStarted = true
					}
					after += grapheme
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
