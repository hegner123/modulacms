import type { PluginConfig } from './types.js'

let apiKey: string | null = null

export function configure(config: PluginConfig): void {
  apiKey = config.apiKey
}

export function resetConfig(): void {
  apiKey = null
}

export function getApiKey(): string | null {
  return apiKey
}
