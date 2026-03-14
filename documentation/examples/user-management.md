# User Management

Recipes for managing users, roles, permissions, tokens, SSH keys, and sessions. ModulaCMS uses role-based access control (RBAC) with `resource:operation` granular permissions.

For background, see [authentication and authorization](../guides/authentication.md).

## Bootstrap Roles

ModulaCMS ships with three built-in roles. These are system-protected and cannot be deleted or renamed.

| Role | Permissions | Description |
|------|-------------|-------------|
| `admin` | All 72 permissions | Full access, bypasses permission checks |
| `editor` | 36 permissions | Content management, media, routes, datatypes, fields, field types |
| `viewer` | 5 permissions | Read-only access (content, media, routes, field types) |

## Create a User

New users created via the admin API can be assigned any role. Users created via `/auth/register` are always assigned the `viewer` role.

**curl:**

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "jdoe",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "secure-password-123",
    "role": "editor"
  }'
```

Response (201):

```json
{
  "user_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
  "username": "jdoe",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "role": "editor",
  "date_created": "2026-01-15T10:00:00Z",
  "date_modified": "2026-01-15T10:00:00Z"
}
```

**Go SDK:**

```go
user, err := client.Users.Create(ctx, modula.CreateUserParams{
    Username: "jdoe",
    Name:     "Jane Doe",
    Email:    modula.Email("jane@example.com"),
    Password: "secure-password-123",
    Role:     "editor",
})
if err != nil {
    // handle error
}

fmt.Printf("Created user %s with role %s\n", user.UserID, user.Role)
```

**TypeScript SDK (admin):**

```typescript
const user = await admin.users.create({
  username: 'jdoe',
  name: 'Jane Doe',
  email: 'jane@example.com' as Email,
  password: 'secure-password-123',
  role: 'editor',
})

console.log(`Created user ${user.user_id} with role ${user.role}`)
```

## Assign a Role

Update a user's role by setting the `role` field. The role must be a valid role label.

**curl:**

```bash
curl -X PUT http://localhost:8080/api/v1/users/ \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "username": "jdoe",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "role": "admin"
  }'
```

**Go SDK:**

```go
updated, err := client.Users.Update(ctx, modula.UpdateUserParams{
    UserID:   modula.UserID("01HXK4N2F8RJZGP6VTQY3MCSW9"),
    Username: "jdoe",
    Name:     "Jane Doe",
    Email:    modula.Email("jane@example.com"),
    Role:     "admin",
})
```

**TypeScript SDK (admin):**

```typescript
const updated = await admin.users.update({
  user_id: '01HXK4N2F8RJZGP6VTQY3MCSW9' as UserID,
  username: 'jdoe',
  name: 'Jane Doe',
  email: 'jane@example.com' as Email,
  role: 'admin',
})
```

## Create a Custom Role with Specific Permissions

Custom roles are created in two steps: create the role, then assign permissions to it via the role-permissions junction table.

**Step 1: Create the role.**

**curl:**

```bash
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"label": "contributor"}'
```

**Step 2: Assign permissions.**

```bash
# Get the list of available permissions
curl http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer YOUR_API_KEY"

# Assign specific permissions to the new role
curl -X POST http://localhost:8080/api/v1/role-permissions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"role_id": "01HXK7A1...", "permission_id": "01HXK8B2..."}'
```

**Go SDK:**

```go
// Step 1: Create role
role, err := client.Roles.Create(ctx, modula.CreateRoleParams{
    Label: "contributor",
})
if err != nil {
    // handle error
}

// Step 2: List all available permissions
perms, err := client.Permissions.List(ctx)
if err != nil {
    // handle error
}

// Step 3: Assign selected permissions
wantedPerms := []string{"content:read", "content:create", "media:read", "media:create"}
for _, p := range perms {
    for _, wanted := range wantedPerms {
        if p.Label == wanted {
            _, err := client.RolePermissions.Create(ctx, modula.CreateRolePermissionParams{
                RoleID:       role.RoleID,
                PermissionID: p.PermissionID,
            })
            if err != nil {
                // handle error
            }
        }
    }
}
```

**TypeScript SDK (admin):**

```typescript
// Step 1: Create role
const role = await admin.roles.create({ label: 'contributor' })

// Step 2: List available permissions
const perms = await admin.permissions.list()

// Step 3: Assign selected permissions
const wanted = ['content:read', 'content:create', 'media:read', 'media:create']
for (const p of perms.filter(p => wanted.includes(p.label))) {
  await admin.rolePermissions.create({
    role_id: role.role_id,
    permission_id: p.permission_id,
  })
}
```

## List Permissions for a Role

**curl:**

```bash
curl "http://localhost:8080/api/v1/role-permissions/role/?q=01HXK7A1..." \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
rolePerms, err := client.RolePermissions.ListByRole(ctx, modula.RoleID("01HXK7A1..."))
if err != nil {
    // handle error
}

for _, rp := range rolePerms {
    fmt.Printf("Permission: %s\n", rp.PermissionID)
}
```

**TypeScript SDK (admin):**

```typescript
const rolePerms = await admin.rolePermissions.listByRole('01HXK7A1...' as RoleID)

for (const rp of rolePerms) {
  console.log(`Permission: ${rp.permission_id}`)
}
```

## Generate an API Token

API tokens authenticate programmatic access. The token carries the permissions of the user it belongs to.

**curl:**

```bash
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "token_type": "api_key",
    "token": "my-generated-token-value",
    "issued_at": "2026-01-15T10:00:00Z",
    "expires_at": "2027-01-15T10:00:00Z",
    "revoked": false
  }'
```

**Go SDK:**

```go
userID := modula.UserID("01HXK4N2F8RJZGP6VTQY3MCSW9")
token, err := client.Tokens.Create(ctx, modula.CreateTokenParams{
    UserID:    &userID,
    TokenType: "api_key",
    Token:     "my-generated-token-value",
    IssuedAt:  "2026-01-15T10:00:00Z",
    ExpiresAt: modula.Timestamp("2027-01-15T10:00:00Z"),
    Revoked:   false,
})
if err != nil {
    // handle error
}

fmt.Printf("Token created: %s\n", token.ID)
```

**TypeScript SDK (admin):**

```typescript
const token = await admin.tokens.create({
  user_id: '01HXK4N2F8RJZGP6VTQY3MCSW9' as UserID,
  token_type: 'api_key',
  token: 'my-generated-token-value',
  issued_at: '2026-01-15T10:00:00Z',
  expires_at: '2027-01-15T10:00:00Z',
  revoked: false,
})
```

Revoke a token:

```go
_, err = client.Tokens.Update(ctx, modula.UpdateTokenParams{
    ID:        token.ID,
    Token:     token.Token,
    IssuedAt:  token.IssuedAt,
    ExpiresAt: token.ExpiresAt,
    Revoked:   true,
})
```

## Manage SSH Keys

SSH keys authenticate users for TUI access via the built-in SSH server.

### Add an SSH Key

**curl:**

```bash
curl -X POST http://localhost:8080/api/v1/ssh-keys \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample...", "label": "work laptop"}'
```

**Go SDK:**

```go
key, err := client.SSHKeys.Create(ctx, modula.CreateSSHKeyParams{
    PublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample...",
    Label:     "work laptop",
})
if err != nil {
    // handle error
}

fmt.Printf("Key added: %s (fingerprint: %s)\n", key.Label, key.Fingerprint)
```

**TypeScript SDK (admin):**

```typescript
const key = await admin.sshKeys.create({
  public_key: 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample...',
  label: 'work laptop',
})

console.log(`Key added: ${key.label} (fingerprint: ${key.fingerprint})`)
```

### List SSH Keys

**curl:**

```bash
curl http://localhost:8080/api/v1/ssh-keys \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
keys, err := client.SSHKeys.List(ctx)
if err != nil {
    // handle error
}

for _, k := range keys {
    fmt.Printf("%s  %s  %s  (last used: %s)\n",
        k.SshKeyID, k.KeyType, k.Label, k.LastUsed)
}
```

**TypeScript SDK (admin):**

```typescript
const keys = await admin.sshKeys.list()

for (const k of keys) {
  console.log(`${k.ssh_key_id}  ${k.key_type}  ${k.label}  (last used: ${k.last_used})`)
}
```

### Delete an SSH Key

**curl:**

```bash
curl -X DELETE "http://localhost:8080/api/v1/ssh-keys/01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
err := client.SSHKeys.Delete(ctx, modula.UserSshKeyID("01HXK4N2F8RJZGP6VTQY3MCSW9"))
```

**TypeScript SDK (admin):**

```typescript
await admin.sshKeys.remove('01HXK4N2F8RJZGP6VTQY3MCSW9')
```

## List Active Sessions

Sessions are created on login. List them to see which devices/IPs have active sessions for your users.

**curl:**

```bash
curl http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
sessions, err := client.Sessions.List(ctx)
if err != nil {
    // handle error
}

for _, s := range sessions {
    ip := ""
    if s.IpAddress != nil {
        ip = *s.IpAddress
    }
    ua := ""
    if s.UserAgent != nil {
        ua = *s.UserAgent
    }
    fmt.Printf("Session %s: IP=%s UA=%s expires=%s\n",
        s.SessionID, ip, ua, s.ExpiresAt)
}
```

**TypeScript SDK (admin):**

```typescript
// List is not exposed; sessions are managed via update and remove
// Use the REST API directly to list sessions

// Invalidate a specific session
await admin.sessions.remove('01HXK9C3...' as SessionID)
```

### Invalidate a Session

**curl:**

```bash
curl -X DELETE "http://localhost:8080/api/v1/sessions/?q=01HXK9C3..." \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
err := client.Sessions.Remove(ctx, modula.SessionID("01HXK9C3..."))
```

**TypeScript SDK (admin):**

```typescript
await admin.sessions.remove('01HXK9C3...' as SessionID)
```

## Get a User's Full Profile

Retrieve a user with all associated data (OAuth connections, SSH keys, sessions, tokens).

**curl:**

```bash
curl "http://localhost:8080/api/v1/users/full/?q=01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**TypeScript SDK (admin):**

```typescript
const fullUser = await admin.users.getFull('01HXK4N2F8RJZGP6VTQY3MCSW9' as UserID)
console.log(`User: ${fullUser.username}, Role: ${fullUser.role}`)
```

## List Users with Role Labels

**curl:**

```bash
curl http://localhost:8080/api/v1/users/full \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**TypeScript SDK (admin):**

```typescript
const users = await admin.users.listFull()

for (const u of users) {
  console.log(`${u.username} (${u.role_label})`)
}
```

## Next Steps

- [Authentication](../guides/authentication.md) -- full auth and RBAC documentation
- [Webhook Integration](webhook-integration.md) -- receive notifications on user events
