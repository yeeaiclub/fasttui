package style

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ThemeFile is the on-disk JSON shape for a color theme.
type ThemeFile struct {
	Schema  string         `json:"$schema,omitempty"`
	Name    string         `json:"name"`
	Vars    map[string]any `json:"vars,omitempty"`
	Colors  map[string]any `json:"colors"`
	Export  *ThemeExport   `json:"export,omitempty"`
	Symbols *ThemeSymbols  `json:"symbols,omitempty"`
}

// ThemeExport holds optional HTML export hints.
type ThemeExport struct {
	PageBg any `json:"pageBg,omitempty"`
	CardBg any `json:"cardBg,omitempty"`
	InfoBg any `json:"infoBg,omitempty"`
}

// ThemeSymbols configures embedded symbol overrides in JSON.
type ThemeSymbols struct {
	Preset    *SymbolPreset     `json:"preset,omitempty"`
	Overrides map[string]string `json:"overrides,omitempty"`
}

// ParseThemeJSON validates and decodes theme JSON bytes.
func ParseThemeJSON(data []byte) (*ThemeFile, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var raw struct {
		Schema  string          `json:"$schema,omitempty"`
		Name    string          `json:"name"`
		Vars    json.RawMessage `json:"vars,omitempty"`
		Colors  json.RawMessage `json:"colors"`
		Export  *ThemeExport    `json:"export,omitempty"`
		Symbols *ThemeSymbols   `json:"symbols,omitempty"`
	}
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse theme json: %w", err)
	}
	if raw.Name == "" {
		return nil, fmt.Errorf("theme: missing name")
	}
	if len(raw.Colors) == 0 {
		return nil, fmt.Errorf("theme %q: missing colors", raw.Name)
	}
	var colors map[string]any
	if err := json.Unmarshal(raw.Colors, &colors); err != nil {
		return nil, fmt.Errorf("theme %q: colors: %w", raw.Name, err)
	}
	var missing []string
	for _, k := range requiredThemeColorKeys {
		if _, ok := colors[k]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("theme %q: missing required color tokens: %s", raw.Name, strings.Join(missing, ", "))
	}
	tf := &ThemeFile{
		Schema:  raw.Schema,
		Name:    raw.Name,
		Colors:  colors,
		Export:  raw.Export,
		Symbols: raw.Symbols,
	}
	if len(raw.Vars) > 0 {
		var vars map[string]any
		if err := json.Unmarshal(raw.Vars, &vars); err != nil {
			return nil, fmt.Errorf("theme %q: vars: %w", raw.Name, err)
		}
		tf.Vars = vars
	}
	if tf.Symbols != nil && tf.Symbols.Preset != nil {
		if !ValidSymbolPreset(string(*tf.Symbols.Preset)) {
			return nil, fmt.Errorf("theme %q: invalid symbols.preset %q", raw.Name, *tf.Symbols.Preset)
		}
	}
	return tf, nil
}
