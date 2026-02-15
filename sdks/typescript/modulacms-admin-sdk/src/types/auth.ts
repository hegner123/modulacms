/**
 * Authentication request and response types for login, session, and user identity.
 *
 * @module types/auth
 */

import type { Email, UserID } from './common.js'

/**
 * Credentials payload sent to `POST /auth/login`.
 */
export type LoginRequest = {
  /** The user's email address. */
  email: Email
  /** The user's plaintext password. Transmitted over HTTPS only (unless `allowInsecure` is set). */
  password: string
}

/**
 * Successful login response returned by the server.
 * Contains the authenticated user's basic identity.
 */
export type LoginResponse = {
  /** Unique identifier of the authenticated user. */
  user_id: UserID
  /** Email address of the authenticated user. */
  email: Email
  /** Display username of the authenticated user. */
  username: string
  /** ISO 8601 timestamp of when the user account was created. */
  created_at: string
}

/**
 * Response from `GET /auth/me` representing the currently authenticated user.
 */
export type MeResponse = {
  /** Unique identifier of the current user. */
  user_id: UserID
  /** Email address of the current user. */
  email: Email
  /** Username of the current user. */
  username: string
  /** Display name of the current user. */
  name: string
  /** Role label assigned to the current user. */
  role: string
}
