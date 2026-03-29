package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// Validations grid: 2 columns
//
//	Col 0 (span 3): Validation list
//	Col 1 (span 9): Details (top), Info (bottom)
var validationsGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Validations"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.45, Title: "Details"},
			{Height: 0.55, Title: "Info"},
		}},
	},
}

// ValidationsScreen implements Screen for both VALIDATIONS and ADMINVALIDATIONS pages.
// When AdminMode is true, it operates on admin validations; otherwise regular validations.
type ValidationsScreen struct {
	GridScreen
	AdminMode        bool
	Validations      []db.Validation
	AdminValidations []db.AdminValidation
}

// NewValidationsScreen creates a ValidationsScreen for regular or admin mode.
func NewValidationsScreen(adminMode bool, validations []db.Validation, adminValidations []db.AdminValidation) *ValidationsScreen {
	listLen := len(validations)
	if adminMode {
		listLen = len(adminValidations)
	}
	cursorMax := listLen - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &ValidationsScreen{
		GridScreen: GridScreen{
			Grid:      validationsGrid,
			CursorMax: cursorMax,
		},
		AdminMode:        adminMode,
		Validations:      validations,
		AdminValidations: adminValidations,
	}
}

func (s *ValidationsScreen) PageIndex() PageIndex {
	if s.AdminMode {
		return ADMINVALIDATIONS
	}
	return VALIDATIONS
}

func (s *ValidationsScreen) listLen() int {
	if s.AdminMode {
		return len(s.AdminValidations)
	}
	return len(s.Validations)
}

func (s *ValidationsScreen) updateCursorMax() {
	s.CursorMax = s.listLen() - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.Cursor > s.CursorMax && s.CursorMax >= 0 {
		s.Cursor = s.CursorMax
	}
}

func (s *ValidationsScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// New validation
		if km.Matches(key, config.ActionNew) {
			if s.AdminMode {
				return s, ShowRouteFormDialogCmd(FORMDIALOGCREATEADMINVALIDATION, "New Admin Validation")
			}
			return s, ShowRouteFormDialogCmd(FORMDIALOGCREATEVALIDATION, "New Validation")
		}

		// Edit validation
		if km.Matches(key, config.ActionEdit) {
			if s.AdminMode {
				if len(s.AdminValidations) > 0 && s.Cursor < len(s.AdminValidations) {
					return s, ShowEditAdminValidationDialogCmd(s.AdminValidations[s.Cursor])
				}
			} else {
				if len(s.Validations) > 0 && s.Cursor < len(s.Validations) {
					return s, ShowEditValidationDialogCmd(s.Validations[s.Cursor])
				}
			}
			return s, nil
		}

		// Delete validation
		if km.Matches(key, config.ActionDelete) {
			if s.AdminMode {
				if len(s.AdminValidations) > 0 && s.Cursor < len(s.AdminValidations) {
					v := s.AdminValidations[s.Cursor]
					return s, ShowDeleteAdminValidationDialogCmd(v.AdminValidationID, v.Name)
				}
			} else {
				if len(s.Validations) > 0 && s.Cursor < len(s.Validations) {
					v := s.Validations[s.Cursor]
					return s, ShowDeleteValidationDialogCmd(v.ValidationID, v.Name)
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
	case ValidationsFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			validations, err := d.ListValidations()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if validations == nil {
				return ValidationsFetchResultsMsg{Data: []db.Validation{}}
			}
			return ValidationsFetchResultsMsg{Data: *validations}
		}
	case ValidationsFetchResultsMsg:
		s.Validations = msg.Data
		s.Cursor = 0
		s.updateCursorMax()
		return s, LoadingStopCmd()
	case AdminValidationsFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			adminValidations, err := d.ListAdminValidations()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if adminValidations == nil {
				return AdminValidationsFetchResultsMsg{Data: []db.AdminValidation{}}
			}
			return AdminValidationsFetchResultsMsg{Data: *adminValidations}
		}
	case AdminValidationsFetchResultsMsg:
		s.AdminValidations = msg.Data
		s.Cursor = 0
		s.updateCursorMax()
		return s, LoadingStopCmd()

	// Data refresh messages (from CMS operations)
	case ValidationsSet:
		s.Validations = msg.Validations
		s.updateCursorMax()
		return s, nil
	case AdminValidationsSet:
		s.AdminValidations = msg.AdminValidations
		s.updateCursorMax()
		return s, nil
	}

	return s, nil
}

func (s *ValidationsScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

func (s *ValidationsScreen) View(ctx AppContext) string {
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

func (s *ValidationsScreen) renderList() string {
	if s.AdminMode {
		return s.renderAdminList()
	}
	return s.renderRegularList()
}

func (s *ValidationsScreen) renderRegularList() string {
	if len(s.Validations) == 0 {
		return "(no validations)"
	}
	lines := make([]string, 0, len(s.Validations))
	for i, v := range s.Validations {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s", cursor, v.Name))
	}
	return strings.Join(lines, "\n")
}

func (s *ValidationsScreen) renderAdminList() string {
	if len(s.AdminValidations) == 0 {
		return "(no admin validations)"
	}
	lines := make([]string, 0, len(s.AdminValidations))
	for i, v := range s.AdminValidations {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s", cursor, v.Name))
	}
	return strings.Join(lines, "\n")
}

func (s *ValidationsScreen) renderDetail() string {
	if s.AdminMode {
		return s.renderAdminDetail()
	}
	return s.renderRegularDetail()
}

func (s *ValidationsScreen) renderRegularDetail() string {
	if len(s.Validations) == 0 || s.Cursor >= len(s.Validations) {
		return " No validation selected"
	}
	v := s.Validations[s.Cursor]

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Name         %s", v.Name),
		fmt.Sprintf(" Description  %s", v.Description),
		"",
		faint.Render(fmt.Sprintf(" ID           %s", v.ValidationID)),
		faint.Render(fmt.Sprintf(" Modified     %s", v.DateModified)),
	}
	return strings.Join(lines, "\n")
}

func (s *ValidationsScreen) renderAdminDetail() string {
	if len(s.AdminValidations) == 0 || s.Cursor >= len(s.AdminValidations) {
		return " No admin validation selected"
	}
	v := s.AdminValidations[s.Cursor]

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Name         %s", v.Name),
		fmt.Sprintf(" Description  %s", v.Description),
		"",
		faint.Render(fmt.Sprintf(" ID           %s", v.AdminValidationID)),
		faint.Render(fmt.Sprintf(" Modified     %s", v.DateModified)),
	}
	return strings.Join(lines, "\n")
}

func (s *ValidationsScreen) renderInfo() string {
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	label := "Validations"
	if s.AdminMode {
		label = "Admin Validations"
	}

	lines := []string{
		accent.Render(fmt.Sprintf(" %s: %d", label, s.listLen())),
		"",
		" Validation configs define reusable validation",
		" rules that can be attached to fields. Each",
		" config contains a JSON schema describing the",
		" constraints applied when content is saved.",
	}
	return strings.Join(lines, "\n")
}
