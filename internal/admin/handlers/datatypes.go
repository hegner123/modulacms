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

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Datatypes"}`)
			RenderWithOOB(w, r, pages.DatatypesListContent(list, pg),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.DatatypeCreateDialog(csrfToken)})
			return
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

		// Fetch linked fields by parent_id (fields belong to datatype directly)
		fieldList, err := driver.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: types.DatatypeID(id), Valid: true})
		if err != nil {
			utility.DefaultLogger.Error("failed to list datatype fields", err)
			fieldList = &[]db.Fields{}
		}
		linkedFields := make([]db.Fields, 0)
		if fieldList != nil {
			linkedFields = *fieldList
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Datatype: "+dt.Label)

		if IsNavHTMX(r) {
			safeTitle := "Datatype: " + dt.Label
			w.Header().Set("HX-Trigger", `{"pageTitle": "`+safeTitle+`"}`)
			RenderWithOOB(w, r, pages.DatatypeDetailContent(*dt, linkedFields, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.DatatypeAddFieldDialog(dt.DatatypeID.String(), csrfToken)})
			return
		}
		Render(w, r, pages.DatatypeDetail(layout, *dt, linkedFields, csrfToken))
	}
}

// DatatypeCreateHandler handles POST /admin/schema/datatypes.
// Validates label and type are required, creates via audited context.
func DatatypeCreateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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
		dtype := strings.TrimSpace(r.FormValue("type"))

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if dtype == "" {
			errs["type"] = "Type is required"
		} else if types.IsReservedPrefix(dtype) {
			errs["type"] = "Type names starting with '_' are reserved for system use"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeForm(name, label, dtype, errs, csrfToken))
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		now := types.NewTimestamp(time.Now())
		_, err := driver.CreateDatatype(r.Context(), ac, db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			Name:         name,
			Label:        label,
			Type:         dtype,
			AuthorID:     user.UserID,
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeForm(name, label, dtype, map[string]string{"_": "Failed to create datatype"}, csrfToken))
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
func DatatypeUpdateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

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

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if dtype == "" {
			errs["type"] = "Type is required"
		} else if types.IsReservedPrefix(dtype) {
			errs["type"] = "Type names starting with '_' are reserved for system use"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeEditForm(id, name, label, dtype, errs, csrfToken))
			return
		}

		existing, err := driver.GetDatatype(types.DatatypeID(id))
		if err != nil {
			http.Error(w, "Datatype not found", http.StatusNotFound)
			return
		}

		// Block type changes on system datatypes (label changes are allowed)
		if types.IsReservedPrefix(existing.Type) && dtype != existing.Type {
			w.WriteHeader(http.StatusForbidden)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeEditForm(id, name, label, dtype, map[string]string{"type": "Cannot change type of system datatype"}, csrfToken))
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		_, err = driver.UpdateDatatype(r.Context(), ac, db.UpdateDatatypeParams{
			DatatypeID:   types.DatatypeID(id),
			ParentID:     existing.ParentID,
			Name:         name,
			Label:        label,
			Type:         dtype,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.NewTimestamp(time.Now()),
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to update datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeEditForm(id, name, label, dtype, map[string]string{"_": "Failed to update datatype"}, csrfToken))
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
func DatatypeDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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
			http.Error(w, "Missing datatype ID", http.StatusBadRequest)
			return
		}

		// Block deletion of system datatypes (e.g. _root, _reference)
		existing, lookupErr := driver.GetDatatype(types.DatatypeID(id))
		if lookupErr != nil {
			utility.DefaultLogger.Error("failed to look up datatype for delete", lookupErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Datatype not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if types.IsReservedPrefix(existing.Type) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Cannot delete system datatype", "type": "error"}}`)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

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

// DatatypeCreateFieldHandler handles POST /admin/schema/datatypes/{id}/fields.
// Creates a new field with parent_id set to this datatype.
func DatatypeCreateFieldHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

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

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if fieldType == "" {
			errs["type"] = "Type is required"
		} else if _, lookupErr := driver.GetFieldTypeByType(fieldType); lookupErr != nil {
			errs["type"] = "Invalid field type"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeCreateFieldForm(id, name, label, fieldType, "", "", "", errs, csrfToken))
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		clientIP, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			clientIP = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), clientIP)

		now := types.NewTimestamp(time.Now())
		_, err := driver.CreateField(r.Context(), ac, db.CreateFieldParams{
			FieldID:      types.NewFieldID(),
			ParentID:     types.NullableDatatypeID{ID: types.DatatypeID(id), Valid: true},
			SortOrder:    0,
			Name:         name,
			Label:        label,
			Type:         types.FieldType(fieldType),
			Data:         "{}",
			Validation:   "{}",
			UIConfig:     "{}",
			AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create field for datatype", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.DatatypeCreateFieldForm(id, name, label, fieldType, "", "", "", map[string]string{"_": "Failed to create field"}, csrfToken))
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
func DatatypesJSONHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := driver.ListDatatypes()
		if err != nil {
			utility.DefaultLogger.Error("failed to list datatypes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.Datatypes, 0)
		if items != nil {
			list = *items
		}

		w.Header().Set("Content-Type", "application/json")
		if encErr := json.NewEncoder(w).Encode(list); encErr != nil {
			utility.DefaultLogger.Error("failed to encode datatypes JSON", encErr)
		}
	}
}
