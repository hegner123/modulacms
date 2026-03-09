package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// RBACService manages roles, permissions, role-permission associations,
// system-protected guards, and permission cache refresh.
type RBACService struct {
	driver db.DbDriver
	mgr    *config.Manager
	pc     *middleware.PermissionCache
}

// NewRBACService creates an RBACService with the given dependencies.
func NewRBACService(driver db.DbDriver, mgr *config.Manager, pc *middleware.PermissionCache) *RBACService {
	return &RBACService{driver: driver, mgr: mgr, pc: pc}
}

// --- Param Types ---

// CreateRoleInput holds caller-provided fields for creating a role.
type CreateRoleInput struct {
	Label string
}

// UpdateRoleInput holds caller-provided fields for updating a role.
type UpdateRoleInput struct {
	RoleID types.RoleID
	Label  string
}

// CreatePermissionInput holds caller-provided fields for creating a permission.
type CreatePermissionInput struct {
	Label       string
	Description string
}

// UpdatePermissionInput holds caller-provided fields for updating a permission.
type UpdatePermissionInput struct {
	PermissionID types.PermissionID
	Label        string
	Description  string
}

// --- Role Methods ---

// CreateRole validates the label, creates a non-system-protected role,
// and refreshes the permission cache.
func (s *RBACService) CreateRole(ctx context.Context, ac audited.AuditContext, input CreateRoleInput) (*db.Roles, error) {
	if input.Label == "" {
		return nil, NewValidationError("label", "label is required")
	}

	created, err := s.driver.CreateRole(ctx, ac, db.CreateRoleParams{
		Label:           input.Label,
		SystemProtected: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	s.refreshCache()
	return created, nil
}

// UpdateRole validates input, enforces system-protected label mutation guard,
// updates the role, and refreshes the permission cache.
func (s *RBACService) UpdateRole(ctx context.Context, ac audited.AuditContext, input UpdateRoleInput) (*db.Roles, error) {
	existing, err := s.driver.GetRole(input.RoleID)
	if err != nil {
		return nil, &NotFoundError{Resource: "role", ID: string(input.RoleID)}
	}

	if input.Label == "" {
		return nil, NewValidationError("label", "label is required")
	}

	// System-protected label mutation guard
	if existing.SystemProtected && input.Label != existing.Label {
		return nil, &ForbiddenError{Message: "cannot rename system-protected role"}
	}

	_, err = s.driver.UpdateRole(ctx, ac, db.UpdateRoleParams{
		Label:           input.Label,
		SystemProtected: existing.SystemProtected,
		RoleID:          input.RoleID,
	})
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	updated, err := s.driver.GetRole(input.RoleID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated role: %w", err)
	}

	s.refreshCache()
	return updated, nil
}

// DeleteRole enforces system-protected guard, checks for assigned users,
// cleans up role-permission links, deletes the role, and refreshes the cache.
func (s *RBACService) DeleteRole(ctx context.Context, ac audited.AuditContext, roleID types.RoleID) error {
	existing, err := s.driver.GetRole(roleID)
	if err != nil {
		return &NotFoundError{Resource: "role", ID: string(roleID)}
	}

	if existing.SystemProtected {
		return &ForbiddenError{Message: "cannot delete system-protected role"}
	}

	// Check for users assigned to this role
	conn, _, err := s.driver.GetConnection()
	if err != nil {
		return fmt.Errorf("get connection for user count: %w", err)
	}
	cfg, err := s.mgr.Config()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}
	dialect := db.DialectFromString(string(cfg.Db_Driver))

	userCount, err := db.QCount(ctx, conn, dialect, "users", map[string]any{"role": string(roleID)})
	if err != nil {
		return fmt.Errorf("count users for role: %w", err)
	}
	if userCount > 0 {
		return &ConflictError{
			Resource: "role",
			ID:       string(roleID),
			Detail:   fmt.Sprintf("role has %d user(s) assigned", userCount),
		}
	}

	// Delete role-permission links first (prevents orphaned junction rows)
	if err := s.driver.DeleteRolePermissionsByRoleID(ctx, ac, roleID); err != nil {
		return fmt.Errorf("delete role-permission links: %w", err)
	}

	if err := s.driver.DeleteRole(ctx, ac, roleID); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	s.refreshCache()
	return nil
}

// GetRole retrieves a role by ID with NotFoundError mapping.
func (s *RBACService) GetRole(ctx context.Context, roleID types.RoleID) (*db.Roles, error) {
	role, err := s.driver.GetRole(roleID)
	if err != nil {
		return nil, &NotFoundError{Resource: "role", ID: string(roleID)}
	}
	return role, nil
}

// GetRoleByLabel retrieves a role by label with NotFoundError mapping.
func (s *RBACService) GetRoleByLabel(ctx context.Context, label string) (*db.Roles, error) {
	role, err := s.driver.GetRoleByLabel(label)
	if err != nil {
		return nil, &NotFoundError{Resource: "role", ID: label}
	}
	return role, nil
}

// ListRoles returns all roles.
func (s *RBACService) ListRoles(ctx context.Context) (*[]db.Roles, error) {
	return s.driver.ListRoles()
}

// --- Permission Methods ---

// CreatePermission validates the label format, creates a non-system-protected
// permission, and refreshes the permission cache.
func (s *RBACService) CreatePermission(ctx context.Context, ac audited.AuditContext, input CreatePermissionInput) (*db.Permissions, error) {
	if err := middleware.ValidatePermissionLabel(input.Label); err != nil {
		return nil, NewValidationError("label", "invalid permission label format")
	}

	created, err := s.driver.CreatePermission(ctx, ac, db.CreatePermissionParams{
		Label:           input.Label,
		SystemProtected: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create permission: %w", err)
	}

	s.refreshCache()
	return created, nil
}

// UpdatePermission enforces system-protected label mutation guard, validates
// label format, updates the permission, and refreshes the cache.
func (s *RBACService) UpdatePermission(ctx context.Context, ac audited.AuditContext, input UpdatePermissionInput) (*db.Permissions, error) {
	existing, err := s.driver.GetPermission(input.PermissionID)
	if err != nil {
		return nil, &NotFoundError{Resource: "permission", ID: string(input.PermissionID)}
	}

	// System-protected label mutation guard
	if existing.SystemProtected && input.Label != existing.Label {
		return nil, &ForbiddenError{Message: "cannot rename system-protected permission"}
	}

	if err := middleware.ValidatePermissionLabel(input.Label); err != nil {
		return nil, NewValidationError("label", "invalid permission label format")
	}

	_, err = s.driver.UpdatePermission(ctx, ac, db.UpdatePermissionParams{
		Label:           input.Label,
		SystemProtected: existing.SystemProtected,
		PermissionID:    input.PermissionID,
	})
	if err != nil {
		return nil, fmt.Errorf("update permission: %w", err)
	}

	updated, err := s.driver.GetPermission(input.PermissionID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated permission: %w", err)
	}

	s.refreshCache()
	return updated, nil
}

// DeletePermission enforces system-protected guard, deletes the permission,
// and refreshes the cache.
func (s *RBACService) DeletePermission(ctx context.Context, ac audited.AuditContext, permissionID types.PermissionID) error {
	existing, err := s.driver.GetPermission(permissionID)
	if err != nil {
		return &NotFoundError{Resource: "permission", ID: string(permissionID)}
	}

	if existing.SystemProtected {
		return &ForbiddenError{Message: "cannot delete system-protected permission"}
	}

	if err := s.driver.DeletePermission(ctx, ac, permissionID); err != nil {
		return fmt.Errorf("delete permission: %w", err)
	}

	s.refreshCache()
	return nil
}

// GetPermission retrieves a permission by ID with NotFoundError mapping.
func (s *RBACService) GetPermission(ctx context.Context, permissionID types.PermissionID) (*db.Permissions, error) {
	perm, err := s.driver.GetPermission(permissionID)
	if err != nil {
		return nil, &NotFoundError{Resource: "permission", ID: string(permissionID)}
	}
	return perm, nil
}

// ListPermissions returns all permissions.
func (s *RBACService) ListPermissions(ctx context.Context) (*[]db.Permissions, error) {
	return s.driver.ListPermissions()
}

// --- Role-Permission Methods ---

// CreateRolePermission creates a role-permission link and refreshes the cache.
func (s *RBACService) CreateRolePermission(ctx context.Context, ac audited.AuditContext, params db.CreateRolePermissionParams) (*db.RolePermissions, error) {
	created, err := s.driver.CreateRolePermission(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create role permission: %w", err)
	}

	s.refreshCache()
	return created, nil
}

// DeleteRolePermission enforces system-protected junction guard, deletes the
// link, and refreshes the cache.
func (s *RBACService) DeleteRolePermission(ctx context.Context, ac audited.AuditContext, rpID types.RolePermissionID) error {
	rp, err := s.driver.GetRolePermission(rpID)
	if err != nil {
		return &NotFoundError{Resource: "role_permission", ID: string(rpID)}
	}

	role, err := s.driver.GetRole(rp.RoleID)
	if err != nil {
		return &NotFoundError{Resource: "role", ID: string(rp.RoleID)}
	}

	if role.SystemProtected {
		return &ForbiddenError{Message: "cannot modify permissions on system-protected role"}
	}

	if err := s.driver.DeleteRolePermission(ctx, ac, rpID); err != nil {
		return fmt.Errorf("delete role permission: %w", err)
	}

	s.refreshCache()
	return nil
}

// GetRolePermission retrieves a role-permission link with NotFoundError mapping.
func (s *RBACService) GetRolePermission(ctx context.Context, rpID types.RolePermissionID) (*db.RolePermissions, error) {
	rp, err := s.driver.GetRolePermission(rpID)
	if err != nil {
		return nil, &NotFoundError{Resource: "role_permission", ID: string(rpID)}
	}
	return rp, nil
}

// ListRolePermissions returns all role-permission links.
func (s *RBACService) ListRolePermissions(ctx context.Context) (*[]db.RolePermissions, error) {
	return s.driver.ListRolePermissions()
}

// ListRolePermissionsByRoleID returns role-permission links for a specific role.
func (s *RBACService) ListRolePermissionsByRoleID(ctx context.Context, roleID types.RoleID) (*[]db.RolePermissions, error) {
	if err := roleID.Validate(); err != nil {
		return nil, NewValidationError("role_id", fmt.Sprintf("invalid role_id: %v", err))
	}
	return s.driver.ListRolePermissionsByRoleID(roleID)
}

// SyncRolePermissions deletes all existing role-permission links for the role
// and creates new ones from the provided permission IDs. Failed creates
// (not-found or constraint violations) are collected — the failed PermissionIDs
// are returned as the first value; nil if all succeeded. Infrastructure errors
// (connection failure, context cancellation) abort immediately.
//
// Note: DbDriver methods hardcode d.Connection and cannot participate in an
// external *sql.Tx. The delete and creates run as separate driver calls.
// The practical risk (crash between delete and creates) is acceptable for
// this admin-only operation — the 60s periodic cache refresh recovers.
func (s *RBACService) SyncRolePermissions(ctx context.Context, ac audited.AuditContext, roleID types.RoleID, permissionIDs []types.PermissionID) ([]types.PermissionID, error) {
	// Delete all existing links for the role
	if err := s.driver.DeleteRolePermissionsByRoleID(ctx, ac, roleID); err != nil {
		return nil, fmt.Errorf("delete existing role permissions: %w", err)
	}

	// Create new links
	var failedIDs []types.PermissionID
	for _, pid := range permissionIDs {
		_, err := s.driver.CreateRolePermission(ctx, ac, db.CreateRolePermissionParams{
			RoleID:       roleID,
			PermissionID: pid,
		})
		if err != nil {
			// Distinguish data errors from infrastructure errors
			if isConstraintOrNotFound(err) {
				failedIDs = append(failedIDs, pid)
				continue
			}
			// Infrastructure failure — abort
			return failedIDs, fmt.Errorf("create role permission for %s: %w", pid, err)
		}
	}

	s.refreshCache()
	return failedIDs, nil
}

// --- Private Helpers ---

// refreshCache synchronously refreshes the permission cache.
// Errors are logged and discarded — the mutation succeeded and the
// 60-second periodic refresh (StartPeriodicRefresh) will catch up.
func (s *RBACService) refreshCache() {
	if err := s.pc.Load(s.driver); err != nil {
		utility.DefaultLogger.Error("permission cache refresh failed", err)
	}
}

// isConstraintOrNotFound checks whether an error looks like a not-found
// or constraint violation rather than an infrastructure failure.
func isConstraintOrNotFound(err error) bool {
	msg := err.Error()
	// Common patterns across SQLite, MySQL, PostgreSQL
	for _, pattern := range []string{
		"not found",
		"no rows",
		"UNIQUE constraint",
		"FOREIGN KEY constraint",
		"duplicate key",
		"foreign key constraint",
		"Duplicate entry",
		"violates unique constraint",
		"violates foreign key constraint",
	} {
		if containsCI(msg, pattern) {
			return true
		}
	}
	return false
}

// containsCI checks if s contains substr (case-insensitive).
func containsCI(s, substr string) bool {
	sLen := len(s)
	subLen := len(substr)
	if subLen > sLen {
		return false
	}
	for i := 0; i <= sLen-subLen; i++ {
		match := true
		for j := 0; j < subLen; j++ {
			sc := s[i+j]
			tc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if tc >= 'A' && tc <= 'Z' {
				tc += 32
			}
			if sc != tc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
