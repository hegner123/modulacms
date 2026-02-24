/**
 * Authentication resource providing login, logout, session identity,
 * user registration, and password reset endpoints.
 *
 * @module resources/auth
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions } from '../types/common.js'
import type { LoginRequest, LoginResponse, MeResponse, RequestPasswordResetParams, ConfirmPasswordResetParams, MessageResponse } from '../types/auth.js'
import type { User, CreateUserParams, UpdateUserParams } from '../types/users.js'

/**
 * Authentication operations available on `client.auth`.
 */
type AuthResource = {
  /**
   * Authenticate a user with email and password.
   * @param params - Login credentials.
   * @param opts - Optional request options.
   * @returns The authenticated user's identity.
   * @throws {@link import('../types/common.js').ApiError} on invalid credentials (401).
   */
  login: (params: LoginRequest, opts?: RequestOptions) => Promise<LoginResponse>

  /**
   * End the current authenticated session.
   * @param opts - Optional request options.
   */
  logout: (opts?: RequestOptions) => Promise<void>

  /**
   * Retrieve the currently authenticated user's identity.
   * @param opts - Optional request options.
   * @returns The current user's profile.
   * @throws {@link import('../types/common.js').ApiError} if not authenticated (401).
   */
  me: (opts?: RequestOptions) => Promise<MeResponse>

  /**
   * Register a new user account.
   * @param params - New user details.
   * @param opts - Optional request options.
   * @returns The created user entity.
   */
  register: (params: CreateUserParams, opts?: RequestOptions) => Promise<User>

  /**
   * Reset a user's password or account details.
   * @param params - Updated user details including the new password hash.
   * @param opts - Optional request options.
   * @returns A confirmation message string.
   * @deprecated Use {@link requestPasswordReset} and {@link confirmPasswordReset} instead.
   */
  reset: (params: UpdateUserParams, opts?: RequestOptions) => Promise<string>

  /**
   * Request a password reset email for the given address.
   * Always succeeds regardless of whether the email exists (prevents user enumeration).
   * @param params - The email address to send the reset link to.
   * @param opts - Optional request options.
   */
  requestPasswordReset: (params: RequestPasswordResetParams, opts?: RequestOptions) => Promise<MessageResponse>

  /**
   * Confirm a password reset using a token received via email.
   * @param params - The reset token and new password.
   * @param opts - Optional request options.
   */
  confirmPasswordReset: (params: ConfirmPasswordResetParams, opts?: RequestOptions) => Promise<MessageResponse>
}

/**
 * Create the authentication resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns An {@link AuthResource} with login, logout, me, register, and reset methods.
 * @internal
 */
function createAuthResource(http: HttpClient): AuthResource {
  return {
    login(params: LoginRequest, opts?: RequestOptions): Promise<LoginResponse> {
      return http.post<LoginResponse>('/auth/login', params as Record<string, unknown>, opts)
    },

    async logout(opts?: RequestOptions): Promise<void> {
      await http.post<Record<string, unknown>>('/auth/logout', undefined, opts)
    },

    me(opts?: RequestOptions): Promise<MeResponse> {
      return http.get<MeResponse>('/auth/me', undefined, opts)
    },

    register(params: CreateUserParams, opts?: RequestOptions): Promise<User> {
      return http.post<User>('/auth/register', params as Record<string, unknown>, opts)
    },

    async reset(params: UpdateUserParams, opts?: RequestOptions): Promise<string> {
      return http.post<string>('/auth/reset', params as Record<string, unknown>, opts)
    },

    requestPasswordReset(params: RequestPasswordResetParams, opts?: RequestOptions): Promise<MessageResponse> {
      return http.post<MessageResponse>('/auth/request-password-reset', params as Record<string, unknown>, opts)
    },

    confirmPasswordReset(params: ConfirmPasswordResetParams, opts?: RequestOptions): Promise<MessageResponse> {
      return http.post<MessageResponse>('/auth/confirm-password-reset', params as Record<string, unknown>, opts)
    },
  }
}

export type { AuthResource }
export { createAuthResource }
