package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/yeeaiclub/fasttui/components"
)

func (app *ChatApp) addUserMessage(value string) {
	userMessage := components.NewMarkdown(
		value,
		1,
		1,
		components.WithMarkdownTheme(app.theme),
	)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-2, userMessage)
}

func (app *ChatApp) addBotMessage(text string) {
	botMessage := components.NewMarkdown(
		text,
		1,
		1,
		components.WithMarkdownTheme(app.theme),
	)
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-2, botMessage)
	app.tui.TriggerRender()
}

func (app *ChatApp) addLoader() *components.Loader {
	loader := components.NewLoader(
		app.tui,
		"Thinking...",
		components.WithLoaderSpinnerColor(cyan),
		components.WithLoaderMessageColor(dim),
	)
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
	children := app.tui.GetChildren()
	app.tui.InsertChildAt(len(children)-2, botMessage)

	app.isResponding = false
	app.tui.TriggerRender()
}