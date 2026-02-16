/**
 * Config management types.
 * @module types/config
 */

/** A single config field metadata entry. */
export type ConfigFieldMeta = {
  json_key: string
  label: string
  category: string
  hot_reloadable: boolean
  sensitive: boolean
  required: boolean
  description: string
}

/** Response from GET /api/v1/admin/config. */
export type ConfigGetResponse = {
  config: Record<string, unknown>
}

/** Response from PATCH /api/v1/admin/config. */
export type ConfigUpdateResponse = {
  ok: boolean
  config: Record<string, unknown>
  restart_required?: string[]
  warnings?: string[]
}

/** Response from GET /api/v1/admin/config/meta. */
export type ConfigMetaResponse = {
  fields: ConfigFieldMeta[]
  categories: string[]
}
