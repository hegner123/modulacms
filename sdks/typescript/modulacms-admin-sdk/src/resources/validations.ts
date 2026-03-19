/**
 * Validations resource providing CRUD operations for validation config management
 * plus search-by-name for both public and admin validations.
 *
 * @module resources/validations
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, ValidationID, AdminValidationID } from '../types/common.js'
import type { Validation, AdminValidation } from '@modulacms/types'

// ---------------------------------------------------------------------------
// Request/Response types
// ---------------------------------------------------------------------------

/** Parameters for creating a new validation via `POST /validations`. */
export type CreateValidationParams = {
  /** Human-readable name. */
  name: string
  /** Description of what this validation does. */
  description?: string
  /** JSON-encoded validation rules configuration. */
  config: string
}

/** Parameters for updating a validation via `PUT /validations/{id}`. */
export type UpdateValidationParams = {
  /** ID of the validation to update. */
  validation_id: ValidationID
  /** Updated name. */
  name: string
  /** Updated description. */
  description?: string
  /** Updated JSON-encoded validation rules configuration. */
  config: string
}

/** Parameters for creating a new admin validation via `POST /admin/validations`. */
export type CreateAdminValidationParams = {
  /** Human-readable name. */
  name: string
  /** Description of what this validation does. */
  description?: string
  /** JSON-encoded validation rules configuration. */
  config: string
}

/** Parameters for updating an admin validation via `PUT /admin/validations/{id}`. */
export type UpdateAdminValidationParams = {
  /** ID of the admin validation to update. */
  admin_validation_id: AdminValidationID
  /** Updated name. */
  name: string
  /** Updated description. */
  description?: string
  /** Updated JSON-encoded validation rules configuration. */
  config: string
}

// ---------------------------------------------------------------------------
// Validations resource type
// ---------------------------------------------------------------------------

/** Public validation config management operations available on `client.validations`. */
export type ValidationsResource = {
  /** List all public validations. */
  list: (opts?: RequestOptions) => Promise<Validation[]>
  /** Get a single public validation by ID. */
  get: (id: ValidationID, opts?: RequestOptions) => Promise<Validation>
  /** Create a new public validation. */
  create: (params: CreateValidationParams, opts?: RequestOptions) => Promise<Validation>
  /** Update an existing public validation. */
  update: (params: UpdateValidationParams, opts?: RequestOptions) => Promise<Validation>
  /** Remove a public validation by ID. */
  remove: (id: ValidationID, opts?: RequestOptions) => Promise<void>
  /** Search public validations by name. */
  search: (name: string, opts?: RequestOptions) => Promise<Validation[]>
}

/** Admin validation config management operations available on `client.adminValidations`. */
export type AdminValidationsResource = {
  /** List all admin validations. */
  list: (opts?: RequestOptions) => Promise<AdminValidation[]>
  /** Get a single admin validation by ID. */
  get: (id: AdminValidationID, opts?: RequestOptions) => Promise<AdminValidation>
  /** Create a new admin validation. */
  create: (params: CreateAdminValidationParams, opts?: RequestOptions) => Promise<AdminValidation>
  /** Update an existing admin validation. */
  update: (params: UpdateAdminValidationParams, opts?: RequestOptions) => Promise<AdminValidation>
  /** Remove an admin validation by ID. */
  remove: (id: AdminValidationID, opts?: RequestOptions) => Promise<void>
  /** Search admin validations by name. */
  search: (name: string, opts?: RequestOptions) => Promise<AdminValidation[]>
}

// ---------------------------------------------------------------------------
// Factories
// ---------------------------------------------------------------------------

/**
 * Create the public validations resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link ValidationsResource} with CRUD and search operations.
 * @internal
 */
function createValidationsResource(http: HttpClient): ValidationsResource {
  return {
    list(opts?: RequestOptions): Promise<Validation[]> {
      return http.get<Validation[]>('/validations', undefined, opts)
    },

    get(id: ValidationID, opts?: RequestOptions): Promise<Validation> {
      return http.get<Validation>(`/validations/${String(id)}`, undefined, opts)
    },

    create(params: CreateValidationParams, opts?: RequestOptions): Promise<Validation> {
      return http.post<Validation>('/validations', params as unknown as Record<string, unknown>, opts)
    },

    update(params: UpdateValidationParams, opts?: RequestOptions): Promise<Validation> {
      return http.put<Validation>(`/validations/${String(params.validation_id)}`, params as unknown as Record<string, unknown>, opts)
    },

    remove(id: ValidationID, opts?: RequestOptions): Promise<void> {
      return http.del(`/validations/${String(id)}`, undefined, opts)
    },

    search(name: string, opts?: RequestOptions): Promise<Validation[]> {
      return http.get<Validation[]>('/validations/search', { name }, opts)
    },
  }
}

/**
 * Create the admin validations resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns An {@link AdminValidationsResource} with CRUD and search operations.
 * @internal
 */
function createAdminValidationsResource(http: HttpClient): AdminValidationsResource {
  return {
    list(opts?: RequestOptions): Promise<AdminValidation[]> {
      return http.get<AdminValidation[]>('/admin/validations', undefined, opts)
    },

    get(id: AdminValidationID, opts?: RequestOptions): Promise<AdminValidation> {
      return http.get<AdminValidation>(`/admin/validations/${String(id)}`, undefined, opts)
    },

    create(params: CreateAdminValidationParams, opts?: RequestOptions): Promise<AdminValidation> {
      return http.post<AdminValidation>('/admin/validations', params as unknown as Record<string, unknown>, opts)
    },

    update(params: UpdateAdminValidationParams, opts?: RequestOptions): Promise<AdminValidation> {
      return http.put<AdminValidation>(`/admin/validations/${String(params.admin_validation_id)}`, params as unknown as Record<string, unknown>, opts)
    },

    remove(id: AdminValidationID, opts?: RequestOptions): Promise<void> {
      return http.del(`/admin/validations/${String(id)}`, undefined, opts)
    },

    search(name: string, opts?: RequestOptions): Promise<AdminValidation[]> {
      return http.get<AdminValidation[]>('/admin/validations/search', { name }, opts)
    },
  }
}

export { createValidationsResource, createAdminValidationsResource }
