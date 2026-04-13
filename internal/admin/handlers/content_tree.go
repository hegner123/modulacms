package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/tree/core"
	"github.com/hegner123/modulacms/internal/utility"
)

// DrawerFieldData wraps a content field row with metadata for drawer rendering.
type DrawerFieldData struct {
	Field       db.ContentFieldWithFieldRow
	FieldConfig string
}

// ContentTreePageHandler renders the full content tree two-panel layout.
// On initial load, it populates the sidebar with the route tree.
func ContentTreePageHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		layout := NewAdminData(r, "Content Tree")

		// Build sidebar data
		routes, globals, err := buildRouteTreeNodes(driver)
		if err != nil {
			utility.DefaultLogger.Error("content tree: failed to build route tree", err)
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		// Load only root/global datatypes for the top-level create dialog
		datatypes, dtErr := driver.ListDatatypes()
		var dtList []db.Datatypes
		if dtErr == nil && datatypes != nil {
			for _, dt := range *datatypes {
				if strings.HasPrefix(dt.Type, "_root") || strings.HasPrefix(dt.Type, "_global") {
					dtList = append(dtList, dt)
				}
			}
		}

		sidebar := partials.ContentTreeSidebar(routes, globals, "", csrfToken)
		createDialog := pages.ContentCreateDialog(dtList, csrfToken)
		content := pages.ContentTreeContentWithSidebar(sidebar)
		fullPage := pages.ContentTreeWithSidebar(layout.WithDialogs(createDialog), sidebar)
		RenderNav(w, r, "Content Tree", content, fullPage)
	}
}

// ContentTreeSidebarHandler returns the sidebar HTML for the route tree.
// Supports search filtering via ?search= query parameter.
func ContentTreeSidebarHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		csrfToken := CSRFTokenFromContext(r.Context())

		routes, globals, err := buildRouteTreeNodes(driver)
		if err != nil {
			utility.DefaultLogger.Error("content tree sidebar: failed to build route tree", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to load routes", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if search != "" {
			lower := strings.ToLower(search)
			routes = filterRouteTree(routes, lower)
			globals = filterRouteTree(globals, lower)
		}

		Render(w, r, partials.ContentTreeSidebar(routes, globals, search, csrfToken))
	}
}

// ContentTreeRouteChildrenHandler lazy-loads the content tree under a route.
// For route IDs, it finds the root content_data for that route and returns its child tree.
// For content IDs (when expanding a block subtree), it returns the node's children.
func ContentTreeRouteChildrenHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("routeID")
		if id == "" {
			http.Error(w, "missing ID", http.StatusBadRequest)
			return
		}

		// Try as a route ID first: find content_data with this route_id
		routeID := types.NullableRouteID{Valid: true, ID: types.RouteID(id)}
		contentRows, err := driver.ListContentDataByRoute(routeID)
		if err != nil {
			utility.DefaultLogger.Error("content tree children: list by route failed", err)
			http.Error(w, "failed to load. Click to retry.", http.StatusInternalServerError)
			return
		}

		var rootNodes []db.ContentData
		if contentRows != nil && len(*contentRows) > 0 {
			// Find root nodes for this route (nodes with no parent, or parent_id is null)
			for _, cd := range *contentRows {
				if !cd.ParentID.Valid {
					rootNodes = append(rootNodes, cd)
				}
			}
			// If no parentless nodes found, this means our lookup found content
			// but they all have parents. Return children of the first root.
			if len(rootNodes) == 0 {
				rootNodes = *contentRows
			}
		}

		// If no content found via route, try as a content data ID (block tree expansion)
		if len(rootNodes) == 0 {
			contentID := types.ContentID(id)
			node, getErr := driver.GetContentData(contentID)
			if getErr != nil || node == nil {
				// Neither route nor content found
				Render(w, r, partials.ContentTreeChildren(nil))
				return
			}
			// Return children of this content node
			rootNodes = collectDirectChildren(driver, node.ContentDataID)
		}

		// Build child nodes with resolved display names
		childNodes := make([]partials.ContentTreeChildNode, 0, len(rootNodes))
		for _, cd := range rootNodes {
			childNodes = append(childNodes, partials.ContentTreeChildNode{
				ContentData:   cd,
				DatatypeLabel: resolveDatatypeLabel(driver, cd.DatatypeID),
				DisplayName:   resolveContentDisplayName(driver, cd),
				HasChildren:   cd.FirstChildID.Valid,
			})
		}

		Render(w, r, partials.ContentTreeChildren(childNodes))
	}
}

// ContentTreePageViewHandler loads a content node and renders its block summary cards.
// Also returns OOB swaps for the sidebar (block-tree mode) and breadcrumb.
func ContentTreePageViewHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("contentID")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		content, err := driver.GetContentData(types.ContentID(id))
		if err != nil || content == nil {
			utility.DefaultLogger.Error("content tree page view: failed to get content", err)
			http.NotFound(w, r)
			return
		}

		// Find the tree root: use root_id if set, otherwise content itself is the root
		rootID := content.ContentDataID
		if content.RootID.Valid {
			rootID = content.RootID.ID
		}

		// Build page data from root
		root, _ := driver.GetContentData(rootID)
		pageTitle := resolveContentDisplayName(driver, *root)
		status := root.Status

		// Collect child blocks (exclude root -- root is the sidebar title)
		blocks, allNodes := buildPageBlocks(driver, rootID)
		// Remove the root node itself from the block list (it's the page header, not a block)
		var childBlocks []partials.ContentBlockSummary
		var childNodes []db.ContentData
		for i, b := range blocks {
			if b.ContentDataID != rootID {
				childBlocks = append(childBlocks, b)
				childNodes = append(childNodes, allNodes[i])
			}
		}

		// Batch-load fields for child blocks and attach to summaries
		attachFieldsToBlocks(driver, childBlocks, childNodes)

		hasPublishPerm := HasPermission(r, "content:publish")
		crumbs := buildContentBreadcrumb(driver, content)

		// Build sidebar: root as title, children as draggable items
		rootNode, sidebarChildren := buildBlockTreeSidebar(driver, rootID)

		// Build components
		csrfToken := CSRFTokenFromContext(r.Context())
		pageView := partials.ContentPageView(pageTitle, status, rootID, childBlocks, hasPublishPerm, csrfToken)
		sidebarComp := partials.ContentTreeBlockSidebar(rootNode, sidebarChildren, content.ContentDataID.String(), pageTitle)
		breadcrumbComp := partials.ContentBreadcrumbNav(crumbs)

		if IsHTMX(r) {
			RenderWithOOB(w, r, pageView,
				OOBSwap{TargetID: "content-tree-sidebar", Component: sidebarComp, InnerHTML: true},
				OOBSwap{TargetID: "content-breadcrumb", Component: breadcrumbComp, InnerHTML: true},
			)
			return
		}

		// Full page: render inside the content tree layout
		layout := NewAdminData(r, pageTitle)
		fullPage := pages.ContentTreeWithPageView(layout, sidebarComp, breadcrumbComp, pageView)
		Render(w, r, fullPage)
	}
}

// ContentTreeDrawerHandler returns the field edit drawer for a content block.
func ContentTreeDrawerHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("contentID")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		contentID := types.ContentID(id)
		content, err := driver.GetContentData(contentID)
		if err != nil || content == nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content block not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Load fields with field definitions
		fields, fieldErr := driver.ListContentFieldsWithFieldByContentData(
			types.NullableContentID{ID: contentID, Valid: true},
		)
		if fieldErr != nil {
			utility.DefaultLogger.Error("drawer: failed to load fields", fieldErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to load fields", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var contentFields []db.ContentFieldWithFieldRow
		if fields != nil {
			contentFields = *fields
		}

		// Filter by role access and load full field definitions
		editUser := middleware.AuthenticatedUser(r.Context())
		editIsAdmin := middleware.ContextIsAdmin(r.Context())
		editRoleID := ""
		if editUser != nil {
			editRoleID = editUser.Role
		}

		drawerFields := make([]partials.DrawerField, 0, len(contentFields))
		for _, cf := range contentFields {
			var fieldDef *db.Fields
			if cf.FieldID.Valid {
				fd, fErr := driver.GetField(cf.FieldID.ID)
				if fErr != nil {
					continue
				}
				fieldDef = fd
				if !editIsAdmin && !db.IsFieldAccessible(*fd, editRoleID, editIsAdmin) {
					continue
				}
			}
			drawerFields = append(drawerFields, partials.DrawerField{
				ContentField: cf,
				FieldDef:     fieldDef,
			})
		}

		dtLabel := resolveDatatypeLabel(driver, content.DatatypeID)
		csrfToken := CSRFTokenFromContext(r.Context())

		Render(w, r, partials.ContentDrawer(drawerFields, partials.DrawerMeta{
			DatatypeLabel: dtLabel,
			ContentID:     id,
			CSRFToken:     csrfToken,
		}))
	}
}

// ContentTreeDrawerSaveHandler saves field values from the drawer form.
// Reads form values keyed by "field_{content_field_id}" and updates each field.
// Returns OOB swap for the updated block card and a toast trigger.
func ContentTreeDrawerSaveHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("contentID")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		contentID := types.ContentID(id)

		// Load existing fields (full ContentFields with Locale)
		fields, fieldErr := driver.ListContentFieldsByContentData(
			types.NullableContentID{ID: contentID, Valid: true},
		)
		if fieldErr != nil {
			utility.DefaultLogger.Error("drawer save: failed to load fields", fieldErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to load fields", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if fields != nil {
			now := types.TimestampNow()
			for _, cf := range *fields {
				formKey := "field_" + cf.ContentFieldID.String()
				newValue, exists := r.Form[formKey]
				if !exists {
					continue
				}
				val := ""
				if len(newValue) > 0 {
					val = newValue[0]
				}

				if _, updateErr := driver.UpdateContentField(r.Context(), ac, db.UpdateContentFieldParams{
					ContentFieldID: cf.ContentFieldID,
					RouteID:        cf.RouteID,
					RootID:         cf.RootID,
					ContentDataID:  cf.ContentDataID,
					FieldID:        cf.FieldID,
					FieldValue:     val,
					Locale:         cf.Locale,
					AuthorID:       user.UserID,
					DateCreated:    cf.DateCreated,
					DateModified:   now,
				}); updateErr != nil {
					utility.DefaultLogger.Error("drawer save: field update failed", updateErr, "field_id", cf.ContentFieldID.String())
					w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to save field", "type": "error"}}`)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		// Refresh the inline block section via OOB swap + toast
		content, _ := driver.GetContentData(contentID)
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Saved", "type": "success"}}`)
		if content != nil {
			block := buildSingleBlockSummary(driver, content)
			sectionComp := partials.ContentBlockSectionOOB(block)
			Render(w, r, sectionComp)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// ContentTreeCreateHandler creates a new content block from the datatype card selection.
// Expects parent_id and datatype_id from the POST body. Creates the content node with
// scaffolded fields and redirects to the page view.
func ContentTreeCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parentID := r.FormValue("parent_id")
		datatypeID := r.FormValue("datatype_id")
		if datatypeID == "" {
			http.Error(w, "Datatype is required", http.StatusBadRequest)
			return
		}

		now := types.TimestampNow()
		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		// Resolve parent's route and root for inheritance
		var routeID types.NullableRouteID
		var rootID types.NullableContentID
		if parentID != "" {
			parent, parentErr := driver.GetContentData(types.ContentID(parentID))
			if parentErr == nil && parent != nil {
				routeID = parent.RouteID
				if parent.RootID.Valid {
					rootID = parent.RootID
				} else {
					rootID = types.NullableContentID{ID: parent.ContentDataID, Valid: true}
				}
			}
		}

		params := db.CreateContentDataParams{
			ParentID:     types.NullableContentID{ID: types.ContentID(parentID), Valid: parentID != ""},
			DatatypeID:   types.NullableDatatypeID{ID: types.DatatypeID(datatypeID), Valid: true},
			RouteID:      routeID,
			RootID:       rootID,
			AuthorID:     user.UserID,
			Status:       "draft",
			DateCreated:  now,
			DateModified: now,
		}

		created, createErr := driver.CreateContentData(r.Context(), ac, params)
		if createErr != nil {
			utility.DefaultLogger.Error("content tree create: failed", createErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create block", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Scaffold content fields for each field on the datatype
		dtFields, dtFieldsErr := driver.ListFieldsByDatatypeID(
			types.NullableDatatypeID{ID: types.DatatypeID(datatypeID), Valid: true},
		)
		if dtFieldsErr != nil {
			utility.DefaultLogger.Error("content tree create: failed to list fields", dtFieldsErr)
		} else if dtFields != nil {
			for _, f := range *dtFields {
				_, cfErr := driver.CreateContentField(r.Context(), ac, db.CreateContentFieldParams{
					ContentDataID: types.NullableContentID{ID: created.ContentDataID, Valid: true},
					RouteID:       routeID,
					RootID:        rootID,
					FieldID:       types.NullableFieldID{ID: f.FieldID, Valid: true},
					FieldValue:    "",
					AuthorID:      user.UserID,
					DateCreated:   now,
					DateModified:  now,
				})
				if cfErr != nil {
					utility.DefaultLogger.Error("content tree create: field scaffold failed", cfErr, "field", f.Label)
				}
			}
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Block created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/content-tree/page/"+created.ContentDataID.String())
		w.WriteHeader(http.StatusOK)
	}
}

// ContentTreeCreateOptionsHandler returns contextual datatype cards for the create flow.
// Lists all datatypes (future: filter by parent datatype constraints).
func ContentTreeCreateOptionsHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parentID := r.PathValue("parentID")

		datatypes, err := driver.ListDatatypes()
		if err != nil {
			utility.DefaultLogger.Error("content tree create options: failed to list datatypes", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to load datatypes", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var dtList []db.Datatypes
		if datatypes != nil {
			dtList = *datatypes
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, partials.ContentCreateCards(dtList, parentID, csrfToken))
	}
}

// ContentTreeDeleteBlockHandler deletes a content block and its children recursively.
// Walks the tree bottom-up (leaves first) to avoid orphaning children, since
// content_data self-referential FKs use ON DELETE SET NULL.
func ContentTreeDeleteBlockHandler(svc *service.Registry, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("contentID")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		// Collect all descendants in bottom-up order for safe deletion
		contentID := types.ContentID(id)
		deleteOrder := collectDescendantsBottomUp(svc.Driver(), contentID)
		deleteOrder = append(deleteOrder, contentID) // delete the target node last

		for _, nodeID := range deleteOrder {
			if deleteErr := svc.Driver().DeleteContentData(r.Context(), ac, nodeID); deleteErr != nil {
				utility.DefaultLogger.Error("content tree delete: failed to delete node", deleteErr, "node_id", nodeID.String())
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete content block", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Block deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// buildRouteTreeNodes loads all routes and top-level content, then builds a
// nested tree from route slug segments. Interior nodes are grouping labels
// (e.g. "building-content"), leaf/matching nodes link to their page view.
// Global content (unrouted) is returned separately.
func buildRouteTreeNodes(driver db.DbDriver) (tree []partials.RouteTreeNode, globals []partials.RouteTreeNode, err error) {
	allRoutes, err := driver.ListRoutes()
	if err != nil {
		return nil, nil, fmt.Errorf("list routes: %w", err)
	}

	// Load top-level content to map route_id -> content_data_id
	topLevel, err := driver.ListContentDataTopLevelPaginated(db.PaginationParams{Limit: 10000, Offset: 0})
	if err != nil {
		return nil, nil, fmt.Errorf("list top-level content: %w", err)
	}
	routeContentMap := make(map[string]db.ContentDataTopLevel)
	if topLevel != nil {
		for _, tl := range *topLevel {
			if tl.RouteID.Valid {
				routeContentMap[tl.RouteID.ID.String()] = tl
			} else if strings.HasPrefix(tl.DatatypeType, "_global") {
				globals = append(globals, partials.RouteTreeNode{
					Segment:       tl.DatatypeLabel,
					Label:         tl.DatatypeLabel,
					IsRoute:       false,
					IsGlobal:      true,
					ContentDataID: tl.ContentDataID,
				})
			}
		}
	}

	// Build segment tree from route slugs
	root := &partials.RouteTreeNode{Segment: "", Label: "root"}
	if allRoutes != nil {
		for _, route := range *allRoutes {
			slug := strings.TrimPrefix(string(route.Slug), "/")
			segments := splitSlugSegments(slug)

			node := partials.RouteTreeNode{
				Segment: segments[len(segments)-1],
				Label:   route.Title,
				IsRoute: true,
				RouteID: route.RouteID,
			}
			if tl, ok := routeContentMap[route.RouteID.String()]; ok {
				node.ContentDataID = tl.ContentDataID
			}

			insertIntoTree(root, segments, node)
		}
	}

	tree = root.Children
	return tree, globals, nil
}

// splitSlugSegments splits a slug (without leading /) into path segments.
// The root slug "/" becomes a single segment "".
func splitSlugSegments(slug string) []string {
	if slug == "" {
		return []string{""}
	}
	return strings.Split(slug, "/")
}

// insertIntoTree places a route node at the correct depth in the tree,
// creating intermediate grouping nodes as needed.
func insertIntoTree(root *partials.RouteTreeNode, segments []string, leaf partials.RouteTreeNode) {
	current := root
	for i, seg := range segments {
		isLast := i == len(segments)-1
		if isLast {
			// Check if an intermediate node already exists for this segment
			found := false
			for j := range current.Children {
				if current.Children[j].Segment == seg {
					// Promote the existing grouping node to a route node
					current.Children[j].IsRoute = true
					current.Children[j].RouteID = leaf.RouteID
					current.Children[j].ContentDataID = leaf.ContentDataID
					if leaf.Label != "" {
						current.Children[j].Label = leaf.Label
					}
					found = true
					break
				}
			}
			if !found {
				current.Children = append(current.Children, leaf)
			}
		} else {
			// Find or create intermediate grouping node
			var child *partials.RouteTreeNode
			for j := range current.Children {
				if current.Children[j].Segment == seg {
					child = &current.Children[j]
					break
				}
			}
			if child == nil {
				current.Children = append(current.Children, partials.RouteTreeNode{
					Segment: seg,
					Label:   segmentToLabel(seg),
				})
				child = &current.Children[len(current.Children)-1]
			}
			current = child
		}
	}
}

// segmentToLabel converts a slug segment like "building-content" to "building Content".
func segmentToLabel(seg string) string {
	if seg == "" {
		return "Home"
	}
	words := strings.Split(seg, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// filterRouteTree filters route tree nodes by case-insensitive substring match.
// Keeps a node if it matches OR if any descendant matches (preserving the path).
func filterRouteTree(nodes []partials.RouteTreeNode, lower string) []partials.RouteTreeNode {
	var filtered []partials.RouteTreeNode
	for _, n := range nodes {
		matchesSelf := strings.Contains(strings.ToLower(n.Label), lower) ||
			strings.Contains(strings.ToLower(n.Segment), lower)
		filteredChildren := filterRouteTree(n.Children, lower)
		if matchesSelf || len(filteredChildren) > 0 {
			copy := n
			if !matchesSelf {
				copy.Children = filteredChildren
			}
			filtered = append(filtered, copy)
		}
	}
	return filtered
}

// collectDirectChildren returns the immediate children of a content node.
// First tries the sibling pointer chain (first_child_id -> next_sibling_id).
// Falls back to querying all content for the same route and filtering by parent_id,
// since sibling pointers may not be populated for all content trees.
func collectDirectChildren(driver db.DbDriver, parentID types.ContentID) []db.ContentData {
	parent, err := driver.GetContentData(parentID)
	if err != nil || parent == nil {
		return nil
	}

	// Fast path: walk sibling chain if first_child_id is set
	if parent.FirstChildID.Valid {
		var children []db.ContentData
		currentID := parent.FirstChildID.ID
		for {
			child, childErr := driver.GetContentData(currentID)
			if childErr != nil || child == nil {
				break
			}
			children = append(children, *child)
			if !child.NextSiblingID.Valid {
				break
			}
			currentID = child.NextSiblingID.ID
		}
		if len(children) > 0 {
			return children
		}
	}

	// Fallback: query by route_id and filter by parent_id in memory
	if parent.RouteID.Valid {
		allForRoute, routeErr := driver.ListContentDataByRoute(parent.RouteID)
		if routeErr == nil && allForRoute != nil {
			var children []db.ContentData
			for _, cd := range *allForRoute {
				if cd.ParentID.Valid && cd.ParentID.ID == parentID {
					children = append(children, cd)
				}
			}
			return children
		}
	}

	// Last resort for unrouted content: query by root_id
	rootID := parentID
	if parent.RootID.Valid {
		rootID = parent.RootID.ID
	}
	allForRoot, rootErr := driver.ListContentDataByRootID(types.NullableContentID{ID: rootID, Valid: true})
	if rootErr == nil && allForRoot != nil {
		var children []db.ContentData
		for _, cd := range *allForRoot {
			if cd.ParentID.Valid && cd.ParentID.ID == parentID {
				children = append(children, cd)
			}
		}
		return children
	}

	return nil
}

// resolveDatatypeLabel returns the label for a datatype ID.
func resolveDatatypeLabel(driver db.DbDriver, dtID types.NullableDatatypeID) string {
	if !dtID.Valid {
		return ""
	}
	dt, err := driver.GetDatatype(dtID.ID)
	if err != nil || dt == nil {
		return ""
	}
	return dt.Label
}

// resolveDatatypeType returns the type string for a datatype ID.
func resolveDatatypeType(driver db.DbDriver, dtID types.NullableDatatypeID) string {
	if !dtID.Valid {
		return ""
	}
	dt, err := driver.GetDatatype(dtID.ID)
	if err != nil || dt == nil {
		return ""
	}
	return dt.Type
}

// buildPageBlocks walks the content tree in DFS order starting from rootID
// and returns block summaries with depth info. Also returns the flat list of
// all content data nodes visited (for batch field loading).
// buildPageBlocks uses the core tree to walk all content nodes in DFS order
// and produce block summaries for the page view. Excludes the root node
// (which is the page header, not a block).
func buildPageBlocks(driver db.DbDriver, rootID types.ContentID) ([]partials.ContentBlockSummary, []db.ContentData) {
	rootData, err := driver.GetContentData(rootID)
	if err != nil || rootData == nil {
		return nil, nil
	}

	treeRoot := buildCoreTree(driver, rootData)
	if treeRoot == nil || treeRoot.Node == nil {
		return nil, nil
	}

	var blocks []partials.ContentBlockSummary
	var allNodes []db.ContentData

	var walkSiblings func(node *core.Node, depth int)
	walkSiblings = func(node *core.Node, depth int) {
		for current := node; current != nil; current = current.NextSibling {
			if current.ContentData == nil {
				continue
			}
			allNodes = append(allNodes, *current.ContentData)

			childCount := 0
			for ch := current.FirstChild; ch != nil; ch = ch.NextSibling {
				childCount++
			}

			// Dirty = modified after published_at (or never published)
			cd := current.ContentData
			isDirty := cd.Status == "draft" ||
				(cd.PublishedAt.String() != "" && cd.DateModified.String() > cd.PublishedAt.String())

			blocks = append(blocks, partials.ContentBlockSummary{
				ContentDataID: current.ContentData.ContentDataID,
				DatatypeLabel: current.Datatype.Label,
				DatatypeType:  current.Datatype.Type,
				ChildCount:    childCount,
				Depth:         depth,
				IsDirty:       isDirty,
			})

			if current.FirstChild != nil {
				walkSiblings(current.FirstChild, depth+1)
			}
		}
	}

	// Walk children of root (root itself is the page header)
	walkSiblings(treeRoot.Node.FirstChild, 0)
	return blocks, allNodes
}

// buildBlockTreeSidebar builds a flat list containing the root node and its
// immediate children for the block-tree sidebar.
// buildBlockTreeSidebar returns the root node (for the sidebar title) and its
// direct children (for the draggable block list).
// buildBlockTreeSidebar uses the shared tree/core package to build the full
// content tree from DB rows, then converts to template nodes. Returns the root
// (for the sidebar title) and its recursive children (for the block list).
func buildBlockTreeSidebar(driver db.DbDriver, rootID types.ContentID) (partials.ContentTreeChildNode, []partials.ContentTreeChildNode) {
	// Find the route for this root node
	rootData, err := driver.GetContentData(rootID)
	if err != nil || rootData == nil {
		return partials.ContentTreeChildNode{}, nil
	}

	rootDisplay := partials.ContentTreeChildNode{
		ContentData:   *rootData,
		DatatypeLabel: resolveDatatypeLabel(driver, rootData.DatatypeID),
		DisplayName:   resolveContentDisplayName(driver, *rootData),
	}

	// Build tree using the shared core package
	treeRoot := buildCoreTree(driver, rootData)
	if treeRoot == nil || treeRoot.Node == nil {
		return rootDisplay, nil
	}

	// Convert the root's children to template nodes (recursive)
	children := coreChildrenToTemplateNodes(driver, treeRoot.Node.FirstChild, 0)
	return rootDisplay, children
}

// buildCoreTree loads content tree rows and builds a core.Root using the same
// algorithm as the TUI. Works for both routed content (via route_id) and
// unrouted globals (via root_id).
func buildCoreTree(driver db.DbDriver, rootData *db.ContentData) *core.Root {
	var rows *[]db.GetContentTreeByRouteRow
	var err error

	if rootData.RouteID.Valid {
		rows, err = driver.GetContentTreeByRoute(rootData.RouteID)
	} else {
		// Unrouted content (_global, etc.): load by root_id (self)
		rootID := types.NullableContentID{ID: rootData.ContentDataID, Valid: true}
		rows, err = driver.GetContentTreeByRootID(rootID)
	}
	if err != nil || rows == nil {
		return nil
	}
	treeRoot, _, buildErr := core.BuildFromRows(*rows)
	if buildErr != nil || treeRoot == nil {
		return nil
	}
	return treeRoot
}

// coreChildrenToTemplateNodes recursively converts core.Node siblings into
// ContentTreeChildNode slices with Children populated.
func coreChildrenToTemplateNodes(driver db.DbDriver, node *core.Node, depth int) []partials.ContentTreeChildNode {
	var result []partials.ContentTreeChildNode
	current := node
	for current != nil {
		if current.ContentData == nil {
			current = current.NextSibling
			continue
		}

		displayName := current.Datatype.Label
		// Resolve display name from title field if available
		for _, cf := range current.ContentFields {
			for _, f := range current.Fields {
				if f.FieldID == cf.FieldID.ID && string(f.Type) == "_title" && cf.FieldValue != "" {
					displayName = cf.FieldValue
					break
				}
			}
		}

		cd := current.ContentData
		isDirty := cd.Status == "draft" ||
			(cd.PublishedAt.String() != "" && cd.DateModified.String() > cd.PublishedAt.String())

		tn := partials.ContentTreeChildNode{
			ContentData:   *cd,
			DatatypeLabel: current.Datatype.Label,
			DisplayName:   displayName,
			HasChildren:   current.FirstChild != nil,
			Depth:         depth,
			IsDirty:       isDirty,
		}

		if current.FirstChild != nil {
			tn.Children = coreChildrenToTemplateNodes(driver, current.FirstChild, depth+1)
		}

		result = append(result, tn)
		current = current.NextSibling
	}
	return result
}

// buildContentBreadcrumb constructs the breadcrumb segments for a content node.
func buildContentBreadcrumb(driver db.DbDriver, content *db.ContentData) partials.ContentBreadcrumb {
	crumbs := partials.ContentBreadcrumb{
		Segments: []partials.BreadcrumbSegment{
			{Label: "Content Tree", Href: "/admin/content-tree"},
		},
	}

	// Find the root to get route info
	rootID := content.ContentDataID
	if content.RootID.Valid {
		rootID = content.RootID.ID
	}
	root, _ := driver.GetContentData(rootID)
	if root != nil && root.RouteID.Valid {
		rt, rtErr := driver.GetRoute(root.RouteID.ID)
		if rtErr == nil && rt != nil {
			label := rt.Title
			if label == "" {
				label = string(rt.Slug)
			}
			crumbs.Segments = append(crumbs.Segments, partials.BreadcrumbSegment{
				Label: label,
				Href:  "/admin/content-tree/page/" + rootID.String(),
			})
		}
	} else if root != nil {
		// Unrouted content: use datatype label
		dtLabel := resolveDatatypeLabel(driver, root.DatatypeID)
		if dtLabel != "" {
			crumbs.Segments = append(crumbs.Segments, partials.BreadcrumbSegment{
				Label: dtLabel,
				Href:  "/admin/content-tree/page/" + rootID.String(),
			})
		}
	}

	// If viewing a non-root node, add its label
	if content.ContentDataID != rootID {
		dtLabel := resolveDatatypeLabel(driver, content.DatatypeID)
		crumbs.Segments = append(crumbs.Segments, partials.BreadcrumbSegment{
			Label: dtLabel,
		})
	}

	return crumbs
}

// truncateFieldValue returns the first maxLen characters of a value, appending "..." if truncated.
func truncateFieldValue(val string, maxLen int) string {
	if len(val) <= maxLen {
		return val
	}
	return val[:maxLen] + "..."
}

// attachFieldsToBlocks batch-loads fields for all blocks and attaches them
// as BlockFieldSummary slices. Resolves _id field values to display names.
func attachFieldsToBlocks(driver db.DbDriver, blocks []partials.ContentBlockSummary, nodes []db.ContentData) {
	if len(nodes) == 0 {
		return
	}
	nodeIDs := make([]types.ContentID, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ContentDataID
	}
	fields, fieldErr := driver.ListContentFieldsWithFieldByContentIDs(nodeIDs)
	if fieldErr != nil || fields == nil {
		return
	}

	// Group fields by content_data_id
	fieldMap := make(map[string][]db.ContentFieldWithFieldRow)
	for _, f := range *fields {
		key := f.ContentDataID.ID.String()
		fieldMap[key] = append(fieldMap[key], f)
	}

	for i := range blocks {
		key := blocks[i].ContentDataID.String()
		flds, ok := fieldMap[key]
		if !ok {
			continue
		}

		summaries := make([]partials.BlockFieldSummary, 0, len(flds))
		for _, f := range flds {
			summary := partials.BlockFieldSummary{
				Label: f.FLabel,
				Type:  string(f.FType),
				Value: f.FieldValue,
			}
			// Resolve _id fields to human-readable names
			if string(f.FType) == "_id" && f.FieldValue != "" {
				summary.ResolvedValue = resolveContentIDToLabel(driver, f.FieldValue)
			}
			summaries = append(summaries, summary)
		}
		blocks[i].Fields = summaries

		// For reference-type blocks, set a resolved label from the _id field
		if strings.HasPrefix(blocks[i].DatatypeType, "_reference") {
			for _, s := range summaries {
				if s.Type == "_id" && s.ResolvedValue != "" {
					blocks[i].ResolvedLabel = s.ResolvedValue
					break
				}
			}
		}
	}
}

// resolveContentIDToLabel takes a content_data_id string and resolves it to a
// human-readable label by finding the content's title field or datatype label.
func resolveContentIDToLabel(driver db.DbDriver, contentIDStr string) string {
	cd, err := driver.GetContentData(types.ContentID(contentIDStr))
	if err != nil || cd == nil {
		return ""
	}
	return resolveContentDisplayName(driver, *cd)
}

// collectDescendantsBottomUp collects all descendant IDs of a content node
// in bottom-up (leaves first) order for safe recursive deletion.
func collectDescendantsBottomUp(driver db.DbDriver, rootID types.ContentID) []types.ContentID {
	var result []types.ContentID

	var walk func(nodeID types.ContentID)
	walk = func(nodeID types.ContentID) {
		children := collectDirectChildren(driver, nodeID)
		for _, child := range children {
			walk(child.ContentDataID)
			result = append(result, child.ContentDataID)
		}
	}

	walk(rootID)
	return result
}

// buildSingleBlockSummary creates a ContentBlockSummary for a single content node,
// including its field previews.
func buildSingleBlockSummary(driver db.DbDriver, content *db.ContentData) partials.ContentBlockSummary {
	dtType := resolveDatatypeType(driver, content.DatatypeID)
	block := partials.ContentBlockSummary{
		ContentDataID: content.ContentDataID,
		DatatypeLabel: resolveDatatypeLabel(driver, content.DatatypeID),
		DatatypeType:  dtType,
		ChildCount:    len(collectDirectChildren(driver, content.ContentDataID)),
	}

	fields, err := driver.ListContentFieldsWithFieldByContentData(
		types.NullableContentID{ID: content.ContentDataID, Valid: true},
	)
	if err == nil && fields != nil {
		summaries := make([]partials.BlockFieldSummary, 0, len(*fields))
		for _, f := range *fields {
			s := partials.BlockFieldSummary{
				Label: f.FLabel,
				Type:  string(f.FType),
				Value: f.FieldValue,
			}
			if string(f.FType) == "_id" && f.FieldValue != "" {
				s.ResolvedValue = resolveContentIDToLabel(driver, f.FieldValue)
			}
			summaries = append(summaries, s)
		}
		block.Fields = summaries

		if strings.HasPrefix(dtType, "_reference") {
			for _, s := range summaries {
				if s.Type == "_id" && s.ResolvedValue != "" {
					block.ResolvedLabel = s.ResolvedValue
					break
				}
			}
		}
	}

	return block
}
