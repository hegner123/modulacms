package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminFieldHandler handles CRUD operations for specific field items.
func AdminFieldHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminField(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminField(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminFieldHandler handles CRUD operations for specific field items.
func AdminFieldHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminField(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminField(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminField(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListAdminFieldsPaginated handles GET requests for listing admin fields with pagination
func apiListAdminFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminFieldsPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminFieldsPaginated handles GET requests for listing admin fields with pagination
func apiListAdminFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminFieldsPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminFieldsPaginated handles GET requests for listing admin fields with pagination
func apiListAdminFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminFieldsPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminFieldsPaginated handles GET requests for listing admin fields with pagination
func apiListAdminFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminFieldsPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminFieldsPaginated handles GET requests for listing admin fields with pagination
func apiListAdminFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminFieldsPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminFieldsPaginated handles GET requests for listing admin fields with pagination
func apiListAdminFieldsPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminFieldsPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}
