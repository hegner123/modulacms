package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminFieldTypesHandler handles CRUD operations that do not require a specific admin field type ID.
func AdminFieldTypesHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListAdminFieldTypes(w, c)
	case http.MethodPost:
		apiCreateAdminFieldType(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminFieldTypeHandler handles CRUD operations for specific admin field type items.
func AdminFieldTypeHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminFieldType(w, r, c)
	case http.MethodPut:
		apiUpdateAdminFieldType(w, r, c)
	case http.MethodDelete:
		apiDeleteAdminFieldType(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetAdminFieldType handles GET requests for a single admin field type.
func apiGetAdminFieldType(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	adminFieldType, err := d.GetAdminFieldType(aftID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminFieldType)
	return nil
}

// apiListAdminFieldTypes handles GET requests for listing admin field types.
func apiListAdminFieldTypes(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	adminFieldTypes, err := d.ListAdminFieldTypes()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminFieldTypes)
	return nil
}

// apiCreateAdminFieldType handles POST requests to create a new admin field type.
func apiCreateAdminFieldType(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var params db.CreateAdminFieldTypeParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	created, err := d.CreateAdminFieldType(r.Context(), ac, params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
	return nil
}

// apiUpdateAdminFieldType handles PUT requests to update an existing admin field type.
func apiUpdateAdminFieldType(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var params db.UpdateAdminFieldTypeParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateAdminFieldType(r.Context(), ac, params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetAdminFieldType(params.AdminFieldTypeID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated admin field type", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteAdminFieldType handles DELETE requests for admin field types.
func apiDeleteAdminFieldType(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	aftID := types.AdminFieldTypeID(q)
	if err := aftID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteAdminFieldType(r.Context(), ac, aftID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
