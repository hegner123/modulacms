package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/search"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/tree/core"
	"github.com/hegner123/modulacms/internal/tree/ops"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/hegner123/modulacms/internal/validation"
)

// clientIP extracts the client IP address from the request.
func clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// resolveContentDisplayName returns a human-readable name for a content item.
// Priority: route title + slug > title content field > truncated ID.
func resolveContentDisplayName(driver db.DbDriver, item db.ContentData) string {
	// If a route is assigned, use route title and slug
	if item.RouteID.Valid {
		rt, rtErr := driver.GetRoute(item.RouteID.ID)
		if rtErr == nil && rt != nil {
			return rt.Title
		}
	}

	// Look for a "title" content field
	fields, fieldsErr := driver.ListContentFieldsWithFieldByContentData(
		types.NullableContentID{ID: item.ContentDataID, Valid: true},
	)
	if fieldsErr == nil && fields != nil {
		for _, f := range *fields {
			if strings.EqualFold(f.FLabel, "title") && f.FieldValue != "" {
				return f.FieldValue
			}
		}
	}

	// Fallback: truncated ID
	id := item.ContentDataID.String()
	if len(id) > 12 {
		return id[:8] + "..."
	}
	return id
}

// ContentListHandler lists content with pagination and search.
// HTMX requests return partial table rows; full requests include the complete page layout.
// When a "search" query param is present and a search service is available,
// results are filtered via the full-text search index.
func ContentListHandler(driver db.DbDriver, mgr *config.Manager, searchSvc *search.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)
		statusFilter := r.URL.Query().Get("status")
		searchQuery := strings.TrimSpace(r.URL.Query().Get("search"))

		hasPublishPerm := HasPermission(r, "content:publish")
		var listItems []pages.ContentListItem
		var totalCount int64

		// Search path: use full-text index when a query is provided.
		if searchQuery != "" && searchSvc != nil {
			resp := searchSvc.SearchWithPrefix(searchQuery, search.SearchOptions{
				Limit:  int(limit),
				Offset: int(offset),
			})
			totalCount = int64(resp.Total)

			for _, result := range resp.Results {
				contentID := types.ContentID(result.ContentDataID)
				cd, cdErr := driver.GetContentData(contentID)
				if cdErr != nil || cd == nil {
					continue
				}
				if statusFilter != "" && string(cd.Status) != statusFilter {
					continue
				}
				item := pages.ContentListItem{
					ContentDataTopLevel: db.ContentDataTopLevel{
						ContentData:   *cd,
						AuthorName:    "",
						RouteSlug:     types.Slug(result.RouteSlug),
						RouteTitle:    result.RouteTitle,
						DatatypeLabel: result.DatatypeLabel,
						DatatypeType:  result.DatatypeName,
					},
					HasPublishPerm: hasPublishPerm,
				}
				item.DisplayName = resolveContentDisplayName(driver, *cd)
				if cd.RouteID.Valid {
					rt, rtErr := driver.GetRoute(cd.RouteID.ID)
					if rtErr == nil && rt != nil {
						item.Slug = string(rt.Slug)
					}
				}
				listItems = append(listItems, item)
			}
		} else {
			// Standard list path: paginated DB query.
			var rawItems []db.ContentDataTopLevel
			var total *int64

			if statusFilter != "" {
				items, err := driver.ListContentDataTopLevelPaginatedByStatus(db.PaginationParams{
					Limit:  limit,
					Offset: offset,
				}, types.ContentStatus(statusFilter))
				if err != nil {
					utility.DefaultLogger.Error("failed to list content by status", err)
					http.Error(w, "failed to load content", http.StatusInternalServerError)
					return
				}
				if items != nil {
					rawItems = *items
				}
				cnt, cntErr := driver.CountContentDataTopLevelByStatus(types.ContentStatus(statusFilter))
				if cntErr != nil {
					utility.DefaultLogger.Error("failed to count content by status", cntErr)
					http.Error(w, "failed to load content", http.StatusInternalServerError)
					return
				}
				total = cnt
			} else {
				items, err := driver.ListContentDataTopLevelPaginated(db.PaginationParams{
					Limit:  limit,
					Offset: offset,
				})
				if err != nil {
					utility.DefaultLogger.Error("failed to list content", err)
					http.Error(w, "failed to load content", http.StatusInternalServerError)
					return
				}
				if items != nil {
					rawItems = *items
				}
				cnt, cntErr := driver.CountContentDataTopLevel()
				if cntErr != nil {
					utility.DefaultLogger.Error("failed to count content", cntErr)
					http.Error(w, "failed to load content", http.StatusInternalServerError)
					return
				}
				total = cnt
			}

			totalCount = *total
			listItems = make([]pages.ContentListItem, len(rawItems))
			for i, item := range rawItems {
				listItems[i] = pages.ContentListItem{ContentDataTopLevel: item, HasPublishPerm: hasPublishPerm}
				listItems[i].DisplayName = resolveContentDisplayName(driver, item.ContentData)
				if item.RouteID.Valid {
					rt, rtErr := driver.GetRoute(item.RouteID.ID)
					if rtErr == nil && rt != nil {
						listItems[i].Slug = string(rt.Slug)
					}
				}
			}
		}

		pd := NewPaginationData(totalCount, limit, offset, "#content-table-body", "/admin/content")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsHTMX(r) && !IsNavHTMX(r) {
			Render(w, r, pages.ContentTableRowsPartial(listItems, pg))
			return
		}

		// Load root/global datatypes for the create dialog
		var datatypes []db.Datatypes
		dtList, dtErr := driver.ListDatatypesRoot()
		if dtErr != nil {
			utility.DefaultLogger.Error("failed to list datatypes for content dialog", dtErr)
		} else if dtList != nil {
			datatypes = *dtList
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Content"}`)
			RenderWithOOB(w, r, pages.ContentListContent(listItems, pg),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.ContentCreateDialog(datatypes, csrfToken)})
			return
		}

		layout := NewAdminData(r, "Content")
		Render(w, r, pages.ContentList(layout, listItems, pg, datatypes))
	}
}

// blockFieldData is a JSON-serializable representation of a content field
// for the block editor's side panel. Includes field definition metadata
// (label, type) alongside the content field value.
type blockFieldData struct {
	ContentFieldID string   `json:"contentFieldId"`
	FieldID        string   `json:"fieldId"`
	Label          string   `json:"label"`
	Type           string   `json:"type"`
	Value          string   `json:"value"`
	Toolbar        []string `json:"toolbar,omitempty"`
}

// blockNode is a JSON-serializable representation of a content tree node
// for the block editor web component. Includes all ContentData fields so the
// JS can send full UpdateContentDataParams to the batch API endpoint.
// Nullable fields use empty string for no value. Non-nullable fields always have values.
type blockNode struct {
	ID            string           `json:"id"`
	ParentID      string           `json:"parentId"`
	FirstChildID  string           `json:"firstChildId"`
	NextSiblingID string           `json:"nextSiblingId"`
	PrevSiblingID string           `json:"prevSiblingId"`
	RouteID       string           `json:"routeId"`
	DatatypeID    string           `json:"datatypeId"`
	AuthorID      string           `json:"authorId"`
	Status        string           `json:"status"`
	DateCreated   string           `json:"dateCreated"`
	DateModified  string           `json:"dateModified"`
	Type          string           `json:"type"`
	Label         string           `json:"label"`
	Fields        []blockFieldData `json:"fields"`
}

// nullableIDStr returns the string value of a NullableContentID, or empty string if not valid.
func nullableIDStr(n types.NullableContentID) string {
	if !n.Valid {
		return ""
	}
	return n.ID.String()
}

// emptyEditorState is the JSON for an empty block editor (no children).
const emptyEditorState = `{"blocks":{},"rootId":null}`

// buildTreeJSON loads the descendants of a content node via core.BuildFromRows,
// then returns the block editor state as a JSON string. Uses the core tree package
// for proper sibling ordering, orphan handling, and circular reference detection.
// Always returns valid JSON so the editor can render even with zero children.
func buildTreeJSON(ctx context.Context, driver db.DbDriver, contentID types.ContentID, roleID string, isAdmin bool) string {
	log := utility.DefaultLogger

	// Look up the content node to get its route_id.
	content, contentErr := driver.GetContentData(contentID)
	if contentErr != nil {
		log.Debug("buildTreeJSON: failed to get content node", "contentID", contentID, "err", contentErr)
		return emptyEditorState
	}

	// Load tree rows via the joined query (content_data + datatypes in one query).
	// For routed content use GetContentTreeByRoute; for unrouted content convert
	// descendants to tree rows with datatype lookups.
	var rows []db.GetContentTreeByRouteRow
	if content.RouteID.Valid {
		treeRows, treeErr := driver.GetContentTreeByRoute(content.RouteID)
		if treeErr != nil {
			log.Debug("buildTreeJSON: GetContentTreeByRoute failed", "routeID", content.RouteID.ID, "err", treeErr)
			return emptyEditorState
		}
		if treeRows != nil {
			rows = *treeRows
		}
	} else {
		desc, descErr := driver.GetContentDataDescendants(ctx, contentID)
		if descErr != nil {
			log.Debug("buildTreeJSON: GetContentDataDescendants failed", "contentID", contentID, "err", descErr)
			return emptyEditorState
		}
		if desc != nil {
			rows = contentDataToTreeRows(driver, *desc)
		}
	}

	if len(rows) <= 1 {
		log.Info("buildTreeJSON: no child blocks found", "contentID", contentID, "totalRows", len(rows))
		return emptyEditorState
	}

	// Build the tree using core.BuildFromRows — handles sibling ordering,
	// orphan resolution, and circular reference detection.
	root, stats, buildErr := core.BuildFromRows(rows)
	if buildErr != nil {
		log.Debug("buildTreeJSON: core.BuildFromRows warning", "err", buildErr,
			"orphans", len(stats.FinalOrphans), "circular", len(stats.CircularRefs))
		// Non-fatal: continue with partial tree (orphans are excluded, circular refs broken)
	}
	if root == nil || root.Node == nil {
		log.Info("buildTreeJSON: no root node after tree build")
		return emptyEditorState
	}

	log.Info("buildTreeJSON: tree built", "contentID", contentID,
		"nodes", stats.NodesCount, "orphans", len(stats.FinalOrphans))

	// Load content fields per descendant for the side panel.
	blockFields := make(map[string][]blockFieldData)
	for id, node := range root.NodeIndex {
		if id == contentID {
			continue
		}
		cid := types.NullableContentID{ID: node.ContentData.ContentDataID, Valid: true}
		cfRows, cfErr := driver.ListContentFieldsWithFieldByContentData(cid)
		if cfErr != nil || cfRows == nil {
			continue
		}
		fds := make([]blockFieldData, 0, len(*cfRows))
		for _, row := range *cfRows {
			if !isAdmin && row.FieldID.Valid {
				fieldDef, fieldErr := driver.GetField(row.FieldID.ID)
				if fieldErr != nil {
					continue
				}
				if !db.IsFieldAccessible(*fieldDef, roleID, isAdmin) {
					continue
				}
			}
			fds = append(fds, blockFieldData{
				ContentFieldID: row.ContentFieldID.String(),
				FieldID:        row.FFieldID.String(),
				Label:          row.FLabel,
				Type:           string(row.FType),
				Value:          row.FieldValue,
			})
		}
		blockFields[id.String()] = fds
	}

	// Walk tree nodes to build block editor state. Skip the content node itself.
	type editorState struct {
		Blocks map[string]blockNode `json:"blocks"`
		RootID string               `json:"rootId"`
	}
	state := editorState{Blocks: make(map[string]blockNode)}

	for id, node := range root.NodeIndex {
		if id == contentID {
			continue
		}
		cd := node.ContentData
		label := node.Datatype.Label
		if label == "" {
			label = "Untitled"
		}
		routeID := ""
		if cd.RouteID.Valid {
			routeID = cd.RouteID.ID.String()
		}
		datatypeID := ""
		if cd.DatatypeID.Valid {
			datatypeID = cd.DatatypeID.ID.String()
		}
		nodeIDStr := id.String()
		fields := blockFields[nodeIDStr]
		if fields == nil {
			fields = []blockFieldData{}
		}
		// Root-level blocks have parentId pointing to the content node in the DB,
		// but the content node isn't in the editor's blocks map. Clear parentId
		// for root-level blocks so the JS validator doesn't reject them.
		parentID := nullableIDStr(cd.ParentID)
		if parentID == contentID.String() {
			parentID = ""
		}

		state.Blocks[nodeIDStr] = blockNode{
			ID:            nodeIDStr,
			ParentID:      parentID,
			FirstChildID:  nullableIDStr(cd.FirstChildID),
			NextSiblingID: nullableIDStr(cd.NextSiblingID),
			PrevSiblingID: nullableIDStr(cd.PrevSiblingID),
			RouteID:       routeID,
			DatatypeID:    datatypeID,
			AuthorID:      string(cd.AuthorID),
			Status:        string(cd.Status),
			DateCreated:   cd.DateCreated.Time.UTC().Format(time.RFC3339),
			DateModified:  cd.DateModified.Time.UTC().Format(time.RFC3339),
			Type:          "container",
			Label:         label,
			Fields:        fields,
		}

		// Root block = direct child of content node with no previous sibling
		if parentID == "" && nullableIDStr(cd.PrevSiblingID) == "" {
			state.RootID = nodeIDStr
		}
	}

	if len(state.Blocks) == 0 {
		log.Info("buildTreeJSON: no child blocks after tree walk")
		return emptyEditorState
	}

	if state.RootID == "" {
		log.Debug("buildTreeJSON: no root block found", "contentID", contentID, "blockCount", len(state.Blocks))
	} else {
		log.Info("buildTreeJSON: tree assembled", "rootId", state.RootID, "blockCount", len(state.Blocks))
	}

	data, marshalErr := json.Marshal(state)
	if marshalErr != nil {
		log.Debug("buildTreeJSON: failed to marshal tree JSON", "err", marshalErr)
		return emptyEditorState
	}
	return string(data)
}

// contentDataToTreeRows converts flat ContentData rows (from GetContentDataDescendants)
// into GetContentTreeByRouteRow with datatype labels resolved. Used as a fallback
// for unrouted content that can't use the joined GetContentTreeByRoute query.
func contentDataToTreeRows(driver db.DbDriver, nodes []db.ContentData) []db.GetContentTreeByRouteRow {
	dtCache := make(map[types.DatatypeID]db.Datatypes)
	rows := make([]db.GetContentTreeByRouteRow, 0, len(nodes))
	for _, cd := range nodes {
		row := db.GetContentTreeByRouteRow{
			ContentDataID: cd.ContentDataID,
			ParentID:      cd.ParentID,
			FirstChildID:  cd.FirstChildID,
			NextSiblingID: cd.NextSiblingID,
			PrevSiblingID: cd.PrevSiblingID,
			DatatypeID:    cd.DatatypeID,
			RouteID:       cd.RouteID,
			AuthorID:      cd.AuthorID,
			Status:        cd.Status,
			DateCreated:   cd.DateCreated,
			DateModified:  cd.DateModified,
		}
		if cd.DatatypeID.Valid {
			dt, ok := dtCache[cd.DatatypeID.ID]
			if !ok {
				fetched, fetchErr := driver.GetDatatype(cd.DatatypeID.ID)
				if fetchErr == nil && fetched != nil {
					dt = *fetched
					dtCache[cd.DatatypeID.ID] = dt
				}
			}
			row.DatatypeLabel = dt.Label
			row.DatatypeType = dt.Type
		}
		rows = append(rows, row)
	}
	return rows
}

// ContentEditHandler renders the content editor page.
// Loads content by ID from the URL path and its associated fields.
// Resolves datatype, route, and author IDs to human-readable labels.
// When i18n is enabled, loads enabled locales for the locale tab bar.
func ContentEditHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		content, err := driver.GetContentData(types.ContentID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get content", err)
			http.NotFound(w, r)
			return
		}

		fields, fieldsErr := driver.ListContentFieldsWithFieldByContentData(
			types.NullableContentID{ID: content.ContentDataID, Valid: true},
		)
		if fieldsErr != nil {
			utility.DefaultLogger.Error("failed to get content fields", fieldsErr)
			http.Error(w, "failed to load content fields", http.StatusInternalServerError)
			return
		}

		var contentFields []db.ContentFieldWithFieldRow
		if fields != nil {
			contentFields = *fields
		}

		// Filter content fields by the authenticated user's role.
		// Look up each field definition to check its Roles column.
		editUser := middleware.AuthenticatedUser(r.Context())
		editIsAdmin := middleware.ContextIsAdmin(r.Context())
		editRoleID := ""
		if editUser != nil {
			editRoleID = editUser.Role
		}
		if !editIsAdmin && len(contentFields) > 0 {
			accessible := make([]db.ContentFieldWithFieldRow, 0, len(contentFields))
			for _, cf := range contentFields {
				if !cf.FieldID.Valid {
					accessible = append(accessible, cf)
					continue
				}
				fieldDef, fieldErr := driver.GetField(cf.FieldID.ID)
				if fieldErr != nil {
					// Cannot verify access; fail-closed: skip the field.
					utility.DefaultLogger.Warn("field role check: could not fetch field definition, skipping", fieldErr)
					continue
				}
				if db.IsFieldAccessible(*fieldDef, editRoleID, editIsAdmin) {
					accessible = append(accessible, cf)
				}
			}
			contentFields = accessible
		}

		// Resolve related entities to human-readable labels
		meta := pages.ContentMeta{}
		if content.DatatypeID.Valid {
			dt, dtErr := driver.GetDatatype(content.DatatypeID.ID)
			if dtErr == nil && dt != nil {
				meta.DatatypeLabel = dt.Label
			}
		}
		if content.RouteID.Valid {
			rt, rtErr := driver.GetRoute(content.RouteID.ID)
			if rtErr == nil && rt != nil {
				meta.RouteTitle = rt.Title
			}
		}
		author, authorErr := driver.GetUser(content.AuthorID)
		if authorErr == nil && author != nil {
			meta.AuthorName = author.Username
		}
		meta.HasPublishPerm = HasPermission(r, "content:publish")
		if content.PublishedBy.Valid {
			pubUser, pubErr := driver.GetUser(content.PublishedBy.ID)
			if pubErr == nil && pubUser != nil {
				meta.PublishedByName = pubUser.Username
			}
		}

		// Populate i18n metadata when enabled
		cfg, cfgErr := mgr.Config()
		if cfgErr == nil && cfg.I18nEnabled() {
			meta.I18nEnabled = true
			meta.ActiveLocale = r.URL.Query().Get("locale")
			if meta.ActiveLocale == "" {
				meta.ActiveLocale = cfg.I18nDefaultLocale()
			}
			enabledLocales, locErr := driver.ListEnabledLocales()
			if locErr == nil && enabledLocales != nil {
				meta.EnabledLocales = *enabledLocales
			}
		}

		treeJSON := buildTreeJSON(r.Context(), driver, content.ContentDataID, editRoleID, editIsAdmin)

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Edit Content")
		RenderNav(w, r, "Edit Content",
			pages.ContentEditContent(*content, contentFields, csrfToken, treeJSON, meta),
			pages.ContentEdit(layout, *content, contentFields, treeJSON, meta),
		)
	}
}

// ContentCreateHandler creates new content from a form submission.
// On success, HTMX requests receive an HX-Trigger toast; non-HTMX requests receive a redirect.
func ContentCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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

		status := types.ContentStatus(r.FormValue("status"))
		if status == "" {
			status = "draft"
		}

		parentID := r.FormValue("parent_id")
		datatypeID := r.FormValue("datatype_id")
		slug := r.FormValue("slug")
		title := r.FormValue("title")

		now := types.NewTimestamp(time.Now())
		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		// If a slug was provided, create a new route for this content
		var routeID types.NullableRouteID
		if slug != "" {
			routeTitle := title
			if routeTitle == "" {
				routeTitle = slug
			}
			route, routeErr := driver.CreateRoute(r.Context(), ac, db.CreateRouteParams{
				Slug:         types.Slug(slug),
				Title:        routeTitle,
				Status:       1,
				AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
				DateCreated:  now,
				DateModified: now,
			})
			if routeErr != nil {
				utility.DefaultLogger.Error("failed to create route for content", routeErr)
				if IsHTMX(r) {
					w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create route", "type": "error"}}`)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				http.Error(w, "failed to create route", http.StatusInternalServerError)
				return
			}
			routeID = types.NullableRouteID{ID: route.RouteID, Valid: true}
		}

		params := db.CreateContentDataParams{
			ParentID:     types.NullableContentID{ID: types.ContentID(parentID), Valid: parentID != ""},
			DatatypeID:   types.NullableDatatypeID{ID: types.DatatypeID(datatypeID), Valid: datatypeID != ""},
			RouteID:      routeID,
			AuthorID:     user.UserID,
			Status:       status,
			DateCreated:  now,
			DateModified: now,
		}

		created, createErr := driver.CreateContentData(r.Context(), ac, params)
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create content", createErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create content", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "failed to create content", http.StatusInternalServerError)
			return
		}

		// Create content field rows for every field defined on the datatype.
		// Pre-fill the "title" field if the user provided a name in the dialog.
		if datatypeID != "" {
			dtFields, dtFieldsErr := driver.ListFieldsByDatatypeID(
				types.NullableDatatypeID{ID: types.DatatypeID(datatypeID), Valid: true},
			)
			if dtFieldsErr != nil {
				utility.DefaultLogger.Error("failed to list datatype fields for content field creation", dtFieldsErr)
			} else if dtFields != nil {
				for _, f := range *dtFields {
					value := ""
					if title != "" && strings.EqualFold(f.Label, "title") {
						value = title
					}
					_, cfErr := driver.CreateContentField(r.Context(), ac, db.CreateContentFieldParams{
						ContentDataID: types.NullableContentID{ID: created.ContentDataID, Valid: true},
						FieldID:       types.NullableFieldID{ID: f.FieldID, Valid: true},
						FieldValue:    value,
						AuthorID:      user.UserID,
						DateCreated:   now,
						DateModified:  now,
					})
					if cfErr != nil {
						utility.DefaultLogger.Error("failed to create content field for "+f.Label, cfErr)
					}
				}
			}
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content created", "type": "success"}}`)
			w.Header().Set("HX-Redirect", "/admin/content-tree/page/"+created.ContentDataID.String())
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/content-tree/page/"+created.ContentDataID.String(), http.StatusSeeOther)
	}
}

// ContentUpdateHandler updates existing content from a form submission.
func ContentUpdateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

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

		existing, getErr := driver.GetContentData(types.ContentID(id))
		if getErr != nil {
			utility.DefaultLogger.Error("content not found for update", getErr)
			http.NotFound(w, r)
			return
		}

		status := types.ContentStatus(r.FormValue("status"))
		if status == "" {
			status = existing.Status
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		params := db.UpdateContentDataParams{
			ContentDataID: existing.ContentDataID,
			ParentID:      existing.ParentID,
			FirstChildID:  existing.FirstChildID,
			NextSiblingID: existing.NextSiblingID,
			PrevSiblingID: existing.PrevSiblingID,
			RouteID:       existing.RouteID,
			DatatypeID:    existing.DatatypeID,
			AuthorID:      existing.AuthorID,
			Status:        status,
			DateCreated:   existing.DateCreated,
			DateModified:  types.NewTimestamp(time.Now()),
		}

		if _, updateErr := driver.UpdateContentData(r.Context(), ac, params); updateErr != nil {
			utility.DefaultLogger.Error("failed to update content", updateErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to update content", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "failed to update content", http.StatusInternalServerError)
			return
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content updated", "type": "success"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/content/"+id, http.StatusSeeOther)
	}
}

// ContentDeleteHandler deletes content by ID.
// Only HTMX DELETE requests are supported; non-HTMX requests receive 405.
func ContentDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
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

		if deleteErr := driver.DeleteContentData(r.Context(), ac, types.ContentID(id)); deleteErr != nil {
			utility.DefaultLogger.Error("failed to delete content", deleteErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete content", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// reorderRequest is the JSON payload for the content reorder endpoint.
type reorderRequest struct {
	ParentID   string   `json:"parent_id"`
	OrderedIDs []string `json:"ordered_ids"`
}

// ContentReorderHandler reorders content siblings under a parent.
// Delegates to ContentService.Reorder which executes within a transaction
// with chain validation and assertions.
func ContentReorderHandler(svc *service.Registry, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req reorderRequest
		if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if len(req.OrderedIDs) == 0 {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "no items to reorder", "type": "warning"}}`)
			w.WriteHeader(http.StatusOK)
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

		// Convert string IDs to typed ContentIDs
		orderedIDs := make([]types.ContentID, len(req.OrderedIDs))
		for i, idStr := range req.OrderedIDs {
			orderedIDs[i] = types.ContentID(idStr)
		}

		var parentID ops.NullableID[types.ContentID]
		if req.ParentID != "" {
			parentID = ops.NullID(types.ContentID(req.ParentID))
		}

		_, reorderErr := svc.Content.Reorder(r.Context(), ac, parentID, orderedIDs)
		if reorderErr != nil {
			if ops.IsChainError(reorderErr) {
				utility.DefaultLogger.Error("content tree corrupted during reorder", reorderErr)
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content tree is corrupted. Run heal to repair.", "type": "error"}}`)
				w.WriteHeader(http.StatusConflict)
				return
			}
			utility.DefaultLogger.Error("failed to reorder content", reorderErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Reorder failed", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content reordered", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// moveRequest is the JSON payload for the content move endpoint.
type moveRequest struct {
	ContentID   string `json:"content_id"`
	NewParentID string `json:"new_parent_id"`
	Position    int    `json:"position"`
}

// ContentMoveHandler moves content to a new parent at a given position.
// Delegates to ContentService.Move which executes within a transaction with
// cycle detection, chain validation, and assertions.
func ContentMoveHandler(svc *service.Registry, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req moveRequest
		if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.ContentID == "" {
			http.Error(w, "content_id required", http.StatusBadRequest)
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

		var newParentID ops.NullableID[types.ContentID]
		if req.NewParentID != "" {
			newParentID = ops.NullID(types.ContentID(req.NewParentID))
		}

		moveParams := ops.MoveParams[types.ContentID]{
			NodeID:      types.ContentID(req.ContentID),
			NewParentID: newParentID,
			Position:    req.Position,
		}

		_, moveErr := svc.Content.Move(r.Context(), ac, moveParams)
		if moveErr != nil {
			if ops.IsChainError(moveErr) {
				utility.DefaultLogger.Error("content tree corrupted during move", moveErr)
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content tree is corrupted. Run heal to repair.", "type": "error"}}`)
				w.WriteHeader(http.StatusConflict)
				return
			}
			utility.DefaultLogger.Error("failed to move content", moveErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Move failed", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content moved", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// --- Content Tree Save (bulk creates, updates, deletes) ---

type treeSaveRequest struct {
	ContentID    string            `json:"content_id"`
	Creates      []treeNodeCreate  `json:"creates"`
	Updates      []treeNodeUpdate  `json:"updates"`
	Deletes      []string          `json:"deletes"`
	FieldUpdates []treeFieldUpdate `json:"field_updates"`
}

type treeFieldUpdate struct {
	ContentDataID  string `json:"content_data_id"`
	ContentFieldID string `json:"content_field_id"`
	FieldID        string `json:"field_id"`
	Value          string `json:"value"`
}

type treeNodeCreate struct {
	ClientID      string  `json:"client_id"`
	DatatypeID    string  `json:"datatype_id"`
	ParentID      *string `json:"parent_id"`
	FirstChildID  *string `json:"first_child_id"`
	NextSiblingID *string `json:"next_sibling_id"`
	PrevSiblingID *string `json:"prev_sibling_id"`
}

type treeNodeUpdate struct {
	ContentDataID string  `json:"content_data_id"`
	ParentID      *string `json:"parent_id"`
	FirstChildID  *string `json:"first_child_id"`
	NextSiblingID *string `json:"next_sibling_id"`
	PrevSiblingID *string `json:"prev_sibling_id"`
}

type treeSaveResponse struct {
	Created       int               `json:"created"`
	Updated       int               `json:"updated"`
	Deleted       int               `json:"deleted"`
	FieldsUpdated int               `json:"fields_updated"`
	IDMap         map[string]string `json:"id_map,omitempty"`
	Errors        []string          `json:"errors,omitempty"`
}

// parseNullableID converts a *string JSON pointer into a NullableContentID.
// nil means SQL NULL; non-nil is parsed as a ContentID.
func parseNullableID(s *string) (types.NullableContentID, error) {
	if s == nil {
		return types.NullableContentID{Valid: false}, nil
	}
	id := types.ContentID(*s)
	if err := id.Validate(); err != nil {
		return types.NullableContentID{}, fmt.Errorf("invalid id %q: %w", *s, err)
	}
	return types.NullableContentID{ID: id, Valid: true}, nil
}

// remapPtr replaces a client UUID pointer with its server ULID if mapped.
func remapPtr(ptr *string, idMap map[string]string) *string {
	if ptr == nil {
		return nil
	}
	if mapped, ok := idMap[*ptr]; ok {
		return &mapped
	}
	return ptr
}

// ContentTreeSaveHandler handles POST /admin/content/tree — bulk tree
// creates, pointer updates, and deletes from the block editor.
func ContentTreeSaveHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req treeSaveRequest
		if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log := utility.DefaultLogger
		log.Info("tree-save: request received",
			"contentID", req.ContentID,
			"creates", len(req.Creates),
			"updates", len(req.Updates),
			"deletes", len(req.Deletes),
			"fieldUpdates", len(req.FieldUpdates))

		if req.ContentID == "" {
			http.Error(w, "content_id required", http.StatusBadRequest)
			return
		}

		if len(req.Creates) == 0 && len(req.Updates) == 0 && len(req.Deletes) == 0 && len(req.FieldUpdates) == 0 {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "no changes to save", "type": "info"}}`)
			w.WriteHeader(http.StatusOK)
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

		ctx := r.Context()
		now := types.NewTimestamp(time.Now())
		nullPtr := types.NullableContentID{Valid: false}

		resp := treeSaveResponse{}

		// Resolve routeID from the parent content node for new blocks.
		var parentRouteID types.NullableRouteID
		parentContent, parentErr := driver.GetContentData(types.ContentID(req.ContentID))
		if parentErr != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("get parent %s: %v", req.ContentID, parentErr))
		} else {
			parentRouteID = parentContent.RouteID
		}

		// --- Phase 1a: Create rows with NULL pointers, collect server IDs ---
		idMap := make(map[string]string, len(req.Creates))
		type createdNode struct {
			clientID string
			serverID types.ContentID
			create   treeNodeCreate
		}
		created := make([]createdNode, 0, len(req.Creates))

		for _, cr := range req.Creates {
			if cr.ClientID == "" {
				resp.Errors = append(resp.Errors, "create: missing client_id")
				continue
			}

			var datatypeID types.NullableDatatypeID
			if cr.DatatypeID != "" {
				dtID := types.DatatypeID(cr.DatatypeID)
				if validateErr := dtID.Validate(); validateErr != nil {
					resp.Errors = append(resp.Errors, fmt.Sprintf("create %s: invalid datatype_id: %v", cr.ClientID, validateErr))
					continue
				}
				datatypeID = types.NullableDatatypeID{ID: dtID, Valid: true}
			}

			result, createErr := driver.CreateContentData(ctx, ac, db.CreateContentDataParams{
				ParentID:      nullPtr,
				FirstChildID:  nullPtr,
				NextSiblingID: nullPtr,
				PrevSiblingID: nullPtr,
				RouteID:       parentRouteID,
				DatatypeID:    datatypeID,
				AuthorID:      user.UserID,
				Status:        types.ContentStatusDraft,
				DateCreated:   now,
				DateModified:  now,
			})
			if createErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("tree-save: create %s failed", cr.ClientID), createErr)
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s: %v", cr.ClientID, createErr))
				continue
			}

			serverID := result.ContentDataID
			idMap[cr.ClientID] = serverID.String()
			created = append(created, createdNode{clientID: cr.ClientID, serverID: serverID, create: cr})
			resp.Created++
			log.Info("tree-save: created block",
				"clientID", cr.ClientID,
				"serverID", serverID,
				"datatypeID", cr.DatatypeID,
				"parentID", cr.ParentID)
		}

		if len(idMap) > 0 {
			resp.IDMap = idMap
		}

		// --- Phase 1b: Update newly created rows with remapped pointers ---
		contentParentID := types.NullableContentID{ID: types.ContentID(req.ContentID), Valid: true}

		for _, cn := range created {
			parentID, parseErr := parseNullableID(remapPtr(cn.create.ParentID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
				continue
			}
			firstChildID, parseErr := parseNullableID(remapPtr(cn.create.FirstChildID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
				continue
			}
			nextSiblingID, parseErr := parseNullableID(remapPtr(cn.create.NextSiblingID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
				continue
			}
			prevSiblingID, parseErr := parseNullableID(remapPtr(cn.create.PrevSiblingID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, parseErr))
				continue
			}

			// Root-level blocks in the editor have null parentId; anchor them
			// to the content node so they can be found by route query.
			if !parentID.Valid {
				parentID = contentParentID
				log.Info("tree-save: anchoring root-level block to content node",
					"blockID", cn.serverID, "contentID", req.ContentID)
			}

			log.Info("tree-save: updating pointers",
				"blockID", cn.serverID,
				"parentID", parentID,
				"firstChildID", firstChildID,
				"nextSiblingID", nextSiblingID,
				"prevSiblingID", prevSiblingID)

			existing, getErr := driver.GetContentData(cn.serverID)
			if getErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointer update: %v", cn.clientID, getErr))
				continue
			}

			_, updateErr := driver.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: cn.serverID,
				ParentID:      parentID,
				FirstChildID:  firstChildID,
				NextSiblingID: nextSiblingID,
				PrevSiblingID: prevSiblingID,
				RouteID:       existing.RouteID,
				DatatypeID:    existing.DatatypeID,
				AuthorID:      existing.AuthorID,
				Status:        existing.Status,
				DateCreated:   existing.DateCreated,
				DateModified:  now,
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("tree-save: create %s pointer update failed", cn.clientID), updateErr)
				resp.Errors = append(resp.Errors, fmt.Sprintf("create %s pointers: %v", cn.clientID, updateErr))
			}
		}

		// --- Phase 2: Deletes ---
		for _, idStr := range req.Deletes {
			id := types.ContentID(idStr)
			if validateErr := id.Validate(); validateErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("invalid delete id %s: %v", idStr, validateErr))
				continue
			}
			if deleteErr := driver.DeleteContentData(ctx, ac, id); deleteErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("tree-save: delete %s failed", idStr), deleteErr)
				resp.Errors = append(resp.Errors, fmt.Sprintf("delete %s: %v", idStr, deleteErr))
				continue
			}
			resp.Deleted++
		}

		// --- Phase 3: Updates ---
		for _, upd := range req.Updates {
			id := types.ContentID(upd.ContentDataID)
			if validateErr := id.Validate(); validateErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("invalid update id %s: %v", upd.ContentDataID, validateErr))
				continue
			}

			existing, getErr := driver.GetContentData(id)
			if getErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("get %s: %v", upd.ContentDataID, getErr))
				continue
			}

			parentID, parseErr := parseNullableID(remapPtr(upd.ParentID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
				continue
			}
			firstChildID, parseErr := parseNullableID(remapPtr(upd.FirstChildID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
				continue
			}
			nextSiblingID, parseErr := parseNullableID(remapPtr(upd.NextSiblingID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
				continue
			}
			prevSiblingID, parseErr := parseNullableID(remapPtr(upd.PrevSiblingID, idMap))
			if parseErr != nil {
				resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, parseErr))
				continue
			}

			// Root-level blocks in the editor have null parentId; anchor them
			// to the content node so GetContentDataDescendants can find them.
			if !parentID.Valid {
				parentID = contentParentID
			}

			_, updateErr := driver.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: existing.ContentDataID,
				ParentID:      parentID,
				FirstChildID:  firstChildID,
				NextSiblingID: nextSiblingID,
				PrevSiblingID: prevSiblingID,
				RouteID:       existing.RouteID,
				DatatypeID:    existing.DatatypeID,
				AuthorID:      existing.AuthorID,
				Status:        existing.Status,
				DateCreated:   existing.DateCreated,
				DateModified:  now,
			})
			if updateErr != nil {
				utility.DefaultLogger.Error(fmt.Sprintf("tree-save: update %s failed", upd.ContentDataID), updateErr)
				resp.Errors = append(resp.Errors, fmt.Sprintf("update %s: %v", upd.ContentDataID, updateErr))
				continue
			}
			resp.Updated++
		}

		// --- Phase 4: Field value updates ---
		// Pre-fetch existing content fields for all affected content_data_ids
		// so we can detect existing fields when content_field_id is empty (upsert).
		existingFieldsByCD := make(map[string]map[string]db.ContentFields) // content_data_id -> field_id -> ContentFields
		for _, fu := range req.FieldUpdates {
			cdIDStr := fu.ContentDataID
			if mapped, ok := idMap[cdIDStr]; ok {
				cdIDStr = mapped
			}
			if _, loaded := existingFieldsByCD[cdIDStr]; !loaded {
				cdID := types.NullableContentID{ID: types.ContentID(cdIDStr), Valid: true}
				existing, listErr := driver.ListContentFieldsByContentData(cdID)
				fieldMap := make(map[string]db.ContentFields)
				if listErr == nil && existing != nil {
					for _, cf := range *existing {
						if cf.FieldID.Valid {
							fieldMap[cf.FieldID.ID.String()] = cf
						}
					}
				}
				existingFieldsByCD[cdIDStr] = fieldMap
			}
		}

		saveIsAdmin := middleware.ContextIsAdmin(r.Context())
		saveRoleID := user.Role
		for _, fu := range req.FieldUpdates {
			cdIDStr := fu.ContentDataID
			if mapped, ok := idMap[cdIDStr]; ok {
				cdIDStr = mapped
			}

			// Guard: check field-level role access before allowing updates.
			// Resolve the field definition ID from either the field update or existing content field.
			guardFieldID := fu.FieldID
			if guardFieldID == "" && fu.ContentFieldID != "" {
				existingCF, cfErr := driver.GetContentField(types.ContentFieldID(fu.ContentFieldID))
				if cfErr == nil && existingCF.FieldID.Valid {
					guardFieldID = existingCF.FieldID.ID.String()
				}
			}
			var resolvedFieldDef *db.Fields
			if guardFieldID != "" {
				fieldDef, fieldErr := driver.GetField(types.FieldID(guardFieldID))
				if fieldErr != nil {
					utility.DefaultLogger.Warn(fmt.Sprintf("tree-save: field role check failed for %s, skipping update", guardFieldID), fieldErr)
					resp.Errors = append(resp.Errors, fmt.Sprintf("field access check failed for %s", guardFieldID))
					continue
				}
				if !db.IsFieldAccessible(*fieldDef, saveRoleID, saveIsAdmin) {
					utility.DefaultLogger.Warn("tree-save: field role access denied, skipping update",
						fmt.Errorf("role %s denied for field %s", saveRoleID, guardFieldID))
					continue
				}
				resolvedFieldDef = fieldDef
			}

			// Validate field value against the field definition rules.
			if resolvedFieldDef != nil {
				fe := validation.ValidateField(validation.FieldInput{
					FieldID:    resolvedFieldDef.FieldID,
					Label:      resolvedFieldDef.Label,
					FieldType:  resolvedFieldDef.Type,
					Value:      fu.Value,
					Validation: "", // TODO: resolve config from validation table via resolvedFieldDef.ValidationID
					Data:       resolvedFieldDef.Data,
				})
				if fe != nil {
					resp.Errors = append(resp.Errors, fe.Error())
					continue
				}
			}

			if fu.ContentFieldID != "" {
				// Update existing ContentField
				cfID := types.ContentFieldID(fu.ContentFieldID)
				existing, getErr := driver.GetContentField(cfID)
				if getErr != nil {
					resp.Errors = append(resp.Errors, fmt.Sprintf("get content field %s: %v", fu.ContentFieldID, getErr))
					continue
				}
				_, updateErr := driver.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
					ContentFieldID: cfID,
					ContentDataID:  existing.ContentDataID,
					FieldID:        existing.FieldID,
					FieldValue:     fu.Value,
					AuthorID:       user.UserID,
					RouteID:        existing.RouteID,
					DateCreated:    existing.DateCreated,
					DateModified:   now,
				})
				if updateErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("tree-save: update content field %s failed", fu.ContentFieldID), updateErr)
					resp.Errors = append(resp.Errors, fmt.Sprintf("update field %s: %v", fu.ContentFieldID, updateErr))
					continue
				}
			} else if fu.FieldID != "" {
				// No content_field_id provided — look up the existing content field
				// by content_data + field_id and update it.
				if fieldMap, ok := existingFieldsByCD[cdIDStr]; ok {
					if existingCF, found := fieldMap[fu.FieldID]; found {
						_, updateErr := driver.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
							ContentFieldID: existingCF.ContentFieldID,
							ContentDataID:  existingCF.ContentDataID,
							FieldID:        existingCF.FieldID,
							FieldValue:     fu.Value,
							AuthorID:       user.UserID,
							RouteID:        existingCF.RouteID,
							DateCreated:    existingCF.DateCreated,
							DateModified:   now,
						})
						if updateErr != nil {
							utility.DefaultLogger.Error(fmt.Sprintf("tree-save: upsert-update content field %s failed", existingCF.ContentFieldID), updateErr)
							resp.Errors = append(resp.Errors, fmt.Sprintf("upsert field %s: %v", existingCF.ContentFieldID, updateErr))
							continue
						}
						resp.FieldsUpdated++
						continue
					}
				}

				// Only create content_fields for blocks that were just created in this
				// request (they bypass autoCreateFields). Existing blocks should always
				// have their fields — a missing field is a data integrity issue, not a
				// create opportunity.
				_, isNewBlock := idMap[fu.ContentDataID]
				if !isNewBlock {
					utility.DefaultLogger.Warn(fmt.Sprintf("tree-save: no existing content_field for content_data=%s field=%s, skipping", cdIDStr, fu.FieldID), fmt.Errorf("missing content_field"))
					resp.Errors = append(resp.Errors, fmt.Sprintf("no existing content_field for content_data=%s field=%s", cdIDStr, fu.FieldID))
					continue
				}

				created, createErr := driver.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					ContentDataID: types.NullableContentID{ID: types.ContentID(cdIDStr), Valid: true},
					FieldID:       types.NullableFieldID{ID: types.FieldID(fu.FieldID), Valid: true},
					FieldValue:    fu.Value,
					AuthorID:      user.UserID,
					RouteID:       parentRouteID,
					DateCreated:   now,
					DateModified:  now,
				})
				if createErr != nil {
					utility.DefaultLogger.Error(fmt.Sprintf("tree-save: create content field for %s failed", cdIDStr), createErr)
					resp.Errors = append(resp.Errors, fmt.Sprintf("create field for %s: %v", cdIDStr, createErr))
					continue
				}
				if created != nil && created.ContentFieldID != "" {
					if fieldMap, ok := existingFieldsByCD[cdIDStr]; ok {
						fieldMap[fu.FieldID] = *created
					}
				}
			}
			resp.FieldsUpdated++
		}

		log.Info("tree-save: complete",
			"created", resp.Created,
			"updated", resp.Updated,
			"deleted", resp.Deleted,
			"fieldsUpdated", resp.FieldsUpdated,
			"errors", len(resp.Errors),
			"idMapSize", len(resp.IDMap))
		if len(resp.Errors) > 0 {
			for _, e := range resp.Errors {
				log.Debug("tree-save: error", "detail", e)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Encode error is non-recoverable (client disconnected); response already partially written.
		json.NewEncoder(w).Encode(resp)
	}
}

// DatatypeFieldsJSONHandler returns the field definitions for a datatype as JSON.
// Used by the block editor when a new block is created client-side and needs
// empty field inputs for the side panel.
// GET /admin/api/datatypes/{id}/fields
func DatatypeFieldsJSONHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Datatype ID required", http.StatusBadRequest)
			return
		}

		dtID := types.DatatypeID(id)
		if validateErr := dtID.Validate(); validateErr != nil {
			http.Error(w, "Invalid datatype ID", http.StatusBadRequest)
			return
		}

		fields, fieldsErr := svc.Schema.ListFieldsByDatatypeID(
			r.Context(),
			types.NullableDatatypeID{ID: dtID, Valid: true},
		)
		if fieldsErr != nil {
			utility.DefaultLogger.Error("failed to list fields for datatype", fieldsErr)
			http.Error(w, "failed to load fields", http.StatusInternalServerError)
			return
		}

		// Filter fields by the authenticated user's role.
		dtUser := middleware.AuthenticatedUser(r.Context())
		dtIsAdmin := middleware.ContextIsAdmin(r.Context())
		dtRoleID := ""
		if dtUser != nil {
			dtRoleID = dtUser.Role
		}

		filtered := db.FilterFieldsByRole(fields, dtRoleID, dtIsAdmin)
		result := make([]blockFieldData, 0, len(filtered))
		for _, f := range filtered {
			bfd := blockFieldData{
				FieldID: f.FieldID.String(),
				Label:   f.Label,
				Type:    string(f.Type),
			}
			if f.Type == types.FieldTypeRichText {
				rtCfg, parseErr := types.ParseRichTextConfig(f.Data)
				if parseErr == nil && len(rtCfg.Toolbar) > 0 {
					bfd.Toolbar = rtCfg.Toolbar
				}
			}
			result = append(result, bfd)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// ContentPublishHandler publishes content by creating a snapshot version.
// On success, re-renders the content edit page with updated status.
func ContentPublishHandler(driver db.DbDriver, mgr *config.Manager, dispatcher publishing.WebhookDispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		contentID := types.ContentID(id)
		locale := r.URL.Query().Get("locale")
		publishAll := !cfg.Node_Level_Publish
		_, pubErr := publishing.PublishContent(r.Context(), driver, contentID, locale, user.UserID, ac, cfg.VersionMaxPerContent(), publishAll, dispatcher, nil)
		if pubErr != nil {
			utility.DefaultLogger.Error("admin publish content failed", pubErr)
			toastMsg := fmt.Sprintf(`{"showToast": {"message": "Publish failed: %s", "type": "error"}}`, pubErr.Error())
			w.Header().Set("HX-Trigger", toastMsg)
			// Re-render the edit page to show current state
			renderContentEditPage(w, r, driver, contentID)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content published", "type": "success"}}`)
		renderContentEditPage(w, r, driver, contentID)
	}
}

// ContentUnpublishHandler unpublishes content by clearing the published flag.
// On success, re-renders the content edit page with updated status.
func ContentUnpublishHandler(driver db.DbDriver, mgr *config.Manager, dispatcher publishing.WebhookDispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		contentID := types.ContentID(id)
		locale := r.URL.Query().Get("locale")
		unpubErr := publishing.UnpublishContent(r.Context(), driver, contentID, locale, user.UserID, ac, dispatcher, nil)
		if unpubErr != nil {
			utility.DefaultLogger.Error("admin unpublish content failed", unpubErr)
			toastMsg := fmt.Sprintf(`{"showToast": {"message": "Unpublish failed: %s", "type": "error"}}`, unpubErr.Error())
			w.Header().Set("HX-Trigger", toastMsg)
			renderContentEditPage(w, r, driver, contentID)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content unpublished", "type": "success"}}`)
		renderContentEditPage(w, r, driver, contentID)
	}
}

// ContentVersionsHandler returns the version list partial for a content item.
func ContentVersionsHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		contentID := types.ContentID(id)
		versions, err := driver.ListContentVersionsByContent(contentID)
		if err != nil {
			utility.DefaultLogger.Error("failed to list versions", err)
			http.Error(w, "failed to load versions", http.StatusInternalServerError)
			return
		}

		var versionList []db.ContentVersion
		if versions != nil {
			versionList = *versions
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, partials.VersionList(versionList, id, csrfToken))
	}
}

// ContentCreateVersionHandler creates a manual snapshot version for a content item.
func ContentCreateVersionHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		contentID := types.ContentID(id)
		locale := r.URL.Query().Get("locale")

		// Build snapshot from live tables
		snapshot, snapErr := publishing.BuildSnapshot(driver, r.Context(), contentID, locale)
		if snapErr != nil {
			utility.DefaultLogger.Error("failed to build snapshot for manual version", snapErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create version", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		snapshotBytes, marshalErr := json.Marshal(snapshot)
		if marshalErr != nil {
			utility.DefaultLogger.Error("failed to marshal snapshot", marshalErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create version", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		maxVersion, maxErr := driver.GetMaxVersionNumber(contentID, locale)
		if maxErr != nil {
			utility.DefaultLogger.Error("failed to get max version number", maxErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create version", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)
		now := types.TimestampNow()

		label := fmt.Sprintf("v%d-%s", maxVersion+1, id[:8])

		_, createErr := driver.CreateContentVersion(r.Context(), ac, db.CreateContentVersionParams{
			ContentDataID: contentID,
			VersionNumber: maxVersion + 1,
			Locale:        locale,
			Snapshot:      string(snapshotBytes),
			Trigger:       "manual",
			Label:         label,
			Published:     false,
			PublishedBy:   types.NullableUserID{ID: user.UserID, Valid: true},
			DateCreated:   now,
		})
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create manual version", createErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create version", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Async prune
		retentionCap := cfg.VersionMaxPerContent()
		go publishing.PruneExcessVersions(driver, contentID, "", retentionCap)

		// Re-render the version list
		versions, listErr := driver.ListContentVersionsByContent(contentID)
		if listErr != nil {
			utility.DefaultLogger.Error("failed to re-list versions after create", listErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Version created", "type": "success"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}

		var versionList []db.ContentVersion
		if versions != nil {
			versionList = *versions
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Version created", "type": "success"}}`)
		Render(w, r, partials.VersionList(versionList, id, csrfToken))
	}
}

// ContentRestoreVersionHandler restores content from a saved version snapshot.
func ContentRestoreVersionHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		versionID := r.FormValue("version_id")
		if versionID == "" {
			http.Error(w, "version_id required", http.StatusBadRequest)
			return
		}

		contentID := types.ContentID(id)
		cvID := types.ContentVersionID(versionID)

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		result, restoreErr := publishing.RestoreContent(r.Context(), driver, contentID, cvID, user.UserID, ac)
		if restoreErr != nil {
			utility.DefaultLogger.Error("admin restore content failed", restoreErr)
			toastMsg := fmt.Sprintf(`{"showToast": {"message": "restore failed: %s", "type": "error"}}`, restoreErr.Error())
			w.Header().Set("HX-Trigger", toastMsg)
			renderContentEditPage(w, r, driver, contentID)
			return
		}

		toastMsg := fmt.Sprintf(`{"showToast": {"message": "Restored %d fields from version", "type": "success"}}`, result.FieldsRestored)
		w.Header().Set("HX-Trigger", toastMsg)
		renderContentEditPage(w, r, driver, contentID)
	}
}

// ContentVersionCompareHandler returns a side-by-side diff of two version snapshots.
func ContentVersionCompareHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Content ID required", http.StatusBadRequest)
			return
		}

		aID := r.URL.Query().Get("a")
		bID := r.URL.Query().Get("b")
		if aID == "" || bID == "" {
			http.Error(w, "Both version IDs (a and b) are required", http.StatusBadRequest)
			return
		}

		versionA, errA := driver.GetContentVersion(types.ContentVersionID(aID))
		if errA != nil {
			utility.DefaultLogger.Error("failed to get version A", errA)
			http.Error(w, "failed to load version A", http.StatusInternalServerError)
			return
		}

		versionB, errB := driver.GetContentVersion(types.ContentVersionID(bID))
		if errB != nil {
			utility.DefaultLogger.Error("failed to get version B", errB)
			http.Error(w, "failed to load version B", http.StatusInternalServerError)
			return
		}

		Render(w, r, partials.VersionDiff(*versionA, *versionB))
	}
}

// renderContentEditPage is a shared helper that re-renders the content edit page.
// Used by publish/unpublish/restore handlers to refresh the page after mutations.
// If the request originated from the content tree, redirects back there instead
// of rendering the old block editor page.
func renderContentEditPage(w http.ResponseWriter, r *http.Request, driver db.DbDriver, contentID types.ContentID) {
	// If the request came from the content tree page, redirect back to it
	currentURL := r.Header.Get("HX-Current-URL")
	if strings.Contains(currentURL, "/admin/content-tree/") {
		w.Header().Set("HX-Redirect", "/admin/content-tree/page/"+contentID.String())
		w.WriteHeader(http.StatusOK)
		return
	}
	content, err := driver.GetContentData(contentID)
	if err != nil {
		utility.DefaultLogger.Error("failed to get content for re-render", err)
		http.Error(w, "failed to load content", http.StatusInternalServerError)
		return
	}

	fields, fieldsErr := driver.ListContentFieldsWithFieldByContentData(
		types.NullableContentID{ID: content.ContentDataID, Valid: true},
	)
	if fieldsErr != nil {
		utility.DefaultLogger.Error("failed to get content fields for re-render", fieldsErr)
		http.Error(w, "failed to load content fields", http.StatusInternalServerError)
		return
	}

	var contentFields []db.ContentFieldWithFieldRow
	if fields != nil {
		contentFields = *fields
	}

	// Filter content fields by the authenticated user's role.
	reUser := middleware.AuthenticatedUser(r.Context())
	reIsAdmin := middleware.ContextIsAdmin(r.Context())
	reRoleID := ""
	if reUser != nil {
		reRoleID = reUser.Role
	}
	if !reIsAdmin && len(contentFields) > 0 {
		accessible := make([]db.ContentFieldWithFieldRow, 0, len(contentFields))
		for _, cf := range contentFields {
			if !cf.FieldID.Valid {
				accessible = append(accessible, cf)
				continue
			}
			fieldDef, fieldErr := driver.GetField(cf.FieldID.ID)
			if fieldErr != nil {
				continue
			}
			if db.IsFieldAccessible(*fieldDef, reRoleID, reIsAdmin) {
				accessible = append(accessible, cf)
			}
		}
		contentFields = accessible
	}

	meta := pages.ContentMeta{}
	if content.DatatypeID.Valid {
		dt, dtErr := driver.GetDatatype(content.DatatypeID.ID)
		if dtErr == nil && dt != nil {
			meta.DatatypeLabel = dt.Label
		}
	}
	if content.RouteID.Valid {
		rt, rtErr := driver.GetRoute(content.RouteID.ID)
		if rtErr == nil && rt != nil {
			meta.RouteTitle = rt.Title
		}
	}
	author, authorErr := driver.GetUser(content.AuthorID)
	if authorErr == nil && author != nil {
		meta.AuthorName = author.Username
	}
	meta.HasPublishPerm = HasPermission(r, "content:publish")
	if content.PublishedBy.Valid {
		pubUser, pubErr := driver.GetUser(content.PublishedBy.ID)
		if pubErr == nil && pubUser != nil {
			meta.PublishedByName = pubUser.Username
		}
	}

	treeJSON := buildTreeJSON(r.Context(), driver, content.ContentDataID, reRoleID, reIsAdmin)
	csrfToken := CSRFTokenFromContext(r.Context())

	Render(w, r, pages.ContentEditContent(*content, contentFields, csrfToken, treeJSON, meta))
}
