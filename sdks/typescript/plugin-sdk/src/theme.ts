import type { ThemeTokens } from './types.js'

export const DEFAULT_THEME: ThemeTokens = {
  colors: {
    primary: "oklch(0.65 0.15 250)",
    surface: "oklch(0.98 0 0)",
    border: "oklch(0.85 0 0)",
    text: "oklch(0.2 0 0)",
    error: "oklch(0.55 0.2 25)",
    success: "oklch(0.6 0.15 145)",
    warning: "oklch(0.7 0.15 85)",
  },
  spacing: {
    xs: "0.25rem",
    sm: "0.5rem",
    md: "1rem",
    lg: "1.5rem",
    xl: "2rem",
    xl2: "3rem",
  },
  radii: {
    sm: "0.25rem",
    md: "0.5rem",
    lg: "1rem",
    full: "9999px",
  },
  typography: {
    fontFamily: "system-ui, -apple-system, sans-serif",
    fontSizeSm: "0.875rem",
    fontSizeMd: "1rem",
    fontSizeLg: "1.25rem",
    lineHeight: "1.5",
  },
}

function deepMerge(target: Record<string, unknown>, source: Record<string, unknown>): Record<string, unknown> {
  const result: Record<string, unknown> = {}
  for (const key of Object.keys(target)) {
    const tVal = target[key]
    const sVal = source[key]
    if (sVal === undefined) {
      result[key] = tVal
    } else if (
      typeof tVal === 'object' && tVal !== null && !Array.isArray(tVal) &&
      typeof sVal === 'object' && sVal !== null && !Array.isArray(sVal)
    ) {
      result[key] = deepMerge(tVal as Record<string, unknown>, sVal as Record<string, unknown>)
    } else {
      result[key] = sVal
    }
  }
  return result
}

export function injectTheme(el: HTMLElement, theme: ThemeTokens): void {
  el.style.setProperty('--mcms-color-primary', theme.colors.primary)
  el.style.setProperty('--mcms-color-surface', theme.colors.surface)
  el.style.setProperty('--mcms-color-border', theme.colors.border)
  el.style.setProperty('--mcms-color-text', theme.colors.text)
  el.style.setProperty('--mcms-color-error', theme.colors.error)
  el.style.setProperty('--mcms-color-success', theme.colors.success)
  el.style.setProperty('--mcms-color-warning', theme.colors.warning)

  el.style.setProperty('--mcms-spacing-xs', theme.spacing.xs)
  el.style.setProperty('--mcms-spacing-sm', theme.spacing.sm)
  el.style.setProperty('--mcms-spacing-md', theme.spacing.md)
  el.style.setProperty('--mcms-spacing-lg', theme.spacing.lg)
  el.style.setProperty('--mcms-spacing-xl', theme.spacing.xl)
  el.style.setProperty('--mcms-spacing-2xl', theme.spacing.xl2)

  el.style.setProperty('--mcms-radius-sm', theme.radii.sm)
  el.style.setProperty('--mcms-radius-md', theme.radii.md)
  el.style.setProperty('--mcms-radius-lg', theme.radii.lg)
  el.style.setProperty('--mcms-radius-full', theme.radii.full)

  el.style.setProperty('--mcms-font-family', theme.typography.fontFamily)
  el.style.setProperty('--mcms-font-size-sm', theme.typography.fontSizeSm)
  el.style.setProperty('--mcms-font-size-md', theme.typography.fontSizeMd)
  el.style.setProperty('--mcms-font-size-lg', theme.typography.fontSizeLg)
  el.style.setProperty('--mcms-line-height', theme.typography.lineHeight)
}

const MCMS_PROPERTIES = [
  '--mcms-color-primary', '--mcms-color-surface', '--mcms-color-border',
  '--mcms-color-text', '--mcms-color-error', '--mcms-color-success', '--mcms-color-warning',
  '--mcms-spacing-xs', '--mcms-spacing-sm', '--mcms-spacing-md',
  '--mcms-spacing-lg', '--mcms-spacing-xl', '--mcms-spacing-2xl',
  '--mcms-radius-sm', '--mcms-radius-md', '--mcms-radius-lg', '--mcms-radius-full',
  '--mcms-font-family', '--mcms-font-size-sm', '--mcms-font-size-md',
  '--mcms-font-size-lg', '--mcms-line-height',
]

export function clearTheme(el: HTMLElement): void {
  for (const prop of MCMS_PROPERTIES) {
    el.style.removeProperty(prop)
  }
}

export function parseThemeOverrides(jsonStr: string, tag: string): ThemeTokens {
  try {
    const parsed: unknown = JSON.parse(jsonStr)
    if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
      console.warn(`Invalid theme JSON on <${tag}>, using defaults: expected object`)
      return DEFAULT_THEME
    }
    return deepMerge(
      DEFAULT_THEME as unknown as Record<string, unknown>,
      parsed as Record<string, unknown>,
    ) as unknown as ThemeTokens
  } catch (e: unknown) {
    const message = e instanceof Error ? e.message : String(e)
    console.warn(`Invalid theme JSON on <${tag}>, using defaults: ${message}`)
    return DEFAULT_THEME
  }
}
