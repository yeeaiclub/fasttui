package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

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

// Create default markdown theme
func createDefaultMarkdownTheme() *components.MarkdownTheme {
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

func main() {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, false)

	isResponding := false
	theme := createDefaultMarkdownTheme()

	// Welcome messages
	welcomeText := components.NewText(bold("Welcome to Simple Chat!"), 1, 1, nil)
	instructionsText := components.NewText("Type your messages below. Commands: /delete (remove last message), /clear (clear all messages)", 1, 0, nil)

	tui.AddChild(welcomeText)
	tui.AddChild(instructionsText)

	// Create editor
	editor := components.NewEditor(term, func(value string) {
		// Prevent submission if already responding
		if isResponding {
			return
		}

		trimmed := strings.TrimSpace(value)

		// Handle slash commands
		if trimmed == "/delete" {
			children := tui.GetChildren()
			// Remove component before editor (if there are any besides the initial text)
			if len(children) > 3 {
				// children[0] = "Welcome to Simple Chat!"
				// children[1] = "Type your messages below..."
				// children[2...n-1] = messages
				// children[n] = editor
				tui.RemoveChildAt(len(children) - 2)
			}
			tui.RequestRender(false)
			return
		}

		if trimmed == "/clear" {
			children := tui.GetChildren()
			// Remove all messages but keep the welcome text and editor
			for len(children) > 3 {
				tui.RemoveChildAt(2)
				children = tui.GetChildren()
			}
			tui.RequestRender(false)
			return
		}

		if trimmed != "" {
			isResponding = true

			// Add user message
			userMessage := components.NewMarkdown(value, 1, 1, theme, nil)
			children := tui.GetChildren()
			tui.InsertChildAt(len(children)-1, userMessage)

			// Add loader
			loader := components.NewLoader(
				tui,
				cyan,
				dim,
				"Thinking...",
			)
			children = tui.GetChildren()
			tui.InsertChildAt(len(children)-1, loader)

			tui.RequestRender(false)

			// Simulate async response
			go func() {
				time.Sleep(1 * time.Second)

				// Remove loader
				tui.RemoveChild(loader)
				loader.Stop()

				// Simulate a response
				responses := []string{
					"That's interesting! Tell me more.",
					"I see what you mean.",
					"Fascinating perspective!",
					"Could you elaborate on that?",
					"That makes sense to me.",
					"I hadn't thought of it that way.",
					"Great point!",
					"Thanks for sharing that.",
				}
				randomResponse := responses[rand.Intn(len(responses))]

				// Add assistant message
				botMessage := components.NewMarkdown(randomResponse, 1, 1, theme, nil)
				children := tui.GetChildren()
				tui.InsertChildAt(len(children)-1, botMessage)

				// Re-enable submit
				isResponding = false

				// Request render
				tui.RequestRender(false)
			}()
		}
	})

	tui.AddChild(editor)
	tui.SetFocus(editor)

	// Start TUI
	tui.Start()

	// Keep running
	select {}
}
