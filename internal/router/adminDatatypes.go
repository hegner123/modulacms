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

// AdminDatatypesHandler handles CRUD operations that do not require a specific datatype ID.
func AdminDatatypesHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		err := apiListAdminDatatypes(w, c)
		fmt.Println(err)
	case http.MethodPost:
		apiCreateAdminDatatype(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminDatatypeHandler handles CRUD operations for specific datatype items.
func AdminDatatypeHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminDatatype(w, r, c)
	case http.MethodPut:
		apiUpdateAdminDatatype(w, r, c)
	case http.MethodDelete:
		apiDeleteAdminDatatype(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetAdminDatatype handles GET requests for a single admin datatype
func apiGetAdminDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	adminDatatype, err := d.GetAdminDatatypeById(adtID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminDatatype)
	return nil
}

// apiListAdminDatatypes handles GET requests for listing admin datatypes
func apiListAdminDatatypes(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	adminDatatypes, err := d.ListAdminDatatypes()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminDatatypes)
	return nil
}

// apiCreateAdminDatatype handles POST requests to create a new admin datatype
func apiCreateAdminDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newAdminDatatype db.CreateAdminDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&newAdminDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdAdminDatatype, err := d.CreateAdminDatatype(r.Context(), ac, newAdminDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAdminDatatype)
	return nil
}

// apiUpdateAdminDatatype handles PUT requests to update an existing admin datatype
func apiUpdateAdminDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateAdminDatatype db.UpdateAdminDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	updatedAdminDatatype, err := d.UpdateAdminDatatype(r.Context(), ac, updateAdminDatatype)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedAdminDatatype)
	return nil
}

// apiDeleteAdminDatatype handles DELETE requests for admin datatypes
func apiDeleteAdminDatatype(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteAdminDatatype(r.Context(), ac, adtID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
