# OAuth

Set up third-party login with Google, GitHub, Azure AD, or any OpenID Connect-compatible provider.

## How OAuth Login Works

ModulaCMS supports OAuth 2.0 with PKCE (Proof Key for Code Exchange). The flow works like this:

1. A user visits your login page and clicks "Sign in with Google" (or another provider).
2. Your frontend redirects to `GET /api/v1/auth/oauth/login`.
3. ModulaCMS generates a state parameter and PKCE verifier, then redirects the user to the provider's authorization page.
4. The user authenticates with the provider and grants access.
5. The provider redirects back to `GET /api/v1/auth/oauth/callback` with an authorization code.
6. ModulaCMS exchanges the code for an access token, retrieves the user's profile, creates or links a local account, starts a session, and redirects to `oauth_success_redirect`.

> **Good to know**: ModulaCMS automatically provisions a local account for first-time OAuth users. If a user with the same email already exists, the OAuth identity is linked to the existing account. New OAuth users are assigned the **viewer** role. Only administrators can change a user's role after login.

## Configuration

Set these fields in `modula.config.json`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `oauth_client_id` | string | `""` | OAuth client ID from your provider |
| `oauth_client_secret` | string | `""` | OAuth client secret from your provider |
| `oauth_scopes` | string[] | `["openid","profile","email"]` | OAuth scopes to request |
| `oauth_provider_name` | string | `""` | Provider name (for display in the admin panel) |
| `oauth_redirect_url` | string | `""` | Callback URL registered with your provider |
| `oauth_success_redirect` | string | `"/"` | URL to redirect after successful login |
| `oauth_endpoint` | object | `{}` | Provider endpoint URLs (see provider examples) |

The `oauth_endpoint` object has three required keys:

| Key | Description |
|-----|-------------|
| `oauth_auth_url` | Provider's authorization endpoint |
| `oauth_token_url` | Provider's token exchange endpoint |
| `oauth_userinfo_url` | Provider's user info endpoint |

All OAuth fields are hot-reloadable. OAuth is optional -- ModulaCMS works with local password authentication when OAuth is not configured.

## Set Up Google

1. Go to the [Google Cloud Console](https://console.cloud.google.com/apis/credentials).
2. Create an OAuth 2.0 Client ID.
3. Set the authorized redirect URI to `https://your-cms-domain.com/api/v1/auth/oauth/callback`.
4. Copy the client ID and client secret.

```json
{
  "oauth_client_id": "123456789-abcdefg.apps.googleusercontent.com",
  "oauth_client_secret": "GOCSPX-your-client-secret",
  "oauth_scopes": ["openid", "email", "profile"],
  "oauth_provider_name": "google",
  "oauth_redirect_url": "https://your-cms-domain.com/api/v1/auth/oauth/callback",
  "oauth_success_redirect": "/admin/",
  "oauth_endpoint": {
    "oauth_auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
    "oauth_token_url": "https://oauth2.googleapis.com/token",
    "oauth_userinfo_url": "https://openidconnect.googleapis.com/v1/userinfo"
  }
}
```

## Set Up GitHub

1. Go to [GitHub Developer Settings > OAuth Apps](https://github.com/settings/developers).
2. Create a new OAuth App.
3. Set the authorization callback URL to `https://your-cms-domain.com/api/v1/auth/oauth/callback`.
4. Copy the client ID and generate a client secret.

```json
{
  "oauth_client_id": "Iv1.abc123def456",
  "oauth_client_secret": "your-github-client-secret",
  "oauth_scopes": ["read:user", "user:email"],
  "oauth_provider_name": "github",
  "oauth_redirect_url": "https://your-cms-domain.com/api/v1/auth/oauth/callback",
  "oauth_success_redirect": "/admin/",
  "oauth_endpoint": {
    "oauth_auth_url": "https://github.com/login/oauth/authorize",
    "oauth_token_url": "https://github.com/login/oauth/access_token",
    "oauth_userinfo_url": "https://api.github.com/user"
  }
}
```

> **Good to know**: GitHub's OAuth Apps use `read:user` and `user:email` scopes instead of OpenID Connect scopes. ModulaCMS handles both conventions.

## Set Up Azure AD

1. Go to the [Azure Portal > App registrations](https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps).
2. Register a new application.
3. Add the redirect URI `https://your-cms-domain.com/api/v1/auth/oauth/callback` under **Authentication**.
4. Create a client secret under **Certificates & secrets**.

```json
{
  "oauth_client_id": "your-azure-application-id",
  "oauth_client_secret": "your-azure-client-secret",
  "oauth_scopes": ["openid", "profile", "email"],
  "oauth_provider_name": "azure",
  "oauth_redirect_url": "https://your-cms-domain.com/api/v1/auth/oauth/callback",
  "oauth_success_redirect": "/admin/",
  "oauth_endpoint": {
    "oauth_auth_url": "https://login.microsoftonline.com/YOUR_TENANT_ID/oauth2/v2.0/authorize",
    "oauth_token_url": "https://login.microsoftonline.com/YOUR_TENANT_ID/oauth2/v2.0/token",
    "oauth_userinfo_url": "https://graph.microsoft.com/oidc/userinfo"
  }
}
```

Replace `YOUR_TENANT_ID` with your Azure AD tenant ID. Use `common` for multi-tenant apps.

## Use Any OpenID Connect Provider

ModulaCMS works with any provider that exposes standard OAuth 2.0 / OpenID Connect endpoints. Set the three `oauth_endpoint` URLs and adjust `oauth_scopes` as needed for your provider.

## Initiate the Login Flow

Redirect your frontend users to the OAuth login endpoint:

```bash
# Browser redirect or link
GET http://localhost:8080/api/v1/auth/oauth/login
```

This redirects to the provider's authorization page. After authentication, the user is redirected back through the callback endpoint and lands at `oauth_success_redirect` with an active session.

## Token Refresh

ModulaCMS refreshes OAuth tokens transparently during session validation. If a token refresh fails, the session remains valid -- the user is not logged out.

## Local Development

For local development, use `http://localhost:8080/api/v1/auth/oauth/callback` as the redirect URL. Most providers allow `localhost` redirect URIs for development apps.

```json
{
  "oauth_redirect_url": "http://localhost:8080/api/v1/auth/oauth/callback",
  "oauth_success_redirect": "/admin/"
}
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/oauth/login` | Initiate OAuth flow (redirects to provider) |
| GET | `/api/v1/auth/oauth/callback` | OAuth provider callback (handles token exchange) |

Both endpoints are public and rate-limited to 10 requests per minute per IP.

## Next Steps

- [Authentication guide](/docs/custom-admin/authentication) -- password login, API keys, RBAC
- [Email](/docs/integrations/email) -- configure email for password reset flows
- [Configuration reference](/docs/getting-started/configuration) -- all config fields
