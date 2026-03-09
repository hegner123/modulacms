package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// FieldTypesHandler handles CRUD operations that do not require a specific field type ID.
func FieldTypesHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListFieldTypes(w, r, svc)
	case http.MethodPost:
		apiCreateFieldType(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// FieldTypeHandler handles CRUD operations for specific field type items.
func FieldTypeHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetFieldType(w, r, svc)
	case http.MethodPut:
		apiUpdateFieldType(w, r, svc)
	case http.MethodDelete:
		apiDeleteFieldType(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetFieldType handles GET requests for a single field type.
func apiGetFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	ftID := types.FieldTypeID(q)
	if err := ftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	fieldType, err := svc.Schema.GetFieldType(r.Context(), ftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fieldType)
	return nil
}

// apiListFieldTypes handles GET requests for listing field types.
func apiListFieldTypes(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	fieldTypes, err := svc.Schema.ListFieldTypes(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fieldTypes)
	return nil
}

// apiCreateFieldType handles POST requests to create a new field type.
func apiCreateFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var params db.CreateFieldTypeParams
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

	created, err := svc.Schema.CreateFieldType(r.Context(), ac, params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
	return nil
}

// apiUpdateFieldType handles PUT requests to update an existing field type.
func apiUpdateFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var params db.UpdateFieldTypeParams
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

	updated, err := svc.Schema.UpdateFieldType(r.Context(), ac, params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteFieldType handles DELETE requests for field types.
func apiDeleteFieldType(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	ftID := types.FieldTypeID(q)
	if err := ftID.Validate(); err != nil {
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

	err = svc.Schema.DeleteFieldType(r.Context(), ac, ftID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
