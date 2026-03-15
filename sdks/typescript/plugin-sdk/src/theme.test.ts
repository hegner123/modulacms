import { describe, it, expect, vi, afterEach } from 'vitest'
import { DEFAULT_THEME, injectTheme, clearTheme, parseThemeOverrides } from './theme.js'

describe('DEFAULT_THEME', () => {
  it('has OKLCH color values', () => {
    expect(DEFAULT_THEME.colors.primary).toBe("oklch(0.65 0.15 250)")
    expect(DEFAULT_THEME.colors.surface).toBe("oklch(0.98 0 0)")
    expect(DEFAULT_THEME.colors.border).toBe("oklch(0.85 0 0)")
    expect(DEFAULT_THEME.colors.text).toBe("oklch(0.2 0 0)")
    expect(DEFAULT_THEME.colors.error).toBe("oklch(0.55 0.2 25)")
    expect(DEFAULT_THEME.colors.success).toBe("oklch(0.6 0.15 145)")
    expect(DEFAULT_THEME.colors.warning).toBe("oklch(0.7 0.15 85)")
  })

  it('has spacing values', () => {
    expect(DEFAULT_THEME.spacing.xs).toBe("0.25rem")
    expect(DEFAULT_THEME.spacing.sm).toBe("0.5rem")
    expect(DEFAULT_THEME.spacing.md).toBe("1rem")
    expect(DEFAULT_THEME.spacing.lg).toBe("1.5rem")
    expect(DEFAULT_THEME.spacing.xl).toBe("2rem")
    expect(DEFAULT_THEME.spacing.xl2).toBe("3rem")
  })

  it('has radii values', () => {
    expect(DEFAULT_THEME.radii.sm).toBe("0.25rem")
    expect(DEFAULT_THEME.radii.md).toBe("0.5rem")
    expect(DEFAULT_THEME.radii.lg).toBe("1rem")
    expect(DEFAULT_THEME.radii.full).toBe("9999px")
  })

  it('has typography values', () => {
    expect(DEFAULT_THEME.typography.fontFamily).toBe("system-ui, -apple-system, sans-serif")
    expect(DEFAULT_THEME.typography.fontSizeSm).toBe("0.875rem")
    expect(DEFAULT_THEME.typography.fontSizeMd).toBe("1rem")
    expect(DEFAULT_THEME.typography.fontSizeLg).toBe("1.25rem")
    expect(DEFAULT_THEME.typography.lineHeight).toBe("1.5")
  })
})

describe('injectTheme', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('sets CSS custom properties on element', () => {
    const el = document.createElement('div')
    document.body.appendChild(el)

    injectTheme(el, DEFAULT_THEME)

    expect(el.style.getPropertyValue('--mcms-color-primary')).toBe("oklch(0.65 0.15 250)")
    expect(el.style.getPropertyValue('--mcms-color-surface')).toBe("oklch(0.98 0 0)")
    expect(el.style.getPropertyValue('--mcms-color-border')).toBe("oklch(0.85 0 0)")
    expect(el.style.getPropertyValue('--mcms-color-text')).toBe("oklch(0.2 0 0)")
    expect(el.style.getPropertyValue('--mcms-color-error')).toBe("oklch(0.55 0.2 25)")
    expect(el.style.getPropertyValue('--mcms-color-success')).toBe("oklch(0.6 0.15 145)")
    expect(el.style.getPropertyValue('--mcms-color-warning')).toBe("oklch(0.7 0.15 85)")

    expect(el.style.getPropertyValue('--mcms-spacing-xs')).toBe("0.25rem")
    expect(el.style.getPropertyValue('--mcms-spacing-sm')).toBe("0.5rem")
    expect(el.style.getPropertyValue('--mcms-spacing-md')).toBe("1rem")
    expect(el.style.getPropertyValue('--mcms-spacing-lg')).toBe("1.5rem")
    expect(el.style.getPropertyValue('--mcms-spacing-xl')).toBe("2rem")
    expect(el.style.getPropertyValue('--mcms-spacing-2xl')).toBe("3rem")

    expect(el.style.getPropertyValue('--mcms-radius-sm')).toBe("0.25rem")
    expect(el.style.getPropertyValue('--mcms-radius-md')).toBe("0.5rem")
    expect(el.style.getPropertyValue('--mcms-radius-lg')).toBe("1rem")
    expect(el.style.getPropertyValue('--mcms-radius-full')).toBe("9999px")

    expect(el.style.getPropertyValue('--mcms-font-family')).toBe("system-ui, -apple-system, sans-serif")
    expect(el.style.getPropertyValue('--mcms-font-size-sm')).toBe("0.875rem")
    expect(el.style.getPropertyValue('--mcms-font-size-md')).toBe("1rem")
    expect(el.style.getPropertyValue('--mcms-font-size-lg')).toBe("1.25rem")
    expect(el.style.getPropertyValue('--mcms-line-height')).toBe("1.5")
  })
})

describe('clearTheme', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('removes all --mcms- CSS custom properties', () => {
    const el = document.createElement('div')
    document.body.appendChild(el)

    injectTheme(el, DEFAULT_THEME)
    expect(el.style.getPropertyValue('--mcms-color-primary')).not.toBe('')

    clearTheme(el)

    expect(el.style.getPropertyValue('--mcms-color-primary')).toBe('')
    expect(el.style.getPropertyValue('--mcms-spacing-xs')).toBe('')
    expect(el.style.getPropertyValue('--mcms-radius-sm')).toBe('')
    expect(el.style.getPropertyValue('--mcms-font-family')).toBe('')
  })
})

describe('parseThemeOverrides', () => {
  it('deep-merges partial color overrides into defaults', () => {
    const json = JSON.stringify({ colors: { primary: "oklch(0.8 0.2 300)" } })
    const result = parseThemeOverrides(json, 'test-tag')

    expect(result.colors.primary).toBe("oklch(0.8 0.2 300)")
    // Other colors should remain as defaults
    expect(result.colors.surface).toBe(DEFAULT_THEME.colors.surface)
    expect(result.colors.border).toBe(DEFAULT_THEME.colors.border)
    expect(result.colors.text).toBe(DEFAULT_THEME.colors.text)
  })

  it('deep-merges spacing overrides', () => {
    const json = JSON.stringify({ spacing: { xs: "0.5rem" } })
    const result = parseThemeOverrides(json, 'test-tag')

    expect(result.spacing.xs).toBe("0.5rem")
    expect(result.spacing.sm).toBe(DEFAULT_THEME.spacing.sm)
  })

  it('returns DEFAULT_THEME for malformed JSON', () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    const result = parseThemeOverrides('not valid json{{{', 'bad-tag')

    expect(result).toEqual(DEFAULT_THEME)
    expect(warnSpy).toHaveBeenCalledTimes(1)
    expect(warnSpy.mock.calls[0]![0]).toContain('Invalid theme JSON on <bad-tag>')

    warnSpy.mockRestore()
  })

  it('returns DEFAULT_THEME for non-object JSON', () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    const result = parseThemeOverrides('"just a string"', 'arr-tag')

    expect(result).toEqual(DEFAULT_THEME)
    expect(warnSpy).toHaveBeenCalledTimes(1)

    warnSpy.mockRestore()
  })

  it('returns DEFAULT_THEME for JSON array', () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    const result = parseThemeOverrides('[1, 2, 3]', 'arr-tag')

    expect(result).toEqual(DEFAULT_THEME)
    expect(warnSpy).toHaveBeenCalledTimes(1)

    warnSpy.mockRestore()
  })

  it('preserves all defaults when override is empty object', () => {
    const result = parseThemeOverrides('{}', 'empty-tag')
    expect(result).toEqual(DEFAULT_THEME)
  })
})

describe('theme-changed event', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('plugin can listen for mcms-theme-changed events', () => {
    const el = document.createElement('div')
    document.body.appendChild(el)

    injectTheme(el, DEFAULT_THEME)

    let themeChanged = false
    el.addEventListener('mcms-theme-changed', () => {
      themeChanged = true
    })

    // Simulate host dispatching theme change
    el.dispatchEvent(new CustomEvent('mcms-theme-changed', {
      detail: { colors: { primary: "oklch(0.8 0.2 300)" } },
    }))

    expect(themeChanged).toBe(true)
  })
})
