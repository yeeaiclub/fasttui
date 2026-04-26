package style

// SymbolPreset selects which glyph set to use for UI symbols.
type SymbolPreset string

const (
	SymbolPresetUnicode SymbolPreset = "unicode"
	SymbolPresetNerd    SymbolPreset = "nerd"
	SymbolPresetAscii   SymbolPreset = "ascii"
)

// UnicodeSymbolMap is the default Unicode / emoji-heavy symbol table.
var UnicodeSymbolMap = map[string]string{
	// Status
	"status.success":  "✔",
	"status.error":    "✘",
	"status.warning":  "⚠",
	"status.info":     "ⓘ",
	"status.pending":  "⏳",
	"status.disabled": "⦸",
	"status.enabled":  "●",
	"status.running":  "⟳",
	"status.shadowed": "◌",
	"status.aborted":  "⏹",
	// Navigation
	"nav.cursor":   "❯",
	"nav.selected": "➤",
	"nav.expand":   "▸",
	"nav.collapse": "▾",
	"nav.back":     "⟵",
	// Tree
	"tree.branch":     "├─",
	"tree.last":       "└─",
	"tree.vertical":   "│",
	"tree.horizontal": "─",
	"tree.hook":       "└",
	// Box (rounded)
	"boxRound.topLeft":     "╭",
	"boxRound.topRight":    "╮",
	"boxRound.bottomLeft":  "╰",
	"boxRound.bottomRight": "╯",
	"boxRound.horizontal":  "─",
	"boxRound.vertical":    "│",
	// Box (sharp)
	"boxSharp.topLeft":     "┌",
	"boxSharp.topRight":    "┐",
	"boxSharp.bottomLeft":  "└",
	"boxSharp.bottomRight": "┘",
	"boxSharp.horizontal":  "─",
	"boxSharp.vertical":    "│",
	"boxSharp.cross":       "┼",
	"boxSharp.teeDown":     "┬",
	"boxSharp.teeUp":       "┴",
	"boxSharp.teeRight":    "├",
	"boxSharp.teeLeft":     "┤",
	// Separators (powerline-ish, but pure Unicode)
	"sep.powerline":          "▕",
	"sep.powerlineThin":      "┆",
	"sep.powerlineLeft":      "▶",
	"sep.powerlineRight":     "◀",
	"sep.powerlineThinLeft":  ">",
	"sep.powerlineThinRight": "<",
	"sep.block":              "▌",
	"sep.space":              " ",
	"sep.asciiLeft":          ">",
	"sep.asciiRight":         "<",
	"sep.dot":                " · ",
	"sep.slash":              " / ",
	"sep.pipe":               " │ ",
	// Icons
	"icon.model":                 "⬢",
	"icon.plan":                  "🗺",
	"icon.folder":                "📁",
	"icon.file":                  "📄",
	"icon.git":                   "⎇",
	"icon.branch":                "⑂",
	"icon.pr":                    "⤴",
	"icon.tokens":                "🪙",
	"icon.context":               "◫",
	"icon.cost":                  "💲",
	"icon.time":                  "⏱",
	"icon.pi":                    "π",
	"icon.agents":                "👥",
	"icon.cache":                 "💾",
	"icon.input":                 "⤵",
	"icon.output":                "⤴",
	"icon.host":                  "🖥",
	"icon.session":               "🆔",
	"icon.package":               "📦",
	"icon.warning":               "⚠",
	"icon.rewind":                "↶",
	"icon.auto":                  "⟲",
	"icon.fast":                  "⚡",
	"icon.extensionSkill":        "✦",
	"icon.extensionTool":         "🛠",
	"icon.extensionSlashCommand": "⌘",
	"icon.extensionMcp":          "🔌",
	"icon.extensionRule":         "⚖",
	"icon.extensionHook":         "🪝",
	"icon.extensionPrompt":       "✎",
	"icon.extensionContextFile":  "📎",
	"icon.extensionInstruction":  "📘",
	// STT
	"icon.mic": "🎤",
	// Thinking levels
	"thinking.minimal": "◔ min",
	"thinking.low":     "◑ low",
	"thinking.medium":  "◒ med",
	"thinking.high":    "◕ high",
	"thinking.xhigh":   "◉ xhi",
	// Checkboxes
	"checkbox.checked":   "☑",
	"checkbox.unchecked": "☐",
	// Formatting
	"format.bullet":       "•",
	"format.dash":         "—",
	"format.bracketLeft":  "⟦",
	"format.bracketRight": "⟧",
	// Markdown
	"md.quoteBorder": "▏",
	"md.hrChar":      "─",
	"md.bullet":      "•",
	// Language/file icons (emoji-centric, no Nerd Font required)
	"lang.default":    "⌘",
	"lang.typescript": "🟦",
	"lang.javascript": "🟨",
	"lang.python":     "🐍",
	"lang.rust":       "🦀",
	"lang.go":         "🐹",
	"lang.java":       "☕",
	"lang.c":          "Ⓒ",
	"lang.cpp":        "➕",
	"lang.csharp":     "♯",
	"lang.ruby":       "💎",
	"lang.php":        "🐘",
	"lang.swift":      "🕊",
	"lang.kotlin":     "🅺",
	"lang.shell":      "💻",
	"lang.html":       "🌐",
	"lang.css":        "🎨",
	"lang.json":       "🧾",
	"lang.yaml":       "📋",
	"lang.markdown":   "📝",
	"lang.sql":        "🗄",
	"lang.docker":     "🐳",
	"lang.lua":        "🌙",
	"lang.text":       "🗒",
	"lang.env":        "🔧",
	"lang.toml":       "🧾",
	"lang.xml":        "⟨⟩",
	"lang.ini":        "⚙",
	"lang.conf":       "⚙",
	"lang.log":        "📜",
	"lang.csv":        "📑",
	"lang.tsv":        "📑",
	"lang.image":      "🖼",
	"lang.pdf":        "📕",
	"lang.archive":    "🗜",
	"lang.binary":     "⚙",
	// Settings tabs
	"tab.appearance":  "🎨",
	"tab.model":       "🤖",
	"tab.interaction": "⌨",
	"tab.context":     "📋",
	"tab.editing":     "💻",
	"tab.tools":       "🔧",
	"tab.tasks":       "📦",
	"tab.providers":   "🌐",
}

// SymbolMap is an alias of [UnicodeSymbolMap] for backward compatibility.
var SymbolMap = UnicodeSymbolMap

var symbolPresetMaps = map[SymbolPreset]map[string]string{
	SymbolPresetUnicode: UnicodeSymbolMap,
	SymbolPresetNerd:    nerdSymbolMap,
	SymbolPresetAscii:   asciiSymbolMap,
}

// AvailableSymbolPresets lists supported symbol presets.
func AvailableSymbolPresets() []SymbolPreset {
	return []SymbolPreset{SymbolPresetUnicode, SymbolPresetNerd, SymbolPresetAscii}
}

// ValidSymbolPreset reports whether s is a known preset name.
func ValidSymbolPreset(s string) bool {
	switch SymbolPreset(s) {
	case SymbolPresetUnicode, SymbolPresetNerd, SymbolPresetAscii:
		return true
	default:
		return false
	}
}
