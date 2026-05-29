package fasttui

import (
	"github.com/clipperhouse/uax29/v2/graphemes"
)

// appendTextSegments scans text from pos, appending ANSI codes and graphemes.
// Scanning is O(n): only ESC bytes trigger ExtractAnsiCode.
func appendTextSegments(segments []textSegment, text string, pos int) ([]textSegment, int) {
	for pos < len(text) {
		if text[pos] != '\x1b' {
			start := pos
			for pos < len(text) && text[pos] != '\x1b' {
				pos++
			}
			g := graphemes.FromString(text[start:pos])
			for g.Next() {
				segments = append(segments, textSegment{
					segType: segmentTypeGrapheme,
					value:   g.Value(),
				})
			}
			continue
		}

		code, length, ok := ExtractAnsiCode(text, pos)
		if ok {
			segments = append(segments, textSegment{
				segType: segmentTypeAnsi,
				value:   code,
			})
			pos += length
			continue
		}

		segments = append(segments, textSegment{
			segType: segmentTypeGrapheme,
			value:   text[pos : pos+1],
		})
		pos++
	}

	return segments, pos
}

func scanTextSegments(text string) []textSegment {
	segments := acquireTextSegments()
	segments, _ = appendTextSegments(segments, text, 0)
	out := make([]textSegment, len(segments))
	copy(out, segments)
	releaseTextSegments(segments)
	return out
}
