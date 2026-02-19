package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// FieldsHandler handles CRUD operations that do not require a specific field ID.
func FieldsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListFieldsPaginated(w, r, c)
		} else {
			apiListFields(w, c)
		}
	case http.MethodPost:
		apiCreateField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// FieldHandler handles CRUD operations for specific field items.
func FieldHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetField(w, r, c)
	case http.MethodPut:
		apiUpdateField(w, r, c)
	case http.MethodDelete:
		apiDeleteField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetField handles GET requests for a single field
func apiGetField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	fID := types.FieldID(q)
	if err := fID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	field, err := d.GetField(fID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(field)
	return nil
}

// apiListFields handles GET requests for listing fields
func apiListFields(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	fields, err := d.ListFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fields)
	return nil
}

// apiCreateField handles POST requests to create a new field
func apiCreateField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newField db.CreateFieldParams
	err := json.NewDecoder(r.Body).Decode(&newField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if newField.FieldID.IsZero() {
		newField.FieldID = types.NewFieldID()
	}
	now := types.NewTimestamp(time.Now().UTC())
	if !newField.DateCreated.Valid {
		newField.DateCreated = now
	}
	if !newField.DateModified.Valid {
		newField.DateModified = now
	}
	if newField.Validation == "" {
		newField.Validation = types.EmptyJSON
	}
	if newField.UIConfig == "" {
		newField.UIConfig = types.EmptyJSON
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdField, err := d.CreateField(r.Context(), ac, newField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdField)
	return nil
}

// apiUpdateField handles PUT requests to update an existing field
func apiUpdateField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateField db.UpdateFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if updateField.Validation == "" {
		updateField.Validation = types.EmptyJSON
	}
	if updateField.UIConfig == "" {
		updateField.UIConfig = types.EmptyJSON
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateField(r.Context(), ac, updateField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetField(updateField.FieldID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated field", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteField handles DELETE requests for fields
func apiDeleteField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	fID := types.FieldID(q)
	if err := fID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteField(r.Context(), ac, fID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListFieldsPaginated handles GET requests for listing fields with pagination.
func apiListFieldsPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListFieldsPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.Fields]{
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
