package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminFieldsHandler handles CRUD operations that do not require a specific field ID.
func AdminFieldsHandler(w http.ResponseWriter, r *http.Request, c config.Config, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListAdminFieldsPaginated(w, r, svc)
		} else {
			apiListAdminFields(w, r, svc)
		}
	case http.MethodPost:
		apiCreateAdminField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminFieldHandler handles CRUD operations for specific field items.
func AdminFieldHandler(w http.ResponseWriter, r *http.Request, c config.Config, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminField(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminField(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetAdminField handles GET requests for a single admin field
func apiGetAdminField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	afID := types.AdminFieldID(q)
	if err := afID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	adminField, err := svc.Schema.GetAdminField(r.Context(), afID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminField)
	return nil
}

// apiListAdminFields handles GET requests for listing admin fields
func apiListAdminFields(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	adminFields, err := svc.Schema.ListAdminFields(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminFields)
	return nil
}

// apiCreateAdminField handles POST requests to create a new admin field
func apiCreateAdminField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var newAdminField db.CreateAdminFieldParams
	err := json.NewDecoder(r.Body).Decode(&newAdminField)
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

	createdAdminField, err := svc.Schema.CreateAdminField(r.Context(), ac, newAdminField)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAdminField)
	return nil
}

// apiUpdateAdminField handles PUT requests to update an existing admin field
func apiUpdateAdminField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var updateAdminField db.UpdateAdminFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminField)
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

	updated, err := svc.Schema.UpdateAdminField(r.Context(), ac, updateAdminField)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteAdminField handles DELETE requests for admin fields
func apiDeleteAdminField(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	afID := types.AdminFieldID(q)
	if err := afID.Validate(); err != nil {
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

	err = svc.Schema.DeleteAdminField(r.Context(), ac, afID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListAdminFieldsPaginated handles GET requests for listing admin fields with pagination
func apiListAdminFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminFieldsPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}
