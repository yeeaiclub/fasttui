package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/components"
	"github.com/yeeaiclub/fasttui/terminal"
)

func main() {
	// Create sample items
	items := []components.SelectItem{
		{Label: "Git Status", Value: "git-status", Description: "Show the working tree status"},
		{Label: "Git Commit", Value: "git-commit", Description: "Record changes to the repository"},
		{Label: "Git Push", Value: "git-push", Description: "Update remote refs along with associated objects"},
		{Label: "Git Pull", Value: "git-pull", Description: "Fetch from and integrate with another repository"},
		{Label: "Git Branch", Value: "git-branch", Description: "List, create, or delete branches"},
		{Label: "Git Checkout", Value: "git-checkout", Description: "Switch branches or restore working tree files"},
		{Label: "Git Merge", Value: "git-merge", Description: "Join two or more development histories together"},
		{Label: "Git Log", Value: "git-log", Description: "Show commit logs"},
		{Label: "Git Diff", Value: "git-diff", Description: "Show changes between commits, commit and working tree, etc"},
		{Label: "Git Stash", Value: "git-stash", Description: "Stash the changes in a dirty working directory away"},
	}

	// Create theme
	theme := components.SelectListTheme{
		SelectedPrefix: "→ ",
		NormalPrefix:   "  ",
		NoMatch: func(s string) string {
			return "\x1b[2m" + s + "\x1b[0m" // Dim text
		},
		ScrollInfo: func(s string) string {
			return "\x1b[2m" + s + "\x1b[0m" // Dim text
		},
		Description: func(s string) string {
			return "\x1b[2m" + s + "\x1b[0m" // Dim text
		},
	}

	// Create terminal and TUI
	term := terminal.NewProcessTerminal()
	tui := fasttui.NewTUI(term, true)

	done := make(chan struct{})
	var selected *components.SelectItem
	var cancelled bool
	quit := func() {
		select {
		case <-done:
		default:
			close(done)
		}
	}

	// Add title text
	titleText := components.NewText(
		"Git Command Selector",
		2,
		1,
		components.WithTextBackground(func(s string) string {
			return "\x1b[44m\x1b[97m" + s + "\x1b[0m" // Blue background, white text
		}),
	)
	tui.AddChild(titleText)

	// Add border
	tui.AddChild(components.NewDynamicBorder(components.WithBorderColor(func(s string) string { return s })))

	// Create select list
	selectList := components.NewSelectList(items, 8, components.WithSelectListTheme(theme))

	// Set up callbacks — signal main instead of calling Stop from the event loop.
	selectList.SetOnSelect(func(item components.SelectItem) {
		selected = &item
		quit()
	})

	selectList.SetOnCancel(func() {
		cancelled = true
		quit()
	})

	selectList.SetOnSelectionChange(func(item components.SelectItem) {
		// Optional: do something when selection changes
	})

	// Add to TUI
	tui.AddChild(selectList)
	tui.SetFocus(selectList)

	// Add bottom border
	tui.AddChild(components.NewDynamicBorder(components.WithBorderColor(func(s string) string { return s })))

	// Add help text at bottom
	helpText := components.NewText("↑/↓: Navigate | Enter: Select | Esc/Ctrl+C: Cancel", 1, 0)
	tui.AddChild(helpText)

	tui.Start()
	defer tui.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigChan:
	case <-done:
	}

	if selected != nil {
		fmt.Printf("\nYou selected: %s (%s)\n", selected.Label, selected.Value)
	} else if cancelled {
		fmt.Println("\nSelection cancelled")
	}
}
