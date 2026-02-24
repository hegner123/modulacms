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

// DatatypesListHandler handles GET /admin/schema/datatypes.
// Lists all datatypes with pagination. HTMX requests receive partial table rows only.
func DatatypesListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, err := driver.ListDatatypesPaginated(db.PaginationParams{Limit: limit, Offset: offset})
		if err != nil {
			utility.DefaultLogger.Error("failed to list datatypes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		total, err := driver.CountDatatypes()
		if err != nil {
			utility.DefaultLogger.Error("failed to count datatypes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.Datatypes, 0)
		if items != nil {
			list = *items
		}

		pd := NewPaginationData(*total, limit, offset, "#datatypes-table-body", "/admin/schema/datatypes")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsHTMX(r) {
			Render(w, r, partials.DatatypesTableRows(list, pg))
			return
		}

		layout := NewAdminData(r, "Datatypes")
		Render(w, r, pages.DatatypesList(layout, list, pg))
	}
}

// DatatypeDetailHandler handles GET /admin/schema/datatypes/{id}.
// Shows datatype detail with linked fields list.
func DatatypeDetailHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing datatype ID", http.StatusBadRequest)
			return
		}

		dt, err := driver.GetDatatype(types.DatatypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get datatype", err)
			http.Error(w, "Datatype not found", http.StatusNotFound)
			return
		}

		// Fetch linked fields via datatype-field junction
		links, err := driver.ListDatatypeFieldByDatatypeID(types.DatatypeID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to list datatype fields", err)
			links = &[]db.DatatypeFields{}
		}

		// Resolve field details for each link
		linkedFields := make([]db.Fields, 0)
		if links != nil {
			for _, link := range *links {
				f, fErr := driver.GetField(link.FieldID)
				if fErr != nil {
					continue
				}
				linkedFields = append(linkedFields, *f)
			}
		}

		// Fetch all fields for the "add field" selector
		allFields, err := driver.ListFieldsPaginated(db.PaginationParams{Limit: 1000, Offset: 0})
		if err != nil {
			utility.DefaultLogger.Error("failed to list all fields", err)
			allFields = &[]db.Fields{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Datatype: "+dt.Label)
		Render(w, r, pages.DatatypeDetail(layout, *dt, linkedFields, *allFields, csrfToken))
	}
}

// DatatypeCreateHandler handles POST /admin/schema/datatypes.
// Validates label and type are required, creates via audited context.
func DatatypeCreateHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		label := strings.TrimSpace(r.FormValue("label"))
		dtype := strings.TrimSpace(r.FormValue("type"))

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if dtype == "" {
			errs["type"] = "Type is required"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeForm(label, dtype, errs, csrfToken))
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		now := types.NewTimestamp(time.Now())
		_, err := driver.CreateDatatype(r.Context(), ac, db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			Label:        label,
			Type:         dtype,
			AuthorID:     user.UserID,
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create datatype", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeForm(label, dtype, map[string]string{"_": "Failed to create datatype"}, csrfToken))
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
// Updates label and type via audited context.
func DatatypeUpdateHandler(driver db.DbDriver) http.HandlerFunc {
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

		label := strings.TrimSpace(r.FormValue("label"))
		dtype := strings.TrimSpace(r.FormValue("type"))

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if dtype == "" {
			errs["type"] = "Type is required"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeEditForm(id, label, dtype, errs, csrfToken))
			return
		}

		existing, err := driver.GetDatatype(types.DatatypeID(id))
		if err != nil {
			http.Error(w, "Datatype not found", http.StatusNotFound)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		_, err = driver.UpdateDatatype(r.Context(), ac, db.UpdateDatatypeParams{
			DatatypeID:   types.DatatypeID(id),
			ParentID:     existing.ParentID,
			Label:        label,
			Type:         dtype,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.NewTimestamp(time.Now()),
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to update datatype", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeEditForm(id, label, dtype, map[string]string{"_": "Failed to update datatype"}, csrfToken))
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
func DatatypeDeleteHandler(driver db.DbDriver) http.HandlerFunc {
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

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		err := driver.DeleteDatatype(r.Context(), ac, types.DatatypeID(id))
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

// DatatypeLinkFieldHandler handles POST /admin/schema/datatypes/{id}/fields.
// Links a field to a datatype via the junction table.
func DatatypeLinkFieldHandler(driver db.DbDriver) http.HandlerFunc {
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

		fieldID := strings.TrimSpace(r.FormValue("field_id"))
		if fieldID == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field is required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		_, err := driver.CreateDatatypeField(r.Context(), ac, db.CreateDatatypeFieldParams{
			ID:         string(types.NewDatatypeFieldID()),
			DatatypeID: types.DatatypeID(id),
			FieldID:    types.FieldID(fieldID),
			SortOrder:  0,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to link field to datatype", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to link field", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/schema/datatypes/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field linked", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/schema/datatypes/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// DatatypeUnlinkFieldHandler handles DELETE /admin/schema/datatypes/{id}/fields/{linkId}.
// Removes a field-to-datatype link from the junction table.
func DatatypeUnlinkFieldHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		dtID := r.PathValue("id")
		linkID := r.PathValue("linkId")
		if dtID == "" || linkID == "" {
			http.Error(w, "Missing IDs", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		err := driver.DeleteDatatypeField(r.Context(), ac, linkID)
		if err != nil {
			utility.DefaultLogger.Error("failed to unlink field from datatype", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to unlink field", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Field unlinked", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
