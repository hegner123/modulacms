package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// 3/9 grid: left = webhook list, right = detail (top) + info (bottom)
var webhooksGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Webhooks"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Details"},
			{Height: 0.4, Title: "Info"},
		}},
	},
}

// WebhooksScreen implements Screen for the webhooks list page.
type WebhooksScreen struct {
	GridScreen
	WebhooksList []db.Webhook
}

// NewWebhooksScreen creates a WebhooksScreen with the given webhooks data.
func NewWebhooksScreen(webhooks []db.Webhook) *WebhooksScreen {
	cursorMax := len(webhooks) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &WebhooksScreen{
		GridScreen: GridScreen{
			Grid:      webhooksGrid,
			CursorMax: cursorMax,
		},
		WebhooksList: webhooks,
	}
}

func (s *WebhooksScreen) PageIndex() PageIndex { return WEBHOOKSPAGE }

func (s *WebhooksScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		cursorMax := len(s.WebhooksList) - 1
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
	case WebhooksFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			list, err := d.ListWebhooks()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			data := make([]db.Webhook, 0)
			if list != nil {
				data = *list
			}
			return WebhooksFetchResultsMsg{Data: data}
		}

	// Fetch result messages
	case WebhooksFetchResultsMsg:
		s.WebhooksList = msg.Data
		s.Cursor = 0
		s.CursorMax = len(s.WebhooksList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		return s, LoadingStopCmd()
	}

	return s, nil
}

func (s *WebhooksScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *WebhooksScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList(), TotalLines: len(s.WebhooksList), ScrollOffset: ClampScroll(s.Cursor, len(s.WebhooksList), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

// renderList renders the webhook list for the left panel.
func (s *WebhooksScreen) renderList() string {
	if len(s.WebhooksList) == 0 {
		return "(no webhooks)"
	}

	lines := make([]string, 0, len(s.WebhooksList))
	for i, wh := range s.WebhooksList {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		status := "off"
		if wh.IsActive {
			status = "on"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]", cursor, wh.Name, status))
	}
	return strings.Join(lines, "\n")
}

// renderDetail renders the selected webhook details.
func (s *WebhooksScreen) renderDetail() string {
	if len(s.WebhooksList) == 0 || s.Cursor >= len(s.WebhooksList) {
		return " No webhook selected"
	}

	wh := s.WebhooksList[s.Cursor]
	active := "No"
	if wh.IsActive {
		active = "Yes"
	}
	lines := []string{
		fmt.Sprintf(" Name     %s", wh.Name),
		fmt.Sprintf(" URL      %s", wh.URL),
		fmt.Sprintf(" Active   %s", active),
		fmt.Sprintf(" Events   %s", strings.Join(wh.Events, ", ")),
		"",
		fmt.Sprintf(" Created  %s", wh.DateCreated.String()),
		fmt.Sprintf(" Modified %s", wh.DateModified.String()),
	}
	return strings.Join(lines, "\n")
}

// renderInfo renders the webhook summary.
func (s *WebhooksScreen) renderInfo() string {
	active := 0
	for _, wh := range s.WebhooksList {
		if wh.IsActive {
			active++
		}
	}
	lines := []string{
		" Webhook Manager",
		"",
		fmt.Sprintf("   Total:  %d", len(s.WebhooksList)),
		fmt.Sprintf("   Active: %d", active),
	}
	return strings.Join(lines, "\n")
}
