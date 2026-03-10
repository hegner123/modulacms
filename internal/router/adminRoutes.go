package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

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
