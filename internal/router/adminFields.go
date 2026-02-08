package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminFieldsHandler handles CRUD operations that do not require a specific field ID.
func AdminFieldsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListAdminFields(w, r, c)
	case http.MethodPost:
		apiCreateAdminField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminFieldHandler handles CRUD operations for specific field items.
func AdminFieldHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminField(w, r, c)
	case http.MethodPut:
		apiUpdateAdminField(w, r, c)
	case http.MethodDelete:
		apiDeleteAdminField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetAdminField handles GET requests for a single admin field
func apiGetAdminField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	afID := types.AdminFieldID(q)
	if err := afID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	adminField, err := d.GetAdminField(afID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminField)
	return nil
}

// apiListAdminFields handles GET requests for listing admin fields
func apiListAdminFields(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	if r == nil {
		err := fmt.Errorf("request error")
		http.Error(w, "request error", http.StatusInternalServerError)
		return err
	}

	adminFields, err := d.ListAdminFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminFields)
	return nil
}

// apiCreateAdminField handles POST requests to create a new admin field
func apiCreateAdminField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newAdminField db.CreateAdminFieldParams
	err := json.NewDecoder(r.Body).Decode(&newAdminField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdAdminField, err := d.CreateAdminField(r.Context(), ac, newAdminField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAdminField)
	return nil
}

// apiUpdateAdminField handles PUT requests to update an existing admin field
func apiUpdateAdminField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateAdminField db.UpdateAdminFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	updatedAdminField, err := d.UpdateAdminField(r.Context(), ac, updateAdminField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedAdminField)
	return nil
}

// apiDeleteAdminField handles DELETE requests for admin fields
func apiDeleteAdminField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	afID := types.AdminFieldID(q)
	if err := afID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteAdminField(r.Context(), ac, afID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
