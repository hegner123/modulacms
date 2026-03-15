import type { PluginAuth, PluginApiClient, PluginContext, ThemeTokens } from './types.js'
import type { JsonValue } from './json.js'
import { PluginError, PluginApiError } from './errors.js'
import { getApiKey } from './configure.js'
import { DEFAULT_THEME, parseThemeOverrides } from './theme.js'

function normalizePath(path: string): string {
  if (path.length > 0 && path.charCodeAt(0) === 47) { // '/'
    return path.slice(1)
  }
  return path
}

function buildUrl(baseUrl: string, pluginName: string, path: string, params?: Record<string, string>): string {
  const normalized = normalizePath(path)
  const url = `${baseUrl}/api/v1/plugins/${pluginName}/${normalized}`
  if (params !== undefined) {
    const searchParams = new URLSearchParams(params)
    const qs = searchParams.toString()
    if (qs.length > 0) {
      return `${url}?${qs}`
    }
  }
  return url
}

function composeSignals(contextSignal: AbortSignal, requestSignal?: AbortSignal): AbortSignal {
  if (requestSignal === undefined) {
    return contextSignal
  }
  return AbortSignal.any([contextSignal, requestSignal])
}

function createApiClient(baseUrl: string, pluginName: string, contextSignal: AbortSignal): PluginApiClient {
  function getHeaders(): Record<string, string> {
    const key = getApiKey()
    const headers: Record<string, string> = {}
    if (key !== null) {
      headers['Authorization'] = `Bearer ${key}`
    }
    return headers
  }

  function buildInit(method: string, headers: Record<string, string>, signal: AbortSignal, body?: string): RequestInit {
    const key = getApiKey()
    const init: RequestInit = { method, headers, signal }
    if (key === null) {
      init.credentials = 'include'
    }
    if (body !== undefined) {
      init.body = body
    }
    return init
  }

  async function handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      let message = response.statusText
      try {
        const body = await response.text()
        if (body.length > 0) {
          message = body
        }
      } catch {
        // use statusText as fallback
      }
      throw new PluginApiError(response.status, message)
    }
    return response.json() as Promise<T>
  }

  async function handleVoidResponse(response: Response): Promise<void> {
    if (!response.ok) {
      let message = response.statusText
      try {
        const body = await response.text()
        if (body.length > 0) {
          message = body
        }
      } catch {
        // use statusText as fallback
      }
      throw new PluginApiError(response.status, message)
    }
  }

  return {
    async get<T = unknown>(path: string, params?: Record<string, string>, signal?: AbortSignal): Promise<T> {
      const url = buildUrl(baseUrl, pluginName, path, params)
      const init = buildInit('GET', getHeaders(), composeSignals(contextSignal, signal))
      const response = await fetch(url, init)
      return handleResponse<T>(response)
    },

    async post<T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal): Promise<T> {
      const url = buildUrl(baseUrl, pluginName, path)
      const headers = { ...getHeaders(), 'Content-Type': 'application/json' }
      const serialized = body !== undefined ? JSON.stringify(body) : undefined
      const init = buildInit('POST', headers, composeSignals(contextSignal, signal), serialized)
      const response = await fetch(url, init)
      return handleResponse<T>(response)
    },

    async put<T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal): Promise<T> {
      const url = buildUrl(baseUrl, pluginName, path)
      const headers = { ...getHeaders(), 'Content-Type': 'application/json' }
      const serialized = body !== undefined ? JSON.stringify(body) : undefined
      const init = buildInit('PUT', headers, composeSignals(contextSignal, signal), serialized)
      const response = await fetch(url, init)
      return handleResponse<T>(response)
    },

    async patch<T = unknown>(path: string, body?: JsonValue, signal?: AbortSignal): Promise<T> {
      const url = buildUrl(baseUrl, pluginName, path)
      const headers = { ...getHeaders(), 'Content-Type': 'application/json' }
      const serialized = body !== undefined ? JSON.stringify(body) : undefined
      const init = buildInit('PATCH', headers, composeSignals(contextSignal, signal), serialized)
      const response = await fetch(url, init)
      return handleResponse<T>(response)
    },

    async del(path: string, params?: Record<string, string>, signal?: AbortSignal): Promise<void> {
      const url = buildUrl(baseUrl, pluginName, path, params)
      const init = buildInit('DELETE', getHeaders(), composeSignals(contextSignal, signal))
      const response = await fetch(url, init)
      return handleVoidResponse(response)
    },

    async raw(path: string, init?: RequestInit): Promise<Response> {
      const url = buildUrl(baseUrl, pluginName, path)
      const key = getApiKey()
      const mergedInit: RequestInit = { ...init }
      if (key !== null) {
        const existingHeaders = new Headers(init?.headers)
        if (!existingHeaders.has('Authorization')) {
          existingHeaders.set('Authorization', `Bearer ${key}`)
        }
        mergedInit.headers = existingHeaders
      } else if (mergedInit.credentials === undefined) {
        mergedInit.credentials = 'include'
      }
      return fetch(url, mergedInit)
    },
  }
}

export function createContext(el: HTMLElement, signal: AbortSignal): PluginContext {
  const tag = el.tagName.toLowerCase()

  const baseUrl = el.getAttribute('base-url')
  if (baseUrl === null || baseUrl === '') {
    throw new PluginError(`Missing required attribute 'base-url' on <${tag}>. Plugin cannot initialize.`)
  }

  const pluginName = el.getAttribute('plugin-name')
  if (pluginName === null || pluginName === '') {
    throw new PluginError(`Missing required attribute 'plugin-name' on <${tag}>. Plugin cannot initialize.`)
  }

  const authUserId = el.getAttribute('auth-user-id')
  const authUsername = el.getAttribute('auth-username')
  const authRole = el.getAttribute('auth-role')

  let auth: PluginAuth | null = null
  if (authUserId !== null && authUsername !== null && authRole !== null) {
    auth = {
      userId: authUserId,
      username: authUsername,
      role: authRole,
    }
  }

  const themeAttr = el.getAttribute('theme')
  let theme: ThemeTokens = DEFAULT_THEME
  if (themeAttr !== null && themeAttr !== '') {
    theme = parseThemeOverrides(themeAttr, tag)
  }

  const api = createApiClient(baseUrl, pluginName, signal)

  function navigate(path: string): void {
    el.dispatchEvent(new CustomEvent(`mcms-navigate:${pluginName}`, {
      bubbles: true,
      composed: true,
      detail: { path },
    }))
  }

  function onNavigate(handler: (path: string) => void): () => void {
    const eventType = `mcms-navigate-to:${pluginName}`
    const listener = (event: Event): void => {
      const detail = (event as CustomEvent).detail as { path: string }
      handler(detail.path)
    }
    el.addEventListener(eventType, listener, { signal })
    return () => {
      el.removeEventListener(eventType, listener)
    }
  }

  return {
    pluginName,
    baseUrl,
    api,
    auth,
    theme,
    onNavigate,
    navigate,
    element: el,
    signal,
  }
}
