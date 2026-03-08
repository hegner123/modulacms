/**
 * Config management types.
 *
 * @remarks
 * Config fields are organized into categories and have behavioral flags that
 * control how updates are applied. See {@link ConfigFieldMeta} for details on
 * the `hot_reloadable` and `sensitive` flags.
 *
 * @module types/config
 */

/**
 * Metadata for a single config field, describing its identity, category,
 * validation constraints, and behavioral flags.
 *
 * Returned by the config meta endpoint to enable dynamic form generation
 * and to inform consumers about the behavior of each field.
 */
export type ConfigFieldMeta = {
  /** The JSON key used to read and write this field (e.g. `'port'`, `'cors_allowed_origins'`). */
  json_key: string
  /** Human-readable display label for this field (e.g. `'HTTP Port'`, `'CORS Allowed Origins'`). */
  label: string
  /** Organizational category (e.g. `'server'`, `'database'`, `'media'`, `'auth'`). */
  category: string
  /**
   * Whether changes to this field take effect immediately without a server restart.
   *
   * When `true`, the new value is applied as soon as the update is persisted.
   * When `false`, the value is persisted but the running server continues using
   * the old value until restarted.
   */
  hot_reloadable: boolean
  /**
   * Whether this field contains sensitive data (passwords, secret keys, tokens).
   *
   * Sensitive fields are redacted (replaced with `"********"`) in GET responses
   * but can be written normally via the update endpoint.
   */
  sensitive: boolean
  /** Whether this field must have a non-empty value for the CMS to start. */
  required: boolean
  /** Human-readable description of this field's purpose and valid values. */
  description: string
}

/**
 * Response from `GET /api/v1/admin/config`.
 *
 * @remarks Sensitive fields are redacted and appear as `"********"`.
 */
export type ConfigGetResponse = {
  /** The current config as a flat key-value object. Keys are the `json_key` values from field metadata. */
  config: Record<string, unknown>
}

/**
 * Response from `PATCH /api/v1/admin/config`.
 *
 * @remarks
 * After a successful update, check `restart_required` to determine if the
 * server needs to be restarted for all changes to take effect.
 */
export type ConfigUpdateResponse = {
  /** Whether the update was persisted successfully. */
  ok: boolean
  /** The full config after applying the update (sensitive fields redacted). */
  config: Record<string, unknown>
  /**
   * List of field keys that were updated but require a server restart to take effect.
   * Empty or absent if all updated fields are hot-reloadable.
   */
  restart_required?: string[]
  /** Non-fatal warnings generated during the update (e.g. deprecated field usage). */
  warnings?: string[]
}

/**
 * Response from `GET /api/v1/admin/config/meta`.
 *
 * Provides the full field metadata registry and available category names
 * for building config management UIs.
 */
export type ConfigMetaResponse = {
  /** All config field metadata entries. */
  fields: ConfigFieldMeta[]
  /** Distinct category names across all fields, for building category filters. */
  categories: string[]
}
