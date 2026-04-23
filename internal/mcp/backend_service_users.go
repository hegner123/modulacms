package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// UserBackend
// ---------------------------------------------------------------------------

type svcUserBackend struct {
	svc *service.Registry
}

func (b *svcUserBackend) Whoami(ctx context.Context) (json.RawMessage, error) {
	// Whoami identifies the MCP operator. In direct mode the audit context
	// carries the user identity set at construction time.
	user, err := b.svc.Users.GetUser(ctx, AuditContextFromMCP(ctx).UserID)
	if err != nil {
		return nil, err
	}
	user.Hash = "" // never expose password hash
	return json.Marshal(user)
}

func (b *svcUserBackend) ListUsers(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Users.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	if result != nil {
		sanitizeUserList(*result)
	}
	return json.Marshal(result)
}

func (b *svcUserBackend) GetUser(ctx context.Context, id string) (json.RawMessage, error) {
	user, err := b.svc.Users.GetUser(ctx, types.UserID(id))
	if err != nil {
		return nil, err
	}
	user.Hash = "" // never expose password hash
	return json.Marshal(user)
}

func (b *svcUserBackend) CreateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateUserInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user params: %w", err)
	}
	// MCP operates as admin
	p.IsAdmin = true
	result, err := b.svc.Users.CreateUser(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	result.Hash = "" // never expose password hash
	return json.Marshal(result)
}

func (b *svcUserBackend) UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateUserInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user params: %w", err)
	}
	// MCP operates as admin
	p.IsAdmin = true
	result, err := b.svc.Users.UpdateUser(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	result.Hash = "" // never expose password hash
	return json.Marshal(result)
}

func (b *svcUserBackend) DeleteUser(ctx context.Context, id string) error {
	return b.svc.Users.DeleteUser(ctx, AuditContextFromMCP(ctx), types.UserID(id))
}

func (b *svcUserBackend) ListUsersFull(ctx context.Context) (json.RawMessage, error) {
	// ListUsersFull assembles full user views with related entities.
	users, err := b.svc.Users.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return json.Marshal([]any{})
	}
	views := make([]db.UserFullView, 0, len(*users))
	for _, u := range *users {
		view, viewErr := b.svc.Users.GetUserFull(ctx, u.UserID)
		if viewErr != nil {
			continue
		}
		views = append(views, *view)
	}
	return json.Marshal(views)
}

func (b *svcUserBackend) GetUserFull(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Users.GetUserFull(ctx, types.UserID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcUserBackend) ReassignAndDeleteUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.ReassignDeleteInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal reassign and delete user params: %w", err)
	}
	result, err := b.svc.Users.ReassignDelete(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcUserBackend) ListUserSessions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Sessions.GetSessionByUser(ctx, AuditContextFromMCP(ctx).UserID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// sanitizeUserList zeroes the Hash field on all users in a slice.
func sanitizeUserList(users []db.Users) {
	for i := range users {
		users[i].Hash = ""
	}
}

// ---------------------------------------------------------------------------
// RBACBackend
// ---------------------------------------------------------------------------

type svcRBACBackend struct {
	svc *service.Registry
}

func (b *svcRBACBackend) ListRoles(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) GetRole(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.GetRole(ctx, types.RoleID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) CreateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateRoleInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create role params: %w", err)
	}
	result, err := b.svc.RBAC.CreateRole(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) UpdateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdateRoleInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update role params: %w", err)
	}
	result, err := b.svc.RBAC.UpdateRole(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) DeleteRole(ctx context.Context, id string) error {
	return b.svc.RBAC.DeleteRole(ctx, AuditContextFromMCP(ctx), types.RoleID(id))
}

func (b *svcRBACBackend) ListPermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) GetPermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.GetPermission(ctx, types.PermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) CreatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreatePermissionInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create permission params: %w", err)
	}
	result, err := b.svc.RBAC.CreatePermission(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) UpdatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.UpdatePermissionInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update permission params: %w", err)
	}
	result, err := b.svc.RBAC.UpdatePermission(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) DeletePermission(ctx context.Context, id string) error {
	return b.svc.RBAC.DeletePermission(ctx, AuditContextFromMCP(ctx), types.PermissionID(id))
}

func (b *svcRBACBackend) AssignRolePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateRolePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal assign role permission params: %w", err)
	}
	result, err := b.svc.RBAC.CreateRolePermission(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) RemoveRolePermission(ctx context.Context, id string) error {
	return b.svc.RBAC.DeleteRolePermission(ctx, AuditContextFromMCP(ctx), types.RolePermissionID(id))
}

func (b *svcRBACBackend) ListRolePermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListRolePermissions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) GetRolePermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.GetRolePermission(ctx, types.RolePermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcRBACBackend) ListRolePermissionsByRole(ctx context.Context, roleID string) (json.RawMessage, error) {
	result, err := b.svc.RBAC.ListRolePermissionsByRoleID(ctx, types.RoleID(roleID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// SessionBackend
// ---------------------------------------------------------------------------

type svcSessionBackend struct {
	svc *service.Registry
}

func (b *svcSessionBackend) ListSessions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Sessions.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSessionBackend) GetSession(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Sessions.GetSession(ctx, types.SessionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSessionBackend) UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateSessionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update session params: %w", err)
	}
	result, err := b.svc.Sessions.UpdateSession(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSessionBackend) DeleteSession(ctx context.Context, id string) error {
	return b.svc.Sessions.DeleteSession(ctx, AuditContextFromMCP(ctx), types.SessionID(id))
}

// ---------------------------------------------------------------------------
// TokenBackend
// ---------------------------------------------------------------------------

type svcTokenBackend struct {
	svc *service.Registry
}

func (b *svcTokenBackend) ListTokens(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Tokens.ListTokens(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTokenBackend) GetToken(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Tokens.GetToken(ctx, id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTokenBackend) CreateToken(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.CreateTokenInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create token params: %w", err)
	}
	result, err := b.svc.Tokens.CreateToken(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTokenBackend) DeleteToken(ctx context.Context, id string) error {
	return b.svc.Tokens.DeleteToken(ctx, AuditContextFromMCP(ctx), id)
}

// ---------------------------------------------------------------------------
// SSHKeyBackend
// ---------------------------------------------------------------------------

type svcSSHKeyBackend struct {
	svc *service.Registry
}

func (b *svcSSHKeyBackend) ListSSHKeys(ctx context.Context) (json.RawMessage, error) {
	// List SSH keys for the audit context user.
	ac := AuditContextFromMCP(ctx)
	userID := types.NullableUserID{ID: ac.UserID, Valid: !ac.UserID.IsZero()}
	result, err := b.svc.SSHKeys.ListKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSSHKeyBackend) CreateSSHKey(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.AddSSHKeyInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create ssh key params: %w", err)
	}
	// Default to the audit context user if not set.
	if p.UserID.IsZero() {
		p.UserID = AuditContextFromMCP(ctx).UserID
	}
	result, err := b.svc.SSHKeys.AddKey(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSSHKeyBackend) DeleteSSHKey(ctx context.Context, id string) error {
	ac := AuditContextFromMCP(ctx)
	return b.svc.SSHKeys.DeleteKey(ctx, ac, ac.UserID, id)
}

// ---------------------------------------------------------------------------
// OAuthBackend
// ---------------------------------------------------------------------------

type svcOAuthBackend struct {
	svc *service.Registry
}

func (b *svcOAuthBackend) ListUsersOAuth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Driver().ListUserOauths()
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) GetUserOAuth(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Driver().GetUserOauth(types.UserOauthID(id))
	if err != nil {
		return nil, &service.NotFoundError{Resource: "user_oauth", ID: id}
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) CreateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user oauth params: %w", err)
	}
	result, err := b.svc.OAuth.CreateUserOauth(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) UpdateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user oauth params: %w", err)
	}
	result, err := b.svc.OAuth.UpdateUserOauth(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	// UpdateUserOauth returns *string. Fetch full entity if we have an ID.
	if result != nil {
		entity, fetchErr := b.svc.Driver().GetUserOauth(p.UserOauthID)
		if fetchErr == nil {
			return json.Marshal(entity)
		}
	}
	return json.Marshal(result)
}

func (b *svcOAuthBackend) DeleteUserOAuth(ctx context.Context, id string) error {
	return b.svc.OAuth.DeleteUserOauth(ctx, AuditContextFromMCP(ctx), types.UserOauthID(id))
}
