package fasttui

import (
	"strings"
)

type StdinBuffer struct {
	buffer    strings.Builder
	pasteMode bool
	OnData    func(seq string)
}

func NewStdinBuffer() *StdinBuffer {
	return &StdinBuffer{}
}

func (s *StdinBuffer) Process(data string) {
	s.buffer.WriteString(data)
}

func (s *StdinBuffer) Flush() []string {
	if s.buffer.Len() == 0 {
		return nil
	}
	content := s.buffer.String()
	s.buffer.Reset()
	return []string{content}
}

func (s *StdinBuffer) clear() {
	s.buffer.Reset()
	s.pasteMode = false
}
