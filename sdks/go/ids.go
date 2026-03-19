package modula

import "time"

// Branded ID Types
//
// All entity IDs in the ModulaCMS Go SDK are 26-character ULIDs (Universally
// Unique Lexicographically Sortable Identifiers) wrapped in distinct Go types.
// This provides compile-time type safety: a [ContentID] cannot be accidentally
// passed where a [UserID] is expected, even though both are string-based.
//
// Each ID type implements two methods:
//
//   - String() string -- returns the raw 26-character ULID string.
//   - IsZero() bool   -- returns true if the ID is empty (unset).
//
// IDs are generated server-side and returned in API responses. Client code
// typically receives them from create/list/get operations and passes them
// back in update/delete calls. The ULID format encodes a millisecond-precision
// timestamp in its first 10 characters, making IDs naturally sortable by
// creation time.
//
// For working with timestamps directly, see [Timestamp], [NewTimestamp], and
// [TimestampNow].

// --- Content IDs ---

// ContentID identifies a content entry in the CMS.
// Content entries hold the user-authored data (field values) for a given datatype.
type ContentID string

// String returns the raw ULID string for this content ID.
func (id ContentID) String() string { return string(id) }

// IsZero reports whether the content ID is empty (unset).
func (id ContentID) IsZero() bool { return id == "" }

// ContentFieldID identifies a content field value within a content entry.
// Each content entry stores one or more field values; this ID references a
// specific field-value row.
type ContentFieldID string

// String returns the raw ULID string for this content field ID.
func (id ContentFieldID) String() string { return string(id) }

// IsZero reports whether the content field ID is empty (unset).
func (id ContentFieldID) IsZero() bool { return id == "" }

// ContentRelationID identifies a relationship between two content entries.
// Content relations link parent and child nodes in the content tree.
type ContentRelationID string

// String returns the raw ULID string for this content relation ID.
func (id ContentRelationID) String() string { return string(id) }

// IsZero reports whether the content relation ID is empty (unset).
func (id ContentRelationID) IsZero() bool { return id == "" }

// --- Admin Content IDs ---
//
// Admin content IDs mirror the public content IDs but reference entries in the
// admin-only content tables. The admin content system provides a separate
// workspace for CMS configuration content (e.g., navigation menus, settings)
// that is managed through the admin panel rather than the public API.

// AdminContentID identifies an admin content entry.
// Admin content is CMS-internal configuration content, separate from
// user-facing published content.
type AdminContentID string

// String returns the raw ULID string for this admin content ID.
func (id AdminContentID) String() string { return string(id) }

// IsZero reports whether the admin content ID is empty (unset).
func (id AdminContentID) IsZero() bool { return id == "" }

// AdminContentFieldID identifies a field value within an admin content entry.
type AdminContentFieldID string

// String returns the raw ULID string for this admin content field ID.
func (id AdminContentFieldID) String() string { return string(id) }

// IsZero reports whether the admin content field ID is empty (unset).
func (id AdminContentFieldID) IsZero() bool { return id == "" }

// AdminContentRelationID identifies a relationship between two admin content entries.
type AdminContentRelationID string

// String returns the raw ULID string for this admin content relation ID.
func (id AdminContentRelationID) String() string { return string(id) }

// IsZero reports whether the admin content relation ID is empty (unset).
func (id AdminContentRelationID) IsZero() bool { return id == "" }

// --- Schema IDs ---
//
// Schema IDs reference the content model definitions. Datatypes define the
// structure of content (like "BlogPost" or "Product"), fields define the
// individual properties within a datatype (like "title" or "price"), and
// field types define the kind of input a field accepts (text, number, media, etc.).

// DatatypeID identifies a datatype definition in the content schema.
// Datatypes are the top-level content model templates (e.g., "BlogPost",
// "Product", "Page") that determine which fields a content entry contains.
type DatatypeID string

// String returns the raw ULID string for this datatype ID.
func (id DatatypeID) String() string { return string(id) }

// IsZero reports whether the datatype ID is empty (unset).
func (id DatatypeID) IsZero() bool { return id == "" }

// FieldID identifies a field definition in the content schema.
// Fields represent individual properties on a datatype, such as "title",
// "body", or "featured_image".
type FieldID string

// String returns the raw ULID string for this field ID.
func (id FieldID) String() string { return string(id) }

// IsZero reports whether the field ID is empty (unset).
func (id FieldID) IsZero() bool { return id == "" }

// FieldTypeID identifies a field type definition.
// Field types describe the kind of input a field accepts (e.g., text, rich_text,
// number, boolean, media, reference). The CMS ships with built-in field types
// and plugins can register additional ones.
type FieldTypeID string

// String returns the raw ULID string for this field type ID.
func (id FieldTypeID) String() string { return string(id) }

// IsZero reports whether the field type ID is empty (unset).
func (id FieldTypeID) IsZero() bool { return id == "" }

// DatatypeFieldID identifies a row in the datatype-field junction table.
// This links a specific field definition to a specific datatype, establishing
// which fields belong to which datatype along with ordering and configuration.
type DatatypeFieldID string

// String returns the raw ULID string for this datatype-field junction ID.
func (id DatatypeFieldID) String() string { return string(id) }

// IsZero reports whether the datatype-field junction ID is empty (unset).
func (id DatatypeFieldID) IsZero() bool { return id == "" }

// --- Admin Schema IDs ---
//
// Admin schema IDs mirror the public schema IDs but reference the admin-only
// content model. The admin schema defines datatypes and fields used for
// CMS-internal configuration content, separate from the user-defined public schema.

// AdminDatatypeID identifies a datatype in the admin content schema.
// Admin datatypes define the structure of CMS configuration content
// (e.g., navigation items, site settings).
type AdminDatatypeID string

// String returns the raw ULID string for this admin datatype ID.
func (id AdminDatatypeID) String() string { return string(id) }

// IsZero reports whether the admin datatype ID is empty (unset).
func (id AdminDatatypeID) IsZero() bool { return id == "" }

// AdminFieldID identifies a field definition in the admin content schema.
type AdminFieldID string

// String returns the raw ULID string for this admin field ID.
func (id AdminFieldID) String() string { return string(id) }

// IsZero reports whether the admin field ID is empty (unset).
func (id AdminFieldID) IsZero() bool { return id == "" }

// AdminFieldTypeID identifies a field type in the admin content schema.
type AdminFieldTypeID string

// String returns the raw ULID string for this admin field type ID.
func (id AdminFieldTypeID) String() string { return string(id) }

// IsZero reports whether the admin field type ID is empty (unset).
func (id AdminFieldTypeID) IsZero() bool { return id == "" }

// AdminDatatypeFieldID identifies a row in the admin datatype-field junction table.
// This links a field definition to an admin datatype.
type AdminDatatypeFieldID string

// String returns the raw ULID string for this admin datatype-field junction ID.
func (id AdminDatatypeFieldID) String() string { return string(id) }

// IsZero reports whether the admin datatype-field junction ID is empty (unset).
func (id AdminDatatypeFieldID) IsZero() bool { return id == "" }

// --- Media IDs ---

// MediaID identifies a media asset (image, video, document, etc.) stored in the CMS.
// Media assets are uploaded via the admin panel or API and stored in the configured
// storage backend (local filesystem or S3-compatible object storage).
type MediaID string

// String returns the raw ULID string for this media ID.
func (id MediaID) String() string { return string(id) }

// IsZero reports whether the media ID is empty (unset).
func (id MediaID) IsZero() bool { return id == "" }

// MediaDimensionID identifies a media dimension preset.
// Dimension presets define named image sizes (e.g., "thumbnail", "hero")
// with width and height constraints for automatic image resizing on upload.
type MediaDimensionID string

// String returns the raw ULID string for this media dimension ID.
func (id MediaDimensionID) String() string { return string(id) }

// IsZero reports whether the media dimension ID is empty (unset).
func (id MediaDimensionID) IsZero() bool { return id == "" }

// MediaFolderID identifies a media folder for organizing assets.
// Folders create a hierarchical structure for media items, with each folder
// optionally nested under a parent folder.
type MediaFolderID string

// String returns the raw ULID string for this media folder ID.
func (id MediaFolderID) String() string { return string(id) }

// IsZero reports whether the media folder ID is empty (unset).
func (id MediaFolderID) IsZero() bool { return id == "" }

// --- Admin Media IDs ---
//
// Admin media IDs mirror the public media IDs but reference assets stored in
// the admin media library. The admin media system provides a separate media
// namespace for CMS-internal assets (e.g., admin UI images, system icons)
// managed independently from user-uploaded public media.

// AdminMediaID identifies a media asset in the admin media library.
// Admin media assets are managed separately from public media and are
// used for CMS-internal purposes.
type AdminMediaID string

// String returns the raw ULID string for this admin media ID.
func (id AdminMediaID) String() string { return string(id) }

// IsZero reports whether the admin media ID is empty (unset).
func (id AdminMediaID) IsZero() bool { return id == "" }

// AdminMediaFolderID identifies a folder in the admin media library.
// Admin media folders organize admin-specific assets in a hierarchy
// separate from the public media folder tree.
type AdminMediaFolderID string

// String returns the raw ULID string for this admin media folder ID.
func (id AdminMediaFolderID) String() string { return string(id) }

// IsZero reports whether the admin media folder ID is empty (unset).
func (id AdminMediaFolderID) IsZero() bool { return id == "" }

// --- Auth and RBAC IDs ---
//
// These IDs reference authentication and authorization entities. ModulaCMS uses
// role-based access control (RBAC) with granular "resource:operation" permissions.
// Users are assigned roles, and roles are linked to permissions via a junction table.

// UserID identifies a CMS user account.
// Users authenticate via password or OAuth and are assigned a role that
// determines their permissions.
type UserID string

// String returns the raw ULID string for this user ID.
func (id UserID) String() string { return string(id) }

// IsZero reports whether the user ID is empty (unset).
func (id UserID) IsZero() bool { return id == "" }

// RoleID identifies a role in the RBAC system.
// Built-in roles include "admin" (full access), "editor" (content management),
// and "viewer" (read-only). Custom roles can be created via the API.
type RoleID string

// String returns the raw ULID string for this role ID.
func (id RoleID) String() string { return string(id) }

// IsZero reports whether the role ID is empty (unset).
func (id RoleID) IsZero() bool { return id == "" }

// SessionID identifies an authenticated user session.
// Sessions are created on login and invalidated on logout or expiration.
type SessionID string

// String returns the raw ULID string for this session ID.
func (id SessionID) String() string { return string(id) }

// IsZero reports whether the session ID is empty (unset).
func (id SessionID) IsZero() bool { return id == "" }

// TokenID identifies an API token for programmatic access.
// Tokens provide stateless authentication for API clients and CI/CD pipelines.
type TokenID string

// String returns the raw ULID string for this token ID.
func (id TokenID) String() string { return string(id) }

// IsZero reports whether the token ID is empty (unset).
func (id TokenID) IsZero() bool { return id == "" }

// UserOauthID identifies a user's OAuth provider connection.
// A single user may have multiple OAuth connections (e.g., Google and GitHub)
// linked to the same account.
type UserOauthID string

// String returns the raw ULID string for this user OAuth ID.
func (id UserOauthID) String() string { return string(id) }

// IsZero reports whether the user OAuth ID is empty (unset).
func (id UserOauthID) IsZero() bool { return id == "" }

// UserSshKeyID identifies an SSH public key registered to a user.
// SSH keys are used for authentication when connecting to the CMS via the
// SSH-based TUI interface.
type UserSshKeyID string

// String returns the raw ULID string for this user SSH key ID.
func (id UserSshKeyID) String() string { return string(id) }

// IsZero reports whether the user SSH key ID is empty (unset).
func (id UserSshKeyID) IsZero() bool { return id == "" }

// PermissionID identifies a granular permission in the RBAC system.
// Permissions follow the "resource:operation" format (e.g., "content:read",
// "media:create", "users:delete").
type PermissionID string

// String returns the raw ULID string for this permission ID.
func (id PermissionID) String() string { return string(id) }

// IsZero reports whether the permission ID is empty (unset).
func (id PermissionID) IsZero() bool { return id == "" }

// RolePermissionID identifies a row in the role-permission junction table.
// Each row links a [RoleID] to a [PermissionID], granting that permission
// to all users assigned the role.
type RolePermissionID string

// String returns the raw ULID string for this role-permission junction ID.
func (id RolePermissionID) String() string { return string(id) }

// IsZero reports whether the role-permission junction ID is empty (unset).
func (id RolePermissionID) IsZero() bool { return id == "" }

// --- Routing IDs ---
//
// Routes map URL paths to content entries. The CMS supports multiple route types
// including page routes, redirect routes, and proxy routes. Admin routes are
// managed separately from public routes.

// RouteID identifies a public route in the CMS routing table.
// Routes map URL slugs to content entries, defining how content is accessed
// by frontend clients.
type RouteID string

// String returns the raw ULID string for this route ID.
func (id RouteID) String() string { return string(id) }

// IsZero reports whether the route ID is empty (unset).
func (id RouteID) IsZero() bool { return id == "" }

// AdminRouteID identifies a route in the admin routing table.
// Admin routes are separate from public routes and serve the admin panel
// navigation structure.
type AdminRouteID string

// String returns the raw ULID string for this admin route ID.
func (id AdminRouteID) String() string { return string(id) }

// IsZero reports whether the admin route ID is empty (unset).
func (id AdminRouteID) IsZero() bool { return id == "" }

// --- Infrastructure IDs ---

// TableID identifies a database table registration in the CMS metadata.
// The CMS tracks which tables exist for schema management and migration purposes.
type TableID string

// String returns the raw ULID string for this table ID.
func (id TableID) String() string { return string(id) }

// IsZero reports whether the table ID is empty (unset).
func (id TableID) IsZero() bool { return id == "" }

// EventID identifies a change event in the audit log.
// Every audited mutation (create, update, delete) generates a change event
// that records the operation type, old and new values, and request metadata
// for audit trail and replication.
type EventID string

// String returns the raw ULID string for this event ID.
func (id EventID) String() string { return string(id) }

// IsZero reports whether the event ID is empty (unset).
func (id EventID) IsZero() bool { return id == "" }

// BackupID identifies a CMS backup snapshot.
// Backups include a SQL dump and optionally media assets, stored locally
// or in S3-compatible object storage.
type BackupID string

// String returns the raw ULID string for this backup ID.
func (id BackupID) String() string { return string(id) }

// IsZero reports whether the backup ID is empty (unset).
func (id BackupID) IsZero() bool { return id == "" }

// BackupSetID identifies a backup set (collection of related backups).
type BackupSetID string

// String returns the raw ULID string for this backup set ID.
func (id BackupSetID) String() string { return string(id) }

// IsZero reports whether the backup set ID is empty (unset).
func (id BackupSetID) IsZero() bool { return id == "" }

// NodeID identifies a node in the content tree.
type NodeID string

// String returns the raw ULID string for this node ID.
func (id NodeID) String() string { return string(id) }

// IsZero reports whether the node ID is empty (unset).
func (id NodeID) IsZero() bool { return id == "" }

// PipelineID identifies a deployment pipeline.
type PipelineID string

// String returns the raw ULID string for this pipeline ID.
func (id PipelineID) String() string { return string(id) }

// IsZero reports whether the pipeline ID is empty (unset).
func (id PipelineID) IsZero() bool { return id == "" }

// PluginID identifies a registered plugin.
type PluginID string

// String returns the raw ULID string for this plugin ID.
func (id PluginID) String() string { return string(id) }

// IsZero reports whether the plugin ID is empty (unset).
func (id PluginID) IsZero() bool { return id == "" }

// VerificationID identifies a backup verification record.
type VerificationID string

// String returns the raw ULID string for this verification ID.
func (id VerificationID) String() string { return string(id) }

// IsZero reports whether the verification ID is empty (unset).
func (id VerificationID) IsZero() bool { return id == "" }

// --- Locale IDs ---

// LocaleID identifies a locale for internationalized content.
// Locales represent language/region combinations (e.g., "en-US", "fr-FR")
// used to serve content in multiple languages.
type LocaleID string

// String returns the raw ULID string for this locale ID.
func (id LocaleID) String() string { return string(id) }

// IsZero reports whether the locale ID is empty (unset).
func (id LocaleID) IsZero() bool { return id == "" }

// --- Version IDs ---
//
// Version IDs reference immutable snapshots of content at a point in time.
// The CMS supports content versioning, allowing users to publish, revert,
// and compare different versions of a content entry.

// ContentVersionID identifies a version snapshot of a public content entry.
// Each time content is published or a draft is saved, a new version is created
// with its own ID, preserving the full edit history.
type ContentVersionID string

// String returns the raw ULID string for this content version ID.
func (id ContentVersionID) String() string { return string(id) }

// IsZero reports whether the content version ID is empty (unset).
func (id ContentVersionID) IsZero() bool { return id == "" }

// AdminContentVersionID identifies a version snapshot of an admin content entry.
// Admin content versions work identically to public content versions but
// track the history of CMS configuration content.
type AdminContentVersionID string

// String returns the raw ULID string for this admin content version ID.
func (id AdminContentVersionID) String() string { return string(id) }

// IsZero reports whether the admin content version ID is empty (unset).
func (id AdminContentVersionID) IsZero() bool { return id == "" }

// --- Webhook IDs ---

// WebhookID identifies a webhook subscription.
// Webhooks allow external services to receive HTTP POST notifications when
// CMS events occur (e.g., content published, media uploaded, user created).
type WebhookID string

// String returns the raw ULID string for this webhook ID.
func (id WebhookID) String() string { return string(id) }

// IsZero reports whether the webhook ID is empty (unset).
func (id WebhookID) IsZero() bool { return id == "" }

// WebhookDeliveryID identifies an individual webhook delivery attempt.
// Each time a webhook fires, a delivery record is created to track the
// HTTP request/response, status code, and any retry attempts.
type WebhookDeliveryID string

// String returns the raw ULID string for this webhook delivery ID.
func (id WebhookDeliveryID) String() string { return string(id) }

// IsZero reports whether the webhook delivery ID is empty (unset).
func (id WebhookDeliveryID) IsZero() bool { return id == "" }

// --- Validation IDs ---

// ValidationID identifies a reusable validation configuration.
// Validations define rules that fields can reference to validate user input
// (e.g., required, min/max length, regex patterns, custom validators).
type ValidationID string

// String returns the raw ULID string for this validation ID.
func (id ValidationID) String() string { return string(id) }

// IsZero reports whether the validation ID is empty (unset).
func (id ValidationID) IsZero() bool { return id == "" }

// AdminValidationID identifies an admin-side validation configuration.
// Admin validations serve the same purpose as public validations but operate
// within the admin content namespace.
type AdminValidationID string

// String returns the raw ULID string for this admin validation ID.
func (id AdminValidationID) String() string { return string(id) }

// IsZero reports whether the admin validation ID is empty (unset).
func (id AdminValidationID) IsZero() bool { return id == "" }

// --- Value Types ---
//
// Value types are branded string wrappers for non-ID values that benefit from
// type safety. Unlike ID types, these do not represent ULIDs but rather
// domain-specific string formats.

// Slug represents a URL-safe slug used in content routing.
// Slugs are lowercase, hyphen-separated strings derived from content titles
// (e.g., "my-blog-post"). They appear in URL paths when resolving content
// via the content delivery API.
type Slug string

// String returns the raw slug string.
func (s Slug) String() string { return string(s) }

// IsZero reports whether the slug is empty (unset).
func (s Slug) IsZero() bool { return s == "" }

// Email represents an email address associated with a user account.
type Email string

// String returns the raw email string.
func (e Email) String() string { return string(e) }

// IsZero reports whether the email is empty (unset).
func (e Email) IsZero() bool { return e == "" }

// URL represents a fully-qualified URL string.
type URL string

// String returns the raw URL string.
func (u URL) String() string { return string(u) }

// IsZero reports whether the URL is empty (unset).
func (u URL) IsZero() bool { return u == "" }

// Timestamp represents an RFC 3339 UTC timestamp string as returned by the API.
// All timestamps from the ModulaCMS API are in UTC with second precision
// (e.g., "2026-03-07T15:04:05Z").
//
// Use [Timestamp.Time] to parse into a [time.Time] for calculations.
// Use [NewTimestamp] to create a Timestamp from an existing [time.Time].
// Use [TimestampNow] to get the current time as a Timestamp.
type Timestamp string

// String returns the raw RFC 3339 timestamp string.
func (t Timestamp) String() string { return string(t) }

// IsZero reports whether the timestamp is empty (unset).
func (t Timestamp) IsZero() bool { return t == "" }

// Time parses the RFC 3339 timestamp string into a [time.Time] value.
// Returns an error if the timestamp string is malformed.
func (t Timestamp) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(t))
}

// NewTimestamp creates a [Timestamp] from a [time.Time] value.
// The time is converted to UTC and formatted as RFC 3339.
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp(t.UTC().Format(time.RFC3339))
}

// TimestampNow returns the current time as a [Timestamp].
// This is a convenience wrapper around NewTimestamp(time.Now()).
func TimestampNow() Timestamp {
	return NewTimestamp(time.Now())
}
