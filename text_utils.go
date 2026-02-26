package fasttui

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func ApplyBackgroundToLine(line string, width int, bgFn func(string) string) string {
	visibleLen := VisibleWidth(line)
	paddingNeeded := max(width-visibleLen, 0)
	padding := repeatSpaces(paddingNeeded)

	withPadding := line + padding
	return bgFn(withPadding)
}

func IsWhitespaceChar(char string) bool {
	if len(char) == 0 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(char)
	return unicode.IsSpace(r)
}

func IsPunctuationChar(char string) bool {
	if len(char) == 0 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(char)
	punctuation := "(){}[]<>.,;:'\"!?+-*/\\|&%^$#@~`"
	for _, p := range punctuation {
		if r == p {
			return true
		}
	}
	return false
}

func trimRight(s string) string {
	return strings.TrimRight(s, " ")
}

func repeatSpaces(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(" ", count)
}
