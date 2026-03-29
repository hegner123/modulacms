package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

var auditGrid = Grid{
	Columns: []GridColumn{
		{Span: 4, Cells: []GridCell{
			{Height: 1.0, Title: "Events"},
		}},
		{Span: 8, Cells: []GridCell{
			{Height: 0.35, Title: "Details"},
			{Height: 0.65, Title: "Values"},
		}},
	},
}

const auditPageSize int64 = 50

// AuditFetchMsg requests fetching audit log events.
type AuditFetchMsg struct{}

// AuditFetchResultsMsg returns fetched change events and total count.
type AuditFetchResultsMsg struct {
	Events []db.ChangeEvent
	Total  int64
}

// AuditScreen implements Screen for the audit log viewer.
type AuditScreen struct {
	GridScreen
	Events []db.ChangeEvent
	Total  int64
	Page   int64 // 0-indexed page number
}

func NewAuditScreen() *AuditScreen {
	return &AuditScreen{
		GridScreen: GridScreen{
			Grid: auditGrid,
		},
	}
}

func (s *AuditScreen) PageIndex() PageIndex { return AUDITPAGE }

func (s *AuditScreen) selectedEvent() *db.ChangeEvent {
	if len(s.Events) == 0 || s.Cursor >= len(s.Events) {
		return nil
	}
	return &s.Events[s.Cursor]
}

func (s *AuditScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Next/prev page
		if key == "]" && (s.Page+1)*auditPageSize < s.Total {
			s.Page++
			s.Cursor = 0
			return s, s.fetchPage(ctx)
		}
		if key == "[" && s.Page > 0 {
			s.Page--
			s.Cursor = 0
			return s, s.fetchPage(ctx)
		}

		cursorMax := len(s.Events) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		s.CursorMax = cursorMax
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	case AuditFetchMsg:
		return s, s.fetchPage(ctx)

	case AuditFetchResultsMsg:
		s.Events = msg.Events
		s.Total = msg.Total
		s.CursorMax = len(s.Events) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		return s, LoadingStopCmd()
	}

	return s, nil
}

func (s *AuditScreen) fetchPage(ctx AppContext) tea.Cmd {
	d := ctx.DB
	if d == nil {
		return func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
	}
	page := s.Page
	return func() tea.Msg {
		total, err := d.CountChangeEvents()
		if err != nil {
			return FetchErrMsg{Error: err}
		}
		t := int64(0)
		if total != nil {
			t = *total
		}
		events, err := d.ListChangeEvents(db.ListChangeEventsParams{
			Limit:  auditPageSize,
			Offset: page * auditPageSize,
		})
		if err != nil {
			return FetchErrMsg{Error: err}
		}
		data := make([]db.ChangeEvent, 0)
		if events != nil {
			data = *events
		}
		return AuditFetchResultsMsg{Events: data, Total: t}
	}
}

func (s *AuditScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{"[/]", "page"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

func (s *AuditScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList(), TotalLines: len(s.Events), ScrollOffset: ClampScroll(s.Cursor, len(s.Events), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderValues()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *AuditScreen) renderList() string {
	if len(s.Events) == 0 {
		return "(no events)"
	}

	totalPages := (s.Total + auditPageSize - 1) / auditPageSize
	header := fmt.Sprintf(" Page %d/%d (%d total)\n", s.Page+1, totalPages, s.Total)

	lines := make([]string, 0, len(s.Events)+1)
	lines = append(lines, header)
	for i, ev := range s.Events {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		ts := ev.WallTimestamp.String()
		if len(ts) > 16 {
			ts = ts[11:16] // HH:MM
		}
		lines = append(lines, fmt.Sprintf("%s %s %s %s", cursor, ts, ev.Operation, ev.TableName))
	}
	return strings.Join(lines, "\n")
}

func (s *AuditScreen) renderDetail() string {
	ev := s.selectedEvent()
	if ev == nil {
		return " No event selected"
	}

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	userStr := "(system)"
	if ev.UserID.Valid {
		uid := string(ev.UserID.ID)
		if len(uid) > 12 {
			uid = uid[:12] + "..."
		}
		userStr = uid
	}

	ip := ""
	if ev.IP.Valid {
		ip = ev.IP.String
	}

	lines := []string{
		accent.Render(fmt.Sprintf(" %s %s", ev.Operation, ev.TableName)),
		fmt.Sprintf(" Record   %s", ev.RecordID),
		fmt.Sprintf(" User     %s", userStr),
		fmt.Sprintf(" Time     %s", ev.WallTimestamp.String()),
	}
	if ip != "" {
		lines = append(lines, fmt.Sprintf(" IP       %s", ip))
	}
	lines = append(lines, faint.Render(fmt.Sprintf(" Event    %s", ev.EventID)))

	return strings.Join(lines, "\n")
}

func (s *AuditScreen) renderValues() string {
	ev := s.selectedEvent()
	if ev == nil {
		return " No event selected"
	}

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	var lines []string

	if !ev.OldValues.IsZero() {
		lines = append(lines, accent.Render(" Old Values:"))
		lines = append(lines, formatJSONCompact(ev.OldValues.String(), " ")...)
		lines = append(lines, "")
	}

	if !ev.NewValues.IsZero() {
		lines = append(lines, accent.Render(" New Values:"))
		lines = append(lines, formatJSONCompact(ev.NewValues.String(), " ")...)
	}

	if len(lines) == 0 {
		return " (no values recorded)"
	}

	return strings.Join(lines, "\n")
}

// formatJSONCompact pretty-prints JSON with indentation, prefixed with a margin.
func formatJSONCompact(raw string, prefix string) []string {
	if raw == "" || raw == "null" {
		return []string{prefix + "(null)"}
	}

	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		// Not valid JSON — show raw
		return []string{prefix + raw}
	}

	pretty, err := json.MarshalIndent(parsed, prefix, "  ")
	if err != nil {
		return []string{prefix + raw}
	}

	return strings.Split(string(pretty), "\n")
}
