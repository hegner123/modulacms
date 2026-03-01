/**
 * Locales resource providing CRUD operations for locale management
 * and translation creation for i18n content.
 *
 * @module resources/locales
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, LocaleID } from '../types/common.js'
import type { Locale } from '@modulacms/types'

// ---------------------------------------------------------------------------
// Request/Response types
// ---------------------------------------------------------------------------

/** Parameters for creating a new locale via `POST /locales`. */
export type CreateLocaleParams = {
  /** BCP 47 language code (e.g. `"en"`, `"fr"`). */
  code: string
  /** Human-readable label (e.g. `"English"`, `"French"`). */
  label: string
  /** Whether this is the default locale. */
  is_default: boolean
  /** Whether this locale is currently enabled. */
  is_enabled: boolean
  /** BCP 47 code of the fallback locale, or empty string. */
  fallback_code?: string
  /** Display ordering position. */
  sort_order: number
}

/** Parameters for updating a locale via `PUT /locales/`. */
export type UpdateLocaleParams = {
  /** ID of the locale to update. */
  locale_id: LocaleID
  /** Updated BCP 47 language code. */
  code: string
  /** Updated human-readable label. */
  label: string
  /** Whether this is the default locale. */
  is_default: boolean
  /** Whether this locale is enabled. */
  is_enabled: boolean
  /** Updated fallback locale code, or empty string. */
  fallback_code?: string
  /** Updated display ordering position. */
  sort_order: number
}

/** Request body for creating translations for a content item. */
export type CreateTranslationRequest = {
  /** Locale code to create translations for. */
  locale: string
}

/** Response from a translation creation operation. */
export type CreateTranslationResponse = {
  /** Locale code of the created translations. */
  locale: string
  /** Number of content fields created. */
  fields_created: number
}

// ---------------------------------------------------------------------------
// Locales resource type
// ---------------------------------------------------------------------------

/** Locale management operations available on `client.locales`. */
export type LocalesResource = {
  /** List all locales. */
  list: (opts?: RequestOptions) => Promise<Locale[]>
  /** Get a single locale by ID. */
  get: (id: LocaleID, opts?: RequestOptions) => Promise<Locale>
  /** Create a new locale. */
  create: (params: CreateLocaleParams, opts?: RequestOptions) => Promise<Locale>
  /** Update an existing locale. */
  update: (params: UpdateLocaleParams, opts?: RequestOptions) => Promise<Locale>
  /** Remove a locale by ID. */
  remove: (id: LocaleID, opts?: RequestOptions) => Promise<void>
  /** Create translations for a content data node in a given locale. */
  createTranslation: (contentDataID: string, req: CreateTranslationRequest, opts?: RequestOptions) => Promise<CreateTranslationResponse>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the locales resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link LocalesResource} with CRUD and translation operations.
 * @internal
 */
function createLocalesResource(http: HttpClient): LocalesResource {
  return {
    list(opts?: RequestOptions): Promise<Locale[]> {
      return http.get<Locale[]>('/locales', undefined, opts)
    },

    get(id: LocaleID, opts?: RequestOptions): Promise<Locale> {
      return http.get<Locale>('/locales/', { q: String(id) }, opts)
    },

    create(params: CreateLocaleParams, opts?: RequestOptions): Promise<Locale> {
      return http.post<Locale>('/locales', params as unknown as Record<string, unknown>, opts)
    },

    update(params: UpdateLocaleParams, opts?: RequestOptions): Promise<Locale> {
      return http.put<Locale>('/locales/', params as unknown as Record<string, unknown>, opts)
    },

    remove(id: LocaleID, opts?: RequestOptions): Promise<void> {
      return http.del('/locales/', { q: String(id) }, opts)
    },

    createTranslation(contentDataID: string, req: CreateTranslationRequest, opts?: RequestOptions): Promise<CreateTranslationResponse> {
      return http.post<CreateTranslationResponse>(`/admin/contentdata/${contentDataID}/translations`, req as unknown as Record<string, unknown>, opts)
    },
  }
}

export { createLocalesResource }
