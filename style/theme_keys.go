package style

// ThemeColor names foreground / non-background tokens in theme JSON.
type ThemeColor string

const (
	ColorAccent              ThemeColor = "accent"
	ColorBorder              ThemeColor = "border"
	ColorBorderAccent        ThemeColor = "borderAccent"
	ColorBorderMuted         ThemeColor = "borderMuted"
	ColorSuccess             ThemeColor = "success"
	ColorError               ThemeColor = "error"
	ColorWarning             ThemeColor = "warning"
	ColorMuted               ThemeColor = "muted"
	ColorDim                 ThemeColor = "dim"
	ColorText                ThemeColor = "text"
	ColorThinkingText        ThemeColor = "thinkingText"
	ColorUserMessageText     ThemeColor = "userMessageText"
	ColorCustomMessageText   ThemeColor = "customMessageText"
	ColorCustomMessageLabel  ThemeColor = "customMessageLabel"
	ColorToolTitle           ThemeColor = "toolTitle"
	ColorToolOutput          ThemeColor = "toolOutput"
	ColorMdHeading           ThemeColor = "mdHeading"
	ColorMdLink              ThemeColor = "mdLink"
	ColorMdLinkURL           ThemeColor = "mdLinkUrl"
	ColorMdCode              ThemeColor = "mdCode"
	ColorMdCodeBlock         ThemeColor = "mdCodeBlock"
	ColorMdCodeBlockBorder   ThemeColor = "mdCodeBlockBorder"
	ColorMdQuote             ThemeColor = "mdQuote"
	ColorMdQuoteBorder       ThemeColor = "mdQuoteBorder"
	ColorMdHr                ThemeColor = "mdHr"
	ColorMdListBullet        ThemeColor = "mdListBullet"
	ColorToolDiffAdded       ThemeColor = "toolDiffAdded"
	ColorToolDiffRemoved     ThemeColor = "toolDiffRemoved"
	ColorToolDiffContext     ThemeColor = "toolDiffContext"
	ColorSyntaxComment       ThemeColor = "syntaxComment"
	ColorSyntaxKeyword       ThemeColor = "syntaxKeyword"
	ColorSyntaxFunction      ThemeColor = "syntaxFunction"
	ColorSyntaxVariable      ThemeColor = "syntaxVariable"
	ColorSyntaxString        ThemeColor = "syntaxString"
	ColorSyntaxNumber        ThemeColor = "syntaxNumber"
	ColorSyntaxType          ThemeColor = "syntaxType"
	ColorSyntaxOperator      ThemeColor = "syntaxOperator"
	ColorSyntaxPunctuation   ThemeColor = "syntaxPunctuation"
	ColorThinkingOff         ThemeColor = "thinkingOff"
	ColorThinkingMinimal     ThemeColor = "thinkingMinimal"
	ColorThinkingLow         ThemeColor = "thinkingLow"
	ColorThinkingMedium      ThemeColor = "thinkingMedium"
	ColorThinkingHigh        ThemeColor = "thinkingHigh"
	ColorThinkingXhigh       ThemeColor = "thinkingXhigh"
	ColorBashMode            ThemeColor = "bashMode"
	ColorPythonMode          ThemeColor = "pythonMode"
	ColorStatusLineSep       ThemeColor = "statusLineSep"
	ColorStatusLineModel     ThemeColor = "statusLineModel"
	ColorStatusLinePath      ThemeColor = "statusLinePath"
	ColorStatusLineGitClean  ThemeColor = "statusLineGitClean"
	ColorStatusLineGitDirty  ThemeColor = "statusLineGitDirty"
	ColorStatusLineContext   ThemeColor = "statusLineContext"
	ColorStatusLineSpend     ThemeColor = "statusLineSpend"
	ColorStatusLineStaged    ThemeColor = "statusLineStaged"
	ColorStatusLineDirty     ThemeColor = "statusLineDirty"
	ColorStatusLineUntracked ThemeColor = "statusLineUntracked"
	ColorStatusLineOutput    ThemeColor = "statusLineOutput"
	ColorStatusLineCost      ThemeColor = "statusLineCost"
	ColorStatusLineSubagents ThemeColor = "statusLineSubagents"
)

// ThemeBg names background tokens resolved to ANSI background sequences.
type ThemeBg string

const (
	BgSelected      ThemeBg = "selectedBg"
	BgUserMessage   ThemeBg = "userMessageBg"
	BgCustomMessage ThemeBg = "customMessageBg"
	BgToolPending   ThemeBg = "toolPendingBg"
	BgToolSuccess   ThemeBg = "toolSuccessBg"
	BgToolError     ThemeBg = "toolErrorBg"
	BgStatusLine    ThemeBg = "statusLineBg"
)

var themeFgKeySet = map[string]struct{}{
	string(ColorAccent): {}, string(ColorBorder): {}, string(ColorBorderAccent): {}, string(ColorBorderMuted): {},
	string(ColorSuccess): {}, string(ColorError): {}, string(ColorWarning): {}, string(ColorMuted): {}, string(ColorDim): {},
	string(ColorText): {}, string(ColorThinkingText): {}, string(ColorUserMessageText): {}, string(ColorCustomMessageText): {},
	string(ColorCustomMessageLabel): {}, string(ColorToolTitle): {}, string(ColorToolOutput): {},
	string(ColorMdHeading): {}, string(ColorMdLink): {}, string(ColorMdLinkURL): {}, string(ColorMdCode): {},
	string(ColorMdCodeBlock): {}, string(ColorMdCodeBlockBorder): {}, string(ColorMdQuote): {}, string(ColorMdQuoteBorder): {},
	string(ColorMdHr): {}, string(ColorMdListBullet): {}, string(ColorToolDiffAdded): {}, string(ColorToolDiffRemoved): {},
	string(ColorToolDiffContext): {}, string(ColorSyntaxComment): {}, string(ColorSyntaxKeyword): {}, string(ColorSyntaxFunction): {},
	string(ColorSyntaxVariable): {}, string(ColorSyntaxString): {}, string(ColorSyntaxNumber): {}, string(ColorSyntaxType): {},
	string(ColorSyntaxOperator): {}, string(ColorSyntaxPunctuation): {}, string(ColorThinkingOff): {}, string(ColorThinkingMinimal): {},
	string(ColorThinkingLow): {}, string(ColorThinkingMedium): {}, string(ColorThinkingHigh): {}, string(ColorThinkingXhigh): {},
	string(ColorBashMode): {}, string(ColorPythonMode): {}, string(ColorStatusLineSep): {}, string(ColorStatusLineModel): {},
	string(ColorStatusLinePath): {}, string(ColorStatusLineGitClean): {}, string(ColorStatusLineGitDirty): {},
	string(ColorStatusLineContext): {}, string(ColorStatusLineSpend): {}, string(ColorStatusLineStaged): {},
	string(ColorStatusLineDirty): {}, string(ColorStatusLineUntracked): {}, string(ColorStatusLineOutput): {},
	string(ColorStatusLineCost): {}, string(ColorStatusLineSubagents): {},
}

var themeBgKeySet = map[string]struct{}{
	string(BgSelected): {}, string(BgUserMessage): {}, string(BgCustomMessage): {},
	string(BgToolPending): {}, string(BgToolSuccess): {}, string(BgToolError): {}, string(BgStatusLine): {},
}

var requiredThemeColorKeys = []string{
	"accent", "border", "borderAccent", "borderMuted", "success", "error", "warning", "muted", "dim", "text", "thinkingText",
	"selectedBg", "userMessageBg", "userMessageText", "customMessageBg", "customMessageText", "customMessageLabel",
	"toolPendingBg", "toolSuccessBg", "toolErrorBg", "toolTitle", "toolOutput",
	"mdHeading", "mdLink", "mdLinkUrl", "mdCode", "mdCodeBlock", "mdCodeBlockBorder", "mdQuote", "mdQuoteBorder", "mdHr", "mdListBullet",
	"toolDiffAdded", "toolDiffRemoved", "toolDiffContext",
	"syntaxComment", "syntaxKeyword", "syntaxFunction", "syntaxVariable", "syntaxString", "syntaxNumber",
	"syntaxType", "syntaxOperator", "syntaxPunctuation",
	"thinkingOff", "thinkingMinimal", "thinkingLow", "thinkingMedium", "thinkingHigh", "thinkingXhigh",
	"bashMode", "pythonMode",
	"statusLineBg", "statusLineSep", "statusLineModel", "statusLinePath", "statusLineGitClean", "statusLineGitDirty",
	"statusLineContext", "statusLineSpend", "statusLineStaged", "statusLineDirty", "statusLineUntracked",
	"statusLineOutput", "statusLineCost", "statusLineSubagents",
}

// ValidThemeColor reports whether name is a known foreground theme token.
func ValidThemeColor(name string) bool {
	_, ok := themeFgKeySet[name]
	return ok
}

// ValidThemeBg reports whether name is a known background theme token.
func ValidThemeBg(name string) bool {
	_, ok := themeBgKeySet[name]
	return ok
}
