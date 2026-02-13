package fasttui

// RenderState tracks the state of the terminal rendering between frames
type RenderState struct {
	previousLines       []string
	previousWidth       int
	previousViewportTop int
	maxLinesRendered    int
	fullRedrawCount     int
}

func newRenderState() *RenderState {
	return &RenderState{
		previousLines:       nil,
		previousWidth:       0,
		previousViewportTop: 0,
		maxLinesRendered:    0,
		fullRedrawCount:     0,
	}
}

func (rs *RenderState) Reset() {
	rs.previousLines = nil
	rs.previousWidth = -1
	rs.maxLinesRendered = 0
	rs.previousViewportTop = 0
}

func (rs *RenderState) UpdateAfterRender(lines []string, width, height int) {
	rs.previousLines = lines
	rs.previousWidth = width
	rs.maxLinesRendered = max(rs.maxLinesRendered, len(lines))
	rs.previousViewportTop = max(0, rs.maxLinesRendered-height)
}

func (rs *RenderState) UpdateAfterClear(lines []string, height int) {
	rs.previousLines = lines
	rs.maxLinesRendered = len(lines)
	rs.previousViewportTop = max(0, rs.maxLinesRendered-height)
	rs.fullRedrawCount++
}

func (rs *RenderState) GetViewportTop(height int) int {
	return max(0, rs.maxLinesRendered-height)
}

func (rs *RenderState) FindChangedLineRange(newLines []string) (firstChanged, lastChanged int) {
	firstChanged = -1
	lastChanged = -1
	maxLines := max(len(newLines), len(rs.previousLines))

	for i := range maxLines {
		oldLine := ""
		newLine := ""
		if i < len(rs.previousLines) {
			oldLine = rs.previousLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			if firstChanged == -1 {
				firstChanged = i
			}
			lastChanged = i
		}
	}
	return firstChanged, lastChanged
}
