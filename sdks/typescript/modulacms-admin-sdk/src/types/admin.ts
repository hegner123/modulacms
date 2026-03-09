/**
 * Admin-side entity types and their create/update parameter shapes.
 *
 * ModulaCMS maintains two parallel content systems: **public** (user-facing content
 * served via the delivery API) and **admin** (internal content used by the CMS admin
 * interface itself, such as the admin panel layout, dashboard widgets, and system pages).
 *
 * The admin content system uses the same tree-based structure as public content but
 * with a separate set of tables (`admin_content_data`, `admin_datatypes`, `admin_fields`,
 * `admin_routes`). This separation ensures that admin UI customizations never interfere
 * with published content and vice versa.
 *
 * @remarks
 * **Tree structure:** Admin content is organized as an ordered tree using sibling
 * pointers. Each node has:
 * - `parent_id` -- pointer to the parent node (null for root nodes)
 * - `first_child_id` -- pointer to the leftmost child node
 * - `next_sibling_id` -- pointer to the next sibling (right neighbor)
 * - `prev_sibling_id` -- pointer to the previous sibling (left neighbor)
 *
 * This linked-list structure enables O(1) insertions, deletions, and reordering
 * without updating sort_order on every sibling. To traverse children of a node,
 * follow `first_child_id` then walk `next_sibling_id` until null.
 *
 * **Admin routes:** Admin routes serve as top-level containers (namespaces) for
 * admin content trees. Each admin route has its own independent tree of content nodes.
 * The `admin_route_id` field on content nodes identifies which route (and therefore
 * which tree) a node belongs to.
 *
 * @module types/admin
 */

import type {
  AdminContentFieldID,
  AdminContentID,
  AdminContentRelationID,
  AdminDatatypeID,
  AdminFieldID,
  AdminRouteID,
  ContentStatus,
  FieldType,
  Slug,
  UserID,
} from './common.js'

// ---------------------------------------------------------------------------
// Entity types
// ---------------------------------------------------------------------------

/**
 * An admin-side route that serves as the top-level container for admin content trees.
 * Identified by both a unique ID and a URL slug.
 */
export type AdminRoute = {
  /** Unique identifier for this admin route. */
  admin_route_id: AdminRouteID
  /** URL-safe slug for this route. */
  slug: Slug
  /** Human-readable title of the route. */
  title: string
  /** Numeric status flag (0 = inactive, 1 = active). */
  status: number
  /** ID of the user who created this route, or `null` if system-generated. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * An admin content data node in the content tree.
 *
 * Uses a linked-list tree structure with sibling pointers for O(1) reordering.
 * To enumerate a node's children: follow {@link first_child_id}, then walk
 * {@link next_sibling_id} until `null`. To move backward, use {@link prev_sibling_id}.
 *
 * @remarks
 * When creating or moving nodes, the server automatically maintains the sibling
 * pointer chain. Use the `move` and `reorder` endpoints rather than manually
 * updating pointer fields, which risks breaking the linked list invariants.
 */
export type AdminContentData = {
  /** Unique identifier for this content node. */
  admin_content_data_id: AdminContentID
  /**
   * Parent node ID, or `null` for root-level nodes.
   *
   * Root nodes are direct children of the admin route (no parent in the tree).
   */
  parent_id: AdminContentID | null
  /**
   * ID of this node's first (leftmost) child, or `null` if the node is a leaf.
   *
   * To get all children, follow this pointer then walk {@link next_sibling_id}.
   */
  first_child_id: string | null
  /**
   * ID of the next sibling in the ordered child list, or `null` if this is the last sibling.
   *
   * Together with {@link prev_sibling_id}, forms a doubly-linked list of siblings
   * under the same parent.
   */
  next_sibling_id: string | null
  /**
   * ID of the previous sibling in the ordered child list, or `null` if this is the first sibling.
   *
   * Together with {@link next_sibling_id}, forms a doubly-linked list of siblings
   * under the same parent.
   */
  prev_sibling_id: string | null
  /**
   * The admin route (tree namespace) this content belongs to.
   *
   * Every content node belongs to exactly one admin route. The route determines
   * which admin content tree this node is part of.
   */
  admin_route_id: AdminRouteID
  /** The datatype that defines this content's schema, or `null` if untyped. */
  admin_datatype_id: AdminDatatypeID | null
  /** ID of the user who created this content, or `null`. */
  author_id: UserID | null
  /** Publication lifecycle status. */
  status: ContentStatus
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * A field value belonging to an admin content data node.
 * Links a content node to a specific field definition and stores the value.
 */
export type AdminContentField = {
  /** Unique identifier for this field value. */
  admin_content_field_id: AdminContentFieldID
  /** The admin route this field belongs to, or `null`. */
  admin_route_id: AdminRouteID | null
  /** The content data node this field value belongs to. */
  admin_content_data_id: string
  /** The field definition this value corresponds to, or `null`. */
  admin_field_id: AdminFieldID | null
  /** The stored value as a string (serialized based on field type). */
  admin_field_value: string
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
 * An admin-side datatype definition that describes the schema of a content type.
 */
export type AdminDatatype = {
  /** Unique identifier for this datatype. */
  admin_datatype_id: AdminDatatypeID
  /** Parent content ID for hierarchical datatypes, or `null`. */
  parent_id: AdminContentID | null
  /** Display ordering position. */
  sort_order: number
  /** Machine-readable name used as JSON key. If empty, derived from label. */
  name: string
  /** Human-readable label for this datatype. */
  label: string
  /** Datatype category (e.g. `'page'`, `'component'`). */
  type: string
  /** ID of the user who created this datatype, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * An admin-side field definition that belongs to a datatype.
 * Defines the name, type, validation, and UI configuration of a content field.
 */
export type AdminField = {
  /** Unique identifier for this field definition. */
  admin_field_id: AdminFieldID
  /** Parent datatype ID this field belongs to, or `null`. */
  parent_id: AdminDatatypeID | null
  /** Display ordering position within the datatype. */
  sort_order: number
  /** Machine-readable name used as JSON key. If empty, derived from label. */
  name: string
  /** Human-readable field label. */
  label: string
  /** Additional field data (JSON-encoded metadata). */
  data: string
  /** Validation rules (JSON-encoded). */
  validation: string
  /** UI widget configuration (JSON-encoded). */
  ui_config: string
  /** The data type of this field. */
  type: FieldType
  /** Whether this field is translatable (0 = no, 1 = yes). */
  translatable: number
  /** Role names that can access this field, or `null` for unrestricted. */
  roles: string[] | null
  /** ID of the user who created this field, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modification timestamp. */
  date_modified: string
}

/**
 * A directional relation between two admin content nodes via a specific field.
 */
export type AdminContentRelation = {
  /** Unique identifier for this relation. */
  admin_content_relation_id: AdminContentRelationID
  /** The content node that owns this relation. */
  source_content_id: AdminContentID
  /** The content node being referenced. */
  target_content_id: AdminContentID
  /** The field definition that established this relation. */
  admin_field_id: AdminFieldID
  /** Display ordering position. */
  sort_order: number
  /** ISO 8601 creation timestamp. */
  date_created: string
}

// ---------------------------------------------------------------------------
// Create params
// ---------------------------------------------------------------------------

/**
 * Parameters for creating a new admin route via `POST /adminroutes`.
 *
 * All fields are required. The `slug` must be unique across all admin routes
 * and will serve as the URL path segment for this route's content tree.
 */
export type CreateAdminRouteParams = {
  /** URL-safe slug for the new route. Must be unique across admin routes. */
  slug: Slug
  /** Human-readable title. */
  title: string
  /** Numeric status flag (0 = inactive, 1 = active). */
  status: number
  /** Author user ID, or `null` for system-generated routes. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. Typically set to the current time. */
  date_created: string
  /** ISO 8601 modification timestamp. Typically set to the current time on creation. */
  date_modified: string
}

/**
 * Parameters for creating a new admin content data node via `POST /admincontentdatas`.
 *
 * @remarks
 * The tree pointer fields (`first_child_id`, `next_sibling_id`, `prev_sibling_id`)
 * should typically be set to `null` on creation. The server will update sibling
 * pointers on adjacent nodes to maintain the linked list. For positioning, prefer
 * using the `move` endpoint after creation rather than manually setting pointers.
 *
 * The `admin_route_id` is required and determines which admin content tree this
 * node belongs to.
 */
export type CreateAdminContentDataParams = {
  /** Parent node ID, or `null` to create a root-level node. */
  parent_id: AdminContentID | null
  /** First child node ID, or `null` (typically `null` for new nodes). */
  first_child_id: string | null
  /** Next sibling node ID, or `null` (typically `null` for new nodes). */
  next_sibling_id: string | null
  /** Previous sibling node ID, or `null` (typically `null` for new nodes). */
  prev_sibling_id: string | null
  /** Admin route this content belongs to. Required -- determines the tree namespace. */
  admin_route_id: AdminRouteID
  /** Datatype ID defining this content's schema, or `null` for untyped nodes. */
  admin_datatype_id: AdminDatatypeID | null
  /** Author user ID, or `null` for system-generated nodes. */
  author_id: UserID | null
  /** Initial publication status (e.g. `'draft'`, `'published'`). */
  status: ContentStatus
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/**
 * Parameters for creating a new admin content field value via `POST /admincontentfields`.
 *
 * Each field value links a content node to a specific field definition from its
 * datatype and stores the serialized value. Multiple field values can exist for the
 * same field definition if the content has locale variants.
 */
export type CreateAdminContentFieldParams = {
  /** Admin route this field belongs to, or `null`. Should match the content node's route. */
  admin_route_id: AdminRouteID | null
  /** Content data node this field value belongs to. Required. */
  admin_content_data_id: string
  /** Field definition this value corresponds to, or `null` for freeform fields. */
  admin_field_id: AdminFieldID | null
  /** The field value serialized as a string. Format depends on the field's type. */
  admin_field_value: string
  /** Author user ID, or `null` for system-generated values. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/**
 * Parameters for creating a new admin datatype definition via `POST /admindatatypes`.
 *
 * Admin datatypes define the schema for admin content nodes (what fields they have,
 * what type of content they represent). This is distinct from public datatypes.
 */
export type CreateAdminDatatypeParams = {
  /** Parent content ID for hierarchical datatype grouping, or `null` for top-level. */
  parent_id: AdminContentID | null
  /** Display ordering position among peer datatypes. Lower values sort first. */
  sort_order: number
  /** Machine-readable name used as JSON key. If empty string, the server derives it from label. */
  name: string
  /** Human-readable label displayed in the admin UI. */
  label: string
  /** Datatype category (e.g. `'page'`, `'component'`). Controls how the CMS treats this type. */
  type: string
  /** Author user ID, or `null` for system-defined datatypes. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/**
 * Parameters for creating a new admin field definition via `POST /adminfields`.
 *
 * Fields define the individual data entry points within a datatype. Each field
 * has a type that determines how its value is stored, validated, and rendered.
 *
 * @remarks
 * The `data`, `validation`, and `ui_config` fields accept JSON-encoded strings.
 * Pass `'{}'` for empty/default configuration. The `roles` field is optional and
 * defaults to unrestricted access when omitted or `null`.
 */
export type CreateAdminFieldParams = {
  /** Parent datatype ID this field belongs to, or `null`. Required for most use cases. */
  parent_id: AdminDatatypeID | null
  /** Display ordering position within the datatype. Lower values sort first. */
  sort_order: number
  /** Machine-readable name used as JSON key. If empty string, the server derives it from label. */
  name: string
  /** Human-readable field label displayed in the admin UI. */
  label: string
  /** Additional field metadata (JSON-encoded string). Pass `'{}'` for defaults. */
  data: string
  /** Validation rules (JSON-encoded string, e.g. `'{"required":true,"maxLength":255}'`). */
  validation: string
  /** UI widget configuration (JSON-encoded string, e.g. `'{"placeholder":"Enter title"}'`). */
  ui_config: string
  /** The data type of this field (e.g. `'text'`, `'number'`, `'richtext'`, `'media'`). */
  type: FieldType
  /** Role names that can access this field, or `null`/omitted for unrestricted access. */
  roles?: string[] | null
  /** Author user ID, or `null` for system-defined fields. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

// ---------------------------------------------------------------------------
// Update params
// ---------------------------------------------------------------------------

/**
 * Parameters for updating an admin route via `PUT /adminroutes/`.
 *
 * This is a full replacement update -- all fields must be provided.
 * The `slug_2` field carries the current slug for the WHERE clause,
 * allowing the `slug` field to be changed (renamed) in the same operation.
 *
 * @remarks
 * To rename a route's slug, set `slug` to the new value and `slug_2` to the
 * current value. If not renaming, both should be the same value.
 */
export type UpdateAdminRouteParams = {
  /** New slug value. Set to a different value than `slug_2` to rename the route. */
  slug: Slug
  /** Updated title. */
  title: string
  /** Updated status flag (0 = inactive, 1 = active). */
  status: number
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp (preserved from original record). */
  date_created: string
  /** ISO 8601 modification timestamp (set to current time). */
  date_modified: string
  /** Current slug used to locate the record (WHERE clause). Must match the existing slug. */
  slug_2: Slug
}

/**
 * Parameters for updating an admin content data node via `PUT /admincontentdatas/`.
 *
 * This is a full replacement update. All fields must be provided.
 *
 * @remarks
 * **Prefer `move` and `reorder` endpoints** over manually updating tree pointer
 * fields (`parent_id`, `first_child_id`, `next_sibling_id`, `prev_sibling_id`).
 * Directly updating pointers without also updating adjacent nodes will break
 * the doubly-linked sibling list invariants.
 */
export type UpdateAdminContentDataParams = {
  /** ID of the content node to update. */
  admin_content_data_id: AdminContentID
  /** Updated parent node ID, or `null` for root-level. */
  parent_id: AdminContentID | null
  /** Updated first child ID, or `null`. Prefer the `move` endpoint over manual changes. */
  first_child_id: string | null
  /** Updated next sibling ID, or `null`. Prefer the `reorder` endpoint over manual changes. */
  next_sibling_id: string | null
  /** Updated previous sibling ID, or `null`. Prefer the `reorder` endpoint over manual changes. */
  prev_sibling_id: string | null
  /** Admin route this content belongs to. Cannot be changed after creation. */
  admin_route_id: AdminRouteID
  /** Updated datatype ID, or `null`. */
  admin_datatype_id: AdminDatatypeID | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** Updated publication status. */
  status: ContentStatus
  /** ISO 8601 creation timestamp (preserved from original record). */
  date_created: string
  /** ISO 8601 modification timestamp (set to current time). */
  date_modified: string
}

/** Parameters for updating an admin content field value via `PUT /admincontentfields/`. */
export type UpdateAdminContentFieldParams = {
  /** ID of the field value to update. */
  admin_content_field_id: AdminContentFieldID
  /** Admin route, or `null`. */
  admin_route_id: AdminRouteID | null
  /** Content data node this field belongs to. */
  admin_content_data_id: string
  /** Field definition, or `null`. */
  admin_field_id: AdminFieldID | null
  /** Updated field value. */
  admin_field_value: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for updating an admin datatype definition via `PUT /admindatatypes/`. */
export type UpdateAdminDatatypeParams = {
  /** ID of the datatype to update. */
  admin_datatype_id: AdminDatatypeID
  /** Updated parent content ID, or `null`. */
  parent_id: AdminContentID | null
  /** Display ordering position. */
  sort_order: number
  /** Machine-readable name used as JSON key. If empty, derived from label. */
  name: string
  /** Updated label. */
  label: string
  /** Updated category type. */
  type: string
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

/** Parameters for updating an admin field definition via `PUT /adminfields/`. */
export type UpdateAdminFieldParams = {
  /** ID of the field to update. */
  admin_field_id: AdminFieldID
  /** Updated parent datatype ID, or `null`. */
  parent_id: AdminDatatypeID | null
  /** Display ordering position within the datatype. */
  sort_order: number
  /** Machine-readable name used as JSON key. If empty, derived from label. */
  name: string
  /** Updated label. */
  label: string
  /** Updated metadata (JSON-encoded). */
  data: string
  /** Updated validation rules (JSON-encoded). */
  validation: string
  /** Updated UI configuration (JSON-encoded). */
  ui_config: string
  /** Updated field type. */
  type: FieldType
  /** Role names that can access this field, or `null` for unrestricted. */
  roles?: string[] | null
  /** Author user ID, or `null`. */
  author_id: UserID | null
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 modification timestamp. */
  date_modified: string
}

// ---------------------------------------------------------------------------
// Admin content data move/reorder
// ---------------------------------------------------------------------------

/** Parameters for moving an admin content data node to a new parent via `POST /admincontentdatas/move`. */
export type MoveAdminContentDataParams = {
  /** ID of the node to move. */
  node_id: AdminContentID
  /** New parent node ID, or `null` for root level. */
  new_parent_id: AdminContentID | null
  /** 0-indexed position within the new parent's children. */
  position: number
}

/** Response from the move admin content data endpoint. */
export type MoveAdminContentDataResponse = {
  /** ID of the moved node. */
  node_id: AdminContentID
  /** Previous parent node ID, or `null`. */
  old_parent_id: AdminContentID | null
  /** New parent node ID, or `null`. */
  new_parent_id: AdminContentID | null
  /** Position within the new parent's children. */
  position: number
}

/** Parameters for reordering admin content data siblings via `POST /admincontentdatas/reorder`. */
export type ReorderAdminContentDataParams = {
  /** Parent node ID, or `null` for root-level siblings. */
  parent_id: AdminContentID | null
  /** Ordered list of sibling admin content data IDs in the desired sequence. */
  ordered_ids: AdminContentID[]
}

/** Response from the reorder admin content data endpoint. */
export type ReorderAdminContentDataResponse = {
  /** Number of nodes updated. */
  updated: number
  /** Parent node ID, or `null`. */
  parent_id: AdminContentID | null
}
