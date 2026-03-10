package main

import "github.com/yeeaiclub/fasttui/components"

const (
	ansiReset = "\x1b[0m"
	ansiCyan  = "\x1b[36m"
	ansiBold  = "\x1b[1m"
	ansiWhite = "\x1b[37m"
	ansiDim   = "\x1b[2m"
)

var (
	cyan = func(s string) string {
		return ansiCyan + s + ansiReset
	}
	dim = func(s string) string {
		return ansiDim + s + ansiReset
	}
)

func bold(s string) string {
	return ansiBold + s + ansiReset
}

func ThemeFg(colorName string, text string) string {
	switch colorName {
	case "accent":
		return ansiCyan + ansiBold + text + ansiReset
	case "text":
		return ansiWhite + text + ansiReset
	default:
		return text
	}
}

func CreateDefaultMarkdownTheme() *components.MarkdownTheme {
	return &components.MarkdownTheme{
		Heading:         bold,
		Link:            cyan,
		LinkURL:         dim,
		Code:            func(s string) string { return s },
		CodeBlock:       func(s string) string { return s },
		CodeBlockBorder: dim,
		Quote:           func(s string) string { return s },
		QuoteBorder:     dim,
		HR:              func(s string) string { return "---" },
		ListBullet:      func(s string) string { return s },
		Bold:            bold,
		Italic:          func(s string) string { return s },
		Strikethrough:   func(s string) string { return s },
		Underline:       func(s string) string { return s },
	}
}
