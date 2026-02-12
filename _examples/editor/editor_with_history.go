package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

// ChatApp wraps the editor and history display
type ChatApp struct {
	editor      *components.Editor
	historyText *components.Text
	tui         *fasttui.TUI
	messages    []string
}

func NewChatApp(tui *fasttui.TUI, term fasttui.Terminal) *ChatApp {
	app := &ChatApp{
		tui:      tui,
		messages: []string{},
	}

	// Create history display
	app.historyText = components.NewText("", 2, 1, nil)

	// Create editor
	app.editor = components.NewEditor()
	app.editor.SetTerminal(term)
	app.editor.SetPaddingX(2)

	// Set up submit callback
	app.editor.OnSubmit = func(text string) {
		if strings.TrimSpace(text) != "" {
			app.addMessage("You: " + text)
			app.updateHistoryDisplay()
		}
	}

	// Initialize with empty editor
	app.editor.SetText([]string{""})
	app.editor.SetCursor(0, 0)

	return app
}

func (app *ChatApp) addMessage(msg string) {
	app.messages = append(app.messages, msg)
	// Keep only last 20 messages
	if len(app.messages) > 20 {
		app.messages = app.messages[len(app.messages)-20:]
	}
}

func (app *ChatApp) updateHistoryDisplay() {
	if len(app.messages) == 0 {
		app.historyText.SetText("No messages yet. Type something and press Enter to submit.")
	} else {
		// Join messages with newlines
		historyContent := strings.Join(app.messages, "\n")
		app.historyText.SetText(historyContent)
	}
	app.tui.RequestRender(false)
}

func (app *ChatApp) Render(width int) []string {
	var result []string

	// Render history
	historyLines := app.historyText.Render(width)
	result = append(result, historyLines...)

	// Add separator
	separator := strings.Repeat("─", width)
	result = append(result, "\x1b[36m"+separator+"\x1b[0m")

	// Render editor
	editorLines := app.editor.Render(width)
	result = append(result, editorLines...)

	return result
}

func (app *ChatApp) HandleInput(data string) {
	// Handle Ctrl+Q to quit
	if data == "\x11" { // Ctrl+Q
		app.tui.Stop()
		fmt.Println("\nChat closed")
		os.Exit(0)
	}

	// Handle Ctrl+L to clear history
	if data == "\x0c" { // Ctrl+L
		app.messages = []string{}
		app.updateHistoryDisplay()
		return
	}

	// Pass to editor
	app.editor.HandleInput(data)
	app.tui.RequestRender(false)
}

func (app *ChatApp) WantsKeyRelease() bool {
	return app.editor.WantsKeyRelease()
}

func (app *ChatApp) Invalidate() {
	app.editor.Invalidate()
	app.historyText.Invalidate()
}

func (app *ChatApp) SetFocused(focused bool) {
	app.editor.SetFocused(focused)
}

func (app *ChatApp) IsFocused() bool {
	return app.editor.IsFocused()
}

func main() {
	// Create terminal and TUI
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)

	// Create chat app
	chatApp := NewChatApp(tui, term)

	// Add initial welcome message
	chatApp.addMessage("Welcome to the chat editor!")
	chatApp.addMessage("Type a message and press Enter to submit.")
	chatApp.addMessage("Press Ctrl+L to clear history, Ctrl+Q to quit.")
	chatApp.updateHistoryDisplay()

	// Add chat app to TUI
	tui.AddChild(chatApp)
	tui.SetFocus(chatApp)

	// Add help text
	helpText := components.NewText("Ctrl+Q: Quit | Ctrl+L: Clear | Enter: Submit", 1, 0, func(s string) string {
		return "\x1b[2m" + s + "\x1b[0m" // Dim text
	})
	tui.AddChild(helpText)

	// Start TUI
	tui.Start()

	// Keep running
	select {}
}
