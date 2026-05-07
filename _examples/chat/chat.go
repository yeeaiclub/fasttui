package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strings"
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

func (app *ChatApp) handleDeleteCommand() {
	// Remove the last message (before editor)
	children := app.tui.GetChildren()
	// Structure: welcome, instructions, [messages...], editor, footer
	// Messages start at index 2, editor is at len-2, footer at len-1
	// Last message is at len-3
	if len(children) > 4 {
		app.tui.RemoveChildAt(len(children) - 3)
	}
	app.tui.TriggerRender()
}

func (app *ChatApp) handleClearCommand() {
	// Structure: welcome, instructions, [messages...], editor, footer
	// Keep welcome(0), instructions(1), editor(len-2), footer(len-1)
	children := app.tui.GetChildren()
	for len(children) > 4 {
		app.tui.RemoveChildAt(2)
		children = app.tui.GetChildren()
	}
	app.tui.TriggerRender()
}

func (app *ChatApp) addUserMessage(value string) {
	userMessage := components.NewMarkdown(
		value,
		1,
		1,
		components.WithMarkdownTheme(app.theme),
	)
	// Insert before editor (editor is at len-2, footer at len-1)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-2, userMessage)
}

func (app *ChatApp) addLoader() *components.Loader {
	loader := components.NewLoader(
		app.tui,
		"Thinking...",
		components.WithLoaderSpinnerColor(cyan),
		components.WithLoaderMessageColor(dim),
	)
	// Insert before editor (editor is at len-2, footer at len-1)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-2, loader)
	app.tui.TriggerRender()
	return loader
}

func (app *ChatApp) simulateResponse(loader *components.Loader) {
	time.Sleep(1 * time.Second)

	app.tui.RemoveChild(loader)
	loader.Stop()

	responses := []string{
		"WRAP_EN " + strings.Repeat("The_quick_brown_fox_jumps_over_the_lazy_dog_0123456789_", 14),
		"WRAP_ZH " + strings.Repeat("这是一段连续中文用于测试换行切割与终端列宽", 12) + " 结束",
		"WRAP_MIX " + strings.Repeat("混合Mixed中文English数字123标点。", 10),
		"**" + strings.Repeat("加粗中文折行测试一二三四五六七八九十。", 8) + "**",
		"## That's interesting! Tell me more.",
		"**I see what you mean.**",
		"#### Fascinating perspective!",
		"`Could you elaborate on that?`",
		"## That makes sense to me.",
		"## I hadn't thought of it that way.",
		"## Great point!",
		"**Thanks for sharing that.**\nyes yes yes",
	}
	randomResponse := responses[rand.Intn(len(responses))]

	botMessage := components.NewMarkdown(
		randomResponse,
		1,
		1,
		components.WithMarkdownTheme(app.theme),
	)
	// Insert before editor (editor is at len-2, footer at len-1)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-2, botMessage)

	app.isResponding = false
	app.tui.TriggerRender()
}

func (app *ChatApp) showGitStatusConfirm() {
	app.selector = NewExtensionSelectorComponent(
		"Execute git status?",
		[]string{"Yes", "No"},
		func(option string) {
			app.hideSelector()
			if option == "Yes" {
				app.addUserMessage("git status")
				app.addBotMessage("Executing `git status`... (simulated)")
			}
			app.isResponding = false
		},
		func() {
			app.hideSelector()
			app.isResponding = false
		},
		nil,
	)
	app.tui.AddChild(app.selector)
	app.tui.SetFocus(app.selector)
	app.tui.TriggerRender()
}

func (app *ChatApp) hideSelector() {
	if app.selector != nil {
		app.selector.Dispose()
		app.tui.RemoveChild(app.selector)
		app.selector = nil
	}
	app.tui.SetFocus(app.editor)
	app.tui.TriggerRender()
}

func (app *ChatApp) addBotMessage(text string) {
	botMessage := components.NewMarkdown(
		text,
		1,
		1,
		components.WithMarkdownTheme(app.theme),
	)
	// Insert before editor (editor is at len-2, footer at len-1)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-2, botMessage)
	app.tui.TriggerRender()
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
	case "/git status":
		app.isResponding = true
		app.showGitStatusConfirm()
		return
	}

	if trimmed != "" {
		app.isResponding = true
		app.addUserMessage(value)
		loader := app.addLoader()
		go app.simulateResponse(loader)
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
		// Double Ctrl+C to exit
		now := time.Now()
		if app.cancelCount > 0 && now.Sub(app.lastCancelTime) < 2*time.Second {
			// Second Ctrl+C within 2 seconds - exit gracefully
			app.exit()
			return
		}
		// First Ctrl+C or timeout - show hint and reset
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
	// Give terminal time to fully restore state
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
