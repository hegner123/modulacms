package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminDatatypesListHandler handles GET /admin/admin-schema/datatypes.
// Lists all admin datatypes with pagination, search, type filter, and sorting.
// HTMX requests receive partial table rows only.
func AdminDatatypesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		typeFilter := r.URL.Query().Get("type")
		sortBy := r.URL.Query().Get("sort")

		// Fetch all admin datatypes and filter in Go (table is small).
		all, err := svc.Schema.ListAdminDatatypes(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin datatypes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Apply type filter first
		list := make([]db.AdminDatatypes, 0)
		for _, dt := range all {
			if typeFilter != "" && dt.Type != typeFilter {
				continue
			}
			list = append(list, dt)
		}

		// Fuzzy search: ranked results supersede user sort (intentional --
		// relevance order is more useful than alphabetical when searching).
		if search != "" {
			results := utility.FuzzyFind(search, list, func(dt db.AdminDatatypes) []string {
				return []string{dt.Name, dt.Label}
			})
			ranked := make([]db.AdminDatatypes, len(results))
			for i, r := range results {
				ranked[i] = list[r.Index]
			}
			list = ranked
			sortBy = ""
		}

		// Sort (skipped when fuzzy search is active)
		switch sortBy {
		case "name-asc":
			sortAdminDatatypes(list, func(a, b db.AdminDatatypes) bool { return strings.ToLower(a.Name) < strings.ToLower(b.Name) })
		case "name-desc":
			sortAdminDatatypes(list, func(a, b db.AdminDatatypes) bool { return strings.ToLower(a.Name) > strings.ToLower(b.Name) })
		case "modified-asc":
			sortAdminDatatypes(list, func(a, b db.AdminDatatypes) bool { return a.DateModified.Time.Before(b.DateModified.Time) })
		case "modified-desc":
			sortAdminDatatypes(list, func(a, b db.AdminDatatypes) bool { return a.DateModified.Time.After(b.DateModified.Time) })
		case "type-asc":
			sortAdminDatatypes(list, func(a, b db.AdminDatatypes) bool { return a.Type < b.Type })
		case "type-desc":
			sortAdminDatatypes(list, func(a, b db.AdminDatatypes) bool { return a.Type > b.Type })
		}

		// Paginate the filtered results
		limit, offset := ParsePagination(r)
		total := int64(len(list))
		off := int(offset)
		lim := int(limit)
		if off < len(list) {
			end := off + lim
			if end > len(list) {
				end = len(list)
			}
			list = list[off:end]
		} else {
			list = nil
		}

		// Collect distinct types for the filter dropdown
		typeSet := make(map[string]bool)
		for _, dt := range all {
			if dt.Type != "" {
				typeSet[dt.Type] = true
			}
		}
		distinctTypes := make([]string, 0, len(typeSet))
		for t := range typeSet {
			distinctTypes = append(distinctTypes, t)
		}
		sortStrings(distinctTypes)

		pd := NewPaginationData(total, limit, offset, "#admin-datatypes-table-body", "/admin/admin-schema/datatypes")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Admin Datatypes"}`)
			RenderWithOOB(w, r, pages.AdminDatatypesListContent(list, pg, distinctTypes),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.AdminDatatypeCreateDialog(csrfToken, all)})
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.AdminDatatypesTableRows(list, pg))
			return
		}

		layout := NewAdminData(r, "Admin Datatypes")
		Render(w, r, pages.AdminDatatypesList(layout, list, pg, distinctTypes, all))
	}
}

func sortAdminDatatypes(s []db.AdminDatatypes, less func(a, b db.AdminDatatypes) bool) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && less(s[j], s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// AdminDatatypeDetailHandler handles GET /admin/admin-schema/datatypes/{id}.
// Shows admin datatype detail with linked admin fields list.
func AdminDatatypeDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing admin datatype ID", http.StatusBadRequest)
			return
		}

		dt, err := svc.Schema.GetAdminDatatype(r.Context(), types.AdminDatatypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get admin datatype", err)
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin datatype not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Fetch linked admin fields by parent_id (fields belong to datatype directly)
		linkedFieldsPtr, fieldErr := svc.Driver().ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(id), Valid: true})
		var linkedFields []db.AdminFields
		if fieldErr != nil {
			utility.DefaultLogger.Error("failed to list admin datatype fields", fieldErr)
			linkedFields = []db.AdminFields{}
		} else if linkedFieldsPtr != nil {
			linkedFields = *linkedFieldsPtr
		} else {
			linkedFields = []db.AdminFields{}
		}

		// Fetch all admin datatypes for the parent select dropdown
		allDatatypes, dtErr := svc.Schema.ListAdminDatatypes(r.Context())
		if dtErr != nil {
			utility.DefaultLogger.Error("failed to list admin datatypes for parent select", dtErr)
			allDatatypes = []db.AdminDatatypes{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Admin Datatype: "+dt.Label)

		if IsNavHTMX(r) {
			safeTitle := "Admin Datatype: " + dt.Label
			w.Header().Set("HX-Trigger", `{"pageTitle": "`+safeTitle+`"}`)
			RenderWithOOB(w, r, pages.AdminDatatypeDetailContent(*dt, linkedFields, allDatatypes, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.AdminDatatypeAddFieldDialog(dt.AdminDatatypeID.String(), csrfToken)})
			return
		}
		Render(w, r, pages.AdminDatatypeDetail(layout, *dt, linkedFields, allDatatypes, csrfToken))
	}
}

// AdminDatatypeCreateHandler handles POST /admin/admin-schema/datatypes.
// Creates an admin datatype via the service layer, which validates label and type.
func AdminDatatypeCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		label := strings.TrimSpace(r.FormValue("label"))
		dtype := strings.TrimSpace(r.FormValue("type"))
		parentIDStr := strings.TrimSpace(r.FormValue("parent_id"))

		var parentID types.NullableAdminDatatypeID
		if parentIDStr != "" {
			parentID = types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(parentIDStr), Valid: true}
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		maxSort, sortErr := svc.Schema.GetMaxAdminDatatypeSortOrder(r.Context(), parentID)
		if sortErr != nil {
			maxSort = -1
		}
		_, err := svc.Schema.CreateAdminDatatype(r.Context(), ac, db.CreateAdminDatatypeParams{
			ParentID:  parentID,
			SortOrder: maxSort + 1,
			Name:      name,
			Label:     label,
			Type:      dtype,
			AuthorID:  user.UserID,
		})
		if err != nil {
			allDatatypes, _ := svc.Schema.ListAdminDatatypes(r.Context())
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.AdminDatatypeForm(name, label, dtype, parentIDStr, allDatatypes, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create admin datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminDatatypeForm(name, label, dtype, parentIDStr, allDatatypes, map[string]string{"_": "Failed to create admin datatype"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-schema/datatypes", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin datatype created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-schema/datatypes")
		w.WriteHeader(http.StatusOK)
	}
}

// AdminDatatypeUpdateHandler handles POST /admin/admin-schema/datatypes/{id}.
// Updates user-editable fields via the service layer, which preserves immutable fields.
func AdminDatatypeUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing admin datatype ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		label := strings.TrimSpace(r.FormValue("label"))
		dtype := strings.TrimSpace(r.FormValue("type"))
		parentIDStr := strings.TrimSpace(r.FormValue("parent_id"))

		var parentID types.NullableAdminDatatypeID
		if parentIDStr != "" {
			parentID = types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(parentIDStr), Valid: true}
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err := svc.Schema.UpdateAdminDatatype(r.Context(), ac, db.UpdateAdminDatatypeParams{
			AdminDatatypeID: types.AdminDatatypeID(id),
			ParentID:        parentID,
			Name:            name,
			Label:           label,
			Type:            dtype,
		})
		if err != nil {
			allDatatypes, _ := svc.Schema.ListAdminDatatypes(r.Context())
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.AdminDatatypeEditForm(id, name, label, dtype, parentIDStr, allDatatypes, errs, csrfToken))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin datatype not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update admin datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminDatatypeEditForm(id, name, label, dtype, parentIDStr, allDatatypes, map[string]string{"_": "Failed to update admin datatype"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-schema/datatypes/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin datatype updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-schema/datatypes/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminDatatypeDeleteHandler handles DELETE /admin/admin-schema/datatypes/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func AdminDatatypeDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing admin datatype ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin datatype", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Schema.DeleteAdminDatatype(r.Context(), ac, types.AdminDatatypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete admin datatype", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin datatype", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin datatype deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminDatatypeCreateFieldHandler handles POST /admin/admin-schema/datatypes/{id}/fields.
// Creates a new admin field with parent_id set to this admin datatype via the service layer.
func AdminDatatypeCreateFieldHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing admin datatype ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		label := strings.TrimSpace(r.FormValue("label"))
		fieldType := strings.TrimSpace(r.FormValue("type"))
		data := strings.TrimSpace(r.FormValue("data"))
		validationJSON := strings.TrimSpace(r.FormValue("validation"))
		uiConfig := strings.TrimSpace(r.FormValue("ui_config"))

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		_, err := svc.Schema.CreateAdminField(r.Context(), ac, db.CreateAdminFieldParams{
			ParentID:   types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(id), Valid: true},
			Name:       name,
			Label:      label,
			Type:       types.FieldType(fieldType),
			Data:       data,
			Validation: validationJSON,
			UIConfig:   uiConfig,
			AuthorID:   types.NullableUserID{ID: user.UserID, Valid: true},
		})
		if err != nil {
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.AdminDatatypeCreateFieldForm(id, name, label, fieldType, data, validationJSON, uiConfig, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create admin field for admin datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminDatatypeCreateFieldForm(id, name, label, fieldType, data, validationJSON, uiConfig, map[string]string{"_": "Failed to create field"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-schema/datatypes/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-schema/datatypes/"+id)
		w.WriteHeader(http.StatusOK)
	}
}
