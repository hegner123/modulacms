package handlers

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
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

		pd := NewPaginationData(*total, limit, offset, "#fields-table-body", "/admin/schema/fields")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
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
func FieldDetailHandler(driver db.DbDriver) http.HandlerFunc {
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

		// Fetch linked datatypes via junction table
		links, err := driver.ListDatatypeFieldByFieldID(types.FieldID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to list field datatypes", err)
			links = &[]db.DatatypeFields{}
		}

		linkedDatatypes := make([]db.Datatypes, 0)
		if links != nil {
			for _, link := range *links {
				dt, dtErr := driver.GetDatatype(link.DatatypeID)
				if dtErr != nil {
					continue
				}
				linkedDatatypes = append(linkedDatatypes, *dt)
			}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Field: "+field.Label)
		Render(w, r, pages.FieldDetail(layout, *field, linkedDatatypes, csrfToken))
	}
}

// FieldCreateHandler handles POST /admin/schema/fields.
// Validates label and type are required, creates via audited context.
func FieldCreateHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

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
		} else if validateErr := types.FieldType(fieldType).Validate(); validateErr != nil {
			errs["type"] = "Invalid field type"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldForm(label, fieldType, data, validation, uiConfig, errs, csrfToken))
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		now := types.NewTimestamp(time.Now())
		_, err := driver.CreateField(r.Context(), ac, db.CreateFieldParams{
			FieldID:      types.NewFieldID(),
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         data,
			Validation:   validation,
			UIConfig:     uiConfig,
			AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create field", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldForm(label, fieldType, data, validation, uiConfig, map[string]string{"_": "Failed to create field"}, csrfToken))
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
func FieldUpdateHandler(driver db.DbDriver) http.HandlerFunc {
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
		} else if validateErr := types.FieldType(fieldType).Validate(); validateErr != nil {
			errs["type"] = "Invalid field type"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldEditForm(id, label, fieldType, data, validation, uiConfig, errs, csrfToken))
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
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		_, err = driver.UpdateField(r.Context(), ac, db.UpdateFieldParams{
			FieldID:      types.FieldID(id),
			ParentID:     existing.ParentID,
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         data,
			Validation:   validation,
			UIConfig:     uiConfig,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.NewTimestamp(time.Now()),
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to update field", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.FieldEditForm(id, label, fieldType, data, validation, uiConfig, map[string]string{"_": "Failed to update field"}, csrfToken))
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
func FieldDeleteHandler(driver db.DbDriver) http.HandlerFunc {
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

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

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
