/**
 * Internal HTTP client layer built on the Fetch API.
 * Handles authentication, base URL prefixing, timeout merging,
 * JSON parsing, and structured error creation.
 *
 * @remarks This module is internal to the SDK and not part of the public API.
 *
 * @module http
 * @internal
 */

import type { ApiError, RequestOptions } from './types/common.js'

// ---------------------------------------------------------------------------
// Internal types (not exported in public API)
// ---------------------------------------------------------------------------

/**
 * Configuration required to create an {@link HttpClient}.
 * @internal
 */
type HttpClientConfig = {
  /** Fully qualified origin (e.g. `https://cms.example.com`). No trailing slash. */
  baseUrl: string
  /** Bearer token for the `Authorization` header. Omit for cookie-only auth. */
  apiKey?: string
  /** Default request timeout in milliseconds. Applied via `AbortSignal.timeout()`. */
  defaultTimeout: number
  /** Fetch `credentials` mode (e.g. `'include'`, `'same-origin'`). */
  credentials: RequestCredentials
}

/**
 * Low-level HTTP methods used internally by resource factories.
 * All paths are automatically prefixed with `/api/v1`.
 * @internal
 */
type HttpClient = {
  /**
   * Perform a GET request and parse the JSON response.
   * @typeParam T - Expected response body type.
   * @param path - API path (e.g. `'/users'`).
   * @param params - Optional query parameters appended as `?key=value`.
   * @param opts - Per-request options (abort signal).
   * @returns Parsed JSON response body.
   * @throws {@link ApiError} on non-2xx or non-JSON responses.
   * @throws `TypeError` on network failure.
   */
  get: <T>(path: string, params?: Record<string, string>, opts?: RequestOptions) => Promise<T>

  /**
   * Perform a POST request with a JSON body and parse the response.
   * @typeParam T - Expected response body type.
   * @param path - API path.
   * @param body - JSON-serializable request body.
   * @param opts - Per-request options.
   * @returns Parsed JSON response body.
   * @throws {@link ApiError} on non-2xx or non-JSON responses.
   * @throws `TypeError` on network failure.
   */
  post: <T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions) => Promise<T>

  /**
   * Perform a PUT request with a JSON body and parse the response.
   * @typeParam T - Expected response body type.
   * @param path - API path.
   * @param body - JSON-serializable request body.
   * @param opts - Per-request options.
   * @returns Parsed JSON response body.
   * @throws {@link ApiError} on non-2xx or non-JSON responses.
   * @throws `TypeError` on network failure.
   */
  put: <T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions) => Promise<T>

  /**
   * Perform a PATCH request with a JSON body and parse the response.
   * @typeParam T - Expected response body type.
   * @param path - API path.
   * @param body - JSON-serializable request body.
   * @param opts - Per-request options.
   * @returns Parsed JSON response body.
   * @throws {@link ApiError} on non-2xx or non-JSON responses.
   * @throws `TypeError` on network failure.
   */
  patch: <T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions) => Promise<T>

  /**
   * Perform a DELETE request. Returns void on success.
   * @param path - API path.
   * @param params - Optional query parameters.
   * @param opts - Per-request options.
   * @throws {@link ApiError} on non-2xx responses.
   * @throws `TypeError` on network failure.
   */
  del: (path: string, params?: Record<string, string>, opts?: RequestOptions) => Promise<void>

  /**
   * Perform a raw fetch request with full control over `RequestInit`.
   * Used for non-JSON payloads like multipart file uploads.
   * @param path - API path (prefixed with base URL and `/api/v1`).
   * @param init - Standard `RequestInit` options passed directly to `fetch()`.
   * @returns The raw `Response` object.
   */
  raw: (path: string, init: RequestInit) => Promise<Response>
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Build a full URL by joining the base URL, API version prefix, path, and optional query params.
 * @param baseUrl - Origin without trailing slash.
 * @param path - API path (e.g. `'/users'`).
 * @param params - Optional query parameters.
 * @returns Fully qualified URL string.
 * @internal
 */
function buildUrl(baseUrl: string, path: string, params?: Record<string, string>): string {
  const url = baseUrl + '/api/v1' + path
  if (!params) return url
  const search = new URLSearchParams(params)
  const separator = url.includes('?') ? '&' : '?'
  return url + separator + search.toString()
}

/**
 * Merge the SDK's default timeout signal with an optional user-provided abort signal.
 * Either signal aborting will cancel the request.
 * @param defaultTimeout - Timeout in milliseconds.
 * @param opts - Optional request options containing a user abort signal.
 * @returns A combined `AbortSignal`.
 * @internal
 */
function mergeSignals(defaultTimeout: number, opts?: RequestOptions): AbortSignal {
  const timeoutSignal = AbortSignal.timeout(defaultTimeout)
  if (!opts?.signal) return timeoutSignal

  const controller = new AbortController()

  const onAbort = (source: AbortSignal) => () => {
    if (!controller.signal.aborted) controller.abort(source.reason)
  }

  timeoutSignal.addEventListener('abort', onAbort(timeoutSignal), { once: true })
  opts.signal.addEventListener('abort', onAbort(opts.signal), { once: true })

  // If the user signal is already aborted, propagate immediately
  if (opts.signal.aborted) controller.abort(opts.signal.reason)

  return controller.signal
}

/**
 * Construct a tagged {@link ApiError} object.
 * @param status - HTTP status code.
 * @param message - HTTP status text.
 * @param body - Parsed JSON response body, if available.
 * @returns A new `ApiError` with `_tag: 'ApiError'`.
 * @internal
 */
function createApiError(status: number, message: string, body?: unknown): ApiError {
  return {
    _tag: 'ApiError' as const,
    status,
    message,
    body,
  }
}

/**
 * Check whether a response has a JSON content type.
 * @param response - The fetch `Response` to inspect.
 * @returns `true` if the `Content-Type` header includes `application/json`.
 * @internal
 */
function isJsonResponse(response: Response): boolean {
  const ct = response.headers.get('content-type')
  return ct !== null && ct.includes('application/json')
}

/**
 * Handle a fetch response by validating status and parsing JSON.
 * @typeParam T - Expected parsed response type.
 * @param response - The fetch `Response`.
 * @returns Parsed JSON body.
 * @throws {@link ApiError} on non-2xx status or when response is not JSON.
 * @internal
 */
async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    if (isJsonResponse(response)) {
      const body: unknown = await response.json()
      throw createApiError(response.status, response.statusText, body)
    }
    throw createApiError(response.status, response.statusText, undefined)
  }

  if (!isJsonResponse(response)) {
    throw createApiError(response.status, response.statusText, undefined)
  }

  return response.json() as Promise<T>
}

/**
 * Handle a fetch response that is expected to return no body (e.g. DELETE).
 * @param response - The fetch `Response`.
 * @throws {@link ApiError} on non-2xx status.
 * @internal
 */
async function handleVoidResponse(response: Response): Promise<void> {
  if (!response.ok) {
    if (isJsonResponse(response)) {
      const body: unknown = await response.json()
      throw createApiError(response.status, response.statusText, body)
    }
    throw createApiError(response.status, response.statusText, undefined)
  }
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create an {@link HttpClient} instance configured with authentication,
 * base URL, timeout, and credential settings.
 *
 * @param config - HTTP client configuration.
 * @returns A configured `HttpClient` with `get`, `post`, `put`, `del`, and `raw` methods.
 * @internal
 */
function createHttpClient(config: HttpClientConfig): HttpClient {
  const { baseUrl, apiKey, defaultTimeout, credentials } = config

  /**
   * Build default headers including Content-Type and optional Authorization.
   * @returns Header record.
   */
  function headers(): Record<string, string> {
    const h: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    if (apiKey) {
      h['Authorization'] = `Bearer ${apiKey}`
    }
    return h
  }

  return {
    async get<T>(path: string, params?: Record<string, string>, opts?: RequestOptions): Promise<T> {
      const url = buildUrl(baseUrl, path, params)
      const response = await fetch(url, {
        method: 'GET',
        headers: headers(),
        credentials,
        signal: mergeSignals(defaultTimeout, opts),
      })
      return handleResponse<T>(response)
    },

    async post<T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions): Promise<T> {
      const url = buildUrl(baseUrl, path)
      const response = await fetch(url, {
        method: 'POST',
        headers: headers(),
        credentials,
        signal: mergeSignals(defaultTimeout, opts),
        ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
      })
      return handleResponse<T>(response)
    },

    async put<T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions): Promise<T> {
      const url = buildUrl(baseUrl, path)
      const response = await fetch(url, {
        method: 'PUT',
        headers: headers(),
        credentials,
        signal: mergeSignals(defaultTimeout, opts),
        ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
      })
      return handleResponse<T>(response)
    },

    async patch<T>(path: string, body?: Record<string, unknown>, opts?: RequestOptions): Promise<T> {
      const url = buildUrl(baseUrl, path)
      const response = await fetch(url, {
        method: 'PATCH',
        headers: headers(),
        credentials,
        signal: mergeSignals(defaultTimeout, opts),
        ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
      })
      return handleResponse<T>(response)
    },

    async del(path: string, params?: Record<string, string>, opts?: RequestOptions): Promise<void> {
      const url = buildUrl(baseUrl, path, params)
      const response = await fetch(url, {
        method: 'DELETE',
        headers: headers(),
        credentials,
        signal: mergeSignals(defaultTimeout, opts),
      })
      return handleVoidResponse(response)
    },

    async raw(path: string, init: RequestInit): Promise<Response> {
      const url = baseUrl + '/api/v1' + path
      return fetch(url, init)
    },
  }
}

export type { HttpClient, HttpClientConfig }
export { createHttpClient }
