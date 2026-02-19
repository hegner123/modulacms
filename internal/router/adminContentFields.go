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

// AdminContentFieldsHandler handles CRUD operations that do not require a specific ID.
func AdminContentFieldsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			err := apiListAdminContentFieldsPaginated(w, r, c)
			if err != nil {
				return
			}
		} else {
			err := apiListAdminContentFields(w, r, c)
			if err != nil {
				return
			}
		}
	case http.MethodPost:
		err := apiCreateAdminContentField(w, r, c)
		if err != nil {
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminContentFieldHandler handles CRUD operations for specific items.
func AdminContentFieldHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		err := apiGetAdminContentField(w, r, c)
		if err != nil {
			return
		}
	case http.MethodPut:
		err := apiUpdateAdminContentField(w, r, c)
		if err != nil {
			return
		}
	case http.MethodDelete:
		err := apiDeleteAdminContentField(w, r, c)
		if err != nil {
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListAdminContentFields handles GET requests for listing admin content fields
func apiListAdminContentFields(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	if r == nil {
		err := fmt.Errorf("request error")
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err

	}

	adminContentFields, err := d.ListAdminContentFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(adminContentFields)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiCreateAdminContentField handles POST requests to create a new admin content field
func apiCreateAdminContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newAdminContentField db.CreateAdminContentFieldParams
	err := json.NewDecoder(r.Body).Decode(&newAdminContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdAdminContentField, err := d.CreateAdminContentField(r.Context(), ac, newAdminContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdAdminContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiUpdateAdminContentField handles PUT requests to update an existing admin content field
func apiUpdateAdminContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateAdminContentField db.UpdateAdminContentFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateAdminContentField(r.Context(), ac, updateAdminContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetAdminContentField(updateAdminContentField.AdminContentFieldID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated admin content field", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiGetAdminContentField handles GET requests for a single admin content field
func apiGetAdminContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	acfID := types.AdminContentFieldID(q)
	if err := acfID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	adminContentField, err := d.GetAdminContentField(acfID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(adminContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiDeleteAdminContentField handles DELETE requests for admin content fields
func apiDeleteAdminContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	cfID := types.AdminContentFieldID(q)
	if err := cfID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteAdminContentField(r.Context(), ac, cfID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListAdminContentFieldsPaginated handles GET requests for listing admin content fields with pagination
func apiListAdminContentFieldsPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListAdminContentFieldsPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountAdminContentFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.AdminContentFields]{
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
