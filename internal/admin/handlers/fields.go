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

// FieldsListHandler handles GET /admin/fields.
// Lists all fields with pagination, search, and sorting.
// HTMX requests receive partial table rows only.
func FieldsListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		sortBy := r.URL.Query().Get("sort")

		user := middleware.AuthenticatedUser(r.Context())
		isAdmin := middleware.ContextIsAdmin(r.Context())
		roleID := ""
		if user != nil {
			roleID = user.Role
		}

		all, err := svc.Schema.ListFieldsFiltered(r.Context(), roleID, isAdmin)
		if err != nil {
			utility.DefaultLogger.Error("failed to list fields", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.Fields, len(all))
		copy(list, all)

		// Fuzzy search: ranked results supersede user sort.
		if search != "" {
			results := utility.FuzzyFind(search, list, func(f db.Fields) []string {
				return []string{f.Name, f.Label, f.Type.String()}
			})
			ranked := make([]db.Fields, len(results))
			for i, r := range results {
				ranked[i] = list[r.Index]
			}
			list = ranked
			sortBy = ""
		}

		// Sort (skipped when fuzzy search is active)
		switch sortBy {
		case "name-asc":
			sortFields(list, func(a, b db.Fields) bool { return strings.ToLower(a.Name) < strings.ToLower(b.Name) })
		case "name-desc":
			sortFields(list, func(a, b db.Fields) bool { return strings.ToLower(a.Name) > strings.ToLower(b.Name) })
		case "type-asc":
			sortFields(list, func(a, b db.Fields) bool { return a.Type.String() < b.Type.String() })
		case "type-desc":
			sortFields(list, func(a, b db.Fields) bool { return a.Type.String() > b.Type.String() })
		case "modified-desc":
			sortFields(list, func(a, b db.Fields) bool { return a.DateModified.Time.After(b.DateModified.Time) })
		case "modified-asc":
			sortFields(list, func(a, b db.Fields) bool { return a.DateModified.Time.Before(b.DateModified.Time) })
		}

		// Paginate
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

		pd := NewPaginationData(total, limit, offset, "#fields-table-body", "/admin/fields")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Fields"}`)
			Render(w, r, pages.FieldsListContent(list, pg))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.FieldsTableRows(list, pg))
			return
		}

		layout := NewAdminData(r, "Fields")
		Render(w, r, pages.FieldsList(layout, list, pg))
	}
}

func sortFields(s []db.Fields, less func(a, b db.Fields) bool) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && less(s[j], s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// fetchFieldTypes loads public field types from the database, returning an empty
// slice on error so callers can continue rendering without blocking the page.
func fetchFieldTypes(svc *service.Registry, r *http.Request) []db.FieldTypes {
	fieldTypes, err := svc.Schema.ListFieldTypes(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("failed to list field types for dropdown", err)
		return []db.FieldTypes{}
	}
	return fieldTypes
}

// fetchValidations loads public validations from the database, returning an empty
// slice on error so callers can continue rendering without blocking the page.
func fetchValidations(svc *service.Registry) []db.Validation {
	validationsPtr, err := svc.Driver().ListValidations()
	if err != nil {
		utility.DefaultLogger.Error("failed to list validations for dropdown", err)
		return []db.Validation{}
	}
	if validationsPtr == nil {
		return []db.Validation{}
	}
	return *validationsPtr
}

// FieldCreatePageHandler handles GET /admin/fields/new.
// Renders the full field creation page with validation and UI config editors.
func FieldCreatePageHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allRoles := make([]db.Roles, 0)
		rolesList, rolesErr := svc.Driver().ListRoles()
		if rolesErr == nil && rolesList != nil {
			allRoles = *rolesList
		}

		fieldTypes := fetchFieldTypes(svc, r)

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "New Field")
		RenderNav(w, r, "New Field",
			pages.FieldCreateContent("", "", "", "", "", "", allRoles, nil, csrfToken, fieldTypes),
			pages.FieldCreate(layout, allRoles, csrfToken, fieldTypes))
	}
}

// FieldCreateHandler handles POST /admin/fields.
// Creates a field via the service layer with full validation, data, and UI config.
func FieldCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		label := strings.TrimSpace(r.FormValue("label"))
		fieldType := strings.TrimSpace(r.FormValue("type"))
		data := strings.TrimSpace(r.FormValue("data"))
		validationIDStr := strings.TrimSpace(r.FormValue("validation_id"))
		uiConfig := strings.TrimSpace(r.FormValue("ui_config"))

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		var authorID types.NullableUserID
		if user != nil {
			authorID = types.NullableUserID{ID: user.UserID, Valid: true}
		}

		var rolesParam types.NullableString
		if selectedRoles := r.Form["roles"]; len(selectedRoles) > 0 {
			rolesJSON, marshalErr := json.Marshal(selectedRoles)
			if marshalErr == nil {
				rolesParam = types.NewNullableString(string(rolesJSON))
			}
		}

		created, err := svc.Schema.CreateField(r.Context(), ac, db.CreateFieldParams{
			Name:         name,
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         data,
			ValidationID: parseNullableValidationID(validationIDStr),
			UIConfig:     uiConfig,
			Roles:        rolesParam,
			AuthorID:     authorID,
		})
		if err != nil {
			fieldTypes := fetchFieldTypes(svc, r)
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				allRoles := make([]db.Roles, 0)
				rolesList, rolesErr := svc.Driver().ListRoles()
				if rolesErr == nil && rolesList != nil {
					allRoles = *rolesList
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, pages.FieldCreateForm(name, label, fieldType, data, validationIDStr, uiConfig, allRoles, errs, csrfToken, fieldTypes))
				return
			}
			utility.DefaultLogger.Error("failed to create field", err)
			allRoles := make([]db.Roles, 0)
			rolesList, rolesErr := svc.Driver().ListRoles()
			if rolesErr == nil && rolesList != nil {
				allRoles = *rolesList
			}
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, pages.FieldCreateForm(name, label, fieldType, data, validationIDStr, uiConfig, allRoles, map[string]string{"_": "Failed to create field"}, csrfToken, fieldTypes))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/fields/"+created.FieldID.String(), http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/fields/"+created.FieldID.String())
		w.WriteHeader(http.StatusOK)
	}
}

// FieldDetailHandler handles GET /admin/fields/{id}.
// Shows field detail with configuration, validation, and linked datatypes.
// When i18n is enabled, shows a "Translatable" checkbox on the edit form.
func FieldDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing field ID", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		isAdmin := middleware.ContextIsAdmin(r.Context())
		roleID := ""
		if user != nil {
			roleID = user.Role
		}

		field, err := svc.Schema.GetField(r.Context(), types.FieldID(id), roleID, isAdmin)
		if err != nil {
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Field not found", http.StatusNotFound)
				return
			}
			var fe *service.ForbiddenError
			if errors.As(err, &fe) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			utility.DefaultLogger.Error("failed to get field", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		i18nEnabled := false
		cfg, cfgErr := svc.Config()
		if cfgErr == nil {
			i18nEnabled = cfg.I18nEnabled()
		}

		// Fetch all roles for the roles multi-select dropdown.
		allRoles := make([]db.Roles, 0)
		rolesList, rolesErr := svc.Driver().ListRoles()
		if rolesErr == nil && rolesList != nil {
			allRoles = *rolesList
		}

		fieldTypes := fetchFieldTypes(svc, r)
		validations := fetchValidations(svc)

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Field: "+field.Label)
		RenderNav(w, r, "Field: "+field.Label,
			pages.FieldDetailContent(*field, allRoles, csrfToken, i18nEnabled, fieldTypes, validations),
			pages.FieldDetail(layout, *field, allRoles, csrfToken, i18nEnabled, fieldTypes, validations))
	}
}

// FieldUpdateHandler handles POST /admin/fields/{id}.
// Updates field properties via the service layer, which validates and preserves immutable fields.
func FieldUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing field ID", http.StatusBadRequest)
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
		validationIDStr := strings.TrimSpace(r.FormValue("validation_id"))
		uiConfig := strings.TrimSpace(r.FormValue("ui_config"))

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Parse translatable flag from form (checkbox: "1" = true, hidden "0" = false).
		// When i18n is disabled the form omits the field entirely, defaulting to false.
		var translatable bool
		if r.Form.Has("translatable") {
			translatable = r.FormValue("translatable") == "1"
		}

		// Parse roles multi-select: marshal selected role IDs to a JSON array.
		// An empty selection clears the roles restriction (NULL = unrestricted).
		var rolesParam types.NullableString
		if selectedRoles := r.Form["roles"]; len(selectedRoles) > 0 {
			rolesJSON, marshalErr := json.Marshal(selectedRoles)
			if marshalErr == nil {
				rolesParam = types.NewNullableString(string(rolesJSON))
			}
		}

		_, err := svc.Schema.UpdateField(r.Context(), ac, db.UpdateFieldParams{
			FieldID:      types.FieldID(id),
			Name:         name,
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         data,
			ValidationID: parseNullableValidationID(validationIDStr),
			UIConfig:     uiConfig,
			Translatable: translatable,
			Roles:        rolesParam,
		})
		if err != nil {
			fieldTypes := fetchFieldTypes(svc, r)
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.FieldEditForm(id, name, label, fieldType, data, validationIDStr, uiConfig, errs, csrfToken, fieldTypes))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Field not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update field", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldEditForm(id, name, label, fieldType, data, validationIDStr, uiConfig, map[string]string{"_": "Failed to update field"}, csrfToken, fieldTypes))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/fields/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/fields/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// FieldDeleteHandler handles DELETE /admin/fields/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func FieldDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing field ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete field", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Schema.DeleteField(r.Context(), ac, types.FieldID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete field", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete field", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// parseNullableValidationID converts a form string to a NullableValidationID.
func parseNullableValidationID(s string) types.NullableValidationID {
	if s == "" {
		return types.NullableValidationID{}
	}
	return types.NullableValidationID{ID: types.ValidationID(s), Valid: true}
}
