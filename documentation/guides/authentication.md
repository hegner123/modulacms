# Authentication and Authorization

ModulaCMS authenticates users through password-based login, OAuth providers (Google, GitHub, Azure AD), or API keys. Once authenticated, access is controlled by a role-based permission system (RBAC) that maps roles to granular `resource:operation` permissions. All admin API endpoints and admin panel pages require authentication; public content delivery endpoints do not.

## Concepts

**Session** -- A server-side record created on login, tied to a user and stored in the database. The session token is sent to the client as an HTTP-only cookie. Sessions expire after 24 hours.

**Role** -- A named group of permissions assigned to a user. ModulaCMS ships with three bootstrap roles: admin, editor, and viewer. You can create custom roles and assign any combination of permissions.

**Permission** -- A string in `resource:operation` format (e.g., `content:read`, `media:create`) that grants access to a specific action on a specific resource. Permissions are checked on every authenticated API request.

**API key** -- A token of type `api_key` stored in the tokens table, used for programmatic access. API keys authenticate via the `Authorization: Bearer <key>` header and carry the permissions of the user they belong to.

**Permission cache** -- An in-memory mapping of roles to permissions, loaded at startup and refreshed every 60 seconds. Changes to role-permission assignments take effect within one refresh cycle without requiring a restart.

## Authentication Methods

### Password Login

Authenticate with email and password. On success, the server creates a session and sets an HTTP-only cookie.

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "your-password"}'
```

Response (HTTP 200):

```json
{
  "user_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "email": "admin@example.com",
  "username": "admin",
  "created_at": "2026-01-15T10:00:00Z"
}
```

The response includes a `Set-Cookie` header with the session cookie. Include this cookie in subsequent requests. The cookie name is configured via the `cookie_name` field in `modula.config.json`.

### OAuth Login

ModulaCMS supports OAuth 2.0 with PKCE (Proof Key for Code Exchange) for Google, GitHub, Azure AD, or any standard OAuth provider. The flow uses state parameters for CSRF protection and PKCE code verifiers for authorization code interception prevention.

**Initiate the OAuth flow:**

```bash
# Redirects the user to the OAuth provider's login page
curl -L http://localhost:8080/api/v1/auth/oauth/login
```

The server generates a state parameter and PKCE verifier, then redirects to the provider's authorization URL. After the user authenticates with the provider, the callback endpoint exchanges the authorization code for an access token, provisions or links the user account, creates a session, and redirects to the configured success URL.

**OAuth configuration in `modula.config.json`:**

```json
{
  "oauth_client_id": "your-client-id",
  "oauth_client_secret": "your-client-secret",
  "oauth_scopes": ["openid", "email", "profile"],
  "oauth_provider_name": "google",
  "oauth_redirect_url": "http://localhost:8080/api/v1/auth/oauth/callback",
  "oauth_success_redirect": "/admin/",
  "oauth_endpoint": {
    "oauth_auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
    "oauth_token_url": "https://oauth2.googleapis.com/token",
    "oauth_userinfo_url": "https://openidconnect.googleapis.com/v1/userinfo"
  }
}
```

OAuth users who do not yet have a local account are automatically provisioned and linked. Token refresh is handled transparently during session validation.

### API Key Authentication

For programmatic access (CI/CD, SDKs, external integrations), create an API key token and use it in the `Authorization` header. API keys authenticate the request as the user they belong to, inheriting that user's role and permissions.

```bash
curl http://localhost:8080/api/v1/media \
  -H "Authorization: Bearer mcms_01JMKX5V6QNPZ3R8W4T2YH9B0D"
```

API key validation checks:

1. Token exists in the database.
2. Token type is `api_key`.
3. Token is not revoked.
4. Token is not expired.
5. Token is associated with a valid user.

If any check fails, the request falls through as unauthenticated. Create API keys via the tokens API or the admin panel.

### Session Validation

The authentication middleware evaluates each request in this order:

1. Check for a session cookie (configured name from `cookie_name`).
2. Validate the cookie's session token against the database.
3. Verify the session has not expired.
4. If cookie auth fails or is absent, check the `Authorization: Bearer` header for an API key.

Unauthenticated requests to protected endpoints receive HTTP 403.

## User Registration

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

New users are always assigned the **viewer** role, regardless of any role specified in the request body. Only administrators can change a user's role after registration.

## Password Reset

ModulaCMS supports token-based password reset with optional email delivery.

**Request a reset:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/request-password-reset \
  -H "Content-Type: application/json" \
  -d '{"email": "jane@example.com"}'
```

This always returns HTTP 200 with a generic message to prevent user enumeration. If the email exists and the email service is configured, a reset link is sent. The reset token expires in 1 hour.

**Confirm the reset:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/confirm-password-reset \
  -H "Content-Type: application/json" \
  -d '{"token": "a1b2c3d4...", "password": "new-secure-password"}'
```

The token is validated and revoked after use. All other password reset tokens for the same user are also cleaned up.

## Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Clears the session cookie. Returns HTTP 200 regardless of whether a valid session existed.

## Current User

Retrieve the authenticated user's profile:

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

## Role-Based Access Control (RBAC)

### Permission Format

Permissions follow the `resource:operation` format. The resource part allows lowercase alphanumeric characters and underscores. The operation part allows lowercase alphabetic characters only.

Examples: `content:read`, `media:create`, `users:admin`, `admin_tree:update`.

### HTTP Method Mapping

Endpoints using resource-level permission checks automatically map HTTP methods to operations:

| HTTP Method | Operation |
|-------------|-----------|
| GET | `read` |
| POST | `create` |
| PUT / PATCH | `update` |
| DELETE | `delete` |

A GET request to an endpoint guarded by `RequireResourcePermission("media")` checks for `media:read`.

### Bootstrap Roles

ModulaCMS creates three system-protected roles during installation. These roles cannot be deleted or renamed.

**Admin** -- Has all 72 RBAC permissions and bypasses permission checks entirely. The admin bypass is checked before any permission evaluation, so admin users always have full access even if new permissions are added.

**Editor** -- 36 permissions. Full CRUD on content-related resources, read-only on user management:

| Resource | Permissions |
|----------|------------|
| content | read, create, update, delete |
| datatypes | read, create, update, delete |
| fields | read, create, update, delete |
| media | read, create, update, delete |
| routes | read, create, update, delete |
| admin_tree | read, create, update, delete |
| field_types | read, create, update, delete |
| admin_field_types | read, create, update, delete |
| users | read |
| sessions | read |
| ssh_keys | read |
| config | read |

**Viewer** -- 5 read-only permissions:

| Resource | Permissions |
|----------|------------|
| content | read |
| media | read |
| routes | read |
| field_types | read |
| admin_field_types | read |

### System Permissions

The full set of 72 RBAC permissions covers these resources:

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

All bootstrap permissions are system-protected and cannot be deleted or renamed via the API.

### Custom Roles

Create custom roles and assign any combination of permissions:

```bash
# Create a role
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"label": "content_manager"}'

# Assign a permission to the role
curl -X POST http://localhost:8080/api/v1/role-permissions \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"role_id": "ROLE_ID", "permission_id": "PERMISSION_ID"}'

# List permissions for a role
curl "http://localhost:8080/api/v1/role-permissions/role/?q=ROLE_ID" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Changes to role-permission mappings are picked up by the permission cache within 60 seconds.

### Authorization Behavior

- **Fail-closed.** If no permission set is found in the request context (unauthenticated or misconfigured), the request is denied with HTTP 403.
- **Admin bypass.** Admin role users bypass all permission checks. This is implemented as an explicit boolean check, not a wildcard in the permission set.
- **Rate limiting.** Auth endpoints (login, register, OAuth, password reset) are rate-limited to 10 requests per minute per IP.

## Session Configuration

Session behavior is configured in `modula.config.json`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `cookie_name` | string | -- | Name of the session cookie |
| `cookie_duration` | string | -- | Session duration |
| `cookie_secure` | bool | `false` | Restrict cookie to HTTPS connections |
| `cookie_samesite` | string | `"lax"` | SameSite attribute (`"lax"`, `"strict"`, `"none"`) |

Cookies are always set with `HttpOnly: true` to prevent JavaScript access. The `Path` is set to `/` for application-wide access.

## API Reference

All auth endpoints are prefixed with `/api/v1`. Auth endpoints are public (no authentication required) and rate-limited.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/login` | Password-based login |
| POST | `/auth/logout` | End session, clear cookie |
| GET | `/auth/me` | Get current authenticated user |
| POST | `/auth/register` | Register a new user (assigned viewer role) |
| POST | `/auth/reset` | Reset password (authenticated) |
| POST | `/auth/request-password-reset` | Request a password reset email |
| POST | `/auth/confirm-password-reset` | Confirm reset with token and new password |
| GET | `/auth/oauth/login` | Initiate OAuth flow (redirects to provider) |
| GET | `/auth/oauth/callback` | OAuth provider callback |

Authorization management endpoints (require authentication and appropriate permissions):

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/roles` | `roles:read` | List roles |
| POST | `/roles` | `roles:create` | Create a role |
| GET | `/roles/` | `roles:read` | Get a role (`?q=ROLE_ID`) |
| PUT | `/roles/` | `roles:update` | Update a role |
| DELETE | `/roles/` | `roles:delete` | Delete a role |
| GET | `/permissions` | `permissions:read` | List permissions |
| POST | `/permissions` | `permissions:create` | Create a permission |
| DELETE | `/permissions/` | `permissions:delete` | Delete a permission (`?q=PERMISSION_ID`) |
| GET | `/role-permissions` | `roles:read` | List role-permission mappings |
| POST | `/role-permissions` | `roles:create` | Assign a permission to a role |
| DELETE | `/role-permissions/` | `roles:delete` | Remove a permission from a role |
| GET | `/role-permissions/role/` | `roles:read` | List permissions for a specific role |
| POST | `/tokens` | `tokens:create` | Create a token (API key) |
| GET | `/tokens/` | `tokens:read` | Get a single token (`?q=TOKEN_ID`) |
| PUT | `/tokens/` | `tokens:update` | Update a token |
| DELETE | `/tokens/` | `tokens:delete` | Delete a token (`?q=TOKEN_ID`) |

Note: Token and import permissions (`tokens:*`, `import:*`, `plugins:*`) are enforced at the route level but are not included in the bootstrap editor or viewer roles. Only admin users can access these endpoints by default unless you create the permissions and assign them to a role.

## Notes

- **User enumeration prevention.** The password reset request endpoint always returns HTTP 200 with the same message regardless of whether the email exists.
- **System-protected records.** The three bootstrap roles (admin, editor, viewer) and all 72 bootstrap permissions cannot be deleted or renamed. Attempts return an error.
- **Permission cache latency.** After changing role-permission assignments, up to 60 seconds may elapse before the change takes effect. The cache uses a build-then-swap strategy that does not block reads during refresh.
- **OAuth token refresh.** OAuth access tokens are refreshed transparently during session validation. If refresh fails, the session remains valid -- the user is not logged out.
- **Non-admin role assignment.** Non-admin users cannot set or change roles during registration or profile updates. The API ignores any role field in the request body.
