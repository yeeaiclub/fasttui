package terminal

import (
	"strconv"
	"strings"
)

type Event struct {
	Type string
	Data string
}

type StdinBuffer struct {
	OnData func(seq string)
	evChan chan Event
	buffer string
}

func NewStdinBuffer() *StdinBuffer {
	st := &StdinBuffer{}
	go st.ProcessEvent()
	return st
}

func (s *StdinBuffer) Process(data string) {
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

	startIndex := strings.Index(s.buffer, BRACKETED_PASTE_START)
	if startIndex != -1 {
		if startIndex > 0 {

		}
	}
}

func (s *StdinBuffer) ProcessEvent() {
	for {
		select {
		case seq := <-s.evChan:
			s.process(seq)
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

func extractCompleteSequences(buffer string) []string {
	return nil
}
