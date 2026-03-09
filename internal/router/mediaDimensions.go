package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// MediaDimensionsHandler handles CRUD operations that do not require a specific dimension ID.
func MediaDimensionsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListMediaDimensions(w, r, svc)
	case http.MethodPost:
		apiCreateMediaDimension(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MediaDimensionHandler handles CRUD operations for specific dimension items.
func MediaDimensionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetMediaDimension(w, r, svc)
	case http.MethodPut:
		apiUpdateMediaDimension(w, r, svc)
	case http.MethodDelete:
		apiDeleteMediaDimension(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetMediaDimension(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	mdID := r.URL.Query().Get("q")
	if mdID == "" {
		http.Error(w, "missing media dimension ID", http.StatusBadRequest)
		return
	}

	dim, err := svc.Media.GetMediaDimension(r.Context(), mdID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dim)
}

func apiListMediaDimensions(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	dims, err := svc.Media.ListMediaDimensions(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dims)
}

func apiCreateMediaDimension(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.CreateMediaDimensionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	created, err := svc.Media.CreateMediaDimension(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func apiUpdateMediaDimension(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.UpdateMediaDimensionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.Media.UpdateMediaDimension(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func apiDeleteMediaDimension(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	mdID := r.URL.Query().Get("q")
	if mdID == "" {
		http.Error(w, fmt.Sprintf("missing media dimension ID"), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.Media.DeleteMediaDimension(r.Context(), ac, mdID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
