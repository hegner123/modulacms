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

// PermissionsHandler handles CRUD operations that do not require a specific permission ID.
func PermissionsHandler(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) {
	switch r.Method {
	case http.MethodGet:
		apiListPermissions(w, c)
	case http.MethodPost:
		apiCreatePermission(w, r, c, pc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// PermissionHandler handles CRUD operations for specific permission items.
func PermissionHandler(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) {
	switch r.Method {
	case http.MethodGet:
		apiGetPermission(w, r, c)
	case http.MethodPut:
		apiUpdatePermission(w, r, c, pc)
	case http.MethodDelete:
		apiDeletePermission(w, r, c, pc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetPermission handles GET requests for a single permission
func apiGetPermission(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	pID := types.PermissionID(q)
	if err := pID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	permission, err := d.GetPermission(pID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(permission)
	return nil
}

// apiListPermissions handles GET requests for listing permissions
func apiListPermissions(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	permissions, err := d.ListPermissions()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(permissions)
	return nil
}

// apiCreatePermission handles POST requests to create a new permission
func apiCreatePermission(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) error {
	d := db.ConfigDB(c)

	var newPermission db.CreatePermissionParams
	err := json.NewDecoder(r.Body).Decode(&newPermission)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Validate permission label format
	if err := middleware.ValidatePermissionLabel(newPermission.Label); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid permission label"})
		return nil
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdPermission, err := d.CreatePermission(r.Context(), ac, newPermission)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPermission)

	go func() {
		if loadErr := pc.Load(db.ConfigDB(c)); loadErr != nil {
			utility.DefaultLogger.Error("permission cache refresh failed after permission create", loadErr)
		}
	}()

	return nil
}

// apiUpdatePermission handles PUT requests to update an existing permission
func apiUpdatePermission(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) error {
	d := db.ConfigDB(c)

	var updatePermission db.UpdatePermissionParams
	err := json.NewDecoder(r.Body).Decode(&updatePermission)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// System-protected label mutation guard
	existing, err := d.GetPermission(updatePermission.PermissionID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	if existing.SystemProtected && updatePermission.Label != existing.Label {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "forbidden", "detail": "cannot rename system-protected record"})
		return nil
	}

	// Validate permission label format
	if err := middleware.ValidatePermissionLabel(updatePermission.Label); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid permission label"})
		return nil
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdatePermission(r.Context(), ac, updatePermission)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetPermission(updatePermission.PermissionID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated permission", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)

	go func() {
		if loadErr := pc.Load(db.ConfigDB(c)); loadErr != nil {
			utility.DefaultLogger.Error("permission cache refresh failed after permission update", loadErr)
		}
	}()

	return nil
}

// apiDeletePermission handles DELETE requests for permissions
func apiDeletePermission(w http.ResponseWriter, r *http.Request, c config.Config, pc *middleware.PermissionCache) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	pID := types.PermissionID(q)
	if err := pID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// System-protected guard
	existing, err := d.GetPermission(pID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	if existing.SystemProtected {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "forbidden", "detail": "cannot delete system-protected record"})
		return nil
	}

	ac := middleware.AuditContextFromRequest(r, c)
	err = d.DeletePermission(r.Context(), ac, pID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	go func() {
		if loadErr := pc.Load(db.ConfigDB(c)); loadErr != nil {
			utility.DefaultLogger.Error("permission cache refresh failed after permission delete", loadErr)
		}
	}()

	return nil
}
