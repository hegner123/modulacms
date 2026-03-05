package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// ContentScreen implements Screen for both CONTENT and ADMINCONTENT pages.
// It supports multiple phases: route list, tree browsing, field editing,
// and version management. AdminMode selects admin vs regular data and IDs.
type ContentScreen struct {
	AdminMode bool

	// Cursors and focus
	Cursor      int
	FieldCursor int
	PanelFocus  FocusPanel

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
	SelectedContentFields  []ContentFieldDisplay
	AdminSelectedFields    []AdminContentFieldDisplay
	PendingCursorContentID types.ContentID

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
	return &ContentScreen{
		AdminMode:               adminMode,
		Cursor:                  0,
		FieldCursor:             0,
		PanelFocus:              TreePanel,
		RootContentSummary:      rootContentSummary,
		AdminRootContentSummary: adminRootContentSummary,
		RootDatatypes:           rootDatatypes,
		AdminRootDatatypes:      adminRootDatatypes,
		PageRouteId:             pageRouteId,
	}
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
		switch s.PanelFocus {
		case RoutePanel:
			return []KeyHint{
				{km.HintString(config.ActionEdit), "edit"},
				{km.HintString(config.ActionNew), "new"},
				{km.HintString(config.ActionDelete), "del"},
				{km.HintString(config.ActionNextPanel), "panel"},
				{km.HintString(config.ActionBack), "back"},
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
	// Route list phase
	return []KeyHint{
		{km.HintString(config.ActionSelect), "select"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
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

// routeListLen returns the number of items in the route/content list.
func (s *ContentScreen) routeListLen() int {
	if s.AdminMode {
		return len(s.AdminRootContentSummary)
	}
	return len(s.RootContentSummary)
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
			switch s.PanelFocus {
			case TreePanel:
				return s.handleTreeKeys(ctx, key, km)
			case RoutePanel:
				return s.handleFieldPanelKeys(ctx, key, km)
			}
			// ContentPanel is skipped in focus cycle, but handle basic nav
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
		return s, nil
	case AdminContentDataSet:
		s.AdminRootContentSummary = msg.AdminContentData
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
		s.FieldCursor = 0
		// Load fields for first node
		return s, s.loadFieldsForCurrentNode(ctx)

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
		s.FieldCursor = 0
		return s, s.loadAdminFieldsForCurrentNode(ctx)

	// Field loading results
	case LoadContentFieldsMsg:
		s.SelectedContentFields = msg.Fields
		s.FieldCursor = 0
		return s, nil
	case AdminLoadContentFieldsMsg:
		s.AdminSelectedFields = msg.Fields
		s.FieldCursor = 0
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
		s.PanelFocus = RoutePanel
		return s, nil
	case AdminVersionsListedMsg:
		s.clearError()
		s.ShowVersionList = true
		s.AdminVersions = msg.Versions
		s.VersionCursor = 0
		s.AdminVersionContentID = msg.AdminContentID
		s.AdminVersionRouteID = msg.AdminRouteID
		s.PanelFocus = RoutePanel
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

	// Content field operation results → reload fields
	case AdminContentFieldUpdatedMsg:
		s.clearError()
		return s, s.loadAdminFieldsForCurrentNode(ctx)
	case AdminContentFieldAddedMsg:
		s.clearError()
		return s, s.loadAdminFieldsForCurrentNode(ctx)
	case AdminContentFieldDeletedMsg:
		s.clearError()
		return s, s.loadAdminFieldsForCurrentNode(ctx)

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

	// Panel navigation
	if km.Matches(key, config.ActionNextPanel) {
		s.PanelFocus = (s.PanelFocus + 1) % 3
		return s, nil
	}
	if km.Matches(key, config.ActionPrevPanel) {
		s.PanelFocus = (s.PanelFocus + 2) % 3
		return s, nil
	}

	// Cursor movement
	if km.Matches(key, config.ActionUp) {
		if s.Cursor > 0 {
			s.Cursor--
		}
		return s, nil
	}
	if km.Matches(key, config.ActionDown) {
		max := s.routeListLen() - 1
		if max < 0 {
			max = 0
		}
		if s.Cursor < max {
			s.Cursor++
		}
		return s, nil
	}

	// Select route → load tree
	if km.Matches(key, config.ActionSelect) {
		if s.AdminMode {
			return s.handleAdminRouteSelect(ctx)
		}
		return s.handleRegularRouteSelect(ctx)
	}

	// New content
	if km.Matches(key, config.ActionNew) {
		if s.AdminMode {
			// Admin: new content requires an admin route and datatypes
			// For now, show a dialog — this is handled through the existing form system
			return s, nil
		}
		if len(s.RootDatatypes) == 0 {
			return s, ShowDialog("Info", "No root datatypes available", false)
		}
		return s, ShowCreateRouteWithContentDialogCmd(s.RootDatatypes)
	}

	// Edit content from route list
	if km.Matches(key, config.ActionEdit) {
		if !s.AdminMode {
			if len(s.RootContentSummary) > 0 && s.Cursor < len(s.RootContentSummary) {
				content := s.RootContentSummary[s.Cursor]
				if content.RouteID.Valid && content.DatatypeID.Valid {
					return s, FetchContentForEditCmd(
						content.ContentDataID,
						content.DatatypeID.ID,
						content.RouteID.ID,
						fmt.Sprintf("Edit: %s", content.RouteTitle),
					)
				}
			}
		}
		return s, nil
	}

	// Delete content from route list
	if km.Matches(key, config.ActionDelete) {
		if s.AdminMode {
			if len(s.AdminRootContentSummary) > 0 && s.Cursor < len(s.AdminRootContentSummary) {
				content := s.AdminRootContentSummary[s.Cursor]
				hasChildren := content.FirstChildID.Valid
				return s, ShowDeleteAdminContentDialogCmd(
					content.AdminContentDataID,
					content.AdminRouteID.ID,
					content.DatatypeLabel,
					hasChildren,
				)
			}
		}
		return s, nil
	}

	return s, nil
}

// handleRegularRouteSelect handles selecting a route from the regular content list.
func (s *ContentScreen) handleRegularRouteSelect(ctx AppContext) (Screen, tea.Cmd) {
	if len(s.RootContentSummary) == 0 || s.Cursor >= len(s.RootContentSummary) {
		return s, nil
	}
	content := s.RootContentSummary[s.Cursor]
	if !content.RouteID.Valid {
		return s, nil
	}
	s.PageRouteId = content.RouteID.ID
	if content.DatatypeID.Valid {
		s.SelectedDatatypeID = content.DatatypeID.ID
	}
	s.Cursor = 0
	return s, tea.Batch(
		LoadingStartCmd(),
		ReloadContentTreeCmd(ctx.Config, content.RouteID.ID),
	)
}

// handleAdminRouteSelect handles selecting from the admin content list.
func (s *ContentScreen) handleAdminRouteSelect(ctx AppContext) (Screen, tea.Cmd) {
	if len(s.AdminRootContentSummary) == 0 || s.Cursor >= len(s.AdminRootContentSummary) {
		return s, nil
	}
	content := s.AdminRootContentSummary[s.Cursor]
	if !content.AdminRouteID.Valid {
		return s, nil
	}
	s.AdminPageRouteId = content.AdminRouteID.ID
	s.Cursor = 0
	return s, tea.Batch(
		LoadingStartCmd(),
		ReloadAdminContentTreeCmd(ctx.Config, content.AdminRouteID.ID),
	)
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

	// Panel navigation (skip ContentPanel)
	if km.Matches(key, config.ActionNextPanel) {
		if s.PanelFocus == TreePanel {
			s.PanelFocus = RoutePanel
		} else {
			s.PanelFocus = TreePanel
		}
		return s, nil
	}
	if km.Matches(key, config.ActionPrevPanel) {
		if s.PanelFocus == RoutePanel {
			s.PanelFocus = TreePanel
		} else {
			s.PanelFocus = RoutePanel
		}
		return s, nil
	}

	// Right arrow: TreePanel → RoutePanel
	if key == "l" || key == "right" {
		if s.PanelFocus == TreePanel {
			s.PanelFocus = RoutePanel
			s.FieldCursor = 0
			return s, nil
		}
	}

	// Cursor up/down
	if km.Matches(key, config.ActionUp) {
		if s.Cursor > 0 {
			s.Cursor--
			s.FieldCursor = 0
			return s, s.loadFieldsForCurrentNodeAuto(ctx)
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
			s.FieldCursor = 0
			return s, s.loadFieldsForCurrentNodeAuto(ctx)
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
				s.FieldCursor = 0
				return s, s.loadFieldsForCurrentNodeAuto(ctx)
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
					s.FieldCursor = 0
					return s, s.loadFieldsForCurrentNodeAuto(ctx)
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

// handleFieldPanelKeys handles key input when the field panel (RoutePanel) is focused.
func (s *ContentScreen) handleFieldPanelKeys(ctx AppContext, key string, km config.KeyMap) (Screen, tea.Cmd) {
	if km.Matches(key, config.ActionQuit) {
		return s, tea.Quit
	}

	// Back/dismiss: move focus back to TreePanel
	if km.Matches(key, config.ActionDismiss) || km.Matches(key, config.ActionBack) {
		s.PanelFocus = TreePanel
		return s, nil
	}

	// Panel navigation (skip ContentPanel)
	if km.Matches(key, config.ActionNextPanel) {
		s.PanelFocus = TreePanel
		return s, nil
	}
	if km.Matches(key, config.ActionPrevPanel) {
		s.PanelFocus = TreePanel
		return s, nil
	}

	// Cursor up/down on fields
	if km.Matches(key, config.ActionUp) {
		if s.FieldCursor > 0 {
			s.FieldCursor--
		}
		return s, nil
	}
	if km.Matches(key, config.ActionDown) {
		max := s.fieldsLen() - 1
		if max < 0 {
			max = 0
		}
		if s.FieldCursor < max {
			s.FieldCursor++
		}
		return s, nil
	}

	// Edit field
	if km.Matches(key, config.ActionEdit) {
		if s.AdminMode {
			return s.handleAdminFieldEdit(ctx)
		}
		return s.handleRegularFieldEdit(ctx)
	}

	// Add field
	if km.Matches(key, config.ActionNew) {
		if s.AdminMode {
			return s.handleAdminFieldAdd(ctx)
		}
		return s.handleRegularFieldAdd(ctx)
	}

	// Delete field
	if km.Matches(key, config.ActionDelete) {
		if s.AdminMode {
			return s.handleAdminFieldDelete(ctx)
		}
		return s.handleRegularFieldDelete(ctx)
	}

	// Reorder fields
	if km.Matches(key, config.ActionReorderUp) {
		return s.handleFieldReorder(ctx, "up")
	}
	if km.Matches(key, config.ActionReorderDown) {
		return s.handleFieldReorder(ctx, "down")
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
	if s.PanelFocus != TreePanel {
		s.PanelFocus = TreePanel
		return s, nil
	}
	if s.inTreePhase() {
		// Go back to route list
		s.Root = tree.Root{}
		s.PageRouteId = types.RouteID("")
		s.AdminPageRouteId = types.AdminRouteID("")
		s.SelectedContentFields = nil
		s.AdminSelectedFields = nil
		s.Cursor = 0
		s.FieldCursor = 0
		s.PanelFocus = TreePanel
		return s, nil
	}
	return s, HistoryPopCmd()
}

// fieldsLen returns the number of fields for the current mode.
func (s *ContentScreen) fieldsLen() int {
	if s.AdminMode {
		return len(s.AdminSelectedFields)
	}
	return len(s.SelectedContentFields)
}

// versionListLen returns the number of versions for the current mode.
func (s *ContentScreen) versionListLen() int {
	if s.AdminMode {
		return len(s.AdminVersions)
	}
	return len(s.Versions)
}

// loadFieldsForCurrentNode loads fields for the node at the current cursor position (regular mode).
func (s *ContentScreen) loadFieldsForCurrentNode(ctx AppContext) tea.Cmd {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return nil
	}
	return LoadContentFieldsCmd(ctx.Config, node.Instance.ContentDataID, node.Instance.DatatypeID)
}

// loadAdminFieldsForCurrentNode loads fields for the node at the current cursor position (admin mode).
func (s *ContentScreen) loadAdminFieldsForCurrentNode(ctx AppContext) tea.Cmd {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return nil
	}
	adminID := types.AdminContentID(node.Instance.ContentDataID)
	dtID := types.NullableAdminDatatypeID{}
	if node.Instance.DatatypeID.Valid {
		dtID = types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(node.Instance.DatatypeID.ID), Valid: true}
	}
	return LoadAdminContentFieldsCmd(ctx.Config, adminID, dtID, ctx.ActiveLocale)
}

// loadFieldsForCurrentNodeAuto dispatches to the correct mode-specific loader.
func (s *ContentScreen) loadFieldsForCurrentNodeAuto(ctx AppContext) tea.Cmd {
	if s.AdminMode {
		return s.loadAdminFieldsForCurrentNode(ctx)
	}
	return s.loadFieldsForCurrentNode(ctx)
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

// handleRegularFieldEdit opens a single-field edit dialog (regular mode).
func (s *ContentScreen) handleRegularFieldEdit(ctx AppContext) (Screen, tea.Cmd) {
	if len(s.SelectedContentFields) == 0 || s.FieldCursor >= len(s.SelectedContentFields) {
		return s, ShowDialog("Info", "No field selected", false)
	}
	cf := s.SelectedContentFields[s.FieldCursor]
	if cf.ContentFieldID.IsZero() {
		return s, ShowDialog("Info", "Field has no value yet. Use 'n' to add.", false)
	}
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, nil
	}
	return s, ShowEditSingleFieldDialogCmd(cf, node.Instance.ContentDataID, s.PageRouteId, node.Instance.DatatypeID)
}

// handleAdminFieldEdit opens a single-field edit dialog (admin mode).
func (s *ContentScreen) handleAdminFieldEdit(ctx AppContext) (Screen, tea.Cmd) {
	if len(s.AdminSelectedFields) == 0 || s.FieldCursor >= len(s.AdminSelectedFields) {
		return s, ShowDialog("Info", "No field selected", false)
	}
	cf := s.AdminSelectedFields[s.FieldCursor]
	if cf.AdminContentFieldID.IsZero() {
		return s, ShowDialog("Info", "Field has no value yet. Use 'n' to add.", false)
	}
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, nil
	}
	adminContentID := types.AdminContentID(node.Instance.ContentDataID)
	dtID := types.NullableAdminDatatypeID{}
	if node.Instance.DatatypeID.Valid {
		dtID = types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(node.Instance.DatatypeID.ID), Valid: true}
	}
	return s, ShowEditAdminSingleFieldDialogCmd(cf, adminContentID, s.AdminPageRouteId, dtID)
}

// handleRegularFieldAdd adds a field to the current content node (regular mode).
func (s *ContentScreen) handleRegularFieldAdd(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, nil
	}

	// Find fields not yet populated
	var missing []ContentFieldDisplay
	for _, cf := range s.SelectedContentFields {
		if cf.ContentFieldID.IsZero() {
			missing = append(missing, cf)
		}
	}
	if len(missing) == 0 {
		return s, ShowDialog("Info", "All fields already populated", false)
	}

	options := make([]huh.Option[string], 0, len(missing))
	for _, mf := range missing {
		options = append(options, huh.NewOption(mf.Label, string(mf.FieldID)))
	}
	return s, ShowAddContentFieldDialogCmd(options, node.Instance.ContentDataID, s.PageRouteId, node.Instance.DatatypeID)
}

// handleAdminFieldAdd adds a field to the current content node (admin mode).
func (s *ContentScreen) handleAdminFieldAdd(ctx AppContext) (Screen, tea.Cmd) {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, nil
	}
	adminContentID := types.AdminContentID(node.Instance.ContentDataID)
	dtID := types.NullableAdminDatatypeID{}
	if node.Instance.DatatypeID.Valid {
		dtID = types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(node.Instance.DatatypeID.ID), Valid: true}
	}

	// Find fields not yet populated
	var missing []AdminContentFieldDisplay
	for _, cf := range s.AdminSelectedFields {
		if cf.AdminContentFieldID.IsZero() {
			missing = append(missing, cf)
		}
	}
	if len(missing) == 0 {
		return s, ShowDialog("Info", "All fields already populated", false)
	}

	options := make([]huh.Option[string], 0, len(missing))
	for _, mf := range missing {
		options = append(options, huh.NewOption(mf.Label, string(mf.AdminFieldID)))
	}
	return s, ShowAddAdminContentFieldDialogCmd(options, adminContentID, s.AdminPageRouteId, dtID)
}

// handleRegularFieldDelete deletes the selected field (regular mode).
func (s *ContentScreen) handleRegularFieldDelete(ctx AppContext) (Screen, tea.Cmd) {
	if len(s.SelectedContentFields) == 0 || s.FieldCursor >= len(s.SelectedContentFields) {
		return s, nil
	}
	cf := s.SelectedContentFields[s.FieldCursor]
	if cf.ContentFieldID.IsZero() {
		return s, nil
	}
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, nil
	}
	return s, ShowDeleteContentFieldDialogCmd(cf, node.Instance.ContentDataID, s.PageRouteId, node.Instance.DatatypeID)
}

// handleAdminFieldDelete deletes the selected field (admin mode).
func (s *ContentScreen) handleAdminFieldDelete(ctx AppContext) (Screen, tea.Cmd) {
	if len(s.AdminSelectedFields) == 0 || s.FieldCursor >= len(s.AdminSelectedFields) {
		return s, nil
	}
	cf := s.AdminSelectedFields[s.FieldCursor]
	if cf.AdminContentFieldID.IsZero() {
		return s, nil
	}
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil || node.Instance == nil {
		return s, nil
	}
	adminContentID := types.AdminContentID(node.Instance.ContentDataID)
	dtID := types.NullableAdminDatatypeID{}
	if node.Instance.DatatypeID.Valid {
		dtID = types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(node.Instance.DatatypeID.ID), Valid: true}
	}
	return s, func() tea.Msg {
		return ShowDeleteAdminContentFieldDialogMsg{
			AdminContentFieldID: cf.AdminContentFieldID,
			AdminContentID:      adminContentID,
			AdminRouteID:        s.AdminPageRouteId,
			AdminDatatypeID:     dtID,
			Label:               cf.Label,
		}
	}
}

// handleFieldReorder handles reordering fields (stub for now).
func (s *ContentScreen) handleFieldReorder(ctx AppContext, direction string) (Screen, tea.Cmd) {
	// Field reordering not yet implemented for content fields
	return s, nil
}
