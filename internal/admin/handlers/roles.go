package handlers

import (
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// RolesListHandler handles GET /admin/users/roles.
// Lists roles with a sidebar; the first role's detail is rendered by default.
func RolesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles, err := svc.RBAC.ListRoles(r.Context())
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
		perms, permsErr := svc.RBAC.ListPermissions(r.Context())
		if permsErr != nil {
			utility.DefaultLogger.Error("failed to list permissions", permsErr)
			perms = &[]db.Permissions{}
		}

		// Load all role-permission links
		rolePerms, rpErr := svc.RBAC.ListRolePermissions(r.Context())
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
func RoleDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing role ID", http.StatusBadRequest)
			return
		}

		role, err := svc.RBAC.GetRole(r.Context(), types.RoleID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get role", err)
			http.Error(w, "Role not found", http.StatusNotFound)
			return
		}

		perms, permsErr := svc.RBAC.ListPermissions(r.Context())
		if permsErr != nil {
			utility.DefaultLogger.Error("failed to list permissions", permsErr)
			perms = &[]db.Permissions{}
		}

		rolePerms, rpErr := svc.RBAC.ListRolePermissions(r.Context())
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
func RoleNewFormHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		perms, permsErr := svc.RBAC.ListPermissions(r.Context())
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
// Creates a new role via RBACService (which refreshes the permission cache).
func RoleCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		label := strings.TrimSpace(r.FormValue("label"))

		ac := middleware.AuditContextFromRequest(r, *c)

		_, err := svc.RBAC.CreateRole(r.Context(), ac, service.CreateRoleInput{
			Label: label,
		})
		if err != nil {
			if service.IsValidation(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				errs := serviceValidationToMap(err)
				Render(w, r, partials.RoleForm(label, errs, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create role", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RoleForm(label, map[string]string{"_": "failed to create role"}, csrfToken))
			return
		}

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
// Updates role label and syncs permission assignments via RBACService.
func RoleUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing role ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		label := strings.TrimSpace(r.FormValue("label"))
		permissionIDs := r.Form["permissions"]

		ac := middleware.AuditContextFromRequest(r, *c)

		// Update role label
		_, err := svc.RBAC.UpdateRole(r.Context(), ac, service.UpdateRoleInput{
			RoleID: types.RoleID(id),
			Label:  label,
		})
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		// Sync permissions
		var pids []types.PermissionID
		for _, pid := range permissionIDs {
			trimmed := strings.TrimSpace(pid)
			if trimmed != "" {
				pids = append(pids, types.PermissionID(trimmed))
			}
		}

		failedIDs, syncErr := svc.RBAC.SyncRolePermissions(r.Context(), ac, types.RoleID(id), pids)
		if syncErr != nil {
			utility.DefaultLogger.Error("failed to sync role permissions", syncErr)
		}
		if len(failedIDs) > 0 {
			utility.DefaultLogger.Error("some permissions failed to assign", nil, "failed_ids", failedIDs)
		}

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
// HTMX-only endpoint. Delegates system-protected guard and cleanup to RBACService.
func RoleDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing role ID", http.StatusBadRequest)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *c)
		err := svc.RBAC.DeleteRole(r.Context(), ac, types.RoleID(id))
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

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
