package handlers

import (
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

// AdminFieldTypesListHandler handles GET /admin/admin-field-types.
// Lists all admin field types with a create dialog.
func AdminFieldTypesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.Schema.ListAdminFieldTypes(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin field types", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Admin Field Types")

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Admin Field Types"}`)
			RenderWithOOB(w, r, pages.AdminFieldTypesListContent(list),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.AdminFieldTypeCreateDialog(csrfToken)})
			return
		}

		Render(w, r, pages.AdminFieldTypesList(layout, list))
	}
}

// AdminFieldTypeDetailHandler handles GET /admin/admin-field-types/{id}.
// Shows admin field type detail with an edit form.
func AdminFieldTypeDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing admin field type ID", http.StatusBadRequest)
			return
		}

		ft, err := svc.Schema.GetAdminFieldType(r.Context(), types.AdminFieldTypeID(id))
		if err != nil {
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin field type not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to get admin field type", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Admin Field Type: "+ft.Label)

		RenderNav(w, r, "Admin Field Type: "+ft.Label,
			pages.AdminFieldTypeDetailContent(*ft, csrfToken),
			pages.AdminFieldTypeDetail(layout, *ft, csrfToken))
	}
}

// AdminFieldTypeCreateHandler handles POST /admin/admin-field-types.
// Creates an admin field type via the service layer.
func AdminFieldTypeCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		ftType := strings.TrimSpace(r.FormValue("type"))
		label := strings.TrimSpace(r.FormValue("label"))

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err := svc.Schema.CreateAdminFieldType(r.Context(), ac, db.CreateAdminFieldTypeParams{
			Type:  ftType,
			Label: label,
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
				Render(w, r, partials.AdminFieldTypeForm(ftType, label, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create admin field type", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminFieldTypeForm(ftType, label, map[string]string{"_": "Failed to create admin field type"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-field-types", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin field type created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-field-types")
		w.WriteHeader(http.StatusOK)
	}
}

// AdminFieldTypeUpdateHandler handles POST /admin/admin-field-types/{id}.
// Updates an admin field type via the service layer.
func AdminFieldTypeUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing admin field type ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		ftType := strings.TrimSpace(r.FormValue("type"))
		label := strings.TrimSpace(r.FormValue("label"))

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, err := svc.Schema.UpdateAdminFieldType(r.Context(), ac, db.UpdateAdminFieldTypeParams{
			AdminFieldTypeID: types.AdminFieldTypeID(id),
			Type:             ftType,
			Label:            label,
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
				Render(w, r, partials.AdminFieldTypeEditForm(id, ftType, label, errs, csrfToken))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Admin field type not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update admin field type", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminFieldTypeEditForm(id, ftType, label, map[string]string{"_": "Failed to update admin field type"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-field-types/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin field type updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-field-types/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminFieldTypeDeleteHandler handles DELETE /admin/admin-field-types/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func AdminFieldTypeDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing admin field type ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin field type", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Schema.DeleteAdminFieldType(r.Context(), ac, types.AdminFieldTypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete admin field type", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin field type", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin field type deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
