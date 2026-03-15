import type { PluginDefinition } from './types.js'
import { PluginError } from './errors.js'
import { createContext } from './context.js'
import { injectTheme, clearTheme, parseThemeOverrides, DEFAULT_THEME } from './theme.js'

const RESERVED_NAMES = [
  'annotation-xml',
  'color-profile',
  'font-face',
  'font-face-src',
  'font-face-name',
  'font-face-format',
  'font-face-uri',
  'missing-glyph',
]

function isLowercaseLetter(code: number): boolean {
  return code >= 97 && code <= 122 // a-z
}

function isDigit(code: number): boolean {
  return code >= 48 && code <= 57 // 0-9
}

function isHyphen(code: number): boolean {
  return code === 45 // -
}

export function validateTag(tag: string): void {
  if (tag.length === 0) {
    throw new PluginError("Tag must not be empty.")
  }

  if (tag.length > 64) {
    throw new PluginError(`Tag must not exceed 64 characters, got ${tag.length}.`)
  }

  const firstChar = tag.charCodeAt(0)
  if (!isLowercaseLetter(firstChar)) {
    throw new PluginError("Tag must start with a lowercase ASCII letter [a-z].")
  }

  let hasHyphen = false
  for (let i = 0; i < tag.length; i++) {
    const code = tag.charCodeAt(i)
    if (isHyphen(code)) {
      hasHyphen = true
    } else if (!isLowercaseLetter(code) && !isDigit(code)) {
      throw new PluginError(`Tag contains invalid character '${tag[i]}' at position ${i}. Only lowercase alphanumeric [a-z0-9] and hyphens are allowed.`)
    }
  }

  if (!hasHyphen) {
    throw new PluginError("Tag must contain at least one hyphen.")
  }

  for (const reserved of RESERVED_NAMES) {
    if (tag === reserved) {
      throw new PluginError(`Tag '${tag}' is a reserved HTML custom element name.`)
    }
  }

  if (customElements.get(tag) !== undefined) {
    throw new PluginError(`Tag '${tag}' already registered. Hot-reload is not supported — reload the page to re-register.`)
  }
}

export function definePlugin(definition: PluginDefinition): void {
  validateTag(definition.tag)

  const { tag, setup, mount, destroy } = definition

  class PluginElement extends HTMLElement {
    _initialized = false
    _abortController: AbortController | null = null

    async connectedCallback(): Promise<void> {
      if (this._initialized) {
        return
      }
      this._initialized = true

      const controller = new AbortController()
      this._abortController = controller

      const pluginName = this.getAttribute('plugin-name')
      if (pluginName !== null) {
        this.setAttribute('data-mcms-plugin', pluginName)
      }

      this.style.contain = 'layout style'

      const themeAttr = this.getAttribute('theme')
      let theme = DEFAULT_THEME
      if (themeAttr !== null && themeAttr !== '') {
        theme = parseThemeOverrides(themeAttr, tag)
      }
      injectTheme(this, theme)

      try {
        const ctx = createContext(this, controller.signal)
        await setup(ctx, this)

        if (!this.isConnected) {
          return
        }

        await mount(ctx, this)
      } catch (err: unknown) {
        const error = err instanceof Error ? err : new Error(String(err))
        console.error(error)
        const pluginName = this.getAttribute('plugin-name')
        this.dispatchEvent(new CustomEvent('mcms-error', {
          bubbles: true,
          composed: true,
          detail: { error, pluginName },
        }))
        this.innerHTML = `<div role="alert" style="color: red; padding: 1em;">Plugin error: ${error.message}</div>`
      }
    }

    disconnectedCallback(): void {
      if (destroy !== undefined) {
        try {
          destroy(this)
        } catch (err: unknown) {
          const error = err instanceof Error ? err : new Error(String(err))
          console.error(error)
          const pluginName = this.getAttribute('plugin-name')
          this.dispatchEvent(new CustomEvent('mcms-error', {
            bubbles: true,
            composed: true,
            detail: { error, pluginName },
          }))
        }
      }

      clearTheme(this)
      this.innerHTML = ''

      if (this._abortController !== null) {
        this._abortController.abort()
        this._abortController = null
      }
    }
  }

  customElements.define(tag, PluginElement)
}
