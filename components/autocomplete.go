package components

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// AutocompleteItem represents a single autocomplete suggestion
type AutocompleteItem struct {
	Value       string
	Label       string
	Description string
}

// SlashCommand represents a command that can be autocompleted
type SlashCommand struct {
	Name        string
	Description string
	// Function to get argument completions for this command
	// Returns nil if no argument completion is available
	GetArgumentCompletions func(argumentPrefix string) []AutocompleteItem
}

// AutocompleteProvider interface for getting autocomplete suggestions
type AutocompleteProvider interface {
	// GetSuggestions returns autocomplete suggestions for current text/cursor position
	// Returns nil if no suggestions available
	GetSuggestions(lines []string, cursorLine, cursorCol int) *AutocompleteSuggestions

	// ApplyCompletion applies the selected item and returns new text and cursor position
	ApplyCompletion(lines []string, cursorLine, cursorCol int, item AutocompleteItem, prefix string) *AutocompleteResult
}

// AutocompleteSuggestions contains the suggestions and the prefix being matched
type AutocompleteSuggestions struct {
	Items  []AutocompleteItem
	Prefix string
}

// AutocompleteResult contains the result of applying a completion
type AutocompleteResult struct {
	Lines      []string
	CursorLine int
	CursorCol  int
}

// CombinedAutocompleteProvider handles both slash commands and file paths
type CombinedAutocompleteProvider struct {
	commands []interface{} // Can be SlashCommand or AutocompleteItem
	basePath string
	fdPath   string
}

// NewCombinedAutocompleteProvider creates a new combined autocomplete provider
func NewCombinedAutocompleteProvider(commands []interface{}, basePath, fdPath string) *CombinedAutocompleteProvider {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &CombinedAutocompleteProvider{
		commands: commands,
		basePath: basePath,
		fdPath:   fdPath,
	}
}

// GetSuggestions implements AutocompleteProvider
func (p *CombinedAutocompleteProvider) GetSuggestions(lines []string, cursorLine, cursorCol int) *AutocompleteSuggestions {
	if cursorLine >= len(lines) {
		return nil
	}

	currentLine := lines[cursorLine]
	if cursorCol > len(currentLine) {
		cursorCol = len(currentLine)
	}

	textBeforeCursor := currentLine[:cursorCol]

	// Check for @ file reference (fuzzy search) - must be after a space or at start
	if atMatch := extractAtMatch(textBeforeCursor); atMatch != "" {
		prefix := atMatch   // The @... part
		query := prefix[1:] // Remove the @
		suggestions := p.getFuzzyFileSuggestions(query)
		if len(suggestions) == 0 {
			return nil
		}

		return &AutocompleteSuggestions{
			Items:  suggestions,
			Prefix: prefix,
		}
	}

	// Check for slash commands
	if strings.HasPrefix(textBeforeCursor, "/") {
		spaceIndex := strings.Index(textBeforeCursor, " ")

		if spaceIndex == -1 {
			// No space yet - complete command names with fuzzy matching
			prefix := textBeforeCursor[1:] // Remove the "/"
			commandItems := p.getCommandItems()

			filtered := FuzzyFilter(commandItems, prefix, func(item commandItem) string {
				return item.name
			})

			if len(filtered) == 0 {
				return nil
			}

			items := make([]AutocompleteItem, len(filtered))
			for i, cmd := range filtered {
				items[i] = AutocompleteItem{
					Value:       cmd.name,
					Label:       cmd.label,
					Description: cmd.description,
				}
			}

			return &AutocompleteSuggestions{
				Items:  items,
				Prefix: textBeforeCursor,
			}
		} else {
			// Space found - complete command arguments
			commandName := textBeforeCursor[1:spaceIndex]   // Command without "/"
			argumentText := textBeforeCursor[spaceIndex+1:] // Text after space

			command := p.findCommand(commandName)
			if command == nil {
				return nil
			}

			slashCmd, ok := command.(SlashCommand)
			if !ok || slashCmd.GetArgumentCompletions == nil {
				return nil
			}

			argumentSuggestions := slashCmd.GetArgumentCompletions(argumentText)
			if len(argumentSuggestions) == 0 {
				return nil
			}

			return &AutocompleteSuggestions{
				Items:  argumentSuggestions,
				Prefix: argumentText,
			}
		}
	}

	// Check for file paths
	pathMatch := p.extractPathPrefix(textBeforeCursor, false)
	if pathMatch != "" {
		suggestions := p.getFileSuggestions(pathMatch)
		if len(suggestions) == 0 {
			return nil
		}

		return &AutocompleteSuggestions{
			Items:  suggestions,
			Prefix: pathMatch,
		}
	}

	return nil
}

// ApplyCompletion implements AutocompleteProvider
func (p *CombinedAutocompleteProvider) ApplyCompletion(lines []string, cursorLine, cursorCol int, item AutocompleteItem, prefix string) *AutocompleteResult {
	if cursorLine >= len(lines) {
		return nil
	}

	currentLine := lines[cursorLine]
	if cursorCol > len(currentLine) {
		cursorCol = len(currentLine)
	}

	beforePrefix := currentLine[:cursorCol-len(prefix)]
	afterCursor := currentLine[cursorCol:]

	// Check if we're completing a slash command
	isSlashCommand := strings.HasPrefix(prefix, "/") && strings.TrimSpace(beforePrefix) == "" && !strings.Contains(prefix[1:], "/")
	if isSlashCommand {
		newLine := beforePrefix + "/" + item.Value + " " + afterCursor
		newLines := make([]string, len(lines))
		copy(newLines, lines)
		newLines[cursorLine] = newLine

		return &AutocompleteResult{
			Lines:      newLines,
			CursorLine: cursorLine,
			CursorCol:  len(beforePrefix) + len(item.Value) + 2, // +2 for "/" and space
		}
	}

	// Check if we're completing a file attachment (prefix starts with "@")
	if strings.HasPrefix(prefix, "@") {
		isDirectory := strings.HasSuffix(item.Value, "/")
		suffix := ""
		if !isDirectory {
			suffix = " "
		}
		newLine := beforePrefix + item.Value + suffix + afterCursor
		newLines := make([]string, len(lines))
		copy(newLines, lines)
		newLines[cursorLine] = newLine

		return &AutocompleteResult{
			Lines:      newLines,
			CursorLine: cursorLine,
			CursorCol:  len(beforePrefix) + len(item.Value) + len(suffix),
		}
	}

	// Check if we're in a slash command context
	textBeforeCursor := currentLine[:cursorCol]
	if strings.Contains(textBeforeCursor, "/") && strings.Contains(textBeforeCursor, " ") {
		newLine := beforePrefix + item.Value + afterCursor
		newLines := make([]string, len(lines))
		copy(newLines, lines)
		newLines[cursorLine] = newLine

		return &AutocompleteResult{
			Lines:      newLines,
			CursorLine: cursorLine,
			CursorCol:  len(beforePrefix) + len(item.Value),
		}
	}

	// For file paths, complete the path
	newLine := beforePrefix + item.Value + afterCursor
	newLines := make([]string, len(lines))
	copy(newLines, lines)
	newLines[cursorLine] = newLine

	return &AutocompleteResult{
		Lines:      newLines,
		CursorLine: cursorLine,
		CursorCol:  len(beforePrefix) + len(item.Value),
	}
}

// GetForceFileSuggestions forces file completion (called on Tab key)
func (p *CombinedAutocompleteProvider) GetForceFileSuggestions(lines []string, cursorLine, cursorCol int) *AutocompleteSuggestions {
	if cursorLine >= len(lines) {
		return nil
	}

	currentLine := lines[cursorLine]
	if cursorCol > len(currentLine) {
		cursorCol = len(currentLine)
	}

	textBeforeCursor := currentLine[:cursorCol]

	// Don't trigger if we're typing a slash command at the start of the line
	trimmed := strings.TrimSpace(textBeforeCursor)
	if strings.HasPrefix(trimmed, "/") && !strings.Contains(trimmed, " ") {
		return nil
	}

	// Force extract path prefix
	pathMatch := p.extractPathPrefix(textBeforeCursor, true)
	if pathMatch != "" {
		suggestions := p.getFileSuggestions(pathMatch)
		if len(suggestions) == 0 {
			return nil
		}

		return &AutocompleteSuggestions{
			Items:  suggestions,
			Prefix: pathMatch,
		}
	}

	return nil
}

// ShouldTriggerFileCompletion checks if we should trigger file completion
func (p *CombinedAutocompleteProvider) ShouldTriggerFileCompletion(lines []string, cursorLine, cursorCol int) bool {
	if cursorLine >= len(lines) {
		return false
	}

	currentLine := lines[cursorLine]
	if cursorCol > len(currentLine) {
		cursorCol = len(currentLine)
	}

	textBeforeCursor := currentLine[:cursorCol]

	// Don't trigger if we're typing a slash command at the start of the line
	trimmed := strings.TrimSpace(textBeforeCursor)
	if strings.HasPrefix(trimmed, "/") && !strings.Contains(trimmed, " ") {
		return false
	}

	return true
}

// Helper types and functions

type commandItem struct {
	name        string
	label       string
	description string
}

func (p *CombinedAutocompleteProvider) getCommandItems() []commandItem {
	items := make([]commandItem, 0, len(p.commands))
	for _, cmd := range p.commands {
		switch c := cmd.(type) {
		case SlashCommand:
			items = append(items, commandItem{
				name:        c.Name,
				label:       c.Name,
				description: c.Description,
			})
		case AutocompleteItem:
			items = append(items, commandItem{
				name:        c.Value,
				label:       c.Label,
				description: c.Description,
			})
		}
	}
	return items
}

func (p *CombinedAutocompleteProvider) findCommand(name string) interface{} {
	for _, cmd := range p.commands {
		switch c := cmd.(type) {
		case SlashCommand:
			if c.Name == name {
				return c
			}
		case AutocompleteItem:
			if c.Value == name {
				return c
			}
		}
	}
	return nil
}

// extractAtMatch extracts @ file reference pattern
func extractAtMatch(text string) string {
	// Match @... pattern after space or at start
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == '@' {
			if i == 0 || text[i-1] == ' ' {
				return text[i:]
			}
		}
		if text[i] == ' ' {
			break
		}
	}
	return ""
}

// extractPathPrefix extracts a path-like prefix from the text
func (p *CombinedAutocompleteProvider) extractPathPrefix(text string, forceExtract bool) string {
	// Check for @ file attachment syntax first
	if atMatch := extractAtMatch(text); atMatch != "" {
		return atMatch
	}

	// Find the last delimiter
	lastDelimiterIndex := -1
	delimiters := []byte{' ', '\t', '"', '\'', '='}
	for i := len(text) - 1; i >= 0; i-- {
		for _, delim := range delimiters {
			if text[i] == delim {
				lastDelimiterIndex = i
				break
			}
		}
		if lastDelimiterIndex != -1 {
			break
		}
	}

	var pathPrefix string
	if lastDelimiterIndex == -1 {
		pathPrefix = text
	} else {
		pathPrefix = text[lastDelimiterIndex+1:]
	}

	// For forced extraction (Tab key), always return something
	if forceExtract {
		return pathPrefix
	}

	// For natural triggers, return if it looks like a path
	if strings.Contains(pathPrefix, "/") || strings.HasPrefix(pathPrefix, ".") || strings.HasPrefix(pathPrefix, "~/") {
		return pathPrefix
	}

	// Return empty string only after a space
	if pathPrefix == "" && strings.HasSuffix(text, " ") {
		return pathPrefix
	}

	return ""
}

// expandHomePath expands ~/ to actual home path
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		expanded := filepath.Join(homeDir, path[2:])
		// Preserve trailing slash
		if strings.HasSuffix(path, "/") && !strings.HasSuffix(expanded, "/") {
			return expanded + "/"
		}
		return expanded
	} else if path == "~" {
		homeDir, _ := os.UserHomeDir()
		return homeDir
	}
	return path
}

// getFileSuggestions gets file/directory suggestions for a given path prefix
func (p *CombinedAutocompleteProvider) getFileSuggestions(prefix string) []AutocompleteItem {
	var searchDir, searchPrefix string
	expandedPrefix := prefix
	isAtPrefix := false

	// Handle @ file attachment prefix
	if strings.HasPrefix(prefix, "@") {
		isAtPrefix = true
		expandedPrefix = prefix[1:] // Remove the @
	}

	// Handle home directory expansion
	if strings.HasPrefix(expandedPrefix, "~") {
		expandedPrefix = expandHomePath(expandedPrefix)
	}

	// Determine search directory and prefix
	if expandedPrefix == "" || expandedPrefix == "./" || expandedPrefix == "../" ||
		expandedPrefix == "~" || expandedPrefix == "~/" || expandedPrefix == "/" || prefix == "@" {
		// Complete from specified position
		if strings.HasPrefix(prefix, "~") || expandedPrefix == "/" {
			searchDir = expandedPrefix
		} else {
			searchDir = filepath.Join(p.basePath, expandedPrefix)
		}
		searchPrefix = ""
	} else if strings.HasSuffix(expandedPrefix, "/") {
		// If prefix ends with /, show contents of that directory
		if strings.HasPrefix(prefix, "~") || strings.HasPrefix(expandedPrefix, "/") {
			searchDir = expandedPrefix
		} else {
			searchDir = filepath.Join(p.basePath, expandedPrefix)
		}
		searchPrefix = ""
	} else {
		// Split into directory and file prefix
		dir := filepath.Dir(expandedPrefix)
		file := filepath.Base(expandedPrefix)
		if strings.HasPrefix(prefix, "~") || strings.HasPrefix(expandedPrefix, "/") {
			searchDir = dir
		} else {
			searchDir = filepath.Join(p.basePath, dir)
		}
		searchPrefix = file
	}

	// Read directory entries
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil
	}

	suggestions := make([]AutocompleteItem, 0)
	lowerSearchPrefix := strings.ToLower(searchPrefix)

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(strings.ToLower(name), lowerSearchPrefix) {
			continue
		}

		// Check if entry is a directory
		isDirectory := entry.IsDir()
		if !isDirectory && entry.Type()&os.ModeSymlink != 0 {
			// Check if symlink points to directory
			fullPath := filepath.Join(searchDir, name)
			if info, err := os.Stat(fullPath); err == nil {
				isDirectory = info.IsDir()
			}
		}

		// Build relative path
		var relativePath string
		if isAtPrefix {
			pathWithoutAt := expandedPrefix
			if strings.HasSuffix(pathWithoutAt, "/") {
				relativePath = "@" + pathWithoutAt + name
			} else if strings.Contains(pathWithoutAt, "/") {
				if strings.HasPrefix(pathWithoutAt, "~/") {
					homeRelativeDir := pathWithoutAt[2:]
					dir := filepath.Dir(homeRelativeDir)
					if dir == "." {
						relativePath = "@~/" + name
					} else {
						relativePath = "@~/" + filepath.Join(dir, name)
					}
				} else {
					relativePath = "@" + filepath.Join(filepath.Dir(pathWithoutAt), name)
				}
			} else {
				if strings.HasPrefix(pathWithoutAt, "~") {
					relativePath = "@~/" + name
				} else {
					relativePath = "@" + name
				}
			}
		} else if strings.HasSuffix(prefix, "/") {
			relativePath = prefix + name
		} else if strings.Contains(prefix, "/") {
			if strings.HasPrefix(prefix, "~/") {
				homeRelativeDir := prefix[2:]
				dir := filepath.Dir(homeRelativeDir)
				if dir == "." {
					relativePath = "~/" + name
				} else {
					relativePath = "~/" + filepath.Join(dir, name)
				}
			} else if strings.HasPrefix(prefix, "/") {
				dir := filepath.Dir(prefix)
				if dir == "/" {
					relativePath = "/" + name
				} else {
					relativePath = dir + "/" + name
				}
			} else {
				relativePath = filepath.Join(filepath.Dir(prefix), name)
			}
		} else {
			if strings.HasPrefix(prefix, "~") {
				relativePath = "~/" + name
			} else {
				relativePath = name
			}
		}

		// Add trailing slash for directories
		if isDirectory {
			relativePath += "/"
		}

		suggestions = append(suggestions, AutocompleteItem{
			Value: relativePath,
			Label: name + map[bool]string{true: "/", false: ""}[isDirectory],
		})
	}

	// Sort directories first, then alphabetically
	sort.Slice(suggestions, func(i, j int) bool {
		iIsDir := strings.HasSuffix(suggestions[i].Value, "/")
		jIsDir := strings.HasSuffix(suggestions[j].Value, "/")
		if iIsDir && !jIsDir {
			return true
		}
		if !iIsDir && jIsDir {
			return false
		}
		return suggestions[i].Label < suggestions[j].Label
	})

	return suggestions
}

// scoreEntry scores an entry against the query (higher = better match)
func scoreEntry(filePath, query string, isDirectory bool) int {
	fileName := filepath.Base(filePath)
	lowerFileName := strings.ToLower(fileName)
	lowerQuery := strings.ToLower(query)

	score := 0

	// Exact filename match (highest)
	if lowerFileName == lowerQuery {
		score = 100
	} else if strings.HasPrefix(lowerFileName, lowerQuery) {
		// Filename starts with query
		score = 80
	} else if strings.Contains(lowerFileName, lowerQuery) {
		// Substring match in filename
		score = 50
	} else if strings.Contains(strings.ToLower(filePath), lowerQuery) {
		// Substring match in full path
		score = 30
	}

	// Directories get a bonus to appear first
	if isDirectory && score > 0 {
		score += 10
	}

	return score
}

// walkDirectoryWithFd uses fd to walk directory tree
func walkDirectoryWithFd(baseDir, fdPath, query string, maxResults int) []struct {
	path        string
	isDirectory bool
} {
	args := []string{
		"--base-directory", baseDir,
		"--max-results", string(rune(maxResults)),
		"--type", "f",
		"--type", "d",
		"--full-path",
	}

	if query != "" {
		args = append(args, query)
	}

	cmd := exec.Command(fdPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	results := make([]struct {
		path        string
		isDirectory bool
	}, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}
		// fd outputs directories with trailing /
		isDirectory := strings.HasSuffix(line, "/")
		results = append(results, struct {
			path        string
			isDirectory bool
		}{
			path:        line,
			isDirectory: isDirectory,
		})
	}

	return results
}

// getFuzzyFileSuggestions performs fuzzy file search using fd
func (p *CombinedAutocompleteProvider) getFuzzyFileSuggestions(query string) []AutocompleteItem {
	if p.fdPath == "" {
		return nil
	}

	entries := walkDirectoryWithFd(p.basePath, p.fdPath, query, 100)

	// Score entries
	type scoredEntry struct {
		path        string
		isDirectory bool
		score       int
	}

	scoredEntries := make([]scoredEntry, 0)
	for _, entry := range entries {
		score := 1
		if query != "" {
			score = scoreEntry(entry.path, query, entry.isDirectory)
		}
		if score > 0 {
			scoredEntries = append(scoredEntries, scoredEntry{
				path:        entry.path,
				isDirectory: entry.isDirectory,
				score:       score,
			})
		}
	}

	// Sort by score (descending) and take top 20
	sort.Slice(scoredEntries, func(i, j int) bool {
		return scoredEntries[i].score > scoredEntries[j].score
	})

	if len(scoredEntries) > 20 {
		scoredEntries = scoredEntries[:20]
	}

	// Build suggestions
	suggestions := make([]AutocompleteItem, 0, len(scoredEntries))
	for _, entry := range scoredEntries {
		pathWithoutSlash := entry.path
		if entry.isDirectory {
			pathWithoutSlash = entry.path[:len(entry.path)-1]
		}
		entryName := filepath.Base(pathWithoutSlash)

		suggestions = append(suggestions, AutocompleteItem{
			Value:       "@" + entry.path,
			Label:       entryName + map[bool]string{true: "/", false: ""}[entry.isDirectory],
			Description: pathWithoutSlash,
		})
	}

	return suggestions
}
