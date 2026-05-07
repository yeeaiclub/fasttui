package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
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
		"CSI_CURSOR `ESC[A ESC[B ESC[C ESC[D` 光标上下左右移动序列测试",
		"CSI_SAVE `ESC[s ESC[u` 保存恢复光标位置测试",
		"CSI_HIDE `\x1b[?25l隐藏光标\x1b[?25h显示光标` 终端光标控制",
		"LONG_ANSI " + strings.Repeat("\x1b[31m红\x1b[32m绿\x1b[33m黄\x1b[34m蓝\x1b[35m紫", 20) + " 频繁颜色切换",
		"MIX_TABLE " + "| Name | Value | Desc |\n|------|-------|------|\n| A | \x1b[B\x1b[C | cursor |\n| B | \x1b[s\x1b[u | save/restore |",
		"EMOJI_MIX " + strings.Repeat("🎉🔥💻🚀⭐🎯✅❌⚠️📝🔧", 15) + " emoji密集测试",
		"WIDE_CHAR " + strings.Repeat("中文ｶﾞﾛﾝｸﾞ文字ＡＢＣＤ全角", 10) + " 全半角混排",
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
				go app.executeExternalCommand("git", "status")
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

func (app *ChatApp) executeExternalCommand(name string, args ...string) {
	loader := app.addLoader()

	var stdoutBuf, stderrBuf strings.Builder
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdoutStr := strings.TrimSpace(stdoutBuf.String())
	stderrStr := strings.TrimSpace(stderrBuf.String())

	outputParts := []string{fmt.Sprintf("`$ %s %s`", name, strings.Join(args, " "))}
	if stdoutStr != "" {
		outputParts = append(outputParts, stdoutStr)
	}
	if stderrStr != "" {
		outputParts = append(outputParts, fmt.Sprintf("**stderr:** %s", stderrStr))
	}
	if err != nil {
		outputParts = append(outputParts, fmt.Sprintf("❌ Error: %v", err))
	} else {
		outputParts = append(outputParts, "✅ Completed.")
	}

	app.tui.RemoveChild(loader)
	loader.Stop()
	app.addBotMessage(strings.Join(outputParts, "\n\n"))

	app.isResponding = false
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
	case "/git log":
		app.isResponding = true
		app.addUserMessage("git log --oneline -5")
		go app.executeExternalCommand("git", "log", "--oneline", "-5")
		return
	case "/ls":
		app.isResponding = true
		app.addUserMessage("ls -la")
		go app.executeExternalCommand("ls", "-la")
		return
	case "/build":
		app.isResponding = true
		app.addUserMessage("go build ./...")
		go app.executeExternalCommand("go", "build", "./...")
		return
	case "/env":
		app.isResponding = true
		app.addUserMessage("env | head -10")
		go app.executeExternalCommand("sh", "-c", "env | head -10")
		return
	case "/echo":
		app.isResponding = true
		testStr := "CSI test: \x1b[A\x1b[B\x1b[C\x1b[D\x1b[s\x1b[u\x1b[?25l\x1b[?25h done"
		app.addUserMessage("/echo " + testStr)
		go app.executeExternalCommand("echo", "-e", testStr)
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
		components.SlashCommand{
			Name:        "git status",
			Description: "Execute git status (external command)",
		},
		components.SlashCommand{
			Name:        "git log",
			Description: "Show recent git commits (external command)",
		},
		components.SlashCommand{
			Name:        "ls",
			Description: "List files (external command, tests output isolation)",
		},
		components.SlashCommand{
			Name:        "build",
			Description: "Run go build (external command, tests output isolation)",
		},
		components.SlashCommand{
			Name:        "env",
			Description: "Show environment variables (external command)",
		},
		components.SlashCommand{
			Name:        "echo",
			Description: "Echo with CSI sequences (tests ANSI handling)",
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
