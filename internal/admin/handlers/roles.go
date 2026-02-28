package handlers

import (
	"net"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// RolesListHandler handles GET /admin/users/roles.
// Lists roles with a sidebar; the first role's detail is rendered by default.
func RolesListHandler(driver db.DbDriver, pc *middleware.PermissionCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles, err := driver.ListRoles()
		if err != nil {
			utility.DefaultLogger.Error("failed to list roles", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		roleList := make([]db.Roles, 0)
		if roles != nil {
			roleList = *roles
		}

		// Load all permissions for the permission assignment UI
		perms, permsErr := driver.ListPermissions()
		if permsErr != nil {
			utility.DefaultLogger.Error("failed to list permissions", permsErr)
			perms = &[]db.Permissions{}
		}

		// Load all role-permission links
		rolePerms, rpErr := driver.ListRolePermissions()
		if rpErr != nil {
			utility.DefaultLogger.Error("failed to list role permissions", rpErr)
			rolePerms = &[]db.RolePermissions{}
		}

		rolePermMap := buildRolePermMap(*rolePerms)

		if IsNavHTMX(r) {
			var defaultMatrix partials.PermissionMatrix
			if len(roleList) > 0 {
				defaultMatrix = partials.BuildPermissionMatrix(*perms, rolePermMap[roleList[0].RoleID])
			}
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Roles"}`)
			Render(w, r, pages.RolesListContent(roleList, defaultMatrix, csrfToken))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.RolesTableRows(roleList, *perms, rolePermMap))
			return
		}

		// Build the permission matrix for the first (default-selected) role
		var defaultMatrix partials.PermissionMatrix
		if len(roleList) > 0 {
			defaultMatrix = partials.BuildPermissionMatrix(*perms, rolePermMap[roleList[0].RoleID])
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Roles")
		Render(w, r, pages.RolesList(layout, roleList, defaultMatrix, csrfToken))
	}
}

// RoleDetailHandler handles GET /admin/users/roles/{id}.
// Returns the detail partial for a single role (HTMX swap target).
func RoleDetailHandler(driver db.DbDriver, pc *middleware.PermissionCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing role ID", http.StatusBadRequest)
			return
		}

		role, err := driver.GetRole(types.RoleID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get role", err)
			http.Error(w, "Role not found", http.StatusNotFound)
			return
		}

		perms, permsErr := driver.ListPermissions()
		if permsErr != nil {
			utility.DefaultLogger.Error("failed to list permissions", permsErr)
			perms = &[]db.Permissions{}
		}

		rolePerms, rpErr := driver.ListRolePermissions()
		if rpErr != nil {
			utility.DefaultLogger.Error("failed to list role permissions", rpErr)
			rolePerms = &[]db.RolePermissions{}
		}

		rolePermMap := buildRolePermMap(*rolePerms)
		matrix := partials.BuildPermissionMatrix(*perms, rolePermMap[role.RoleID])

		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, partials.RoleDetail(*role, matrix, csrfToken))
	}
}

// RoleNewFormHandler handles GET /admin/users/roles/new.
// Returns the new-role form partial (HTMX swap target).
func RoleNewFormHandler(driver db.DbDriver, pc *middleware.PermissionCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		perms, permsErr := driver.ListPermissions()
		if permsErr != nil {
			utility.DefaultLogger.Error("failed to list permissions", permsErr)
			perms = &[]db.Permissions{}
		}

		// Empty active permissions — new role starts with nothing
		matrix := partials.BuildPermissionMatrix(*perms, nil)

		csrfToken := CSRFTokenFromContext(r.Context())
		Render(w, r, partials.RoleNewForm(matrix, csrfToken))
	}
}

// RoleCreateHandler handles POST /admin/users/roles.
// Creates a new role and refreshes the permission cache.
func RoleCreateHandler(driver db.DbDriver, pc *middleware.PermissionCache, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		label := strings.TrimSpace(r.FormValue("label"))

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RoleForm(label, errs, csrfToken))
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		_, err := driver.CreateRole(r.Context(), ac, db.CreateRoleParams{
			Label:           label,
			SystemProtected: false,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create role", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RoleForm(label, map[string]string{"_": "Failed to create role"}, csrfToken))
			return
		}

		// Refresh permission cache asynchronously
		go pc.Load(driver)

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/users/roles", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Role created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/users/roles")
		w.WriteHeader(http.StatusOK)
	}
}

// RoleUpdateHandler handles POST /admin/users/roles/{id}.
// Updates a role label and manages permission assignments.
// System-protected roles cannot be renamed.
func RoleUpdateHandler(driver db.DbDriver, pc *middleware.PermissionCache, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing role ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		existing, err := driver.GetRole(types.RoleID(id))
		if err != nil {
			http.Error(w, "Role not found", http.StatusNotFound)
			return
		}

		label := strings.TrimSpace(r.FormValue("label"))
		permissionIDs := r.Form["permissions"]

		errs := make(map[string]string)
		if label == "" {
			errs["label"] = "Label is required"
		}
		if existing.SystemProtected && label != existing.Label {
			errs["label"] = "System-protected roles cannot be renamed"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "`+errs["label"]+`", "type": "error"}}`)
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			http.Error(w, errs["label"], http.StatusUnprocessableEntity)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		// Update role label
		_, err = driver.UpdateRole(r.Context(), ac, db.UpdateRoleParams{
			RoleID:          types.RoleID(id),
			Label:           label,
			SystemProtected: existing.SystemProtected,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to update role", err)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update role", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to update role", http.StatusInternalServerError)
			return
		}

		// Sync permissions: delete all existing, then re-create from form
		if deleteErr := driver.DeleteRolePermissionsByRoleID(r.Context(), ac, types.RoleID(id)); deleteErr != nil {
			utility.DefaultLogger.Error("failed to clear role permissions", deleteErr)
		}

		for _, permID := range permissionIDs {
			trimmed := strings.TrimSpace(permID)
			if trimmed == "" {
				continue
			}
			_, createErr := driver.CreateRolePermission(r.Context(), ac, db.CreateRolePermissionParams{
				RoleID:       types.RoleID(id),
				PermissionID: types.PermissionID(trimmed),
			})
			if createErr != nil {
				utility.DefaultLogger.Error("failed to assign permission to role", createErr, "permission_id", trimmed)
			}
		}

		// Refresh permission cache asynchronously
		go pc.Load(driver)

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/users/roles", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Role updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/users/roles")
		w.WriteHeader(http.StatusOK)
	}
}

// RoleDeleteHandler handles DELETE /admin/users/roles/{id}.
// HTMX-only endpoint. Cannot delete system-protected roles.
func RoleDeleteHandler(driver db.DbDriver, pc *middleware.PermissionCache, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing role ID", http.StatusBadRequest)
			return
		}

		// Check if role is system-protected
		role, err := driver.GetRole(types.RoleID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get role", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Role not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if role.SystemProtected {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Cannot delete system-protected role", "type": "error"}}`)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		// Delete role-permission links first
		if deleteErr := driver.DeleteRolePermissionsByRoleID(r.Context(), ac, types.RoleID(id)); deleteErr != nil {
			utility.DefaultLogger.Error("failed to clear role permissions before delete", deleteErr)
		}

		err = driver.DeleteRole(r.Context(), ac, types.RoleID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete role", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete role", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Refresh permission cache asynchronously
		go pc.Load(driver)

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Role deleted", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/users/roles")
		w.WriteHeader(http.StatusOK)
	}
}

// buildRolePermMap builds a map of role ID to a set of permission IDs.
func buildRolePermMap(rolePerms []db.RolePermissions) map[types.RoleID]map[types.PermissionID]bool {
	m := make(map[types.RoleID]map[types.PermissionID]bool)
	for _, rp := range rolePerms {
		if m[rp.RoleID] == nil {
			m[rp.RoleID] = make(map[types.PermissionID]bool)
		}
		m[rp.RoleID][rp.PermissionID] = true
	}
	return m
}
