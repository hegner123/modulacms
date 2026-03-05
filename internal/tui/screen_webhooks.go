package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// WebhooksScreen implements Screen for the webhooks list page.
type WebhooksScreen struct {
	Cursor       int
	PanelFocus   FocusPanel
	WebhooksList []db.Webhook
}

// NewWebhooksScreen creates a WebhooksScreen with the given webhooks data.
func NewWebhooksScreen(webhooks []db.Webhook) *WebhooksScreen {
	return &WebhooksScreen{
		Cursor:       0,
		PanelFocus:   TreePanel,
		WebhooksList: webhooks,
	}
}

func (s *WebhooksScreen) PageIndex() PageIndex { return WEBHOOKSPAGE }

func (s *WebhooksScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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
		cursorMax := len(s.WebhooksList) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, cursorMax)
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
	left := s.renderList()
	center := s.renderDetail()
	right := s.renderInfo()

	layout := layoutForPage(WEBHOOKSPAGE)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	listLen := len(s.WebhooksList)

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

// renderDetail renders the selected webhook details for the center panel.
func (s *WebhooksScreen) renderDetail() string {
	if len(s.WebhooksList) == 0 || s.Cursor >= len(s.WebhooksList) {
		return "No webhook selected"
	}

	wh := s.WebhooksList[s.Cursor]
	active := "No"
	if wh.IsActive {
		active = "Yes"
	}
	lines := []string{
		fmt.Sprintf("Name:     %s", wh.Name),
		fmt.Sprintf("URL:      %s", wh.URL),
		fmt.Sprintf("Active:   %s", active),
		fmt.Sprintf("Events:   %s", strings.Join(wh.Events, ", ")),
		"",
		fmt.Sprintf("Created:  %s", wh.DateCreated.String()),
		fmt.Sprintf("Modified: %s", wh.DateModified.String()),
	}
	return strings.Join(lines, "\n")
}

// renderInfo renders the webhook summary for the right panel.
func (s *WebhooksScreen) renderInfo() string {
	active := 0
	for _, wh := range s.WebhooksList {
		if wh.IsActive {
			active++
		}
	}
	lines := []string{
		"Webhook Manager",
		"",
		fmt.Sprintf("  Total:  %d", len(s.WebhooksList)),
		fmt.Sprintf("  Active: %d", active),
	}
	return strings.Join(lines, "\n")
}
