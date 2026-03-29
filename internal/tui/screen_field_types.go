package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// Field types grid: 2 columns
//
//	Col 0 (span 3): Field type list
//	Col 1 (span 9): Details (top), Info (bottom)
var fieldTypesGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Field Types"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.45, Title: "Details"},
			{Height: 0.55, Title: "Info"},
		}},
	},
}

// FieldTypesScreen implements Screen for both FIELDTYPES and ADMINFIELDTYPES pages.
// When AdminMode is true, it operates on admin field types; otherwise regular field types.
type FieldTypesScreen struct {
	GridScreen
	AdminMode       bool
	FieldTypes      []db.FieldTypes
	AdminFieldTypes []db.AdminFieldTypes
}

// NewFieldTypesScreen creates a FieldTypesScreen for regular or admin mode.
func NewFieldTypesScreen(adminMode bool, fieldTypes []db.FieldTypes, adminFieldTypes []db.AdminFieldTypes) *FieldTypesScreen {
	listLen := len(fieldTypes)
	if adminMode {
		listLen = len(adminFieldTypes)
	}
	cursorMax := listLen - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &FieldTypesScreen{
		GridScreen: GridScreen{
			Grid:      fieldTypesGrid,
			CursorMax: cursorMax,
		},
		AdminMode:       adminMode,
		FieldTypes:      fieldTypes,
		AdminFieldTypes: adminFieldTypes,
	}
}

func (s *FieldTypesScreen) PageIndex() PageIndex {
	if s.AdminMode {
		return ADMINFIELDTYPES
	}
	return FIELDTYPES
}

func (s *FieldTypesScreen) listLen() int {
	if s.AdminMode {
		return len(s.AdminFieldTypes)
	}
	return len(s.FieldTypes)
}

func (s *FieldTypesScreen) updateCursorMax() {
	s.CursorMax = s.listLen() - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.Cursor > s.CursorMax && s.CursorMax >= 0 {
		s.Cursor = s.CursorMax
	}
}

func (s *FieldTypesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// New field type
		if km.Matches(key, config.ActionNew) {
			if s.AdminMode {
				return s, ShowRouteFormDialogCmd(FORMDIALOGCREATEADMINFIELDTYPE, "New Admin Field Type")
			}
			return s, ShowRouteFormDialogCmd(FORMDIALOGCREATEFIELDTYPE, "New Field Type")
		}

		// Edit field type
		if km.Matches(key, config.ActionEdit) {
			if s.AdminMode {
				if len(s.AdminFieldTypes) > 0 && s.Cursor < len(s.AdminFieldTypes) {
					return s, ShowEditAdminFieldTypeDialogCmd(s.AdminFieldTypes[s.Cursor])
				}
			} else {
				if len(s.FieldTypes) > 0 && s.Cursor < len(s.FieldTypes) {
					return s, ShowEditFieldTypeDialogCmd(s.FieldTypes[s.Cursor])
				}
			}
			return s, nil
		}

		// Delete field type
		if km.Matches(key, config.ActionDelete) {
			if s.AdminMode {
				if len(s.AdminFieldTypes) > 0 && s.Cursor < len(s.AdminFieldTypes) {
					ft := s.AdminFieldTypes[s.Cursor]
					return s, ShowDeleteAdminFieldTypeDialogCmd(ft.AdminFieldTypeID, ft.Label)
				}
			} else {
				if len(s.FieldTypes) > 0 && s.Cursor < len(s.FieldTypes) {
					ft := s.FieldTypes[s.Cursor]
					return s, ShowDeleteFieldTypeDialogCmd(ft.FieldTypeID, ft.Label)
				}
			}
			return s, nil
		}

		// Common keys (quit, back, cursor)
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case FieldTypesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			fieldTypes, err := d.ListFieldTypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if fieldTypes == nil {
				return FieldTypesFetchResultsMsg{Data: []db.FieldTypes{}}
			}
			return FieldTypesFetchResultsMsg{Data: *fieldTypes}
		}
	case FieldTypesFetchResultsMsg:
		s.FieldTypes = msg.Data
		s.Cursor = 0
		s.updateCursorMax()
		return s, LoadingStopCmd()
	case AdminFieldTypesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			adminFieldTypes, err := d.ListAdminFieldTypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if adminFieldTypes == nil {
				return AdminFieldTypesFetchResultsMsg{Data: []db.AdminFieldTypes{}}
			}
			return AdminFieldTypesFetchResultsMsg{Data: *adminFieldTypes}
		}
	case AdminFieldTypesFetchResultsMsg:
		s.AdminFieldTypes = msg.Data
		s.Cursor = 0
		s.updateCursorMax()
		return s, LoadingStopCmd()

	// Data refresh messages (from CMS operations)
	case FieldTypesSet:
		s.FieldTypes = msg.FieldTypes
		s.updateCursorMax()
		return s, nil
	case AdminFieldTypesSet:
		s.AdminFieldTypes = msg.AdminFieldTypes
		s.updateCursorMax()
		return s, nil
	}

	return s, nil
}

func (s *FieldTypesScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

func (s *FieldTypesScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList()},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

// ---------------------------------------------------------------------------
// Render helpers
// ---------------------------------------------------------------------------

func (s *FieldTypesScreen) renderList() string {
	if s.AdminMode {
		return s.renderAdminList()
	}
	return s.renderRegularList()
}

func (s *FieldTypesScreen) renderRegularList() string {
	if len(s.FieldTypes) == 0 {
		return "(no field types)"
	}
	lines := make([]string, 0, len(s.FieldTypes))
	for i, ft := range s.FieldTypes {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s [%s]", cursor, ft.Label, ft.Type))
	}
	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderAdminList() string {
	if len(s.AdminFieldTypes) == 0 {
		return "(no admin field types)"
	}
	lines := make([]string, 0, len(s.AdminFieldTypes))
	for i, ft := range s.AdminFieldTypes {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s [%s]", cursor, ft.Label, ft.Type))
	}
	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderDetail() string {
	if s.AdminMode {
		return s.renderAdminDetail()
	}
	return s.renderRegularDetail()
}

func (s *FieldTypesScreen) renderRegularDetail() string {
	if len(s.FieldTypes) == 0 || s.Cursor >= len(s.FieldTypes) {
		return " No field type selected"
	}
	ft := s.FieldTypes[s.Cursor]

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Label  %s", ft.Label),
		fmt.Sprintf(" Type   %s", ft.Type),
		"",
		faint.Render(fmt.Sprintf(" ID     %s", ft.FieldTypeID)),
	}
	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderAdminDetail() string {
	if len(s.AdminFieldTypes) == 0 || s.Cursor >= len(s.AdminFieldTypes) {
		return " No admin field type selected"
	}
	ft := s.AdminFieldTypes[s.Cursor]

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Label  %s", ft.Label),
		fmt.Sprintf(" Type   %s", ft.Type),
		"",
		faint.Render(fmt.Sprintf(" ID     %s", ft.AdminFieldTypeID)),
	}
	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderInfo() string {
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	label := "Field Types"
	if s.AdminMode {
		label = "Admin Field Types"
	}

	lines := []string{
		accent.Render(fmt.Sprintf(" %s: %d", label, s.listLen())),
		"",
		" Field types define the available input types",
		" for datatype fields. Each field type maps to a",
		" specific editor widget in the admin panel and",
		" determines how content values are stored and",
		" validated.",
	}
	return strings.Join(lines, "\n")
}
