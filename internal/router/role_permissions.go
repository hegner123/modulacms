package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// RolePermissionsHandler handles GET (list all) and POST (create) for role_permissions.
func RolePermissionsHandler(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) {
	switch r.Method {
	case http.MethodGet:
		apiListRolePermissions(w, c)
	case http.MethodPost:
		apiCreateRolePermission(w, r, c, pc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RolePermissionHandler handles GET (read) and DELETE (delete) for a specific role_permission.
func RolePermissionHandler(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) {
	switch r.Method {
	case http.MethodGet:
		apiGetRolePermission(w, r, c)
	case http.MethodDelete:
		apiDeleteRolePermission(w, r, c, pc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RolePermissionsByRoleHandler handles GET to list permissions by role ID.
func RolePermissionsByRoleHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListRolePermissionsByRole(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiListRolePermissions(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)
	rps, err := d.ListRolePermissions()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rps)
	return nil
}

func apiGetRolePermission(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	q := r.URL.Query().Get("q")
	rpID := types.RolePermissionID(q)
	if err := rpID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	rp, err := d.GetRolePermission(rpID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rp)
	return nil
}

func apiCreateRolePermission(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) error {
	d := db.ConfigDB(c)
	var params db.CreateRolePermissionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	created, err := d.CreateRolePermission(r.Context(), ac, params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)

	// Async cache refresh
	go func() {
		if loadErr := pc.Load(db.ConfigDB(c)); loadErr != nil {
			utility.DefaultLogger.Error("permission cache refresh failed after role_permission create", loadErr)
		}
	}()
	return nil
}

func apiDeleteRolePermission(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) error {
	d := db.ConfigDB(c)
	q := r.URL.Query().Get("q")
	rpID := types.RolePermissionID(q)
	if err := rpID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// System-protected junction row guard
	rp, err := d.GetRolePermission(rpID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	role, err := d.GetRole(rp.RoleID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	if role.SystemProtected {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "forbidden",
			"detail": "cannot modify permissions on system-protected role",
		})
		return nil
	}

	ac := middleware.AuditContextFromRequest(r, c)
	if err := d.DeleteRolePermission(r.Context(), ac, rpID); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Async cache refresh
	go func() {
		if loadErr := pc.Load(db.ConfigDB(c)); loadErr != nil {
			utility.DefaultLogger.Error("permission cache refresh failed after role_permission delete", loadErr)
		}
	}()
	return nil
}

func apiListRolePermissionsByRole(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	q := r.URL.Query().Get("q")
	roleID := types.RoleID(q)
	if err := roleID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	rps, err := d.ListRolePermissionsByRoleID(roleID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rps)
	return nil
}
