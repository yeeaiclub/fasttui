package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

type ChatApp struct {
	tui          *fasttui.TUI
	term         *terminal.ProcessTerminal
	isResponding bool
	theme        *components.MarkdownTheme
}

func NewChatApp() *ChatApp {
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, false)

	return &ChatApp{
		tui:   tui,
		term:  term,
		theme: CreateDefaultMarkdownTheme(),
	}
}

func (app *ChatApp) handleDeleteCommand() {
	children := app.tui.GetChildren()
	if len(children) > 3 {
		app.tui.RemoveChildAt(len(children) - 2)
	}
	app.tui.RequestRender(false)
}

func (app *ChatApp) handleClearCommand() {
	children := app.tui.GetChildren()
	for len(children) > 3 {
		app.tui.RemoveChildAt(2)
		children = app.tui.GetChildren()
	}
	app.tui.RequestRender(false)
}

func (app *ChatApp) addUserMessage(value string) {
	userMessage := components.NewMarkdown(value, 1, 1, app.theme, nil)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-1, userMessage)
}

func (app *ChatApp) addLoader() *components.Loader {
	loader := components.NewLoader(
		app.tui,
		cyan,
		dim,
		"Thinking...",
	)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-1, loader)
	app.tui.RequestRender(false)
	return loader
}

func (app *ChatApp) simulateResponse(loader *components.Loader) {
	go func() {
		time.Sleep(1 * time.Second)

		app.tui.RemoveChild(loader)
		loader.Stop()

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

		botMessage := components.NewMarkdown(randomResponse, 1, 1, app.theme, nil)
		children := app.tui.GetChildren()
		app.tui.InsertChildAt(len(children)-1, botMessage)

		app.isResponding = false
		app.tui.RequestRender(false)
	}()
}

func (app *ChatApp) handleSubmit(value string) {
	if app.isResponding {
		return
	}

	trimmed := strings.TrimSpace(value)

	switch trimmed {
	case "/delete":
		app.handleDeleteCommand()
		return
	case "/clear":
		app.handleClearCommand()
		return
	}

	if trimmed != "" {
		app.isResponding = true
		app.addUserMessage(value)
		loader := app.addLoader()
		app.simulateResponse(loader)
	}
}

func (app *ChatApp) setupEditor() {
	editor := components.NewEditor(app.term, app.handleSubmit)
	app.tui.AddChild(editor)
	app.tui.SetFocus(editor)
}

func (app *ChatApp) Run() {
	welcomeText := components.NewText(bold("Welcome to Simple Chat!"), 1, 1, nil)
	instructionsText := components.NewText("Type your messages below. Commands: /delete (remove last message), /clear (clear all messages)", 1, 0, nil)

	app.tui.AddChild(welcomeText)
	app.tui.AddChild(instructionsText)
	app.setupEditor()
	app.tui.Start()
	select {}
}

func main() {
	app := NewChatApp()
	app.Run()
}
