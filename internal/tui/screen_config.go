package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
)

// 3/9 grid: left = categories list, right = fields (top) + detail (bottom)
var configGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Categories"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Fields"},
			{Height: 0.4, Title: "Detail"},
		}},
	},
}

// ConfigScreen implements Screen for the configuration page (CONFIGPAGE).
type ConfigScreen struct {
	GridScreen
	ConfigFieldCursor    int // field cursor within selected category
	ConfigCategory       config.FieldCategory
	ConfigCategoryFields []config.FieldMeta
	Viewport             viewport.Model
}

// NewConfigScreen creates a ConfigScreen with initial state.
func NewConfigScreen(category config.FieldCategory, categoryFields []config.FieldMeta, configFieldCursor int) *ConfigScreen {
	menuItems := ConfigCategoryMenuInit()
	cursorMax := len(menuItems) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &ConfigScreen{
		GridScreen: GridScreen{
			Grid:      configGrid,
			CursorMax: cursorMax,
		},
		ConfigFieldCursor:    configFieldCursor,
		ConfigCategory:       category,
		ConfigCategoryFields: categoryFields,
		Viewport:             viewport.Model{},
	}
}

func (s *ConfigScreen) PageIndex() PageIndex { return CONFIGPAGE }

func (s *ConfigScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	// Raw JSON viewport mode: when fields cell is focused and category is
	// "raw_json", delegate all non-navigation keys to the viewport for scrolling.
	if s.ConfigCategory == "raw_json" && s.FocusIndex == 1 {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			km := ctx.Config.KeyBindings
			key := msg.String()

			if km.Matches(key, config.ActionQuit) {
				return s, tea.Quit
			}
			if s.HandleFocusNav(key, km) {
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.FocusIndex = 0
				s.ConfigCategory = ""
				s.ConfigCategoryFields = nil
				return s, nil
			}
		}
		// All other messages (including remaining key events) delegate to viewport
		var cmd tea.Cmd
		s.Viewport, cmd = s.Viewport.Update(msg)
		return s, cmd
	}

	menuItems := ConfigCategoryMenuInit()

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		switch s.FocusIndex {
		case 0: // Categories
			if km.Matches(key, config.ActionUp) {
				if s.Cursor > 0 {
					s.Cursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				if s.Cursor < len(menuItems)-1 {
					s.Cursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				return s, HistoryPopCmd()
			}
			if km.Matches(key, config.ActionSelect) {
				categories := config.AllCategories()
				if s.Cursor < len(categories) {
					s.ConfigCategory = categories[s.Cursor]
					s.ConfigCategoryFields = config.FieldsByCategory(s.ConfigCategory)
					s.ConfigFieldCursor = 0
					s.FocusIndex = 1
					return s, nil
				}
				// Last item: "View Raw JSON"
				if s.Cursor == len(menuItems)-1 {
					content, err := configFormatJSON(ctx.Config)
					if err == nil {
						s.Viewport.SetContent(content)
					}
					s.ConfigCategory = "raw_json"
					s.ConfigCategoryFields = nil
					s.ConfigFieldCursor = 0
					s.FocusIndex = 1
					return s, nil
				}
			}

		case 1: // Fields
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.FocusIndex = 0
				s.ConfigCategory = ""
				s.ConfigCategoryFields = nil
				return s, nil
			}
			if km.Matches(key, config.ActionUp) {
				if s.ConfigFieldCursor > 0 {
					s.ConfigFieldCursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				if s.ConfigFieldCursor < len(s.ConfigCategoryFields)-1 {
					s.ConfigFieldCursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionEdit) || km.Matches(key, config.ActionSelect) {
				if len(s.ConfigCategoryFields) > 0 && s.ConfigFieldCursor < len(s.ConfigCategoryFields) {
					field := s.ConfigCategoryFields[s.ConfigFieldCursor]
					currentValue := config.ConfigFieldString(*ctx.Config, field.JSONKey)
					if field.Sensitive {
						currentValue = ""
					}
					return s, ShowConfigFieldEditDialogCmd(field, currentValue)
				}
			}

		case 2: // Detail
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.FocusIndex = 1
				return s, nil
			}
		}

		if km.Matches(key, config.ActionQuit) {
			return s, tea.Quit
		}
	}

	return s, nil
}

func (s *ConfigScreen) KeyHints(km config.KeyMap) []KeyHint {
	switch s.FocusIndex {
	case 1: // Fields
		if s.ConfigCategory == "raw_json" {
			return []KeyHint{
				{"↑↓/pgup/pgdn", "scroll"},
				{km.HintString(config.ActionNextPanel), "panel"},
				{km.HintString(config.ActionBack), "back"},
				{km.HintString(config.ActionQuit), "quit"},
			}
		}
		return []KeyHint{
			{km.HintString(config.ActionEdit), "edit"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	default:
		return []KeyHint{
			{km.HintString(config.ActionSelect), "select"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	}
}

func (s *ConfigScreen) View(ctx AppContext) string {
	catItems := ConfigCategoryMenuInit()
	fieldsContent := s.renderFields(ctx)

	// Actual inner heights for each cell (accounting for borders/title).
	catInnerH := s.Grid.CellInnerHeight(0, ctx.Height)
	fieldsInnerH := s.Grid.CellInnerHeight(1, ctx.Height)

	// Fields render as multi-line: 2-line header + 3 lines per field.
	// Convert field cursor to line offset for scrolling.
	fieldsTotalLines := strings.Count(fieldsContent, "\n") + 1
	fieldScrollLine := 0
	if len(s.ConfigCategoryFields) > 0 {
		fieldScrollLine = 2 + s.ConfigFieldCursor*3 // header(2) + 3 lines per field
	}

	cells := []CellContent{
		{Content: s.renderCategories(), TotalLines: len(catItems), ScrollOffset: ClampScroll(s.Cursor, len(catItems), catInnerH)},
		{Content: fieldsContent, TotalLines: fieldsTotalLines, ScrollOffset: ClampScroll(fieldScrollLine, fieldsTotalLines, fieldsInnerH)},
		{Content: s.renderDetail(ctx)},
	}
	return s.RenderGrid(ctx, cells)
}

// renderCategories renders the category list for the left panel.
func (s *ConfigScreen) renderCategories() string {
	categories := config.AllCategories()
	items := make([]string, 0, len(categories)+1)
	for _, cat := range categories {
		items = append(items, config.CategoryLabel(cat))
	}
	items = append(items, "View Raw JSON")

	lines := make([]string, 0, len(items))
	for i, label := range items {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		// Highlight active category
		active := ""
		if i < len(categories) && categories[i] == s.ConfigCategory {
			active = " *"
		}
		if i == len(items)-1 && s.ConfigCategory == "raw_json" {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, label, active))
	}
	return strings.Join(lines, "\n")
}

// renderFields renders the config fields for the fields cell.
func (s *ConfigScreen) renderFields(ctx AppContext) string {
	if s.ConfigCategory == "" {
		return " Select a category"
	}

	if s.ConfigCategory == "raw_json" {
		return s.Viewport.View()
	}

	if len(s.ConfigCategoryFields) == 0 {
		return " (no fields)"
	}

	title := config.CategoryLabel(s.ConfigCategory)
	labelStyle := lipgloss.NewStyle().Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	lines := []string{labelStyle.Render(title), ""}

	for i, field := range s.ConfigCategoryFields {
		value := config.ConfigFieldString(*ctx.Config, field.JSONKey)
		if field.Sensitive && value != "" {
			value = "********"
		}

		restartMark := ""
		if !field.HotReloadable {
			restartMark = lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Render(" [restart]")
		}

		if s.FocusIndex == 1 && i == s.ConfigFieldCursor {
			lines = append(lines, cursorStyle.Render("> ")+labelStyle.Render(field.Label)+restartMark)
			lines = append(lines, fmt.Sprintf("    %s", value))
		} else {
			lines = append(lines, fmt.Sprintf("  %s%s", field.Label, restartMark))
			lines = append(lines, fmt.Sprintf("    %s", value))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderDetail renders the field detail for the bottom-right cell.
func (s *ConfigScreen) renderDetail(ctx AppContext) string {
	if s.ConfigCategory == "" {
		return " Config\n\n  Select a category to\n  view its fields."
	}

	if s.ConfigCategory == "raw_json" {
		return " Raw JSON\n\n  Scroll with up/down\n  or pgup/pgdn."
	}

	if len(s.ConfigCategoryFields) == 0 || s.ConfigFieldCursor >= len(s.ConfigCategoryFields) {
		return ""
	}

	field := s.ConfigCategoryFields[s.ConfigFieldCursor]
	value := config.ConfigFieldString(*ctx.Config, field.JSONKey)
	if field.Sensitive && value != "" {
		value = "********"
	}

	lines := []string{
		fmt.Sprintf(" Field  %s", field.Label),
		fmt.Sprintf(" Key    %s", field.JSONKey),
		"",
		fmt.Sprintf(" Value  %s", value),
	}

	if field.Description != "" {
		lines = append(lines, "", fmt.Sprintf(" %s", field.Description))
	}

	lines = append(lines, "")
	if field.Sensitive {
		lines = append(lines, "   (sensitive)")
	}
	if field.HotReloadable {
		lines = append(lines, "   Hot-reloadable")
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Render("   Requires restart"))
	}

	if field.Example != "" {
		lines = append(lines, fmt.Sprintf("   Example: %s", field.Example))
	}

	lines = append(lines, "", "   Press e to edit")

	return strings.Join(lines, "\n")
}

// configFormatJSON marshals config to formatted JSON for the raw JSON view.
func configFormatJSON(c *config.Config) (string, error) {
	formatted, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}
