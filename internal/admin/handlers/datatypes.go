package handlers

import (
	"encoding/json"
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

// DatatypesListHandler handles GET /admin/schema/datatypes.
// Lists all datatypes with pagination, search, type filter, and sorting.
// HTMX requests receive partial table rows only.
func DatatypesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		typeFilter := r.URL.Query().Get("type")
		sortBy := r.URL.Query().Get("sort")

		// Fetch all datatypes and filter in Go (table is small).
		all, err := svc.Schema.ListDatatypes(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list datatypes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Apply type filter first
		list := make([]db.Datatypes, 0)
		for _, dt := range all {
			if typeFilter != "" && dt.Type != typeFilter {
				continue
			}
			list = append(list, dt)
		}

		// Fuzzy search: ranked results supersede user sort (intentional —
		// relevance order is more useful than alphabetical when searching).
		if search != "" {
			results := utility.FuzzyFind(search, list, func(dt db.Datatypes) []string {
				return []string{dt.Name, dt.Label}
			})
			ranked := make([]db.Datatypes, len(results))
			for i, r := range results {
				ranked[i] = list[r.Index]
			}
			list = ranked
			sortBy = ""
		}

		// Sort (skipped when fuzzy search is active)
		switch sortBy {
		case "name-asc":
			sortDatatypes(list, func(a, b db.Datatypes) bool { return strings.ToLower(a.Name) < strings.ToLower(b.Name) })
		case "name-desc":
			sortDatatypes(list, func(a, b db.Datatypes) bool { return strings.ToLower(a.Name) > strings.ToLower(b.Name) })
		case "modified-asc":
			sortDatatypes(list, func(a, b db.Datatypes) bool { return a.DateModified.Time.Before(b.DateModified.Time) })
		case "modified-desc":
			sortDatatypes(list, func(a, b db.Datatypes) bool { return a.DateModified.Time.After(b.DateModified.Time) })
		case "type-asc":
			sortDatatypes(list, func(a, b db.Datatypes) bool { return a.Type < b.Type })
		case "type-desc":
			sortDatatypes(list, func(a, b db.Datatypes) bool { return a.Type > b.Type })
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

		pd := NewPaginationData(total, limit, offset, "#datatypes-table-body", "/admin/schema/datatypes")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Datatypes"}`)
			RenderWithOOB(w, r, pages.DatatypesListContent(list, pg, distinctTypes),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.DatatypeCreateDialog(csrfToken, all)})
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.DatatypesTableRows(list, pg))
			return
		}

		layout := NewAdminData(r, "Datatypes")
		Render(w, r, pages.DatatypesList(layout, list, pg, distinctTypes, all))
	}
}

func sortDatatypes(s []db.Datatypes, less func(a, b db.Datatypes) bool) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && less(s[j], s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// DatatypeDetailHandler handles GET /admin/schema/datatypes/{id}.
// Shows datatype detail with linked fields list.
func DatatypeDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing datatype ID", http.StatusBadRequest)
			return
		}

		dt, err := svc.Schema.GetDatatype(r.Context(), types.DatatypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get datatype", err)
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Datatype not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Fetch linked fields by parent_id (fields belong to datatype directly)
		linkedFields, err := svc.Schema.ListFieldsByDatatypeID(r.Context(), types.NullableDatatypeID{ID: types.DatatypeID(id), Valid: true})
		if err != nil {
			utility.DefaultLogger.Error("failed to list datatype fields", err)
			linkedFields = []db.Fields{}
		}

		// Fetch all datatypes for the parent select dropdown
		allDatatypes, dtErr := svc.Schema.ListDatatypes(r.Context())
		if dtErr != nil {
			utility.DefaultLogger.Error("failed to list datatypes for parent select", dtErr)
			allDatatypes = []db.Datatypes{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Datatype: "+dt.Label)

		if IsNavHTMX(r) {
			safeTitle := "Datatype: " + dt.Label
			w.Header().Set("HX-Trigger", `{"pageTitle": "`+safeTitle+`"}`)
			RenderWithOOB(w, r, pages.DatatypeDetailContent(*dt, linkedFields, allDatatypes, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.DatatypeAddFieldDialog(dt.DatatypeID.String(), csrfToken)})
			return
		}
		Render(w, r, pages.DatatypeDetail(layout, *dt, linkedFields, allDatatypes, csrfToken))
	}
}

// DatatypeCreateHandler handles POST /admin/schema/datatypes.
// Creates a datatype via the service layer, which validates label and type.
func DatatypeCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		label := strings.TrimSpace(r.FormValue("label"))
		dtype := strings.TrimSpace(r.FormValue("type"))
		parentIDStr := strings.TrimSpace(r.FormValue("parent_id"))

		var parentID types.NullableDatatypeID
		if parentIDStr != "" {
			parentID = types.NullableDatatypeID{ID: types.DatatypeID(parentIDStr), Valid: true}
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		maxSort, sortErr := svc.Schema.GetMaxDatatypeSortOrder(r.Context(), parentID)
		if sortErr != nil {
			maxSort = -1
		}
		_, err := svc.Schema.CreateDatatype(r.Context(), ac, db.CreateDatatypeParams{
			ParentID:  parentID,
			SortOrder: maxSort + 1,
			Name:      name,
			Label:     label,
			Type:      dtype,
			AuthorID:  user.UserID,
		})
		if err != nil {
			allDatatypes, _ := svc.Schema.ListDatatypes(r.Context())
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.DatatypeForm(name, label, dtype, parentIDStr, allDatatypes, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeForm(name, label, dtype, parentIDStr, allDatatypes, map[string]string{"_": "Failed to create datatype"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/schema/datatypes", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Datatype created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/schema/datatypes")
		w.WriteHeader(http.StatusOK)
	}
}

// DatatypeUpdateHandler handles POST /admin/schema/datatypes/{id}.
// Updates user-editable fields via the service layer, which preserves immutable fields.
func DatatypeUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing datatype ID", http.StatusBadRequest)
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

		var parentID types.NullableDatatypeID
		if parentIDStr != "" {
			parentID = types.NullableDatatypeID{ID: types.DatatypeID(parentIDStr), Valid: true}
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err := svc.Schema.UpdateDatatype(r.Context(), ac, db.UpdateDatatypeParams{
			DatatypeID: types.DatatypeID(id),
			ParentID:   parentID,
			Name:       name,
			Label:      label,
			Type:       dtype,
		})
		if err != nil {
			allDatatypes, _ := svc.Schema.ListDatatypes(r.Context())
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.DatatypeEditForm(id, name, label, dtype, parentIDStr, allDatatypes, errs, csrfToken))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Datatype not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeEditForm(id, name, label, dtype, parentIDStr, allDatatypes, map[string]string{"_": "Failed to update datatype"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/schema/datatypes/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Datatype updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/schema/datatypes/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// DatatypeDeleteHandler handles DELETE /admin/schema/datatypes/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func DatatypeDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing datatype ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete datatype", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Schema.DeleteDatatype(r.Context(), ac, types.DatatypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete datatype", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete datatype", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Datatype deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// DatatypeCreateFieldHandler handles POST /admin/schema/datatypes/{id}/fields.
// Creates a new field with parent_id set to this datatype via the service layer.
func DatatypeCreateFieldHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing datatype ID", http.StatusBadRequest)
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
		_, err := svc.Schema.CreateField(r.Context(), ac, db.CreateFieldParams{
			ParentID:   types.NullableDatatypeID{ID: types.DatatypeID(id), Valid: true},
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
				Render(w, r, partials.DatatypeCreateFieldForm(id, name, label, fieldType, data, validationJSON, uiConfig, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create field for datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeCreateFieldForm(id, name, label, fieldType, data, validationJSON, uiConfig, map[string]string{"_": "Failed to create field"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/schema/datatypes/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/schema/datatypes/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// DatatypesJSONHandler handles GET /admin/api/datatypes.
// Returns all datatypes as JSON for the block editor insert dialog.
func DatatypesJSONHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.Schema.ListDatatypes(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list datatypes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if encErr := json.NewEncoder(w).Encode(list); encErr != nil {
			utility.DefaultLogger.Error("failed to encode datatypes JSON", encErr)
		}
	}
}
