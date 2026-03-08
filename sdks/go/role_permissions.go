package modula

import (
	"context"
	"fmt"
	"net/url"
)

// RolePermissionsResource provides operations for managing role-permission associations
// in the RBAC (Role-Based Access Control) system. Each association grants a specific
// permission to a role. Users inherit permissions through their assigned role.
//
// Permissions use the "resource:operation" format (e.g., "content:read", "media:create").
// System-protected associations (bootstrap data) cannot be deleted.
// It is accessed via [Client].RolePermissions.
type RolePermissionsResource struct {
	http *httpClient
}

// List returns all role-permission associations across all roles. Each
// [RolePermission] contains the role ID, permission ID, and their respective labels.
func (r *RolePermissionsResource) List(ctx context.Context) ([]RolePermission, error) {
	var result []RolePermission
	if err := r.http.get(ctx, "/api/v1/role-permissions", nil, &result); err != nil {
		return nil, fmt.Errorf("list role-permissions: %w", err)
	}
	return result, nil
}

// Get returns a single role-permission association by its ID.
// Returns an [*ApiError] with status 404 if the association does not exist.
func (r *RolePermissionsResource) Get(ctx context.Context, id RolePermissionID) (*RolePermission, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result RolePermission
	if err := r.http.get(ctx, "/api/v1/role-permissions/", params, &result); err != nil {
		return nil, fmt.Errorf("get role-permission %s: %w", string(id), err)
	}
	return &result, nil
}

// Create grants a permission to a role by creating a new role-permission association.
// The [CreateRolePermissionParams] must specify both the role ID and permission ID.
// Returns an error if the association already exists or either ID is invalid.
func (r *RolePermissionsResource) Create(ctx context.Context, params CreateRolePermissionParams) (*RolePermission, error) {
	var result RolePermission
	if err := r.http.post(ctx, "/api/v1/role-permissions", params, &result); err != nil {
		return nil, fmt.Errorf("create role-permission: %w", err)
	}
	return &result, nil
}

// Delete removes a role-permission association, revoking the permission from the role.
// Returns an error if the association is system-protected (bootstrap data) or does not exist.
func (r *RolePermissionsResource) Delete(ctx context.Context, id RolePermissionID) error {
	params := url.Values{}
	params.Set("q", string(id))
	if err := r.http.del(ctx, "/api/v1/role-permissions/", params); err != nil {
		return fmt.Errorf("delete role-permission %s: %w", string(id), err)
	}
	return nil
}

// ListByRole returns all role-permission associations for a specific role,
// effectively listing all permissions granted to that role.
func (r *RolePermissionsResource) ListByRole(ctx context.Context, roleID RoleID) ([]RolePermission, error) {
	params := url.Values{}
	params.Set("q", string(roleID))
	var result []RolePermission
	if err := r.http.get(ctx, "/api/v1/role-permissions/role/", params, &result); err != nil {
		return nil, fmt.Errorf("list role-permissions by role %s: %w", string(roleID), err)
	}
	return result, nil
}
