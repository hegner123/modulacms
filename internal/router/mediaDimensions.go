package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediaDimensionsHandler handles CRUD operations that do not require a specific dimension ID.
func MediaDimensionsHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListMediaDimensions(w, c)
	case http.MethodPost:
		apiCreateMediaDimension(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MediaDimensionHandler handles CRUD operations for specific dimension items.
func MediaDimensionHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetMediaDimension(w, r, c)
	case http.MethodPut:
		apiUpdateMediaDimension(w, r, c)
	case http.MethodDelete:
		apiDeleteMediaDimension(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetMediaDimension handles GET requests for a single media dimension
func apiGetMediaDimension(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	mdID := r.URL.Query().Get("q")
	if mdID == "" {
		err := fmt.Errorf("missing media dimension ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	mediaDimension, err := d.GetMediaDimension(mdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mediaDimension)
	return nil
}

// apiListMediaDimensions handles GET requests for listing media dimensions
func apiListMediaDimensions(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	mediaDimensionsList, err := d.ListMediaDimensions()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mediaDimensionsList)
	return nil
}

// apiCreateMediaDimension handles POST requests to create a new media dimension
func apiCreateMediaDimension(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newMediaDimension db.CreateMediaDimensionParams
	err := json.NewDecoder(r.Body).Decode(&newMediaDimension)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdMediaDimension, err := d.CreateMediaDimension(r.Context(), ac, newMediaDimension)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdMediaDimension)
	return nil
}

// apiUpdateMediaDimension handles PUT requests to update an existing media dimension
func apiUpdateMediaDimension(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateMediaDimension db.UpdateMediaDimensionParams
	err := json.NewDecoder(r.Body).Decode(&updateMediaDimension)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateMediaDimension(r.Context(), ac, updateMediaDimension)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetMediaDimension(updateMediaDimension.MdID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated media dimension", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteMediaDimension handles DELETE requests for media dimensions
func apiDeleteMediaDimension(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	mdID := r.URL.Query().Get("q")
	if mdID == "" {
		err := fmt.Errorf("missing media dimension ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteMediaDimension(r.Context(), ac, mdID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
