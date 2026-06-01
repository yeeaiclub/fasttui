package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

type ChatApp struct {
	tui            *fasttui.TUI
	term           *terminal.ProcessTerminal
	isResponding   bool
	theme          *components.MarkdownTheme
	cancelCount    int
	lastCancelTime time.Time
	selector       *ExtensionSelectorComponent
	editor         *components.Editor
	footer         *FooterComponent
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

func (app *ChatApp) setupEditor() {
	app.editor = components.NewEditor(app.term, app.handleSubmit)

	commands := []any{
		components.SlashCommand{
			Name:        "delete",
			Description: "Remove last message",
		},
		components.SlashCommand{
			Name:        "clear",
			Description: "Clear all messages",
		},
		components.AutocompleteItem{
			Value:       "git status",
			Label:       "git status",
			Description: "Show the working tree status",
		},
	}

	selectTheme := components.SelectListTheme{
		SelectedPrefix: "→ ",
		NormalPrefix:   "  ",
		NoMatch: func(s string) string {
			return "\x1b[2m" + s + "\x1b[0m"
		},
		ScrollInfo: func(s string) string {
			return "\x1b[2m" + s + "\x1b[0m"
		},
		Description: func(s string) string {
			return "\x1b[2m" + s + "\x1b[0m"
		},
	}

	provider := components.NewCombinedAutocompleteProvider(commands, "", "fd")
	app.editor.SetAutocomplete(provider, selectTheme, 8)

	app.editor.OnCancel = func() {
		now := time.Now()
		if app.cancelCount > 0 && now.Sub(app.lastCancelTime) < 2*time.Second {
			app.exit()
			return
		}
		app.cancelCount = 1
		app.lastCancelTime = now
	}
	app.tui.AddChild(app.editor)
	app.tui.SetFocus(app.editor)
}

func (app *ChatApp) setupFooter() {
	app.footer = NewFooterComponent(0, "gpt-4")
	app.tui.AddChild(app.footer)
}

func (app *ChatApp) exit() {
	app.tui.Stop()
	time.Sleep(200 * time.Millisecond)
	os.Exit(0)
}

func (app *ChatApp) Run() {
	welcomeText := components.NewText(bold("Welcome to Simple Chat!"), 1, 1)
	instructionsText := components.NewText("Type your messages below. Commands: /delete (remove last message), /clear (clear all messages)", 1, 0)

	app.tui.AddChild(welcomeText)
	app.tui.AddChild(instructionsText)
	app.setupEditor()
	app.setupFooter()
	app.tui.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}

func main() {
	app := NewChatApp()
	app.Run()
}
