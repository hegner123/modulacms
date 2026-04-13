package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

type proxyUserBackend struct{ p *proxyBackends }

func (b *proxyUserBackend) Whoami(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.Whoami(ctx)
}
func (b *proxyUserBackend) ListUsers(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.ListUsers(ctx)
}
func (b *proxyUserBackend) GetUser(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.GetUser(ctx, id)
}
func (b *proxyUserBackend) CreateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.CreateUser(ctx, params)
}
func (b *proxyUserBackend) UpdateUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.UpdateUser(ctx, params)
}
func (b *proxyUserBackend) DeleteUser(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Users.DeleteUser(ctx, id)
}
func (b *proxyUserBackend) ListUsersFull(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.ListUsersFull(ctx)
}
func (b *proxyUserBackend) GetUserFull(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.GetUserFull(ctx, id)
}
func (b *proxyUserBackend) ReassignAndDeleteUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.ReassignAndDeleteUser(ctx, params)
}
func (b *proxyUserBackend) ListUserSessions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Users.ListUserSessions(ctx)
}

// ---------------------------------------------------------------------------
// RBAC
// ---------------------------------------------------------------------------

type proxyRBACBackend struct{ p *proxyBackends }

func (b *proxyRBACBackend) ListRoles(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListRoles(ctx)
}
func (b *proxyRBACBackend) GetRole(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.GetRole(ctx, id)
}
func (b *proxyRBACBackend) CreateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.CreateRole(ctx, params)
}
func (b *proxyRBACBackend) UpdateRole(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.UpdateRole(ctx, params)
}
func (b *proxyRBACBackend) DeleteRole(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.RBAC.DeleteRole(ctx, id)
}
func (b *proxyRBACBackend) ListPermissions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListPermissions(ctx)
}
func (b *proxyRBACBackend) GetPermission(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.GetPermission(ctx, id)
}
func (b *proxyRBACBackend) CreatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.CreatePermission(ctx, params)
}
func (b *proxyRBACBackend) UpdatePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.UpdatePermission(ctx, params)
}
func (b *proxyRBACBackend) DeletePermission(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.RBAC.DeletePermission(ctx, id)
}
func (b *proxyRBACBackend) AssignRolePermission(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.AssignRolePermission(ctx, params)
}
func (b *proxyRBACBackend) RemoveRolePermission(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.RBAC.RemoveRolePermission(ctx, id)
}
func (b *proxyRBACBackend) ListRolePermissions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListRolePermissions(ctx)
}
func (b *proxyRBACBackend) GetRolePermission(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.GetRolePermission(ctx, id)
}
func (b *proxyRBACBackend) ListRolePermissionsByRole(ctx context.Context, roleID string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.RBAC.ListRolePermissionsByRole(ctx, roleID)
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

type proxySessionBackend struct{ p *proxyBackends }

func (b *proxySessionBackend) ListSessions(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Sessions.ListSessions(ctx)
}
func (b *proxySessionBackend) GetSession(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Sessions.GetSession(ctx, id)
}
func (b *proxySessionBackend) UpdateSession(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Sessions.UpdateSession(ctx, params)
}
func (b *proxySessionBackend) DeleteSession(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Sessions.DeleteSession(ctx, id)
}

// ---------------------------------------------------------------------------
// Tokens
// ---------------------------------------------------------------------------

type proxyTokenBackend struct{ p *proxyBackends }

func (b *proxyTokenBackend) ListTokens(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tokens.ListTokens(ctx)
}
func (b *proxyTokenBackend) GetToken(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tokens.GetToken(ctx, id)
}
func (b *proxyTokenBackend) CreateToken(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tokens.CreateToken(ctx, params)
}
func (b *proxyTokenBackend) DeleteToken(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Tokens.DeleteToken(ctx, id)
}

// ---------------------------------------------------------------------------
// SSHKeys
// ---------------------------------------------------------------------------

type proxySSHKeyBackend struct{ p *proxyBackends }

func (b *proxySSHKeyBackend) ListSSHKeys(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.SSHKeys.ListSSHKeys(ctx)
}
func (b *proxySSHKeyBackend) CreateSSHKey(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.SSHKeys.CreateSSHKey(ctx, params)
}
func (b *proxySSHKeyBackend) DeleteSSHKey(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.SSHKeys.DeleteSSHKey(ctx, id)
}

// ---------------------------------------------------------------------------
// OAuth
// ---------------------------------------------------------------------------

type proxyOAuthBackend struct{ p *proxyBackends }

func (b *proxyOAuthBackend) ListUsersOAuth(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.ListUsersOAuth(ctx)
}
func (b *proxyOAuthBackend) GetUserOAuth(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.GetUserOAuth(ctx, id)
}
func (b *proxyOAuthBackend) CreateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.CreateUserOAuth(ctx, params)
}
func (b *proxyOAuthBackend) UpdateUserOAuth(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.OAuth.UpdateUserOAuth(ctx, params)
}
func (b *proxyOAuthBackend) DeleteUserOAuth(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.OAuth.DeleteUserOAuth(ctx, id)
}
