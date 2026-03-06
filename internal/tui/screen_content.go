package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// Grid definitions for content screen phases.
var contentSelectGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Content"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.65, Title: "Details"},
			{Height: 0.35, Title: "Stats"},
		}},
	},
}

var contentTreeGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Tree"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 1.0, Title: "Preview"},
		}},
	},
}

// ContentScreen implements Screen for both CONTENT and ADMINCONTENT pages.
// It supports multiple phases: route list, tree browsing, field editing,
// and version management. AdminMode selects admin vs regular data and IDs.
type ContentScreen struct {
	GridScreen
	AdminMode bool

	// Cursors
	Cursor int

	// Select phase (slug tree)
	SelectTree     []*ContentSelectNode
	FlatSelectList []*ContentSelectNode

	// Route list phase
	RootContentSummary      []db.ContentDataTopLevel
	AdminRootContentSummary []db.AdminContentDataTopLevel
	RootDatatypes           []db.Datatypes
	AdminRootDatatypes      []db.AdminDatatypes

	// Tree browsing phase (entered after selecting a route)
	PageRouteId            types.RouteID
	AdminPageRouteId       types.AdminRouteID
	Root                   tree.Root
	SelectedDatatypeID     types.DatatypeID
	PendingCursorContentID types.ContentID

	// Batch-loaded fields for all tree nodes (used by grid preview)
	AllFields      map[types.ContentID][]ContentFieldDisplay
	AllAdminFields map[types.AdminContentID][]AdminContentFieldDisplay

	// Version state
	ShowVersionList       bool
	Versions              []db.ContentVersion
	AdminVersions         []db.AdminContentVersion
	VersionCursor         int
	VersionContentID      types.ContentID
	AdminVersionContentID types.AdminContentID
	VersionRouteID        types.RouteID
	AdminVersionRouteID   types.AdminRouteID

	// Error state
	LastError    error
	ErrorContext string
}

// NewContentScreen creates a ContentScreen for regular or admin mode.
func NewContentScreen(
	adminMode bool,
	rootContentSummary []db.ContentDataTopLevel,
	adminRootContentSummary []db.AdminContentDataTopLevel,
	rootDatatypes []db.Datatypes,
	adminRootDatatypes []db.AdminDatatypes,
	pageRouteId types.RouteID,
) *ContentScreen {
	s := &ContentScreen{
		GridScreen: GridScreen{
			Grid: contentSelectGrid,
		},
		AdminMode:               adminMode,
		Cursor:                  0,
		RootContentSummary:      rootContentSummary,
		AdminRootContentSummary: adminRootContentSummary,
		RootDatatypes:           rootDatatypes,
		AdminRootDatatypes:      adminRootDatatypes,
		PageRouteId:             pageRouteId,
	}
	s.rebuildSelectTree()
	return s
}

func (s *ContentScreen) KeyHints(km config.KeyMap) []KeyHint {
	if s.ShowVersionList {
		return []KeyHint{
			{km.HintString(config.ActionSelect), "restore"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	}
	if s.inTreePhase() {
		return []KeyHint{
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionSelect), "expand"},
			{km.HintString(config.ActionEdit), "edit"},
			{km.HintString(config.ActionNew), "new"},
			{km.HintString(config.ActionDelete), "del"},
			{km.HintString(config.ActionCopy), "copy"},
			{km.HintString(config.ActionMove), "move"},
			{km.HintString(config.ActionPublish), "publish"},
			{km.HintString(config.ActionVersions), "versions"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	}
	// Select phase
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionSelect), "select"},
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *ContentScreen) PageIndex() PageIndex {
	if s.AdminMode {
		return ADMINCONTENT
	}
	return CONTENT
}

func (s *ContentScreen) setError(context string, err error) {
	s.LastError = err
	s.ErrorContext = context
}

func (s *ContentScreen) clearError() {
	s.LastError = nil
	s.ErrorContext = ""
}

// inTreePhase returns true when a tree has been loaded (user selected a route).
func (s *ContentScreen) inTreePhase() bool {
	return s.Root.Root != nil
}

// inRouteListPhase returns true when showing the route/content list (no tree loaded).
func (s *ContentScreen) inRouteListPhase() bool {
	if s.AdminMode {
		return !s.inTreePhase() && s.AdminPageRouteId.IsZero()
	}
	return !s.inTreePhase() && s.PageRouteId.IsZero()
}

// visibleNodeCount returns the number of visible nodes in the tree.
func (s *ContentScreen) visibleNodeCount() int {
	if s.Root.Root == nil {
		return 0
	}
	return len(s.Root.FlattenVisible())
}

// Update processes messages for the ContentScreen.
func (s *ContentScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Clear error on any key press
		if s.LastError != nil {
			s.clearError()
		}

		// Version list mode
		if s.ShowVersionList {
			return s.handleVersionListKeys(ctx, key, km)
		}

		// Tree phase
		if s.inTreePhase() {
			return s.handleTreeKeys(ctx, key, km)
		}

		// Route list phase
		return s.handleRouteListKeys(ctx, key, km)

	// Fetch request messages
	case RootContentSummaryFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			summary, err := d.ListContentDataTopLevelPaginated(db.PaginationParams{Limit: 10000, Offset: 0})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if summary == nil {
				return RootContentSummaryFetchResultsMsg{Data: []db.ContentDataTopLevel{}}
			}
			return RootContentSummaryFetchResultsMsg{Data: *summary}
		}
	case RootContentSummaryFetchResultsMsg:
		s.RootContentSummary = msg.Data
		s.Cursor = 0
		s.rebuildSelectTree()
		return s, LoadingStopCmd()
	case AdminContentDataFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			contentData, err := d.ListAdminContentDataTopLevelPaginated(db.PaginationParams{Limit: 10000, Offset: 0})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if contentData == nil {
				return AdminContentDataFetchResultsMsg{Data: []db.AdminContentDataTopLevel{}}
			}
			return AdminContentDataFetchResultsMsg{Data: *contentData}
		}
	case AdminContentDataFetchResultsMsg:
		s.AdminRootContentSummary = msg.Data
		s.Cursor = 0
		s.rebuildSelectTree()
		return s, LoadingStopCmd()
	case RootDatatypesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			datatypes, err := d.ListDatatypesRoot()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if datatypes == nil {
				return RootDatatypesFetchResultsMsg{Data: []db.Datatypes{}}
			}
			return RootDatatypesFetchResultsMsg{Data: *datatypes}
		}
	case RootDatatypesFetchResultsMsg:
		s.RootDatatypes = msg.Data
		return s, LoadingStopCmd()
	case FetchChildDatatypesMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		routeID := msg.RouteID
		rootDatatypeID := msg.ParentDatatypeID
		return s, func() tea.Msg {
			all, err := d.ListDatatypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if all == nil || len(*all) == 0 {
				return ActionResultMsg{
					Title:   "No Datatypes",
					Message: "No datatypes are defined.",
				}
			}
			filtered := filterChildDatatypes(*all, rootDatatypeID)
			if len(filtered) == 0 {
				return ActionResultMsg{
					Title:   "No Datatypes",
					Message: "No eligible child datatypes for this root type.",
				}
			}
			return ShowChildDatatypeDialogMsg{
				ChildDatatypes: filtered,
				RouteID:        string(routeID),
			}
		}
	case FetchContentFieldsMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		datatypeID := msg.DatatypeID
		routeID := msg.RouteID
		parentID := msg.ParentID
		title := msg.Title
		return s, func() tea.Msg {
			fieldList, err := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: datatypeID, Valid: true})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			var fields []db.Fields
			if fieldList != nil {
				fields = *fieldList
			}
			if len(fields) == 0 {
				return ActionResultMsg{
					Title:   "No Fields",
					Message: "This datatype has no fields defined.",
				}
			}
			return ShowContentFormDialogMsg{
				Action:     FORMDIALOGCREATECONTENT,
				Title:      title,
				DatatypeID: datatypeID,
				RouteID:    routeID,
				ParentID:   parentID,
				Fields:     fields,
			}
		}
	case RoutesByDatatypeFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		datatypeID := msg.DatatypeID
		return s, func() tea.Msg {
			routes, err := d.ListRoutesByDatatype(datatypeID)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if routes == nil {
				return RoutesFetchResultsMsg{Data: []db.Routes{}}
			}
			return RoutesFetchResultsMsg{Data: *routes}
		}

	// Data refresh messages (from CMS operations)
	case RootContentSummarySet:
		s.RootContentSummary = msg.RootContentSummary
		s.rebuildSelectTree()
		return s, nil
	case AdminContentDataSet:
		s.AdminRootContentSummary = msg.AdminContentData
		s.rebuildSelectTree()
		return s, nil
	case RootDatatypesSet:
		s.RootDatatypes = msg.RootDatatypes
		return s, nil
	case AdminRootDatatypesFetchResultsMsg:
		s.AdminRootDatatypes = msg.RootDatatypes
		return s, nil

	// Tree loading results
	case TreeLoadedMsg:
		s.clearError()
		if msg.RootNode != nil {
			s.Root = *msg.RootNode
		}
		if !s.PendingCursorContentID.IsZero() {
			idx := s.Root.FindVisibleIndex(s.PendingCursorContentID)
			if idx >= 0 {
				s.Cursor = idx
			}
			s.PendingCursorContentID = types.ContentID("")
		} else {
			s.Cursor = 0
		}
		return s, BatchLoadContentFieldsCmd(ctx.Config, s.PageRouteId, s.collectDatatypeIDs(), ctx.ActiveLocale)

	case AdminTreeLoadedMsg:
		s.clearError()
		if msg.RootNode != nil {
			s.Root = *msg.RootNode
		}
		if !s.PendingCursorContentID.IsZero() {
			idx := s.Root.FindVisibleIndex(s.PendingCursorContentID)
			if idx >= 0 {
				s.Cursor = idx
			}
			s.PendingCursorContentID = types.ContentID("")
		} else {
			s.Cursor = 0
		}
		return s, BatchLoadAdminContentFieldsCmd(ctx.Config, s.AdminPageRouteId, s.collectAdminDatatypeIDs(), ctx.ActiveLocale)

	// Batch field loading results
	case BatchContentFieldsLoadedMsg:
		s.AllFields = msg.Fields
		return s, nil
	case BatchAdminContentFieldsLoadedMsg:
		s.AllAdminFields = msg.Fields
		return s, nil

	// Content CRUD completion messages → reload tree
	case ContentCreatedMsg:
		s.clearError()
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)
	case AdminContentCreatedMsg:
		s.clearError()
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)
	case ContentCreatedWithErrorsMsg:
		s.clearError()
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)

	case AdminContentUpdatedFromDialogMsg:
		s.clearError()
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)

	case AdminContentDeletedMsg:
		s.clearError()
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)

	case ContentReorderedMsg:
		s.clearError()
		s.PendingCursorContentID = msg.ContentID
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)
	case AdminContentReorderedMsg:
		s.clearError()
		s.PendingCursorContentID = types.ContentID(msg.AdminContentID)
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)

	case ContentCopiedMsg:
		s.clearError()
		s.PendingCursorContentID = msg.NewContentID
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)
	case AdminContentCopiedMsg:
		s.clearError()
		s.PendingCursorContentID = types.ContentID(msg.NewID)
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)

	case AdminContentMovedMsg:
		s.clearError()
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)

	// Publish/Unpublish completion
	case PublishCompletedMsg:
		s.clearError()
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)
	case AdminPublishCompletedMsg:
		s.clearError()
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)
	case UnpublishCompletedMsg:
		s.clearError()
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)
	case AdminUnpublishCompletedMsg:
		s.clearError()
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)

	// Version list results
	case VersionsListedMsg:
		s.clearError()
		s.ShowVersionList = true
		s.Versions = msg.Versions
		s.VersionCursor = 0
		s.VersionContentID = msg.ContentID
		s.VersionRouteID = msg.RouteID
		return s, nil
	case AdminVersionsListedMsg:
		s.clearError()
		s.ShowVersionList = true
		s.AdminVersions = msg.Versions
		s.VersionCursor = 0
		s.AdminVersionContentID = msg.AdminContentID
		s.AdminVersionRouteID = msg.AdminRouteID
		return s, nil

	// Version restore results
	case VersionRestoredMsg:
		s.clearError()
		s.ShowVersionList = false
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)
	case AdminVersionRestoredMsg:
		s.clearError()
		s.ShowVersionList = false
		return s, ReloadAdminContentTreeCmd(ctx.Config, msg.AdminRouteID)

	// Content field operation results → reload batch fields
	case AdminContentFieldUpdatedMsg:
		s.clearError()
		return s, BatchLoadAdminContentFieldsCmd(ctx.Config, s.AdminPageRouteId, s.collectAdminDatatypeIDs(), ctx.ActiveLocale)
	case AdminContentFieldAddedMsg:
		s.clearError()
		return s, BatchLoadAdminContentFieldsCmd(ctx.Config, s.AdminPageRouteId, s.collectAdminDatatypeIDs(), ctx.ActiveLocale)
	case AdminContentFieldDeletedMsg:
		s.clearError()
		return s, BatchLoadAdminContentFieldsCmd(ctx.Config, s.AdminPageRouteId, s.collectAdminDatatypeIDs(), ctx.ActiveLocale)

	// Content publish toggled (legacy compat)
	case ContentPublishToggledMsg:
		s.clearError()
		return s, ReloadContentTreeCmd(ctx.Config, msg.RouteID)
	}

	return s, nil
}

// handleRouteListKeys handles key input in the route/content list phase.
func (s *ContentScreen) handleRouteListKeys(ctx AppContext, key string, km config.KeyMap) (Screen, tea.Cmd) {
	if km.Matches(key, config.ActionQuit) {
		return s, tea.Quit
	}
	if km.Matches(key, config.ActionDismiss) || km.Matches(key, config.ActionBack) {
		return s, HistoryPopCmd()
	}

	// Cursor movement (using FlatSelectList)
	if km.Matches(key, config.ActionUp) {
		if s.Cursor > 0 {
			s.Cursor--
		}
		return s, nil
	}
	if km.Matches(key, config.ActionDown) {
		max := len(s.FlatSelectList) - 1
		if max < 0 {
			max = 0
		}
		if s.Cursor < max {
			s.Cursor++
		}
		return s, nil
	}

	// Expand/collapse group nodes
	if km.Matches(key, config.ActionExpand) || km.Matches(key, config.ActionCollapse) {
		if s.Cursor < len(s.FlatSelectList) {
			node := s.FlatSelectList[s.Cursor]
			if node.Kind == NodeGroup || node.hasChildren() {
				node.Expand = !node.Expand
				s.FlatSelectList = FlattenSelectTree(s.SelectTree)
			}
		}
		return s, nil
	}

	// Select content → load tree
	if km.Matches(key, config.ActionSelect) {
		return s.handleSelectPhaseEnter(ctx)
	}

	// New content
	if km.Matches(key, config.ActionNew) {
		if s.AdminMode {
			return s, nil
		}
		if len(s.RootDatatypes) == 0 {
			return s, ShowDialog("Info", "No root datatypes available", false)
		}
		return s, ShowCreateRouteWithContentDialogCmd(s.RootDatatypes)
	}

	// Edit content from select list
	if km.Matches(key, config.ActionEdit) {
		if s.Cursor < len(s.FlatSelectList) {
			node := s.FlatSelectList[s.Cursor]
			if node.Content != nil && node.Content.RouteID.Valid && node.Content.DatatypeID.Valid {
				return s, FetchContentForEditCmd(
					node.Content.ContentDataID,
					node.Content.DatatypeID.ID,
					node.Content.RouteID.ID,
					fmt.Sprintf("Edit: %s", node.Content.RouteTitle),
				)
			}
		}
		return s, nil
	}

	// Delete content from select list
	if km.Matches(key, config.ActionDelete) {
		if s.Cursor < len(s.FlatSelectList) {
			node := s.FlatSelectList[s.Cursor]
			if s.AdminMode && node.AdminContent != nil {
				hasChildren := node.AdminContent.FirstChildID.Valid
				return s, ShowDeleteAdminContentDialogCmd(
					node.AdminContent.AdminContentDataID,
					node.AdminContent.AdminRouteID.ID,
					node.AdminContent.DatatypeLabel,
					hasChildren,
				)
			}
		}
		return s, nil
	}

	return s, nil
}

// handleSelectPhaseEnter handles selecting an item from the slug tree to enter tree phase.
func (s *ContentScreen) handleSelectPhaseEnter(ctx AppContext) (Screen, tea.Cmd) {
	if s.Cursor >= len(s.FlatSelectList) {
		return s, nil
	}
	node := s.FlatSelectList[s.Cursor]

	// Toggle expand/collapse on group nodes
	if node.Kind == NodeGroup && (node.Content == nil && node.AdminContent == nil) {
		node.Expand = !node.Expand
		s.FlatSelectList = FlattenSelectTree(s.SelectTree)
		return s, nil
	}

	// Regular content
	if node.Content != nil {
		if !node.Content.RouteID.Valid {
			return s, nil
		}
		s.PageRouteId = node.Content.RouteID.ID
		if node.Content.DatatypeID.Valid {
			s.SelectedDatatypeID = node.Content.DatatypeID.ID
		}
		s.Cursor = 0
		s.Grid = contentTreeGrid
		return s, tea.Batch(
			LoadingStartCmd(),
			ReloadContentTreeCmd(ctx.Config, node.Content.RouteID.ID),
		)
	}

	// Admin content
	if node.AdminContent != nil {
		if !node.AdminContent.AdminRouteID.Valid {
			return s, nil
		}
		s.AdminPageRouteId = node.AdminContent.AdminRouteID.ID
		s.Cursor = 0
		s.Grid = contentTreeGrid
		return s, tea.Batch(
			LoadingStartCmd(),
			ReloadAdminContentTreeCmd(ctx.Config, node.AdminContent.AdminRouteID.ID),
		)
	}

	return s, nil
}

// handleTreeKeys handles key input in the tree browsing phase.
func (s *ContentScreen) handleTreeKeys(ctx AppContext, key string, km config.KeyMap) (Screen, tea.Cmd) {
	if km.Matches(key, config.ActionQuit) {
		return s, tea.Quit
	}

	// Back: tree → route list
	if km.Matches(key, config.ActionDismiss) || km.Matches(key, config.ActionBack) {
		return s.handleBackKey(ctx)
	}

	// Cursor up/down
	if km.Matches(key, config.ActionUp) {
		if s.Cursor > 0 {
			s.Cursor--
		}
		return s, nil
	}
	if km.Matches(key, config.ActionDown) {
		max := s.visibleNodeCount() - 1
		if max < 0 {
			max = 0
		}
		if s.Cursor < max {
			s.Cursor++
		}
		return s, nil
	}

	// Expand/Collapse
	if km.Matches(key, config.ActionSelect) {
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.FirstChild != nil {
			node.Expand = !node.Expand
			return s, nil
		}
	}
	if km.Matches(key, config.ActionExpand) {
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.FirstChild != nil {
			node.Expand = true
			return s, nil
		}
	}
	if km.Matches(key, config.ActionCollapse) {
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.FirstChild != nil {
			node.Expand = false
			return s, nil
		}
	}

	// Navigate to parent
	if km.Matches(key, config.ActionGoParent) {
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.Parent != nil && node.Parent.Instance != nil {
			idx := s.Root.FindVisibleIndex(node.Parent.Instance.ContentDataID)
			if idx >= 0 {
				s.Cursor = idx
			}
		}
		return s, nil
	}
	// Navigate to first child
	if km.Matches(key, config.ActionGoChild) {
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.FirstChild != nil {
			node.Expand = true
			if node.FirstChild.Instance != nil {
				idx := s.Root.FindVisibleIndex(node.FirstChild.Instance.ContentDataID)
				if idx >= 0 {
					s.Cursor = idx
				}
			}
		}
		return s, nil
	}

	// New content node
	if km.Matches(key, config.ActionNew) {
		if s.AdminMode {
			return s.handleAdminTreeNew(ctx)
		}
		return s.handleRegularTreeNew(ctx)
	}

	// Edit content node
	if km.Matches(key, config.ActionEdit) {
		if s.AdminMode {
			return s.handleAdminTreeEdit(ctx)
		}
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.Instance != nil {
			return s, FetchContentForEditCmd(
				node.Instance.ContentDataID,
				node.Datatype.DatatypeID,
				s.PageRouteId,
				fmt.Sprintf("Edit: %s", node.Datatype.Label),
			)
		}
		return s, ShowDialog("Error", "Please select a content node first", false)
	}

	// Delete content node
	if km.Matches(key, config.ActionDelete) {
		if s.AdminMode {
			return s.handleAdminTreeDelete(ctx)
		}
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.Instance != nil {
			contentName := DecideNodeName(*node)
			hasChildren := node.FirstChild != nil
			return s, ShowDeleteContentDialogCmd(
				string(node.Instance.ContentDataID),
				contentName,
				hasChildren,
			)
		}
		return s, ShowDialog("Error", "Please select a content node first", false)
	}

	// Move content
	if km.Matches(key, config.ActionMove) {
		if s.AdminMode {
			return s.handleAdminTreeMove(ctx)
		}
		return s.handleRegularTreeMove(ctx)
	}

	// Reorder up/down
	if km.Matches(key, config.ActionReorderUp) {
		if s.AdminMode {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil && node.PrevSibling != nil {
				adminID := types.AdminContentID(node.Instance.ContentDataID)
				return s, tea.Batch(LoadingStartCmd(), AdminReorderSiblingCmd(adminID, s.AdminPageRouteId, "up"))
			}
		} else {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil && node.PrevSibling != nil {
				return s, tea.Batch(LoadingStartCmd(), ReorderSiblingCmd(node.Instance.ContentDataID, s.PageRouteId, "up"))
			}
		}
		return s, nil
	}
	if km.Matches(key, config.ActionReorderDown) {
		if s.AdminMode {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil && node.NextSibling != nil {
				adminID := types.AdminContentID(node.Instance.ContentDataID)
				return s, tea.Batch(LoadingStartCmd(), AdminReorderSiblingCmd(adminID, s.AdminPageRouteId, "down"))
			}
		} else {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil && node.NextSibling != nil {
				return s, tea.Batch(LoadingStartCmd(), ReorderSiblingCmd(node.Instance.ContentDataID, s.PageRouteId, "down"))
			}
		}
		return s, nil
	}

	// Copy
	if km.Matches(key, config.ActionCopy) {
		if s.AdminMode {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil {
				adminID := types.AdminContentID(node.Instance.ContentDataID)
				return s, tea.Batch(LoadingStartCmd(), AdminCopyContentCmd(adminID, s.AdminPageRouteId))
			}
		} else {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil {
				return s, tea.Batch(LoadingStartCmd(), CopyContentCmd(node.Instance.ContentDataID, s.PageRouteId))
			}
		}
		return s, nil
	}

	// Publish
	if km.Matches(key, config.ActionPublish) {
		if s.AdminMode {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil {
				adminID := types.AdminContentID(node.Instance.ContentDataID)
				name := DecideNodeName(*node)
				isPublished := node.Instance.Status == types.ContentStatusPublished
				return s, ShowPublishAdminContentDialogCmd(adminID, s.AdminPageRouteId, name, isPublished)
			}
		} else {
			if s.Root.Root != nil && s.Root.Root.Instance != nil {
				return s, TogglePublishCmd(s.Root.Root.Instance.ContentDataID, s.PageRouteId)
			}
		}
		return s, nil
	}

	// Versions
	if km.Matches(key, config.ActionVersions) {
		if s.AdminMode {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil {
				adminID := types.AdminContentID(node.Instance.ContentDataID)
				return s, tea.Batch(LoadingStartCmd(), AdminListVersionsCmd(adminID, s.AdminPageRouteId))
			}
		} else {
			node := s.Root.NodeAtIndex(s.Cursor)
			if node != nil && node.Instance != nil {
				return s, tea.Batch(LoadingStartCmd(), ListVersionsCmd(node.Instance.ContentDataID, s.PageRouteId))
			}
		}
		return s, nil
	}

	// Locale switch
	if km.Matches(key, config.ActionLocale) {
		if ctx.Config.I18nEnabled() && ctx.DB != nil {
			return s, LoadEnabledLocalesCmd(ctx.DB)
		}
	}

	return s, nil
}

// handleVersionListKeys handles key input when viewing the version list.
func (s *ContentScreen) handleVersionListKeys(ctx AppContext, key string, km config.KeyMap) (Screen, tea.Cmd) {
	if km.Matches(key, config.ActionQuit) {
		return s, tea.Quit
	}

	// Dismiss/back: close version list
	if km.Matches(key, config.ActionDismiss) || km.Matches(key, config.ActionBack) {
		s.ShowVersionList = false
		return s, nil
	}

	// Cursor movement
	if km.Matches(key, config.ActionUp) {
		if s.VersionCursor > 0 {
			s.VersionCursor--
		}
		return s, nil
	}
	if km.Matches(key, config.ActionDown) {
		max := s.versionListLen() - 1
		if max < 0 {
			max = 0
		}
		if s.VersionCursor < max {
			s.VersionCursor++
		}
		return s, nil
	}

	// Select version → restore
	if km.Matches(key, config.ActionSelect) {
		if s.AdminMode {
			if s.VersionCursor < len(s.AdminVersions) {
				v := s.AdminVersions[s.VersionCursor]
				return s, ShowRestoreAdminVersionDialogCmd(
					s.AdminVersionContentID,
					v.AdminContentVersionID,
					s.AdminVersionRouteID,
					v.VersionNumber,
				)
			}
		} else {
			if s.VersionCursor < len(s.Versions) {
				v := s.Versions[s.VersionCursor]
				return s, ShowRestoreVersionDialogCmd(
					s.VersionContentID,
					v.ContentVersionID,
					s.VersionRouteID,
					v.VersionNumber,
				)
			}
		}
	}

	return s, nil
}

// handleBackKey navigates back from tree to route list, or pops history.
func (s *ContentScreen) handleBackKey(ctx AppContext) (Screen, tea.Cmd) {
	if s.inTreePhase() {
		// Go back to select phase
		s.Root = tree.Root{}
		s.PageRouteId = types.RouteID("")
		s.AdminPageRouteId = types.AdminRouteID("")
		s.AllFields = nil
		s.AllAdminFields = nil
		s.Cursor = 0
		s.Grid = contentSelectGrid
		s.FocusIndex = 0
		return s, nil
	}
	return s, HistoryPopCmd()
}

// versionListLen returns the number of versions for the current mode.
func (s *ContentScreen) versionListLen() int {
	if s.AdminMode {
		return len(s.AdminVersions)
	}
	return len(s.Versions)
}

// handleRegularTreeNew handles creating new content in regular mode.
func (s *ContentScreen) handleRegularTreeNew(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil {
		return s, ShowDialog("Error", "Please select a content node first", false)
	}
	rootDatatypeID := node.Datatype.DatatypeID
	if s.Root.Root != nil {
		rootDatatypeID = s.Root.Root.Datatype.DatatypeID
	}
	return s, ShowChildDatatypeDialogCmd(rootDatatypeID, s.PageRouteId)
}

// handleAdminTreeNew handles creating new content in admin mode.
func (s *ContentScreen) handleAdminTreeNew(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil {
		return s, ShowDialog("Error", "Please select a content node first", false)
	}
	dtID := types.AdminDatatypeID(node.Datatype.DatatypeID)
	return s, AdminFetchFieldsForFormCmd(ctx.DB, dtID, s.AdminPageRouteId)
}

// handleAdminTreeEdit handles editing content in admin mode.
func (s *ContentScreen) handleAdminTreeEdit(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, ShowDialog("Error", "Please select a content node first", false)
	}
	adminContentID := types.AdminContentID(node.Instance.ContentDataID)
	adminDatatypeID := types.AdminDatatypeID(node.Datatype.DatatypeID)
	title := fmt.Sprintf("Edit: %s", node.Datatype.Label)
	return s, func() tea.Msg {
		return AdminFetchContentForEditMsg{
			AdminContentID:  adminContentID,
			AdminDatatypeID: adminDatatypeID,
			AdminRouteID:    s.AdminPageRouteId,
			Title:           title,
		}
	}
}

// handleAdminTreeDelete handles deleting content in admin mode.
func (s *ContentScreen) handleAdminTreeDelete(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, ShowDialog("Error", "Please select a content node first", false)
	}
	adminID := types.AdminContentID(node.Instance.ContentDataID)
	contentName := DecideNodeName(*node)
	hasChildren := node.FirstChild != nil
	return s, ShowDeleteAdminContentDialogCmd(adminID, s.AdminPageRouteId, contentName, hasChildren)
}

// handleRegularTreeMove handles moving content in regular mode.
func (s *ContentScreen) handleRegularTreeMove(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, ShowDialog("Error", "Please select a content node first", false)
	}
	allVisible := s.Root.FlattenVisible()
	targets := make([]ParentOption, 0)
	for _, candidate := range allVisible {
		if candidate.Instance == nil {
			continue
		}
		if candidate.Instance.ContentDataID == node.Instance.ContentDataID {
			continue
		}
		if tree.IsDescendantOf(candidate, node) {
			continue
		}
		label := DecideNodeName(*candidate)
		targets = append(targets, ParentOption{
			Label: label,
			Value: string(candidate.Instance.ContentDataID),
		})
	}
	if len(targets) == 0 {
		return s, ShowDialog("Cannot Move", "No valid move targets", false)
	}
	return s, ShowMoveContentDialogCmd(node, s.PageRouteId, targets)
}

// handleAdminTreeMove handles moving content in admin mode.
func (s *ContentScreen) handleAdminTreeMove(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, ShowDialog("Error", "Please select a content node first", false)
	}
	allVisible := s.Root.FlattenVisible()
	targets := make([]ParentOption, 0)
	for _, candidate := range allVisible {
		if candidate.Instance == nil {
			continue
		}
		if candidate.Instance.ContentDataID == node.Instance.ContentDataID {
			continue
		}
		if tree.IsDescendantOf(candidate, node) {
			continue
		}
		label := DecideNodeName(*candidate)
		targets = append(targets, ParentOption{
			Label: label,
			Value: string(candidate.Instance.ContentDataID),
		})
	}
	if len(targets) == 0 {
		return s, ShowDialog("Cannot Move", "No valid move targets", false)
	}
	return s, ShowMoveAdminContentDialogCmd(node, s.AdminPageRouteId, targets)
}

// rebuildSelectTree rebuilds the slug-grouped select tree and flat cursor list
// from the current content summary data.
func (s *ContentScreen) rebuildSelectTree() {
	if s.AdminMode {
		s.SelectTree = BuildAdminContentSelectTree(s.AdminRootContentSummary)
	} else {
		s.SelectTree = BuildContentSelectTree(s.RootContentSummary)
	}
	s.FlatSelectList = FlattenSelectTree(s.SelectTree)
}

// collectDatatypeIDs returns distinct DatatypeIDs from all nodes in the tree.
func (s *ContentScreen) collectDatatypeIDs() []types.DatatypeID {
	if s.Root.Root == nil {
		return nil
	}
	seen := make(map[types.DatatypeID]struct{})
	var ids []types.DatatypeID
	for _, n := range s.Root.NodeIndex {
		dtID := n.Datatype.DatatypeID
		if _, ok := seen[dtID]; !ok {
			seen[dtID] = struct{}{}
			ids = append(ids, dtID)
		}
	}
	return ids
}

// collectAdminDatatypeIDs returns distinct AdminDatatypeIDs from all tree nodes.
func (s *ContentScreen) collectAdminDatatypeIDs() []types.AdminDatatypeID {
	if s.Root.Root == nil {
		return nil
	}
	seen := make(map[types.DatatypeID]struct{})
	var ids []types.AdminDatatypeID
	for _, n := range s.Root.NodeIndex {
		dtID := n.Datatype.DatatypeID
		if _, ok := seen[dtID]; !ok {
			seen[dtID] = struct{}{}
			ids = append(ids, types.AdminDatatypeID(dtID))
		}
	}
	return ids
}
