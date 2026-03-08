package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// SelectPluginAndNavigateMsg is emitted by the PluginsScreen when the user
// selects a plugin. It sets Model.SelectedPlugin and then navigates to the
// plugin detail page. Handled in the ActiveScreen dispatch block in update.go.
type SelectPluginAndNavigateMsg struct {
	PluginName string
}

// 3/9 grid: left = plugin list, right = detail (top) + info (bottom)
var pluginsGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Plugins"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Details"},
			{Height: 0.4, Title: "Info"},
		}},
	},
}

// PluginsScreen implements Screen for the plugins list page (PLUGINSPAGE).
type PluginsScreen struct {
	GridScreen
	Plugins []PluginDisplay
}

// NewPluginsScreen creates a PluginsScreen with the given plugin list.
func NewPluginsScreen(plugins []PluginDisplay) *PluginsScreen {
	cursorMax := len(plugins) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &PluginsScreen{
		GridScreen: GridScreen{
			Grid:      pluginsGrid,
			CursorMax: cursorMax,
		},
		Plugins: plugins,
	}
}

func (s *PluginsScreen) PageIndex() PageIndex { return PLUGINSPAGE }

func (s *PluginsScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Select plugin: navigate to detail page
		if km.Matches(key, config.ActionSelect) {
			if len(s.Plugins) > 0 && s.Cursor < len(s.Plugins) {
				name := s.Plugins[s.Cursor].Name
				return s, func() tea.Msg {
					return SelectPluginAndNavigateMsg{PluginName: name}
				}
			}
		}

		// Common keys (quit, back, cursor) -- LAST
		cursorMax := len(s.Plugins) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		s.CursorMax = cursorMax
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case PluginsFetchMsg:
		mgr := ctx.PluginManager
		if mgr == nil {
			return s, func() tea.Msg {
				return PluginsFetchResultsMsg{Data: []PluginDisplay{}}
			}
		}
		return s, func() tea.Msg {
			instances := mgr.ListPlugins()
			displays := make([]PluginDisplay, 0, len(instances))
			for _, inst := range instances {
				cbState := "closed"
				if inst.CB != nil {
					cbState = inst.CB.State().String()
				}
				displays = append(displays, PluginDisplay{
					Name:             inst.Info.Name,
					Version:          inst.Info.Version,
					State:            inst.State.String(),
					CBState:          cbState,
					Description:      inst.Info.Description,
					ManifestDrift:    inst.ManifestDrift,
					CapabilityDrifts: len(inst.CapabilityDrift),
					SchemaDrifts:     len(inst.SchemaDrift),
				})
			}
			return PluginsFetchResultsMsg{Data: displays}
		}
	case PluginsFetchResultsMsg:
		s.Plugins = msg.Data
		s.Cursor = 0
		s.CursorMax = len(s.Plugins) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		return s, LoadingStopCmd()

	// Data refresh (from CMS operations)
	case PluginsListSet:
		s.Plugins = msg.PluginsList
		s.CursorMax = len(s.Plugins) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		if s.Cursor > s.CursorMax && s.CursorMax >= 0 {
			s.Cursor = s.CursorMax
		}
		return s, nil
	}

	return s, nil
}

func (s *PluginsScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionSelect), "view"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *PluginsScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList(), TotalLines: len(s.Plugins), ScrollOffset: ClampScroll(s.Cursor, len(s.Plugins), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *PluginsScreen) renderList() string {
	if len(s.Plugins) == 0 {
		return "(no plugins)"
	}

	lines := make([]string, 0, len(s.Plugins))
	for i, p := range s.Plugins {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		stateIndicator := p.State
		if p.CBState == "open" {
			stateIndicator = "tripped"
		}
		drift := ""
		if p.ManifestDrift || p.CapabilityDrifts > 0 || p.SchemaDrifts > 0 {
			drift = " [drift]"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]%s", cursor, p.Name, stateIndicator, drift))
	}
	return strings.Join(lines, "\n")
}

func (s *PluginsScreen) renderDetail() string {
	if len(s.Plugins) == 0 || s.Cursor >= len(s.Plugins) {
		return " No plugin selected"
	}

	p := s.Plugins[s.Cursor]
	lines := []string{
		fmt.Sprintf(" Name        %s", p.Name),
		fmt.Sprintf(" Version     %s", p.Version),
		fmt.Sprintf(" State       %s", p.State),
		fmt.Sprintf(" Circuit     %s", p.CBState),
	}
	if p.Description != "" {
		lines = append(lines, "", fmt.Sprintf(" %s", p.Description))
	}

	// Drift warnings
	if p.ManifestDrift || p.CapabilityDrifts > 0 || p.SchemaDrifts > 0 {
		lines = append(lines, "", " --- Drift Detected ---")
		if p.ManifestDrift {
			lines = append(lines, "   Manifest changed (hash differs)")
		}
		if p.CapabilityDrifts > 0 {
			lines = append(lines, fmt.Sprintf("   %d capability change(s)", p.CapabilityDrifts))
		}
		if p.SchemaDrifts > 0 {
			lines = append(lines, fmt.Sprintf("   %d schema drift(s)", p.SchemaDrifts))
		}
	}

	return strings.Join(lines, "\n")
}

func (s *PluginsScreen) renderInfo() string {
	lines := []string{
		" Plugin Manager",
		"",
		fmt.Sprintf("   Total: %d", len(s.Plugins)),
	}

	running := 0
	failed := 0
	stopped := 0
	for _, p := range s.Plugins {
		switch p.State {
		case "running":
			running++
		case "failed":
			failed++
		case "stopped":
			stopped++
		}
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("   Running: %d", running))
	if failed > 0 {
		lines = append(lines, fmt.Sprintf("   Failed:  %d", failed))
	}
	if stopped > 0 {
		lines = append(lines, fmt.Sprintf("   Stopped: %d", stopped))
	}

	return strings.Join(lines, "\n")
}
