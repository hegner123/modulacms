package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/search"
)

var searchGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Actions"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.5, Title: "Index Stats"},
			{Height: 0.5, Title: "Info"},
		}},
	},
}

type searchAction struct {
	Label string
	Desc  string
}

var searchActions = []searchAction{
	{"View Stats", "Refresh index statistics"},
	{"Rebuild Index", "Rebuild the full-text search index from published content"},
}

// SearchStatsFetchMsg requests fetching search index stats.
type SearchStatsFetchMsg struct{}

// SearchStatsFetchResultsMsg returns the fetched stats.
type SearchStatsFetchResultsMsg struct {
	Stats   search.IndexStats
	Enabled bool
}

// SearchRebuildRequestMsg triggers a search index rebuild.
type SearchRebuildRequestMsg struct{}

// SearchRebuildCompleteMsg carries the rebuild result.
type SearchRebuildCompleteMsg struct {
	Stats *search.IndexStats
	Err   error
}

// SearchScreen implements Screen for search index management.
type SearchScreen struct {
	GridScreen
	Stats      *search.IndexStats
	Enabled    bool
	Rebuilding bool
	RebuildErr error
}

func NewSearchScreen() *SearchScreen {
	return &SearchScreen{
		GridScreen: GridScreen{
			Grid:      searchGrid,
			CursorMax: len(searchActions) - 1,
		},
	}
}

func (s *SearchScreen) PageIndex() PageIndex { return SEARCHPAGE }

func (s *SearchScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Select action
		if km.Matches(key, config.ActionSelect) {
			switch s.Cursor {
			case 0: // View Stats
				return s, s.fetchStats(ctx)
			case 1: // Rebuild Index
				if !s.Enabled {
					return s, ShowDialogCmd("search Disabled", "search is not enabled in the configuration.", false, DIALOGGENERIC)
				}
				s.Rebuilding = true
				s.RebuildErr = nil
				return s, s.runRebuild(ctx)
			}
		}

		s.CursorMax = len(searchActions) - 1
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	case SearchStatsFetchMsg:
		return s, s.fetchStats(ctx)

	case SearchStatsFetchResultsMsg:
		s.Enabled = msg.Enabled
		s.Stats = &msg.Stats
		return s, LoadingStopCmd()

	case SearchRebuildCompleteMsg:
		s.Rebuilding = false
		if msg.Err != nil {
			s.RebuildErr = msg.Err
		} else if msg.Stats != nil {
			s.Stats = msg.Stats
		}
		return s, LoadingStopCmd()
	}

	return s, nil
}

func (s *SearchScreen) fetchStats(ctx AppContext) tea.Cmd {
	cfg := ctx.Config
	if cfg == nil {
		return func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("config not loaded")} }
	}
	d := ctx.DB
	if d == nil {
		return func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
	}
	enabled := cfg.SearchEnabled()
	searchPath := cfg.Search_Path
	return func() tea.Msg {
		if !enabled {
			return SearchStatsFetchResultsMsg{Enabled: false}
		}
		searchCfg := search.DefaultConfig()
		searchCfg.IndexPath = searchPath
		svc := search.NewService(d, searchCfg)
		stats := svc.Stats()
		return SearchStatsFetchResultsMsg{Stats: stats, Enabled: true}
	}
}

func (s *SearchScreen) runRebuild(ctx AppContext) tea.Cmd {
	cfg := ctx.Config
	d := ctx.DB
	searchPath := cfg.Search_Path
	return func() tea.Msg {
		searchCfg := search.DefaultConfig()
		searchCfg.IndexPath = searchPath
		svc := search.NewService(d, searchCfg)
		if err := svc.Rebuild(); err != nil {
			return SearchRebuildCompleteMsg{Err: err}
		}
		stats := svc.Stats()
		return SearchRebuildCompleteMsg{Stats: &stats}
	}
}

func (s *SearchScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionSelect), "run"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

func (s *SearchScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderActions()},
		{Content: s.renderStats()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *SearchScreen) renderActions() string {
	lines := make([]string, 0, len(searchActions))
	for i, a := range searchActions {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, a.Label))
	}
	return strings.Join(lines, "\n")
}

func (s *SearchScreen) renderStats() string {
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	if !s.Enabled {
		return faint.Render(" Search is disabled in configuration")
	}

	if s.Stats == nil {
		return faint.Render(" Press enter on 'View Stats' to load")
	}

	st := s.Stats
	lines := []string{
		accent.Render(" Index Statistics"),
		"",
		fmt.Sprintf("   Documents: %d", st.Documents),
		fmt.Sprintf("   Terms:     %d", st.Terms),
		fmt.Sprintf("   Postings:  %d", st.Postings),
		fmt.Sprintf("   Fields:    %d", st.Fields),
		fmt.Sprintf("   Memory:    %s", formatBytes(st.MemEstimate)),
	}

	return strings.Join(lines, "\n")
}

func (s *SearchScreen) renderInfo() string {
	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		" Search Management",
		"",
	}

	if s.Rebuilding {
		lines = append(lines, " Rebuilding index...")
	} else if s.RebuildErr != nil {
		lines = append(lines, fmt.Sprintf(" Rebuild failed: %s", s.RebuildErr))
	} else if s.Stats != nil && s.Stats.Documents > 0 {
		lines = append(lines, " Index loaded and ready.")
	}

	a := searchActions[s.Cursor]
	lines = append(lines, "", faint.Render(fmt.Sprintf(" %s", a.Desc)))

	return strings.Join(lines, "\n")
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
