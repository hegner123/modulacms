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
			{Height: 0.30, Title: "Details"},
			{Height: 0.30, Title: "Connections"},
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

// UserSshKeysFetchedMsg delivers SSH keys for a user.
type UserSshKeysFetchedMsg struct {
	UserID types.UserID
	Data   []db.UserSshKeys
	Err    error
}

// UserSshKeyDeletedMsg is sent after an SSH key is removed.
type UserSshKeyDeletedMsg struct {
	UserID   types.UserID
	SshKeyID string
}

// ShowDeleteSshKeyDialogMsg triggers showing a delete SSH key dialog.
type ShowDeleteSshKeyDialogMsg struct {
	SshKeyID    string
	UserID      types.UserID
	Fingerprint string
}

// ShowDeleteSshKeyDialogCmd creates a command to show the delete dialog.
func ShowDeleteSshKeyDialogCmd(sshKeyID string, userID types.UserID, fingerprint string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteSshKeyDialogMsg{SshKeyID: sshKeyID, UserID: userID, Fingerprint: fingerprint}
	}
}

// DeleteSshKeyRequestMsg triggers SSH key deletion.
type DeleteSshKeyRequestMsg struct {
	SshKeyID string
	UserID   types.UserID
}

// DeleteSshKeyCmd creates a command to delete an SSH key.
func DeleteSshKeyCmd(sshKeyID string, userID types.UserID) tea.Cmd {
	return func() tea.Msg {
		return DeleteSshKeyRequestMsg{SshKeyID: sshKeyID, UserID: userID}
	}
}

// UsersScreen implements Screen for the USERSADMIN page.
type UsersScreen struct {
	GridScreen
	UsersList        []db.UserWithRoleLabelRow
	RolesList        []db.Roles
	PermissionCache  map[types.RoleID][]string          // role ID -> permission labels
	LastPermRoleID   types.RoleID                        // role ID of last permission fetch
	OauthCache       map[types.UserID][]db.UserOauth     // user ID -> OAuth connections
	LastOauthUserID  types.UserID                        // user ID of last OAuth fetch
	SshKeyCache      map[types.UserID][]db.UserSshKeys   // user ID -> SSH keys
	LastSshKeyUserID types.UserID                        // user ID of last SSH key fetch
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
		SshKeyCache:     make(map[types.UserID][]db.UserSshKeys),
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

// fetchSshKeysIfNeeded returns a command to fetch SSH keys for the
// selected user, or nil if already cached.
func (s *UsersScreen) fetchSshKeysIfNeeded(driver db.DbDriver) tea.Cmd {
	user := s.selectedUser()
	if user == nil || driver == nil {
		return nil
	}
	userID := user.UserID
	if userID == s.LastSshKeyUserID {
		return nil
	}
	s.LastSshKeyUserID = userID
	if _, ok := s.SshKeyCache[userID]; ok {
		return nil
	}
	return func() tea.Msg {
		result, err := driver.ListUserSshKeys(types.NullableUserID{ID: userID, Valid: true})
		if err != nil {
			return UserSshKeysFetchedMsg{UserID: userID, Data: []db.UserSshKeys{}}
		}
		if result == nil {
			return UserSshKeysFetchedMsg{UserID: userID, Data: []db.UserSshKeys{}}
		}
		return UserSshKeysFetchedMsg{UserID: userID, Data: *result}
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

		// Remove SSH key
		if key == "s" {
			if user := s.selectedUser(); user != nil {
				if keys, ok := s.SshKeyCache[user.UserID]; ok && len(keys) > 0 {
					k := keys[0]
					return s, ShowDeleteSshKeyDialogCmd(k.SshKeyID, user.UserID, k.Fingerprint)
				}
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
				if sshCmd := s.fetchSshKeysIfNeeded(ctx.DB); sshCmd != nil {
					cmds = append(cmds, sshCmd)
				}
				if len(cmds) > 0 {
					return s, tea.Batch(cmds...)
				}
				return s, nil
			}
			return s, cmd
		}

	// SSH keys fetched
	case UserSshKeysFetchedMsg:
		if msg.Err == nil {
			s.SshKeyCache[msg.UserID] = msg.Data
		}
		return s, nil

	// SSH key deleted — invalidate cache and re-fetch
	case UserSshKeyDeletedMsg:
		delete(s.SshKeyCache, msg.UserID)
		s.LastSshKeyUserID = ""
		return s, s.fetchSshKeysIfNeeded(ctx.DB)

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
		s.LastSshKeyUserID = ""
		cmds := []tea.Cmd{LoadingStopCmd()}
		if permCmd := s.fetchPermissionsIfNeeded(ctx.DB); permCmd != nil {
			cmds = append(cmds, permCmd)
		}
		if oauthCmd := s.fetchOauthIfNeeded(ctx.DB); oauthCmd != nil {
			cmds = append(cmds, oauthCmd)
		}
		if sshCmd := s.fetchSshKeysIfNeeded(ctx.DB); sshCmd != nil {
			cmds = append(cmds, sshCmd)
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
		{"s", "rm ssh key"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *UsersScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderUsersList()},
		{Content: s.renderUserDetail()},
		{Content: s.renderConnections()},
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

func (s *UsersScreen) renderConnections() string {
	user := s.selectedUser()
	if user == nil {
		return " No user selected"
	}

	faint := lipgloss.NewStyle().Faint(true)
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	var lines []string

	// SSH Keys
	sshKeys, sshOk := s.SshKeyCache[user.UserID]
	if !sshOk {
		lines = append(lines, faint.Render(" Loading SSH keys..."))
	} else if len(sshKeys) == 0 {
		lines = append(lines, faint.Render(" No SSH keys"))
	} else {
		for _, k := range sshKeys {
			label := k.Label
			if label == "" {
				label = k.KeyType
			}
			lines = append(lines, accent.Render(fmt.Sprintf(" SSH: %s", label)))
			fp := k.Fingerprint
			if len(fp) > 40 {
				fp = fp[:40] + "..."
			}
			lines = append(lines, fmt.Sprintf("   %s", fp))
		}
	}

	// OAuth
	oauths, oauthOk := s.OauthCache[user.UserID]
	if !oauthOk {
		lines = append(lines, faint.Render(" Loading OAuth..."))
	} else if len(oauths) > 0 {
		for _, oa := range oauths {
			lines = append(lines, accent.Render(fmt.Sprintf(" OAuth: %s", oa.OauthProvider)))
			lines = append(lines, fmt.Sprintf("   %s", oa.OauthProviderUserID))
		}
	}

	if len(lines) == 0 {
		return faint.Render(" No connections")
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
