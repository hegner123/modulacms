package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// ContentDatasHandler handles CRUD operations that do not require a specific data ID.
func ContentDatasHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListContentData(w, c)
	case http.MethodPost:
		apiCreateContentData(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ContentDataHandler handles CRUD operations for specific content data items.
func ContentDataHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetContentData(w, r, c)
	case http.MethodPost:
		apiCreateContentData(w, r, c)
	case http.MethodPut:
		apiUpdateContentData(w, r, c)
	case http.MethodDelete:
		apiDeleteContentData(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetContentData handles GET requests for a single content data
func apiGetContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	contentData, err := d.GetContentData(cdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contentData)
	return nil
}

// apiListContentData handles GET requests for listing content data
func apiListContentData(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	contentDataList, err := d.ListContentData()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contentDataList)
	return nil
}

// apiCreateContentData handles POST requests to create new content data
func apiCreateContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var newContentData db.CreateContentDataParams
	err = json.NewDecoder(r.Body).Decode(&newContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	createdContentData := d.CreateContentData(newContentData)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdContentData)
	return nil
}

// apiUpdateContentData handles PUT requests to update existing content data
func apiUpdateContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	var updateContentData db.UpdateContentDataParams
	err = json.NewDecoder(r.Body).Decode(&updateContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	updatedContentData, err := d.UpdateContentData(updateContentData)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedContentData)
	return nil
}

// apiDeleteContentData handles DELETE requests for content data
func apiDeleteContentData(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer con.Close()

	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	err = d.DeleteContentData(cdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
