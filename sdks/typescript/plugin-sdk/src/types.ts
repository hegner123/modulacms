import type { JsonValue } from './json.js'

export type PluginAuth = {
  userId: string
  username: string
  role: string
}

export type ThemeColors = {
  primary: string
  surface: string
  border: string
  text: string
  error: string
  success: string
  warning: string
}

export type ThemeSpacing = {
  xs: string
  sm: string
  md: string
  lg: string
  xl: string
  xl2: string
}

export type ThemeRadii = {
  sm: string
  md: string
  lg: string
  full: string
}

export type ThemeTypography = {
  fontFamily: string
  fontSizeSm: string
  fontSizeMd: string
  fontSizeLg: string
  lineHeight: string
}

export type ThemeTokens = {
  colors: ThemeColors
  spacing: ThemeSpacing
  radii: ThemeRadii
  typography: ThemeTypography
}

/**
 * Scoped API client for plugin HTTP routes.
 *
 * Generic parameter T on get/post/put/patch is a trust-based cast — the SDK does NOT validate
 * response shapes at runtime. Consumers are responsible for verifying that API responses
 * match T. For runtime validation, use raw() and validate the Response body yourself.
 */
export type PluginApiClient = {
  /** GET request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
  get: <T = unknown>(path: string, params?: Record<string, string>, signal?: AbortSignal) => Promise<T>
  /** POST request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
  post: <T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal) => Promise<T>
  /** PUT request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
  put: <T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal) => Promise<T>
  /** PATCH request. Throws PluginApiError on non-2xx. T is not runtime-validated. */
  patch: <T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal) => Promise<T>
  /** DELETE request. Throws PluginApiError on non-2xx. Returns void — error info only via thrown PluginApiError. */
  del: (path: string, params?: Record<string, string>, signal?: AbortSignal) => Promise<void>
  /** Raw fetch. Does NOT throw on non-2xx — caller must check response.ok. */
  raw: (path: string, init?: RequestInit) => Promise<Response>
}

export type PluginContext = {
  pluginName: string
  baseUrl: string
  api: PluginApiClient
  auth: PluginAuth | null
  theme: ThemeTokens
  onNavigate: (handler: (path: string) => void) => () => void
  navigate: (path: string) => void
  element: HTMLElement
  signal: AbortSignal
}

export type PluginDefinition = {
  tag: string
  setup: (ctx: PluginContext, el: HTMLElement) => void | Promise<void>
  mount: (ctx: PluginContext, el: HTMLElement) => void | Promise<void>
  destroy?: (el: HTMLElement) => void
}

export type PluginConfig = {
  apiKey: string
}
