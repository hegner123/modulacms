package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/definitions"
)

// Quickstart grid: 2 columns
//
//	Col 0 (span 3): Schema list
//	Col 1 (span 9): Details (top), Info (bottom)
var quickstartGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Schemas"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.45, Title: "Details"},
			{Height: 0.55, Title: "Info"},
		}},
	},
}

// QuickstartScreen implements Screen for the quickstart schema page.
type QuickstartScreen struct {
	GridScreen
}

// NewQuickstartScreen creates a QuickstartScreen.
func NewQuickstartScreen() *QuickstartScreen {
	defs := definitions.List()
	cursorMax := len(defs) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &QuickstartScreen{
		GridScreen: GridScreen{
			Grid:      quickstartGrid,
			CursorMax: cursorMax,
		},
	}
}

func (s *QuickstartScreen) PageIndex() PageIndex { return QUICKSTARTPAGE }

func (s *QuickstartScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Select schema
		if km.Matches(key, config.ActionSelect) {
			defs := definitions.List()
			if s.Cursor < len(defs) {
				return s, func() tea.Msg {
					return QuickstartConfirmMsg{SchemaIndex: s.Cursor}
				}
			}
		}

		// Common keys (quit, back, cursor)
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
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
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *QuickstartScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderSchemas()},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

// ---------------------------------------------------------------------------
// Render helpers
// ---------------------------------------------------------------------------

func (s *QuickstartScreen) renderSchemas() string {
	labels := QuickstartMenuLabels()
	if len(labels) == 0 {
		return " (no schemas)"
	}
	lines := make([]string, 0, len(labels))
	for i, label := range labels {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

func (s *QuickstartScreen) renderDetail() string {
	defs := definitions.List()
	if len(defs) == 0 || s.Cursor >= len(defs) {
		return " No schema selected"
	}
	def := defs[s.Cursor]

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Label   %s", def.Label),
		fmt.Sprintf(" Format  %s", def.Format),
		"",
		faint.Render(fmt.Sprintf(" Slug    %s", def.Name)),
		"",
		fmt.Sprintf(" %s", def.Description),
	}
	return strings.Join(lines, "\n")
}

func (s *QuickstartScreen) renderInfo() string {
	defs := definitions.List()
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	lines := []string{
		accent.Render(fmt.Sprintf(" Schemas: %d", len(defs))),
		"",
		" Install a predefined schema definition",
		" to quickly set up your content structure.",
		"",
		" Press enter to install the selected schema.",
	}
	return strings.Join(lines, "\n")
}
