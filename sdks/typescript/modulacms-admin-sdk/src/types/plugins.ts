/**
 * Type definitions for the plugin management, route approval, and hook approval APIs.
 *
 * @module types/plugins
 */

// ---------------------------------------------------------------------------
// Plugin list / info
// ---------------------------------------------------------------------------

/** Summary item returned by the plugin list endpoint. */
export type PluginListItem = {
  name: string
  version: string
  description: string
  state: string
  circuit_breaker_state?: string
}

/** Schema drift entry describing a missing or extra column. */
export type DriftEntry = {
  table: string
  kind: 'missing' | 'extra'
  column: string
}

/** Detailed plugin information returned by the get endpoint. */
export type PluginInfo = {
  name: string
  version: string
  description: string
  author?: string
  license?: string
  state: string
  failed_reason?: string
  circuit_breaker_state?: string
  circuit_breaker_errors?: number
  vms_available: number
  vms_total: number
  dependencies?: string[]
  schema_drift?: DriftEntry[]
}

// ---------------------------------------------------------------------------
// Action responses
// ---------------------------------------------------------------------------

/** Response from reload action. */
export type PluginActionResponse = {
  ok: true
  plugin: string
}

/** Response from enable/disable actions (includes new state). */
export type PluginStateResponse = PluginActionResponse & {
  state: string
}

// ---------------------------------------------------------------------------
// Cleanup
// ---------------------------------------------------------------------------

/** Response from cleanup dry-run (GET). */
export type CleanupDryRunResponse = {
  orphaned_tables: string[]
  count: number
  action: 'dry_run'
}

/** Parameters for cleanup drop (POST). */
export type CleanupDropParams = {
  confirm: true
  tables: string[]
}

/** Response from cleanup drop (POST). */
export type CleanupDropResponse = {
  dropped: string[]
  count: number
}

// ---------------------------------------------------------------------------
// Route approval
// ---------------------------------------------------------------------------

/** A registered plugin route with its approval status. */
export type PluginRoute = {
  plugin: string
  method: string
  path: string
  public: boolean
  approved: boolean
}

/** Identifies a specific route for approval/revocation. */
export type RouteApprovalItem = {
  plugin: string
  method: string
  path: string
}

// ---------------------------------------------------------------------------
// Hook approval
// ---------------------------------------------------------------------------

/** A registered plugin hook with its approval status. */
export type PluginHook = {
  plugin_name: string
  event: string
  table: string
  priority: number
  approved: boolean
  is_wildcard: boolean
}

/** Identifies a specific hook for approval/revocation. */
export type HookApprovalItem = {
  plugin: string
  event: string
  table: string
}
