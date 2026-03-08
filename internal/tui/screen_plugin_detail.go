package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// 3/9 grid: left = plugin info, right = actions (top) + action info (bottom)
var pluginDetailGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Plugin"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Actions"},
			{Height: 0.4, Title: "Info"},
		}},
	},
}

// PluginDetailScreen implements Screen for the plugin detail page.
type PluginDetailScreen struct {
	GridScreen
	SelectedPlugin string
	PluginsList    []PluginDisplay
	ActionCursor   int
}

const pluginDetailActionCount = 6

// NewPluginDetailScreen creates a PluginDetailScreen for the named plugin.
func NewPluginDetailScreen(name string, plugins []PluginDisplay) *PluginDetailScreen {
	return &PluginDetailScreen{
		GridScreen: GridScreen{
			Grid:      pluginDetailGrid,
			CursorMax: pluginDetailActionCount - 1,
		},
		SelectedPlugin: name,
		PluginsList:    plugins,
	}
}

func (s *PluginDetailScreen) PageIndex() PageIndex { return PLUGINDETAILPAGE }

func (s *PluginDetailScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		switch s.FocusIndex {
		case 0: // Plugin info (read-only, just nav)
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				return s, HistoryPopCmd()
			}
		case 1: // Actions list
			if km.Matches(key, config.ActionUp) {
				if s.ActionCursor > 0 {
					s.ActionCursor--
				}
				return s, nil
			}
			if km.Matches(key, config.ActionDown) {
				if s.ActionCursor < pluginDetailActionCount-1 {
					s.ActionCursor++
				}
				return s, nil
			}
			if km.Matches(key, config.ActionSelect) {
				return s, s.executeAction(ctx)
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				return s, HistoryPopCmd()
			}
		case 2: // Action info (read-only, actions still work)
			if km.Matches(key, config.ActionSelect) {
				return s, s.executeAction(ctx)
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				s.FocusIndex = 1
				return s, nil
			}
		}

		if km.Matches(key, config.ActionQuit) {
			return s, tea.Quit
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
		s.PluginsList = msg.Data
		return s, LoadingStopCmd()

	// Data refresh (from CMS operations)
	case PluginsListSet:
		s.PluginsList = msg.PluginsList
		return s, nil
	}

	return s, nil
}

// executeAction dispatches the plugin action for the current ActionCursor.
func (s *PluginDetailScreen) executeAction(ctx AppContext) tea.Cmd {
	name := s.SelectedPlugin
	if name == "" {
		return nil
	}
	switch s.ActionCursor {
	case 0:
		return func() tea.Msg {
			return PluginActionRequestMsg{Name: name, Action: PluginActionEnable}
		}
	case 1:
		return func() tea.Msg {
			return PluginActionRequestMsg{Name: name, Action: PluginActionDisable}
		}
	case 2:
		return func() tea.Msg {
			return PluginActionRequestMsg{Name: name, Action: PluginActionReload}
		}
	case 3:
		return FetchPendingRoutesForApprovalScreenCmd(ctx.PluginManager, name)
	case 4:
		return FetchPendingHooksForApprovalScreenCmd(ctx.PluginManager, name)
	case 5:
		return func() tea.Msg {
			return PluginSyncCapabilitiesRequestMsg{Name: name}
		}
	}
	return nil
}

func (s *PluginDetailScreen) KeyHints(km config.KeyMap) []KeyHint {
	switch s.FocusIndex {
	case 1:
		return []KeyHint{
			{km.HintString(config.ActionSelect), "run"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
		}
	default:
		return []KeyHint{
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	}
}

func (s *PluginDetailScreen) View(ctx AppContext) string {
	infoInnerH := s.Grid.CellInnerHeight(0, ctx.Height)
	actionsInnerH := s.Grid.CellInnerHeight(1, ctx.Height)

	infoContent := s.renderInfo()
	infoTotalLines := strings.Count(infoContent, "\n") + 1

	actionsContent := s.renderActions()

	cells := []CellContent{
		{Content: infoContent, TotalLines: infoTotalLines, ScrollOffset: ClampScroll(0, infoTotalLines, infoInnerH)},
		{Content: actionsContent, TotalLines: pluginDetailActionCount, ScrollOffset: ClampScroll(s.ActionCursor, pluginDetailActionCount, actionsInnerH)},
		{Content: s.renderActionInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *PluginDetailScreen) renderInfo() string {
	if s.SelectedPlugin == "" {
		return " No plugin selected"
	}
	var found *PluginDisplay
	for i := range s.PluginsList {
		if s.PluginsList[i].Name == s.SelectedPlugin {
			found = &s.PluginsList[i]
			break
		}
	}
	if found == nil {
		return fmt.Sprintf(" Plugin: %s\n\n  (not found)", s.SelectedPlugin)
	}
	lines := []string{
		fmt.Sprintf(" Name        %s", found.Name),
		fmt.Sprintf(" Version     %s", found.Version),
		fmt.Sprintf(" State       %s", found.State),
		fmt.Sprintf(" Circuit     %s", found.CBState),
	}
	if found.Description != "" {
		lines = append(lines, "", fmt.Sprintf(" %s", found.Description))
	}
	if found.ManifestDrift || found.CapabilityDrifts > 0 || found.SchemaDrifts > 0 {
		lines = append(lines, "", " --- Drift Detected ---")
		if found.ManifestDrift {
			lines = append(lines, "   Manifest changed (hash differs)")
		}
		if found.CapabilityDrifts > 0 {
			lines = append(lines, fmt.Sprintf("   %d capability change(s)", found.CapabilityDrifts))
		}
		if found.SchemaDrifts > 0 {
			lines = append(lines, fmt.Sprintf("   %d schema drift(s)", found.SchemaDrifts))
		}
	}
	return strings.Join(lines, "\n")
}

func (s *PluginDetailScreen) renderActions() string {
	actions := []string{
		"Enable Plugin",
		"Disable Plugin",
		"Reload Plugin",
		"Approve Routes",
		"Approve Hooks",
		"Sync Capabilities",
	}
	lines := make([]string, 0, len(actions))
	for i, action := range actions {
		cursor := "   "
		if s.ActionCursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, action))
	}
	return strings.Join(lines, "\n")
}

func (s *PluginDetailScreen) renderActionInfo() string {
	descriptions := []string{
		" Enable this plugin to activate its features.",
		" Disable this plugin to deactivate its features.",
		" Reload the plugin to apply configuration changes.",
		" Review and approve registered routes.",
		" Review and approve registered hooks.",
		" Sync capabilities with the plugin manifest.",
	}
	if s.ActionCursor < 0 || s.ActionCursor >= len(descriptions) {
		return ""
	}
	return descriptions[s.ActionCursor]
}
