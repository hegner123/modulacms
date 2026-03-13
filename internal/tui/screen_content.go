package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
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
	TitleMap       map[string]string // ContentDataID → resolved _title field value

	// Route list phase
	RootContentSummary      []db.ContentDataTopLevel
	AdminRootContentSummary []db.AdminContentDataTopLevel
	RootDatatypes           []db.Datatypes
	AdminRootDatatypes      []db.AdminDatatypes

	// Tree browsing phase (entered after selecting a route or root_id)
	PageRouteId      types.RouteID
	AdminPageRouteId types.AdminRouteID
	PageRootId       types.ContentID
	AdminPageRootId  types.AdminContentID
	Root             tree.Root
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
			{km.HintString(config.ActionExpand) + "/" + km.HintString(config.ActionCollapse), "+/-"},
			{km.HintString(config.ActionEdit), "edit"},
			{km.HintString(config.ActionNew), "new"},
			{km.HintString(config.ActionDelete), "del"},
			{km.HintString(config.ActionCopy), "copy"},
			{km.HintString(config.ActionMove), "move"},
			{km.HintString(config.ActionPublish), "publish"},
			{km.HintString(config.ActionReorderUp) + "/" + km.HintString(config.ActionReorderDown), "reorder"},
			{km.HintString(config.ActionVersions), "versions"},
			{km.HintString(config.ActionBack), "back"},
			{km.HintString(config.ActionQuit), "quit"},
		}
	}
	// Select phase
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionSelect), "select"},
		{km.HintString(config.ActionExpand) + "/" + km.HintString(config.ActionCollapse), "+/-"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionPublish), "publish"},
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
		return !s.inTreePhase() && s.AdminPageRouteId.IsZero() && s.AdminPageRootId.IsZero()
	}
	return !s.inTreePhase() && s.PageRouteId.IsZero() && s.PageRootId.IsZero()
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
	case tea.KeyPressMsg:
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
			titleMap := resolveTitleFields(d, *summary)
			return RootContentSummaryFetchResultsMsg{Data: *summary, TitleMap: titleMap}
		}
	case RootContentSummaryFetchResultsMsg:
		s.RootContentSummary = msg.Data
		s.TitleMap = msg.TitleMap
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
			titleMap := resolveAdminTitleFields(d, *contentData)
			return AdminContentDataFetchResultsMsg{Data: *contentData, TitleMap: titleMap}
		}
	case AdminContentDataFetchResultsMsg:
		s.AdminRootContentSummary = msg.Data
		s.TitleMap = msg.TitleMap
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
	case ChildDatatypeSelectedMsg:
		// User selected a child datatype — fetch its fields to build the content form.
		// The parent content ID comes from the currently selected tree node.
		var parentID types.NullableContentID
		node := s.Root.NodeAtIndex(s.Cursor)
		if node != nil && node.Instance != nil {
			parentID = types.NullableContentID{ID: node.Instance.ContentDataID, Valid: true}
		}
		return s, FetchContentFieldsCmd(
			msg.DatatypeID,
			msg.RouteID,
			parentID,
			"New Content",
		)
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
		return s, tea.Batch(s.batchLoadFieldsCmd(ctx), s.refreshSelectCmd())

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
		return s, tea.Batch(s.batchLoadFieldsCmd(ctx), s.refreshSelectCmd())

	// Batch field loading results
	case BatchContentFieldsLoadedMsg:
		s.AllFields = msg.Fields
		return s, nil
	case BatchAdminContentFieldsLoadedMsg:
		s.AllAdminFields = msg.Fields
		return s, nil

	// Content CRUD completion messages → reload tree + refresh select list
	case ContentCreatedMsg:
		s.clearError()
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())
	case ContentCreatedWithErrorsMsg:
		s.clearError()
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())

	case AdminContentUpdatedFromDialogMsg:
		s.clearError()
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())

	case ContentDeletedMsg:
		s.clearError()
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())

	case ContentReorderedMsg:
		s.clearError()
		s.PendingCursorContentID = msg.ContentID
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())

	case ContentCopiedMsg:
		s.clearError()
		s.PendingCursorContentID = msg.NewContentID
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())

	case ContentMovedMsg:
		s.clearError()
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())

	// Publish/Unpublish completion
	case PublishCompletedMsg:
		s.clearError()
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())
	case UnpublishCompletedMsg:
		s.clearError()
		return s, tea.Batch(s.reloadTreeCmd(ctx.Config), s.refreshSelectCmd())

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
		return s, s.reloadTreeCmd(ctx.Config)

	// Content field operation results → reload batch fields
	case ContentFieldUpdatedMsg:
		s.clearError()
		return s, s.batchLoadFieldsCmd(ctx)
	case ContentFieldAddedMsg:
		s.clearError()
		return s, s.batchLoadFieldsCmd(ctx)
	case ContentFieldDeletedMsg:
		s.clearError()
		return s, s.batchLoadFieldsCmd(ctx)

	// Content publish toggled (legacy compat)
	case ContentPublishToggledMsg:
		s.clearError()
		return s, s.reloadTreeCmd(ctx.Config)
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
			if len(s.AdminRootDatatypes) == 0 {
				return s, ShowDialog("Info", "No admin root datatypes available", false)
			}
			return s, ShowCreateAdminRouteWithContentDialogCmd(s.AdminRootDatatypes)
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
			if s.AdminMode && node.AdminContent != nil && node.AdminContent.AdminRouteID.Valid && node.AdminContent.AdminDatatypeID.Valid {
				adminContentID := node.AdminContent.AdminContentDataID
				adminDatatypeID := node.AdminContent.AdminDatatypeID.ID
				adminRouteID := node.AdminContent.AdminRouteID.ID
				title := fmt.Sprintf("Edit: %s", node.AdminContent.RouteTitle)
				return s, func() tea.Msg {
					return AdminFetchContentForEditMsg{
						AdminContentID:  adminContentID,
						AdminDatatypeID: adminDatatypeID,
						AdminRouteID:    adminRouteID,
						Title:           title,
					}
				}
			}
			if !s.AdminMode && node.Content != nil && node.Content.RouteID.Valid && node.Content.DatatypeID.Valid {
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
			if !s.AdminMode && node.Content != nil {
				contentName := node.Label
				hasChildren := node.Content.FirstChildID.Valid
				return s, ShowDeleteContentDialogCmd(
					string(node.Content.ContentDataID),
					contentName,
					hasChildren,
				)
			}
		}
		return s, nil
	}

	// Publish from select list
	if km.Matches(key, config.ActionPublish) {
		if s.Cursor < len(s.FlatSelectList) {
			node := s.FlatSelectList[s.Cursor]
			if s.AdminMode && node.AdminContent != nil {
				adminID := types.AdminContentID(node.AdminContent.AdminContentDataID)
				isPublished := node.AdminContent.Status == types.ContentStatusPublished
				return s, ShowPublishAdminContentDialogCmd(adminID, node.AdminContent.AdminRouteID.ID, node.AdminContent.DatatypeLabel, isPublished)
			}
			if node.Content != nil && node.Content.RouteID.Valid {
				isPublished := node.Content.Status == types.ContentStatusPublished
				return s, ShowPublishDialogCmd(node.Content.ContentDataID, node.Content.RouteID.ID, string(node.Content.RouteTitle), isPublished)
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
		if node.Content.DatatypeID.Valid {
			s.SelectedDatatypeID = node.Content.DatatypeID.ID
		}

		if node.Content.RouteID.Valid {
			s.Cursor = 0
			s.Grid = contentTreeGrid
			s.PageRouteId = node.Content.RouteID.ID
			return s, tea.Batch(
				LoadingStartCmd(),
				ReloadContentTreeCmd(ctx.Config, node.Content.RouteID.ID),
			)
		}
		if node.Content.RootID.Valid {
			s.Cursor = 0
			s.Grid = contentTreeGrid
			s.PageRootId = node.Content.RootID.ID
			return s, tea.Batch(
				LoadingStartCmd(),
				ReloadContentTreeByRootIDCmd(ctx.Config, node.Content.RootID.ID),
			)
		}
		return s, nil
	}

	// Admin content
	if node.AdminContent != nil {
		if node.AdminContent.AdminRouteID.Valid {
			s.Cursor = 0
			s.Grid = contentTreeGrid
			s.AdminPageRouteId = node.AdminContent.AdminRouteID.ID
			return s, tea.Batch(
				LoadingStartCmd(),
				ReloadAdminContentTreeCmd(ctx.Config, node.AdminContent.AdminRouteID.ID),
			)
		}
		if node.AdminContent.RootID.Valid {
			s.Cursor = 0
			s.Grid = contentTreeGrid
			s.AdminPageRootId = node.AdminContent.RootID.ID
			return s, tea.Batch(
				LoadingStartCmd(),
				ReloadAdminContentTreeByRootIDCmd(ctx.Config, node.AdminContent.RootID.ID),
			)
		}
		return s, nil
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
				name := DecideNodeName(*s.Root.Root)
				isPublished := s.Root.Root.Instance.Status == types.ContentStatusPublished
				return s, ShowPublishDialogCmd(s.Root.Root.Instance.ContentDataID, s.PageRouteId, name, isPublished)
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
		s.PageRootId = types.ContentID("")
		s.AdminPageRootId = types.AdminContentID("")
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

// reloadTreeCmd returns the command to reload the current tree by route_id or root_id.
func (s *ContentScreen) reloadTreeCmd(cfg *config.Config) tea.Cmd {
	if s.AdminMode {
		if !s.AdminPageRouteId.IsZero() {
			return ReloadAdminContentTreeCmd(cfg, s.AdminPageRouteId)
		}
		if !s.AdminPageRootId.IsZero() {
			return ReloadAdminContentTreeByRootIDCmd(cfg, s.AdminPageRootId)
		}
		return nil
	}
	if !s.PageRouteId.IsZero() {
		return ReloadContentTreeCmd(cfg, s.PageRouteId)
	}
	if !s.PageRootId.IsZero() {
		return ReloadContentTreeByRootIDCmd(cfg, s.PageRootId)
	}
	return nil
}

// batchLoadFieldsCmd returns the command to batch-load fields using route_id or root_id.
func (s *ContentScreen) batchLoadFieldsCmd(ctx AppContext) tea.Cmd {
	if s.AdminMode {
		if !s.AdminPageRouteId.IsZero() {
			return BatchLoadAdminContentFieldsCmd(ctx.Config, s.AdminPageRouteId, s.collectAdminDatatypeIDs(), ctx.ActiveLocale)
		}
		if !s.AdminPageRootId.IsZero() {
			return BatchLoadAdminContentFieldsByRootIDCmd(ctx.Config, s.AdminPageRootId, s.collectAdminDatatypeIDs(), ctx.ActiveLocale)
		}
		return nil
	}
	if !s.PageRouteId.IsZero() {
		return BatchLoadContentFieldsCmd(ctx.Config, s.PageRouteId, s.collectDatatypeIDs(), ctx.ActiveLocale)
	}
	if !s.PageRootId.IsZero() {
		return BatchLoadContentFieldsByRootIDCmd(ctx.Config, s.PageRootId, s.collectDatatypeIDs(), ctx.ActiveLocale)
	}
	return nil
}

// refreshSelectCmd returns the command to refetch the select list data for the
// current mode. Batch this into completion handlers so the select phase stays current.
func (s *ContentScreen) refreshSelectCmd() tea.Cmd {
	if s.AdminMode {
		return AdminContentDataFetchCmd()
	}
	return RootContentSummaryFetchCmd()
}

// rebuildSelectTree rebuilds the slug-grouped select tree and flat cursor list
// from the current content summary data.
func (s *ContentScreen) rebuildSelectTree() {
	if s.AdminMode {
		s.SelectTree = BuildAdminContentSelectTree(s.AdminRootContentSummary, s.TitleMap)
	} else {
		s.SelectTree = BuildContentSelectTree(s.RootContentSummary, s.TitleMap)
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
