import { createAdminClient } from '@modulacms/admin-sdk'
import type { ModulaCMSAdminClient } from '@modulacms/admin-sdk'
import { ModulaClient } from '@modulacms/sdk'

const baseUrl = import.meta.env.VITE_CMS_URL || window.location.origin
const apiKey = import.meta.env.VITE_CMS_API_KEY

// Session-only client: sends cookies, no API key header
export const sessionSdk = createAdminClient({
  baseUrl,
  credentials: 'include',
  allowInsecure: import.meta.env.DEV,
})

// API key client: sends API key header (+ cookies)
const apiKeySdk = apiKey
  ? createAdminClient({
      baseUrl,
      apiKey,
      credentials: 'include',
      allowInsecure: import.meta.env.DEV,
    })
  : sessionSdk

// Active client — starts with API key if configured, switches to session when user logs in
let active: ModulaCMSAdminClient = apiKey ? apiKeySdk : sessionSdk

export function activateSessionAuth() {
  active = sessionSdk
}

export function activateApiKeyAuth() {
  active = apiKeySdk
}

// Proxy so all imports of `sdk` always see the current active client
export const sdk: ModulaCMSAdminClient = new Proxy({} as ModulaCMSAdminClient, {
  get(_target, prop, receiver) {
    return Reflect.get(active, prop, receiver)
  },
})

export const cms = new ModulaClient({
  baseUrl,
  apiKey,
  credentials: 'include',
})
