# fasttui

fasttui is a Go port of @mariozechner/pi-tui: the same differential renderer, bracketed paste handling, autocomplete, and component value semantics

## Features

- **Differential Rendering**: Three-strategy rendering system that only updates what changed
- **Synchronized Output**: Uses CSI 2026 for atomic screen updates (no flicker)
- **Bracketed Paste Mode**: Handles large pastes correctly with markers for >10 line pastes
- **Component-based**: Simple Component interface with render() method
- **Theme Support**: Components accept theme interfaces for customizable styling
- **Built-in Components**: Text, TruncatedText, Input, Editor, Markdown, Loader, SelectList, SettingsList, Spacer, Image, Box, Container
- **Inline Images**: Renders images in terminals that support Kitty or iTerm2 graphics protocols
- **Autocomplete Support**: File paths and slash commands

## Installation

```bash
go get github.com/yeeaiclub/fasttui
```

## Quick Start

```go
package main

import (
    "github.com/yeeaiclub/fasttui"
    "github.com/yeeaiclub/fasttui/components"
    "github.com/yeeaiclub/fasttui/terminal"
)

func main() {
    // Create terminal and TUI
    term := terminal.NewProcessTerminal()
    tui := fasttui.NewTUI(term, false)

    // Add components
    text := components.NewText("Hello, fasttui!", 1, 1, nil)
    tui.AddChild(text)

    // Start TUI
    tui.Start()
    select {}
}
```

## Examples

### Select List

A fuzzy searchable select list with keyboard navigation:

```go
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
        {Label: "Git Push", Value: "git-push", Description: "Update remote refs"},
    }

    // Create theme
    theme := components.SelectListTheme{
        SelectedPrefix: "→ ",
        NormalPrefix:   "  ",
        NoMatch: func(s string) string {
            return "\x1b[2m" + s + "\x1b[0m" // Dim text
        },
    }

    // Create terminal and TUI
    term := terminal.NewProcessTerminal()
    tui := fasttui.NewTUI(term, true)

    // Create select list
    selectList := components.NewSelectList(items, 8, theme)
    
    // Set up callbacks
    selectList.SetOnSelect(func(item components.SelectItem) {
        tui.Stop()
        fmt.Printf("\nYou selected: %s\n", item.Label)
        os.Exit(0)
    })

    selectList.SetOnCancel(func() {
        tui.Stop()
        fmt.Println("\nSelection cancelled")
        os.Exit(0)
    })

    // Add to TUI and set focus
    tui.AddChild(selectList)
    tui.SetFocus(selectList)

    // Start TUI
    tui.Start()
    select {}
}
```

### Chat Interface

A simple chat interface with an editor and markdown rendering:

```go
type ChatApp struct {
    tui  *fasttui.TUI
    term *terminal.ProcessTerminal
}

func NewChatApp() *ChatApp {
    term := terminal.NewProcessTerminal()
    tui := fasttui.NewTUI(term, false)

    return &ChatApp{
        tui:  tui,
        term: term,
    }
}

func (app *ChatApp) handleSubmit(value string) {
    // Handle user input
    userMessage := components.NewMarkdown(value, 1, 1, theme, nil)
    app.tui.AddChild(userMessage)
    app.tui.RequestRender(false)
}

func (app *ChatApp) Run() {
    // Add welcome text
    welcomeText := components.NewText("Welcome to Chat!", 1, 1, nil)
    app.tui.AddChild(welcomeText)

    // Setup editor for input
    editor := components.NewEditor(app.term, app.handleSubmit)
    app.tui.AddChild(editor)
    app.tui.SetFocus(editor)

    // Start TUI
    app.tui.Start()
    select {}
}
```

### Key Logger

Test keyboard input and see key codes:

```go
type KeyLogger struct {
    tui *fasttui.TUI
}

func (k *KeyLogger) HandleInput(data string) {
    if keys.MatchesKey(data, "ctrl+c") {
        k.tui.Stop()
        os.Exit(0)
    }
    // Process key input...
}

func main() {
    term := terminal.NewProcessTerminal()
    tui := fasttui.NewTUI(term, false)
    
    keyLogger := &KeyLogger{tui: tui}
    tui.AddChild(keyLogger)
    
    tui.Start()
    select {}
}
```

## Built-in Components

### Text
Display static text with optional styling:

```go
text := components.NewText("Hello World", paddingTop, paddingBottom, styleFunc)
```

### Editor
Multi-line text input with submit handling:

```go
editor := components.NewEditor(terminal, onSubmitFunc)
tui.SetFocus(editor)
```

### SelectList
Fuzzy searchable list with keyboard navigation:

```go
items := []components.SelectItem{
    {Label: "Option 1", Value: "opt1", Description: "Description"},
}
selectList := components.NewSelectList(items, visibleItems, theme)
```

### Markdown
Render markdown content:

```go
md := components.NewMarkdown(content, paddingTop, paddingBottom, theme, imageFetcher)
```

### Loader
Animated loading indicator:

```go
loader := components.NewLoader(tui, colorFunc, dimFunc, "Loading...")
loader.Stop()
```

### Border
Add decorative borders:

```go
border := components.NewDynamicBorder(styleFunc)
tui.AddChild(border)
```

## Styling

Use ANSI escape codes for styling:

```go
func cyan(s string) string {
    return "\x1b[36m" + s + "\x1b[0m"
}

func bold(s string) string {
    return "\x1b[1m" + s + "\x1b[0m"
}

func dim(s string) string {
    return "\x1b[2m" + s + "\x1b[0m"
}
```

Common ANSI codes:
- `\x1b[1m` - Bold
- `\x1b[2m` - Dim
- `\x1b[3m` - Italic
- `\x1b[4m` - Underline
- `\x1b[30m` to `\x1b[37m` - Foreground colors (black to white)
- `\x1b[40m` to `\x1b[47m` - Background colors
- `\x1b[0m` - Reset

## Running Examples

```bash
# Select list example
go run _examples/select/select.go

# Chat example
go run _examples/chat/chat.go _examples/chat/theme.go

# Key logger example
go run _examples/key/key_logger.go
```

## License

Apache License 2.0
