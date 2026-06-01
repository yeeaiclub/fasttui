package main

import (
	"strings"
)

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

func (app *ChatApp) handleDeleteCommand() {
	children := app.tui.GetChildren()
	if len(children) > 4 {
		app.tui.RemoveChildAt(len(children) - 3)
	}
	app.tui.TriggerRender()
}

func (app *ChatApp) handleClearCommand() {
	children := app.tui.GetChildren()
	for len(children) > 4 {
		app.tui.RemoveChildAt(2)
		children = app.tui.GetChildren()
	}
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