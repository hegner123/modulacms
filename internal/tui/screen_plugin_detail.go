package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// PluginDetailScreen implements Screen for the plugin detail page.
type PluginDetailScreen struct {
	SelectedPlugin string
	PluginsList    []PluginDisplay
	Cursor         int
	CursorMax      int // 5 (6 action items: 0-5)
	PanelFocus     FocusPanel
}

// NewPluginDetailScreen creates a PluginDetailScreen for the named plugin.
func NewPluginDetailScreen(name string, plugins []PluginDisplay) *PluginDetailScreen {
	return &PluginDetailScreen{
		SelectedPlugin: name,
		PluginsList:    plugins,
		Cursor:         0,
		CursorMax:      5,
		PanelFocus:     ContentPanel,
	}
}

func (s *PluginDetailScreen) PageIndex() PageIndex { return PLUGINDETAILPAGE }

func (s *PluginDetailScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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

		// Select action
		if km.Matches(key, config.ActionSelect) {
			name := s.SelectedPlugin
			if name == "" {
				return s, nil
			}
			switch s.Cursor {
			case 0:
				return s, func() tea.Msg {
					return PluginActionRequestMsg{Name: name, Action: PluginActionEnable}
				}
			case 1:
				return s, func() tea.Msg {
					return PluginActionRequestMsg{Name: name, Action: PluginActionDisable}
				}
			case 2:
				return s, func() tea.Msg {
					return PluginActionRequestMsg{Name: name, Action: PluginActionReload}
				}
			case 3:
				return s, FetchPendingRoutesForApprovalScreenCmd(ctx.PluginManager, name)
			case 4:
				return s, FetchPendingHooksForApprovalScreenCmd(ctx.PluginManager, name)
			case 5:
				return s, func() tea.Msg {
					return PluginSyncCapabilitiesRequestMsg{Name: name}
				}
			}
		}

		// Title font cycling
		if km.Matches(key, config.ActionTitlePrev) {
			return s, TitleFontPreviousCmd()
		}
		if km.Matches(key, config.ActionTitleNext) {
			return s, TitleFontNextCmd()
		}

		// Common keys (quit, back, cursor)
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
		s.PluginsList = msg.Data
		return s, LoadingStopCmd()

	// Data refresh (from CMS operations)
	case PluginsListSet:
		s.PluginsList = msg.PluginsList
		return s, nil
	}

	return s, nil
}

func (s *PluginDetailScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionSelect), "run"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *PluginDetailScreen) View(ctx AppContext) string {
	left := s.renderInfo()
	center := s.renderActions()
	right := s.renderActionInfo()

	layout := layoutForPage(PLUGINDETAILPAGE)
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
		panels = append(panels, Panel{Title: "Plugin", Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: "Actions", Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: "Info", Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

func (s *PluginDetailScreen) renderInfo() string {
	if s.SelectedPlugin == "" {
		return "No plugin selected"
	}
	var found *PluginDisplay
	for i := range s.PluginsList {
		if s.PluginsList[i].Name == s.SelectedPlugin {
			found = &s.PluginsList[i]
			break
		}
	}
	if found == nil {
		return fmt.Sprintf("Plugin: %s\n\n  (not found)", s.SelectedPlugin)
	}
	lines := []string{
		fmt.Sprintf("Name:    %s", found.Name),
		fmt.Sprintf("Version: %s", found.Version),
		fmt.Sprintf("State:   %s", found.State),
		fmt.Sprintf("Circuit: %s", found.CBState),
	}
	if found.ManifestDrift || found.CapabilityDrifts > 0 || found.SchemaDrifts > 0 {
		lines = append(lines, "", "Drift detected")
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
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, action))
	}
	return strings.Join(lines, "\n")
}

func (s *PluginDetailScreen) renderActionInfo() string {
	descriptions := []string{
		"Enable this plugin to activate its features",
		"Disable this plugin to deactivate its features",
		"Reload the plugin to apply configuration changes",
		"Review and approve registered routes",
		"Review and approve registered hooks",
		"Sync capabilities with the plugin manifest",
	}
	if s.Cursor < 0 || s.Cursor >= len(descriptions) {
		return ""
	}
	return descriptions[s.Cursor]
}
