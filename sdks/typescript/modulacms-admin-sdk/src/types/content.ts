/**
 * Public-facing content entity types and their create/update parameter shapes.
 * Entity types are re-exported from @modulacms/types; param types are local.
 *
 * @module types/content
 */

import type {
  ContentFieldID,
  ContentID,
  ContentStatus,
  DatatypeID,
  FieldID,
  RouteID,
  UserID,
} from './common.js'

// Re-export shared entity types
export type { ContentData, ContentField, ContentRelation } from '@modulacms/types'

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/** Parameters for creating a new public content data node via `POST /contentdata`. */
export type CreateContentDataParams = {
  /** Public route this content belongs to, or `null`. */
  route_id: RouteID | null
  /** Parent node ID, or `null` for root nodes. */
  parent_id: ContentID | null
  /** First child node ID, or `null`. */
  first_child_id: string | null
  /** Next sibling node ID, or `null`. */
  next_sibling_id: string | null
  /** Previous sibling node ID, or `null`. */
  prev_sibling_id: string | null
  /** Datatype ID, or `null`. */
  datatype_id: DatatypeID | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** Publication lifecycle status. */
  status: ContentStatus
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for creating a new public content field value via `POST /contentfields`. */
export type CreateContentFieldParams = {
  /** Public route, or `null`. */
  route_id: RouteID | null
  /** Content data node this field belongs to, or `null`. */
  content_data_id: ContentID | null
  /** Field definition, or `null`. */
  field_id: FieldID | null
  /** The field value as a serialized string. */
  field_value: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/** Parameters for updating a public content data node via `PUT /contentdata/`. */
export type UpdateContentDataParams = {
  /** ID of the content node to update. */
  content_data_id: ContentID
  /** Updated parent node ID, or `null`. */
  parent_id: ContentID | null
  /** Updated first child ID, or `null`. */
  first_child_id: string | null
  /** Updated next sibling ID, or `null`. */
  next_sibling_id: string | null
  /** Updated previous sibling ID, or `null`. */
  prev_sibling_id: string | null
  /** Updated route, or `null`. */
  route_id: RouteID | null
  /** Updated datatype, or `null`. */
  datatype_id: DatatypeID | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** Updated publication status. */
  status: ContentStatus
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for reordering content data siblings via `POST /contentdata/reorder`. */
export type ReorderContentDataParams = {
  /** Parent node ID, or `null` for root-level siblings. */
  parent_id: ContentID | null
  /** Ordered list of sibling content data IDs in the desired sequence. */
  ordered_ids: ContentID[]
}

/** Response from the reorder content data endpoint. */
export type ReorderContentDataResponse = {
  /** Number of nodes updated. */
  updated: number
  /** Parent node ID, or `null`. */
  parent_id: ContentID | null
}

/** Parameters for updating a public content field value via `PUT /contentfields/`. */
export type UpdateContentFieldParams = {
  /** ID of the field value to update. */
  content_field_id: ContentFieldID
  /** Updated route, or `null`. */
  route_id: RouteID | null
  /** Updated content data node, or `null`. */
  content_data_id: ContentID | null
  /** Updated field definition, or `null`. */
  field_id: FieldID | null
  /** Updated field value. */
  field_value: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}
