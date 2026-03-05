package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// ConfigScreen implements Screen for the configuration page (CONFIGPAGE).
// It displays config categories in a left panel and fields/raw JSON in a
// center panel. The right panel ratio is 0 (2-panel layout).
type ConfigScreen struct {
	Cursor               int // category cursor (TreePanel)
	ConfigFieldCursor    int // field cursor within selected category (ContentPanel)
	PanelFocus           FocusPanel
	ConfigCategory       config.FieldCategory
	ConfigCategoryFields []config.FieldMeta
	Viewport             viewport.Model
}

// NewConfigScreen creates a ConfigScreen with initial state from Model fields.
func NewConfigScreen(category config.FieldCategory, categoryFields []config.FieldMeta, configFieldCursor int) *ConfigScreen {
	return &ConfigScreen{
		Cursor:               0,
		ConfigFieldCursor:    configFieldCursor,
		PanelFocus:           TreePanel,
		ConfigCategory:       category,
		ConfigCategoryFields: categoryFields,
		Viewport:             viewport.Model{},
	}
}

func (s *ConfigScreen) PageIndex() PageIndex { return CONFIGPAGE }

func (s *ConfigScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	// Raw JSON viewport mode: when center panel is focused and category is
	// "raw_json", delegate all non-navigation keys to the viewport for scrolling.
	if s.ConfigCategory == "raw_json" && s.PanelFocus == ContentPanel {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			km := ctx.Config.KeyBindings
			key := msg.String()

			if km.Matches(key, config.ActionQuit) {
				return s, tea.Quit
			}
			if km.Matches(key, config.ActionNextPanel) {
				s.PanelFocus = (s.PanelFocus + 1) % 3
				return s, nil
			}
			if km.Matches(key, config.ActionPrevPanel) {
				s.PanelFocus = (s.PanelFocus + 2) % 3
				return s, nil
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.PanelFocus = TreePanel
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
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Panel navigation
		if km.Matches(key, config.ActionNextPanel) {
			s.PanelFocus = (s.PanelFocus + 1) % 3
			return s, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			s.PanelFocus = (s.PanelFocus + 2) % 3
			return s, nil
		}

		switch s.PanelFocus {
		case TreePanel:
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
					s.PanelFocus = ContentPanel
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
					s.PanelFocus = ContentPanel
					return s, nil
				}
			}

		case ContentPanel:
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.PanelFocus = TreePanel
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

		case RoutePanel:
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.PanelFocus = ContentPanel
				return s, nil
			}
		}

		// Common keys LAST (quit, back handled per-panel above; cursor handled per-panel)
		if km.Matches(key, config.ActionQuit) {
			return s, tea.Quit
		}
	}

	return s, nil
}

func (s *ConfigScreen) KeyHints(km config.KeyMap) []KeyHint {
	switch s.PanelFocus {
	case ContentPanel:
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
	left := s.renderCategories(ctx)
	center := s.renderFields(ctx)
	right := s.renderFieldDetail(ctx)

	layout := layoutForPage(CONFIGPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	catLen := len(config.AllCategories())
	fieldLen := len(s.ConfigCategoryFields)

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel, TotalLines: catLen, ScrollOffset: ClampScroll(s.Cursor, catLen, innerH)}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[1], Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel, TotalLines: fieldLen, ScrollOffset: ClampScroll(s.ConfigFieldCursor, fieldLen, innerH)}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

// renderCategories renders the category list for the left panel.
func (s *ConfigScreen) renderCategories(ctx AppContext) string {
	categories := config.AllCategories()
	items := make([]string, 0, len(categories)+1)
	for _, cat := range categories {
		items = append(items, config.CategoryLabel(cat))
	}
	items = append(items, "View Raw JSON")

	lines := make([]string, 0, len(items))
	for i, label := range items {
		cursor := "   "
		if s.PanelFocus == TreePanel && s.Cursor == i {
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

// renderFields renders the config fields for the center panel.
func (s *ConfigScreen) renderFields(ctx AppContext) string {
	if s.ConfigCategory == "" {
		return "Select a category"
	}

	if s.ConfigCategory == "raw_json" {
		return s.Viewport.View()
	}

	if len(s.ConfigCategoryFields) == 0 {
		return "(no fields)"
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

		if s.PanelFocus == ContentPanel && i == s.ConfigFieldCursor {
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

// renderFieldDetail renders the detail view for the right panel.
func (s *ConfigScreen) renderFieldDetail(ctx AppContext) string {
	if s.ConfigCategory == "" {
		return "Config\n\n  Select a category to\n  view its fields."
	}

	if s.ConfigCategory == "raw_json" {
		return "Raw JSON\n\n  Scroll with up/down\n  or pgup/pgdn."
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
		fmt.Sprintf("Field: %s", field.Label),
		fmt.Sprintf("Key:   %s", field.JSONKey),
		"",
		fmt.Sprintf("Value: %s", value),
		"",
	}

	if field.Sensitive {
		lines = append(lines, "  (sensitive)")
	}
	if field.HotReloadable {
		lines = append(lines, "  Hot-reloadable")
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Render("  Requires restart"))
	}

	lines = append(lines, "")
	lines = append(lines, "  Press e to edit")

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
