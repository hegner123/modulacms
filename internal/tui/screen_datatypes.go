package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// DatatypesScreen implements Screen for both DATATYPES and ADMINDATATYPES pages.
// When AdminMode is true, it operates on admin datatypes; otherwise regular datatypes.
// This is a 3-panel screen with TWO cursors:
//   - Cursor navigates the datatypes list (TreePanel / left)
//   - FieldCursor navigates the fields list (ContentPanel / center)
type DatatypesScreen struct {
	AdminMode      bool
	Cursor         int
	FieldCursor    int
	PanelFocus     FocusPanel
	Datatypes      []db.Datatypes
	AdminDatatypes []db.AdminDatatypes
	Fields         []db.Fields
	AdminFields    []db.AdminFields
	SelectedID     types.DatatypeID // last selected datatype ID (regular mode)
}

// NewDatatypesScreen creates a DatatypesScreen for regular or admin mode.
func NewDatatypesScreen(
	adminMode bool,
	datatypes []db.Datatypes,
	adminDatatypes []db.AdminDatatypes,
	fields []db.Fields,
	adminFields []db.AdminFields,
	selectedDatatype types.DatatypeID,
) *DatatypesScreen {
	return &DatatypesScreen{
		AdminMode:      adminMode,
		Cursor:         0,
		FieldCursor:    0,
		PanelFocus:     TreePanel,
		Datatypes:      datatypes,
		AdminDatatypes: adminDatatypes,
		Fields:         fields,
		AdminFields:    adminFields,
		SelectedID:     selectedDatatype,
	}
}

func (s *DatatypesScreen) PageIndex() PageIndex {
	if s.AdminMode {
		return ADMINDATATYPES
	}
	return DATATYPES
}

// datatypesLen returns the number of datatypes in the current mode.
func (s *DatatypesScreen) datatypesLen() int {
	if s.AdminMode {
		return len(s.AdminDatatypes)
	}
	return len(s.Datatypes)
}

// fieldsLen returns the number of fields in the current mode.
func (s *DatatypesScreen) fieldsLen() int {
	if s.AdminMode {
		return len(s.AdminFields)
	}
	return len(s.Fields)
}

func (s *DatatypesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Dismiss: esc goes back to tree panel before quitting
		if km.Matches(key, config.ActionDismiss) {
			if s.PanelFocus != TreePanel {
				s.PanelFocus = TreePanel
				return s, nil
			}
			return s, tea.Quit
		}

		// Quit always quits
		if km.Matches(key, config.ActionQuit) {
			return s, tea.Quit
		}

		// Panel navigation with field cursor reset on entering ContentPanel
		if km.Matches(key, config.ActionNextPanel) {
			s.PanelFocus = (s.PanelFocus + 1) % 3
			if s.PanelFocus == ContentPanel {
				s.FieldCursor = 0
			}
			return s, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			s.PanelFocus = (s.PanelFocus + 2) % 3
			if s.PanelFocus == ContentPanel {
				s.FieldCursor = 0
			}
			return s, nil
		}

		// Up/Down are panel-aware
		if km.Matches(key, config.ActionUp) {
			return s.handleUp()
		}
		if km.Matches(key, config.ActionDown) {
			return s.handleDown()
		}

		// Back: move focus to TreePanel first, then pop history
		if km.Matches(key, config.ActionBack) {
			if s.PanelFocus != TreePanel {
				s.PanelFocus = TreePanel
				return s, nil
			}
			return s, HistoryPopCmd()
		}

		// Right arrow moves to next panel
		if key == "l" || key == "right" {
			if s.PanelFocus == TreePanel {
				s.PanelFocus = ContentPanel
				s.FieldCursor = 0
				return s, nil
			}
			if s.PanelFocus == ContentPanel {
				s.PanelFocus = RoutePanel
				return s, nil
			}
		}

		// New (panel-aware)
		if km.Matches(key, config.ActionNew) {
			return s.handleNew()
		}

		// Edit (panel-aware)
		if km.Matches(key, config.ActionEdit) {
			return s.handleEdit()
		}

		// Delete (panel-aware)
		if km.Matches(key, config.ActionDelete) {
			return s.handleDelete()
		}

		// UI Config (regular mode only, ContentPanel only)
		if key == "u" && !s.AdminMode {
			return s.handleUIConfig()
		}

		// Enter/select (panel-aware)
		if key == "enter" {
			return s.handleSelect()
		}

	// Fetch request messages
	case AllDatatypesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			datatypes, err := d.ListDatatypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if datatypes == nil {
				return AllDatatypesFetchResultsMsg{Data: []db.Datatypes{}}
			}
			return AllDatatypesFetchResultsMsg{Data: *datatypes}
		}
	case AllDatatypesFetchResultsMsg:
		s.Datatypes = msg.Data
		s.Cursor = 0
		s.FieldCursor = 0
		cmds := []tea.Cmd{LoadingStopCmd()}
		if len(msg.Data) > 0 {
			cmds = append(cmds, DatatypeFieldsFetchCmd(msg.Data[0].DatatypeID))
		}
		return s, tea.Batch(cmds...)
	case AdminAllDatatypesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			datatypes, err := d.ListAdminDatatypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if datatypes == nil {
				return AdminAllDatatypesFetchResultsMsg{Data: []db.AdminDatatypes{}}
			}
			return AdminAllDatatypesFetchResultsMsg{Data: *datatypes}
		}
	case AdminAllDatatypesFetchResultsMsg:
		s.AdminDatatypes = msg.Data
		s.Cursor = 0
		s.FieldCursor = 0
		cmds := []tea.Cmd{LoadingStopCmd()}
		if len(msg.Data) > 0 {
			cmds = append(cmds, AdminDatatypeFieldsFetchCmd(msg.Data[0].AdminDatatypeID))
		}
		return s, tea.Batch(cmds...)
	case DatatypeFieldsFetchMsg:
		d := ctx.DB
		datatypeID := msg.DatatypeID
		return s, func() tea.Msg {
			fields, err := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: datatypeID, Valid: true})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if fields == nil || len(*fields) == 0 {
				return DatatypeFieldsFetchResultsMsg{Fields: []db.Fields{}}
			}
			return DatatypeFieldsFetchResultsMsg{Fields: *fields}
		}
	case DatatypeFieldsFetchResultsMsg:
		s.Fields = msg.Fields
		s.FieldCursor = 0
		return s, nil
	case AdminDatatypeFieldsFetchMsg:
		d := ctx.DB
		datatypeID := msg.AdminDatatypeID
		return s, func() tea.Msg {
			fields, err := d.ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID{ID: datatypeID, Valid: true})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if fields == nil || len(*fields) == 0 {
				return AdminDatatypeFieldsFetchResultsMsg{Fields: []db.AdminFields{}}
			}
			return AdminDatatypeFieldsFetchResultsMsg{Fields: *fields}
		}
	case AdminDatatypeFieldsFetchResultsMsg:
		s.AdminFields = msg.Fields
		s.FieldCursor = 0
		return s, nil

	// Data refresh messages (from CMS operations)
	case AllDatatypesSet:
		s.Datatypes = msg.AllDatatypes
		return s, nil
	case AdminAllDatatypesSet:
		s.AdminDatatypes = msg.AdminAllDatatypes
		return s, nil
	case DatatypeFieldsSet:
		s.Fields = msg.Fields
		return s, nil
	case AdminDatatypeFieldsSet:
		s.AdminFields = msg.Fields
		return s, nil
	}

	return s, nil
}

// handleUp processes upward cursor movement based on active panel.
func (s *DatatypesScreen) handleUp() (Screen, tea.Cmd) {
	switch s.PanelFocus {
	case TreePanel:
		if s.Cursor > 0 {
			s.Cursor--
			s.FieldCursor = 0
			return s, s.fetchFieldsForCurrent()
		}
	case ContentPanel:
		if s.FieldCursor > 0 {
			s.FieldCursor--
		}
	}
	return s, nil
}

// handleDown processes downward cursor movement based on active panel.
func (s *DatatypesScreen) handleDown() (Screen, tea.Cmd) {
	switch s.PanelFocus {
	case TreePanel:
		if s.Cursor < s.datatypesLen()-1 {
			s.Cursor++
			s.FieldCursor = 0
			return s, s.fetchFieldsForCurrent()
		}
	case ContentPanel:
		if s.FieldCursor < s.fieldsLen()-1 {
			s.FieldCursor++
		}
	}
	return s, nil
}

// fetchFieldsForCurrent returns a command to fetch fields for the datatype at the current cursor position.
func (s *DatatypesScreen) fetchFieldsForCurrent() tea.Cmd {
	if s.AdminMode {
		if s.Cursor < len(s.AdminDatatypes) {
			return AdminDatatypeFieldsFetchCmd(s.AdminDatatypes[s.Cursor].AdminDatatypeID)
		}
		return nil
	}
	if s.Cursor < len(s.Datatypes) {
		return DatatypeFieldsFetchCmd(s.Datatypes[s.Cursor].DatatypeID)
	}
	return nil
}

// handleNew handles creation key based on active panel and mode.
func (s *DatatypesScreen) handleNew() (Screen, tea.Cmd) {
	switch s.PanelFocus {
	case TreePanel:
		if s.AdminMode {
			return s, ShowAdminFormDialogCmd(FORMDIALOGCREATEADMINDATATYPE, "New Admin Datatype", s.AdminDatatypes)
		}
		return s, CmsDefineDatatypeLoadCmd()
	case ContentPanel:
		if s.AdminMode {
			if len(s.AdminDatatypes) > 0 && s.Cursor < len(s.AdminDatatypes) {
				dtID := string(s.AdminDatatypes[s.Cursor].AdminDatatypeID)
				return s, ShowFieldFormDialogCmd(FORMDIALOGCREATEADMINFIELD, "New Admin Field", dtID)
			}
		} else {
			if len(s.Datatypes) > 0 && s.Cursor < len(s.Datatypes) {
				dtID := string(s.Datatypes[s.Cursor].DatatypeID)
				return s, ShowFieldFormDialogCmd(FORMDIALOGCREATEFIELD, "New Field", dtID)
			}
		}
	}
	return s, nil
}

// handleEdit handles edit key based on active panel and mode.
func (s *DatatypesScreen) handleEdit() (Screen, tea.Cmd) {
	switch s.PanelFocus {
	case TreePanel:
		if s.AdminMode {
			if len(s.AdminDatatypes) > 0 && s.Cursor < len(s.AdminDatatypes) {
				dt := s.AdminDatatypes[s.Cursor]
				return s, ShowEditAdminDatatypeDialogCmd(dt, s.AdminDatatypes)
			}
		} else {
			if len(s.Datatypes) > 0 && s.Cursor < len(s.Datatypes) {
				return s, ShowEditDatatypeDialogCmd(s.Datatypes[s.Cursor], s.Datatypes)
			}
		}
	case ContentPanel:
		if s.AdminMode {
			if len(s.AdminFields) > 0 && s.FieldCursor < len(s.AdminFields) {
				return s, ShowEditAdminFieldDialogCmd(s.AdminFields[s.FieldCursor])
			}
		} else {
			if len(s.Fields) > 0 && s.FieldCursor < len(s.Fields) {
				return s, ShowEditFieldDialogCmd(s.Fields[s.FieldCursor])
			}
		}
	}
	return s, nil
}

// handleDelete handles deletion key based on active panel and mode.
func (s *DatatypesScreen) handleDelete() (Screen, tea.Cmd) {
	switch s.PanelFocus {
	case TreePanel:
		if s.AdminMode {
			if len(s.AdminDatatypes) > 0 && s.Cursor < len(s.AdminDatatypes) {
				dt := s.AdminDatatypes[s.Cursor]
				hasChildren := false
				for _, other := range s.AdminDatatypes {
					if other.ParentID.Valid && string(other.ParentID.ID) == string(dt.AdminDatatypeID) {
						hasChildren = true
						break
					}
				}
				return s, ShowDeleteAdminDatatypeDialogCmd(dt.AdminDatatypeID, dt.Label, hasChildren)
			}
		} else {
			if len(s.Datatypes) > 0 && s.Cursor < len(s.Datatypes) {
				dt := s.Datatypes[s.Cursor]
				hasChildren := false
				for _, other := range s.Datatypes {
					if other.ParentID.Valid && types.DatatypeID(other.ParentID.ID) == dt.DatatypeID {
						hasChildren = true
						break
					}
				}
				return s, ShowDeleteDatatypeDialogCmd(dt.DatatypeID, dt.Label, hasChildren)
			}
		}
	case ContentPanel:
		if s.AdminMode {
			if len(s.AdminFields) > 0 && s.FieldCursor < len(s.AdminFields) {
				field := s.AdminFields[s.FieldCursor]
				var datatypeID types.AdminDatatypeID
				if len(s.AdminDatatypes) > 0 && s.Cursor < len(s.AdminDatatypes) {
					datatypeID = s.AdminDatatypes[s.Cursor].AdminDatatypeID
				}
				return s, ShowDeleteAdminFieldDialogCmd(field.AdminFieldID, datatypeID, field.Label)
			}
		} else {
			if len(s.Fields) > 0 && s.FieldCursor < len(s.Fields) {
				field := s.Fields[s.FieldCursor]
				var datatypeID types.DatatypeID
				if len(s.Datatypes) > 0 && s.Cursor < len(s.Datatypes) {
					datatypeID = s.Datatypes[s.Cursor].DatatypeID
				}
				return s, ShowDeleteFieldDialogCmd(field.FieldID, datatypeID, field.Label)
			}
		}
	}
	return s, nil
}

// handleUIConfig handles the 'u' key to open the UIConfig dialog for a field (regular mode only).
func (s *DatatypesScreen) handleUIConfig() (Screen, tea.Cmd) {
	if s.PanelFocus != ContentPanel {
		return s, nil
	}
	if len(s.Fields) == 0 || s.FieldCursor >= len(s.Fields) {
		return s, nil
	}

	field := s.Fields[s.FieldCursor]
	fieldID := string(field.FieldID)

	uc, err := types.ParseUIConfig(field.UIConfig)
	if err != nil {
		return s, LogMessageCmd(fmt.Sprintf("Failed to parse UIConfig: %v", err))
	}

	isZero := uc.Widget == "" && uc.Placeholder == "" && uc.HelpText == "" && !uc.Hidden
	if isZero {
		return s, ShowUIConfigFormDialogCmd("UI Config: "+field.Label, fieldID)
	}
	return s, ShowEditUIConfigFormDialogCmd("UI Config: "+field.Label, fieldID, uc)
}

// handleSelect handles enter/selection key based on active panel and mode.
func (s *DatatypesScreen) handleSelect() (Screen, tea.Cmd) {
	switch s.PanelFocus {
	case TreePanel:
		if s.AdminMode {
			if len(s.AdminDatatypes) > 0 && s.Cursor < len(s.AdminDatatypes) {
				dt := s.AdminDatatypes[s.Cursor]
				s.PanelFocus = ContentPanel
				s.FieldCursor = 0
				return s, LogMessageCmd(fmt.Sprintf("Admin datatype selected: %s (%s)", dt.Label, dt.AdminDatatypeID))
			}
		} else {
			if len(s.Datatypes) > 0 && s.Cursor < len(s.Datatypes) {
				dt := s.Datatypes[s.Cursor]
				s.PanelFocus = ContentPanel
				s.FieldCursor = 0
				return s, LogMessageCmd(fmt.Sprintf("Datatype selected: %s (%s)", dt.Label, dt.DatatypeID))
			}
		}
	case ContentPanel:
		if s.AdminMode {
			if len(s.AdminFields) > 0 && s.FieldCursor < len(s.AdminFields) {
				field := s.AdminFields[s.FieldCursor]
				return s, LogMessageCmd(fmt.Sprintf("Admin field selected: %s [%s]", field.Label, field.Type))
			}
		} else {
			if len(s.Fields) > 0 && s.FieldCursor < len(s.Fields) {
				field := s.Fields[s.FieldCursor]
				return s, LogMessageCmd(fmt.Sprintf("Field selected: %s [%s]", field.Label, field.Type))
			}
		}
	}
	return s, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) KeyHints(km config.KeyMap) []KeyHint {
	switch s.PanelFocus {
	case ContentPanel:
		return []KeyHint{
			{km.HintString(config.ActionNew), "new field"},
			{km.HintString(config.ActionEdit), "edit"},
			{km.HintString(config.ActionDelete), "del"},
			{"u", "ui config"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	default:
		return []KeyHint{
			{km.HintString(config.ActionNew), "new"},
			{km.HintString(config.ActionEdit), "edit"},
			{km.HintString(config.ActionDelete), "del"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	}
}

func (s *DatatypesScreen) View(ctx AppContext) string {
	left := s.renderList()
	center := s.renderFields()
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
	dtLen := s.datatypesLen()
	fLen := s.fieldsLen()
	// Fields panel has 2 header lines ("Fields for: ..." + blank) before items
	fieldsTotalLines := fLen + 2

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel, TotalLines: dtLen, ScrollOffset: ClampScroll(s.Cursor, dtLen, innerH)}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[1], Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel, TotalLines: fieldsTotalLines, ScrollOffset: ClampScroll(s.FieldCursor+2, fieldsTotalLines, innerH)}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

// ---------------------------------------------------------------------------
// Render helpers: List (left panel)
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) renderList() string {
	if s.AdminMode {
		return s.renderAdminDatatypesList()
	}
	return s.renderDatatypesList()
}

func (s *DatatypesScreen) renderDatatypesList() string {
	if len(s.Datatypes) == 0 {
		return "(no datatypes)"
	}

	lines := make([]string, 0, len(s.Datatypes))
	for i, dt := range s.Datatypes {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		parent := ""
		if dt.ParentID.Valid {
			parent = fmt.Sprintf(" (child of %s)", dt.ParentID.ID)
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]%s", cursor, dt.Label, dt.Type, parent))
	}
	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderAdminDatatypesList() string {
	if len(s.AdminDatatypes) == 0 {
		return "(no admin datatypes)"
	}

	lines := make([]string, 0, len(s.AdminDatatypes))
	for i, dt := range s.AdminDatatypes {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		parent := ""
		if dt.ParentID.Valid {
			parent = fmt.Sprintf(" (child of %s)", dt.ParentID.ID)
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]%s", cursor, dt.Label, dt.Type, parent))
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Render helpers: Fields (center panel)
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) renderFields() string {
	if s.AdminMode {
		return s.renderAdminDatatypeFields()
	}
	return s.renderDatatypeFields()
}

func (s *DatatypesScreen) renderDatatypeFields() string {
	if len(s.Datatypes) == 0 || s.Cursor >= len(s.Datatypes) {
		return "No datatype selected"
	}

	dt := s.Datatypes[s.Cursor]
	lines := []string{
		fmt.Sprintf("Fields for: %s", dt.Label),
		"",
	}

	if len(s.Fields) == 0 {
		if s.PanelFocus == ContentPanel {
			lines = append(lines, " -> (empty)")
		} else {
			lines = append(lines, "    (empty)")
		}
		lines = append(lines, "")
		lines = append(lines, "Press 'n' to add a field")
	} else {
		for i, field := range s.Fields {
			cursor := "   "
			if s.PanelFocus == ContentPanel && s.FieldCursor == i {
				cursor = " ->"
			}
			lines = append(lines, fmt.Sprintf("%s %d. %s [%s]", cursor, i+1, field.Label, field.Type))
		}
	}

	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderAdminDatatypeFields() string {
	if len(s.AdminDatatypes) == 0 || s.Cursor >= len(s.AdminDatatypes) {
		return "No admin datatype selected"
	}

	dt := s.AdminDatatypes[s.Cursor]
	lines := []string{
		fmt.Sprintf("Fields for: %s", dt.Label),
		"",
	}

	if len(s.AdminFields) == 0 {
		if s.PanelFocus == ContentPanel {
			lines = append(lines, " -> (empty)")
		} else {
			lines = append(lines, "    (empty)")
		}
		lines = append(lines, "")
		lines = append(lines, "Press 'n' to add a field")
	} else {
		for i, field := range s.AdminFields {
			cursor := "   "
			if s.PanelFocus == ContentPanel && s.FieldCursor == i {
				cursor = " ->"
			}
			lines = append(lines, fmt.Sprintf("%s %d. %s [%s]", cursor, i+1, field.Label, field.Type))
		}
	}

	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Render helpers: Actions (right panel)
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) renderActions() string {
	if s.AdminMode {
		return s.renderAdminDatatypeActions()
	}
	return s.renderDatatypeActions()
}

func (s *DatatypesScreen) renderDatatypeActions() string {
	lines := []string{
		"Actions",
		"",
	}

	switch s.PanelFocus {
	case TreePanel:
		lines = append(lines,
			"Datatypes Panel",
			"",
			"  n: New datatype",
			"  e: Edit datatype",
			"  d: Delete datatype",
			"",
			"  enter: Select",
			"  tab: Switch panel",
		)
	case ContentPanel:
		lines = append(lines,
			"Fields Panel",
			"",
			"  n: New field",
			"  e: Edit field",
			"  d: Delete field",
			"  u: UI Config",
			"",
			"  esc/h: Back to datatypes",
			"  tab: Switch panel",
		)
	default:
		lines = append(lines,
			"  n: New",
			"  e: Edit",
			"  d: Delete",
		)
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Datatypes: %d", s.datatypesLen()))
	if s.datatypesLen() > 0 && s.Cursor < s.datatypesLen() {
		lines = append(lines, fmt.Sprintf("Fields: %d", s.fieldsLen()))
	}

	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderAdminDatatypeActions() string {
	lines := []string{
		"Actions",
		"",
	}

	switch s.PanelFocus {
	case TreePanel:
		lines = append(lines,
			"Datatypes Panel",
			"",
			"  n: New datatype",
			"  e: Edit datatype",
			"  d: Delete datatype",
			"",
			"  enter: Select",
			"  tab: Switch panel",
		)
	case ContentPanel:
		lines = append(lines,
			"Fields Panel",
			"",
			"  n: New field",
			"  e: Edit field",
			"  d: Delete field",
			"",
			"  esc/h: Back to datatypes",
			"  tab: Switch panel",
		)
	default:
		lines = append(lines,
			"  n: New",
			"  e: Edit",
			"  d: Delete",
		)
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Datatypes: %d", s.datatypesLen()))
	if s.datatypesLen() > 0 && s.Cursor < s.datatypesLen() {
		lines = append(lines, fmt.Sprintf("Fields: %d", s.fieldsLen()))
	}

	return strings.Join(lines, "\n")
}
