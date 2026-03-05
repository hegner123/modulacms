package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/definitions"
)

// QuickstartScreen implements Screen for the quickstart schema page.
type QuickstartScreen struct {
	Cursor     int
	PanelFocus FocusPanel
}

// NewQuickstartScreen creates a QuickstartScreen.
func NewQuickstartScreen() *QuickstartScreen {
	return &QuickstartScreen{
		Cursor:     0,
		PanelFocus: ContentPanel,
	}
}

func (s *QuickstartScreen) PageIndex() PageIndex { return QUICKSTARTPAGE }

func (s *QuickstartScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	labels := QuickstartMenuLabels()
	cursorMax := len(labels) - 1

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

		// Select schema
		if km.Matches(key, config.ActionSelect) {
			if s.Cursor < len(labels) {
				return s, func() tea.Msg {
					return QuickstartConfirmMsg{SchemaIndex: s.Cursor}
				}
			}
		}

		// Common keys (quit, back, cursor)
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, cursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}
	}

	return s, nil
}

func (s *QuickstartScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionSelect), "install"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *QuickstartScreen) View(ctx AppContext) string {
	left := s.renderSchemas()
	center := s.renderDetail()
	right := s.renderStatus()

	layout := layoutForPage(QUICKSTARTPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	// Single-panel page: render center panel full width
	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: "Schemas", Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: "Details", Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: "Status", Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

func (s *QuickstartScreen) renderSchemas() string {
	labels := QuickstartMenuLabels()
	if len(labels) == 0 {
		return "(no schemas)"
	}
	lines := make([]string, 0, len(labels))
	for i, label := range labels {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

func (s *QuickstartScreen) renderDetail() string {
	defs := definitions.List()
	if len(defs) == 0 || s.Cursor >= len(defs) {
		return "No schema selected"
	}
	def := defs[s.Cursor]
	lines := []string{
		fmt.Sprintf("Name: %s", def.Label),
		fmt.Sprintf("Slug: %s", def.Name),
		"",
		def.Description,
	}
	return strings.Join(lines, "\n")
}

func (s *QuickstartScreen) renderStatus() string {
	lines := []string{
		"Quickstart",
		"",
		"  Install a predefined",
		"  schema definition to",
		"  quickly set up your",
		"  content structure.",
		"",
		"  Press enter to install",
		"  the selected schema.",
	}
	return strings.Join(lines, "\n")
}
