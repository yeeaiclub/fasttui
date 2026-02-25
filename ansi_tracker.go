package fasttui

import (
	"strconv"
	"strings"
)

type AnsiCodeTracker struct {
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	blink         bool
	inverse       bool
	hidden        bool
	strikethrough bool
	fgColor       string
	bgColor       string
}

func NewAnsiCodeTracker() *AnsiCodeTracker {
	return &AnsiCodeTracker{}
}

func (t *AnsiCodeTracker) Process(ansiCode string) {
	if len(ansiCode) == 0 || ansiCode[len(ansiCode)-1] != 'm' {
		return
	}

	params := ansiCode[2 : len(ansiCode)-1]
	if params == "" || params == "0" {
		t.Reset()
		return
	}

	parts := strings.Split(params, ";")
	i := 0
	for i < len(parts) {
		code, _ := strconv.Atoi(parts[i])
		if code == 38 || code == 48 {
			if i+2 < len(parts) && parts[i+1] == "5" {
				colorCode := parts[i] + ";" + parts[i+1] + ";" + parts[i+2]
				if code == 38 {
					t.fgColor = colorCode
				} else {
					t.bgColor = colorCode
				}
				i += 3
				continue
			} else if i+4 < len(parts) && parts[i+1] == "2" {
				colorCode := parts[i] + ";" + parts[i+1] + ";" + parts[i+2] + ";" + parts[i+3] + ";" + parts[i+4]
				if code == 38 {
					t.fgColor = colorCode
				} else {
					t.bgColor = colorCode
				}
				i += 5
				continue
			}
		}

		switch code {
		case 0:
			t.Reset()
		case 1:
			t.bold = true
		case 2:
			t.dim = true
		case 3:
			t.italic = true
		case 4:
			t.underline = true
		case 5:
			t.blink = true
		case 7:
			t.inverse = true
		case 8:
			t.hidden = true
		case 9:
			t.strikethrough = true
		case 21:
			t.bold = false
		case 22:
			t.bold = false
			t.dim = false
		case 23:
			t.italic = false
		case 24:
			t.underline = false
		case 25:
			t.blink = false
		case 27:
			t.inverse = false
		case 28:
			t.hidden = false
		case 29:
			t.strikethrough = false
		case 39:
			t.fgColor = ""
		case 49:
			t.bgColor = ""
		default:
			if code >= 30 && code <= 37 {
				t.fgColor = parts[i]
			} else if code >= 40 && code <= 47 {
				t.bgColor = parts[i]
			} else if code >= 90 && code <= 97 {
				t.fgColor = parts[i]
			} else if code >= 100 && code <= 107 {
				t.bgColor = parts[i]
			}
		}
		i++
	}
}

func (t *AnsiCodeTracker) Reset() {
	t.bold = false
	t.dim = false
	t.italic = false
	t.underline = false
	t.blink = false
	t.inverse = false
	t.hidden = false
	t.strikethrough = false
	t.fgColor = ""
	t.bgColor = ""
}

func (t *AnsiCodeTracker) Clear() {
	t.Reset()
}

func (t *AnsiCodeTracker) GetActiveCodes() string {
	if !t.HasActiveCodes() {
		return ""
	}

	codes := make([]string, 0, 10)
	if t.bold {
		codes = append(codes, "1")
	}
	if t.dim {
		codes = append(codes, "2")
	}
	if t.italic {
		codes = append(codes, "3")
	}
	if t.underline {
		codes = append(codes, "4")
	}
	if t.blink {
		codes = append(codes, "5")
	}
	if t.inverse {
		codes = append(codes, "7")
	}
	if t.hidden {
		codes = append(codes, "8")
	}
	if t.strikethrough {
		codes = append(codes, "9")
	}
	if t.fgColor != "" {
		codes = append(codes, t.fgColor)
	}
	if t.bgColor != "" {
		codes = append(codes, t.bgColor)
	}

	var builder strings.Builder
	builder.WriteString("\x1b[")
	builder.WriteString(strings.Join(codes, ";"))
	builder.WriteString("m")
	return builder.String()
}

func (t *AnsiCodeTracker) HasActiveCodes() bool {
	return t.bold || t.dim || t.italic || t.underline ||
		t.blink || t.inverse || t.hidden || t.strikethrough ||
		t.fgColor != "" || t.bgColor != ""
}

func (t *AnsiCodeTracker) GetLineEndReset() string {
	if t.underline {
		return "\x1b[24m"
	}
	return ""
}
