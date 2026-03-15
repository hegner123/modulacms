import { describe, it, expect, vi, afterEach } from 'vitest'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { definePlugin, validateTag } from './define-plugin.js'
import { PluginError } from './errors.js'

// Shared fixture for tag validation
const fixturePath = path.resolve(import.meta.dirname, '../../../../testdata/tag-validation-cases.json')
const cases = JSON.parse(readFileSync(fixturePath, 'utf-8')) as Array<{ tag: string; valid: boolean; reason?: string }>

let tagCounter = 0

function uniqueTag(prefix = 'dp'): string {
  tagCounter++
  return `${prefix}-test-${tagCounter}`
}

function mountPlugin(tag: string, attrs: Record<string, string> = {}): { el: HTMLElement; cleanup: () => void } {
  const defaults: Record<string, string> = {
    'base-url': 'https://example.com',
    'plugin-name': 'test_plugin',
  }
  const merged = { ...defaults, ...attrs }

  const el = document.createElement(tag)
  for (const [key, value] of Object.entries(merged)) {
    el.setAttribute(key, value)
  }
  document.body.appendChild(el)

  return {
    el,
    cleanup: () => {
      if (el.parentNode !== null) {
        el.parentNode.removeChild(el)
      }
    },
  }
}

afterEach(() => {
  document.body.innerHTML = ''
  vi.restoreAllMocks()
})

// Stub fetch globally for lifecycle tests
vi.stubGlobal('fetch', async () => {
  return new Response(JSON.stringify({}), { status: 200 })
})

describe('validateTag — shared fixture', () => {
  for (const c of cases) {
    it(`${c.valid ? 'accepts' : 'rejects'}: "${c.tag}" (${c.reason})`, () => {
      // We need a way to test without the customElements.get() check interfering.
      // For valid tags, validateTag should not throw (ignoring already-registered check).
      // For invalid tags, validateTag should throw PluginError.
      if (c.valid) {
        // Valid tags should pass validation. If the tag was previously registered
        // in another test, we skip the already-registered check by catching that specific error.
        try {
          validateTag(c.tag)
        } catch (e) {
          if (e instanceof PluginError && (e.message).includes('already registered')) {
            // This is fine - tag is valid but was registered by a previous test
          } else {
            throw e
          }
        }
      } else {
        expect(() => validateTag(c.tag)).toThrow(PluginError)
      }
    })
  }
})

describe('validateTag — additional rules', () => {
  it('rejects already registered tags', () => {
    const tag = uniqueTag('val')
    definePlugin({
      tag,
      setup: () => {},
      mount: () => {},
    })

    expect(() => validateTag(tag)).toThrow('already registered')
  })
})

describe('definePlugin — registration', () => {
  it('registers a custom element', () => {
    const tag = uniqueTag()
    definePlugin({
      tag,
      setup: () => {},
      mount: () => {},
    })

    expect(customElements.get(tag)).toBeDefined()
  })

  it('throws on invalid tag', () => {
    expect(() => definePlugin({
      tag: 'nohyphen',
      setup: () => {},
      mount: () => {},
    })).toThrow(PluginError)
  })
})

describe('definePlugin — lifecycle', () => {
  it('calls setup before mount', async () => {
    const tag = uniqueTag()
    const order: string[] = []

    definePlugin({
      tag,
      setup: () => { order.push('setup') },
      mount: () => { order.push('mount') },
    })

    mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    expect(order).toEqual(['setup', 'mount'])
  })

  it('supports async setup and mount', async () => {
    const tag = uniqueTag()
    const order: string[] = []

    definePlugin({
      tag,
      setup: async () => {
        await new Promise(r => setTimeout(r, 5))
        order.push('setup')
      },
      mount: async () => {
        await new Promise(r => setTimeout(r, 5))
        order.push('mount')
      },
    })

    mountPlugin(tag)
    await new Promise(r => setTimeout(r, 50))

    expect(order).toEqual(['setup', 'mount'])
  })

  it('does not call mount if setup throws', async () => {
    const tag = uniqueTag()
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    let mountCalled = false

    definePlugin({
      tag,
      setup: () => { throw new Error('setup failed') },
      mount: () => { mountCalled = true },
    })

    mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    expect(mountCalled).toBe(false)
    consoleSpy.mockRestore()
  })

  it('double-init guard prevents re-initialization', async () => {
    const tag = uniqueTag()
    let setupCount = 0

    definePlugin({
      tag,
      setup: () => { setupCount++ },
      mount: () => {},
    })

    const { el } = mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    // Force re-trigger connectedCallback by removing and re-adding
    document.body.removeChild(el)
    document.body.appendChild(el)
    await new Promise(r => setTimeout(r, 10))

    // setup should only be called once because _initialized flag and re-append is no-op
    expect(setupCount).toBe(1)
  })

  it('re-append after disconnect is a no-op', async () => {
    const tag = uniqueTag()
    let setupCount = 0
    let mountCount = 0

    definePlugin({
      tag,
      setup: () => { setupCount++ },
      mount: () => { mountCount++ },
    })

    const { el } = mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    expect(setupCount).toBe(1)
    expect(mountCount).toBe(1)

    // Disconnect
    document.body.removeChild(el)

    // Re-append
    document.body.appendChild(el)
    await new Promise(r => setTimeout(r, 10))

    // Should not have been called again
    expect(setupCount).toBe(1)
    expect(mountCount).toBe(1)
  })

  it('sets data-mcms-plugin attribute', async () => {
    const tag = uniqueTag()

    definePlugin({
      tag,
      setup: () => {},
      mount: () => {},
    })

    const { el } = mountPlugin(tag, { 'plugin-name': 'my_plugin' })
    await new Promise(r => setTimeout(r, 10))

    expect(el.getAttribute('data-mcms-plugin')).toBe('my_plugin')
  })

  it('applies CSS containment', async () => {
    const tag = uniqueTag()

    definePlugin({
      tag,
      setup: () => {},
      mount: () => {},
    })

    const { el } = mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    expect(el.style.contain).toBe('layout style')
  })

  it('signal is aborted on disconnect', async () => {
    const tag = uniqueTag()
    let capturedSignal: AbortSignal | undefined

    definePlugin({
      tag,
      setup: (ctx) => { capturedSignal = ctx.signal },
      mount: () => {},
    })

    const { el, cleanup } = mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    expect(capturedSignal).toBeDefined()
    expect(capturedSignal!.aborted).toBe(false)

    cleanup()

    expect(capturedSignal!.aborted).toBe(true)
  })

  it('disconnect during async setup skips mount', async () => {
    const tag = uniqueTag()
    let mountCalled = false
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    definePlugin({
      tag,
      setup: async (_ctx, el) => {
        // Simulate long async operation
        await new Promise(r => setTimeout(r, 20))
        // Element might be disconnected by now
      },
      mount: () => { mountCalled = true },
    })

    const { el, cleanup } = mountPlugin(tag)

    // Disconnect immediately during async setup
    await new Promise(r => setTimeout(r, 5))
    cleanup()

    // Wait for setup to complete
    await new Promise(r => setTimeout(r, 50))

    // mount should not have been called because isConnected is false
    expect(mountCalled).toBe(false)
    consoleSpy.mockRestore()
  })

  it('calls destroy on disconnect', async () => {
    const tag = uniqueTag()
    let destroyCalled = false

    definePlugin({
      tag,
      setup: () => {},
      mount: () => {},
      destroy: () => { destroyCalled = true },
    })

    const { cleanup } = mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    cleanup()

    expect(destroyCalled).toBe(true)
  })

  it('clears innerHTML on disconnect', async () => {
    const tag = uniqueTag()

    definePlugin({
      tag,
      setup: () => {},
      mount: (_ctx, el) => { el.innerHTML = '<p>content</p>' },
    })

    const { el, cleanup } = mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    expect(el.innerHTML).toBe('<p>content</p>')

    cleanup()

    expect(el.innerHTML).toBe('')
  })

  it('clears theme properties on disconnect', async () => {
    const tag = uniqueTag()

    definePlugin({
      tag,
      setup: () => {},
      mount: () => {},
    })

    const { el, cleanup } = mountPlugin(tag)
    await new Promise(r => setTimeout(r, 10))

    expect(el.style.getPropertyValue('--mcms-color-primary')).not.toBe('')

    cleanup()

    expect(el.style.getPropertyValue('--mcms-color-primary')).toBe('')
  })
})
