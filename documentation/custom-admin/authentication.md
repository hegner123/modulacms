# Authentication and Access Control

ModulaCMS authenticates users through password login, API keys, or OAuth providers, then controls access with role-based permissions.

## Log in with a password

Authenticate with email and password. The server creates a session and returns an HTTP-only cookie valid for 24 hours.

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "your-password"}'
```

```json
{
  "user_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "email": "admin@example.com",
  "username": "admin",
  "created_at": "2026-01-15T10:00:00Z"
}
```

The response includes a `Set-Cookie` header with the session token. Include this cookie in subsequent requests.

> **Good to know**: The login endpoint is rate-limited to 10 attempts per minute per IP address.

## Use an API key

For programmatic access (CI/CD, SDKs, external integrations), create an API key and use it as a Bearer token:

```bash
curl http://localhost:8080/api/v1/media \
  -H "Authorization: Bearer mcms_01JMKX5V6QNPZ3R8W4T2YH9B0D"
```

API keys inherit the permissions of the user they belong to. Create them via the tokens endpoint or the built-in admin panel.

### Create an API key

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

> **Good to know**: Token management permissions (`tokens:read`, `tokens:create`, `tokens:delete`) are not included in the bootstrap editor or viewer roles. Only admin users can manage tokens by default.

## Use OAuth

ModulaCMS supports OAuth 2.0 with Google, GitHub, Azure AD, or any standard OAuth provider. Initiate the flow by redirecting users to `/api/v1/auth/oauth/login`. After the user authenticates with the provider, ModulaCMS provisions or links the user account, creates a session, and redirects to your configured success URL.

For setup details, see [OAuth integration](/docs/integrations/oauth).

## Get the current user

Retrieve the authenticated user's profile and role:

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

```json
{
  "user_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "email": "admin@example.com",
  "username": "admin",
  "name": "System Administrator",
  "role": "01JMKW8N3QRYZ7T1B5K6F2P4HE"
}
```

## Log out

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Clears the session cookie and invalidates the session.

## Register a user

Register a new user via the public registration endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "jdoe",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "secure-password"
  }'
```

ModulaCMS always assigns the **viewer** role to new users, regardless of any role specified in the request body. Only administrators can change a user's role after registration.

## Create and manage users

Admins can create users with any role through the users API:

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

### Update a user's role

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

### Get a user's full profile

Retrieve a user with all associated data (OAuth connections, SSH keys, sessions, tokens):

```bash
curl "http://localhost:8080/api/v1/users/full/?q=01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Delete a user

Delete a user and reassign their content to another user:

```bash
curl -X POST http://localhost:8080/api/v1/users/reassign-delete \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "01HXK...",
    "reassign_to": "01HXK..."
  }'
```

## Reset a password

Request a password reset link:

```bash
curl -X POST http://localhost:8080/api/v1/auth/request-password-reset \
  -H "Content-Type: application/json" \
  -d '{"email": "jane@example.com"}'
```

This always returns HTTP 200 with a generic message to prevent user enumeration. If the email exists and you have configured the email service, ModulaCMS sends a reset link. The reset token expires in 1 hour.

Confirm the reset:

```bash
curl -X POST http://localhost:8080/api/v1/auth/confirm-password-reset \
  -H "Content-Type: application/json" \
  -d '{"token": "a1b2c3d4...", "password": "new-secure-password"}'
```

## Manage sessions

List active sessions to see which devices and IPs have active sessions:

```bash
curl http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Invalidate a specific session:

```bash
curl -X DELETE "http://localhost:8080/api/v1/sessions/?q=01HXK9C3..." \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## Roles and permissions

ModulaCMS uses role-based access control (RBAC) to protect every API endpoint. Assign a role to each user, attach permissions to that role, and ModulaCMS enforces permission checks on every request.

### Built-in roles

ModulaCMS ships with three system-protected roles created during installation. You cannot delete or rename them.

| Role | Permissions | Description |
|------|-------------|-------------|
| `admin` | All 72 | Full access, bypasses all permission checks |
| `editor` | 36 | Content, media, routes, datatypes, fields, field types (full CRUD) |
| `viewer` | 5 | Read-only access to content, media, routes, and field types |

Admin users bypass permission checks entirely. This is a role-level bypass -- admins are not checked against the permission system at all.

### Permission format

Permissions follow the `resource:operation` format. The resource identifies what is being accessed, and the operation identifies the action.

Examples: `content:read`, `media:create`, `users:delete`, `config:update`.

HTTP methods map to operations automatically:

| HTTP Method | Operation |
|-------------|-----------|
| GET | `read` |
| POST | `create` |
| PUT / PATCH | `update` |
| DELETE | `delete` |

### Create a custom role

Create roles with any combination of permissions:

```bash
# Create the role
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"label": "contributor"}'

# List available permissions
curl http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer YOUR_API_KEY"

# Assign permissions to the role
curl -X POST http://localhost:8080/api/v1/role-permissions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"role_id": "01HXK7A1...", "permission_id": "01HXK8B2..."}'
```

### List permissions for a role

```bash
curl "http://localhost:8080/api/v1/role-permissions/role/?q=01HXK7A1..." \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Revoke a permission

```bash
curl -X DELETE "http://localhost:8080/api/v1/role-permissions/?q=ROLE_PERMISSION_ID" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

> **Good to know**: Permission changes take effect within 60 seconds. The permission cache refreshes automatically on a regular interval.

### All available permissions

The full set of 72 permissions covers these resources:

| Resource | Operations |
|----------|-----------|
| content | read, create, update, delete, publish, admin |
| datatypes | read, create, update, delete, admin |
| fields | read, create, update, delete, admin |
| media | read, create, update, delete, admin |
| routes | read, create, update, delete, admin |
| users | read, create, update, delete, admin |
| roles | read, create, update, delete, admin |
| permissions | read, create, update, delete, admin |
| sessions | read, delete, admin |
| ssh_keys | read, create, delete, admin |
| config | read, update, admin |
| admin_tree | read, create, update, delete, admin |
| field_types | read, create, update, delete, admin |
| admin_field_types | read, create, update, delete, admin |
| deploy | read, create |
| webhook | create, read, update, delete |

> **Good to know**: All bootstrap permissions are system-protected. You cannot delete or rename them, but you can create additional custom permissions for use with custom roles.

## Configure sessions

Configure session behavior in `modula.config.json`:

| Field | Default | Description |
|-------|---------|-------------|
| `cookie_name` | -- | Name of the session cookie |
| `cookie_duration` | -- | Session duration |
| `cookie_secure` | `false` | Restrict cookie to HTTPS connections |
| `cookie_samesite` | `"lax"` | SameSite attribute (`"lax"`, `"strict"`, `"none"`) |

ModulaCMS always sets cookies with `HttpOnly` enabled and `Path` set to `/`.

## Next steps

- [Build a custom admin interface](/docs/custom-admin/building-interfaces) -- create admin screens and manage content via the API
- [OAuth integration](/docs/integrations/oauth) -- set up Google, GitHub, or Azure AD login
