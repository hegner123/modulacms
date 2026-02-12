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

// AdminDatatypeFieldsHandler handles CRUD operations that do not require a specific admin datatype field ID.
func AdminDatatypeFieldsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListAdminDatatypeFieldPaginated(w, r, c)
		} else {
			apiListAdminDatatypeFields(w, r, c)
		}
	case http.MethodPost:
		apiCreateAdminDatatypeField(w, r, c)
	case http.MethodDelete:
		apiDeleteAdminDatatypeField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminDatatypeFieldHandler handles CRUD operations for specific admin datatype field items.
func AdminDatatypeFieldHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodPut:
		apiUpdateAdminDatatypeField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListAdminDatatypeFields handles GET requests for listing admin datatype fields.
// Supports query filters: ?admin_datatype_id=<id>, ?admin_field_id=<id>, or no filter for all.
func apiListAdminDatatypeFields(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	adminDatatypeIDStr := r.URL.Query().Get("admin_datatype_id")
	adminFieldIDStr := r.URL.Query().Get("admin_field_id")

	var adminDatatypeFields *[]db.AdminDatatypeFields
	var err error

	if adminDatatypeIDStr != "" {
		adminDatatypeFields, err = d.ListAdminDatatypeFieldByDatatypeID(types.AdminDatatypeID(adminDatatypeIDStr))
	} else if adminFieldIDStr != "" {
		adminDatatypeFields, err = d.ListAdminDatatypeFieldByFieldID(types.AdminFieldID(adminFieldIDStr))
	} else {
		adminDatatypeFields, err = d.ListAdminDatatypeField()
	}

	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminDatatypeFields)
	return nil
}

// apiCreateAdminDatatypeField handles POST requests to create a new admin datatype field
func apiCreateAdminDatatypeField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newAdminDatatypeField db.CreateAdminDatatypeFieldParams
	err := json.NewDecoder(r.Body).Decode(&newAdminDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdAdminDatatypeField, err := d.CreateAdminDatatypeField(r.Context(), ac, newAdminDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAdminDatatypeField)
	return nil
}

// apiUpdateAdminDatatypeField handles PUT requests to update an existing admin datatype field
func apiUpdateAdminDatatypeField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateAdminDatatypeField db.UpdateAdminDatatypeFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	updatedAdminDatatypeField, err := d.UpdateAdminDatatypeField(r.Context(), ac, updateAdminDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedAdminDatatypeField)
	return nil
}

// apiDeleteAdminDatatypeField handles DELETE requests for admin datatype fields
func apiDeleteAdminDatatypeField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("id")
	if q == "" {
		err := fmt.Errorf("missing required query parameter: id")
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteAdminDatatypeField(r.Context(), ac, q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	res := fmt.Sprintf("Deleted %s", q)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	return nil
}

// apiListAdminDatatypeFieldPaginated handles GET requests for listing admin datatype fields with pagination
func apiListAdminDatatypeFieldPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListAdminDatatypeFieldPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountAdminDatatypeFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.AdminDatatypeFields]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}
