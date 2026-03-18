package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// resolveAdminContentDisplayName returns a human-readable name for an admin content item.
// Priority: route title > datatype label > truncated ID.
func resolveAdminContentDisplayName(item db.AdminContentDataTopLevel) string {
	if item.RouteTitle != "" {
		return item.RouteTitle
	}
	if item.DatatypeLabel != "" {
		return item.DatatypeLabel
	}
	id := item.AdminContentDataID.String()
	if len(id) > 12 {
		return id[:8] + "..."
	}
	return id
}

// AdminContentListHandler lists admin content with pagination.
// HTMX requests return partial table rows; full requests include the complete page layout.
func AdminContentListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)
		driver := svc.Driver()

		items, err := driver.ListAdminContentDataTopLevelPaginated(db.PaginationParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin content", err)
			http.Error(w, "Failed to load admin content", http.StatusInternalServerError)
			return
		}

		var rawItems []db.AdminContentDataTopLevel
		if items != nil {
			rawItems = *items
		}

		cnt, cntErr := driver.CountAdminContentDataTopLevel()
		if cntErr != nil {
			utility.DefaultLogger.Error("failed to count admin content", cntErr)
			http.Error(w, "Failed to load admin content", http.StatusInternalServerError)
			return
		}

		hasPublishPerm := HasPermission(r, "admin_content:publish")

		listItems := make([]pages.AdminContentListItem, len(rawItems))
		for i, item := range rawItems {
			listItems[i] = pages.AdminContentListItem{
				AdminContentDataTopLevel: item,
				HasPublishPerm:           hasPublishPerm,
			}
			listItems[i].DisplayName = resolveAdminContentDisplayName(item)
		}

		pd := NewPaginationData(*cnt, limit, offset, "#admin-content-table-body", "/admin/admin-content")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsHTMX(r) && !IsNavHTMX(r) {
			Render(w, r, pages.AdminContentTableRowsPartial(listItems, pg))
			return
		}

		// Load admin datatypes for the create dialog
		var datatypes []db.AdminDatatypes
		dtList, dtErr := svc.Schema.ListAdminDatatypes(r.Context())
		if dtErr != nil {
			utility.DefaultLogger.Error("failed to list admin datatypes for create dialog", dtErr)
		} else {
			datatypes = dtList
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Admin Content"}`)
			RenderWithOOB(w, r, pages.AdminContentListContent(listItems, pg),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.AdminContentCreateDialog(datatypes, csrfToken)})
			return
		}

		layout := NewAdminData(r, "Admin Content")
		Render(w, r, pages.AdminContentList(layout, listItems, pg, datatypes))
	}
}

// AdminContentEditHandler renders the admin content edit page.
// Shows metadata, status, publish controls, and version history access.
func AdminContentEditHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Content ID required", http.StatusBadRequest)
			return
		}

		content, err := svc.AdminContent.Get(r.Context(), types.AdminContentID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get admin content", err)
			http.NotFound(w, r)
			return
		}

		// Resolve datatype label for display
		datatypeLabel := ""
		if content.AdminDatatypeID.Valid {
			dt, dtErr := svc.Schema.GetAdminDatatype(r.Context(), content.AdminDatatypeID.ID)
			if dtErr == nil && dt != nil {
				datatypeLabel = dt.Label
			}
		}

		hasPublishPerm := HasPermission(r, "admin_content:publish")
		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Edit Admin Content")

		RenderNav(w, r, "Edit Admin Content",
			pages.AdminContentEditContent(*content, datatypeLabel, csrfToken, hasPublishPerm),
			pages.AdminContentEdit(layout, *content, datatypeLabel, csrfToken, hasPublishPerm),
		)
	}
}

// AdminContentCreateHandler creates new admin content from a form submission.
// On success, HTMX requests receive an HX-Trigger toast and redirect.
func AdminContentCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		datatypeID := r.FormValue("admin_datatype_id")
		now := types.NewTimestamp(time.Now())

		params := db.CreateAdminContentDataParams{
			AdminDatatypeID: types.NullableAdminDatatypeID{
				ID:    types.AdminDatatypeID(datatypeID),
				Valid: datatypeID != "",
			},
			AuthorID:     user.UserID,
			Status:       types.ContentStatusDraft,
			DateCreated:  now,
			DateModified: now,
		}

		created, createErr := svc.AdminContent.Create(r.Context(), ac, params)
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create admin content", createErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to create admin content", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to create admin content", http.StatusInternalServerError)
			return
		}

		// Create content field rows for every field defined on the admin datatype.
		if datatypeID != "" {
			dtFields, dtFieldsErr := svc.Driver().ListAdminFieldsByDatatypeID(
				types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(datatypeID), Valid: true},
			)
			if dtFieldsErr != nil {
				utility.DefaultLogger.Error("failed to list admin datatype fields for content field creation", dtFieldsErr)
			} else if dtFields != nil {
				for _, f := range *dtFields {
					_, cfErr := svc.Driver().CreateAdminContentField(r.Context(), ac, db.CreateAdminContentFieldParams{
						AdminContentDataID: types.NullableAdminContentID{ID: created.AdminContentDataID, Valid: true},
						AdminFieldID:       types.NullableAdminFieldID{ID: f.AdminFieldID, Valid: true},
						AdminFieldValue:    "",
						AuthorID:           user.UserID,
						DateCreated:        now,
						DateModified:       now,
					})
					if cfErr != nil {
						utility.DefaultLogger.Error("failed to create admin content field for "+f.Label, cfErr)
					}
				}
			}
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin content created", "type": "success"}}`)
			w.Header().Set("HX-Redirect", "/admin/admin-content/"+created.AdminContentDataID.String())
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/admin-content/"+created.AdminContentDataID.String(), http.StatusSeeOther)
	}
}

// AdminContentUpdateHandler updates existing admin content from a form submission.
func AdminContentUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Content ID required", http.StatusBadRequest)
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

		existing, getErr := svc.AdminContent.Get(r.Context(), types.AdminContentID(id))
		if getErr != nil {
			utility.DefaultLogger.Error("admin content not found for update", getErr)
			http.NotFound(w, r)
			return
		}

		status := types.ContentStatus(r.FormValue("status"))
		if status == "" {
			status = existing.Status
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		params := db.UpdateAdminContentDataParams{
			AdminContentDataID: existing.AdminContentDataID,
			ParentID:           existing.ParentID,
			FirstChildID:       existing.FirstChildID,
			NextSiblingID:      existing.NextSiblingID,
			PrevSiblingID:      existing.PrevSiblingID,
			RootID:             existing.RootID,
			AdminRouteID:       existing.AdminRouteID,
			AdminDatatypeID:    existing.AdminDatatypeID,
			AuthorID:           existing.AuthorID,
			Status:             status,
			DateCreated:        existing.DateCreated,
			DateModified:       types.NewTimestamp(time.Now()),
		}

		if _, updateErr := svc.AdminContent.Update(r.Context(), ac, params, existing.Revision); updateErr != nil {
			utility.DefaultLogger.Error("failed to update admin content", updateErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update admin content", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to update admin content", http.StatusInternalServerError)
			return
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin content updated", "type": "success"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/admin-content/"+id, http.StatusSeeOther)
	}
}

// AdminContentDeleteHandler deletes admin content by ID.
// Only HTMX DELETE requests are supported; non-HTMX requests receive 405.
func AdminContentDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Content ID required", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		if _, deleteErr := svc.AdminContent.Delete(r.Context(), ac, types.AdminContentID(id), false); deleteErr != nil {
			utility.DefaultLogger.Error("failed to delete admin content", deleteErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin content", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin content deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminContentPublishHandler publishes admin content.
// On success, re-renders the edit page with updated status.
func AdminContentPublishHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Content ID required", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		adminContentID := types.AdminContentID(id)
		pubErr := svc.AdminContent.Publish(r.Context(), ac, adminContentID, "", user.UserID)
		if pubErr != nil {
			utility.DefaultLogger.Error("admin publish admin content failed", pubErr)
			toastMsg := fmt.Sprintf(`{"showToast": {"message": "Publish failed: %s", "type": "error"}}`, pubErr.Error())
			w.Header().Set("HX-Trigger", toastMsg)
			renderAdminContentEditPage(w, r, svc, adminContentID)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin content published", "type": "success"}}`)
		renderAdminContentEditPage(w, r, svc, adminContentID)
	}
}

// AdminContentUnpublishHandler unpublishes admin content.
// On success, re-renders the edit page with updated status.
func AdminContentUnpublishHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Content ID required", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		adminContentID := types.AdminContentID(id)
		unpubErr := svc.AdminContent.Unpublish(r.Context(), ac, adminContentID, "", user.UserID)
		if unpubErr != nil {
			utility.DefaultLogger.Error("admin unpublish admin content failed", unpubErr)
			toastMsg := fmt.Sprintf(`{"showToast": {"message": "Unpublish failed: %s", "type": "error"}}`, unpubErr.Error())
			w.Header().Set("HX-Trigger", toastMsg)
			renderAdminContentEditPage(w, r, svc, adminContentID)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin content unpublished", "type": "success"}}`)
		renderAdminContentEditPage(w, r, svc, adminContentID)
	}
}

// AdminContentVersionsHandler returns the version list partial for an admin content item.
func AdminContentVersionsHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Content ID required", http.StatusBadRequest)
			return
		}

		adminContentID := types.AdminContentID(id)
		versions, err := svc.AdminContent.ListVersions(r.Context(), adminContentID)
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin content versions", err)
			http.Error(w, "Failed to load versions", http.StatusInternalServerError)
			return
		}

		var versionList []db.AdminContentVersion
		if versions != nil {
			versionList = *versions
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, partials.AdminVersionList(versionList, id, csrfToken))
	}
}

// renderAdminContentEditPage re-renders the admin content edit page after a state change.
func renderAdminContentEditPage(w http.ResponseWriter, r *http.Request, svc *service.Registry, adminContentID types.AdminContentID) {
	content, err := svc.AdminContent.Get(r.Context(), adminContentID)
	if err != nil {
		utility.DefaultLogger.Error("failed to reload admin content for re-render", err)
		http.Error(w, "Failed to reload admin content", http.StatusInternalServerError)
		return
	}

	datatypeLabel := ""
	if content.AdminDatatypeID.Valid {
		dt, dtErr := svc.Schema.GetAdminDatatype(r.Context(), content.AdminDatatypeID.ID)
		if dtErr == nil && dt != nil {
			datatypeLabel = dt.Label
		}
	}

	hasPublishPerm := HasPermission(r, "admin_content:publish")
	csrfToken := CSRFTokenFromContext(r.Context())

	Render(w, r, pages.AdminContentEditContent(*content, datatypeLabel, csrfToken, hasPublishPerm))
}
