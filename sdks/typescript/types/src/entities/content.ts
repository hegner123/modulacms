/**
 * Public-facing content entity types.
 * These represent the published content tree visible to site visitors.
 *
 * @module entities/content
 */

import type {
  AdminContentID,
  AdminContentVersionID,
  ContentFieldID,
  ContentID,
  ContentRelationID,
  ContentVersionID,
  DatatypeID,
  FieldID,
  RouteID,
  UserID,
} from '../ids.js'
import type { ContentStatus } from '../enums.js'

/**
 * A public content data node in the content tree.
 * Uses a linked-list tree structure (parent, first child, siblings).
 */
export type ContentData = {
  /** Unique identifier for this content node. */
  content_data_id: ContentID
  /** Parent node ID, or `null` for root nodes. */
  parent_id: ContentID | null
  /** First child node ID, or `null` if no children. */
  first_child_id: string | null
  /** Next sibling node ID, or `null` if last sibling. */
  next_sibling_id: string | null
  /** Previous sibling node ID, or `null` if first sibling. */
  prev_sibling_id: string | null
  /** The public route this content belongs to, or `null`. */
  route_id: RouteID | null
  /** The datatype defining this content's schema, or `null`. */
  datatype_id: DatatypeID | null
  /** ID of the user who created this content, or `null`. */
  author_id: UserID | null
  /** Publication lifecycle status. */
  status: ContentStatus
  /** ISO 8601 timestamp when this content was last published, or `null`. */
  published_at?: string | null
  /** User ID of the person who last published this content, or `null`. */
  published_by?: string | null
  /** ISO 8601 timestamp for scheduled publication, or `null`. */
  publish_at?: string | null
  /** Monotonically increasing revision counter. */
  revision: number
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * A snapshot version of a content data node at a point in time.
 */
export type ContentVersion = {
  /** Unique identifier for this version. */
  content_version_id: ContentVersionID
  /** The content data node this version belongs to. */
  content_data_id: ContentID
  /** Sequential version number. */
  version_number: number
  /** Locale of the content at snapshot time. */
  locale: string
  /** JSON-serialized snapshot of content data and fields. */
  snapshot: string
  /** What triggered this version (e.g. `"publish"`, `"manual"`, `"auto"`). */
  trigger: string
  /** Human-readable label for this version. */
  label: string
  /** Whether this version represents a published state. */
  published: boolean
  /** User ID of the person who published this version, or `null`. */
  published_by?: string | null
  /** ISO 8601 creation timestamp. */
  date_created: string
}

/**
 * A snapshot version of an admin content data node at a point in time.
 */
export type AdminContentVersion = {
  /** Unique identifier for this admin version. */
  admin_content_version_id: AdminContentVersionID
  /** The admin content data node this version belongs to. */
  admin_content_data_id: AdminContentID
  /** Sequential version number. */
  version_number: number
  /** Locale of the content at snapshot time. */
  locale: string
  /** JSON-serialized snapshot of content data and fields. */
  snapshot: string
  /** What triggered this version (e.g. `"publish"`, `"manual"`, `"auto"`). */
  trigger: string
  /** Human-readable label for this version. */
  label: string
  /** Whether this version represents a published state. */
  published: boolean
  /** User ID of the person who published this version, or `null`. */
  published_by?: string | null
  /** ISO 8601 creation timestamp. */
  date_created: string
}

/**
 * A field value belonging to a public content data node.
 * Links a content node to a specific field definition and stores the value.
 */
export type ContentField = {
  /** Unique identifier for this field value. */
  content_field_id: ContentFieldID
  /** The public route this field belongs to, or `null`. */
  route_id: RouteID | null
  /** The content data node this field value belongs to, or `null`. */
  content_data_id: ContentID | null
  /** The field definition this value corresponds to, or `null`. */
  field_id: FieldID | null
  /** The stored value as a serialized string. */
  field_value: string
  /** Locale code for this field value (e.g. `"en"`, `"fr"`). */
  locale: string
  /** ID of the user who set this value, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * A directional relation between two public content nodes via a specific field.
 */
export type ContentRelation = {
  /** Unique identifier for this relation. */
  content_relation_id: ContentRelationID
  /** The content node that owns this relation. */
  source_content_id: ContentID
  /** The content node being referenced. */
  target_content_id: ContentID
  /** The field definition that established this relation. */
  field_id: FieldID
  /** Display ordering position. */
  sort_order: number
  /** ISO 8601 creation timestamp. */
  date_created: string
}
