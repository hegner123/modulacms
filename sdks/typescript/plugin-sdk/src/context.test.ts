import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { createContext } from './context.js'
import { configure, resetConfig } from './configure.js'
import { PluginApiError } from './errors.js'

describe('createContext', () => {
  afterEach(() => {
    resetConfig()
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  function makeElement(attrs: Record<string, string> = {}): HTMLElement {
    const el = document.createElement('div')
    const defaults: Record<string, string> = {
      'base-url': 'https://example.com',
      'plugin-name': 'test_plugin',
    }
    const merged = { ...defaults, ...attrs }
    for (const [key, value] of Object.entries(merged)) {
      el.setAttribute(key, value)
    }
    return el
  }

  describe('attribute reading', () => {
    it('reads base-url and plugin-name', () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      expect(ctx.baseUrl).toBe('https://example.com')
      expect(ctx.pluginName).toBe('test_plugin')
      controller.abort()
    })

    it('reads auth attributes', () => {
      const el = makeElement({
        'auth-user-id': '01HXK123',
        'auth-username': 'admin',
        'auth-role': 'admin',
      })
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      expect(ctx.auth).not.toBeNull()
      expect(ctx.auth!.userId).toBe('01HXK123')
      expect(ctx.auth!.username).toBe('admin')
      expect(ctx.auth!.role).toBe('admin')
      controller.abort()
    })

    it('sets auth to null when auth attributes are absent', () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      expect(ctx.auth).toBeNull()
      controller.abort()
    })

    it('exposes element reference', () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      expect(ctx.element).toBe(el)
      controller.abort()
    })

    it('exposes signal', () => {
      const controller = new AbortController()
      const el = makeElement()
      const ctx = createContext(el, controller.signal)

      expect(ctx.signal).toBe(controller.signal)
      expect(ctx.signal.aborted).toBe(false)
      controller.abort()
      expect(ctx.signal.aborted).toBe(true)
    })
  })

  describe('required attribute validation', () => {
    it('throws on missing base-url', () => {
      const el = document.createElement('div')
      el.setAttribute('plugin-name', 'test_plugin')
      const controller = new AbortController()

      expect(() => createContext(el, controller.signal)).toThrow(
        "Missing required attribute 'base-url'"
      )
      controller.abort()
    })

    it('throws on missing plugin-name', () => {
      const el = document.createElement('div')
      el.setAttribute('base-url', 'https://example.com')
      const controller = new AbortController()

      expect(() => createContext(el, controller.signal)).toThrow(
        "Missing required attribute 'plugin-name'"
      )
      controller.abort()
    })

    it('throws on empty base-url', () => {
      const el = document.createElement('div')
      el.setAttribute('base-url', '')
      el.setAttribute('plugin-name', 'test_plugin')
      const controller = new AbortController()

      expect(() => createContext(el, controller.signal)).toThrow(
        "Missing required attribute 'base-url'"
      )
      controller.abort()
    })
  })

  describe('API client', () => {
    beforeEach(() => {
      vi.stubGlobal('fetch', async (url: string, init?: RequestInit) => {
        return new Response(JSON.stringify({ url, method: init?.method }), { status: 200 })
      })
    })

    it('prefixes paths with /api/v1/plugins/<name>/', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedUrl = ''
      vi.stubGlobal('fetch', async (url: string, _init?: RequestInit) => {
        capturedUrl = url
        return new Response(JSON.stringify({}), { status: 200 })
      })

      await ctx.api.get('tasks')
      expect(capturedUrl).toBe('https://example.com/api/v1/plugins/test_plugin/tasks')
      controller.abort()
    })

    it('strips leading slash from path', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedUrl = ''
      vi.stubGlobal('fetch', async (url: string, _init?: RequestInit) => {
        capturedUrl = url
        return new Response(JSON.stringify({}), { status: 200 })
      })

      await ctx.api.get('/tasks')
      expect(capturedUrl).toBe('https://example.com/api/v1/plugins/test_plugin/tasks')
      controller.abort()
    })

    it('appends query params on GET', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedUrl = ''
      vi.stubGlobal('fetch', async (url: string, _init?: RequestInit) => {
        capturedUrl = url
        return new Response(JSON.stringify({}), { status: 200 })
      })

      await ctx.api.get('tasks', { status: 'active', page: '1' })
      expect(capturedUrl).toContain('?')
      expect(capturedUrl).toContain('status=active')
      expect(capturedUrl).toContain('page=1')
      controller.abort()
    })

    it('appends query params on DEL', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedUrl = ''
      vi.stubGlobal('fetch', async (url: string, _init?: RequestInit) => {
        capturedUrl = url
        return new Response(null, { status: 204 })
      })

      await ctx.api.del('tasks', { id: '123' })
      expect(capturedUrl).toContain('id=123')
      controller.abort()
    })

    it('sends JSON body on POST', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedInit: RequestInit | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedInit = init
        return new Response(JSON.stringify({ id: 1 }), { status: 201 })
      })

      await ctx.api.post('tasks', { title: "New Task" })
      expect(capturedInit!.method).toBe('POST')
      expect(capturedInit!.body).toBe(JSON.stringify({ title: "New Task" }))
      const headers = capturedInit!.headers as Record<string, string>
      expect(headers['Content-Type']).toBe('application/json')
      controller.abort()
    })

    it('sends JSON body on PUT', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedInit: RequestInit | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedInit = init
        return new Response(JSON.stringify({ id: 1 }), { status: 200 })
      })

      await ctx.api.put('tasks/1', { title: "Updated" })
      expect(capturedInit!.method).toBe('PUT')
      expect(capturedInit!.body).toBe(JSON.stringify({ title: "Updated" }))
      controller.abort()
    })

    it('sends JSON body on PATCH', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedInit: RequestInit | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedInit = init
        return new Response(JSON.stringify({ id: 1 }), { status: 200 })
      })

      await ctx.api.patch('tasks/1', { status: "done" })
      expect(capturedInit!.method).toBe('PATCH')
      expect(capturedInit!.body).toBe(JSON.stringify({ status: "done" }))
      controller.abort()
    })

    it('throws PluginApiError on non-2xx GET', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      vi.stubGlobal('fetch', async () => {
        return new Response('Not Found', { status: 404, statusText: 'Not Found' })
      })

      await expect(ctx.api.get('missing')).rejects.toThrow(PluginApiError)
      try {
        await ctx.api.get('missing')
      } catch (e) {
        const err = e as PluginApiError
        expect(err.status).toBe(404)
        expect(err.message).toBe('Not Found')
      }
      controller.abort()
    })

    it('throws PluginApiError on non-2xx POST', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      vi.stubGlobal('fetch', async () => {
        return new Response('Bad Request', { status: 400, statusText: 'Bad Request' })
      })

      await expect(ctx.api.post('tasks', { bad: "data" })).rejects.toThrow(PluginApiError)
      controller.abort()
    })

    it('throws PluginApiError on non-2xx DEL', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      vi.stubGlobal('fetch', async () => {
        return new Response('Forbidden', { status: 403, statusText: 'Forbidden' })
      })

      await expect(ctx.api.del('tasks/1')).rejects.toThrow(PluginApiError)
      controller.abort()
    })

    it('raw does NOT throw on non-2xx', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      vi.stubGlobal('fetch', async () => {
        return new Response('Error', { status: 500, statusText: 'Internal Server Error' })
      })

      const response = await ctx.api.raw('tasks')
      expect(response.ok).toBe(false)
      expect(response.status).toBe(500)
      controller.abort()
    })
  })

  describe('configure auth', () => {
    it('uses Bearer header when API key is configured', async () => {
      configure({ apiKey: 'secret-key' })

      let capturedInit: RequestInit | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedInit = init
        return new Response(JSON.stringify({}), { status: 200 })
      })

      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      await ctx.api.get('tasks')

      const headers = capturedInit!.headers as Record<string, string>
      expect(headers['Authorization']).toBe('Bearer secret-key')
      expect(capturedInit!.credentials).toBeUndefined()
      controller.abort()
    })

    it('uses credentials include when no API key', async () => {
      let capturedInit: RequestInit | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedInit = init
        return new Response(JSON.stringify({}), { status: 200 })
      })

      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      await ctx.api.get('tasks')

      expect(capturedInit!.credentials).toBe('include')
      controller.abort()
    })
  })

  describe('navigation', () => {
    it('navigate dispatches CustomEvent on element', () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      const events: CustomEvent[] = []
      el.addEventListener('mcms-navigate:test_plugin', ((e: Event) => {
        events.push(e as CustomEvent)
      }) as EventListener)

      ctx.navigate('/dashboard')

      expect(events).toHaveLength(1)
      expect((events[0]!.detail as { path: string }).path).toBe('/dashboard')
      expect(events[0]!.bubbles).toBe(true)
      expect(events[0]!.composed).toBe(true)
      controller.abort()
    })

    it('onNavigate receives inbound navigation events', () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      const paths: string[] = []
      ctx.onNavigate((path) => {
        paths.push(path)
      })

      // Host dispatches navigate-to event
      el.dispatchEvent(new CustomEvent('mcms-navigate-to:test_plugin', {
        detail: { path: '/settings' },
      }))

      expect(paths).toEqual(['/settings'])
      controller.abort()
    })

    it('onNavigate returns unsubscribe function', () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      const paths: string[] = []
      const unsub = ctx.onNavigate((path) => {
        paths.push(path)
      })

      el.dispatchEvent(new CustomEvent('mcms-navigate-to:test_plugin', {
        detail: { path: '/first' },
      }))

      unsub()

      el.dispatchEvent(new CustomEvent('mcms-navigate-to:test_plugin', {
        detail: { path: '/second' },
      }))

      expect(paths).toEqual(['/first'])
      controller.abort()
    })

    it('onNavigate auto-cleans up when signal aborts', () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      const paths: string[] = []
      ctx.onNavigate((path) => {
        paths.push(path)
      })

      el.dispatchEvent(new CustomEvent('mcms-navigate-to:test_plugin', {
        detail: { path: '/before' },
      }))

      controller.abort()

      el.dispatchEvent(new CustomEvent('mcms-navigate-to:test_plugin', {
        detail: { path: '/after' },
      }))

      expect(paths).toEqual(['/before'])
    })
  })

  describe('AbortSignal', () => {
    it('passes context signal to fetch', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)

      let capturedSignal: AbortSignal | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedSignal = init?.signal as AbortSignal | undefined
        return new Response(JSON.stringify({}), { status: 200 })
      })

      await ctx.api.get('tasks')
      expect(capturedSignal).toBeDefined()
      expect(capturedSignal!.aborted).toBe(false)
      controller.abort()
    })

    it('composes context signal with per-request signal via AbortSignal.any', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)
      const requestController = new AbortController()

      let capturedSignal: AbortSignal | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedSignal = init?.signal as AbortSignal | undefined
        return new Response(JSON.stringify({}), { status: 200 })
      })

      await ctx.api.get('tasks', undefined, requestController.signal)

      expect(capturedSignal).toBeDefined()
      // The composed signal should not be the same as either individual signal
      expect(capturedSignal).not.toBe(controller.signal)
      expect(capturedSignal).not.toBe(requestController.signal)
      // Aborting the request signal should abort the composed signal
      requestController.abort()
      expect(capturedSignal!.aborted).toBe(true)
      controller.abort()
    })

    it('context signal abort cancels composed signal', async () => {
      const el = makeElement()
      const controller = new AbortController()
      const ctx = createContext(el, controller.signal)
      const requestController = new AbortController()

      let capturedSignal: AbortSignal | undefined
      vi.stubGlobal('fetch', async (_url: string, init?: RequestInit) => {
        capturedSignal = init?.signal as AbortSignal | undefined
        return new Response(JSON.stringify({}), { status: 200 })
      })

      await ctx.api.post('tasks', { title: "test" }, requestController.signal)

      expect(capturedSignal).toBeDefined()
      controller.abort()
      expect(capturedSignal!.aborted).toBe(true)
    })
  })
})
