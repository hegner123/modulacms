package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// Pipelines grid: 3 columns
//
//	Col 0 (span 3): Pipeline Chains
//	Col 1 (span 6): Chain Info (65%), Help (35%)
//	Col 2 (span 3): Registry (55%), By Table (45%)
var pipelinesGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1, Title: "Pipeline Chains"},
		}},
		{Span: 6, Cells: []GridCell{
			{Height: 0.65, Title: "Chain Info"},
			{Height: 0.35, Title: "Help"},
		}},
		{Span: 3, Cells: []GridCell{
			{Height: 0.55, Title: "Registry"},
			{Height: 0.45, Title: "By Table"},
		}},
	},
}

// PipelinesScreen implements Screen for the pipelines list page.
type PipelinesScreen struct {
	GridScreen
	PipelinesList []PipelineDisplay
}

// NewPipelinesScreen creates a PipelinesScreen for the pipeline chain list.
func NewPipelinesScreen(pipelines []PipelineDisplay) *PipelinesScreen {
	max := len(pipelines) - 1
	if max < 0 {
		max = 0
	}
	return &PipelinesScreen{
		GridScreen: GridScreen{
			Grid:      pipelinesGrid,
			CursorMax: max,
		},
		PipelinesList: pipelines,
	}
}

func (s *PipelinesScreen) PageIndex() PageIndex { return PIPELINESPAGE }

func (s *PipelinesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Select pipeline: emit message to set key on Model and navigate
		if km.Matches(key, config.ActionSelect) {
			if s.FocusIndex == 0 && len(s.PipelinesList) > 0 && s.Cursor < len(s.PipelinesList) {
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
	}
}

func (s *PipelinesScreen) View(ctx AppContext) string {
	listLen := len(s.PipelinesList)
	innerH := s.Grid.CellInnerHeight(0, ctx.Height)

	cells := []CellContent{
		{Content: s.renderChains(), TotalLines: listLen, ScrollOffset: ClampScroll(s.Cursor, listLen, innerH)},
		{Content: s.renderChainInfo()},
		{Content: s.renderHelp()},
		{Content: s.renderRegistry()},
		{Content: s.renderByTable()},
	}
	return s.RenderGrid(ctx, cells)
}

// renderChains renders the pipeline chain list for the left panel.
func (s *PipelinesScreen) renderChains() string {
	if len(s.PipelinesList) == 0 {
		return " (no pipeline chains)"
	}

	lines := make([]string, 0, len(s.PipelinesList))
	for i, p := range s.PipelinesList {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s.%s_%s (%d)", cursor, p.Table, p.Phase, p.Operation, p.Count))
	}
	return strings.Join(lines, "\n")
}

// renderChainInfo renders detailed info about the selected pipeline chain.
func (s *PipelinesScreen) renderChainInfo() string {
	if len(s.PipelinesList) == 0 || s.Cursor >= len(s.PipelinesList) {
		return " No pipeline selected"
	}

	p := s.PipelinesList[s.Cursor]
	accent := lipgloss.NewStyle().Bold(true)

	lines := []string{
		accent.Render(" " + p.Key),
		"",
		fmt.Sprintf(" Table      %s", p.Table),
		fmt.Sprintf(" Operation  %s", p.Operation),
		fmt.Sprintf(" Phase      %s", p.Phase),
		fmt.Sprintf(" Entries    %d", p.Count),
	}

	return strings.Join(lines, "\n")
}

// renderHelp renders usage hints.
func (s *PipelinesScreen) renderHelp() string {
	lines := []string{
		" Press Enter to view pipeline entries.",
		" Use Tab to switch between panels.",
		"",
		" Pipeline chains are registered by plugins",
		" and run before/after database operations.",
	}
	return strings.Join(lines, "\n")
}

// renderRegistry renders aggregate stats.
func (s *PipelinesScreen) renderRegistry() string {
	totalEntries := 0
	for _, p := range s.PipelinesList {
		totalEntries += p.Count
	}

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

	lines := []string{
		fmt.Sprintf(" Chains   %d", len(s.PipelinesList)),
		fmt.Sprintf(" Entries  %d", totalEntries),
		"",
		fmt.Sprintf(" Before   %d", before),
		fmt.Sprintf(" After    %d", after),
	}
	return strings.Join(lines, "\n")
}

// renderByTable groups chains by table name.
func (s *PipelinesScreen) renderByTable() string {
	if len(s.PipelinesList) == 0 {
		return " (no chains)"
	}

	// Preserve insertion order
	var tableOrder []string
	counts := make(map[string]int)
	for _, p := range s.PipelinesList {
		if _, exists := counts[p.Table]; !exists {
			tableOrder = append(tableOrder, p.Table)
		}
		counts[p.Table]++
	}

	lines := make([]string, 0, len(tableOrder))
	for _, table := range tableOrder {
		lines = append(lines, fmt.Sprintf(" %-12s %d chains", table, counts[table]))
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// PipelineDetailScreen — entries for a selected pipeline (PIPELINEDETAILPAGE)
// ---------------------------------------------------------------------------

// Pipeline detail grid: 3 columns
//
//	Col 0 (span 3): All Chains (read-only context)
//	Col 1 (span 6): Entries (70%), Entry Detail (30%)
//	Col 2 (span 3): Chain Status (50%), Execution Order (50%)
var pipelineDetailGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1, Title: "All Chains"},
		}},
		{Span: 6, Cells: []GridCell{
			{Height: 0.70, Title: "Entries"},
			{Height: 0.30, Title: "Entry Detail"},
		}},
		{Span: 3, Cells: []GridCell{
			{Height: 0.50, Title: "Chain Status"},
			{Height: 0.50, Title: "Execution Order"},
		}},
	},
}

// PipelineDetailScreen implements Screen for the pipeline detail page.
type PipelineDetailScreen struct {
	GridScreen
	PipelinesList       []PipelineDisplay
	PipelineEntries     []PipelineEntryDisplay
	SelectedPipelineKey string
}

// NewPipelineDetailScreen creates a PipelineDetailScreen for a selected pipeline chain.
func NewPipelineDetailScreen(pipelines []PipelineDisplay, entries []PipelineEntryDisplay, selectedKey string) *PipelineDetailScreen {
	max := len(entries) - 1
	if max < 0 {
		max = 0
	}
	return &PipelineDetailScreen{
		GridScreen: GridScreen{
			Grid:       pipelineDetailGrid,
			FocusIndex: 1, // start focused on entries list
			CursorMax:  max,
		},
		PipelinesList:       pipelines,
		PipelineEntries:     entries,
		SelectedPipelineKey: selectedKey,
	}
}

func (s *PipelineDetailScreen) PageIndex() PageIndex { return PIPELINEDETAILPAGE }

func (s *PipelineDetailScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Common keys (quit, back, cursor) — only move cursor when focused on entries (cell 1)
		if s.FocusIndex == 1 {
			newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
			if handled {
				s.Cursor = newCursor
				return s, cmd
			}
		} else {
			// Still handle quit/back from any cell
			if km.Matches(key, config.ActionQuit) {
				return s, tea.Quit
			}
			if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
				return s, HistoryPopCmd()
			}
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
	}
}

func (s *PipelineDetailScreen) View(ctx AppContext) string {
	entriesLen := len(s.PipelineEntries)
	innerH := s.Grid.CellInnerHeight(1, ctx.Height)

	cells := []CellContent{
		{Content: s.renderAllChains()},
		{Content: s.renderEntries(), TotalLines: entriesLen, ScrollOffset: ClampScroll(s.Cursor, entriesLen, innerH)},
		{Content: s.renderEntryDetail()},
		{Content: s.renderChainStatus()},
		{Content: s.renderExecution()},
	}
	return s.RenderGrid(ctx, cells)
}

// renderAllChains renders the pipeline chain list with the selected chain highlighted.
func (s *PipelineDetailScreen) renderAllChains() string {
	if len(s.PipelinesList) == 0 {
		return " (no pipeline chains)"
	}

	lines := make([]string, 0, len(s.PipelinesList))
	for _, p := range s.PipelinesList {
		marker := "  "
		if p.Key == s.SelectedPipelineKey {
			marker = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s (%d)", marker, p.Key, p.Count))
	}
	return strings.Join(lines, "\n")
}

// renderEntries renders the entries list for the selected pipeline chain.
func (s *PipelineDetailScreen) renderEntries() string {
	if len(s.PipelineEntries) == 0 {
		return " No entries"
	}

	lines := make([]string, 0, len(s.PipelineEntries))
	for i, e := range s.PipelineEntries {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		enabled := "on"
		if !e.Enabled {
			enabled = "off"
		}
		lines = append(lines, fmt.Sprintf(" %s %s -> %s [%s]", cursor, e.PluginName, e.Handler, enabled))
	}
	return strings.Join(lines, "\n")
}

// renderEntryDetail renders full details of the selected entry.
func (s *PipelineDetailScreen) renderEntryDetail() string {
	if len(s.PipelineEntries) == 0 || s.Cursor >= len(s.PipelineEntries) {
		return " No entry selected"
	}

	e := s.PipelineEntries[s.Cursor]
	accent := lipgloss.NewStyle().Bold(true)

	status := "Enabled"
	if !e.Enabled {
		status = "Disabled"
	}

	lines := []string{
		accent.Render(" " + e.PluginName),
		"",
		fmt.Sprintf(" Handler   %s", e.Handler),
		fmt.Sprintf(" Priority  %d", e.Priority),
		fmt.Sprintf(" Status    %s", status),
		fmt.Sprintf(" ID        %s", e.PipelineID),
	}
	return strings.Join(lines, "\n")
}

// renderChainStatus renders enabled/disabled/total counts.
func (s *PipelineDetailScreen) renderChainStatus() string {
	if s.SelectedPipelineKey == "" {
		return ""
	}

	enabledCount := 0
	for _, e := range s.PipelineEntries {
		if e.Enabled {
			enabledCount++
		}
	}

	lines := []string{
		fmt.Sprintf(" Total     %d", len(s.PipelineEntries)),
		fmt.Sprintf(" Enabled   %d", enabledCount),
		fmt.Sprintf(" Disabled  %d", len(s.PipelineEntries)-enabledCount),
	}
	return strings.Join(lines, "\n")
}

// renderExecution renders the execution order of enabled entries.
func (s *PipelineDetailScreen) renderExecution() string {
	if len(s.PipelineEntries) == 0 {
		return " (no entries)"
	}

	var lines []string
	step := 1
	for _, e := range s.PipelineEntries {
		if !e.Enabled {
			continue
		}
		lines = append(lines, fmt.Sprintf(" %d. %s", step, e.Handler))
		step++
	}

	if len(lines) == 0 {
		return " (no enabled entries)"
	}
	return strings.Join(lines, "\n")
}
