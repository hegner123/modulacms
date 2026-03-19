package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// 3/9 grid: left = session list, right = detail (top) + info (bottom)
var sessionsGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Sessions"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Details"},
			{Height: 0.4, Title: "Info"},
		}},
	},
}

// SessionsScreen implements Screen for the active sessions management page.
type SessionsScreen struct {
	GridScreen
	SessionsList []db.Sessions
	// UserNames maps UserID → username for display.
	UserNames map[string]string
	// CurrentSessionID is the session of the logged-in user, shown as indicator.
	CurrentSessionID types.SessionID
}

// NewSessionsScreen creates a SessionsScreen with the given data.
func NewSessionsScreen(sessions []db.Sessions) *SessionsScreen {
	cursorMax := len(sessions) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &SessionsScreen{
		GridScreen: GridScreen{
			Grid:      sessionsGrid,
			CursorMax: cursorMax,
		},
		SessionsList: sessions,
		UserNames:    make(map[string]string),
	}
}

func (s *SessionsScreen) PageIndex() PageIndex { return SESSIONSPAGE }

// selectedSession returns the session at the current cursor, or nil.
func (s *SessionsScreen) selectedSession() *db.Sessions {
	if len(s.SessionsList) == 0 || s.Cursor >= len(s.SessionsList) {
		return nil
	}
	return &s.SessionsList[s.Cursor]
}

func (s *SessionsScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Delete (revoke) session
		if km.Matches(key, config.ActionDelete) {
			if sess := s.selectedSession(); sess != nil {
				label := s.sessionLabel(sess)
				return s, ShowDeleteSessionDialogCmd(sess.SessionID, label)
			}
		}

		// Common keys (quit, back, cursor)
		cursorMax := len(s.SessionsList) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		s.CursorMax = cursorMax
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request
	case SessionsFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			list, err := d.ListSessions()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			data := make([]db.Sessions, 0)
			if list != nil {
				data = *list
			}
			// Build user name map
			names := make(map[string]string)
			for _, sess := range data {
				if sess.UserID.Valid {
					uid := string(sess.UserID.ID)
					if _, ok := names[uid]; !ok {
						user, err := d.GetUser(sess.UserID.ID)
						if err == nil && user != nil {
							names[uid] = user.Username
						} else {
							names[uid] = uid[:8] + "..."
						}
					}
				}
			}
			return sessionsFetchResultWithNames{Data: data, Names: names}
		}

	case sessionsFetchResultWithNames:
		s.SessionsList = msg.Data
		s.UserNames = msg.Names
		s.Cursor = 0
		s.CursorMax = len(s.SessionsList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		// Detect current user's session
		for _, sess := range s.SessionsList {
			if sess.UserID.Valid && sess.UserID.ID == ctx.UserID {
				s.CurrentSessionID = sess.SessionID
				break
			}
		}
		return s, LoadingStopCmd()

	// Session deleted — refresh list
	case SessionDeletedMsg:
		s.Cursor = 0
		return s, SessionsFetchCmd()
	}

	return s, nil
}

// sessionsFetchResultWithNames carries sessions + resolved user names.
type sessionsFetchResultWithNames struct {
	Data  []db.Sessions
	Names map[string]string
}

func (s *SessionsScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionDelete), "revoke"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *SessionsScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList(), TotalLines: len(s.SessionsList), ScrollOffset: ClampScroll(s.Cursor, len(s.SessionsList), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

// sessionLabel returns a short display label for a session.
func (s *SessionsScreen) sessionLabel(sess *db.Sessions) string {
	name := s.resolveUser(sess)
	ip := "unknown"
	if sess.IpAddress.Valid {
		ip = sess.IpAddress.String
	}
	return fmt.Sprintf("%s @ %s", name, ip)
}

// resolveUser returns the username for a session's user ID.
func (s *SessionsScreen) resolveUser(sess *db.Sessions) string {
	if !sess.UserID.Valid {
		return "(anonymous)"
	}
	uid := string(sess.UserID.ID)
	if name, ok := s.UserNames[uid]; ok {
		return name
	}
	if len(uid) > 8 {
		return uid[:8] + "..."
	}
	return uid
}

// renderList renders the session list for the left panel.
func (s *SessionsScreen) renderList() string {
	if len(s.SessionsList) == 0 {
		return "(no sessions)"
	}

	lines := make([]string, 0, len(s.SessionsList))
	for i, sess := range s.SessionsList {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		name := s.resolveUser(&sess)
		current := ""
		if sess.SessionID == s.CurrentSessionID {
			current = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, name, current))
	}
	return strings.Join(lines, "\n")
}

// renderDetail renders the selected session's details.
func (s *SessionsScreen) renderDetail() string {
	if len(s.SessionsList) == 0 || s.Cursor >= len(s.SessionsList) {
		return " No session selected"
	}

	sess := s.SessionsList[s.Cursor]
	name := s.resolveUser(&sess)
	ip := "(none)"
	if sess.IpAddress.Valid {
		ip = sess.IpAddress.String
	}
	ua := "(none)"
	if sess.UserAgent.Valid {
		ua = sess.UserAgent.String
	}
	// Truncate long user agents
	if len(ua) > 70 {
		ua = ua[:67] + "..."
	}

	current := ""
	if sess.SessionID == s.CurrentSessionID {
		current = "  (current session)"
	}

	lines := []string{
		fmt.Sprintf(" Session  %s%s", sess.SessionID, current),
		fmt.Sprintf(" User     %s", name),
		fmt.Sprintf(" IP       %s", ip),
		fmt.Sprintf(" Agent    %s", ua),
		"",
		fmt.Sprintf(" Created  %s", sess.DateCreated.String()),
		fmt.Sprintf(" Expires  %s", sess.ExpiresAt.String()),
		fmt.Sprintf(" Active   %s", sess.LastAccess.String()),
	}

	return strings.Join(lines, "\n")
}

// renderInfo renders the session summary.
func (s *SessionsScreen) renderInfo() string {
	// Count unique users
	users := make(map[string]bool)
	for _, sess := range s.SessionsList {
		if sess.UserID.Valid {
			users[string(sess.UserID.ID)] = true
		}
	}
	lines := []string{
		" Session Manager",
		"",
		fmt.Sprintf("   Active:  %d", len(s.SessionsList)),
		fmt.Sprintf("   Users:   %d", len(users)),
	}
	if s.CurrentSessionID != "" {
		lines = append(lines, "", "   * = your session")
	}
	return strings.Join(lines, "\n")
}
