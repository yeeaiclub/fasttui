package fasttui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/yeeaiclub/fasttui/keys"
)

// InputHandler processes terminal input and cell size queries
type InputHandler struct {
	cellSizeQueryPending bool
	inputBuffer          strings.Builder
	terminal             Terminal
	onInvalidate         func()
}

func newInputHandler(terminal Terminal, onInvalidate func()) *InputHandler {
	return &InputHandler{
		cellSizeQueryPending: false,
		terminal:             terminal,
		onInvalidate:         onInvalidate,
	}
}

func (ih *InputHandler) QueryCellSize() {
	if !ih.terminal.IsKittyProtocolActive() {
		return
	}
	ih.cellSizeQueryPending = true
	ih.terminal.Write("\x1b[16t")
}

func (ih *InputHandler) ProcessInput(data string, focusedComponent Component) string {
	if ih.cellSizeQueryPending {
		ih.inputBuffer.WriteString(data)
		filtered := ih.parseCellSizeResponse()
		if filtered == "" {
			return ""
		}
		data = filtered
	}

	if focusedComponent != nil {
		if keys.IsKeyRelease(data) && !focusedComponent.WantsKeyRelease() {
			return ""
		}
	}

	return data
}

func (ih *InputHandler) parseCellSizeResponse() string {
	data := ih.inputBuffer.String()

	responsePattern := `\x1b\[6;(\d+);(\d+)t`
	re := regexp.MustCompile(responsePattern)
	matches := re.FindStringSubmatch(data)

	if len(matches) == 3 {
		heightPx, err1 := strconv.Atoi(matches[1])
		widthPx, err2 := strconv.Atoi(matches[2])

		if err1 == nil && err2 == nil && heightPx > 0 && widthPx > 0 {
			ih.onInvalidate()

			ih.inputBuffer.Reset()
			ih.cellSizeQueryPending = false
			return ""
		}
	}

	partialPattern := `\x1b(\[6?;?[\d;]*)?$`
	rePartial := regexp.MustCompile(partialPattern)
	if rePartial.MatchString(data) {
		if len(data) > 0 {
			lastChar := data[len(data)-1]
			if !((lastChar >= 'a' && lastChar <= 'z') || (lastChar >= 'A' && lastChar <= 'Z') || lastChar == '~') {
				return ""
			}
		}
	}

	result := ih.inputBuffer.String()
	ih.inputBuffer.Reset()
	ih.cellSizeQueryPending = false
	return result
}
