package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// AdminContentFieldsHandler handles CRUD operations that do not require a specific ID.
func AdminContentFieldsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListAdminContentFieldsPaginated(w, r, svc)
		} else {
			apiListAdminContentFields(w, r, svc)
		}
	case http.MethodPost:
		apiCreateAdminContentField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminContentFieldHandler handles CRUD operations for specific items.
func AdminContentFieldHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminContentField(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminContentField(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminContentField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListAdminContentFields handles GET requests for listing admin content fields
func apiListAdminContentFields(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	fields, err := svc.AdminContent.ListFields(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	writeJSON(w, fields)
}

// apiCreateAdminContentField handles POST requests to create a new admin content field
func apiCreateAdminContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.CreateAdminContentFieldParams
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

	created, err := svc.AdminContent.CreateField(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// apiUpdateAdminContentField handles PUT requests to update an existing admin content field
func apiUpdateAdminContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.UpdateAdminContentFieldParams
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

	updated, err := svc.AdminContent.UpdateField(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, updated)
}

// apiGetAdminContentField handles GET requests for a single admin content field
func apiGetAdminContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	acfID := types.AdminContentFieldID(q)
	if err := acfID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cf, err := svc.AdminContent.GetField(r.Context(), acfID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, cf)
}

// apiDeleteAdminContentField handles DELETE requests for admin content fields
func apiDeleteAdminContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cfID := types.AdminContentFieldID(q)
	if err := cfID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.AdminContent.DeleteField(r.Context(), ac, cfID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// apiListAdminContentFieldsPaginated handles GET requests for listing admin content fields with pagination
func apiListAdminContentFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	params := ParsePaginationParams(r)

	result, err := svc.AdminContent.ListFieldsPaginated(r.Context(), params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, result)
}
