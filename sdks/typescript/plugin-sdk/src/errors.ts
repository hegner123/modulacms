export class PluginError extends Error {
  readonly _tag = "PluginError" as const

  constructor(message: string) {
    super(message)
    this.name = "PluginError"
  }
}

export class PluginApiError extends Error {
  readonly _tag = "PluginApiError" as const
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = "PluginApiError"
    this.status = status
  }
}

export function isPluginError(err: unknown): err is PluginError {
  if (err === null || err === undefined) return false
  if (typeof err !== 'object') return false
  return (err as PluginError)._tag === "PluginError"
}

export function isPluginApiError(err: unknown): err is PluginApiError {
  if (err === null || err === undefined) return false
  if (typeof err !== 'object') return false
  return (err as PluginApiError)._tag === "PluginApiError"
}
