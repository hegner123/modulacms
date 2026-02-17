package modulacms

import (
	"context"
	"fmt"
	"net/url"
)

// RolePermissionsResource provides operations for managing role-permission associations.
type RolePermissionsResource struct {
	http *httpClient
}

// List returns all role-permission associations.
func (r *RolePermissionsResource) List(ctx context.Context) ([]RolePermission, error) {
	var result []RolePermission
	if err := r.http.get(ctx, "/api/v1/role-permissions", nil, &result); err != nil {
		return nil, fmt.Errorf("list role-permissions: %w", err)
	}
	return result, nil
}

// Get returns a single role-permission association by ID.
func (r *RolePermissionsResource) Get(ctx context.Context, id RolePermissionID) (*RolePermission, error) {
	params := url.Values{}
	params.Set("q", string(id))
	var result RolePermission
	if err := r.http.get(ctx, "/api/v1/role-permissions/", params, &result); err != nil {
		return nil, fmt.Errorf("get role-permission %s: %w", string(id), err)
	}
	return &result, nil
}

// Create creates a new role-permission association and returns it.
func (r *RolePermissionsResource) Create(ctx context.Context, params CreateRolePermissionParams) (*RolePermission, error) {
	var result RolePermission
	if err := r.http.post(ctx, "/api/v1/role-permissions", params, &result); err != nil {
		return nil, fmt.Errorf("create role-permission: %w", err)
	}
	return &result, nil
}

// Delete removes a role-permission association by ID.
func (r *RolePermissionsResource) Delete(ctx context.Context, id RolePermissionID) error {
	params := url.Values{}
	params.Set("q", string(id))
	if err := r.http.del(ctx, "/api/v1/role-permissions/", params); err != nil {
		return fmt.Errorf("delete role-permission %s: %w", string(id), err)
	}
	return nil
}

// ListByRole returns all role-permission associations for a given role.
func (r *RolePermissionsResource) ListByRole(ctx context.Context, roleID RoleID) ([]RolePermission, error) {
	params := url.Values{}
	params.Set("q", string(roleID))
	var result []RolePermission
	if err := r.http.get(ctx, "/api/v1/role-permissions/role/", params, &result); err != nil {
		return nil, fmt.Errorf("list role-permissions by role %s: %w", string(roleID), err)
	}
	return result, nil
}
