package config

import (
	_ "embed"
	"strings"
)

//go:embed HELP_TEXT.md
var helpTextMD string

// SearchIndexEntry is a single searchable config field document.
// Each entry is self-contained with all metadata a search engine needs.
type SearchIndexEntry struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Category      string `json:"category"`
	CategoryLabel string `json:"category_label"`
	Description   string `json:"description"`
	HelpText      string `json:"help_text"`
	Default       string `json:"default"`
	Example       string `json:"example"`
	HotReloadable bool   `json:"hot_reloadable"`
	Sensitive     bool   `json:"sensitive"`
	Required      bool   `json:"required"`
}

// BuildSearchIndex returns a searchable index of all config fields, combining
// FieldRegistry metadata with rich help text parsed from HELP_TEXT.md.
func BuildSearchIndex() []SearchIndexEntry {
	helpMap, defaultMap := parseHelpText(helpTextMD)

	entries := make([]SearchIndexEntry, 0, len(FieldRegistry))
	for _, f := range FieldRegistry {
		entries = append(entries, SearchIndexEntry{
			Key:           f.JSONKey,
			Label:         f.Label,
			Category:      string(f.Category),
			CategoryLabel: CategoryLabel(f.Category),
			Description:   f.Description,
			HelpText:      helpMap[f.JSONKey],
			Default:       defaultMap[f.JSONKey],
			Example:       f.Example,
			HotReloadable: f.HotReloadable,
			Sensitive:     f.Sensitive,
			Required:      f.Required,
		})
	}
	return entries
}

// parseHelpText extracts per-field help text and default values from the
// HELP_TEXT.md markdown format. Returns two maps keyed by field JSON key.
func parseHelpText(md string) (helpText map[string]string, defaults map[string]string) {
	helpText = make(map[string]string)
	defaults = make(map[string]string)

	lines := strings.Split(md, "\n")
	var currentKey string
	var body strings.Builder

	for _, line := range lines {
		// Field header: ### `field_key`
		if strings.HasPrefix(line, "### `") {
			// Save previous field if any.
			if currentKey != "" {
				saveField(currentKey, body.String(), helpText, defaults)
			}
			// Extract key from ### `key`
			trimmed := strings.TrimPrefix(line, "### `")
			trimmed = strings.TrimSuffix(trimmed, "`")
			trimmed = strings.TrimSpace(trimmed)
			currentKey = trimmed
			body.Reset()
			continue
		}

		// Category headers (## Server, etc.) reset the current field.
		if strings.HasPrefix(line, "## ") {
			if currentKey != "" {
				saveField(currentKey, body.String(), helpText, defaults)
				currentKey = ""
				body.Reset()
			}
			continue
		}

		// Accumulate body lines for current field.
		if currentKey != "" {
			body.WriteString(line)
			body.WriteByte('\n')
		}
	}

	// Save the last field.
	if currentKey != "" {
		saveField(currentKey, body.String(), helpText, defaults)
	}

	return helpText, defaults
}

// saveField extracts the help text and default value from a field's markdown body.
func saveField(key, body string, helpText, defaults map[string]string) {
	// Split into help text (before **Default:**) and default line.
	parts := strings.SplitN(body, "**Default:**", 2)

	help := strings.TrimSpace(parts[0])

	// Skip cross-reference entries like "See [Server > cookie_name](...)"
	if strings.HasPrefix(help, "See [") {
		return
	}

	helpText[key] = help

	if len(parts) == 2 {
		def := strings.TrimSpace(parts[1])
		// Strip inline backticks from default values.
		def = strings.ReplaceAll(def, "`", "")
		defaults[key] = def
	}
}
