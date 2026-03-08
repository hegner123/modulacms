/**
 * User, role, token, OAuth, session, and SSH key entity types
 * with their create/update parameter shapes.
 *
 * @remarks
 * **Credential sensitivity:** Several types in this module contain sensitive
 * credential data that must be handled carefully:
 *
 * - {@link Token} -- contains bearer tokens (`token` field)
 * - {@link UserOauth} -- contains OAuth `access_token` and `refresh_token`
 * - {@link User} -- contains the password `hash` field
 * - {@link CreateUserParams} -- contains the plaintext `password`
 * - {@link CreateTokenParams} / {@link UpdateTokenParams} -- contain token values
 * - {@link CreateUserOauthParams} / {@link UpdateUserOauthParams} -- contain OAuth tokens
 *
 * **Safe error handling patterns:**
 * - When logging API errors from user/token/OAuth endpoints, log the error message
 *   and status code but never the request body or response body
 * - Use the `*View` types ({@link TokenView}, {@link UserOauthView}, {@link SessionView})
 *   for display and logging -- these omit sensitive fields by design
 * - Never include token values, password hashes, or OAuth tokens in error reports,
 *   analytics events, or client-side storage
 * - When serializing {@link UserFullView} for debugging, note that it uses safe
 *   view sub-types and is safe to log
 *
 * @module types/users
 */

import type { Email, PermissionID, RoleID, RolePermissionID, SessionID, UserID, UserOauthID } from './common.js'

// ---------------------------------------------------------------------------
// Entity types
// ---------------------------------------------------------------------------

/**
 * A registered user account.
 *
 * @remarks
 * SENSITIVE -- the `hash` field contains the bcrypt password hash.
 * This type is used for admin CRUD operations where the full record is needed.
 * For display purposes, prefer {@link UserWithRoleLabel} or {@link UserFullView}
 * which omit the hash.
 *
 * Never log, serialize to analytics, or expose the `hash` field to client-side code.
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
  /**
   * Bcrypt password hash. SENSITIVE -- server-side only.
   *
   * Included in the type for admin operations that need the full record.
   * Never log or expose this value.
   */
  hash: string
  /** Role label assigned to this user (e.g. `'admin'`, `'editor'`, `'viewer'`). */
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
}

/**
 * A permission entity with access control information.
 */
export type Permission = {
  /** Unique identifier for this permission. */
  permission_id: PermissionID
  /** Human-readable permission name. */
  label: string
}

/**
 * A junction between a role and a permission.
 */
export type RolePermission = {
  /** Unique identifier for this role-permission association. */
  id: RolePermissionID
  /** The role in this association. */
  role_id: RoleID
  /** The permission in this association. */
  permission_id: PermissionID
}

/**
 * An API token or refresh token issued to a user.
 *
 * @remarks
 * SENSITIVE -- the `token` field contains a bearer credential that grants
 * authenticated access to the API. Treat this entire type as sensitive.
 *
 * For safe display (e.g. listing tokens in a management UI), use {@link TokenView}
 * which omits the `token` field entirely.
 *
 * **Do not:**
 * - Log the `token` field or any object containing it
 * - Store in browser localStorage/sessionStorage
 * - Include in error reports or analytics
 * - Transmit over unencrypted channels
 */
export type Token = {
  /** Unique identifier for this token record. */
  id: string
  /** The user this token belongs to, or `null` for system tokens. */
  user_id: UserID | null
  /** Token category (e.g. `'access'`, `'refresh'`). */
  token_type: string
  /** The bearer token value. SENSITIVE -- treat as a secret. */
  token: string
  /** ISO 8601 timestamp when the token was issued. */
  issued_at: string
  /** ISO 8601 timestamp when the token expires. */
  expires_at: string
  /** Whether this token has been revoked. Revoked tokens are rejected by the server. */
  revoked: boolean
}

/**
 * An OAuth connection linking a user to an external provider.
 *
 * @remarks
 * SENSITIVE -- contains `access_token` and `refresh_token` that grant
 * access to the user's account on the OAuth provider's platform.
 *
 * For safe display (e.g. showing connected providers in a user profile),
 * use {@link UserOauthView} which omits both token fields.
 *
 * **Do not:**
 * - Log the `access_token` or `refresh_token` fields
 * - Store this type in client-side storage
 * - Include in error reports or diagnostics
 */
export type UserOauth = {
  /** Unique identifier for this OAuth connection. */
  user_oauth_id: UserOauthID
  /** The local user linked to this OAuth connection, or `null` if pending linkage. */
  user_id: UserID | null
  /** OAuth provider name (e.g. `'google'`, `'github'`, `'azure'`). */
  oauth_provider: string
  /** User ID on the OAuth provider's platform. */
  oauth_provider_user_id: string
  /** OAuth access token. SENSITIVE -- never log or expose. */
  access_token: string
  /** OAuth refresh token. SENSITIVE -- never log or expose. */
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
// View types (composed responses from /users/full endpoints)
//
// These types intentionally omit sensitive fields (password hashes, tokens,
// session data, public key material) and are safe to log, display, and
// transmit to client-side code.
// ---------------------------------------------------------------------------

/** A user row joined with the role label, returned by `GET /users/full`. */
export type UserWithRoleLabel = {
  /** Unique identifier for this user. */
  user_id: UserID
  /** Login username. */
  username: string
  /** Display name. */
  name: string
  /** Email address. */
  email: Email
  /** Role identifier. */
  role: string
  /** Human-readable role label. */
  role_label: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * Safe subset of an OAuth connection (excludes `access_token` and `refresh_token`).
 * Safe to log and display in management UIs.
 */
export type UserOauthView = {
  /** Unique identifier for this OAuth connection. */
  user_oauth_id: UserOauthID
  /** OAuth provider name. */
  oauth_provider: string
  /** User ID on the provider's platform. */
  oauth_provider_user_id: string
  /** ISO 8601 timestamp when the access token expires. */
  token_expires_at: string
  /** ISO 8601 creation timestamp. */
  date_created: string
}

/**
 * Safe subset of an SSH key (excludes `public_key` material).
 * Safe to log and display in key management UIs.
 */
export type UserSshKeyView = {
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

/**
 * Safe subset of a session (excludes `session_data` payload).
 * Safe to log and display in session management UIs.
 */
export type SessionView = {
  /** Unique identifier for this session. */
  session_id: SessionID
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 expiration timestamp. */
  expires_at: string
  /** ISO 8601 timestamp of last access. */
  last_access: string
  /** Client IP address. */
  ip_address: string
  /** Client user-agent string. */
  user_agent: string
}

/**
 * Safe subset of a token (excludes the bearer `token` value).
 * Safe to log and display in token management UIs.
 */
export type TokenView = {
  /** Unique identifier for this token record. */
  id: string
  /** Token category. */
  token_type: string
  /** ISO 8601 issue timestamp. */
  issued_at: string
  /** ISO 8601 expiration timestamp. */
  expires_at: string
  /** Whether this token has been revoked. */
  revoked: boolean
}

/**
 * A fully composed user response from `GET /users/full/{id}`.
 *
 * Aggregates user info with role label, OAuth connections, SSH keys, sessions,
 * and tokens -- all using safe view sub-types that omit sensitive fields.
 * This type is safe to log and display.
 */
export type UserFullView = {
  /** Unique identifier for this user. */
  user_id: UserID
  /** Login username. */
  username: string
  /** Display name. */
  name: string
  /** Email address. */
  email: Email
  /** Role identifier. */
  role_id: string
  /** Human-readable role label. */
  role_label: string
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
  /** OAuth connection, or absent if none. */
  oauth?: UserOauthView
  /** SSH keys registered to this user. */
  ssh_keys: UserSshKeyView[]
  /** Active session, or absent if none. */
  sessions?: SessionView
  /** API tokens issued to this user. */
  tokens: TokenView[]
}

// ---------------------------------------------------------------------------
// User reassign-delete (POST /users/reassign-delete)
// ---------------------------------------------------------------------------

/**
 * Parameters for reassigning a user's owned content and then deleting the user
 * via `POST /users/reassign-delete`.
 *
 * @remarks
 * This is a destructive operation. The user account, all associated sessions,
 * tokens, and OAuth connections are permanently deleted. Content authored by the
 * user is reassigned to the target user (or orphaned if `reassign_to` is omitted).
 *
 * This operation is audited. Only admin users can perform reassign-delete.
 */
export type UserReassignDeleteParams = {
  /** ID of the user to delete. Cannot be the requesting user's own ID. */
  user_id: UserID
  /** ID of the user to reassign owned content to. If omitted, content `author_id` is set to `null`. */
  reassign_to?: UserID
}

/** Response from the user reassign-delete endpoint. */
export type UserReassignDeleteResponse = {
  /** ID of the user that was deleted. */
  deleted_user_id: string
  /** ID of the user content was reassigned to. */
  reassigned_to: string
  /** Number of content_data rows reassigned. */
  content_data_reassigned: number
  /** Number of datatypes reassigned. */
  datatypes_reassigned: number
  /** Number of admin content_data rows reassigned. */
  admin_content_data_reassigned: number
}

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/**
 * Parameters for registering a new user via `POST /auth/register`.
 *
 * @remarks
 * SENSITIVE -- the `password` field contains the plaintext password.
 * The server hashes it with bcrypt before storage. Never log the request
 * body or persist this type after the API call completes.
 *
 * Non-admin callers cannot set the `role` field; it defaults to `'viewer'`.
 * Only admin users can assign roles other than `'viewer'` during registration.
 */
export type CreateUserParams = {
  /** Login username. Must be unique across all users. */
  username: string
  /** Display name. */
  name: string
  /** Email address. Must be unique across all users. */
  email: Email
  /** Plaintext password. SENSITIVE -- hashed server-side with bcrypt. Never log. */
  password: string
  /** Role label to assign (e.g. `'admin'`, `'editor'`, `'viewer'`). Non-admins default to `'viewer'`. */
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
}

/** Parameters for creating a new permission via `POST /permissions`. */
export type CreatePermissionParams = {
  /** Human-readable permission name. */
  label: string
}

/** Parameters for creating a new role-permission association via `POST /role-permissions`. */
export type CreateRolePermissionParams = {
  /** The role to associate. */
  role_id: RoleID
  /** The permission to associate. */
  permission_id: PermissionID
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

/**
 * Parameters for updating a user via `PUT /users/`.
 *
 * This is a full replacement update -- all fields except `password` must be provided.
 *
 * @remarks
 * The `password` field is optional. When omitted or empty string, the existing
 * password hash is preserved. When provided, the server re-hashes the new value
 * with bcrypt.
 *
 * Only admin users can change the `role` field on other users.
 */
export type UpdateUserParams = {
  /** ID of the user to update. */
  user_id: UserID
  /** Updated username. Must remain unique across all users. */
  username: string
  /** Updated display name. */
  name: string
  /** Updated email. Must remain unique across all users. */
  email: Email
  /** New plaintext password. SENSITIVE. Omit or empty string to keep existing password. */
  password?: string
  /** Updated role label. Only admin users can change roles. */
  role: string
  /** ISO 8601 creation timestamp (preserved from original record). */
  date_created: string
  /** ISO 8601 modification timestamp (set to current time). */
  date_modified: string
}

/** Parameters for updating a role via `PUT /roles/`. */
export type UpdateRoleParams = {
  /** ID of the role to update. */
  role_id: RoleID
  /** Updated label. */
  label: string
}

/** Parameters for updating a permission via `PUT /permissions/`. */
export type UpdatePermissionParams = {
  /** ID of the permission to update. */
  permission_id: PermissionID
  /** Updated label. */
  label: string
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
