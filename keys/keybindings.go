package keys

type EditorAction string

const (
	EditorActionCursorUp                 EditorAction = "cursorUp"
	EditorActionCursorDown               EditorAction = "cursorDown"
	EditorActionCursorLeft               EditorAction = "cursorLeft"
	EditorActionCursorRight              EditorAction = "cursorRight"
	EditorActionCursorWordLeft           EditorAction = "cursorWordLeft"
	EditorActionCursorWordRight          EditorAction = "cursorWordRight"
	EditorActionCursorLineStart          EditorAction = "cursorLineStart"
	EditorActionCursorLineEnd            EditorAction = "cursorLineEnd"
	EditorActionJumpForward              EditorAction = "jumpForward"
	EditorActionJumpBackward             EditorAction = "jumpBackward"
	EditorActionPageUp                   EditorAction = "pageUp"
	EditorActionPageDown                 EditorAction = "pageDown"
	EditorActionDeleteCharBackward       EditorAction = "deleteCharBackward"
	EditorActionDeleteCharForward        EditorAction = "deleteCharForward"
	EditorActionDeleteWordBackward       EditorAction = "deleteWordBackward"
	EditorActionDeleteWordForward        EditorAction = "deleteWordForward"
	EditorActionDeleteToLineStart        EditorAction = "deleteToLineStart"
	EditorActionDeleteToLineEnd          EditorAction = "deleteToLineEnd"
	EditorActionNewLine                  EditorAction = "newLine"
	EditorActionSubmit                   EditorAction = "submit"
	EditorActionTab                      EditorAction = "tab"
	EditorActionSelectUp                 EditorAction = "selectUp"
	EditorActionSelectDown               EditorAction = "selectDown"
	EditorActionSelectPageUp             EditorAction = "selectPageUp"
	EditorActionSelectPageDown           EditorAction = "selectPageDown"
	EditorActionSelectConfirm            EditorAction = "selectConfirm"
	EditorActionSelectCancel             EditorAction = "selectCancel"
	EditorActionCopy                     EditorAction = "copy"
	EditorActionYank                     EditorAction = "yank"
	EditorActionYankPop                  EditorAction = "yankPop"
	EditorActionUndo                     EditorAction = "undo"
	EditorActionExpandTools              EditorAction = "expandTools"
	EditorActionToggleSessionPath        EditorAction = "toggleSessionPath"
	EditorActionToggleSessionSort        EditorAction = "toggleSessionSort"
	EditorActionRenameSession            EditorAction = "renameSession"
	EditorActionDeleteSession            EditorAction = "deleteSession"
	EditorActionDeleteSessionNoninvasive EditorAction = "deleteSessionNoninvasive"
)

type EditorKeybindingsConfig map[EditorAction][]string

var defaultEditorKeybindings = map[EditorAction][]string{
	EditorActionCursorUp:                 {"up"},
	EditorActionCursorDown:               {"down"},
	EditorActionCursorLeft:               {"left", "ctrl+b"},
	EditorActionCursorRight:              {"right", "ctrl+f"},
	EditorActionCursorWordLeft:           {"alt+left", "ctrl+left", "alt+b"},
	EditorActionCursorWordRight:          {"alt+right", "ctrl+right", "alt+f"},
	EditorActionCursorLineStart:          {"home", "ctrl+a"},
	EditorActionCursorLineEnd:            {"end", "ctrl+e"},
	EditorActionJumpForward:              {"ctrl+]"},
	EditorActionJumpBackward:             {"ctrl+alt+]"},
	EditorActionPageUp:                   {"pageUp"},
	EditorActionPageDown:                 {"pageDown"},
	EditorActionDeleteCharBackward:       {"backspace"},
	EditorActionDeleteCharForward:        {"delete", "ctrl+d"},
	EditorActionDeleteWordBackward:       {"ctrl+w", "alt+backspace"},
	EditorActionDeleteWordForward:        {"alt+d", "alt+delete"},
	EditorActionDeleteToLineStart:        {"ctrl+u"},
	EditorActionDeleteToLineEnd:          {"ctrl+k"},
	EditorActionNewLine:                  {"shift+enter"},
	EditorActionSubmit:                   {"enter"},
	EditorActionTab:                      {"tab"},
	EditorActionSelectUp:                 {"up"},
	EditorActionSelectDown:               {"down"},
	EditorActionSelectPageUp:             {"pageUp"},
	EditorActionSelectPageDown:           {"pageDown"},
	EditorActionSelectConfirm:            {"enter"},
	EditorActionSelectCancel:             {"escape", "ctrl+c"},
	EditorActionCopy:                     {"ctrl+c"},
	EditorActionYank:                     {"ctrl+y"},
	EditorActionYankPop:                  {"alt+y"},
	EditorActionUndo:                     {"ctrl+-"},
	EditorActionExpandTools:              {"ctrl+o"},
	EditorActionToggleSessionPath:        {"ctrl+p"},
	EditorActionToggleSessionSort:        {"ctrl+s"},
	EditorActionRenameSession:            {"ctrl+r"},
	EditorActionDeleteSession:            {"ctrl+d"},
	EditorActionDeleteSessionNoninvasive: {"ctrl+backspace"},
}

type EditorKeybindingsManager struct {
	actionToKeys map[EditorAction][]string
}

func NewEditorKeybindingsManager(config EditorKeybindingsConfig) *EditorKeybindingsManager {
	mgr := &EditorKeybindingsManager{
		actionToKeys: make(map[EditorAction][]string),
	}
	mgr.buildMaps(config)
	return mgr
}

func (m *EditorKeybindingsManager) buildMaps(config EditorKeybindingsConfig) {
	for k := range m.actionToKeys {
		delete(m.actionToKeys, k)
	}

	for action, keys := range defaultEditorKeybindings {
		m.actionToKeys[action] = append([]string{}, keys...)
	}

	for action, keys := range config {
		if keys != nil {
			m.actionToKeys[action] = append([]string{}, keys...)
		}
	}
}

func (m *EditorKeybindingsManager) Matches(data string, action EditorAction) bool {
	keys, ok := m.actionToKeys[action]
	if !ok {
		return false
	}
	for _, key := range keys {
		if MatchesKey(data, key) {
			return true
		}
	}
	return false
}

func (m *EditorKeybindingsManager) GetKeys(action EditorAction) []string {
	if keys, ok := m.actionToKeys[action]; ok {
		return append([]string{}, keys...)
	}
	return []string{}
}

func (m *EditorKeybindingsManager) SetConfig(config EditorKeybindingsConfig) {
	m.buildMaps(config)
}

var globalEditorKeybindings *EditorKeybindingsManager

func GetEditorKeybindings() *EditorKeybindingsManager {
	if globalEditorKeybindings == nil {
		globalEditorKeybindings = NewEditorKeybindingsManager(nil)
	}
	return globalEditorKeybindings
}

func SetEditorKeybindings(manager *EditorKeybindingsManager) {
	globalEditorKeybindings = manager
}
