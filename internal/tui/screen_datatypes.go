package tui

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Phase constants for the DatatypesScreen state machine.
const (
	DatatypesPhaseBrowse = 0 // Phase 1: datatype tree browsing
	DatatypesPhaseFields = 1 // Phase 2: field selection for a datatype
)

// Grid definitions for each phase.
var datatypeBrowseGrid = Grid{
	Columns: []GridColumn{
		{Span: 4, Cells: []GridCell{
			{Height: 1.0, Title: "Datatypes"},
		}},
		{Span: 8, Cells: []GridCell{
			{Height: 0.30, Title: "Details"},
			{Height: 0.70, Title: "Fields"},
		}},
	},
}

var datatypeFieldGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Fields"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.65, Title: "Properties"},
			{Height: 0.35, Title: "Context"},
		}},
	},
}

// DatatypesScreen implements Screen for both DATATYPES and ADMINDATATYPES pages.
type DatatypesScreen struct {
	GridScreen
	AdminMode bool

	// Phase tracking
	Phase         int // 0=browse, 1=fields
	SavedDTCursor int // preserved datatype cursor when in Phase 2

	// Phase 1: Datatype browse
	Datatypes      []db.Datatypes
	AdminDatatypes []db.AdminDatatypes
	DatatypeTree   []*DatatypeTreeNode
	FlatDTList     []*DatatypeTreeNode
	Searching      bool
	SearchInput    textinput.Model
	SearchQuery    string

	// Phase 2: Field selection
	SelectedDTNode *DatatypeTreeNode // the datatype entered in Phase 2
	Fields         []db.Fields
	AdminFields    []db.AdminFields
	Properties     []FieldProperty // read-only preview from selected field
}

// NewDatatypesScreen creates a DatatypesScreen for regular or admin mode.
func NewDatatypesScreen(adminMode bool) *DatatypesScreen {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 128

	s := &DatatypesScreen{
		GridScreen: GridScreen{
			Grid: datatypeBrowseGrid,
		},
		AdminMode:   adminMode,
		Phase:       DatatypesPhaseBrowse,
		SearchInput: ti,
	}
	return s
}

func (s *DatatypesScreen) PageIndex() PageIndex {
	if s.AdminMode {
		return ADMINDATATYPES
	}
	return DATATYPES
}

// rebuildTree rebuilds the tree and flat list from current data (filtered if searching).
func (s *DatatypesScreen) rebuildTree() {
	if s.AdminMode {
		filtered := FilterAdminDatatypeList(s.AdminDatatypes, s.SearchQuery)
		s.DatatypeTree = BuildAdminDatatypeTree(filtered)
	} else {
		filtered := FilterDatatypeList(s.Datatypes, s.SearchQuery)
		s.DatatypeTree = BuildDatatypeTree(filtered)
	}
	s.FlatDTList = FlattenDatatypeTree(s.DatatypeTree)
	s.CursorMax = len(s.FlatDTList) - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.Cursor > s.CursorMax {
		s.Cursor = s.CursorMax
	}
}

// flatDTLen returns the length of the flat datatype list.
func (s *DatatypesScreen) flatDTLen() int {
	return len(s.FlatDTList)
}

// fieldsLen returns the field count for current mode.
func (s *DatatypesScreen) fieldsLen() int {
	if s.AdminMode {
		return len(s.AdminFields)
	}
	return len(s.Fields)
}

// selectedDTID returns the datatype ID of the current cursor position in Phase 1.
func (s *DatatypesScreen) selectedDTID() string {
	if s.Cursor >= len(s.FlatDTList) || len(s.FlatDTList) == 0 {
		return ""
	}
	return s.FlatDTList[s.Cursor].DatatypeID()
}

// fetchFieldsForCurrentDT returns a command to fetch fields for the cursor-highlighted datatype.
func (s *DatatypesScreen) fetchFieldsForCurrentDT() tea.Cmd {
	if len(s.FlatDTList) == 0 || s.Cursor >= len(s.FlatDTList) {
		return nil
	}
	node := s.FlatDTList[s.Cursor]
	if s.AdminMode && node.AdminDT != nil {
		return AdminDatatypeFieldsFetchCmd(node.AdminDT.AdminDatatypeID)
	}
	if !s.AdminMode && node.Datatype != nil {
		return DatatypeFieldsFetchCmd(node.Datatype.DatatypeID)
	}
	return nil
}

// enterFieldPhase transitions from Phase 1 to Phase 2.
func (s *DatatypesScreen) enterFieldPhase() tea.Cmd {
	if len(s.FlatDTList) == 0 || s.Cursor >= len(s.FlatDTList) {
		return nil
	}
	s.SelectedDTNode = s.FlatDTList[s.Cursor]
	s.SavedDTCursor = s.Cursor
	s.Phase = DatatypesPhaseFields
	s.Grid = datatypeFieldGrid
	s.Cursor = 0
	maxFields := s.fieldsLen() - 1
	if maxFields < 0 {
		maxFields = 0
	}
	s.CursorMax = maxFields
	s.FocusIndex = 0
	s.Properties = nil

	// Fetch fields for the selected datatype
	if s.AdminMode && s.SelectedDTNode.AdminDT != nil {
		return AdminDatatypeFieldsFetchCmd(s.SelectedDTNode.AdminDT.AdminDatatypeID)
	}
	if !s.AdminMode && s.SelectedDTNode.Datatype != nil {
		return DatatypeFieldsFetchCmd(s.SelectedDTNode.Datatype.DatatypeID)
	}
	return nil
}

// exitFieldPhase transitions from Phase 2 back to Phase 1.
func (s *DatatypesScreen) exitFieldPhase() {
	s.Phase = DatatypesPhaseBrowse
	s.Grid = datatypeBrowseGrid
	s.Cursor = s.SavedDTCursor
	s.CursorMax = len(s.FlatDTList) - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.Cursor > s.CursorMax {
		s.Cursor = s.CursorMax
	}
	s.FocusIndex = 0
	s.SelectedDTNode = nil
	s.Properties = nil
}

// rebuildProperties rebuilds the property list from the currently highlighted field.
func (s *DatatypesScreen) rebuildProperties() {
	if s.AdminMode {
		if s.Cursor < len(s.AdminFields) {
			s.Properties = FieldPropertiesFromAdminField(s.AdminFields[s.Cursor])
			return
		}
	} else {
		if s.Cursor < len(s.Fields) {
			s.Properties = FieldPropertiesFromField(s.Fields[s.Cursor])
			return
		}
	}
	s.Properties = nil
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if s.Phase == DatatypesPhaseBrowse {
			return s.updateBrowse(ctx, msg)
		}
		return s.updateFields(ctx, msg)

	// Fetch: datatypes
	case AllDatatypesFetchMsg:
		return s.handleAllDatatypesFetch(ctx)
	case AllDatatypesFetchResultsMsg:
		return s.handleAllDatatypesFetchResults(msg)
	case AdminAllDatatypesFetchMsg:
		return s.handleAdminAllDatatypesFetch(ctx)
	case AdminAllDatatypesFetchResultsMsg:
		return s.handleAdminAllDatatypesFetchResults(msg)

	// Fetch: fields
	case DatatypeFieldsFetchMsg:
		return s.handleDatatypeFieldsFetch(ctx, msg)
	case DatatypeFieldsFetchResultsMsg:
		return s.handleDatatypeFieldsFetchResults(msg)
	case AdminDatatypeFieldsFetchMsg:
		return s.handleAdminDatatypeFieldsFetch(ctx, msg)
	case AdminDatatypeFieldsFetchResultsMsg:
		return s.handleAdminDatatypeFieldsFetchResults(msg)

	// Reorder results
	case DatatypeReorderedMsg:
		return s.handleDatatypeReordered(msg)
	case AdminDatatypeReorderedMsg:
		return s.handleAdminDatatypeReordered(msg)
	case DatatypeFieldReorderedMsg:
		return s.handleFieldReordered(msg)
	case AdminDatatypeFieldReorderedMsg:
		return s.handleAdminFieldReordered(msg)

	// Data refresh (from CMS operations)
	case AllDatatypesSet:
		s.Datatypes = msg.AllDatatypes
		s.rebuildTree()
		return s, nil
	case AdminAllDatatypesSet:
		s.AdminDatatypes = msg.AdminAllDatatypes
		s.rebuildTree()
		return s, nil
	case DatatypeFieldsSet:
		s.Fields = msg.Fields
		if s.Phase == DatatypesPhaseFields {
			s.CursorMax = len(s.Fields) - 1
			if s.CursorMax < 0 {
				s.CursorMax = 0
			}
			if s.Cursor > s.CursorMax {
				s.Cursor = s.CursorMax
			}
			s.rebuildProperties()
		}
		return s, nil
	case AdminDatatypeFieldsSet:
		s.AdminFields = msg.Fields
		if s.Phase == DatatypesPhaseFields {
			s.CursorMax = len(s.AdminFields) - 1
			if s.CursorMax < 0 {
				s.CursorMax = 0
			}
			if s.Cursor > s.CursorMax {
				s.Cursor = s.CursorMax
			}
			s.rebuildProperties()
		}
		return s, nil
	}

	return s, nil
}

// ---------------------------------------------------------------------------
// Phase 1: Browse
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) updateBrowse(ctx AppContext, msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	km := ctx.Config.KeyBindings
	key := msg.String()

	// Search mode: all input goes to textinput
	if s.Searching {
		return s.updateSearch(msg)
	}

	if s.HandleFocusNav(key, km) {
		return s, nil
	}

	// Start search
	if km.Matches(key, config.ActionSearch) {
		s.Searching = true
		s.SearchInput.Focus()
		return s, nil
	}

	// Expand/Collapse
	if key == "right" || key == "l" {
		if len(s.FlatDTList) > 0 && s.Cursor < len(s.FlatDTList) {
			node := s.FlatDTList[s.Cursor]
			if node.Kind == DatatypeNodeGroup && !node.Expand {
				node.Expand = true
				oldID := s.selectedDTID()
				s.FlatDTList = FlattenDatatypeTree(s.DatatypeTree)
				s.CursorMax = len(s.FlatDTList) - 1
				if s.CursorMax < 0 {
					s.CursorMax = 0
				}
				_ = oldID // cursor position preserved, same node
				return s, nil
			}
		}
	}
	if key == "left" || key == "h" {
		if len(s.FlatDTList) > 0 && s.Cursor < len(s.FlatDTList) {
			node := s.FlatDTList[s.Cursor]
			if node.Kind == DatatypeNodeGroup && node.Expand {
				node.Expand = false
				s.FlatDTList = FlattenDatatypeTree(s.DatatypeTree)
				s.CursorMax = len(s.FlatDTList) - 1
				if s.CursorMax < 0 {
					s.CursorMax = 0
				}
				if s.Cursor > s.CursorMax {
					s.Cursor = s.CursorMax
				}
				return s, s.fetchFieldsForCurrentDT()
			}
		}
	}

	// Enter: select datatype and enter field phase
	if km.Matches(key, config.ActionSelect) {
		cmd := s.enterFieldPhase()
		return s, cmd
	}

	// New datatype
	if km.Matches(key, config.ActionNew) {
		if s.AdminMode {
			return s, ShowAdminFormDialogCmd(FORMDIALOGCREATEADMINDATATYPE, "New Admin Datatype", s.AdminDatatypes)
		}
		return s, ShowFormDialogCmd(FORMDIALOGCREATEDATATYPE, "New Datatype", s.Datatypes)
	}

	// Edit datatype
	if km.Matches(key, config.ActionEdit) {
		if s.AdminMode {
			if len(s.AdminDatatypes) > 0 && s.Cursor < len(s.FlatDTList) {
				node := s.FlatDTList[s.Cursor]
				if node.AdminDT != nil {
					return s, ShowEditAdminDatatypeDialogCmd(*node.AdminDT, s.AdminDatatypes)
				}
			}
		} else {
			if len(s.Datatypes) > 0 && s.Cursor < len(s.FlatDTList) {
				node := s.FlatDTList[s.Cursor]
				if node.Datatype != nil {
					return s, ShowEditDatatypeDialogCmd(*node.Datatype, s.Datatypes)
				}
			}
		}
		return s, nil
	}

	// Delete datatype
	if km.Matches(key, config.ActionDelete) {
		return s.handleDeleteDatatype()
	}

	// Reorder
	if km.Matches(key, config.ActionReorderUp) {
		return s.handleReorderDatatypeUp()
	}
	if km.Matches(key, config.ActionReorderDown) {
		return s.handleReorderDatatypeDown()
	}

	// Cursor movement with field fetch on change
	if km.Matches(key, config.ActionUp) {
		if s.Cursor > 0 {
			oldID := s.selectedDTID()
			s.Cursor--
			newID := s.selectedDTID()
			if oldID != newID {
				return s, s.fetchFieldsForCurrentDT()
			}
		}
		return s, nil
	}
	if km.Matches(key, config.ActionDown) {
		if s.Cursor < s.CursorMax {
			oldID := s.selectedDTID()
			s.Cursor++
			newID := s.selectedDTID()
			if oldID != newID {
				return s, s.fetchFieldsForCurrentDT()
			}
		}
		return s, nil
	}

	// Common keys (quit, back)
	_, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
	if handled {
		return s, cmd
	}

	return s, nil
}

func (s *DatatypesScreen) updateSearch(msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		s.Searching = false
		s.SearchQuery = s.SearchInput.Value()
		s.SearchInput.Blur()
		s.rebuildTree()
		return s, s.fetchFieldsForCurrentDT()
	case "esc":
		s.Searching = false
		s.SearchQuery = ""
		s.SearchInput.SetValue("")
		s.SearchInput.Blur()
		s.rebuildTree()
		return s, s.fetchFieldsForCurrentDT()
	default:
		var cmd tea.Cmd
		s.SearchInput, cmd = s.SearchInput.Update(msg)
		// Live filter on each keystroke
		s.SearchQuery = s.SearchInput.Value()
		s.rebuildTree()
		return s, cmd
	}
}

// ---------------------------------------------------------------------------
// Phase 2: Fields
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) updateFields(ctx AppContext, msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	km := ctx.Config.KeyBindings
	key := msg.String()

	if s.HandleFocusNav(key, km) {
		return s, nil
	}

	// Back: return to browse phase
	if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
		s.exitFieldPhase()
		return s, nil
	}

	// Quit
	if km.Matches(key, config.ActionQuit) {
		return s, tea.Quit
	}

	// Enter: open edit-all-at-once field dialog (Deliverable A behavior)
	if km.Matches(key, config.ActionSelect) {
		if s.AdminMode {
			if len(s.AdminFields) > 0 && s.Cursor < len(s.AdminFields) {
				return s, ShowEditAdminFieldDialogCmd(s.AdminFields[s.Cursor])
			}
		} else {
			if len(s.Fields) > 0 && s.Cursor < len(s.Fields) {
				return s, ShowEditFieldDialogCmd(s.Fields[s.Cursor])
			}
		}
		return s, nil
	}

	// New field
	if km.Matches(key, config.ActionNew) {
		return s.handleNewField()
	}

	// Edit field
	if km.Matches(key, config.ActionEdit) {
		if s.AdminMode {
			if len(s.AdminFields) > 0 && s.Cursor < len(s.AdminFields) {
				return s, ShowEditAdminFieldDialogCmd(s.AdminFields[s.Cursor])
			}
		} else {
			if len(s.Fields) > 0 && s.Cursor < len(s.Fields) {
				return s, ShowEditFieldDialogCmd(s.Fields[s.Cursor])
			}
		}
		return s, nil
	}

	// Delete field
	if km.Matches(key, config.ActionDelete) {
		return s.handleDeleteField()
	}

	// Reorder field
	if km.Matches(key, config.ActionReorderUp) {
		return s.handleReorderFieldUp()
	}
	if km.Matches(key, config.ActionReorderDown) {
		return s.handleReorderFieldDown()
	}

	// Cursor movement with property rebuild
	if km.Matches(key, config.ActionUp) {
		if s.Cursor > 0 {
			s.Cursor--
			s.rebuildProperties()
		}
		return s, nil
	}
	if km.Matches(key, config.ActionDown) {
		if s.Cursor < s.CursorMax {
			s.Cursor++
			s.rebuildProperties()
		}
		return s, nil
	}

	return s, nil
}

// ---------------------------------------------------------------------------
// Action handlers
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) handleDeleteDatatype() (Screen, tea.Cmd) {
	if len(s.FlatDTList) == 0 || s.Cursor >= len(s.FlatDTList) {
		return s, nil
	}
	node := s.FlatDTList[s.Cursor]
	hasChildren := len(node.Children) > 0

	if s.AdminMode && node.AdminDT != nil {
		return s, ShowDeleteAdminDatatypeDialogCmd(node.AdminDT.AdminDatatypeID, node.Label, hasChildren)
	}
	if !s.AdminMode && node.Datatype != nil {
		return s, ShowDeleteDatatypeDialogCmd(node.Datatype.DatatypeID, node.Label, hasChildren)
	}
	return s, nil
}

func (s *DatatypesScreen) handleNewField() (Screen, tea.Cmd) {
	if s.SelectedDTNode == nil {
		return s, nil
	}
	if s.AdminMode && s.SelectedDTNode.AdminDT != nil {
		dtID := string(s.SelectedDTNode.AdminDT.AdminDatatypeID)
		return s, ShowFieldFormDialogCmd(FORMDIALOGCREATEADMINFIELD, "New Admin Field", dtID)
	}
	if !s.AdminMode && s.SelectedDTNode.Datatype != nil {
		dtID := string(s.SelectedDTNode.Datatype.DatatypeID)
		return s, ShowFieldFormDialogCmd(FORMDIALOGCREATEFIELD, "New Field", dtID)
	}
	return s, nil
}

func (s *DatatypesScreen) handleDeleteField() (Screen, tea.Cmd) {
	if s.AdminMode {
		if len(s.AdminFields) > 0 && s.Cursor < len(s.AdminFields) {
			field := s.AdminFields[s.Cursor]
			var datatypeID types.AdminDatatypeID
			if s.SelectedDTNode != nil && s.SelectedDTNode.AdminDT != nil {
				datatypeID = s.SelectedDTNode.AdminDT.AdminDatatypeID
			}
			return s, ShowDeleteAdminFieldDialogCmd(field.AdminFieldID, datatypeID, field.Label)
		}
	} else {
		if len(s.Fields) > 0 && s.Cursor < len(s.Fields) {
			field := s.Fields[s.Cursor]
			var datatypeID types.DatatypeID
			if s.SelectedDTNode != nil && s.SelectedDTNode.Datatype != nil {
				datatypeID = s.SelectedDTNode.Datatype.DatatypeID
			}
			return s, ShowDeleteFieldDialogCmd(field.FieldID, datatypeID, field.Label)
		}
	}
	return s, nil
}

func (s *DatatypesScreen) handleReorderDatatypeUp() (Screen, tea.Cmd) {
	if s.Cursor <= 0 || len(s.FlatDTList) == 0 {
		return s, nil
	}

	// Find siblings at the same depth/parent for reorder
	if s.AdminMode {
		a := s.FlatDTList[s.Cursor]
		b := s.FlatDTList[s.Cursor-1]
		if a.AdminDT != nil && b.AdminDT != nil {
			return s, tea.Batch(LoadingStartCmd(), ReorderAdminDatatypeCmd(
				a.AdminDT.AdminDatatypeID, b.AdminDT.AdminDatatypeID,
				a.AdminDT.SortOrder, b.AdminDT.SortOrder, "up",
			))
		}
	} else {
		a := s.FlatDTList[s.Cursor]
		b := s.FlatDTList[s.Cursor-1]
		if a.Datatype != nil && b.Datatype != nil {
			return s, tea.Batch(LoadingStartCmd(), ReorderDatatypeCmd(
				a.Datatype.DatatypeID, b.Datatype.DatatypeID,
				a.Datatype.SortOrder, b.Datatype.SortOrder, "up",
			))
		}
	}
	return s, nil
}

func (s *DatatypesScreen) handleReorderDatatypeDown() (Screen, tea.Cmd) {
	if s.Cursor >= len(s.FlatDTList)-1 || len(s.FlatDTList) == 0 {
		return s, nil
	}

	if s.AdminMode {
		a := s.FlatDTList[s.Cursor]
		b := s.FlatDTList[s.Cursor+1]
		if a.AdminDT != nil && b.AdminDT != nil {
			return s, tea.Batch(LoadingStartCmd(), ReorderAdminDatatypeCmd(
				a.AdminDT.AdminDatatypeID, b.AdminDT.AdminDatatypeID,
				a.AdminDT.SortOrder, b.AdminDT.SortOrder, "down",
			))
		}
	} else {
		a := s.FlatDTList[s.Cursor]
		b := s.FlatDTList[s.Cursor+1]
		if a.Datatype != nil && b.Datatype != nil {
			return s, tea.Batch(LoadingStartCmd(), ReorderDatatypeCmd(
				a.Datatype.DatatypeID, b.Datatype.DatatypeID,
				a.Datatype.SortOrder, b.Datatype.SortOrder, "down",
			))
		}
	}
	return s, nil
}

// Field reorder is only available in Phase 2.
func (s *DatatypesScreen) handleReorderFieldUp() (Screen, tea.Cmd) {
	if s.Cursor <= 0 {
		return s, nil
	}
	if s.AdminMode && len(s.AdminFields) > 1 && s.Cursor < len(s.AdminFields) {
		a := s.AdminFields[s.Cursor]
		b := s.AdminFields[s.Cursor-1]
		return s, tea.Batch(LoadingStartCmd(), ReorderAdminFieldCmd(
			a.AdminFieldID, b.AdminFieldID,
			a.SortOrder, b.SortOrder, "up",
		))
	}
	if !s.AdminMode && len(s.Fields) > 1 && s.Cursor < len(s.Fields) {
		a := s.Fields[s.Cursor]
		b := s.Fields[s.Cursor-1]
		return s, tea.Batch(LoadingStartCmd(), ReorderFieldCmd(
			a.FieldID, b.FieldID,
			a.SortOrder, b.SortOrder, "up",
		))
	}
	return s, nil
}

func (s *DatatypesScreen) handleReorderFieldDown() (Screen, tea.Cmd) {
	if s.AdminMode {
		if s.Cursor >= len(s.AdminFields)-1 {
			return s, nil
		}
		a := s.AdminFields[s.Cursor]
		b := s.AdminFields[s.Cursor+1]
		return s, tea.Batch(LoadingStartCmd(), ReorderAdminFieldCmd(
			a.AdminFieldID, b.AdminFieldID,
			a.SortOrder, b.SortOrder, "down",
		))
	}
	if s.Cursor >= len(s.Fields)-1 {
		return s, nil
	}
	a := s.Fields[s.Cursor]
	b := s.Fields[s.Cursor+1]
	return s, tea.Batch(LoadingStartCmd(), ReorderFieldCmd(
		a.FieldID, b.FieldID,
		a.SortOrder, b.SortOrder, "up",
	))
}

// ---------------------------------------------------------------------------
// Fetch handlers
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) handleAllDatatypesFetch(ctx AppContext) (Screen, tea.Cmd) {
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
}

func (s *DatatypesScreen) handleAllDatatypesFetchResults(msg AllDatatypesFetchResultsMsg) (Screen, tea.Cmd) {
	s.Datatypes = msg.Data
	s.rebuildTree()
	cmds := []tea.Cmd{LoadingStopCmd()}
	if s.Phase == DatatypesPhaseBrowse && len(s.FlatDTList) > 0 {
		cmds = append(cmds, s.fetchFieldsForCurrentDT())
	}
	return s, tea.Batch(cmds...)
}

func (s *DatatypesScreen) handleAdminAllDatatypesFetch(ctx AppContext) (Screen, tea.Cmd) {
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
}

func (s *DatatypesScreen) handleAdminAllDatatypesFetchResults(msg AdminAllDatatypesFetchResultsMsg) (Screen, tea.Cmd) {
	s.AdminDatatypes = msg.Data
	s.rebuildTree()
	cmds := []tea.Cmd{LoadingStopCmd()}
	if s.Phase == DatatypesPhaseBrowse && len(s.FlatDTList) > 0 {
		cmds = append(cmds, s.fetchFieldsForCurrentDT())
	}
	return s, tea.Batch(cmds...)
}

func (s *DatatypesScreen) handleDatatypeFieldsFetch(ctx AppContext, msg DatatypeFieldsFetchMsg) (Screen, tea.Cmd) {
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
}

func (s *DatatypesScreen) handleDatatypeFieldsFetchResults(msg DatatypeFieldsFetchResultsMsg) (Screen, tea.Cmd) {
	s.Fields = msg.Fields
	if s.Phase == DatatypesPhaseFields {
		s.CursorMax = len(s.Fields) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		if s.Cursor > s.CursorMax {
			s.Cursor = s.CursorMax
		}
		s.rebuildProperties()
	}
	return s, nil
}

func (s *DatatypesScreen) handleAdminDatatypeFieldsFetch(ctx AppContext, msg AdminDatatypeFieldsFetchMsg) (Screen, tea.Cmd) {
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
}

func (s *DatatypesScreen) handleAdminDatatypeFieldsFetchResults(msg AdminDatatypeFieldsFetchResultsMsg) (Screen, tea.Cmd) {
	s.AdminFields = msg.Fields
	if s.Phase == DatatypesPhaseFields {
		s.CursorMax = len(s.AdminFields) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		if s.Cursor > s.CursorMax {
			s.Cursor = s.CursorMax
		}
		s.rebuildProperties()
	}
	return s, nil
}

func (s *DatatypesScreen) handleDatatypeReordered(msg DatatypeReorderedMsg) (Screen, tea.Cmd) {
	if msg.Direction == "up" && s.Cursor > 0 {
		s.Cursor--
	} else if msg.Direction == "down" && s.Cursor < s.CursorMax {
		s.Cursor++
	}
	return s, tea.Batch(LoadingStopCmd(), AllDatatypesFetchCmd())
}

func (s *DatatypesScreen) handleAdminDatatypeReordered(msg AdminDatatypeReorderedMsg) (Screen, tea.Cmd) {
	if msg.Direction == "up" && s.Cursor > 0 {
		s.Cursor--
	} else if msg.Direction == "down" && s.Cursor < s.CursorMax {
		s.Cursor++
	}
	return s, tea.Batch(LoadingStopCmd(), AdminAllDatatypesFetchCmd())
}

func (s *DatatypesScreen) handleFieldReordered(msg DatatypeFieldReorderedMsg) (Screen, tea.Cmd) {
	if msg.Direction == "up" && s.Cursor > 0 {
		s.Cursor--
	} else if msg.Direction == "down" && s.Cursor < s.CursorMax {
		s.Cursor++
	}
	// Re-fetch fields for the selected datatype
	if s.SelectedDTNode != nil && s.SelectedDTNode.Datatype != nil {
		return s, tea.Batch(LoadingStopCmd(), DatatypeFieldsFetchCmd(s.SelectedDTNode.Datatype.DatatypeID))
	}
	return s, LoadingStopCmd()
}

func (s *DatatypesScreen) handleAdminFieldReordered(msg AdminDatatypeFieldReorderedMsg) (Screen, tea.Cmd) {
	if msg.Direction == "up" && s.Cursor > 0 {
		s.Cursor--
	} else if msg.Direction == "down" && s.Cursor < s.CursorMax {
		s.Cursor++
	}
	if s.SelectedDTNode != nil && s.SelectedDTNode.AdminDT != nil {
		return s, tea.Batch(LoadingStopCmd(), AdminDatatypeFieldsFetchCmd(s.SelectedDTNode.AdminDT.AdminDatatypeID))
	}
	return s, LoadingStopCmd()
}
