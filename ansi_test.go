package fasttui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractAnsiCode(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		pos        int
		wantCode   string
		wantLength int
		wantOk     bool
	}{
		{
			name:       "CSI color code",
			input:      "\x1b[31mred text",
			pos:        0,
			wantCode:   "\x1b[31m",
			wantLength: 5,
			wantOk:     true,
		},
		{
			name:       "CSI reset code",
			input:      "\x1b[0m",
			pos:        0,
			wantCode:   "\x1b[0m",
			wantLength: 4,
			wantOk:     true,
		},
		{
			name:       "CSI with multiple parameters",
			input:      "\x1b[1;31;40m",
			pos:        0,
			wantCode:   "\x1b[1;31;40m",
			wantLength: 10,
			wantOk:     true,
		},
		{
			name:       "OSC hyperlink with BEL terminator",
			input:      "\x1b]8;;https://example.com\x07",
			pos:        0,
			wantCode:   "\x1b]8;;https://example.com\x07",
			wantLength: 25,
			wantOk:     true,
		},
		{
			name:       "OSC hyperlink with ESC backslash terminator",
			input:      "\x1b]8;;https://example.com\x1b\\",
			pos:        0,
			wantCode:   "\x1b]8;;https://example.com\x1b\\",
			wantLength: 26,
			wantOk:     true,
		},
		{
			name:       "APC cursor marker with BEL",
			input:      "\x1b_pi:c\x07",
			pos:        0,
			wantCode:   "\x1b_pi:c\x07",
			wantLength: 7,
			wantOk:     true,
		},
		{
			name:       "APC with ESC backslash terminator",
			input:      "\x1b_test\x1b\\",
			pos:        0,
			wantCode:   "\x1b_test\x1b\\",
			wantLength: 8,
			wantOk:     true,
		},
		{
			name:       "no escape at position",
			input:      "normal text",
			pos:        0,
			wantCode:   "",
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "escape at end of string",
			input:      "text\x1b",
			pos:        4,
			wantCode:   "",
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "incomplete CSI sequence",
			input:      "\x1b[31",
			pos:        0,
			wantCode:   "",
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "incomplete OSC sequence",
			input:      "\x1b]8;;url",
			pos:        0,
			wantCode:   "",
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "extract from middle of string",
			input:      "text\x1b[32mgreen",
			pos:        4,
			wantCode:   "\x1b[32m",
			wantLength: 5,
			wantOk:     true,
		},
		{
			name:       "position out of bounds",
			input:      "text",
			pos:        10,
			wantCode:   "",
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "unsupported escape sequence",
			input:      "\x1bX",
			pos:        0,
			wantCode:   "",
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "CSI cursor up A",
			input:      "\x1b[A",
			pos:        0,
			wantCode:   "\x1b[A",
			wantLength: 3,
			wantOk:     true,
		},
		{
			name:       "CSI cursor down B",
			input:      "\x1b[B",
			pos:        0,
			wantCode:   "\x1b[B",
			wantLength: 3,
			wantOk:     true,
		},
		{
			name:       "CSI cursor forward C",
			input:      "\x1b[10C",
			pos:        0,
			wantCode:   "\x1b[10C",
			wantLength: 5,
			wantOk:     true,
		},
		{
			name:       "CSI cursor back D",
			input:      "\x1b[5D",
			pos:        0,
			wantCode:   "\x1b[5D",
			wantLength: 4,
			wantOk:     true,
		},
		{
			name:       "CSI save cursor s",
			input:      "\x1b[s",
			pos:        0,
			wantCode:   "\x1b[s",
			wantLength: 3,
			wantOk:     true,
		},
		{
			name:       "CSI restore cursor u",
			input:      "\x1b[u",
			pos:        0,
			wantCode:   "\x1b[u",
			wantLength: 3,
			wantOk:     true,
		},
		{
			name:       "CSI hide cursor ?25l",
			input:      "\x1b[?25l",
			pos:        0,
			wantCode:   "\x1b[?25l",
			wantLength: 6,
			wantOk:     true,
		},
		{
			name:       "CSI show cursor ?25h",
			input:      "\x1b[?25h",
			pos:        0,
			wantCode:   "\x1b[?25h",
			wantLength: 6,
			wantOk:     true,
		},
		{
			name:       "CSI device attributes c",
			input:      "\x1b[0c",
			pos:        0,
			wantCode:   "\x1b[0c",
			wantLength: 4,
			wantOk:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, gotLength, gotOk := ExtractAnsiCode(tt.input, tt.pos)
			assert.Equal(t, tt.wantCode, gotCode, "code mismatch")
			assert.Equal(t, tt.wantLength, gotLength, "length mismatch")
			assert.Equal(t, tt.wantOk, gotOk, "ok mismatch")
		})
	}
}

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "hello", "hello"},
		{"csi only", "\x1b[31mred\x1b[0m", "red"},
		{"osc bel", "a\x1b]8;;u\x07b", "ab"},
		{"osc st", "x\x1b]0;t\x1b\\y", "xy"},
		{"apc", "a\x1b_pi:c\x07b", "ab"},
		{"mixed", "\x1b[1m\x1b]8;;\x07\x1b[0mok", "ok"},
		{"lone esc", "a\x1bb", "a\x1bb"},
		{"incomplete csi no final byte", "a\x1b[31", "a\x1b[31"},
		{"csi cursor up", "\x1b[Atext", "text"},
		{"csi cursor down", "\x1b[Btext", "text"},
		{"csi cursor forward", "\x1b[10Ctext", "text"},
		{"csi cursor back", "\x1b[5Dtext", "text"},
		{"csi save cursor", "\x1b[stext", "text"},
		{"csi restore cursor", "\x1b[utext", "text"},
		{"csi hide cursor", "\x1b[?25ltext", "text"},
		{"csi show cursor", "\x1b[?25htext", "text"},
		{"csi device attributes", "\x1b[0ctext", "text"},
		{"csi valid b terminator", "a\x1b[31b", "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, StripAnsi(tt.input))
		})
	}
}
