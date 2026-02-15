/**
 * User, role, token, OAuth, session, and SSH key entity types
 * with their create/update parameter shapes.
 *
 * @remarks
 * {@link Token} and {@link UserOauth} contain sensitive credential data.
 * Consumers must treat these as secrets and never log or expose them.
 *
 * @module types/users
 */

import type { Email, RoleID, SessionID, UserID, UserOauthID } from './common.js'

// ---------------------------------------------------------------------------
// Entity types
// ---------------------------------------------------------------------------

/**
 * A registered user account.
 */
export type User = {
  /** Unique identifier for this user. */
  user_id: UserID
  /** Login username. */
  username: string
  /** Display name. */
  name: string
  /** Email address. */
  email: Email
  /** Password hash. Server-side only; included in the type for admin operations. */
  hash: string
  /** Role label assigned to this user. */
  role: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * A permission role that can be assigned to users.
 */
export type Role = {
  /** Unique identifier for this role. */
  role_id: RoleID
  /** Human-readable role name. */
  label: string
  /** JSON-encoded permissions map. */
  permissions: string
}

/**
 * An API token or refresh token issued to a user.
 *
 * @remarks SENSITIVE - the `token` field contains a bearer credential.
 * Never log or expose this value.
 */
export type Token = {
  /** Unique identifier for this token record. */
  id: string
  /** The user this token belongs to, or `null` for system tokens. */
  user_id: UserID | null
  /** Token category (e.g. `'access'`, `'refresh'`). */
  token_type: string
  /** The token value. SENSITIVE - treat as a secret. */
  token: string
  /** ISO 8601 timestamp when the token was issued. */
  issued_at: string
  /** ISO 8601 timestamp when the token expires. */
  expires_at: string
  /** Whether this token has been revoked. */
  revoked: boolean
}

/**
 * An OAuth connection linking a user to an external provider.
 *
 * @remarks SENSITIVE - contains `access_token` and `refresh_token`.
 * Never log or expose these values.
 */
export type UserOauth = {
  /** Unique identifier for this OAuth connection. */
  user_oauth_id: UserOauthID
  /** The local user linked to this OAuth connection, or `null`. */
  user_id: UserID | null
  /** OAuth provider name (e.g. `'google'`, `'github'`). */
  oauth_provider: string
  /** User ID on the OAuth provider's platform. */
  oauth_provider_user_id: string
  /** OAuth access token. SENSITIVE. */
  access_token: string
  /** OAuth refresh token. SENSITIVE. */
  refresh_token: string
  /** ISO 8601 timestamp when the access token expires. */
  token_expires_at: string
  /** ISO 8601 creation timestamp. */
  date_created: string
}

/**
 * An active user session.
 */
export type Session = {
  /** Unique identifier for this session. */
  session_id: SessionID
  /** The user this session belongs to, or `null`. */
  user_id: UserID | null
  /** ISO 8601 timestamp when the session was created. */
  created_at: string
  /** ISO 8601 timestamp when the session expires. */
  expires_at: string
  /** ISO 8601 timestamp of last access, or `null`. */
  last_access: string | null
  /** Client IP address, or `null`. */
  ip_address: string | null
  /** Client user-agent string, or `null`. */
  user_agent: string | null
  /** JSON-encoded session payload, or `null`. */
  session_data: string | null
}

/**
 * A full SSH key record including the public key material.
 * Returned when creating a new SSH key.
 */
export type SshKey = {
  /** Unique identifier for this SSH key. */
  ssh_key_id: string
  /** User this key belongs to, or `null`. */
  user_id: string | null
  /** The public key material (e.g. `ssh-ed25519 AAAA...`). */
  public_key: string
  /** Key algorithm (e.g. `'ssh-ed25519'`, `'ssh-rsa'`). */
  key_type: string
  /** Key fingerprint (e.g. `SHA256:...`). */
  fingerprint: string
  /** Human-readable label for this key. */
  label: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 timestamp of last use. */
  last_used: string
}

/**
 * Summary SSH key record returned by list operations.
 * Omits the `public_key` field for security.
 */
export type SshKeyListItem = {
  /** Unique identifier for this SSH key. */
  ssh_key_id: string
  /** Key algorithm. */
  key_type: string
  /** Key fingerprint. */
  fingerprint: string
  /** Human-readable label. */
  label: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 timestamp of last use. */
  last_used: string
}

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for registering a new user via `POST /auth/register`. */
export type CreateUserParams = {
  /** Login username. */
  username: string
  /** Display name. */
  name: string
  /** Email address. */
  email: Email
  /** Password hash. */
  hash: string
  /** Role label to assign. */
  role: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for creating a new role via `POST /roles`. */
export type CreateRoleParams = {
  /** Human-readable role name. */
  label: string
  /** JSON-encoded permissions map. */
  permissions: string
}

/**
 * Parameters for creating a new token via `POST /tokens`.
 *
 * @remarks The `token` field is SENSITIVE.
 */
export type CreateTokenParams = {
  /** User this token belongs to, or `null`. */
  user_id: UserID | null
  /** Token category. */
  token_type: string
  /** The token value. SENSITIVE. */
  token: string
  /** ISO 8601 issue timestamp. */
  issued_at: string
  /** ISO 8601 expiration timestamp. */
  expires_at: string
  /** Whether the token is created in a revoked state. */
  revoked: boolean
}

/**
 * Parameters for creating a new OAuth connection via `POST /usersoauth`.
 *
 * @remarks Contains SENSITIVE token fields.
 */
export type CreateUserOauthParams = {
  /** Local user to link, or `null`. */
  user_id: UserID | null
  /** OAuth provider name. */
  oauth_provider: string
  /** User ID on the provider's platform. */
  oauth_provider_user_id: string
  /** OAuth access token. SENSITIVE. */
  access_token: string
  /** OAuth refresh token. SENSITIVE. */
  refresh_token: string
  /** ISO 8601 token expiration timestamp. */
  token_expires_at: string
  /** ISO 8601 creation timestamp. */
  date_created: string
}

/** Parameters for registering a new SSH key via `POST /ssh-keys`. */
export type CreateSshKeyRequest = {
  /** The public key material to register. */
  public_key: string
  /** Human-readable label for the key. */
  label: string
}

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/** Parameters for updating a user via `PUT /users/`. */
export type UpdateUserParams = {
  /** ID of the user to update. */
  user_id: UserID
  /** Updated username. */
  username: string
  /** Updated display name. */
  name: string
  /** Updated email. */
  email: Email
  /** Updated password hash. */
  hash: string
  /** Updated role label. */
  role: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for updating a role via `PUT /roles/`. */
export type UpdateRoleParams = {
  /** ID of the role to update. */
  role_id: RoleID
  /** Updated label. */
  label: string
  /** Updated permissions (JSON-encoded). */
  permissions: string
}

/**
 * Parameters for updating a token via `PUT /tokens/`.
 *
 * @remarks The `token` field is SENSITIVE.
 */
export type UpdateTokenParams = {
  /** Updated token value. SENSITIVE. */
  token: string
  /** Updated issue timestamp. */
  issued_at: string
  /** Updated expiration timestamp. */
  expires_at: string
  /** Updated revocation status. */
  revoked: boolean
  /** ID of the token to update. */
  id: string
}

/**
 * Parameters for updating an OAuth connection via `PUT /usersoauth/`.
 *
 * @remarks Contains SENSITIVE token fields.
 */
export type UpdateUserOauthParams = {
  /** Updated access token. SENSITIVE. */
  access_token: string
  /** Updated refresh token. SENSITIVE. */
  refresh_token: string
  /** Updated token expiration. */
  token_expires_at: string
  /** ID of the OAuth connection to update. */
  user_oauth_id: UserOauthID
}

/** Parameters for updating a session via `PUT /sessions/`. */
export type UpdateSessionParams = {
  /** ID of the session to update. */
  session_id: SessionID
  /** Updated user, or `null`. */
  user_id: UserID | null
  /** Updated creation timestamp. */
  created_at: string
  /** Updated expiration timestamp. */
  expires_at: string
  /** Updated last-access timestamp, or `null`. */
  last_access: string | null
  /** Updated IP address, or `null`. */
  ip_address: string | null
  /** Updated user-agent, or `null`. */
  user_agent: string | null
  /** Updated session data (JSON-encoded), or `null`. */
  session_data: string | null
}
