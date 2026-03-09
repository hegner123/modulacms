package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// PermissionsHandler handles CRUD operations that do not require a specific permission ID.
func PermissionsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListPermissions(w, r, svc)
	case http.MethodPost:
		apiCreatePermission(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// PermissionHandler handles CRUD operations for specific permission items.
func PermissionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetPermission(w, r, svc)
	case http.MethodPut:
		apiUpdatePermission(w, r, svc)
	case http.MethodDelete:
		apiDeletePermission(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetPermission(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	pID := types.PermissionID(q)
	if err := pID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	permission, err := svc.RBAC.GetPermission(r.Context(), pID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(permission)
}

func apiListPermissions(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	permissions, err := svc.RBAC.ListPermissions(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(permissions)
}

func apiCreatePermission(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}
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

	created, err := svc.RBAC.CreatePermission(r.Context(), ac, service.CreatePermissionInput{
		Label:       req.Label,
		Description: req.Description,
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func apiUpdatePermission(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req struct {
		PermissionID types.PermissionID `json:"permission_id"`
		Label        string             `json:"label"`
		Description  string             `json:"description"`
	}
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

	updated, err := svc.RBAC.UpdatePermission(r.Context(), ac, service.UpdatePermissionInput{
		PermissionID: req.PermissionID,
		Label:        req.Label,
		Description:  req.Description,
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func apiDeletePermission(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	pID := types.PermissionID(q)
	if err := pID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, cfgErr := svc.Config()
	if cfgErr != nil {
		service.HandleServiceError(w, r, cfgErr)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.RBAC.DeletePermission(r.Context(), ac, pID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
