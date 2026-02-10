package components

type SelectItem struct {
	Label string
	Value string
}

type SelectListTheme struct {
	SelectedPrefix string
	NormalPrefix   string
	NoMatch        func(string) string
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
		if i == s.selectedIndex {
			lines = append(lines, s.theme.SelectedPrefix+item.Label)
		} else {
			lines = append(lines, s.theme.NormalPrefix+item.Label)
		}
	}

	return lines
}
