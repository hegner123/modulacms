package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

func apiGetAdminDatatypeFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.Schema.GetAdminDatatypeFull(r.Context(), adtID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

func apiGetAdminDatatypeFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.Schema.GetAdminDatatypeFull(r.Context(), adtID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

func apiGetAdminDatatypeFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.Schema.GetAdminDatatypeFull(r.Context(), adtID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

func apiGetAdminDatatypeFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.Schema.GetAdminDatatypeFull(r.Context(), adtID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

func apiGetAdminDatatypeFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	adtID := types.AdminDatatypeID(q)
	if err := adtID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.Schema.GetAdminDatatypeFull(r.Context(), adtID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// Admin datatypeCascadeDeleteResponse is the JSON response for DELETE with cascade=true.
type AdminDatatypeCascadeDeleteResponse struct {
	DeletedAdminDatatypeID types.AdminDatatypeID `json:"deleted_datatype_id"`
	ContentDeleted         int                   `json:"content_deleted"`
	Errors                 []string              `json:"errors"`
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// apiListAdminDatatypesPaginated handles GET requests for listing admin datatypes with pagination
func apiListAdminDatatypesPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	params := ParsePaginationParams(r)

	response, err := svc.Schema.ListAdminDatatypesPaginated(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}
