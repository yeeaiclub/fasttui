package terminal

import (
	"regexp"
	"strings"
)

const (
	BRACKETED_PASTE_START = "\x1b[200~"
	BRACKETED_PASTE_END   = "\x1b[201~"
)

var mouseSequenceRegex = regexp.MustCompile(`^<\d+;\d+;\d+[Mm]$`)

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
	st := &StdinBuffer{
		evChan:    make(chan Event, 100),
		pasteMode: false,
	}
	go st.ProcessEvent()
	return st
}

func (s *StdinBuffer) Process(data string) {
	var seq string
	if len(data) == 1 && data[0] > 127 {
		byteValue := data[0]
		seq = "\x1b" + string(byteValue-128)
	} else {
		seq = data
	}
	if len(seq) == 0 && len(data) == 1 {
		s.evChan <- Event{Type: "data", Data: ""}
		return
	}
	s.buffer += seq

	if s.pasteMode {
		s.pasteBuffer += s.buffer
		s.buffer = ""
		endIndex := strings.Index(s.pasteBuffer, BRACKETED_PASTE_END)
		if endIndex != -1 {
			pastedContent := s.pasteBuffer[:endIndex]
			remaining := s.pasteBuffer[endIndex+len(BRACKETED_PASTE_END):]
			s.pasteMode = false
			s.pasteBuffer = ""
			s.evChan <- Event{Type: "paste", Data: pastedContent}
			if len(remaining) > 0 {
				s.Process(remaining)
			}
		}
		return
	}

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

	// handle string and extract complete sequences
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

func (s *StdinBuffer) Flush() []string {
	return nil
}

func (s *StdinBuffer) clear() {
	s.buffer = ""
	s.pasteMode = false
	s.pasteBuffer = ""
}

func extractCompleteSequences(buffer string) ([]string, string) {
	var sequences []string
	var pos int

	for pos < len(buffer) {
		if buffer[pos] == ESC[0] {
			seq, newPos := extractEscapeSequence(buffer, pos)
			if seq == "" {
				return sequences, buffer[pos:]
			}
			sequences = append(sequences, seq)
			pos = newPos
		} else {
			sequences = append(sequences, string(buffer[pos]))
			pos++
		}
	}

	return sequences, ""
}

func extractEscapeSequence(buffer string, pos int) (string, int) {
	for end := pos + 1; end <= len(buffer); end++ {
		candidate := buffer[pos:end]
		status := isCompleteSequence(candidate)

		switch status {
		case "complete":
			return candidate, end
		case "incomplete":
			continue
		default:
			return candidate, end
		}
	}
	return "", pos
}

func isCompleteSequence(candidate string) string {
	if !strings.HasPrefix(candidate, ESC) {
		return "not-escape"
	}
	if len(candidate) == 1 {
		return "incomplete"
	}
	afterEsc := candidate[1:]

	if strings.HasPrefix(afterEsc, "[") {
		if strings.HasPrefix(afterEsc, "[M") {
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
			return checkMouseSequence(payload)
		}
		return "complete"
	}

	return "incomplete"
}

func checkMouseSequence(payload string) string {
	if mouseSequenceRegex.MatchString(payload) {
		return "complete"
	}

	if len(payload) < 4 {
		return "incomplete"
	}

	lastChar := payload[len(payload)-1]
	if lastChar != 'M' && lastChar != 'm' {
		return "incomplete"
	}

	parts := strings.Split(payload[1:len(payload)-1], ";")
	if len(parts) != 3 {
		return "incomplete"
	}

	for _, part := range parts {
		if !isAllDigits(part) {
			return "incomplete"
		}
	}

	return "complete"
}

func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
