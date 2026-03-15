import { describe, it, expect, afterEach, vi } from 'vitest'
import { configure, resetConfig, getApiKey } from './configure.js'

describe('configure', () => {
  afterEach(() => {
    resetConfig()
  })

  it('sets API key', () => {
    configure({ apiKey: 'test-key-123' })
    expect(getApiKey()).toBe('test-key-123')
  })

  it('overwrites previous API key', () => {
    configure({ apiKey: 'first-key' })
    configure({ apiKey: 'second-key' })
    expect(getApiKey()).toBe('second-key')
  })
})

describe('resetConfig', () => {
  it('clears the API key', () => {
    configure({ apiKey: 'test-key' })
    expect(getApiKey()).toBe('test-key')
    resetConfig()
    expect(getApiKey()).toBeNull()
  })

  it('is idempotent', () => {
    resetConfig()
    resetConfig()
    expect(getApiKey()).toBeNull()
  })
})

describe('cookie fallback', () => {
  afterEach(() => {
    resetConfig()
    vi.restoreAllMocks()
  })

  it('uses credentials include when no API key is set', async () => {
    let capturedInit: RequestInit | undefined

    vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
      capturedInit = init
      return new Response(JSON.stringify({ ok: true }), { status: 200 })
    })

    // Import context to create an API client without configuring an API key
    const { createContext } = await import('./context.js')
    const el = document.createElement('div')
    el.setAttribute('base-url', 'https://example.com')
    el.setAttribute('plugin-name', 'test_plugin')
    const controller = new AbortController()
    const ctx = createContext(el, controller.signal)

    await ctx.api.get('tasks')

    expect(capturedInit).toBeDefined()
    expect(capturedInit!.credentials).toBe('include')
    expect(capturedInit!.headers).toBeDefined()
    const headers = capturedInit!.headers as Record<string, string>
    expect(headers['Authorization']).toBeUndefined()

    controller.abort()
  })

  it('uses Bearer header when API key is configured', async () => {
    configure({ apiKey: 'my-secret-key' })

    let capturedInit: RequestInit | undefined

    vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
      capturedInit = init
      return new Response(JSON.stringify({ ok: true }), { status: 200 })
    })

    const { createContext } = await import('./context.js')
    const el = document.createElement('div')
    el.setAttribute('base-url', 'https://example.com')
    el.setAttribute('plugin-name', 'test_plugin')
    const controller = new AbortController()
    const ctx = createContext(el, controller.signal)

    await ctx.api.get('tasks')

    expect(capturedInit).toBeDefined()
    expect(capturedInit!.credentials).toBeUndefined()
    const headers = capturedInit!.headers as Record<string, string>
    expect(headers['Authorization']).toBe('Bearer my-secret-key')

    controller.abort()
  })
})
