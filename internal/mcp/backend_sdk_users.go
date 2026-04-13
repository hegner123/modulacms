package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// UserBackend
// ---------------------------------------------------------------------------

type sdkUserBackend struct {
	client *modula.Client
}

func (b *sdkUserBackend) Whoami(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Auth.Me(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) ListUsers(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) GetUser(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Users.Get(ctx, modula.UserID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) CreateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateUserParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user params: %w", err)
	}
	result, err := b.client.Users.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateUserParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user params: %w", err)
	}
	result, err := b.client.Users.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) DeleteUser(ctx context.Context, id string) error {
	return b.client.Users.Delete(ctx, modula.UserID(id))
}

func (b *sdkUserBackend) ListUsersFull(ctx context.Context) (json.RawMessage, error) {
	return b.client.UsersFull.List(ctx)
}

func (b *sdkUserBackend) GetUserFull(ctx context.Context, id string) (json.RawMessage, error) {
	return b.client.UsersFull.Get(ctx, modula.UserID(id))
}

func (b *sdkUserBackend) ReassignAndDeleteUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UserReassignDeleteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal reassign and delete user params: %w", err)
	}
	result, err := b.client.UserComposite.ReassignDelete(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkUserBackend) ListUserSessions(ctx context.Context) (json.RawMessage, error) {
	me, err := b.client.Auth.Me(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve authenticated user for sessions: %w", err)
	}
	return b.client.UserSessions.GetByUser(ctx, me.UserID)
}

// ---------------------------------------------------------------------------
// RBACBackend
// ---------------------------------------------------------------------------

type sdkRBACBackend struct {
	client *modula.Client
}

func (b *sdkRBACBackend) ListRoles(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Roles.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) GetRole(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Roles.Get(ctx, modula.RoleID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) CreateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateRoleParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create role params: %w", err)
	}
	result, err := b.client.Roles.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) UpdateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateRoleParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update role params: %w", err)
	}
	result, err := b.client.Roles.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) DeleteRole(ctx context.Context, id string) error {
	return b.client.Roles.Delete(ctx, modula.RoleID(id))
}

func (b *sdkRBACBackend) ListPermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Permissions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) GetPermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Permissions.Get(ctx, modula.PermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) CreatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreatePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create permission params: %w", err)
	}
	result, err := b.client.Permissions.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) UpdatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdatePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update permission params: %w", err)
	}
	result, err := b.client.Permissions.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) DeletePermission(ctx context.Context, id string) error {
	return b.client.Permissions.Delete(ctx, modula.PermissionID(id))
}

func (b *sdkRBACBackend) AssignRolePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateRolePermissionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal assign role permission params: %w", err)
	}
	result, err := b.client.RolePermissions.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) RemoveRolePermission(ctx context.Context, id string) error {
	return b.client.RolePermissions.Delete(ctx, modula.RolePermissionID(id))
}

func (b *sdkRBACBackend) ListRolePermissions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.RolePermissions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) GetRolePermission(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.RolePermissions.Get(ctx, modula.RolePermissionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkRBACBackend) ListRolePermissionsByRole(ctx context.Context, roleID string) (json.RawMessage, error) {
	result, err := b.client.RolePermissions.ListByRole(ctx, modula.RoleID(roleID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// SessionBackend
// ---------------------------------------------------------------------------

type sdkSessionBackend struct {
	client *modula.Client
}

func (b *sdkSessionBackend) ListSessions(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Sessions.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSessionBackend) GetSession(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Sessions.Get(ctx, modula.SessionID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSessionBackend) UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateSessionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update session params: %w", err)
	}
	result, err := b.client.Sessions.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSessionBackend) DeleteSession(ctx context.Context, id string) error {
	return b.client.Sessions.Remove(ctx, modula.SessionID(id))
}

// ---------------------------------------------------------------------------
// TokenBackend
// ---------------------------------------------------------------------------

type sdkTokenBackend struct {
	client *modula.Client
}

func (b *sdkTokenBackend) ListTokens(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Tokens.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTokenBackend) GetToken(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Tokens.Get(ctx, modula.TokenID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTokenBackend) CreateToken(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateTokenParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create token params: %w", err)
	}
	result, err := b.client.Tokens.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTokenBackend) DeleteToken(ctx context.Context, id string) error {
	return b.client.Tokens.Delete(ctx, modula.TokenID(id))
}

// ---------------------------------------------------------------------------
// SSHKeyBackend
// ---------------------------------------------------------------------------

type sdkSSHKeyBackend struct {
	client *modula.Client
}

func (b *sdkSSHKeyBackend) ListSSHKeys(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.SSHKeys.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSSHKeyBackend) CreateSSHKey(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateSSHKeyParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create ssh key params: %w", err)
	}
	result, err := b.client.SSHKeys.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSSHKeyBackend) DeleteSSHKey(ctx context.Context, id string) error {
	return b.client.SSHKeys.Delete(ctx, modula.UserSshKeyID(id))
}

// ---------------------------------------------------------------------------
// OAuthBackend
// ---------------------------------------------------------------------------

type sdkOAuthBackend struct {
	client *modula.Client
}

func (b *sdkOAuthBackend) ListUsersOAuth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.UsersOauth.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) GetUserOAuth(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.UsersOauth.Get(ctx, modula.UserOauthID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) CreateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create user oauth params: %w", err)
	}
	result, err := b.client.UsersOauth.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) UpdateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateUserOauthParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update user oauth params: %w", err)
	}
	result, err := b.client.UsersOauth.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkOAuthBackend) DeleteUserOAuth(ctx context.Context, id string) error {
	return b.client.UsersOauth.Delete(ctx, modula.UserOauthID(id))
}
