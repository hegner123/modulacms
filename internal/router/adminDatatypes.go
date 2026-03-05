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

// AdminDatatypesHandler handles CRUD operations that do not require a specific datatype ID.
func AdminDatatypesHandler(w http.ResponseWriter, r *http.Request, c config.Config, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListAdminDatatypesPaginated(w, r, svc)
		} else {
			apiListAdminDatatypes(w, r, svc)
		}
	case http.MethodPost:
		apiCreateAdminDatatype(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminDatatypeHandler handles CRUD operations for specific datatype items.
func AdminDatatypeHandler(w http.ResponseWriter, r *http.Request, c config.Config, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminDatatype(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminDatatype(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminDatatype(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetAdminDatatype handles GET requests for a single admin datatype
func apiGetAdminDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	adminDatatype, err := svc.Schema.GetAdminDatatype(r.Context(), adtID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminDatatype)
	return nil
}

// apiListAdminDatatypes handles GET requests for listing admin datatypes
func apiListAdminDatatypes(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	adminDatatypes, err := svc.Schema.ListAdminDatatypes(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminDatatypes)
	return nil
}

// apiCreateAdminDatatype handles POST requests to create a new admin datatype
func apiCreateAdminDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var newAdminDatatype db.CreateAdminDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&newAdminDatatype)
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

	createdAdminDatatype, err := svc.Schema.CreateAdminDatatype(r.Context(), ac, newAdminDatatype)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAdminDatatype)
	return nil
}

// apiUpdateAdminDatatype handles PUT requests to update an existing admin datatype
func apiUpdateAdminDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var updateAdminDatatype db.UpdateAdminDatatypeParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminDatatype)
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

	updated, err := svc.Schema.UpdateAdminDatatype(r.Context(), ac, updateAdminDatatype)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteAdminDatatype handles DELETE requests for admin datatypes
func apiDeleteAdminDatatype(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
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

	err = svc.Schema.DeleteAdminDatatype(r.Context(), ac, adtID)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}
