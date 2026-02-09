package terminal

import (
	"regexp"
	"strconv"
	"strings"
)

type Event struct {
	Type string
	Data string
}

type StdinBuffer struct {
	OnData      func(seq string)
	OnPaste     func(paste string)
	evChan      chan Event
	buffer      string
	pasteMode   bool
	pasteBuffer string
}

func NewStdinBuffer() *StdinBuffer {
	st := &StdinBuffer{}
	go st.ProcessEvent()
	return st
}

func (s *StdinBuffer) Process(data string) {
	// transform the data to a sequence
	var seq string
	if len(data) == 1 && data[0] > 127 {
		byteValue := data[0]
		seq = "\x1b[" + strconv.Itoa(int(byteValue-128))
	} else {
		seq = data
	}
	if len(seq) == 0 && len(data) == 1 {
		s.evChan <- Event{Type: "data", Data: ""}
		return
	}
	s.buffer += seq

	// handle the before BRACKETED_PASTE_START
	startIndex := strings.Index(s.buffer, BRACKETED_PASTE_START)
	if startIndex != -1 {
		if startIndex > 0 {
			before := s.buffer[:startIndex]
			seqs, _ := extractCompleteSequences(before)
			for _, seq := range seqs {
				s.evChan <- Event{Type: "data", Data: seq}
			}
		}
		s.buffer = s.buffer[startIndex+len(BRACKETED_PASTE_START):]
		s.pasteMode = true
		s.pasteBuffer = s.buffer
		s.buffer = ""
		endIndex := strings.Index(s.pasteBuffer, BRACKETED_PASTE_END)
		if endIndex != -1 {
			paste := s.pasteBuffer[:endIndex]
			remaining := s.pasteBuffer[endIndex+len(BRACKETED_PASTE_END):]
			s.pasteMode = false
			s.pasteBuffer = ""
			s.evChan <- Event{Type: "paste", Data: paste}
			if len(remaining) > 0 {
				s.Process(remaining)
			}
		}
		return
	}
	result, remaining := extractCompleteSequences(s.buffer)
	s.buffer = remaining

	for _, seq := range result {
		s.evChan <- Event{Type: "data", Data: seq}
	}
}

func (s *StdinBuffer) ProcessEvent() {
	for {
		select {
		case seq := <-s.evChan:
			if seq.Type == "data" {
				s.OnData(seq.Data)
			} else if seq.Type == "paste" {
				s.OnPaste(seq.Data)
			}
		default:
		}
	}
}

const (
	BRACKETED_PASTE_START = "\x1b[200~"
	BRACKETED_PASTE_END   = "\x1b[201~"
)

func (s *StdinBuffer) process(event Event) {

}

func (s *StdinBuffer) Flush() []string {
	return nil
}

func (s *StdinBuffer) clear() {
}

func extractCompleteSequences(buffer string) ([]string, string) {
	var seq []string
	pos := 0

	for pos < len(buffer) {
		remaining := buffer[pos:]
		if strings.HasPrefix(remaining, ESC) {
			seqEnd := 1
			for seqEnd < len(remaining) {
				candidate := remaining[:seqEnd]
				status := isCompleteSequence(candidate)
				if status == "complete" {
					seq = append(seq, candidate)
					pos += seqEnd
					break
				} else if status == "incomplete" {
					seqEnd++
				} else {
					seq = append(seq, candidate)
					pos += seqEnd
					break
				}
			}
			if seqEnd > len(remaining) {
				return seq, remaining
			}
		} else {
			seq = append(seq, remaining[:1])
			pos++
		}
	}

	return seq, ""
}

func isCompleteSequence(candidate string) string {
	if !strings.HasPrefix(candidate, ESC) {
		return "complete"
	}
	if len(candidate) == 1 {
		return "incomplete"
	}
	afterEsc := candidate[1:]

	if strings.HasPrefix(afterEsc, "[") {
		if strings.HasSuffix(afterEsc, "[M") {
			if len(afterEsc) >= 6 {
				return "complete"
			}
			return "incomplete"
		}
		return isCompleteCsiSequence(candidate)
	}

	if strings.HasPrefix(afterEsc, "]") {
		return isCompleteOscSequence(candidate)
	}

	if strings.HasPrefix(afterEsc, "P") {
		return isCompleteDcsSequence(candidate)
	}

	if strings.HasPrefix(afterEsc, "_") {
		return isCompleteApcSequence(candidate)
	}

	if strings.HasPrefix(afterEsc, "O") {
		if len(afterEsc) >= 2 {
			return "complete"
		}
		return "incomplete"
	}

	if len(afterEsc) == 1 {
		return "complete"
	}

	return "complete"
}

const ESC = "\x1b"

func isCompleteCsiSequence(data string) string {
	if !strings.HasPrefix(data, ESC+"[") {
		return "complete"
	}

	if len(data) < 3 {
		return "incomplete"
	}

	payload := data[2:]
	if len(payload) == 0 {
		return "incomplete"
	}

	lastChar := payload[len(payload)-1]
	lastCharCode := byte(lastChar)

	if lastCharCode >= 0x40 && lastCharCode <= 0x7e {
		if strings.HasPrefix(payload, "<") {
			mouseMatch := regexp.MustCompile(`^<\d+;\d+;\d+[Mm]$`).MatchString(payload)
			if mouseMatch {
				return "complete"
			}

			if lastChar == 'M' || lastChar == 'm' {
				parts := strings.Split(payload[1:len(payload)-1], ";")
				if len(parts) == 3 {
					allDigits := true
					for _, p := range parts {
						if _, err := strconv.Atoi(p); err != nil {
							allDigits = false
							break
						}
					}
					if allDigits {
						return "complete"
					}
				}
			}

			return "incomplete"
		}

		return "complete"
	}

	return "incomplete"
}

func isCompleteOscSequence(data string) string {
	if !strings.HasPrefix(data, ESC+"]") {
		return "complete"
	}

	if strings.HasSuffix(data, ESC+"\\") || strings.HasSuffix(data, "\x07") {
		return "complete"
	}

	return "incomplete"
}

func isCompleteDcsSequence(data string) string {
	if !strings.HasPrefix(data, ESC+"P") {
		return "complete"
	}

	if strings.HasSuffix(data, ESC+"\\") {
		return "complete"
	}

	return "incomplete"
}

func isCompleteApcSequence(data string) string {
	if !strings.HasPrefix(data, ESC+"_") {
		return "complete"
	}

	if strings.HasSuffix(data, ESC+"\\") {
		return "complete"
	}

	return "incomplete"
}
