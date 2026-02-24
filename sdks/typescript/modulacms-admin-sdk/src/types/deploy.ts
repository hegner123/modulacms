/**
 * Deploy sync types for cross-instance content synchronization.
 *
 * @module types/deploy
 */

/** Health check response from GET /api/v1/deploy/health. */
export type DeployHealthResponse = {
  /** Server status (e.g. "ok"). */
  status: string
  /** ModulaCMS version string. */
  version: string
  /** Unique node identifier for this instance. */
  node_id: string
}

/** Optional request body for POST /api/v1/deploy/export. */
export type DeployExportRequest = {
  /** Table names to export. Omit for default table set. */
  tables?: string[]
}

/** The full sync payload returned by export and consumed by import. */
export type DeploySyncPayload = Record<string, unknown>

/** Per-table error during a sync operation. */
export type DeploySyncError = {
  /** Table where the error occurred. */
  table: string
  /** Phase of the sync: "export", "validate", "truncate", "insert", or "verify". */
  phase: string
  /** Human-readable error message. */
  message: string
  /** Row ID that caused the error, if applicable. */
  row_id?: string
}

/** Result of a sync import operation. */
export type DeploySyncResult = {
  /** Whether the import completed successfully. */
  success: boolean
  /** Whether this was a dry-run (no database changes). */
  dry_run: boolean
  /** Merge strategy used (e.g. "overwrite"). */
  strategy: string
  /** Table names that were affected. */
  tables_affected: string[]
  /** Per-table row counts. */
  row_counts: Record<string, number>
  /** Path to the backup created before import, if any. */
  backup_path: string
  /** Snapshot identifier. */
  snapshot_id: string
  /** Human-readable duration string. */
  duration: string
  /** Errors encountered during the import. */
  errors?: DeploySyncError[]
  /** Warnings generated during the import. */
  warnings?: string[]
}
