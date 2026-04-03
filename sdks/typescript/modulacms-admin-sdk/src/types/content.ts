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
export type { ContentData, ContentField, ContentRelation, ContentVersion, AdminContentVersion } from '@modulacms/types'
import type { ContentData, ContentField } from '@modulacms/types'

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

/** Parameters for moving a content data node to a new parent via `POST /contentdata/move`. */
export type MoveContentDataParams = {
  /** ID of the node to move. */
  node_id: ContentID
  /** New parent node ID, or `null` for root level. */
  new_parent_id: ContentID | null
  /** 0-indexed position within the new parent's children. */
  position: number
}

/** Response from the move content data endpoint. */
export type MoveContentDataResponse = {
  /** ID of the moved node. */
  node_id: ContentID
  /** Previous parent node ID, or `null`. */
  old_parent_id: ContentID | null
  /** New parent node ID, or `null`. */
  new_parent_id: ContentID | null
  /** Position within the new parent's children. */
  position: number
}

// ---------------------------------------------------------------------------
// Batch content update
// ---------------------------------------------------------------------------

/** Parameters for a batch content update via `POST /content/batch`. */
export type BatchContentUpdateParams = {
  /** ID of the content data node to update. */
  content_data_id: ContentID
  /** Optional content data fields to update. */
  content_data?: UpdateContentDataParams
  /** Map of field ID to value for field upserts. */
  fields?: Record<FieldID, string>
}

/** Response from the batch content update endpoint. */
export type BatchContentUpdateResponse = {
  /** ID of the content data node that was updated. */
  content_data_id: ContentID
  /** Whether the content data row was updated. */
  content_data_updated: boolean
  /** Error message if the content data update failed. */
  content_data_error?: string
  /** Number of existing fields updated. */
  fields_updated: number
  /** Number of new fields created. */
  fields_created: number
  /** Number of field operations that failed. */
  fields_failed: number
  /** Individual error messages for partial failures. */
  errors?: string[]
}

// ---------------------------------------------------------------------------
// Content tree heal
// ---------------------------------------------------------------------------

/** A single ID repair made (or that would be made in dry-run mode). */
export type HealRepair = {
  /** Primary key of the row that was repaired. */
  row_id: string
  /** Column name that contained the invalid ID. */
  column: string
  /** The original invalid value. */
  old_value: string
  /** The replacement value (`"null"` for nullable IDs cleared). */
  new_value: string
}

/** A content_field row that was created (or would be in dry-run) for a
 *  content_data node whose datatype requires the field but it was missing. */
export type MissingFieldReport = {
  /** The content_data row that was missing the field. */
  content_data_id: string
  /** The field_id from the datatype definition that was missing. */
  field_id: string
  /** Whether the missing field was actually created (false in dry-run). */
  created: boolean
}

/** A duplicate content_field row that was deleted (or would be in dry-run). */
export type DuplicateFieldReport = {
  /** The content_field_id of the duplicate row. */
  content_field_id: string
  /** The content_data row the duplicate belonged to. */
  content_data_id: string
  /** The field_id that was duplicated. */
  field_id: string
  /** Whether the duplicate was actually deleted (false in dry-run). */
  deleted: boolean
}

/** An orphaned content_field row referencing a field_id no longer in its datatype. */
export type OrphanedFieldReport = {
  content_field_id: string
  content_data_id: string
  field_id: string
  deleted: boolean
}

/** A tree pointer on a content_data row referencing a non-existent content_data_id. */
export type DanglingPointerReport = {
  content_data_id: string
  column: string
  target_id: string
  nulled: boolean
}

/** A content_data row whose route_id references a route that no longer exists. */
export type OrphanedRouteReport = {
  content_data_id: string
  route_id: string
  nulled: boolean
}

/** A root content node with no route_id set. */
export type UnroutedRootReport = {
  content_data_id: string
  datatype_id: string
  datatype_name: string
}

/** A content_data row on a route that has no _root node (inaccessible). */
export type RootlessContentReport = {
  content_data_id: string
  route_id: string
  route_slug: string
  datatype_name: string
  deleted: boolean
}

/** A content row whose author_id or published_by references a non-existent user. */
export type InvalidUserRefReport = {
  table: string
  row_id: string
  column: string
  old_value: string
  new_value: string
  repaired: boolean
}

/** A content_data_id+locale group with more than one published version. */
export type DuplicatePublishedReport = {
  content_data_id: string
  locale: string
  count: number
  kept_version_id: string
  repaired: boolean
}

/** Response from the content heal endpoint (`POST /admin/content/heal`). */
export type HealReport = {
  /** Whether the request was a dry run (no changes written). */
  dry_run: boolean
  /** Number of content_data rows scanned. */
  content_data_scanned: number
  /** Repairs made (or planned) to content_data rows. */
  content_data_repairs: HealRepair[]
  /** Number of content_field rows scanned. */
  content_field_scanned: number
  /** Repairs made (or planned) to content_field rows. */
  content_field_repairs: HealRepair[]
  /** Content fields that were missing and created (or would be in dry-run). */
  missing_fields: MissingFieldReport[]
  /** Duplicate content_field rows that were deleted (or would be in dry-run). */
  duplicate_fields: DuplicateFieldReport[]
  /** Orphaned content_field rows deleted (or would be in dry-run). */
  orphaned_fields: OrphanedFieldReport[]
  /** Tree pointers referencing non-existent content_data rows. */
  dangling_pointers: DanglingPointerReport[]
  /** Route references pointing to deleted routes. */
  orphaned_route_refs: OrphanedRouteReport[]
  /** Root content nodes with no route_id. */
  unrouted_roots: UnroutedRootReport[]
  /** Content on routes with no _root node. */
  rootless_content: RootlessContentReport[]
  /** Content rows with invalid user references. */
  invalid_user_refs: InvalidUserRefReport[]
  /** Number of duplicate-published groups found. */
  versions_scanned: number
  /** Duplicate published version groups that were repaired. */
  duplicate_published: DuplicatePublishedReport[]
}

// ---------------------------------------------------------------------------
// Content tree save (bulk creates, updates, deletes)
// ---------------------------------------------------------------------------

/**
 * A new content_data node to insert via the tree save endpoint.
 *
 * The caller supplies a client-generated ID (e.g. from `crypto.randomUUID()`).
 * The server generates a ULID and returns the mapping in
 * {@link TreeSaveResponse.id_map}. Pointer fields may reference other new
 * nodes by their client IDs — the server remaps them automatically.
 *
 * The server inherits `route_id` from the parent content node, sets `author_id`
 * from the authenticated user, and defaults `status` to `"draft"`.
 */
export type TreeNodeCreate = {
  /** Caller-generated temporary ID for this node. Must be unique within the request. */
  client_id: string
  /** Datatype to assign, or empty string for no datatype. */
  datatype_id: string
  /** Parent content node, or `null` for root-level. May be a client ID of another new node. */
  parent_id: string | null
  /** First child, or `null`. May be a client ID of another new node. */
  first_child_id: string | null
  /** Next sibling, or `null`. May be a client ID of another new node. */
  next_sibling_id: string | null
  /** Previous sibling, or `null`. May be a client ID of another new node. */
  prev_sibling_id: string | null
}

/**
 * Pointer-field changes for an existing content_data node.
 *
 * Only the four tree-pointer fields are updated; all other fields (route,
 * datatype, author, status, dates) are preserved from the existing row.
 * Pointer fields may reference new nodes by client ID — the server remaps
 * them using the ID map built during the creates phase.
 */
export type TreeNodeUpdate = {
  /** ULID of the existing node to update. */
  content_data_id: ContentID
  /** New parent, or `null` for SQL NULL. May be a client ID. */
  parent_id: string | null
  /** New first child, or `null`. May be a client ID. */
  first_child_id: string | null
  /** New next sibling, or `null`. May be a client ID. */
  next_sibling_id: string | null
  /** New previous sibling, or `null`. May be a client ID. */
  prev_sibling_id: string | null
}

/**
 * Request body for the tree save endpoint (`POST /api/v1/content/tree`).
 *
 * Atomically applies creates, deletes, and pointer-field updates to
 * content_data nodes in a single HTTP round-trip. This is the preferred
 * way to persist structural changes from a block editor or tree manipulation UI.
 *
 * Processing order: creates first (with ID remapping), then deletes, then updates.
 */
export type TreeSaveRequest = {
  /** Root content node being edited. Used to resolve route_id for new child nodes. */
  content_id: ContentID
  /** New content_data nodes to insert. */
  creates?: TreeNodeCreate[]
  /** Existing nodes whose pointer fields changed. */
  updates?: TreeNodeUpdate[]
  /** Content data IDs to remove. */
  deletes?: ContentID[]
}

/**
 * Response from the tree save endpoint.
 *
 * Always returns HTTP 200. Check {@link errors} for per-node failure messages.
 * {@link created}, {@link updated}, and {@link deleted} counts reflect only
 * successful operations.
 */
export type TreeSaveResponse = {
  /** Number of nodes successfully created. */
  created: number
  /** Number of nodes successfully updated. */
  updated: number
  /** Number of nodes successfully deleted. */
  deleted: number
  /** Maps client-supplied IDs to server-generated ULIDs. Only present when creates were included. */
  id_map?: Record<string, string>
  /** Per-node error messages for partial failures. Empty when all operations succeeded. */
  errors?: string[]
}

// ---------------------------------------------------------------------------
// Composite content create (POST /content/create)
// ---------------------------------------------------------------------------

/** Parameters for creating content with fields in a single request via `POST /content/create`. */
export type ContentCreateParams = {
  /** Parent content node ID, or `null` for root nodes. */
  parent_id?: string | null
  /** Public route this content belongs to, or `null`. */
  route_id?: string | null
  /** Datatype ID for the new content node. */
  datatype_id: string
  /** Publication lifecycle status (defaults to server default if omitted). */
  status?: string
  /** Map of field name to field value for initial field creation. */
  fields?: Record<string, string>
}

/** Response from the composite content create endpoint (`POST /content/create`). */
export type ContentCreateResponse = {
  /** The newly created content data node. */
  content_data: ContentData
  /** Content fields that were created. */
  fields: ContentField[]
  /** Number of fields successfully created. */
  fields_created: number
  /** Number of fields that failed to create. */
  fields_failed: number
  /** Error messages for partial failures. */
  errors: string[]
}

// ---------------------------------------------------------------------------
// Recursive content delete (DELETE /contentdata/?q=id&recursive=true)
// ---------------------------------------------------------------------------

/** Response from the recursive content delete endpoint. */
export type RecursiveDeleteResponse = {
  /** ID of the root node that was deleted. */
  deleted_root: string
  /** Total number of nodes deleted (including the root). */
  total_deleted: number
  /** IDs of all nodes that were deleted. */
  deleted_ids: string[]
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
