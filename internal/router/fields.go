package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// FieldsHandler handles CRUD operations that do not require a specific field ID.
func FieldsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListFieldsPaginated(w, r, svc)
		} else {
			apiListFields(w, r, svc)
		}
	case http.MethodPost:
		apiCreateField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// FieldHandler handles CRUD operations for specific field items.
func FieldHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetField(w, r, svc)
	case http.MethodPut:
		apiUpdateField(w, r, svc)
	case http.MethodDelete:
		apiDeleteField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListFields handles GET requests for listing fields.
// Filters results by the authenticated user's role.
func apiListFields(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	user := middleware.AuthenticatedUser(r.Context())
	isAdmin := middleware.ContextIsAdmin(r.Context())
	roleID := ""
	if user != nil {
		roleID = user.Role
	}

	filtered, err := svc.Schema.ListFieldsFiltered(r.Context(), roleID, isAdmin)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(filtered)
	return nil
}

// apiListFieldsPaginated handles GET requests for listing fields with pagination.
// Filters results by the authenticated user's role. Total count reflects
// the pre-filter database count; Data contains only accessible fields.
func apiListFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	user := middleware.AuthenticatedUser(r.Context())
	isAdmin := middleware.ContextIsAdmin(r.Context())
	roleID := ""
	if user != nil {
		roleID = user.Role
	}

	response, err := svc.Schema.ListFieldsPaginated(r.Context(), params, roleID, isAdmin)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiGetField handles GET requests for a single field
func apiGetField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	fID := types.FieldID(q)
	if err := fID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Extract role info from middleware context for field-level access check.
	user := middleware.AuthenticatedUser(r.Context())
	isAdmin := middleware.ContextIsAdmin(r.Context())
	roleID := ""
	if user != nil {
		roleID = user.Role
	}

	field, err := svc.Schema.GetField(r.Context(), fID, roleID, isAdmin)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(field)
	return nil
}

// apiCreateField handles POST requests to create a new field
func apiCreateField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var newField db.CreateFieldParams
	err := json.NewDecoder(r.Body).Decode(&newField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	createdField, err := svc.Schema.CreateField(r.Context(), ac, newField)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdField)
	return nil
}

// apiUpdateField handles PUT requests to update an existing field
func apiUpdateField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var updateField db.UpdateFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := svc.Schema.UpdateField(r.Context(), ac, updateField)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteField handles DELETE requests for fields
func apiDeleteField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	fID := types.FieldID(q)
	if err := fID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	err = svc.Schema.DeleteField(r.Context(), ac, fID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
