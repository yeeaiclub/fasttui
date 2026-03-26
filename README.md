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

## quickstart

### Installation

```bash
go get github.com/yourusername/fasttui
```


### components

all components must implement this interface.

```go
type Component interface {
    // render the compoent to lines for the given viewport width
    Render(width int) []string
	// handler for keyboard input when component has focus
    HandleInput(data string)
    // receives key release events
	WantsKeyRelease() bool
    // Invalidate any cached rendering state.
	// Called when theme changes or when component needs to re-render from scratch.
	Invalidate()
}
```

## TUI

Create a **terminal** implementation (stdin/stdout TTY), then a **TUI** that owns rendering and input. Call `Start()` after adding children and optional `SetFocus`.

```go
import (
	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/terminal"
)

term := terminal.NewProcessTerminal()

// Second arg: show hardware cursor (true) vs hide it and rely on drawn UI (false).
tui := fasttui.NewTUI(term, false)
```

## Add Child

`TUI` embeds `Container`, so you use `AddChild` to stack components; the TUI owns the render loop and paints the screen.

```go
tui.AddChild(component) // Append to the vertical layout (order = top to bottom).

// Optional: give keyboard focus to a Focusable component (Input, Editor, lists, …).
tui.SetFocus(component)

tui.Start() // Enter raw mode, start input + differential rendering.

defer tui.Stop() // Restore the terminal when the program exits.
```

## render

Manually trigger a render using `tui.TriggerRender()`.
This is useful when you update component state outside the normal input flow.
```go
tui.TriggerRender()
```

## License

Apache License 2.0
