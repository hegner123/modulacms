import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/**
 * Safely extract a string from a value that may be a plain string, null,
 * or a Go sql.NullString JSON shape: { String: string, Valid: boolean }.
 */
/**
 * Rewrite a minio http://localhost:9000 URL to go through the Vite proxy
 * at /minio/ to avoid mixed-content blocking on HTTPS.
 */

const mediaBasePathname = (() => {
  const base = import.meta.env.VITE_MEDIA_BASE_URL
  if (!base) return ''
  try {
    const parsed = new URL(base)
    // Remove trailing slash for consistent prefix matching
    return parsed.pathname.replace(/\/$/, '')
  } catch {
    return ''
  }
})()

export function extractMediaPath(url: string): { dir: string; filename: string } {
  let pathname: string
  try {
    pathname = new URL(url).pathname
  } catch {
    return { dir: '', filename: url }
  }

  let relative = pathname
  if (mediaBasePathname && pathname.startsWith(mediaBasePathname)) {
    relative = pathname.slice(mediaBasePathname.length)
  }

  // Strip leading slash
  if (relative.startsWith('/')) {
    relative = relative.slice(1)
  }

  const lastSlash = relative.lastIndexOf('/')
  if (lastSlash === -1) {
    return { dir: '', filename: relative }
  }
  return {
    dir: relative.slice(0, lastSlash),
    filename: relative.slice(lastSlash + 1),
  }
}

export function nullStr(v: unknown): string | null {
  if (v == null) return null
  if (typeof v === 'string') return v || null
  if (typeof v === 'object' && 'Valid' in (v as Record<string, unknown>)) {
    const ns = v as { String: string; Valid: boolean }
    return ns.Valid ? ns.String : null
  }
  return String(v)
}
