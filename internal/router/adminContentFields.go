package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

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
