package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
)

// SelectPipelineAndNavigateMsg is emitted by the PipelinesScreen when the user
// selects a pipeline. It sets Model.SelectedPipelineKey and then navigates to
// the pipeline detail page. Handled in the ActiveScreen dispatch block in update.go.
type SelectPipelineAndNavigateMsg struct {
	PipelineKey string
}

// ---------------------------------------------------------------------------
// PipelinesScreen — list of pipeline chains (PIPELINESPAGE)
// ---------------------------------------------------------------------------

// PipelinesScreen implements Screen for the pipelines list page.
type PipelinesScreen struct {
	PipelinesList []PipelineDisplay
	Cursor        int
	CursorMax     int
	PanelFocus    FocusPanel
}

// NewPipelinesScreen creates a PipelinesScreen for the pipeline chain list.
func NewPipelinesScreen(pipelines []PipelineDisplay) *PipelinesScreen {
	max := len(pipelines) - 1
	if max < 0 {
		max = 0
	}
	return &PipelinesScreen{
		PipelinesList: pipelines,
		Cursor:        0,
		CursorMax:     max,
		PanelFocus:    TreePanel,
	}
}

func (s *PipelinesScreen) PageIndex() PageIndex { return PIPELINESPAGE }

func (s *PipelinesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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

		// Select pipeline: emit message to set key on Model and navigate
		if km.Matches(key, config.ActionSelect) {
			if len(s.PipelinesList) > 0 && s.Cursor < len(s.PipelinesList) {
				selectedKey := s.PipelinesList[s.Cursor].Key
				return s, func() tea.Msg {
					return SelectPipelineAndNavigateMsg{PipelineKey: selectedKey}
				}
			}
		}

		// Common keys (quit, back, cursor)
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case PipelinesFetchMsg:
		mgr := ctx.PluginManager
		if mgr == nil {
			return s, func() tea.Msg {
				return PipelinesFetchResultsMsg{Data: []PipelineDisplay{}}
			}
		}
		return s, func() tea.Msg {
			results := mgr.DryRunAllPipelines()
			displays := make([]PipelineDisplay, 0, len(results))
			for _, r := range results {
				displays = append(displays, PipelineDisplay{
					Key:       r.Table + "." + r.Phase + "_" + r.Operation,
					Table:     r.Table,
					Operation: r.Operation,
					Phase:     r.Phase,
					Count:     len(r.Entries),
				})
			}
			return PipelinesFetchResultsMsg{Data: displays}
		}
	case PipelinesFetchResultsMsg:
		s.PipelinesList = msg.Data
		s.CursorMax = len(s.PipelinesList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		s.Cursor = 0
		return s, LoadingStopCmd()

	// Data refresh messages (from CMS operations)
	case PipelinesListSet:
		s.PipelinesList = msg.PipelinesList
		s.CursorMax = len(s.PipelinesList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		if s.Cursor > s.CursorMax {
			s.Cursor = s.CursorMax
		}
		return s, nil
	}

	return s, nil
}

func (s *PipelinesScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionSelect), "view"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *PipelinesScreen) View(ctx AppContext) string {
	left := s.renderList()
	center := s.renderDetail()
	right := s.renderInfo()

	layout := layoutForPage(PIPELINESPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	listLen := len(s.PipelinesList)

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel, TotalLines: listLen, ScrollOffset: ClampScroll(s.Cursor, listLen, innerH)}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[1], Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

// renderList renders the pipeline chain list for the left panel.
func (s *PipelinesScreen) renderList() string {
	if len(s.PipelinesList) == 0 {
		return "(no pipeline chains)"
	}

	lines := make([]string, 0, len(s.PipelinesList))
	for i, p := range s.PipelinesList {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s.%s_%s (%d)", cursor, p.Table, p.Phase, p.Operation, p.Count))
	}
	return strings.Join(lines, "\n")
}

// renderDetail renders the entries for the selected pipeline chain (center panel).
func (s *PipelinesScreen) renderDetail() string {
	if len(s.PipelinesList) == 0 || s.Cursor >= len(s.PipelinesList) {
		return "No pipeline selected"
	}

	return "Press enter to view pipeline entries"
}

// renderInfo renders the pipeline summary for the right panel.
func (s *PipelinesScreen) renderInfo() string {
	totalEntries := 0
	for _, p := range s.PipelinesList {
		totalEntries += p.Count
	}

	lines := []string{
		"Pipeline Registry",
		"",
		fmt.Sprintf("  Chains: %d", len(s.PipelinesList)),
		fmt.Sprintf("  Total entries: %d", totalEntries),
	}

	// Count by phase
	before := 0
	after := 0
	for _, p := range s.PipelinesList {
		switch p.Phase {
		case "before":
			before++
		case "after":
			after++
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Before chains: %d", before))
	lines = append(lines, fmt.Sprintf("  After chains:  %d", after))

	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// PipelineDetailScreen — entries for a selected pipeline (PIPELINEDETAILPAGE)
// ---------------------------------------------------------------------------

// PipelineDetailScreen implements Screen for the pipeline detail page.
type PipelineDetailScreen struct {
	PipelinesList       []PipelineDisplay
	PipelineEntries     []PipelineEntryDisplay
	SelectedPipelineKey string
	Cursor              int
	CursorMax           int
	PanelFocus          FocusPanel
}

// NewPipelineDetailScreen creates a PipelineDetailScreen for a selected pipeline chain.
func NewPipelineDetailScreen(pipelines []PipelineDisplay, entries []PipelineEntryDisplay, selectedKey string) *PipelineDetailScreen {
	max := len(entries) - 1
	if max < 0 {
		max = 0
	}
	return &PipelineDetailScreen{
		PipelinesList:       pipelines,
		PipelineEntries:     entries,
		SelectedPipelineKey: selectedKey,
		Cursor:              0,
		CursorMax:           max,
		PanelFocus:          ContentPanel,
	}
}

func (s *PipelineDetailScreen) PageIndex() PageIndex { return PIPELINEDETAILPAGE }

func (s *PipelineDetailScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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

		// Common keys (quit, back, cursor)
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case PipelineEntriesFetchMsg:
		mgr := ctx.PluginManager
		if mgr == nil {
			return s, func() tea.Msg {
				return PipelineEntriesFetchResultsMsg{Entries: []PipelineEntryDisplay{}}
			}
		}
		key := msg.Key
		return s, func() tea.Msg {
			results := mgr.DryRunAllPipelines()
			var matchedEntries []PipelineEntryDisplay
			for _, r := range results {
				rKey := r.Table + "." + r.Phase + "_" + r.Operation
				if rKey == key {
					for _, e := range r.Entries {
						matchedEntries = append(matchedEntries, PipelineEntryDisplay{
							PipelineID: e.PipelineID,
							PluginName: e.PluginName,
							Handler:    e.Handler,
							Priority:   e.Priority,
							Enabled:    e.Enabled,
						})
					}
					break
				}
			}
			return PipelineEntriesFetchResultsMsg{Entries: matchedEntries}
		}
	case PipelineEntriesFetchResultsMsg:
		s.PipelineEntries = msg.Entries
		s.CursorMax = len(s.PipelineEntries) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		s.Cursor = 0
		return s, nil

	// Data refresh messages (from CMS operations)
	case PipelinesListSet:
		s.PipelinesList = msg.PipelinesList
		return s, nil

	case PipelineEntriesSet:
		s.PipelineEntries = msg.PipelineEntries
		s.CursorMax = len(s.PipelineEntries) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		if s.Cursor > s.CursorMax {
			s.Cursor = s.CursorMax
		}
		return s, nil
	}

	return s, nil
}

func (s *PipelineDetailScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *PipelineDetailScreen) View(ctx AppContext) string {
	left := s.renderPipelinesList()
	center := s.renderEntries()
	right := s.renderStatus()

	layout := layoutForPage(PIPELINEDETAILPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	entriesLen := len(s.PipelineEntries)

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[1], Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel, TotalLines: entriesLen, ScrollOffset: ClampScroll(s.Cursor, entriesLen, innerH)}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

// renderPipelinesList renders the pipeline chain list for the left panel on the detail page.
func (s *PipelineDetailScreen) renderPipelinesList() string {
	if len(s.PipelinesList) == 0 {
		return "(no pipeline chains)"
	}

	lines := make([]string, 0, len(s.PipelinesList))
	for _, p := range s.PipelinesList {
		marker := "   "
		if p.Key == s.SelectedPipelineKey {
			marker = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s.%s_%s (%d)", marker, p.Table, p.Phase, p.Operation, p.Count))
	}
	return strings.Join(lines, "\n")
}

// renderEntries renders the entries for the selected pipeline chain (center panel).
func (s *PipelineDetailScreen) renderEntries() string {
	if len(s.PipelineEntries) == 0 {
		return "No entries (select a chain to view)"
	}

	lines := make([]string, 0, len(s.PipelineEntries)+2)
	lines = append(lines, "Pipeline entries (priority order):")
	lines = append(lines, "")
	for i, e := range s.PipelineEntries {
		enabled := "on"
		if !e.Enabled {
			enabled = "off"
		}
		lines = append(lines, fmt.Sprintf("  %d. %s -> %s (pri:%d, %s)", i+1, e.PluginName, e.Handler, e.Priority, enabled))
	}
	return strings.Join(lines, "\n")
}

// renderStatus renders the status summary for the right panel on the detail page.
func (s *PipelineDetailScreen) renderStatus() string {
	if s.SelectedPipelineKey == "" {
		return ""
	}

	lines := []string{
		fmt.Sprintf("Chain: %s", s.SelectedPipelineKey),
		"",
		fmt.Sprintf("  Entries: %d", len(s.PipelineEntries)),
	}

	enabledCount := 0
	for _, e := range s.PipelineEntries {
		if e.Enabled {
			enabledCount++
		}
	}
	lines = append(lines, fmt.Sprintf("  Enabled: %d", enabledCount))
	lines = append(lines, fmt.Sprintf("  Disabled: %d", len(s.PipelineEntries)-enabledCount))

	return strings.Join(lines, "\n")
}
