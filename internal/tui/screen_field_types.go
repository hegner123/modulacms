package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// FieldTypesScreen implements Screen for both FIELDTYPES and ADMINFIELDTYPES pages.
// When AdminMode is true, it operates on admin field types; otherwise regular field types.
type FieldTypesScreen struct {
	AdminMode       bool
	Cursor          int
	PanelFocus      FocusPanel
	FieldTypes      []db.FieldTypes
	AdminFieldTypes []db.AdminFieldTypes
}

// NewFieldTypesScreen creates a FieldTypesScreen for regular or admin mode.
func NewFieldTypesScreen(adminMode bool, fieldTypes []db.FieldTypes, adminFieldTypes []db.AdminFieldTypes) *FieldTypesScreen {
	return &FieldTypesScreen{
		AdminMode:       adminMode,
		Cursor:          0,
		PanelFocus:      TreePanel,
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

func (s *FieldTypesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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
		cursorMax := s.listLen() - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, cursorMax)
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
		return s, LoadingStopCmd()

	// Data refresh messages (from CMS operations)
	case FieldTypesSet:
		s.FieldTypes = msg.FieldTypes
		return s, nil
	case AdminFieldTypesSet:
		s.AdminFieldTypes = msg.AdminFieldTypes
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
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *FieldTypesScreen) View(ctx AppContext) string {
	left := s.renderList()
	center := s.renderDetail()
	right := s.renderActions()

	layout := layoutForPage(s.PageIndex())
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	listLen := s.listLen()

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

func (s *FieldTypesScreen) renderList() string {
	if s.AdminMode {
		return s.renderAdminFieldTypesList()
	}
	return s.renderFieldTypesList()
}

func (s *FieldTypesScreen) renderFieldTypesList() string {
	if len(s.FieldTypes) == 0 {
		return "(no field types)"
	}

	lines := make([]string, 0, len(s.FieldTypes))
	for i, ft := range s.FieldTypes {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]", cursor, ft.Label, ft.Type))
	}
	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderAdminFieldTypesList() string {
	if len(s.AdminFieldTypes) == 0 {
		return "(no admin field types)"
	}

	lines := make([]string, 0, len(s.AdminFieldTypes))
	for i, ft := range s.AdminFieldTypes {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]", cursor, ft.Label, ft.Type))
	}
	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderDetail() string {
	if s.AdminMode {
		return s.renderAdminFieldTypeDetail()
	}
	return s.renderFieldTypeDetail()
}

func (s *FieldTypesScreen) renderFieldTypeDetail() string {
	if len(s.FieldTypes) == 0 || s.Cursor >= len(s.FieldTypes) {
		return "No field type selected"
	}

	ft := s.FieldTypes[s.Cursor]
	lines := []string{
		fmt.Sprintf("Type:  %s", ft.Type),
		fmt.Sprintf("Label: %s", ft.Label),
		"",
		fmt.Sprintf("ID:    %s", ft.FieldTypeID),
	}

	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderAdminFieldTypeDetail() string {
	if len(s.AdminFieldTypes) == 0 || s.Cursor >= len(s.AdminFieldTypes) {
		return "No admin field type selected"
	}

	ft := s.AdminFieldTypes[s.Cursor]
	lines := []string{
		fmt.Sprintf("Type:  %s", ft.Type),
		fmt.Sprintf("Label: %s", ft.Label),
		"",
		fmt.Sprintf("ID:    %s", ft.AdminFieldTypeID),
	}

	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderActions() string {
	if s.AdminMode {
		return s.renderAdminFieldTypeActions()
	}
	return s.renderFieldTypeActions()
}

func (s *FieldTypesScreen) renderFieldTypeActions() string {
	lines := []string{
		"Actions",
		"",
		"  n: New",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Field Types: %d", len(s.FieldTypes)),
	}

	return strings.Join(lines, "\n")
}

func (s *FieldTypesScreen) renderAdminFieldTypeActions() string {
	lines := []string{
		"Actions",
		"",
		"  n: New",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Admin Field Types: %d", len(s.AdminFieldTypes)),
	}

	return strings.Join(lines, "\n")
}
