package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// ContentFieldsHandler handles CRUD operations that do not require a specific field ID.
func ContentFieldsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListContentFieldsPaginated(w, r, svc)
		} else {
			apiListContentFields(w, r, svc)
		}
	case http.MethodPost:
		apiCreateContentField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ContentFieldHandler handles CRUD operations for specific field items.
func ContentFieldHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetContentField(w, r, svc)
	case http.MethodPut:
		apiUpdateContentField(w, r, svc)
	case http.MethodDelete:
		apiDeleteContentField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetContentField handles GET requests for a single content field
func apiGetContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cfID := types.ContentFieldID(q)
	if err := cfID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cf, err := svc.Content.GetField(r.Context(), cfID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, cf)
}

// apiListContentFields handles GET requests for listing content fields.
// Supports optional locale query parameter to filter by locale code.
func apiListContentFields(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	locale := r.URL.Query().Get("locale")
	contentDataIDStr := r.URL.Query().Get("content_data_id")

	if locale != "" && contentDataIDStr != "" {
		contentDataID := types.NullableContentID{
			ID:    types.ContentID(contentDataIDStr),
			Valid: true,
		}
		fields, err := svc.Content.ListFieldsByContentDataAndLocale(r.Context(), contentDataID, locale)
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}
		writeJSON(w, fields)
		return
	}

	fields, err := svc.Content.ListFields(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, fields)
}

// apiCreateContentField handles POST requests to create a new content field
func apiCreateContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.CreateContentFieldParams
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

	created, err := svc.Content.CreateField(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// apiUpdateContentField handles PUT requests to update an existing content field
func apiUpdateContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.UpdateContentFieldParams
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

	updated, err := svc.Content.UpdateField(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, updated)
}

// apiDeleteContentField handles DELETE requests for content fields
func apiDeleteContentField(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cfID := types.ContentFieldID(q)
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

	if err := svc.Content.DeleteField(r.Context(), ac, cfID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// apiListContentFieldsPaginated handles GET requests for listing content fields with pagination
func apiListContentFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	params := ParsePaginationParams(r)

	result, err := svc.Content.ListFieldsPaginated(r.Context(), params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, result)
}
