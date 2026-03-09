package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// AdminRoutesHandler handles CRUD operations that do not require a specific admin route ID.
func AdminRoutesHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Query().Get("ordered") == "true" {
			apiListOrderedAdminRoutes(w, r, svc)
		} else if HasPaginationParams(r) {
			apiListAdminRoutesPaginated(w, r, svc)
		} else {
			apiListAdminRoutes(w, r, svc)
		}
	case http.MethodPost:
		apiCreateAdminRoute(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminRouteHandler handles CRUD operations for specific admin route items.
func AdminRouteHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminRoute(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminRoute(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminRoute(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetAdminRoute(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	slug := types.Slug(q)
	if err := slug.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adminRoute, err := svc.Routes.GetAdminRouteBySlug(r.Context(), slug)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminRoute)
}

func apiListAdminRoutes(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	adminRoutes, err := svc.Routes.ListAdminRoutes(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminRoutes)
}

func apiListOrderedAdminRoutes(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	result, err := svc.Routes.ListOrderedAdminRoutes(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func apiCreateAdminRoute(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req service.CreateRouteInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	created, err := svc.Routes.CreateAdminRoute(r.Context(), ac, req)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func apiUpdateAdminRoute(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req service.UpdateRouteInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.Routes.UpdateAdminRoute(r.Context(), ac, req)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func apiDeleteAdminRoute(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	id := types.AdminRouteID(q)
	if err := id.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.Routes.DeleteAdminRoute(r.Context(), ac, id); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func apiListAdminRoutesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	params := ParsePaginationParams(r)

	items, total, err := svc.Routes.ListAdminRoutesPaginated(r.Context(), params.Limit, params.Offset)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	response := db.PaginatedResponse[db.AdminRoutes]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
