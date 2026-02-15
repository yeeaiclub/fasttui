package main

import "github.com/yeeaiclub/fasttui/components"

// ANSI color functions
func cyan(s string) string {
	return "\x1b[36m" + s + "\x1b[0m"
}

func dim(s string) string {
	return "\x1b[2m" + s + "\x1b[0m"
}

func bold(s string) string {
	return "\x1b[1m" + s + "\x1b[0m"
}

func italic(s string) string {
	return "\x1b[3m" + s + "\x1b[0m"
}

func underline(s string) string {
	return "\x1b[4m" + s + "\x1b[0m"
}

func strikethrough(s string) string {
	return "\x1b[9m" + s + "\x1b[0m"
}

func green(s string) string {
	return "\x1b[32m" + s + "\x1b[0m"
}

func yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[0m"
}

func blue(s string) string {
	return "\x1b[34m" + s + "\x1b[0m"
}

func magenta(s string) string {
	return "\x1b[35m" + s + "\x1b[0m"
}

// CreateDefaultMarkdownTheme creates the default markdown theme
func CreateDefaultMarkdownTheme() *components.MarkdownTheme {
	return &components.MarkdownTheme{
		Heading:         cyan,
		Link:            blue,
		LinkURL:         dim,
		Code:            yellow,
		CodeBlock:       green,
		CodeBlockBorder: dim,
		Quote:           italic,
		QuoteBorder:     cyan,
		HR:              dim,
		ListBullet:      cyan,
		Bold:            bold,
		Italic:          italic,
		Strikethrough:   strikethrough,
		Underline:       underline,
		CodeBlockIndent: "  ",
	}
}
