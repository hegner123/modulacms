package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Users grid: 2 columns
//
//	Col 0 (span 3): User list
//	Col 1 (span 9): Details (top), Permissions (bottom)
var usersGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Users"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.35, Title: "Details"},
			{Height: 0.25, Title: "OAuth"},
			{Height: 0.40, Title: "Permissions"},
		}},
	},
}

// UserPermissionsFetchedMsg delivers permission labels for a role.
type UserPermissionsFetchedMsg struct {
	RoleID types.RoleID
	Labels []string
	Err    error
}

// UserOauthFetchedMsg delivers OAuth connections for a user.
type UserOauthFetchedMsg struct {
	UserID types.UserID
	Data   []db.UserOauth
	Err    error
}

// UserOauthDeletedMsg is sent after an OAuth connection is unlinked.
type UserOauthDeletedMsg struct {
	UserID      types.UserID
	UserOauthID types.UserOauthID
}

// ShowUnlinkOauthDialogMsg triggers showing an unlink OAuth dialog.
type ShowUnlinkOauthDialogMsg struct {
	UserOauthID types.UserOauthID
	UserID      types.UserID
	Provider    string
}

// ShowUnlinkOauthDialogCmd creates a command to show the unlink dialog.
func ShowUnlinkOauthDialogCmd(oauthID types.UserOauthID, userID types.UserID, provider string) tea.Cmd {
	return func() tea.Msg {
		return ShowUnlinkOauthDialogMsg{UserOauthID: oauthID, UserID: userID, Provider: provider}
	}
}

// UnlinkOauthRequestMsg triggers OAuth connection deletion.
type UnlinkOauthRequestMsg struct {
	UserOauthID types.UserOauthID
	UserID      types.UserID
}

// UnlinkOauthCmd creates a command to unlink an OAuth connection.
func UnlinkOauthCmd(oauthID types.UserOauthID, userID types.UserID) tea.Cmd {
	return func() tea.Msg {
		return UnlinkOauthRequestMsg{UserOauthID: oauthID, UserID: userID}
	}
}

// UsersScreen implements Screen for the USERSADMIN page.
type UsersScreen struct {
	GridScreen
	UsersList       []db.UserWithRoleLabelRow
	RolesList       []db.Roles
	PermissionCache map[types.RoleID][]string       // role ID -> permission labels
	LastPermRoleID  types.RoleID                     // role ID of last permission fetch
	OauthCache      map[types.UserID][]db.UserOauth  // user ID -> OAuth connections
	LastOauthUserID types.UserID                     // user ID of last OAuth fetch
}

// NewUsersScreen creates a UsersScreen with the given users and roles data.
func NewUsersScreen(users []db.UserWithRoleLabelRow, roles []db.Roles) *UsersScreen {
	cursorMax := len(users) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &UsersScreen{
		GridScreen: GridScreen{
			Grid:      usersGrid,
			CursorMax: cursorMax,
		},
		UsersList:       users,
		RolesList:       roles,
		PermissionCache: make(map[types.RoleID][]string),
		OauthCache:      make(map[types.UserID][]db.UserOauth),
	}
}

func (s *UsersScreen) PageIndex() PageIndex { return USERSADMIN }

func (s *UsersScreen) updateCursorMax() {
	s.CursorMax = len(s.UsersList) - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.Cursor > s.CursorMax && s.CursorMax >= 0 {
		s.Cursor = s.CursorMax
	}
}

// selectedUser returns the user at the current cursor, or nil.
func (s *UsersScreen) selectedUser() *db.UserWithRoleLabelRow {
	if len(s.UsersList) == 0 || s.Cursor >= len(s.UsersList) {
		return nil
	}
	return &s.UsersList[s.Cursor]
}

// fetchPermissionsIfNeeded returns a command to fetch permissions for the
// selected user's role, or nil if already cached.
func (s *UsersScreen) fetchPermissionsIfNeeded(driver db.DbDriver) tea.Cmd {
	user := s.selectedUser()
	if user == nil || driver == nil {
		return nil
	}
	roleID := types.RoleID(user.Role)
	if roleID == s.LastPermRoleID {
		return nil
	}
	s.LastPermRoleID = roleID
	if _, ok := s.PermissionCache[roleID]; ok {
		return nil
	}
	return func() tea.Msg {
		labels, err := driver.ListPermissionLabelsByRoleID(roleID)
		if err != nil {
			return UserPermissionsFetchedMsg{RoleID: roleID, Err: err}
		}
		if labels == nil {
			return UserPermissionsFetchedMsg{RoleID: roleID, Labels: []string{}}
		}
		return UserPermissionsFetchedMsg{RoleID: roleID, Labels: *labels}
	}
}

// fetchOauthIfNeeded returns a command to fetch OAuth connections for the
// selected user, or nil if already cached.
func (s *UsersScreen) fetchOauthIfNeeded(driver db.DbDriver) tea.Cmd {
	user := s.selectedUser()
	if user == nil || driver == nil {
		return nil
	}
	userID := user.UserID
	if userID == s.LastOauthUserID {
		return nil
	}
	s.LastOauthUserID = userID
	if _, ok := s.OauthCache[userID]; ok {
		return nil
	}
	return func() tea.Msg {
		result, err := driver.GetUserOauthByUserId(types.NullableUserID{ID: userID, Valid: true})
		if err != nil {
			// No OAuth connection is not an error — return empty
			return UserOauthFetchedMsg{UserID: userID, Data: []db.UserOauth{}}
		}
		if result == nil {
			return UserOauthFetchedMsg{UserID: userID, Data: []db.UserOauth{}}
		}
		return UserOauthFetchedMsg{UserID: userID, Data: []db.UserOauth{*result}}
	}
}

func (s *UsersScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
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

		// Unlink OAuth connection
		if key == "u" {
			if user := s.selectedUser(); user != nil {
				if oauths, ok := s.OauthCache[user.UserID]; ok && len(oauths) > 0 {
					oa := oauths[0]
					return s, ShowUnlinkOauthDialogCmd(oa.UserOauthID, user.UserID, oa.OauthProvider)
				}
			}
		}

		// Delete user dialog
		if km.Matches(key, config.ActionDelete) {
			if len(s.UsersList) > 0 && s.Cursor < len(s.UsersList) {
				user := s.UsersList[s.Cursor]
				return s, ShowDeleteUserDialogCmd(user.UserID, user.Username)
			}
		}

		// Common keys (quit, back, cursor)
		prevCursor := s.Cursor
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			if s.Cursor != prevCursor {
				var cmds []tea.Cmd
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				if permCmd := s.fetchPermissionsIfNeeded(ctx.DB); permCmd != nil {
					cmds = append(cmds, permCmd)
				}
				if oauthCmd := s.fetchOauthIfNeeded(ctx.DB); oauthCmd != nil {
					cmds = append(cmds, oauthCmd)
				}
				if len(cmds) > 0 {
					return s, tea.Batch(cmds...)
				}
				return s, nil
			}
			return s, cmd
		}

	// OAuth connections fetched
	case UserOauthFetchedMsg:
		if msg.Err == nil {
			s.OauthCache[msg.UserID] = msg.Data
		}
		return s, nil

	// OAuth unlinked — invalidate cache and re-fetch
	case UserOauthDeletedMsg:
		delete(s.OauthCache, msg.UserID)
		s.LastOauthUserID = ""
		return s, s.fetchOauthIfNeeded(ctx.DB)

	// Permission labels fetched
	case UserPermissionsFetchedMsg:
		if msg.Err == nil {
			s.PermissionCache[msg.RoleID] = msg.Labels
		}
		return s, nil

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
		s.Cursor = 0
		s.updateCursorMax()
		s.LastPermRoleID = ""
		s.LastOauthUserID = ""
		cmds := []tea.Cmd{LoadingStopCmd()}
		if permCmd := s.fetchPermissionsIfNeeded(ctx.DB); permCmd != nil {
			cmds = append(cmds, permCmd)
		}
		if oauthCmd := s.fetchOauthIfNeeded(ctx.DB); oauthCmd != nil {
			cmds = append(cmds, oauthCmd)
		}
		return s, tea.Batch(cmds...)
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
		s.updateCursorMax()
		s.LastPermRoleID = ""
		return s, s.fetchPermissionsIfNeeded(ctx.DB)
	case RolesListSet:
		s.RolesList = msg.RolesList
		// Invalidate permission cache since roles may have changed
		s.PermissionCache = make(map[types.RoleID][]string)
		s.LastPermRoleID = ""
		return s, s.fetchPermissionsIfNeeded(ctx.DB)
	}

	return s, nil
}

func (s *UsersScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{"u", "unlink oauth"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *UsersScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderUsersList()},
		{Content: s.renderUserDetail()},
		{Content: s.renderOAuth()},
		{Content: s.renderPermissions()},
	}
	return s.RenderGrid(ctx, cells)
}

// ---------------------------------------------------------------------------
// Render helpers
// ---------------------------------------------------------------------------

func (s *UsersScreen) renderUsersList() string {
	if len(s.UsersList) == 0 {
		return " No users found"
	}

	lines := make([]string, 0, len(s.UsersList))
	for i, user := range s.UsersList {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s (%s)", cursor, user.Username, user.RoleLabel))
	}
	return strings.Join(lines, "\n")
}

func (s *UsersScreen) renderUserDetail() string {
	user := s.selectedUser()
	if user == nil {
		return " Select a user"
	}

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Username  %s", user.Username),
		fmt.Sprintf(" Name      %s", user.Name),
		fmt.Sprintf(" Email     %s", string(user.Email)),
		fmt.Sprintf(" Role      %s", user.RoleLabel),
		"",
		faint.Render(fmt.Sprintf(" ID        %s", user.UserID)),
		faint.Render(fmt.Sprintf(" Created   %s", user.DateCreated)),
		faint.Render(fmt.Sprintf(" Modified  %s", user.DateModified)),
	}
	return strings.Join(lines, "\n")
}

func (s *UsersScreen) renderOAuth() string {
	user := s.selectedUser()
	if user == nil {
		return " No user selected"
	}

	faint := lipgloss.NewStyle().Faint(true)
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	oauths, ok := s.OauthCache[user.UserID]
	if !ok {
		return faint.Render(" Loading OAuth connections...")
	}

	if len(oauths) == 0 {
		return faint.Render(" No OAuth connections")
	}

	lines := make([]string, 0)
	for _, oa := range oauths {
		lines = append(lines, accent.Render(fmt.Sprintf(" %s", oa.OauthProvider)))
		lines = append(lines, fmt.Sprintf("   ID      %s", oa.OauthProviderUserID))
		lines = append(lines, faint.Render(fmt.Sprintf("   Linked  %s", oa.DateCreated.String())))
	}
	return strings.Join(lines, "\n")
}

func (s *UsersScreen) renderPermissions() string {
	user := s.selectedUser()
	if user == nil {
		return " No user selected"
	}

	roleID := types.RoleID(user.Role)
	labels, ok := s.PermissionCache[roleID]

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		accent.Render(fmt.Sprintf(" Role: %s", user.RoleLabel)),
		fmt.Sprintf(" Users: %d", len(s.UsersList)),
		"",
	}

	if !ok {
		lines = append(lines, faint.Render(" Loading permissions..."))
		return strings.Join(lines, "\n")
	}

	if len(labels) == 0 {
		lines = append(lines, faint.Render(" No permissions assigned"))
		return strings.Join(lines, "\n")
	}

	// Group permissions by resource (prefix before ':')
	grouped := groupPermissionsByResource(labels)

	// Sort resource names for stable output
	resources := make([]string, 0, len(grouped))
	for r := range grouped {
		resources = append(resources, r)
	}
	sort.Strings(resources)

	for _, resource := range resources {
		ops := grouped[resource]
		sort.Strings(ops)
		lines = append(lines, accent.Render(fmt.Sprintf(" %s", resource)))
		lines = append(lines, fmt.Sprintf("   %s", strings.Join(ops, ", ")))
	}

	return strings.Join(lines, "\n")
}

// groupPermissionsByResource groups "resource:operation" labels by resource.
func groupPermissionsByResource(labels []string) map[string][]string {
	grouped := make(map[string][]string)
	for _, label := range labels {
		parts := strings.SplitN(label, ":", 2)
		if len(parts) == 2 {
			grouped[parts[0]] = append(grouped[parts[0]], parts[1])
		} else {
			grouped["other"] = append(grouped["other"], label)
		}
	}
	return grouped
}
