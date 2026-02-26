package fasttui

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
