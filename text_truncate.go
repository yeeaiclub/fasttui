package fasttui

import (
	"github.com/clipperhouse/uax29/v2/graphemes"
)

// TruncateToWidth truncates text to fit within a maximum visible width, adding ellipsis if needed.
// Optionally pad with spaces to reach exactly maxWidth.
// Properly handles ANSI escape codes (they don't count toward width).
//
// Parameters:
//   - text: Text to truncate (may contain ANSI codes)
//   - maxWidth: Maximum visible width
//   - ellipsis: Ellipsis string to append when truncating (default: "...")
//   - pad: If true, pad result with spaces to exactly maxWidth
//
// Returns: Truncated text, optionally padded to exactly maxWidth
func TruncateToWidth(text string, maxWidth int, ellipsis string, pad bool) string {
	if ellipsis == "" {
		ellipsis = "..."
	}

	textVisibleWidth := VisibleWidth(text)

	// If text fits, return as-is (with optional padding)
	if textVisibleWidth <= maxWidth {
		if pad {
			return text + repeatSpaces(maxWidth-textVisibleWidth)
		}
		return text
	}

	// Calculate target width (accounting for ellipsis)
	ellipsisWidth := VisibleWidth(ellipsis)
	targetWidth := maxWidth - ellipsisWidth

	// If no room for content, just return truncated ellipsis
	if targetWidth <= 0 {
		// Truncate ellipsis itself if needed
		if ellipsisWidth > maxWidth {
			return truncateEllipsis(ellipsis, maxWidth)
		}
		return ellipsis
	}

	// Separate ANSI codes from visible content using grapheme segmentation
	segments := segmentText(text)

	// Build truncated string from segments
	result := ""
	currentWidth := 0

	for _, seg := range segments {
		if seg.segType == segmentTypeAnsi {
			result += seg.value
			continue
		}

		grapheme := seg.value
		// Skip empty graphemes to avoid issues with width calculation
		if grapheme == "" {
			continue
		}

		graphemeWidth := GraphemeWidth(grapheme)
		if currentWidth+graphemeWidth > targetWidth {
			break
		}

		result += grapheme
		currentWidth += graphemeWidth
	}

	// Add reset code before ellipsis to prevent styling leaking into it
	truncated := result + "\x1b[0m" + ellipsis

	if pad {
		truncatedWidth := VisibleWidth(truncated)
		paddingNeeded := maxWidth - truncatedWidth
		if paddingNeeded > 0 {
			return truncated + repeatSpaces(paddingNeeded)
		}
		return truncated
	}

	return truncated
}

// segmentText separates ANSI codes from visible content using grapheme segmentation
func segmentText(text string) []textSegment {
	segments := []textSegment{}
	i := 0

	for i < len(text) {
		// Try to extract ANSI code at current position
		code, length, ok := ExtractAnsiCode(text, i)
		if ok {
			segments = append(segments, textSegment{
				segType: segmentTypeAnsi,
				value:   code,
			})
			i += length
		} else {
			// Find the next ANSI code or end of string
			end := i
			for end < len(text) {
				if _, _, ok := ExtractAnsiCode(text, end); ok {
					break
				}
				end++
			}

			// Segment this non-ANSI portion into graphemes
			textPortion := text[i:end]
			g := graphemes.FromString(textPortion)
			for g.Next() {
				segments = append(segments, textSegment{
					segType: segmentTypeGrapheme,
					value:   g.Value(),
				})
			}

			i = end
		}
	}

	return segments
}

// truncateEllipsis truncates the ellipsis string itself to fit maxWidth
func truncateEllipsis(ellipsis string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	currentWidth := 0
	result := ""

	g := graphemes.FromString(ellipsis)
	for g.Next() {
		grapheme := g.Value()
		graphemeWidth := GraphemeWidth(grapheme)

		if currentWidth+graphemeWidth > maxWidth {
			break
		}

		result += grapheme
		currentWidth += graphemeWidth
	}

	return result
}
