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
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// fetchAdminFieldTypes loads admin field types from the database, returning an empty
// slice on error so callers can continue rendering without blocking the page.
func fetchAdminFieldTypes(svc *service.Registry, r *http.Request) []db.AdminFieldTypes {
	fieldTypes, err := svc.Schema.ListAdminFieldTypes(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("failed to list admin field types for dropdown", err)
		return []db.AdminFieldTypes{}
	}
	return fieldTypes
}

func fetchAdminValidations(svc *service.Registry) []db.AdminValidation {
	validationsPtr, err := svc.Driver().ListAdminValidations()
	if err != nil {
		utility.DefaultLogger.Error("failed to list admin validations for dropdown", err)
		return []db.AdminValidation{}
	}
	if validationsPtr == nil {
		return []db.AdminValidation{}
	}
	return *validationsPtr
}

// AdminFieldDetailHandler handles GET /admin/admin-fields/{id}.
// Shows admin field detail with configuration, validation, and linked datatypes.
// When i18n is enabled, shows a "Translatable" checkbox on the edit form.
func AdminFieldDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing field ID", http.StatusBadRequest)
			return
		}

		field, err := svc.Schema.GetAdminField(r.Context(), types.AdminFieldID(id))
		if err != nil {
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin field not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to get admin field", err)
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

		fieldTypes := fetchAdminFieldTypes(svc, r)
		validations := fetchAdminValidations(svc)

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Admin Field: "+field.Label)
		RenderNav(w, r, "Admin Field: "+field.Label,
			pages.AdminFieldDetailContent(*field, allRoles, csrfToken, i18nEnabled, fieldTypes, validations),
			pages.AdminFieldDetail(layout, *field, allRoles, csrfToken, i18nEnabled, fieldTypes, validations))
	}
}

// AdminFieldUpdateHandler handles POST /admin/admin-fields/{id}.
// Updates admin field properties via the service layer, which validates and preserves immutable fields.
func AdminFieldUpdateHandler(svc *service.Registry) http.HandlerFunc {
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
		validation := strings.TrimSpace(r.FormValue("validation")) // kept for form re-render
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

		_, err := svc.Schema.UpdateAdminField(r.Context(), ac, db.UpdateAdminFieldParams{
			AdminFieldID: types.AdminFieldID(id),
			Name:         name,
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         data,
			ValidationID: types.NullableAdminValidationID{}, // TODO: parse validation_id from form
			UIConfig:     uiConfig,
			Translatable: translatable,
			Roles:        rolesParam,
		})
		if err != nil {
			fieldTypes := fetchAdminFieldTypes(svc, r)
			var ve *service.ValidationError
			if errors.As(err, &ve) {
				errs := make(map[string]string, len(ve.Errors))
				for _, fe := range ve.Errors {
					errs[fe.Field] = fe.Message
				}
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.AdminFieldEditForm(id, name, label, fieldType, data, validation, uiConfig, errs, csrfToken, fieldTypes))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin field not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update admin field", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminFieldEditForm(id, name, label, fieldType, data, validation, uiConfig, map[string]string{"_": "Failed to update field"}, csrfToken, fieldTypes))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-fields/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin field updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-fields/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminFieldDeleteHandler handles DELETE /admin/admin-fields/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func AdminFieldDeleteHandler(svc *service.Registry) http.HandlerFunc {
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
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin field", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Schema.DeleteAdminField(r.Context(), ac, types.AdminFieldID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete admin field", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin field", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin field deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
