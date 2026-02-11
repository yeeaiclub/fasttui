package components

import (
	"fmt"
	"strings"

	"github.com/yeeaiclub/fasttui"
	"github.com/yeeaiclub/fasttui/keys"
)

type SelectItem struct {
	Label       string
	Value       string
	Description string
}

type SelectListTheme struct {
	SelectedPrefix string
	NormalPrefix   string
	NoMatch        func(string) string
	ScrollInfo     func(string) string
	Description    func(string) string
}

type SelectList struct {
	items         []SelectItem
	filteredItems []SelectItem
	selectedIndex int
	maxVisible    int
	theme         SelectListTheme

	onSelect          func(item SelectItem)
	onCancel          func()
	onSelectionChange func(item SelectItem)
}

func NewSelectList(items []SelectItem, maxVisible int, theme SelectListTheme) *SelectList {
	return &SelectList{
		items:         items,
		filteredItems: items,
		selectedIndex: 0,
		maxVisible:    maxVisible,
		theme:         theme,
	}
}

func (s *SelectList) Render(width int) []string {
	var lines []string

	if len(s.filteredItems) == 0 {
		if s.theme.NoMatch != nil {
			lines = append(lines, s.theme.NoMatch("  No matching commands"))
		} else {
			lines = append(lines, "  No matching commands")
		}
		return lines
	}

	startIndex := max(0, min(s.selectedIndex-s.maxVisible/2, len(s.filteredItems)-s.maxVisible))
	endIndex := min(startIndex+s.maxVisible, len(s.filteredItems))

	for i := startIndex; i < endIndex; i++ {
		item := s.filteredItems[i]
		isSelected := i == s.selectedIndex
		line := ""
		if isSelected {
			line = s.handleSelect(item, width)
		} else {
			line = s.handleNoSelect(item, width)
		}
		lines = append(lines, line)
	}

	if startIndex > 0 || endIndex < len(s.filteredItems) {
		rateText := fmt.Sprintf(" (%d/%d)", startIndex, len(s.filteredItems))
		lines = append(lines, fasttui.TruncateToWidth(rateText, width-2, "", false))
	}
	return lines
}

func (s *SelectList) handleSelect(item SelectItem, width int) string {
	prefix := "→ "
	prefixLen := len(prefix)

	display := item.Value
	if item.Label != "" {
		display = item.Label
	}

	descLine := ""
	if item.Description != "" {
		descLine = normalizeToSingleLine(item.Description)
	}

	if width < 40 || descLine == "" {
		maxWidth := width - prefixLen - 2
		return fmt.Sprintf("%s %s", prefix, fasttui.TruncateToWidth(display, maxWidth, "", false))
	}

	maxValueWidth := min(30, width-prefixLen-4)
	truncatedValue := fasttui.TruncateToWidth(display, maxValueWidth, "", false)
	spacing := strings.Repeat(" ", max(1, 32-len(truncatedValue)))
	start := prefixLen + len(truncatedValue) + len(spacing)
	remainingWidth := width - start - 2

	if remainingWidth > 10 && descLine != "" {
		desc := fasttui.TruncateToWidth(descLine, remainingWidth, "", false)
		descText := s.theme.Description(spacing + desc)
		return prefix + truncatedValue + descText
	}
	maxWidth := width - len(prefix) - 2
	return prefix + fasttui.TruncateToWidth(display, maxWidth, "", false)
}

func (s *SelectList) handleNoSelect(item SelectItem, width int) string {
	prefix := "  "
	prefixLen := len(prefix)
	display := item.Value
	if item.Label != "" {
		display = item.Label
	}

	descLine := ""
	if item.Description != "" {
		descLine = normalizeToSingleLine(item.Description)
	}

	if descLine == "" || width < 40 {
		maxWidth := width - len(prefix) - 2
		line := prefix + fasttui.TruncateToWidth(display, maxWidth, "", false)
		return line
	}

	maxValueWidth := min(30, width-prefixLen-4)
	truncatedValue := fasttui.TruncateToWidth(display, maxValueWidth, "", false)
	spacing := strings.Repeat(" ", max(1, 32-len(truncatedValue)))

	start := prefixLen + len(truncatedValue) + len(spacing)
	remainingWidth := width - start - 2

	if remainingWidth > 10 && descLine != "" {
		desc := fasttui.TruncateToWidth(descLine, remainingWidth, "", false)
		descText := s.theme.Description(spacing + desc)
		return prefix + truncatedValue + descText
	}

	maxWidth := width - len(prefix) - 2
	return prefix + fasttui.TruncateToWidth(display, maxWidth, "", false)
}

func (s *SelectList) HandleInput(keyData string) {
	if len(s.filteredItems) == 0 {
		kb := keys.GetEditorKeybindings()
		if kb.Matches(keyData, keys.EditorActionSelectCancel) {
			if s.onCancel != nil {
				s.onCancel()
			}
		}
		return
	}

	kb := keys.GetEditorKeybindings()
	if kb.Matches(keyData, keys.EditorActionSelectUp) {
		if s.selectedIndex == 0 {
			s.selectedIndex = len(s.filteredItems) - 1
		} else {
			s.selectedIndex--
		}
		s.notifySelectionChange()
	}

	if kb.Matches(keyData, keys.EditorActionSelectDown) {
		if s.selectedIndex == len(s.filteredItems)-1 {
			s.selectedIndex = 0
		} else {
			s.selectedIndex++
		}
		s.notifySelectionChange()
	}

	if kb.Matches(keyData, keys.EditorActionSelectConfirm) {
		item := s.getSelectItem()
		if s.onSelect != nil {
			s.onSelect(item)
		}
	}

	if kb.Matches(keyData, keys.EditorActionSelectCancel) {
		if s.onCancel != nil {
			s.onCancel()
		}
	}
}

func (s *SelectList) notifySelectionChange() {
	selectItem := s.filteredItems[s.selectedIndex]
	if s.onSelectionChange != nil {
		s.onSelectionChange(selectItem)
	}
}

func (s *SelectList) getSelectItem() SelectItem {
	item := s.filteredItems[s.selectedIndex]
	return item
}

func normalizeToSingleLine(text string) string {
	replacer := strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")
	return strings.TrimSpace(replacer.Replace(text))
}
