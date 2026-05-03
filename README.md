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

## Theme

Colors and terminal glyphs are provided by the **`style`** subpackage (`github.com/yeeaiclub/fasttui/style`). A theme is a JSON file that lists named color tokens, optional `vars` for indirection, and optional symbol / export settings.

### Loading a theme

```go
import "github.com/yeeaiclub/fasttui/style"

// Built-in names are embedded in the library (e.g. "dark", "light", and many *defaults*).
th, err := style.LoadTheme("dark")
if err != nil {
	// handle error
}
// Optional configuration uses the options pattern, e.g.:
//   style.LoadTheme("dark", style.WithColorMode(style.ColorModeTruecolor))
//   style.LoadTheme("dark", style.WithSymbolPreset(style.SymbolPresetNerd), style.WithColorBlindMode(true))
// Available: [style.WithColorMode], [style.WithSymbolPreset], [style.WithColorBlindMode]
```

- **`LoadThemeFile(name)`** – returns `*ThemeFile` (parsed JSON) without ANSI resolution.
- **`NewTheme(tf, opts ...ThemeOption)`** – builds a `*Theme` from `*ThemeFile` (same options as `LoadTheme`).
- **`ParseThemeJSON(data)`** – parse raw JSON (same validation as file load).

### Where themes are loaded from

1. **Built-in** – themes shipped under `style/theme` (e.g. `dark.json`, `light.json`, `theme/defaults/*.json`) are **embedded** and resolved first by name (filename without `.json`).
2. **User themes** – `<config-dir>/fasttui/themes/<name>.json`. On most systems `<config-dir>` is `$XDG_CONFIG_HOME` or OS user config. Override the directory with **`FASTTUI_THEMES_DIR`**.

**`ListThemeNames()`** returns all built-in names plus any `*.json` in the user themes directory (sorted, de-duplicated).

### Theme JSON

- **`name`** (string) and **`colors`** (object) are required. Every key listed in the schema must be present in `colors` (e.g. `accent`, `userMessageBg`, `syntaxComment`, `statusLineBg`, …).
- **`vars`** (optional) – map of name → hex string, `""`, or 0–255 index. Other fields can reference a var by the **same string** as the key (e.g. `"accent": "teal"` with `"vars": { "teal": "#5a8080" }`).
- **`export`** (optional) – `pageBg` / `cardBg` / `infoBg` for HTML/CSS export helpers; var refs may use a **`$` prefix** (e.g. `"$teal"`) when matching `vars`.
- **`symbols`** (optional) – `preset` (`unicode` | `nerd` | `ascii`) and `overrides` for individual logical keys (e.g. `status.success`).

The JSON schema is embedded as **`style/theme/theme-schema.json`** in the module.

### Using `*style.Theme`

`Theme` exposes **foreground** and **background** tokens as ANSI, plus symbols:

- `Fg(color ThemeColor, text string)`, `Bg(bg ThemeBg, text string)`
- `FgANSI` / `BgANSI` for raw sequences; `Bold`, `Italic`, `Underline`, etc.
- `Symbol(key string)`, `LangIcon(lang string)`, `SpinnerFrames(kind)`, `InputCursor()`, …

For non-TUI use (e.g. HTML), **`ResolvedThemeColors`**, **`ExportColors`**, **`IsLightTheme`**, and **`DefaultThemeName`** (uses `COLORFGBG` when set) are available in the same package.

## License

Apache License 2.0
