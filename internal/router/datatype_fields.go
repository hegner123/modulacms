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

// DatatypeFieldsHandler handles CRUD operations that do not require a specific datatype field ID.
func DatatypeFieldsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListDatatypeFields(w, r, c)
	case http.MethodPost:
		apiCreateDatatypeField(w, r, c)
	case http.MethodDelete:
		apiDeleteDatatypeField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DatatypeFieldHandler handles CRUD operations for specific datatype field items.
func DatatypeFieldHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodPut:
		apiUpdateDatatypeField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListDatatypeFields handles GET requests for listing datatype fields.
// Supports query filters: ?datatype_id=<id>, ?field_id=<id>, or no filter for all.
func apiListDatatypeFields(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	datatypeIDStr := r.URL.Query().Get("datatype_id")
	fieldIDStr := r.URL.Query().Get("field_id")

	var datatypeFields *[]db.DatatypeFields
	var err error

	if datatypeIDStr != "" {
		datatypeFields, err = d.ListDatatypeFieldByDatatypeID(types.DatatypeID(datatypeIDStr))
	} else if fieldIDStr != "" {
		datatypeFields, err = d.ListDatatypeFieldByFieldID(types.FieldID(fieldIDStr))
	} else {
		datatypeFields, err = d.ListDatatypeField()
	}

	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datatypeFields)
	return nil
}

// apiCreateDatatypeField handles POST requests to create a new datatype field
func apiCreateDatatypeField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newDatatypeField db.CreateDatatypeFieldParams
	err := json.NewDecoder(r.Body).Decode(&newDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if newDatatypeField.ID == "" {
		newDatatypeField.ID = string(types.NewDatatypeFieldID())
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdDatatypeField, err := d.CreateDatatypeField(r.Context(), ac, newDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// apiUpdateDatatypeField handles PUT requests to update an existing datatype field
func apiUpdateDatatypeField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateDatatypeField db.UpdateDatatypeFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	updatedDatatypeField, err := d.UpdateDatatypeField(r.Context(), ac, updateDatatypeField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedDatatypeField)
	return nil
}

// apiDeleteDatatypeField handles DELETE requests for datatype fields
func apiDeleteDatatypeField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("id")
	if q == "" {
		err := fmt.Errorf("missing required query parameter: id")
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteDatatypeField(r.Context(), ac, q)
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
