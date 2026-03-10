package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminFieldTypesHandler handles CRUD operations that do not require a specific admin field type ID.
func AdminFieldTypesHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListAdminFieldTypes(w, r, svc)
	case http.MethodPost:
		apiCreateAdminFieldType(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminFieldTypeHandler handles CRUD operations for specific admin field type items.
func AdminFieldTypeHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminFieldType(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminFieldType(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminFieldType(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetAdminFieldType handles GET requests for a single admin field type.
func apiGetAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	adminFieldType, err := svc.Schema.GetAdminFieldType(r.Context(), aftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminFieldType)
	return nil
}

// apiListAdminFieldTypes handles GET requests for listing admin field types.
func apiListAdminFieldTypes(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	adminFieldTypes, err := svc.Schema.ListAdminFieldTypes(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminFieldTypes)
	return nil
}

// apiCreateAdminFieldType handles POST requests to create a new admin field type.
func apiCreateAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var params db.CreateAdminFieldTypeParams
	err := json.NewDecoder(r.Body).Decode(&params)
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

	created, err := svc.Schema.CreateAdminFieldType(r.Context(), ac, params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
	return nil
}

// apiUpdateAdminFieldType handles PUT requests to update an existing admin field type.
func apiUpdateAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var params db.UpdateAdminFieldTypeParams
	err := json.NewDecoder(r.Body).Decode(&params)
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

	updated, err := svc.Schema.UpdateAdminFieldType(r.Context(), ac, params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteAdminFieldType handles DELETE requests for admin field types.
func apiDeleteAdminFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
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

	err = svc.Schema.DeleteAdminFieldType(r.Context(), ac, aftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
