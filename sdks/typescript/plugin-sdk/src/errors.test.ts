import { describe, it, expect, vi, afterEach } from 'vitest'
import { PluginError, PluginApiError, isPluginError, isPluginApiError } from './errors.js'

describe('PluginError', () => {
  it('constructs with message', () => {
    const err = new PluginError("something went wrong")
    expect(err.message).toBe("something went wrong")
    expect(err.name).toBe("PluginError")
    expect(err._tag).toBe("PluginError")
    expect(err).toBeInstanceOf(Error)
  })

  it('has stack trace', () => {
    const err = new PluginError("test")
    expect(err.stack).toBeDefined()
  })
})

describe('PluginApiError', () => {
  it('constructs with status and message', () => {
    const err = new PluginApiError(404, "Not Found")
    expect(err.status).toBe(404)
    expect(err.message).toBe("Not Found")
    expect(err.name).toBe("PluginApiError")
    expect(err._tag).toBe("PluginApiError")
    expect(err).toBeInstanceOf(Error)
  })

  it('preserves status code', () => {
    const err = new PluginApiError(500, "Internal Server Error")
    expect(err.status).toBe(500)
  })
})

describe('isPluginError', () => {
  it('returns true for PluginError instances', () => {
    const err = new PluginError("test")
    expect(isPluginError(err)).toBe(true)
  })

  it('returns false for PluginApiError', () => {
    const err = new PluginApiError(500, "test")
    expect(isPluginError(err)).toBe(false)
  })

  it('returns false for plain Error', () => {
    expect(isPluginError(new Error("test"))).toBe(false)
  })

  it('returns false for null', () => {
    expect(isPluginError(null)).toBe(false)
  })

  it('returns false for undefined', () => {
    expect(isPluginError(undefined)).toBe(false)
  })

  it('returns false for string', () => {
    expect(isPluginError("error")).toBe(false)
  })

  it('returns true for duck-typed object with matching _tag', () => {
    const fake = { _tag: "PluginError" as const, message: "fake" }
    expect(isPluginError(fake)).toBe(true)
  })
})

describe('isPluginApiError', () => {
  it('returns true for PluginApiError instances', () => {
    const err = new PluginApiError(400, "Bad Request")
    expect(isPluginApiError(err)).toBe(true)
  })

  it('returns false for PluginError', () => {
    const err = new PluginError("test")
    expect(isPluginApiError(err)).toBe(false)
  })

  it('returns false for plain Error', () => {
    expect(isPluginApiError(new Error("test"))).toBe(false)
  })

  it('returns false for null', () => {
    expect(isPluginApiError(null)).toBe(false)
  })

  it('returns false for undefined', () => {
    expect(isPluginApiError(undefined)).toBe(false)
  })

  it('returns false for number', () => {
    expect(isPluginApiError(42)).toBe(false)
  })
})

describe('error event dispatch from lifecycle', () => {
  let counter = 0

  function uniqueTag(): string {
    counter++
    return `err-lifecycle-${counter}`
  }

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('dispatches mcms-error event when setup throws', async () => {
    const tag = uniqueTag()
    const { definePlugin } = await import('./define-plugin.js')

    const errorEvents: CustomEvent[] = []
    document.body.addEventListener('mcms-error', ((e: Event) => {
      errorEvents.push(e as CustomEvent)
    }) as EventListener)

    definePlugin({
      tag,
      setup: () => { throw new PluginError("setup failed") },
      mount: () => {},
    })

    const el = document.createElement(tag)
    el.setAttribute('base-url', 'https://example.com')
    el.setAttribute('plugin-name', 'test_plugin')
    document.body.appendChild(el)

    // Wait for async connectedCallback
    await new Promise(r => setTimeout(r, 10))

    expect(errorEvents.length).toBeGreaterThanOrEqual(1)
    const detail = errorEvents[0]!.detail as { error: Error; pluginName: string }
    expect(detail.error.message).toBe("setup failed")
    expect(detail.pluginName).toBe("test_plugin")
  })

  it('renders fallback UI with role="alert" on error', async () => {
    const tag = uniqueTag()
    const { definePlugin } = await import('./define-plugin.js')

    definePlugin({
      tag,
      setup: () => { throw new Error("render failed") },
      mount: () => {},
    })

    const el = document.createElement(tag)
    el.setAttribute('base-url', 'https://example.com')
    el.setAttribute('plugin-name', 'test_plugin')
    document.body.appendChild(el)

    await new Promise(r => setTimeout(r, 10))

    const alert = el.querySelector('[role="alert"]')
    expect(alert).not.toBeNull()
    expect(alert!.textContent).toContain("render failed")
  })

  it('destroy error still cleans up DOM', async () => {
    const tag = uniqueTag()
    const { definePlugin } = await import('./define-plugin.js')

    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    definePlugin({
      tag,
      setup: () => {},
      mount: (_ctx, el) => { el.innerHTML = '<p>content</p>' },
      destroy: () => { throw new Error("destroy failed") },
    })

    const el = document.createElement(tag)
    el.setAttribute('base-url', 'https://example.com')
    el.setAttribute('plugin-name', 'test_plugin')
    document.body.appendChild(el)

    await new Promise(r => setTimeout(r, 10))

    expect(el.innerHTML).toBe('<p>content</p>')

    // Remove from DOM triggers disconnectedCallback
    document.body.removeChild(el)

    // innerHTML should be cleared despite destroy throwing
    expect(el.innerHTML).toBe('')
    expect(consoleSpy).toHaveBeenCalled()

    consoleSpy.mockRestore()
  })
})
