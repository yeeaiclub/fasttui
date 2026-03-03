package main

import (
	"github.com/yeeaiclub/fasttui/keys"
)

func RawKeyHint(key string, description string) string {
	return "[" + key + "] " + description
}

func KeyHint(action string, description string) string {
	kb := keys.GetEditorKeybindings()
	keyStr := ""
	switch action {
	case "selectConfirm":
		keyStr = getFirstKey(kb.GetKeys(keys.EditorActionSelectConfirm))
	case "selectCancel":
		keyStr = getFirstKey(kb.GetKeys(keys.EditorActionSelectCancel))
	default:
		keyStr = action
	}
	return "[" + keyStr + "] " + description
}

func getFirstKey(keys []string) string {
	if len(keys) > 0 {
		return keys[0]
	}
	return "?"
}
