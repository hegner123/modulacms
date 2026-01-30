package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// ContentFieldsHandler handles CRUD operations that do not require a specific field ID.
func ContentFieldsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListContentFields(w, c)
	case http.MethodPost:
		apiCreateContentField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ContentFieldHandler handles CRUD operations for specific field items.
func ContentFieldHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetContentField(w, r, c)
	case http.MethodPut:
		apiUpdateContentField(w, r, c)
	case http.MethodDelete:
		apiDeleteContentField(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetContentField handles GET requests for a single content field
func apiGetContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	cfID := types.ContentFieldID(q)
	if err := cfID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	contentField, err := d.GetContentField(cfID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contentField)
	return nil
}

// apiListContentFields handles GET requests for listing content fields
func apiListContentFields(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	contentFields, err := d.ListContentFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contentFields)
	return nil
}

// apiCreateContentField handles POST requests to create a new content field
func apiCreateContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newContentField db.CreateContentFieldParams
	err := json.NewDecoder(r.Body).Decode(&newContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	createdContentField := d.CreateContentField(newContentField)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdContentField)
	return nil
}

// apiUpdateContentField handles PUT requests to update an existing content field
func apiUpdateContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateContentField db.UpdateContentFieldParams
	err := json.NewDecoder(r.Body).Decode(&updateContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	updatedContentField, err := d.UpdateContentField(updateContentField)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedContentField)
	return nil
}

// apiDeleteContentField handles DELETE requests for content fields
func apiDeleteContentField(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	cfID := types.ContentFieldID(q)
	if err := cfID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	err := d.DeleteContentField(cfID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
