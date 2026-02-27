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
