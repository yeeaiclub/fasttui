package keys

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	kittyProtocolActive              = false
	lastEventType       KeyEventType = "press"
)

type KeyEventType string

const (
	KeyPress   KeyEventType = "press"
	KeyRepeat  KeyEventType = "repeat"
	KeyRelease KeyEventType = "release"
)

type ParsedKittySequence struct {
	Codepoint     int
	ShiftedKey    *int
	BaseLayoutKey *int
	Modifier      int
	EventType     KeyEventType
}

const (
	ModifierShift = 1
	ModifierAlt   = 2
	ModifierCtrl  = 4
	LockMask      = 64 + 128
)

const (
	CodepointEscape    = 27
	CodepointTab       = 9
	CodepointEnter     = 13
	CodepointSpace     = 32
	CodepointBackspace = 127
	CodepointKPEnter   = 57414
)

const (
	ArrowCodepointUp    = -1
	ArrowCodepointDown  = -2
	ArrowCodepointRight = -3
	ArrowCodepointLeft  = -4
)

const (
	FunctionalCodepointDelete   = -10
	FunctionalCodepointInsert   = -11
	FunctionalCodepointPageUp   = -12
	FunctionalCodepointPageDown = -13
	FunctionalCodepointHome     = -14
	FunctionalCodepointEnd      = -15
)

var symbolKeys = map[rune]bool{
	'`': true, '-': true, '=': true, '[': true, ']': true,
	'\\': true, ';': true, '\'': true, ',': true, '.': true,
	'/': true, '!': true, '@': true, '#': true, '$': true,
	'%': true, '^': true, '&': true, '*': true, '(': true,
	')': true, '_': true, '+': true, '|': true, '~': true,
	'{': true, '}': true, ':': true, '<': true, '>': true,
	'?': true,
}

var legacyKeySequences = map[string][]string{
	"up":       {"\x1b[A", "\x1bOA"},
	"down":     {"\x1b[B", "\x1bOB"},
	"right":    {"\x1b[C", "\x1bOC"},
	"left":     {"\x1b[D", "\x1bOD"},
	"home":     {"\x1b[H", "\x1bOH", "\x1b[1~", "\x1b[7~"},
	"end":      {"\x1b[F", "\x1bOF", "\x1b[4~", "\x1b[8~"},
	"insert":   {"\x1b[2~"},
	"delete":   {"\x1b[3~"},
	"pageup":   {"\x1b[5~", "\x1b[[5~"},
	"pagedown": {"\x1b[6~", "\x1b[[6~"},
	"clear":    {"\x1b[E", "\x1bOE"},
	"f1":       {"\x1bOP", "\x1b[11~", "\x1b[[A"},
	"f2":       {"\x1bOQ", "\x1b[12~", "\x1b[[B"},
	"f3":       {"\x1bOR", "\x1b[13~", "\x1b[[C"},
	"f4":       {"\x1bOS", "\x1b[14~", "\x1b[[D"},
	"f5":       {"\x1b[15~", "\x1b[[E"},
	"f6":       {"\x1b[17~"},
	"f7":       {"\x1b[18~"},
	"f8":       {"\x1b[19~"},
	"f9":       {"\x1b[20~"},
	"f10":      {"\x1b[21~"},
	"f11":      {"\x1b[23~"},
	"f12":      {"\x1b[24~"},
}

var legacyShiftSequences = map[string][]string{
	"up":       {"\x1b[a"},
	"down":     {"\x1b[b"},
	"right":    {"\x1b[c"},
	"left":     {"\x1b[d"},
	"clear":    {"\x1b[e"},
	"insert":   {"\x1b[2$"},
	"delete":   {"\x1b[3$"},
	"pageup":   {"\x1b[5$"},
	"pagedown": {"\x1b[6$"},
	"home":     {"\x1b[7$"},
	"end":      {"\x1b[8$"},
}

var legacyCtrlSequences = map[string][]string{
	"up":       {"\x1bOa"},
	"down":     {"\x1bOb"},
	"right":    {"\x1bOc"},
	"left":     {"\x1bOd"},
	"clear":    {"\x1bOe"},
	"insert":   {"\x1b[2^"},
	"delete":   {"\x1b[3^"},
	"pageup":   {"\x1b[5^"},
	"pagedown": {"\x1b[6^"},
	"home":     {"\x1b[7^"},
	"end":      {"\x1b[8^"},
}

var legacySequenceKeyIDs = map[string]string{
	"\x1bOA":   "up",
	"\x1bOB":   "down",
	"\x1bOC":   "right",
	"\x1bOD":   "left",
	"\x1bOH":   "home",
	"\x1bOF":   "end",
	"\x1b[E":   "clear",
	"\x1bOE":   "clear",
	"\x1bOe":   "ctrl+clear",
	"\x1b[e":   "shift+clear",
	"\x1b[2~":  "insert",
	"\x1b[2$":  "shift+insert",
	"\x1b[2^":  "ctrl+insert",
	"\x1b[3$":  "shift+delete",
	"\x1b[3^":  "ctrl+delete",
	"\x1b[[5~": "pageup",
	"\x1b[[6~": "pagedown",
	"\x1b[a":   "shift+up",
	"\x1b[b":   "shift+down",
	"\x1b[c":   "shift+right",
	"\x1b[d":   "shift+left",
	"\x1bOa":   "ctrl+up",
	"\x1bOb":   "ctrl+down",
	"\x1bOc":   "ctrl+right",
	"\x1bOd":   "ctrl+left",
	"\x1b[5$":  "shift+pageup",
	"\x1b[6$":  "shift+pagedown",
	"\x1b[7$":  "shift+home",
	"\x1b[8$":  "shift+end",
	"\x1b[5^":  "ctrl+pageup",
	"\x1b[6^":  "ctrl+pagedown",
	"\x1b[7^":  "ctrl+home",
	"\x1b[8^":  "ctrl+end",
	"\x1bOP":   "f1",
	"\x1bOQ":   "f2",
	"\x1bOR":   "f3",
	"\x1bOS":   "f4",
	"\x1b[11~": "f1",
	"\x1b[12~": "f2",
	"\x1b[13~": "f3",
	"\x1b[14~": "f4",
	"\x1b[[A":  "f1",
	"\x1b[[B":  "f2",
	"\x1b[[C":  "f3",
	"\x1b[[D":  "f4",
	"\x1b[[E":  "f5",
	"\x1b[15~": "f5",
	"\x1b[17~": "f6",
	"\x1b[18~": "f7",
	"\x1b[19~": "f8",
	"\x1b[20~": "f9",
	"\x1b[21~": "f10",
	"\x1b[23~": "f11",
	"\x1b[24~": "f12",
	"\x1bb":    "alt+left",
	"\x1bf":    "alt+right",
	"\x1bp":    "alt+up",
	"\x1bn":    "alt+down",
}

func SetKittyProtocolActive(active bool) {
	kittyProtocolActive = active
}

func IsKittyProtocolActive() bool {
	return kittyProtocolActive
}

func IsKeyRelease(data string) bool {
	if strings.Contains(data, "\x1b[200~") {
		return false
	}

	releasePatterns := []string{":3u", ":3~", ":3A", ":3B", ":3C", ":3D", ":3H", ":3F"}
	for _, pattern := range releasePatterns {
		if strings.Contains(data, pattern) {
			return true
		}
	}
	return false
}

func IsKeyRepeat(data string) bool {
	if strings.Contains(data, "\x1b[200~") {
		return false
	}

	repeatPatterns := []string{":2u", ":2~", ":2A", ":2B", ":2C", ":2D", ":2H", ":2F"}
	for _, pattern := range repeatPatterns {
		if strings.Contains(data, pattern) {
			return true
		}
	}
	return false
}

func parseEventType(eventTypeStr string) KeyEventType {
	if eventTypeStr == "" {
		return KeyPress
	}
	eventType, _ := strconv.Atoi(eventTypeStr)
	switch eventType {
	case 2:
		return KeyRepeat
	case 3:
		return KeyRelease
	default:
		return KeyPress
	}
}

func ParseKittySequence(data string) *ParsedKittySequence {
	csiUMatch := regexp.MustCompile(`^\x1b\[(\d+)(?::(\d*))?(?::(\d+))?(?:;(\d+))?(?::(\d+))?u$`).FindStringSubmatch(data)
	if csiUMatch != nil {
		codepoint, _ := strconv.Atoi(csiUMatch[1])
		var shiftedKey *int
		if csiUMatch[2] != "" {
			sk, _ := strconv.Atoi(csiUMatch[2])
			shiftedKey = &sk
		}
		var baseLayoutKey *int
		if csiUMatch[3] != "" {
			bl, _ := strconv.Atoi(csiUMatch[3])
			baseLayoutKey = &bl
		}
		modValue := 1
		if csiUMatch[4] != "" {
			modValue, _ = strconv.Atoi(csiUMatch[4])
		}
		eventType := parseEventType(csiUMatch[5])
		lastEventType = eventType
		return &ParsedKittySequence{
			Codepoint:     codepoint,
			ShiftedKey:    shiftedKey,
			BaseLayoutKey: baseLayoutKey,
			Modifier:      modValue - 1,
			EventType:     eventType,
		}
	}

	arrowMatch := regexp.MustCompile(`^\x1b\[1;(\d+)(?::(\d+))?([ABCD])$`).FindStringSubmatch(data)
	if arrowMatch != nil {
		modValue, _ := strconv.Atoi(arrowMatch[1])
		eventType := parseEventType(arrowMatch[2])
		arrowCodes := map[string]int{"A": ArrowCodepointUp, "B": ArrowCodepointDown, "C": ArrowCodepointRight, "D": ArrowCodepointLeft}
		lastEventType = eventType
		return &ParsedKittySequence{
			Codepoint: arrowCodes[arrowMatch[3]],
			Modifier:  modValue - 1,
			EventType: eventType,
		}
	}

	funcMatch := regexp.MustCompile(`^\x1b\[(\d+)(?:;(\d+))?(?::(\d+))?~$`).FindStringSubmatch(data)
	if funcMatch != nil {
		keyNum, _ := strconv.Atoi(funcMatch[1])
		modValue := 1
		if funcMatch[2] != "" {
			modValue, _ = strconv.Atoi(funcMatch[2])
		}
		eventType := parseEventType(funcMatch[3])
		funcCodes := map[int]int{
			2: FunctionalCodepointInsert,
			3: FunctionalCodepointDelete,
			5: FunctionalCodepointPageUp,
			6: FunctionalCodepointPageDown,
			7: FunctionalCodepointHome,
			8: FunctionalCodepointEnd,
		}
		if codepoint, ok := funcCodes[keyNum]; ok {
			lastEventType = eventType
			return &ParsedKittySequence{
				Codepoint: codepoint,
				Modifier:  modValue - 1,
				EventType: eventType,
			}
		}
	}

	homeEndMatch := regexp.MustCompile(`^\x1b\[1;(\d+)(?::(\d+))?([HF])$`).FindStringSubmatch(data)
	if homeEndMatch != nil {
		modValue, _ := strconv.Atoi(homeEndMatch[1])
		eventType := parseEventType(homeEndMatch[2])
		codepoint := FunctionalCodepointHome
		if homeEndMatch[3] == "F" {
			codepoint = FunctionalCodepointEnd
		}
		lastEventType = eventType
		return &ParsedKittySequence{
			Codepoint: codepoint,
			Modifier:  modValue - 1,
			EventType: eventType,
		}
	}

	return nil
}

func matchesKittySequence(data string, expectedCodepoint int, expectedModifier int) bool {
	parsed := ParseKittySequence(data)
	if parsed == nil {
		return false
	}
	actualMod := parsed.Modifier & ^LockMask
	expectedMod := expectedModifier & ^LockMask

	if actualMod != expectedMod {
		return false
	}

	if parsed.Codepoint == expectedCodepoint {
		return true
	}

	if parsed.BaseLayoutKey != nil && *parsed.BaseLayoutKey == expectedCodepoint {
		cp := parsed.Codepoint
		isLatinLetter := cp >= 97 && cp <= 122
		isKnownSymbol := symbolKeys[rune(cp)]
		if !isLatinLetter && !isKnownSymbol {
			return true
		}
	}

	return false
}

func matchesModifyOtherKeys(data string, expectedKeycode int, expectedModifier int) bool {
	match := regexp.MustCompile(`^\x1b\[27;(\d+);(\d+)~$`).FindStringSubmatch(data)
	if match == nil {
		return false
	}
	modValue, _ := strconv.Atoi(match[1])
	keycode, _ := strconv.Atoi(match[2])
	actualMod := modValue - 1
	return keycode == expectedKeycode && actualMod == expectedModifier
}

func rawCtrlChar(key string) string {
	char := strings.ToLower(key)
	code := int(char[0])
	if (code >= 97 && code <= 122) || char == "[" || char == "\\" || char == "]" || char == "_" {
		return string(rune(code & 0x1f))
	}
	if char == "-" {
		return string(rune(31))
	}
	return ""
}

func parseKeyId(keyId string) *struct {
	key   string
	ctrl  bool
	shift bool
	alt   bool
} {
	parts := strings.Split(strings.ToLower(keyId), "+")
	if len(parts) == 0 {
		return nil
	}
	key := parts[len(parts)-1]
	if key == "" {
		return nil
	}
	return &struct {
		key   string
		ctrl  bool
		shift bool
		alt   bool
	}{
		key:   key,
		ctrl:  slices.Contains(parts, "ctrl"),
		shift: slices.Contains(parts, "shift"),
		alt:   slices.Contains(parts, "alt"),
	}
}

func matchesLegacySequence(data string, sequences []string) bool {
	for _, seq := range sequences {
		if data == seq {
			return true
		}
	}
	return false
}

func matchesLegacyModifierSequence(data string, key string, modifier int) bool {
	if modifier == ModifierShift {
		if sequences, ok := legacyShiftSequences[key]; ok {
			return matchesLegacySequence(data, sequences)
		}
	}
	if modifier == ModifierCtrl {
		if sequences, ok := legacyCtrlSequences[key]; ok {
			return matchesLegacySequence(data, sequences)
		}
	}
	return false
}

func MatchesKey(data string, keyId string) bool {
	parsed := parseKeyId(keyId)
	if parsed == nil {
		return false
	}

	key := parsed.key
	ctrl := parsed.ctrl
	shift := parsed.shift
	alt := parsed.alt

	modifier := 0
	if shift {
		modifier |= ModifierShift
	}
	if alt {
		modifier |= ModifierAlt
	}
	if ctrl {
		modifier |= ModifierCtrl
	}

	switch key {
	case "escape", "esc":
		if modifier != 0 {
			return false
		}
		return data == "\x1b" || matchesKittySequence(data, CodepointEscape, 0)

	case "space":
		if !kittyProtocolActive {
			if ctrl && !alt && !shift && data == "\x00" {
				return true
			}
			if alt && !ctrl && !shift && data == "\x1b " {
				return true
			}
		}
		if modifier == 0 {
			return data == " " || matchesKittySequence(data, CodepointSpace, 0)
		}
		return matchesKittySequence(data, CodepointSpace, modifier)

	case "tab":
		if shift && !ctrl && !alt {
			return data == "\x1b[Z" || matchesKittySequence(data, CodepointTab, ModifierShift)
		}
		if modifier == 0 {
			return data == "\t" || matchesKittySequence(data, CodepointTab, 0)
		}
		return matchesKittySequence(data, CodepointTab, modifier)

	case "enter", "return":
		if shift && !ctrl && !alt {
			if matchesKittySequence(data, CodepointEnter, ModifierShift) ||
				matchesKittySequence(data, CodepointKPEnter, ModifierShift) {
				return true
			}
			if matchesModifyOtherKeys(data, CodepointEnter, ModifierShift) {
				return true
			}
			if kittyProtocolActive {
				return data == "\x1b\r" || data == "\n"
			}
			return false
		}
		if alt && !ctrl && !shift {
			if matchesKittySequence(data, CodepointEnter, ModifierAlt) ||
				matchesKittySequence(data, CodepointKPEnter, ModifierAlt) {
				return true
			}
			if matchesModifyOtherKeys(data, CodepointEnter, ModifierAlt) {
				return true
			}
			if !kittyProtocolActive {
				return data == "\x1b\r"
			}
			return false
		}
		if modifier == 0 {
			return data == "\r" ||
				(!kittyProtocolActive && data == "\n") ||
				data == "\x1bOM" ||
				matchesKittySequence(data, CodepointEnter, 0) ||
				matchesKittySequence(data, CodepointKPEnter, 0)
		}
		return matchesKittySequence(data, CodepointEnter, modifier) ||
			matchesKittySequence(data, CodepointKPEnter, modifier)

	case "backspace":
		if alt && !ctrl && !shift {
			if data == "\x1b\x7f" || data == "\x1b\b" {
				return true
			}
			return matchesKittySequence(data, CodepointBackspace, ModifierAlt)
		}
		if modifier == 0 {
			return data == "\x7f" || data == "\x08" || matchesKittySequence(data, CodepointBackspace, 0)
		}
		return matchesKittySequence(data, CodepointBackspace, modifier)

	case "insert":
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["insert"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, FunctionalCodepointInsert, 0)
		}
		if matchesLegacyModifierSequence(data, "insert", modifier) {
			return true
		}
		return matchesKittySequence(data, FunctionalCodepointInsert, modifier)

	case "delete":
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["delete"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, FunctionalCodepointDelete, 0)
		}
		if matchesLegacyModifierSequence(data, "delete", modifier) {
			return true
		}
		return matchesKittySequence(data, FunctionalCodepointDelete, modifier)

	case "clear":
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["clear"]; ok {
				return matchesLegacySequence(data, sequences)
			}
		}
		return matchesLegacyModifierSequence(data, "clear", modifier)

	case "home":
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["home"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, FunctionalCodepointHome, 0)
		}
		if matchesLegacyModifierSequence(data, "home", modifier) {
			return true
		}
		return matchesKittySequence(data, FunctionalCodepointHome, modifier)

	case "end":
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["end"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, FunctionalCodepointEnd, 0)
		}
		if matchesLegacyModifierSequence(data, "end", modifier) {
			return true
		}
		return matchesKittySequence(data, FunctionalCodepointEnd, modifier)

	case "pageup", "pagedown":
		lookupKey := key
		if key == "pageup" {
			lookupKey = "pageup"
		} else {
			lookupKey = "pagedown"
		}
		var expectedCodepoint int
		if key == "pageup" {
			expectedCodepoint = FunctionalCodepointPageUp
		} else {
			expectedCodepoint = FunctionalCodepointPageDown
		}
		if modifier == 0 {
			if sequences, ok := legacyKeySequences[lookupKey]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, expectedCodepoint, 0)
		}
		if matchesLegacyModifierSequence(data, lookupKey, modifier) {
			return true
		}
		return matchesKittySequence(data, expectedCodepoint, modifier)

	case "up":
		if alt && !ctrl && !shift {
			return data == "\x1bp" || matchesKittySequence(data, ArrowCodepointUp, ModifierAlt)
		}
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["up"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, ArrowCodepointUp, 0)
		}
		if matchesLegacyModifierSequence(data, "up", modifier) {
			return true
		}
		return matchesKittySequence(data, ArrowCodepointUp, modifier)

	case "down":
		if alt && !ctrl && !shift {
			return data == "\x1bn" || matchesKittySequence(data, ArrowCodepointDown, ModifierAlt)
		}
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["down"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, ArrowCodepointDown, 0)
		}
		if matchesLegacyModifierSequence(data, "down", modifier) {
			return true
		}
		return matchesKittySequence(data, ArrowCodepointDown, modifier)

	case "left":
		if alt && !ctrl && !shift {
			if data == "\x1b[1;3D" ||
				(!kittyProtocolActive && data == "\x1bB") ||
				data == "\x1bb" ||
				matchesKittySequence(data, ArrowCodepointLeft, ModifierAlt) {
				return true
			}
		}
		if ctrl && !alt && !shift {
			if data == "\x1b[1;5D" ||
				matchesLegacyModifierSequence(data, "left", ModifierCtrl) ||
				matchesKittySequence(data, ArrowCodepointLeft, ModifierCtrl) {
				return true
			}
		}
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["left"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, ArrowCodepointLeft, 0)
		}
		if matchesLegacyModifierSequence(data, "left", modifier) {
			return true
		}
		return matchesKittySequence(data, ArrowCodepointLeft, modifier)

	case "right":
		if alt && !ctrl && !shift {
			if data == "\x1b[1;3C" ||
				(!kittyProtocolActive && data == "\x1bF") ||
				data == "\x1bf" ||
				matchesKittySequence(data, ArrowCodepointRight, ModifierAlt) {
				return true
			}
		}
		if ctrl && !alt && !shift {
			if data == "\x1b[1;5C" ||
				matchesLegacyModifierSequence(data, "right", ModifierCtrl) ||
				matchesKittySequence(data, ArrowCodepointRight, ModifierCtrl) {
				return true
			}
		}
		if modifier == 0 {
			if sequences, ok := legacyKeySequences["right"]; ok {
				if matchesLegacySequence(data, sequences) {
					return true
				}
			}
			return matchesKittySequence(data, ArrowCodepointRight, 0)
		}
		if matchesLegacyModifierSequence(data, "right", modifier) {
			return true
		}
		return matchesKittySequence(data, ArrowCodepointRight, modifier)

	case "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12":
		if modifier != 0 {
			return false
		}
		if sequences, ok := legacyKeySequences[key]; ok {
			return matchesLegacySequence(data, sequences)
		}
	}

	if len(key) == 1 && ((key >= "a" && key <= "z") || symbolKeys[rune(key[0])]) {
		codepoint := int(key[0])
		rawCtrl := rawCtrlChar(key)

		if ctrl && alt && !shift && !kittyProtocolActive && rawCtrl != "" {
			return data == "\x1b"+rawCtrl
		}

		if alt && !ctrl && !shift && !kittyProtocolActive && key >= "a" && key <= "z" {
			if data == "\x1b"+key {
				return true
			}
		}

		if ctrl && !shift && !alt {
			if rawCtrl != "" && data == rawCtrl {
				return true
			}
			return matchesKittySequence(data, codepoint, ModifierCtrl)
		}

		if ctrl && shift && !alt {
			return matchesKittySequence(data, codepoint, ModifierShift+ModifierCtrl)
		}

		if shift && !ctrl && !alt {
			if data == strings.ToUpper(key) {
				return true
			}
			return matchesKittySequence(data, codepoint, ModifierShift)
		}

		if modifier != 0 {
			return matchesKittySequence(data, codepoint, modifier)
		}

		return data == key || matchesKittySequence(data, codepoint, 0)
	}

	return false
}

func ParseKey(data string) string {
	kitty := ParseKittySequence(data)
	if kitty != nil {
		var mods []string
		effectiveMod := kitty.Modifier & ^LockMask
		if effectiveMod&ModifierShift != 0 {
			mods = append(mods, "shift")
		}
		if effectiveMod&ModifierCtrl != 0 {
			mods = append(mods, "ctrl")
		}
		if effectiveMod&ModifierAlt != 0 {
			mods = append(mods, "alt")
		}

		isLatinLetter := kitty.Codepoint >= 97 && kitty.Codepoint <= 122
		isKnownSymbol := symbolKeys[rune(kitty.Codepoint)]
		effectiveCodepoint := kitty.Codepoint
		if !isLatinLetter && !isKnownSymbol && kitty.BaseLayoutKey != nil {
			effectiveCodepoint = *kitty.BaseLayoutKey
		}

		var keyName string
		switch effectiveCodepoint {
		case CodepointEscape:
			keyName = "escape"
		case CodepointTab:
			keyName = "tab"
		case CodepointEnter, CodepointKPEnter:
			keyName = "enter"
		case CodepointSpace:
			keyName = "space"
		case CodepointBackspace:
			keyName = "backspace"
		case FunctionalCodepointDelete:
			keyName = "delete"
		case FunctionalCodepointInsert:
			keyName = "insert"
		case FunctionalCodepointHome:
			keyName = "home"
		case FunctionalCodepointEnd:
			keyName = "end"
		case FunctionalCodepointPageUp:
			keyName = "pageup"
		case FunctionalCodepointPageDown:
			keyName = "pagedown"
		case ArrowCodepointUp:
			keyName = "up"
		case ArrowCodepointDown:
			keyName = "down"
		case ArrowCodepointLeft:
			keyName = "left"
		case ArrowCodepointRight:
			keyName = "right"
		default:
			if effectiveCodepoint >= 97 && effectiveCodepoint <= 122 {
				keyName = string(rune(effectiveCodepoint))
			} else if symbolKeys[rune(effectiveCodepoint)] {
				keyName = string(rune(effectiveCodepoint))
			}
		}

		if keyName != "" {
			if len(mods) > 0 {
				return strings.Join(mods, "+") + "+" + keyName
			}
			return keyName
		}
	}

	if kittyProtocolActive {
		if data == "\x1b\r" || data == "\n" {
			return "shift+enter"
		}
	}

	if keyId, ok := legacySequenceKeyIDs[data]; ok {
		return keyId
	}

	if data == "\x1b" {
		return "escape"
	}
	if data == "\x1c" {
		return "ctrl+\\"
	}
	if data == "\x1d" {
		return "ctrl+]"
	}
	if data == "\x1f" {
		return "ctrl+-"
	}
	if data == "\x1b\x1b" {
		return "ctrl+alt+["
	}
	if data == "\x1b\x1c" {
		return "ctrl+alt+\\"
	}
	if data == "\x1b\x1d" {
		return "ctrl+alt+]"
	}
	if data == "\x1b\x1f" {
		return "ctrl+alt+-"
	}
	if data == "\t" {
		return "tab"
	}
	if data == "\r" || (!kittyProtocolActive && data == "\n") || data == "\x1bOM" {
		return "enter"
	}
	if data == "\x00" {
		return "ctrl+space"
	}
	if data == " " {
		return "space"
	}
	if data == "\x7f" || data == "\x08" {
		return "backspace"
	}
	if data == "\x1b[Z" {
		return "shift+tab"
	}
	if !kittyProtocolActive && data == "\x1b\r" {
		return "alt+enter"
	}
	if !kittyProtocolActive && data == "\x1b " {
		return "alt+space"
	}
	if data == "\x1b\x7f" || data == "\x1b\b" {
		return "alt+backspace"
	}
	if !kittyProtocolActive && data == "\x1bB" {
		return "alt+left"
	}
	if !kittyProtocolActive && data == "\x1bF" {
		return "alt+right"
	}
	if !kittyProtocolActive && len(data) == 2 && data[0] == '\x1b' {
		code := int(data[1])
		if code >= 1 && code <= 26 {
			return "ctrl+alt+" + string(rune(code+96))
		}
		if code >= 97 && code <= 122 {
			return "alt+" + string(rune(code))
		}
	}
	if data == "\x1b[A" {
		return "up"
	}
	if data == "\x1b[B" {
		return "down"
	}
	if data == "\x1b[C" {
		return "right"
	}
	if data == "\x1b[D" {
		return "left"
	}
	if data == "\x1b[H" || data == "\x1bOH" {
		return "home"
	}
	if data == "\x1b[F" || data == "\x1bOF" {
		return "end"
	}
	if data == "\x1b[3~" {
		return "delete"
	}
	if data == "\x1b[5~" {
		return "pageup"
	}
	if data == "\x1b[6~" {
		return "pagedown"
	}

	if len(data) == 1 {
		code := int(data[0])
		if code >= 1 && code <= 26 {
			return "ctrl+" + string(rune(code+96))
		}
		if code >= 32 && code <= 126 {
			return data
		}
	}

	return ""
}
