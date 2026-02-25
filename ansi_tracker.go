package fasttui

import (
	"strconv"
	"strings"
)

// AnsiCodeTracker tracks the current state of ANSI escape codes in a text stream.
// It maintains the active text formatting attributes (bold, italic, colors, etc.)
// and can reconstruct the ANSI codes needed to continue formatting on a new line.
//
// Example usage:
//
//	tracker := NewAnsiCodeTracker()
//	tracker.Process("\x1b[1;31m")  // bold + red foreground
//	codes := tracker.GetActiveCodes()  // returns "\x1b[1;31m"
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

// NewAnsiCodeTracker creates a new ANSI code tracker with all formatting disabled.
//
// Example:
//
//	tracker := NewAnsiCodeTracker()
func NewAnsiCodeTracker() *AnsiCodeTracker {
	return &AnsiCodeTracker{}
}

// Process parses an ANSI escape code and updates the tracker's internal state.
// It handles SGR (Select Graphic Rendition) codes for text formatting and colors.
//
// Supported codes:
//   - 0: Reset all attributes
//   - 1: Bold, 2: Dim, 3: Italic, 4: Underline, 5: Blink
//   - 7: Inverse, 8: Hidden, 9: Strikethrough
//   - 21-29: Turn off corresponding attributes
//   - 30-37, 90-97: Foreground colors
//   - 40-47, 100-107: Background colors
//   - 38;5;n: 256-color foreground
//   - 48;5;n: 256-color background
//   - 38;2;r;g;b: RGB foreground
//   - 48;2;r;g;b: RGB background
//
// Example:
//
//	tracker.Process("\x1b[1;31m")     // bold + red
//	tracker.Process("\x1b[38;5;208m") // 256-color orange
//	tracker.Process("\x1b[0m")        // reset all
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

// Reset clears all active formatting attributes, returning the tracker to its initial state.
//
// Example:
//
//	tracker.Process("\x1b[1;31m")  // bold + red
//	tracker.Reset()                 // clear all formatting
//	tracker.HasActiveCodes()        // returns false
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

// Clear is an alias for Reset. It clears all active formatting attributes.
//
// Example:
//
//	tracker.Clear()  // same as tracker.Reset()
func (t *AnsiCodeTracker) Clear() {
	t.Reset()
}

// GetActiveCodes returns an ANSI escape sequence that represents all currently active
// formatting attributes. This is useful for continuing formatting on a new line.
// Returns an empty string if no formatting is active.
//
// Example:
//
//	tracker.Process("\x1b[1;31m")     // bold + red
//	codes := tracker.GetActiveCodes()  // returns "\x1b[1;31m"
//
//	// Use case: wrapping text while preserving formatting
//	line1 := "\x1b[1;31mHello"
//	tracker.Process("\x1b[1;31m")
//	line2 := tracker.GetActiveCodes() + "World\x1b[0m"
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

// HasActiveCodes returns true if any formatting attributes are currently active.
//
// Example:
//
//	tracker := NewAnsiCodeTracker()
//	tracker.HasActiveCodes()        // returns false
//	tracker.Process("\x1b[1m")      // bold
//	tracker.HasActiveCodes()        // returns true
func (t *AnsiCodeTracker) HasActiveCodes() bool {
	return t.bold || t.dim || t.italic || t.underline ||
		t.blink || t.inverse || t.hidden || t.strikethrough ||
		t.fgColor != "" || t.bgColor != ""
}

// GetLineEndReset returns an ANSI code to reset underline formatting at the end of a line.
// This is useful because underline formatting can extend beyond the text content.
// Returns "\x1b[24m" (turn off underline) if underline is active, empty string otherwise.
//
// Example:
//
//	tracker.Process("\x1b[4m")         // underline
//	reset := tracker.GetLineEndReset()  // returns "\x1b[24m"
//	line := "text" + reset              // prevents underline from extending
func (t *AnsiCodeTracker) GetLineEndReset() string {
	if t.underline {
		return "\x1b[24m"
	}
	return ""
}

/*
Complete Usage Example:

	package main

	import (
		"fmt"
		"regexp"
	)

	func main() {
		// Create a new tracker
		tracker := NewAnsiCodeTracker()

		// Example 1: Basic text formatting
		text := "\x1b[1;31mBold Red Text\x1b[0m Normal Text"
		tracker.Process("\x1b[1;31m")
		fmt.Println("Has active codes:", tracker.HasActiveCodes()) // true
		fmt.Println("Active codes:", tracker.GetActiveCodes())     // "\x1b[1;31m"

		// Example 2: Wrapping text while preserving formatting
		longText := "\x1b[1;32mThis is a very long line that needs to be wrapped"
		tracker.Reset()
		tracker.Process("\x1b[1;32m")

		// Split text and continue formatting on next line
		line1 := "This is a very long line"
		line2 := tracker.GetActiveCodes() + "that needs to be wrapped"
		fmt.Println(line1)
		fmt.Println(line2)

		// Example 3: Processing multiple ANSI codes
		tracker.Reset()
		tracker.Process("\x1b[1m")      // bold
		tracker.Process("\x1b[4m")      // underline
		tracker.Process("\x1b[38;5;208m") // 256-color orange
		fmt.Println(tracker.GetActiveCodes()) // "\x1b[1;4;38;5;208m"

		// Example 4: Handling underline at line end
		tracker.Reset()
		tracker.Process("\x1b[4m")
		lineEnd := "underlined text" + tracker.GetLineEndReset()
		fmt.Println(lineEnd) // prevents underline from extending

		// Example 5: Parsing text with ANSI codes
		ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
		styledText := "\x1b[1;31mRed\x1b[0m \x1b[32mGreen\x1b[0m"

		tracker.Reset()
		matches := ansiRegex.FindAllString(styledText, -1)
		for _, code := range matches {
			tracker.Process(code)
			fmt.Printf("Code: %s, Active: %s\n", code, tracker.GetActiveCodes())
		}

		// Example 6: RGB colors
		tracker.Reset()
		tracker.Process("\x1b[38;2;255;100;50m") // RGB foreground
		tracker.Process("\x1b[48;2;0;0;255m")    // RGB background
		fmt.Println(tracker.GetActiveCodes())     // "\x1b[38;2;255;100;50;48;2;0;0;255m"
	}
*/
