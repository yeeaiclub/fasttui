package main

import (
	"fmt"
	"os"

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

	// Add border
	tui.AddChild(components.NewDynamicBorder(func(s string) string { return s }))

	// Create select list
	selectList := components.NewSelectList(items, 8, theme)

	// Set up callbacks
	selectList.SetOnSelect(func(item components.SelectItem) {
		tui.Stop()
		fmt.Printf("\nYou selected: %s (%s)\n", item.Label, item.Value)
		os.Exit(0)
	})

	selectList.SetOnCancel(func() {
		tui.Stop()
		fmt.Println("\nSelection cancelled")
		os.Exit(0)
	})

	selectList.SetOnSelectionChange(func(item components.SelectItem) {
		// Optional: do something when selection changes
	})

	// Add to TUI
	tui.AddChild(selectList)
	tui.SetFocus(selectList)

	// Add bottom border
	tui.AddChild(components.NewDynamicBorder(func(s string) string { return s }))

	// Start TUI
	tui.Start()

	// Keep running
	select {}
}
