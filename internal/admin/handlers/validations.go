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

// ---- Public Validations ----

// ValidationsListHandler handles GET /admin/validations.
// Lists all public validations with a create dialog.
func ValidationsListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.Validations.ListValidations()
		if err != nil {
			utility.DefaultLogger.Error("failed to list validations", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var validations []db.Validation
		if list != nil {
			validations = *list
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Validations")

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Validations"}`)
			RenderWithOOB(w, r, pages.ValidationsListContent(validations),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.ValidationCreateDialog(csrfToken)})
			return
		}

		Render(w, r, pages.ValidationsList(layout, validations))
	}
}

// ValidationDetailHandler handles GET /admin/validations/{id}.
// Shows validation detail with an edit form.
func ValidationDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing validation ID", http.StatusBadRequest)
			return
		}

		v, err := svc.Validations.GetValidation(types.ValidationID(id))
		if err != nil {
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Validation not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to get validation", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Validation: "+v.Name)

		RenderNav(w, r, "Validation: "+v.Name,
			pages.ValidationDetailContent(*v, csrfToken),
			pages.ValidationDetail(layout, *v, csrfToken))
	}
}

// ValidationCreateHandler handles POST /admin/validations.
// Creates a validation via the service layer.
func ValidationCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		description := strings.TrimSpace(r.FormValue("description"))
		config := strings.TrimSpace(r.FormValue("config"))
		if config == "" {
			config = types.EmptyJSON
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())

		_, err := svc.Validations.CreateValidation(r.Context(), ac, db.CreateValidationParams{
			Name:        name,
			Description: description,
			Config:      config,
			AuthorID:    types.NullableUserID{ID: user.UserID, Valid: true},
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
				Render(w, r, partials.ValidationForm(name, description, config, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create validation", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.ValidationForm(name, description, config, map[string]string{"_": "failed to create validation"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/validations", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Validation created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/validations")
		w.WriteHeader(http.StatusOK)
	}
}

// ValidationUpdateHandler handles POST /admin/validations/{id}.
// Updates a validation via the service layer.
func ValidationUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing validation ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		description := strings.TrimSpace(r.FormValue("description"))
		config := strings.TrimSpace(r.FormValue("config"))
		if config == "" {
			config = types.EmptyJSON
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())

		_, err := svc.Validations.UpdateValidation(r.Context(), ac, db.UpdateValidationParams{
			ValidationID: types.ValidationID(id),
			Name:         name,
			Description:  description,
			Config:       config,
			AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
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
				Render(w, r, partials.ValidationEditForm(id, name, description, config, errs, csrfToken))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Validation not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update validation", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.ValidationEditForm(id, name, description, config, map[string]string{"_": "failed to update validation"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/validations/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Validation updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/validations/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// ValidationDeleteHandler handles DELETE /admin/validations/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func ValidationDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing validation ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete validation", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Validations.DeleteValidation(r.Context(), ac, types.ValidationID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete validation", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete validation", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Validation deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// ---- Admin Validations ----

// AdminValidationsListHandler handles GET /admin/admin-validations.
// Lists all admin validations with a create dialog.
func AdminValidationsListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.Validations.ListAdminValidations()
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin validations", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var validations []db.AdminValidation
		if list != nil {
			validations = *list
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Admin Validations")

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Admin Validations"}`)
			RenderWithOOB(w, r, pages.AdminValidationsListContent(validations),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.AdminValidationCreateDialog(csrfToken)})
			return
		}

		Render(w, r, pages.AdminValidationsList(layout, validations))
	}
}

// AdminValidationDetailHandler handles GET /admin/admin-validations/{id}.
// Shows admin validation detail with an edit form.
func AdminValidationDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing admin validation ID", http.StatusBadRequest)
			return
		}

		v, err := svc.Validations.GetAdminValidation(types.AdminValidationID(id))
		if err != nil {
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin validation not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to get admin validation", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Admin Validation: "+v.Name)

		RenderNav(w, r, "Admin Validation: "+v.Name,
			pages.AdminValidationDetailContent(*v, csrfToken),
			pages.AdminValidationDetail(layout, *v, csrfToken))
	}
}

// AdminValidationCreateHandler handles POST /admin/admin-validations.
// Creates an admin validation via the service layer.
func AdminValidationCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		description := strings.TrimSpace(r.FormValue("description"))
		config := strings.TrimSpace(r.FormValue("config"))
		if config == "" {
			config = types.EmptyJSON
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())

		_, err := svc.Validations.CreateAdminValidation(r.Context(), ac, db.CreateAdminValidationParams{
			Name:        name,
			Description: description,
			Config:      config,
			AuthorID:    types.NullableUserID{ID: user.UserID, Valid: true},
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
				Render(w, r, partials.AdminValidationForm(name, description, config, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create admin validation", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminValidationForm(name, description, config, map[string]string{"_": "failed to create admin validation"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-validations", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin validation created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-validations")
		w.WriteHeader(http.StatusOK)
	}
}

// AdminValidationUpdateHandler handles POST /admin/admin-validations/{id}.
// Updates an admin validation via the service layer.
func AdminValidationUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing admin validation ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		description := strings.TrimSpace(r.FormValue("description"))
		config := strings.TrimSpace(r.FormValue("config"))
		if config == "" {
			config = types.EmptyJSON
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())

		_, err := svc.Validations.UpdateAdminValidation(r.Context(), ac, db.UpdateAdminValidationParams{
			AdminValidationID: types.AdminValidationID(id),
			Name:              name,
			Description:       description,
			Config:            config,
			AuthorID:          types.NullableUserID{ID: user.UserID, Valid: true},
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
				Render(w, r, partials.AdminValidationEditForm(id, name, description, config, errs, csrfToken))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin validation not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update admin validation", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminValidationEditForm(id, name, description, config, map[string]string{"_": "failed to update admin validation"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-validations/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin validation updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-validations/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminValidationDeleteHandler handles DELETE /admin/admin-validations/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func AdminValidationDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing admin validation ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete admin validation", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Validations.DeleteAdminValidation(r.Context(), ac, types.AdminValidationID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete admin validation", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete admin validation", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin validation deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
