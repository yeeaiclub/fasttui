package terminal

import (
	"strings"
)

type StdinBuffer struct {
	buffer      chan rune
	pasteBuffer chan strings.Builder
	pasteMode   bool
	OnData      func(seq string)
}

func NewStdinBuffer() *StdinBuffer {
	return &StdinBuffer{}
}

func (s *StdinBuffer) Process(data string) {
}

func (s *StdinBuffer) Flush() []string {
	return nil
}

func (s *StdinBuffer) clear() {
}
