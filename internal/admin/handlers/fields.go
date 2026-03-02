package handlers

import (
	"encoding/json"
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
	"github.com/hegner123/modulacms/internal/utility"
)

// FieldsListHandler handles GET /admin/schema/fields.
// Lists all fields with pagination. HTMX requests receive partial table rows only.
func FieldsListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, err := driver.ListFieldsPaginated(db.PaginationParams{Limit: limit, Offset: offset})
		if err != nil {
			utility.DefaultLogger.Error("failed to list fields", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		total, err := driver.CountFields()
		if err != nil {
			utility.DefaultLogger.Error("failed to count fields", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.Fields, 0)
		if items != nil {
			list = *items
		}

		// Filter fields by the authenticated user's role.
		user := middleware.AuthenticatedUser(r.Context())
		isAdmin := middleware.ContextIsAdmin(r.Context())
		roleID := ""
		if user != nil {
			roleID = user.Role
		}
		list = db.FilterFieldsByRole(list, roleID, isAdmin)

		pd := NewPaginationData(*total, limit, offset, "#fields-table-body", "/admin/schema/fields")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Fields"}`)
			RenderWithOOB(w, r, pages.FieldsListContent(list, pg),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.FieldCreateDialog(csrfToken)})
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

// FieldDetailHandler handles GET /admin/schema/fields/{id}.
// Shows field detail with configuration, validation, and linked datatypes.
// When i18n is enabled, shows a "Translatable" checkbox on the edit form.
func FieldDetailHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing field ID", http.StatusBadRequest)
			return
		}

		field, err := driver.GetField(types.FieldID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get field", err)
			http.Error(w, "Field not found", http.StatusNotFound)
			return
		}

		// Check field-level role access.
		user := middleware.AuthenticatedUser(r.Context())
		isAdmin := middleware.ContextIsAdmin(r.Context())
		roleID := ""
		if user != nil {
			roleID = user.Role
		}
		if !db.IsFieldAccessible(*field, roleID, isAdmin) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		i18nEnabled := false
		cfg, cfgErr := mgr.Config()
		if cfgErr == nil {
			i18nEnabled = cfg.I18nEnabled()
		}

		// Fetch all roles for the roles multi-select dropdown.
		allRoles := make([]db.Roles, 0)
		rolesList, rolesErr := driver.ListRoles()
		if rolesErr == nil && rolesList != nil {
			allRoles = *rolesList
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Field: "+field.Label)
		RenderNav(w, r, "Field: "+field.Label,
			pages.FieldDetailContent(*field, allRoles, csrfToken, i18nEnabled),
			pages.FieldDetail(layout, *field, allRoles, csrfToken, i18nEnabled))
	}
}

// FieldCreateHandler handles POST /admin/schema/fields.
// Validates label and type are required, creates via audited context.
func FieldCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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
		validation := strings.TrimSpace(r.FormValue("validation"))
		uiConfig := strings.TrimSpace(r.FormValue("ui_config"))

		// Default empty JSON objects
		if data == "" {
			data = "{}"
		}
		if validation == "" {
			validation = "{}"
		}
		if uiConfig == "" {
			uiConfig = "{}"
		}

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if fieldType == "" {
			errs["type"] = "Type is required"
		} else if _, lookupErr := driver.GetFieldTypeByType(fieldType); lookupErr != nil {
			errs["type"] = "Invalid field type"
		}

		// Validate the validation config JSON before persisting.
		vc, vcErr := types.ParseValidationConfig(validation)
		if vcErr != nil {
			errs["validation"] = vcErr.Error()
		} else if vcValErr := types.ValidateValidationConfig(vc); vcValErr != nil {
			errs["validation"] = vcValErr.Error()
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldForm(name, label, fieldType, data, validation, uiConfig, errs, csrfToken))
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		// Parse roles multi-select: marshal selected role IDs to a JSON array.
		var rolesParam types.NullableString
		if selectedRoles := r.Form["roles"]; len(selectedRoles) > 0 {
			rolesJSON, marshalErr := json.Marshal(selectedRoles)
			if marshalErr == nil {
				rolesParam = types.NewNullableString(string(rolesJSON))
			}
		}

		now := types.NewTimestamp(time.Now())
		_, err := driver.CreateField(r.Context(), ac, db.CreateFieldParams{
			FieldID:      types.NewFieldID(),
			Name:         name,
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         data,
			Validation:   validation,
			UIConfig:     uiConfig,
			Roles:        rolesParam,
			AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create field", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldForm(name, label, fieldType, data, validation, uiConfig, map[string]string{"_": "Failed to create field"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/schema/fields", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/schema/fields")
		w.WriteHeader(http.StatusOK)
	}
}

// FieldUpdateHandler handles POST /admin/schema/fields/{id}.
// Updates field properties via audited context.
func FieldUpdateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

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
		validation := strings.TrimSpace(r.FormValue("validation"))
		uiConfig := strings.TrimSpace(r.FormValue("ui_config"))

		if data == "" {
			data = "{}"
		}
		if validation == "" {
			validation = "{}"
		}
		if uiConfig == "" {
			uiConfig = "{}"
		}

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if fieldType == "" {
			errs["type"] = "Type is required"
		} else if _, lookupErr := driver.GetFieldTypeByType(fieldType); lookupErr != nil {
			errs["type"] = "Invalid field type"
		}

		// Validate the validation config JSON before persisting.
		vc, vcErr := types.ParseValidationConfig(validation)
		if vcErr != nil {
			errs["validation"] = vcErr.Error()
		} else if vcValErr := types.ValidateValidationConfig(vc); vcValErr != nil {
			errs["validation"] = vcValErr.Error()
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldEditForm(id, name, label, fieldType, data, validation, uiConfig, errs, csrfToken))
			return
		}

		existing, err := driver.GetField(types.FieldID(id))
		if err != nil {
			http.Error(w, "Field not found", http.StatusNotFound)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		// Parse translatable flag from form (checkbox: "1" = true, hidden "0" = false)
		translatable := existing.Translatable
		if r.Form.Has("translatable") {
			if r.FormValue("translatable") == "1" {
				translatable = 1
			} else {
				translatable = 0
			}
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

		_, err = driver.UpdateField(r.Context(), ac, db.UpdateFieldParams{
			FieldID:      types.FieldID(id),
			ParentID:     existing.ParentID,
			Name:         name,
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         data,
			Validation:   validation,
			UIConfig:     uiConfig,
			Translatable: translatable,
			Roles:        rolesParam,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.NewTimestamp(time.Now()),
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to update field", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldEditForm(id, name, label, fieldType, data, validation, uiConfig, map[string]string{"_": "Failed to update field"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/schema/fields/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/schema/fields/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// FieldDeleteHandler handles DELETE /admin/schema/fields/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func FieldDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing field ID", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		err := driver.DeleteField(r.Context(), ac, types.FieldID(id))
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
