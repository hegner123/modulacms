package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// 3/9 grid: left = token list, right = detail (top) + info (bottom)
var tokensGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Tokens"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Details"},
			{Height: 0.4, Title: "Info"},
		}},
	},
}

// TokensScreen implements Screen for the API tokens management page.
type TokensScreen struct {
	GridScreen
	TokensList []db.Tokens

	// RevealToken holds a newly created token value for one-time display.
	// Cleared when the user navigates away or creates another token.
	RevealToken string
}

// NewTokensScreen creates a TokensScreen with the given tokens data.
func NewTokensScreen(tokens []db.Tokens) *TokensScreen {
	cursorMax := len(tokens) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &TokensScreen{
		GridScreen: GridScreen{
			Grid:      tokensGrid,
			CursorMax: cursorMax,
		},
		TokensList: tokens,
	}
}

func (s *TokensScreen) PageIndex() PageIndex { return TOKENSPAGE }

// selectedToken returns the token at the current cursor, or nil.
func (s *TokensScreen) selectedToken() *db.Tokens {
	if len(s.TokensList) == 0 || s.Cursor >= len(s.TokensList) {
		return nil
	}
	return &s.TokensList[s.Cursor]
}

func (s *TokensScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// New token dialog
		if km.Matches(key, config.ActionNew) {
			return s, ShowCreateTokenDialogCmd()
		}

		// Delete token dialog
		if km.Matches(key, config.ActionDelete) {
			if tok := s.selectedToken(); tok != nil {
				label := tok.TokenType + " (" + tok.ID[:8] + "...)"
				return s, ShowDeleteTokenDialogCmd(tok.ID, label)
			}
		}

		// Common keys (quit, back, cursor)
		cursorMax := len(s.TokensList) - 1
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
	case TokensFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			list, err := d.ListTokens()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			data := make([]db.Tokens, 0)
			if list != nil {
				data = *list
			}
			return TokensFetchResultsMsg{Data: data}
		}

	// Fetch result messages
	case TokensFetchResultsMsg:
		s.TokensList = msg.Data
		s.Cursor = 0
		s.CursorMax = len(s.TokensList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		return s, LoadingStopCmd()

	// Token created — store reveal value and refresh list
	case TokenCreatedFromDialogMsg:
		s.RevealToken = msg.RawToken
		return s, TokensFetchCmd()

	// Token deleted — refresh list
	case TokenDeletedMsg:
		s.Cursor = 0
		s.RevealToken = ""
		return s, TokensFetchCmd()
	}

	return s, nil
}

func (s *TokensScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *TokensScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList(), TotalLines: len(s.TokensList), ScrollOffset: ClampScroll(s.Cursor, len(s.TokensList), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

// renderList renders the token list for the left panel.
func (s *TokensScreen) renderList() string {
	if len(s.TokensList) == 0 {
		return "(no tokens)"
	}

	lines := make([]string, 0, len(s.TokensList))
	for i, tok := range s.TokensList {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		status := "active"
		if tok.Revoked {
			status = "revoked"
		}
		// Show abbreviated ID + type
		idShort := tok.ID
		if len(idShort) > 8 {
			idShort = idShort[:8]
		}
		lines = append(lines, fmt.Sprintf("%s %s %s [%s]", cursor, tok.TokenType, idShort, status))
	}
	return strings.Join(lines, "\n")
}

// renderDetail renders the selected token's details.
func (s *TokensScreen) renderDetail() string {
	if len(s.TokensList) == 0 || s.Cursor >= len(s.TokensList) {
		return " No token selected"
	}

	tok := s.TokensList[s.Cursor]
	revoked := "No"
	if tok.Revoked {
		revoked = "Yes"
	}
	lines := []string{
		fmt.Sprintf(" ID        %s", tok.ID),
		fmt.Sprintf(" Type      %s", tok.TokenType),
		fmt.Sprintf(" User ID   %s", tok.UserID.String()),
		fmt.Sprintf(" Revoked   %s", revoked),
		"",
		fmt.Sprintf(" Issued    %s", tok.IssuedAt.String()),
		fmt.Sprintf(" Expires   %s", tok.ExpiresAt.String()),
	}

	// Show one-time reveal token if this was just created
	if s.RevealToken != "" {
		lines = append(lines, "")
		lines = append(lines, " --- NEW TOKEN (copy now, shown once) ---")
		lines = append(lines, fmt.Sprintf(" %s", s.RevealToken))
	}

	return strings.Join(lines, "\n")
}

// renderInfo renders the token summary.
func (s *TokensScreen) renderInfo() string {
	active := 0
	revoked := 0
	for _, tok := range s.TokensList {
		if tok.Revoked {
			revoked++
		} else {
			active++
		}
	}
	lines := []string{
		" Token Manager",
		"",
		fmt.Sprintf("   Total:   %d", len(s.TokensList)),
		fmt.Sprintf("   Active:  %d", active),
		fmt.Sprintf("   Revoked: %d", revoked),
	}
	return strings.Join(lines, "\n")
}
