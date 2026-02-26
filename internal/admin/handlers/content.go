package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/tree/core"
	"github.com/hegner123/modulacms/internal/utility"
)

// clientIP extracts the client IP address from the request.
func clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// ContentListHandler lists content with pagination.
// HTMX requests return partial table rows; full requests include the complete page layout.
func ContentListHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, err := driver.ListContentDataPaginated(db.PaginationParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to list content", err)
			http.Error(w, "Failed to load content", http.StatusInternalServerError)
			return
		}

		total, err := driver.CountContentData()
		if err != nil {
			utility.DefaultLogger.Error("failed to count content", err)
			http.Error(w, "Failed to load content", http.StatusInternalServerError)
			return
		}

		var contentItems []db.ContentData
		if items != nil {
			contentItems = *items
		}

		pd := NewPaginationData(*total, limit, offset, "#content-table-body", "/admin/content")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		// Load tree nodes for sidebar via core.BuildFromRows
		var treeRoot *core.Root
		tree, treeErr := driver.GetContentTreeByRoute(types.NullableRouteID{})
		if treeErr == nil && tree != nil {
			built, _, buildErr := core.BuildFromRows(*tree)
			if buildErr != nil {
				utility.DefaultLogger.Warn("content tree build issue", buildErr)
			}
			treeRoot = built
		}

		if IsHTMX(r) {
			Render(w, r, pages.ContentTableRowsPartial(contentItems, pg))
			return
		}

		layout := NewAdminData(r, "Content")
		Render(w, r, pages.ContentList(layout, contentItems, pg, treeRoot))
	}
}

// ContentEditHandler renders the content editor page.
// Loads content by ID from the URL path and its associated fields.
func ContentEditHandler(driver db.DbDriver) http.HandlerFunc {
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
			http.Error(w, "Failed to load content fields", http.StatusInternalServerError)
			return
		}

		var contentFields []db.ContentFieldWithFieldRow
		if fields != nil {
			contentFields = *fields
		}

		layout := NewAdminData(r, "Edit Content")
		Render(w, r, pages.ContentEdit(layout, *content, contentFields))
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
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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
		routeID := r.FormValue("route_id")

		now := types.NewTimestamp(time.Now())
		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		params := db.CreateContentDataParams{
			ParentID:     types.NullableContentID{ID: types.ContentID(parentID), Valid: parentID != ""},
			DatatypeID:   types.NullableDatatypeID{ID: types.DatatypeID(datatypeID), Valid: datatypeID != ""},
			RouteID:      types.NullableRouteID{ID: types.RouteID(routeID), Valid: routeID != ""},
			AuthorID:     user.UserID,
			Status:       status,
			DateCreated:  now,
			DateModified: now,
		}

		created, createErr := driver.CreateContentData(r.Context(), ac, params)
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create content", createErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to create content", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to create content", http.StatusInternalServerError)
			return
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content created", "type": "success"}}`)
			w.Header().Set("HX-Redirect", "/admin/content/"+created.ContentDataID.String())
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/content/"+created.ContentDataID.String(), http.StatusSeeOther)
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
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update content", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to update content", http.StatusInternalServerError)
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
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete content", "type": "error"}}`)
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
// Accepts JSON with parent_id and ordered_ids, then updates sibling pointers
// for all nodes in the given order.
func ContentReorderHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "No items to reorder", "type": "warning"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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

		now := types.NewTimestamp(time.Now())

		// Update sibling pointers for each node in order.
		// First node: prev=nil, next=second. Last node: prev=second-to-last, next=nil.
		for i, idStr := range req.OrderedIDs {
			content, getErr := driver.GetContentData(types.ContentID(idStr))
			if getErr != nil {
				utility.DefaultLogger.Error("failed to get content for reorder", getErr)
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Reorder failed", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var prevID types.NullableContentID
			var nextID types.NullableContentID

			if i > 0 {
				prevID = types.NullableContentID{
					ID:    types.ContentID(req.OrderedIDs[i-1]),
					Valid: true,
				}
			}
			if i < len(req.OrderedIDs)-1 {
				nextID = types.NullableContentID{
					ID:    types.ContentID(req.OrderedIDs[i+1]),
					Valid: true,
				}
			}

			params := db.UpdateContentDataParams{
				ContentDataID: content.ContentDataID,
				ParentID:      content.ParentID,
				FirstChildID:  content.FirstChildID,
				NextSiblingID: nextID,
				PrevSiblingID: prevID,
				RouteID:       content.RouteID,
				DatatypeID:    content.DatatypeID,
				AuthorID:      content.AuthorID,
				Status:        content.Status,
				DateCreated:   content.DateCreated,
				DateModified:  now,
			}

			if _, updateErr := driver.UpdateContentData(r.Context(), ac, params); updateErr != nil {
				utility.DefaultLogger.Error("failed to update content during reorder", updateErr)
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Reorder failed", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// Update parent's first_child_id to the first item in the new order
		if req.ParentID != "" {
			parent, getErr := driver.GetContentData(types.ContentID(req.ParentID))
			if getErr == nil {
				params := db.UpdateContentDataParams{
					ContentDataID: parent.ContentDataID,
					ParentID:      parent.ParentID,
					FirstChildID: types.NullableContentID{
						ID:    types.ContentID(req.OrderedIDs[0]),
						Valid: true,
					},
					NextSiblingID: parent.NextSiblingID,
					PrevSiblingID: parent.PrevSiblingID,
					RouteID:       parent.RouteID,
					DatatypeID:    parent.DatatypeID,
					AuthorID:      parent.AuthorID,
					Status:        parent.Status,
					DateCreated:   parent.DateCreated,
					DateModified:  now,
				}
				if _, updateErr := driver.UpdateContentData(r.Context(), ac, params); updateErr != nil {
					utility.DefaultLogger.Error("failed to update parent first_child_id", updateErr)
				}
			}
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
// Detaches from old parent's sibling chain, then attaches to new parent.
func ContentMoveHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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

		now := types.NewTimestamp(time.Now())

		content, getErr := driver.GetContentData(types.ContentID(req.ContentID))
		if getErr != nil {
			utility.DefaultLogger.Error("content not found for move", getErr)
			http.NotFound(w, r)
			return
		}

		// Detach from old sibling chain: link prev and next siblings together
		if content.PrevSiblingID.Valid {
			prev, prevErr := driver.GetContentData(content.PrevSiblingID.ID)
			if prevErr == nil {
				prevParams := db.UpdateContentDataParams{
					ContentDataID: prev.ContentDataID,
					ParentID:      prev.ParentID,
					FirstChildID:  prev.FirstChildID,
					NextSiblingID: content.NextSiblingID,
					PrevSiblingID: prev.PrevSiblingID,
					RouteID:       prev.RouteID,
					DatatypeID:    prev.DatatypeID,
					AuthorID:      prev.AuthorID,
					Status:        prev.Status,
					DateCreated:   prev.DateCreated,
					DateModified:  now,
				}
				if _, uErr := driver.UpdateContentData(r.Context(), ac, prevParams); uErr != nil {
					utility.DefaultLogger.Error("failed to update prev sibling during move", uErr)
				}
			}
		}
		if content.NextSiblingID.Valid {
			next, nextErr := driver.GetContentData(content.NextSiblingID.ID)
			if nextErr == nil {
				nextParams := db.UpdateContentDataParams{
					ContentDataID: next.ContentDataID,
					ParentID:      next.ParentID,
					FirstChildID:  next.FirstChildID,
					NextSiblingID: next.NextSiblingID,
					PrevSiblingID: content.PrevSiblingID,
					RouteID:       next.RouteID,
					DatatypeID:    next.DatatypeID,
					AuthorID:      next.AuthorID,
					Status:        next.Status,
					DateCreated:   next.DateCreated,
					DateModified:  now,
				}
				if _, uErr := driver.UpdateContentData(r.Context(), ac, nextParams); uErr != nil {
					utility.DefaultLogger.Error("failed to update next sibling during move", uErr)
				}
			}
		}

		// If content was first child of old parent, update old parent's first_child_id
		if content.ParentID.Valid {
			oldParent, opErr := driver.GetContentData(content.ParentID.ID)
			if opErr == nil && oldParent.FirstChildID.Valid && oldParent.FirstChildID.ID == content.ContentDataID {
				opParams := db.UpdateContentDataParams{
					ContentDataID: oldParent.ContentDataID,
					ParentID:      oldParent.ParentID,
					FirstChildID:  content.NextSiblingID,
					NextSiblingID: oldParent.NextSiblingID,
					PrevSiblingID: oldParent.PrevSiblingID,
					RouteID:       oldParent.RouteID,
					DatatypeID:    oldParent.DatatypeID,
					AuthorID:      oldParent.AuthorID,
					Status:        oldParent.Status,
					DateCreated:   oldParent.DateCreated,
					DateModified:  now,
				}
				if _, uErr := driver.UpdateContentData(r.Context(), ac, opParams); uErr != nil {
					utility.DefaultLogger.Error("failed to update old parent first_child_id", uErr)
				}
			}
		}

		// Attach to new parent as first child (position 0) or at end
		newParentID := types.NullableContentID{
			ID:    types.ContentID(req.NewParentID),
			Valid: req.NewParentID != "",
		}

		moveParams := db.UpdateContentDataParams{
			ContentDataID: content.ContentDataID,
			ParentID:      newParentID,
			FirstChildID:  content.FirstChildID,
			NextSiblingID: types.NullableContentID{},
			PrevSiblingID: types.NullableContentID{},
			RouteID:       content.RouteID,
			DatatypeID:    content.DatatypeID,
			AuthorID:      content.AuthorID,
			Status:        content.Status,
			DateCreated:   content.DateCreated,
			DateModified:  now,
		}

		// If new parent exists, set as first child
		if newParentID.Valid {
			newParent, npErr := driver.GetContentData(newParentID.ID)
			if npErr == nil && newParent.FirstChildID.Valid {
				// New parent already has children; insert at beginning
				moveParams.NextSiblingID = newParent.FirstChildID

				// Update old first child's prev pointer
				oldFirst, ofErr := driver.GetContentData(newParent.FirstChildID.ID)
				if ofErr == nil {
					ofParams := db.UpdateContentDataParams{
						ContentDataID: oldFirst.ContentDataID,
						ParentID:      oldFirst.ParentID,
						FirstChildID:  oldFirst.FirstChildID,
						NextSiblingID: oldFirst.NextSiblingID,
						PrevSiblingID: types.NullableContentID{ID: content.ContentDataID, Valid: true},
						RouteID:       oldFirst.RouteID,
						DatatypeID:    oldFirst.DatatypeID,
						AuthorID:      oldFirst.AuthorID,
						Status:        oldFirst.Status,
						DateCreated:   oldFirst.DateCreated,
						DateModified:  now,
					}
					if _, uErr := driver.UpdateContentData(r.Context(), ac, ofParams); uErr != nil {
						utility.DefaultLogger.Error("failed to update old first child prev pointer", uErr)
					}
				}
			}

			if npErr == nil {
				// Update new parent's first_child_id
				npParams := db.UpdateContentDataParams{
					ContentDataID: newParent.ContentDataID,
					ParentID:      newParent.ParentID,
					FirstChildID:  types.NullableContentID{ID: content.ContentDataID, Valid: true},
					NextSiblingID: newParent.NextSiblingID,
					PrevSiblingID: newParent.PrevSiblingID,
					RouteID:       newParent.RouteID,
					DatatypeID:    newParent.DatatypeID,
					AuthorID:      newParent.AuthorID,
					Status:        newParent.Status,
					DateCreated:   newParent.DateCreated,
					DateModified:  now,
				}
				if _, uErr := driver.UpdateContentData(r.Context(), ac, npParams); uErr != nil {
					utility.DefaultLogger.Error("failed to update new parent first_child_id", uErr)
				}
			}
		}

		if _, updateErr := driver.UpdateContentData(r.Context(), ac, moveParams); updateErr != nil {
			utility.DefaultLogger.Error("failed to move content", updateErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Move failed", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Content moved", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
