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

// FieldTypesListHandler handles GET /admin/field-types.
// Lists all field types with search and sort. HTMX partial requests
// receive table rows only; full/nav requests get the complete page.
func FieldTypesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		sortBy := r.URL.Query().Get("sort")

		list, err := svc.Schema.ListFieldTypes(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list field types", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Fuzzy search (overrides sort with relevance order)
		if search != "" {
			results := utility.FuzzyFind(search, list, func(ft db.FieldTypes) []string {
				return []string{ft.Label, ft.Type}
			})
			ranked := make([]db.FieldTypes, len(results))
			for i, r := range results {
				ranked[i] = list[r.Index]
			}
			list = ranked
			sortBy = ""
		}

		// Sort
		switch sortBy {
		case "label-asc":
			sortFieldTypes(list, func(a, b db.FieldTypes) bool { return strings.ToLower(a.Label) < strings.ToLower(b.Label) })
		case "label-desc":
			sortFieldTypes(list, func(a, b db.FieldTypes) bool { return strings.ToLower(a.Label) > strings.ToLower(b.Label) })
		case "type-asc":
			sortFieldTypes(list, func(a, b db.FieldTypes) bool { return strings.ToLower(a.Type) < strings.ToLower(b.Type) })
		case "type-desc":
			sortFieldTypes(list, func(a, b db.FieldTypes) bool { return strings.ToLower(a.Type) > strings.ToLower(b.Type) })
		}

		// HTMX partial (search/sort toolbar requests)
		if IsHTMX(r) && !IsNavHTMX(r) {
			Render(w, r, partials.FieldTypesTableRows(list))
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Field Types"}`)
			RenderWithOOB(w, r, pages.FieldTypesListContent(list),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.FieldTypeCreateDialog(csrfToken)})
			return
		}

		layout := NewAdminData(r, "Field Types")
		Render(w, r, pages.FieldTypesList(layout, list))
	}
}

func sortFieldTypes(s []db.FieldTypes, less func(a, b db.FieldTypes) bool) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && less(s[j], s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// FieldTypeDetailHandler handles GET /admin/field-types/{id}.
// Shows field type detail with an edit form.
func FieldTypeDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing field type ID", http.StatusBadRequest)
			return
		}

		ft, err := svc.Schema.GetFieldType(r.Context(), types.FieldTypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get field type", err)
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Field type not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Field Type: "+ft.Label)

		if IsNavHTMX(r) {
			safeTitle := "Field Type: " + ft.Label
			w.Header().Set("HX-Trigger", `{"pageTitle": "`+safeTitle+`"}`)
			Render(w, r, pages.FieldTypeDetailContent(*ft, csrfToken))
			return
		}
		Render(w, r, pages.FieldTypeDetail(layout, *ft, csrfToken))
	}
}

// FieldTypeCreateHandler handles POST /admin/field-types.
// Creates a field type via the service layer.
func FieldTypeCreateHandler(svc *service.Registry) http.HandlerFunc {
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

		_, err := svc.Schema.CreateFieldType(r.Context(), ac, db.CreateFieldTypeParams{
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
				Render(w, r, partials.FieldTypeForm(ftType, label, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create field type", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldTypeForm(ftType, label, map[string]string{"_": "failed to create field type"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/field-types", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field type created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/field-types")
		w.WriteHeader(http.StatusOK)
	}
}

// FieldTypeUpdateHandler handles POST /admin/field-types/{id}.
// Updates user-editable fields via the service layer.
func FieldTypeUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing field type ID", http.StatusBadRequest)
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

		_, err := svc.Schema.UpdateFieldType(r.Context(), ac, db.UpdateFieldTypeParams{
			FieldTypeID: types.FieldTypeID(id),
			Type:        ftType,
			Label:       label,
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
				Render(w, r, partials.FieldTypeEditForm(id, ftType, label, errs, csrfToken))
				return
			}
			var nfe *service.NotFoundError
			if errors.As(err, &nfe) {
				http.Error(w, "Field type not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update field type", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldTypeEditForm(id, ftType, label, map[string]string{"_": "failed to update field type"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/field-types/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field type updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/field-types/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// FieldTypeDeleteHandler handles DELETE /admin/field-types/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func FieldTypeDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing field type ID", http.StatusBadRequest)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete field type", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := svc.Schema.DeleteFieldType(r.Context(), ac, types.FieldTypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete field type", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete field type", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field type deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
