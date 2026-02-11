package terminal

import (
	"strings"
)

const (
	BRACKETED_PASTE_START = "\x1b[200~"
	BRACKETED_PASTE_END   = "\x1b[201~"
	ESC                   = "\x1b"
)

type Event struct {
	Type string
	Data string
}

// ParserState represents the current state of the input parser
type ParserState int

const (
	StateNormal ParserState = iota
	StatePaste
	StateEscape
	StateCSI
	StateOSC
	StateDCS
	StateAPC
	StateSS3
)

type StdinBuffer struct {
	OnData      func(seq string)
	OnPaste     func(paste string)
	evChan      chan Event
	buffer      string
	state       ParserState
	pasteBuffer string
}

func NewStdinBuffer() *StdinBuffer {
	st := &StdinBuffer{
		evChan: make(chan Event, 100),
		state:  StateNormal,
	}
	go st.ProcessEvent()
	return st
}

func (s *StdinBuffer) Process(data string) {
	// Normalize high-bit characters
	var seq string
	if len(data) == 1 && data[0] > 127 {
		seq = ESC + string(data[0]-128)
	} else {
		seq = data
	}

	if len(seq) == 0 && len(data) == 1 {
		s.evChan <- Event{Type: "data", Data: ""}
		return
	}

	s.buffer += seq
	s.processBuffer()
}

func (s *StdinBuffer) processBuffer() {
	for len(s.buffer) > 0 {
		consumed := s.processState()
		if !consumed {
			break
		}
	}
}

func (s *StdinBuffer) processState() bool {
	switch s.state {
	case StateNormal:
		return s.processNormal()
	case StatePaste:
		return s.processPaste()
	case StateEscape:
		return s.processEscape()
	case StateCSI:
		return s.processCSI()
	case StateOSC:
		return s.processOSC()
	case StateDCS:
		return s.processDCS()
	case StateAPC:
		return s.processAPC()
	case StateSS3:
		return s.processSS3()
	default:
		s.state = StateNormal
		return true
	}
}

func (s *StdinBuffer) processNormal() bool {
	// Check for bracketed paste start
	if strings.HasPrefix(s.buffer, BRACKETED_PASTE_START) {
		s.buffer = s.buffer[len(BRACKETED_PASTE_START):]
		s.state = StatePaste
		s.pasteBuffer = ""
		return true
	}

	// Check for escape sequence
	if strings.HasPrefix(s.buffer, ESC) {
		s.state = StateEscape
		return true
	}

	// Regular character
	if len(s.buffer) > 0 {
		s.emitData(string(s.buffer[0]))
		s.buffer = s.buffer[1:]
		return true
	}

	return false
}

func (s *StdinBuffer) processPaste() bool {
	endIndex := strings.Index(s.buffer, BRACKETED_PASTE_END)
	if endIndex != -1 {
		s.pasteBuffer += s.buffer[:endIndex]
		s.buffer = s.buffer[endIndex+len(BRACKETED_PASTE_END):]
		s.emitPaste(s.pasteBuffer)
		s.pasteBuffer = ""
		s.state = StateNormal
		return true
	}

	// Accumulate paste data
	s.pasteBuffer += s.buffer
	s.buffer = ""
	return false
}

func (s *StdinBuffer) processEscape() bool {
	if len(s.buffer) < 2 {
		return false
	}

	nextChar := s.buffer[1]
	switch nextChar {
	case '[':
		s.state = StateCSI
		return true
	case ']':
		s.state = StateOSC
		return true
	case 'P':
		s.state = StateDCS
		return true
	case '_':
		s.state = StateAPC
		return true
	case 'O':
		s.state = StateSS3
		return true
	default:
		// Simple escape sequence (ESC + one char)
		s.emitData(s.buffer[:2])
		s.buffer = s.buffer[2:]
		s.state = StateNormal
		return true
	}
}

func (s *StdinBuffer) processCSI() bool {
	// CSI sequences: ESC [ ... final_byte
	// final_byte is in range 0x40-0x7E
	for i := 2; i < len(s.buffer); i++ {
		ch := s.buffer[i]
		if ch >= 0x40 && ch <= 0x7E {
			seq := s.buffer[:i+1]
			s.emitData(seq)
			s.buffer = s.buffer[i+1:]
			s.state = StateNormal
			return true
		}
	}
	return false
}

func (s *StdinBuffer) processOSC() bool {
	// OSC sequences: ESC ] ... (ESC \ or BEL)
	if idx := strings.Index(s.buffer, ESC+"\\"); idx != -1 {
		s.emitData(s.buffer[:idx+2])
		s.buffer = s.buffer[idx+2:]
		s.state = StateNormal
		return true
	}
	if idx := strings.Index(s.buffer, "\x07"); idx != -1 {
		s.emitData(s.buffer[:idx+1])
		s.buffer = s.buffer[idx+1:]
		s.state = StateNormal
		return true
	}
	return false
}

func (s *StdinBuffer) processDCS() bool {
	// DCS sequences: ESC P ... ESC \
	if idx := strings.Index(s.buffer, ESC+"\\"); idx != -1 {
		s.emitData(s.buffer[:idx+2])
		s.buffer = s.buffer[idx+2:]
		s.state = StateNormal
		return true
	}
	return false
}

func (s *StdinBuffer) processAPC() bool {
	// APC sequences: ESC _ ... ESC \
	if idx := strings.Index(s.buffer, ESC+"\\"); idx != -1 {
		s.emitData(s.buffer[:idx+2])
		s.buffer = s.buffer[idx+2:]
		s.state = StateNormal
		return true
	}
	return false
}

func (s *StdinBuffer) processSS3() bool {
	// SS3 sequences: ESC O X (function keys)
	if len(s.buffer) >= 3 {
		s.emitData(s.buffer[:3])
		s.buffer = s.buffer[3:]
		s.state = StateNormal
		return true
	}
	return false
}

func (s *StdinBuffer) emitData(data string) {
	s.evChan <- Event{Type: "data", Data: data}
}

func (s *StdinBuffer) emitPaste(data string) {
	s.evChan <- Event{Type: "paste", Data: data}
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

func (s *StdinBuffer) Clear() {
	s.buffer = ""
	s.state = StateNormal
	s.pasteBuffer = ""
}
