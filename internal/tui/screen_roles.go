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

var rolesGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Roles"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.30, Title: "Details"},
			{Height: 0.70, Title: "Permissions"},
		}},
	},
}

// RolesScreen implements Screen for standalone role management.
type RolesScreen struct {
	GridScreen
	RolesList   []db.Roles
	Permissions []db.Permissions
	Assignments []db.RolePermissions
	// permCountCache maps role ID → number of assigned permissions.
	permCountCache map[types.RoleID]int
	// permLabelsCache maps role ID → sorted permission labels.
	permLabelsCache map[types.RoleID][]string
}

func NewRolesScreen(roles []db.Roles) *RolesScreen {
	cursorMax := len(roles) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &RolesScreen{
		GridScreen: GridScreen{
			Grid:      rolesGrid,
			CursorMax: cursorMax,
		},
		RolesList:       roles,
		permCountCache:  make(map[types.RoleID]int),
		permLabelsCache: make(map[types.RoleID][]string),
	}
}

func (s *RolesScreen) PageIndex() PageIndex { return ROLESPAGE }

func (s *RolesScreen) selectedRole() *db.Roles {
	if len(s.RolesList) == 0 || s.Cursor >= len(s.RolesList) {
		return nil
	}
	return &s.RolesList[s.Cursor]
}

func (s *RolesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		if km.Matches(key, config.ActionNew) {
			return s, ShowCreateRoleDialogCmd()
		}

		if km.Matches(key, config.ActionEdit) {
			if role := s.selectedRole(); role != nil {
				if role.SystemProtected {
					return s, ShowDialogCmd("Protected", "System-protected roles cannot be renamed.", false, DIALOGGENERIC)
				}
				return s, ShowEditRoleDialogCmd(*role)
			}
		}

		if km.Matches(key, config.ActionDelete) {
			if role := s.selectedRole(); role != nil {
				if role.SystemProtected {
					return s, ShowDialogCmd("Protected", "System-protected roles cannot be deleted.", false, DIALOGGENERIC)
				}
				return s, ShowDeleteRoleDialogCmd(role.RoleID, role.Label)
			}
		}

		cursorMax := len(s.RolesList) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		s.CursorMax = cursorMax
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	case RolesScreenFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			roles, err := d.ListRoles()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			perms, err := d.ListPermissions()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			assigns, err := d.ListRolePermissions()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			r := make([]db.Roles, 0)
			if roles != nil {
				r = *roles
			}
			p := make([]db.Permissions, 0)
			if perms != nil {
				p = *perms
			}
			a := make([]db.RolePermissions, 0)
			if assigns != nil {
				a = *assigns
			}
			return RolesScreenFetchResultsMsg{Roles: r, Permissions: p, Assignments: a}
		}

	case RolesScreenFetchResultsMsg:
		s.RolesList = msg.Roles
		s.Permissions = msg.Permissions
		s.Assignments = msg.Assignments
		s.Cursor = 0
		s.CursorMax = len(s.RolesList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		s.buildCaches()
		return s, LoadingStopCmd()

	case RoleCreatedFromDialogMsg, RoleUpdatedFromDialogMsg, RoleDeletedMsg:
		s.Cursor = 0
		return s, RolesScreenFetchCmd()
	}

	return s, nil
}

// buildCaches precomputes permission counts and labels per role.
func (s *RolesScreen) buildCaches() {
	// Build permissionID → label lookup
	permLabelMap := make(map[types.PermissionID]string, len(s.Permissions))
	for _, p := range s.Permissions {
		permLabelMap[p.PermissionID] = p.Label
	}

	s.permCountCache = make(map[types.RoleID]int, len(s.RolesList))
	s.permLabelsCache = make(map[types.RoleID][]string, len(s.RolesList))

	for _, a := range s.Assignments {
		s.permCountCache[a.RoleID]++
		if label, ok := permLabelMap[a.PermissionID]; ok {
			s.permLabelsCache[a.RoleID] = append(s.permLabelsCache[a.RoleID], label)
		}
	}

	for roleID, labels := range s.permLabelsCache {
		sort.Strings(labels)
		s.permLabelsCache[roleID] = labels
	}
}

func (s *RolesScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *RolesScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList(), TotalLines: len(s.RolesList), ScrollOffset: ClampScroll(s.Cursor, len(s.RolesList), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderPermissions()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *RolesScreen) renderList() string {
	if len(s.RolesList) == 0 {
		return "(no roles)"
	}
	lines := make([]string, 0, len(s.RolesList))
	for i, role := range s.RolesList {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		count := s.permCountCache[role.RoleID]
		protected := ""
		if role.SystemProtected {
			protected = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s (%d)", cursor, role.Label, protected, count))
	}
	return strings.Join(lines, "\n")
}

func (s *RolesScreen) renderDetail() string {
	role := s.selectedRole()
	if role == nil {
		return " No role selected"
	}

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	protected := "No"
	if role.SystemProtected {
		protected = accent.Render("Yes (cannot edit/delete)")
	}

	count := s.permCountCache[role.RoleID]

	lines := []string{
		fmt.Sprintf(" Label       %s", role.Label),
		fmt.Sprintf(" Protected   %s", protected),
		fmt.Sprintf(" Permissions %d", count),
		"",
		faint.Render(fmt.Sprintf(" ID          %s", role.RoleID)),
	}

	if role.SystemProtected {
		lines = append(lines, "", faint.Render(" * = system-protected"))
	}

	return strings.Join(lines, "\n")
}

func (s *RolesScreen) renderPermissions() string {
	role := s.selectedRole()
	if role == nil {
		return " No role selected"
	}

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	labels := s.permLabelsCache[role.RoleID]
	if len(labels) == 0 {
		return faint.Render(" No permissions assigned")
	}

	// Group by resource
	grouped := groupPermissionsByResource(labels)
	resources := make([]string, 0, len(grouped))
	for r := range grouped {
		resources = append(resources, r)
	}
	sort.Strings(resources)

	lines := make([]string, 0)
	for _, resource := range resources {
		ops := grouped[resource]
		sort.Strings(ops)
		lines = append(lines, accent.Render(fmt.Sprintf(" %s", resource)))
		lines = append(lines, fmt.Sprintf("   %s", strings.Join(ops, ", ")))
	}
	return strings.Join(lines, "\n")
}
