package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// UsersScreen implements Screen for the USERSADMIN page.
type UsersScreen struct {
	Cursor     int
	CursorMax  int
	PanelFocus FocusPanel
	UsersList  []db.UserWithRoleLabelRow
	RolesList  []db.Roles
}

// NewUsersScreen creates a UsersScreen with the given users and roles data.
func NewUsersScreen(users []db.UserWithRoleLabelRow, roles []db.Roles) *UsersScreen {
	cursorMax := len(users) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &UsersScreen{
		Cursor:     0,
		CursorMax:  cursorMax,
		PanelFocus: TreePanel,
		UsersList:  users,
		RolesList:  roles,
	}
}

func (s *UsersScreen) PageIndex() PageIndex { return USERSADMIN }

func (s *UsersScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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

		// New user dialog
		if km.Matches(key, config.ActionNew) {
			return s, ShowCreateUserDialogCmd(s.RolesList)
		}

		// Edit user dialog
		if km.Matches(key, config.ActionEdit) {
			if len(s.UsersList) > 0 && s.Cursor < len(s.UsersList) {
				return s, ShowEditUserDialogCmd(s.UsersList[s.Cursor], s.RolesList)
			}
		}

		// Delete user dialog
		if km.Matches(key, config.ActionDelete) {
			if len(s.UsersList) > 0 && s.Cursor < len(s.UsersList) {
				user := s.UsersList[s.Cursor]
				return s, ShowDeleteUserDialogCmd(user.UserID, user.Username)
			}
		}

		// Common keys (quit, back, cursor) - LAST
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case UsersFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			users, err := d.ListUsersWithRoleLabel()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if users == nil {
				return UsersFetchResultsMsg{Data: []db.UserWithRoleLabelRow{}}
			}
			return UsersFetchResultsMsg{Data: *users}
		}
	case UsersFetchResultsMsg:
		s.UsersList = msg.Data
		s.CursorMax = len(s.UsersList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		s.Cursor = 0
		return s, LoadingStopCmd()
	case RolesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			roles, err := d.ListRoles()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if roles == nil {
				return RolesFetchResultsMsg{Data: []db.Roles{}}
			}
			return RolesFetchResultsMsg{Data: *roles}
		}
	case RolesFetchResultsMsg:
		s.RolesList = msg.Data
		return s, nil

	// Data refresh messages (from CMS operations)
	case UsersListSet:
		s.UsersList = msg.UsersList
		s.CursorMax = len(s.UsersList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		if s.Cursor > s.CursorMax {
			s.Cursor = s.CursorMax
		}
		return s, nil

	case RolesListSet:
		s.RolesList = msg.RolesList
		return s, nil
	}

	return s, nil
}

func (s *UsersScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *UsersScreen) View(ctx AppContext) string {
	left := s.renderUsersList()
	center := s.renderUserDetail()
	right := s.renderUserPermissions()

	layout := layoutForPage(USERSADMIN)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	listLen := len(s.UsersList)

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

func (s *UsersScreen) renderUsersList() string {
	if len(s.UsersList) == 0 {
		return "No users found"
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary)

	var lines []string
	for i, user := range s.UsersList {
		prefix := "  "
		style := normalStyle
		if i == s.Cursor {
			prefix = "> "
			style = selectedStyle
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%s (%s)", prefix, user.Username, user.RoleLabel)))
	}

	return strings.Join(lines, "\n")
}

func (s *UsersScreen) renderUserDetail() string {
	if len(s.UsersList) == 0 || s.Cursor >= len(s.UsersList) {
		return "Select a user"
	}

	user := s.UsersList[s.Cursor]
	labelStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent).
		Bold(true)

	lines := []string{
		labelStyle.Render("User Details"),
		"",
		fmt.Sprintf("  Username:  %s", user.Username),
		fmt.Sprintf("  Name:      %s", user.Name),
		fmt.Sprintf("  Email:     %s", user.Email),
		fmt.Sprintf("  Role:      %s", user.RoleLabel),
		"",
		fmt.Sprintf("  ID:        %s", user.UserID),
		fmt.Sprintf("  Created:   %s", user.DateCreated),
		fmt.Sprintf("  Modified:  %s", user.DateModified),
	}

	return strings.Join(lines, "\n")
}

func (s *UsersScreen) renderUserPermissions() string {
	if len(s.UsersList) == 0 || s.Cursor >= len(s.UsersList) {
		return "Permissions\n\n  (none)"
	}

	user := s.UsersList[s.Cursor]
	lines := []string{
		"Permissions",
		"",
		fmt.Sprintf("  Role: %s", user.RoleLabel),
		"",
		fmt.Sprintf("  Users: %d", len(s.UsersList)),
	}

	return strings.Join(lines, "\n")
}
