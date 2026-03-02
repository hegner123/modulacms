package main

// Engine describes a database engine for sqlc config generation.
type Engine struct {
	Label      string // "SQLite", "MySQL", "PostgreSQL"
	Name       string // "sqlite", "mysql", "postgresql"
	Package    string // "mdb", "mdbm", "mdbp"
	OutDir     string // "../internal/db-sqlite"
	SchemaGlob string // "schema/**/schema.sql"
	QueryGlob  string // "schema/**/queries.sql"
}

// Rename maps a sqlc name to a Go name.
type Rename struct {
	Comment string // optional YAML comment emitted above this entry
	From    string
	To      string
	Quoted  bool // true: emit "Value", false: emit Value (bare)
}

// Override maps a column to a Go type.
type Override struct {
	Comment  string // optional YAML comment emitted above this entry
	Column   string
	Nullable *bool  // nil=omit, true/false=emit
	Import   string // empty for bare types like int64
	Type     string
}

// TemplateData is the top-level data passed to the template.
type TemplateData struct {
	Engines   []Engine
	Renames   []Rename
	Overrides []Override
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool { return &b }

// Engines defines the three database engines.
var Engines = []Engine{
	{
		Label:      "SQLite",
		Name:       "sqlite",
		Package:    "mdb",
		OutDir:     "../internal/db-sqlite",
		SchemaGlob: "schema/**/schema.sql",
		QueryGlob:  "schema/**/queries.sql",
	},
	{
		Label:      "MySQL",
		Name:       "mysql",
		Package:    "mdbm",
		OutDir:     "../internal/db-mysql",
		SchemaGlob: "schema/**/schema_mysql.sql",
		QueryGlob:  "schema/**/queries_mysql.sql",
	},
	{
		Label:      "PostgreSQL",
		Name:       "postgresql",
		Package:    "mdbp",
		OutDir:     "../internal/db-psql",
		SchemaGlob: "schema/**/schema_psql.sql",
		QueryGlob:  "schema/**/queries_psql.sql",
	},
}

// typesImport is the common import path for all typed IDs.
const typesImport = "github.com/hegner123/modulacms/internal/db/types"

// Renames defines all column/table name mappings shared across engines.
var Renames = []Rename{
	// Base name overrides
	{From: "id", To: "ID", Quoted: true},
	{From: "url", To: "URL", Quoted: true},
	{From: "oauth", To: "OAuth", Quoted: true},
	{From: "ssh", To: "SSH", Quoted: true},
	// Column name ID suffixes (explicit to get uppercase ID)
	{Comment: "Column name ID suffixes (explicit to get uppercase ID)", From: "admin_content_data_id", To: "AdminContentDataID", Quoted: true},
	{From: "admin_content_field_id", To: "AdminContentFieldID", Quoted: true},
	{From: "admin_content_relation_id", To: "AdminContentRelationID", Quoted: true},
	{From: "admin_datatype_id", To: "AdminDatatypeID", Quoted: true},
	{From: "admin_field_id", To: "AdminFieldID", Quoted: true},
	{From: "admin_route_id", To: "AdminRouteID", Quoted: true},
	{From: "author_id", To: "AuthorID", Quoted: true},
	{From: "backup_id", To: "BackupID", Quoted: true},
	{From: "backup_set_id", To: "BackupSetID", Quoted: true},
	{From: "content_data_id", To: "ContentDataID", Quoted: true},
	{From: "content_field_id", To: "ContentFieldID", Quoted: true},
	{From: "content_relation_id", To: "ContentRelationID", Quoted: true},
	{From: "content_version_id", To: "ContentVersionID", Quoted: true},
	{From: "admin_content_version_id", To: "AdminContentVersionID", Quoted: true},
	{From: "datatype_id", To: "DatatypeID", Quoted: true},
	{From: "event_id", To: "EventID", Quoted: true},
	{From: "field_id", To: "FieldID", Quoted: true},
	{From: "field_type_id", To: "FieldTypeID", Quoted: true},
	{From: "admin_field_type_id", To: "AdminFieldTypeID", Quoted: true},
	{From: "first_child_id", To: "FirstChildID", Quoted: true},
	{From: "md_id", To: "MdID", Quoted: true},
	{From: "media_dimension_id", To: "MediaDimensionID", Quoted: true},
	{From: "media_id", To: "MediaID", Quoted: true},
	{From: "next_sibling_id", To: "NextSiblingID", Quoted: true},
	{From: "node_id", To: "NodeID", Quoted: true},
	{From: "oauth_provider_user_id", To: "OAuthProviderUserID", Quoted: true},
	{From: "parent_id", To: "ParentID", Quoted: true},
	{From: "permission_id", To: "PermissionID", Quoted: true},
	{From: "prev_sibling_id", To: "PrevSiblingID", Quoted: true},
	{From: "record_id", To: "RecordID", Quoted: true},
	{From: "role_id", To: "RoleID", Quoted: true},
	{From: "route_id", To: "RouteID", Quoted: true},
	{From: "session_id", To: "SessionID", Quoted: true},
	{From: "source_content_id", To: "SourceContentID", Quoted: true},
	{From: "ssh_key_id", To: "SSHKeyID", Quoted: true},
	{From: "table_id", To: "TableID", Quoted: true},
	{From: "target_content_id", To: "TargetContentID", Quoted: true},
	{From: "user_id", To: "UserID", Quoted: true},
	{From: "user_oauth_id", To: "UserOAuthID", Quoted: true},
	{From: "user_ssh_key_id", To: "UserSSHKeyID", Quoted: true},
	{From: "verification_id", To: "VerificationID", Quoted: true},
	{From: "plugin_id", To: "PluginID", Quoted: true},
	{From: "pipeline_id", To: "PipelineID", Quoted: true},
	{From: "locale_id", To: "LocaleID", Quoted: true},
	{From: "webhook_id", To: "WebhookID", Quoted: true},
	{From: "delivery_id", To: "DeliveryID", Quoted: true},
	// Table-to-struct names (sqlc singularizes first, then applies rename)
	{Comment: "Table-to-struct names (sqlc singularizes first, then applies rename)", From: "admin_content_datum", To: "AdminContentData"},
	{From: "admin_content_field", To: "AdminContentFields"},
	{From: "admin_content_relation", To: "AdminContentRelations"},
	{From: "admin_content_version", To: "AdminContentVersions"},
	{From: "admin_datatype", To: "AdminDatatypes"},
	{From: "admin_field", To: "AdminFields"},
	{From: "admin_route", To: "AdminRoutes"},
	{From: "content_datum", To: "ContentData"},
	{From: "content_field", To: "ContentFields"},
	{From: "content_relation", To: "ContentRelations"},
	{From: "content_version", To: "ContentVersions"},
	{From: "datatype", To: "Datatypes"},
	{From: "field", To: "Fields"},
	{From: "field_type", To: "FieldTypes"},
	{From: "admin_field_type", To: "AdminFieldTypes"},
	{From: "medium", To: "Media"},
	{From: "media_dimension", To: "MediaDimensions"},
	{From: "permission", To: "Permissions"},
	{From: "pipeline", To: "Pipelines"},
	{From: "plugin", To: "Plugins"},
	{From: "role", To: "Roles"},
	{From: "role_permission", To: "RolePermissions"},
	{From: "route", To: "Routes"},
	{From: "session", To: "Sessions"},
	{From: "table", To: "Tables"},
	{From: "token", To: "Tokens"},
	{From: "user", To: "Users"},
	{From: "user_ssh_key", To: "UserSshKeys"},
	{From: "webhook", To: "Webhooks"},
	{From: "webhook_delivery", To: "WebhookDeliveries"},
}

// Overrides defines all column-to-Go-type overrides shared across engines.
var Overrides = []Override{
	// PRIMARY ID COLUMNS
	{Comment: "PRIMARY ID COLUMNS", Column: "datatypes.datatype_id", Import: typesImport, Type: "DatatypeID"},
	{Column: "users.user_id", Import: typesImport, Type: "UserID"},
	{Column: "roles.role_id", Import: typesImport, Type: "RoleID"},
	{Column: "permissions.permission_id", Import: typesImport, Type: "PermissionID"},
	{Column: "fields.field_id", Import: typesImport, Type: "FieldID"},
	{Column: "content_data.content_data_id", Import: typesImport, Type: "ContentID"},
	{Column: "content_fields.content_field_id", Import: typesImport, Type: "ContentFieldID"},
	{Column: "media.media_id", Import: typesImport, Type: "MediaID"},
	{Column: "media_dimension.media_dimension_id", Import: typesImport, Type: "MediaDimensionID"},
	{Column: "sessions.session_id", Import: typesImport, Type: "SessionID"},
	{Column: "tokens.token_id", Import: typesImport, Type: "TokenID"},
	{Column: "routes.route_id", Import: typesImport, Type: "RouteID"},
	{Column: "admin_routes.admin_route_id", Import: typesImport, Type: "AdminRouteID"},
	{Column: "tables.table_id", Import: typesImport, Type: "TableID"},
	{Column: "user_oauth.user_oauth_id", Import: typesImport, Type: "UserOauthID"},
	{Column: "user_ssh_keys.user_ssh_key_id", Import: typesImport, Type: "UserSshKeyID"},
	{Column: "admin_datatypes.admin_datatype_id", Import: typesImport, Type: "AdminDatatypeID"},
	{Column: "admin_fields.admin_field_id", Import: typesImport, Type: "AdminFieldID"},
	{Column: "admin_content_data.admin_content_data_id", Import: typesImport, Type: "AdminContentID"},
	{Column: "admin_content_fields.admin_content_field_id", Import: typesImport, Type: "AdminContentFieldID"},
	// LOCALES
	{Comment: "LOCALES", Column: "locales.locale_id", Import: typesImport, Type: "LocaleID"},
	// FIELD TYPES
	{Comment: "FIELD TYPES", Column: "field_types.field_type_id", Import: typesImport, Type: "FieldTypeID"},
	{Column: "admin_field_types.admin_field_type_id", Import: typesImport, Type: "AdminFieldTypeID"},
	// ROLE PERMISSIONS
	{Comment: "ROLE PERMISSIONS", Column: "role_permissions.id", Import: typesImport, Type: "RolePermissionID"},
	// CONTENT RELATIONS - PRIMARY IDs
	{Comment: "CONTENT RELATIONS \u2014 PRIMARY IDs", Column: "content_relations.content_relation_id", Import: typesImport, Type: "ContentRelationID"},
	{Column: "admin_content_relations.admin_content_relation_id", Import: typesImport, Type: "AdminContentRelationID"},
	// CONTENT RELATIONS - FK columns
	{Comment: "CONTENT RELATIONS \u2014 FK columns", Column: "content_relations.source_content_id", Import: typesImport, Type: "ContentID"},
	{Column: "content_relations.target_content_id", Import: typesImport, Type: "ContentID"},
	{Column: "admin_content_relations.source_content_id", Import: typesImport, Type: "AdminContentID"},
	{Column: "admin_content_relations.target_content_id", Import: typesImport, Type: "AdminContentID"},
	// TABLE-SPECIFIC FK OVERRIDES (override wildcards for NOT NULL columns)
	{Comment: "TABLE-SPECIFIC FK OVERRIDES (override wildcards for NOT NULL columns)", Column: "content_relations.field_id", Import: typesImport, Type: "FieldID"},
	{Column: "admin_content_relations.admin_field_id", Import: typesImport, Type: "AdminFieldID"},
	{Column: "role_permissions.role_id", Import: typesImport, Type: "RoleID"},
	{Column: "role_permissions.permission_id", Import: typesImport, Type: "PermissionID"},
	{Column: "content_versions.content_data_id", Import: typesImport, Type: "ContentID"},
	{Column: "admin_content_versions.admin_content_data_id", Import: typesImport, Type: "AdminContentID"},
	// NOT NULL author_id columns
	{Comment: "NOT NULL author_id columns", Column: "content_data.author_id", Import: typesImport, Type: "UserID"},
	{Column: "datatypes.author_id", Import: typesImport, Type: "UserID"},
	{Column: "admin_content_data.author_id", Import: typesImport, Type: "UserID"},
	{Column: "admin_datatypes.author_id", Import: typesImport, Type: "UserID"},
	{Column: "content_fields.author_id", Import: typesImport, Type: "UserID"},
	{Column: "admin_content_fields.author_id", Import: typesImport, Type: "UserID"},
	// FOREIGN KEY COLUMNS
	{Comment: "FOREIGN KEY COLUMNS", Column: "*.datatype_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableDatatypeID"},
	{Column: "*.datatype_id", Nullable: boolPtr(false), Import: typesImport, Type: "DatatypeID"},
	{Column: "*.user_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableUserID"},
	{Column: "*.user_id", Nullable: boolPtr(false), Import: typesImport, Type: "UserID"},
	{Column: "*.author_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableUserID"},
	{Column: "*.author_id", Nullable: boolPtr(false), Import: typesImport, Type: "UserID"},
	{Column: "*.role_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableRoleID"},
	{Column: "*.role_id", Nullable: boolPtr(false), Import: typesImport, Type: "RoleID"},
	{Column: "*.permission_id", Nullable: boolPtr(false), Import: typesImport, Type: "PermissionID"},
	{Column: "*.content_data_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableContentID"},
	{Column: "*.content_data_id", Nullable: boolPtr(false), Import: typesImport, Type: "ContentID"},
	// content_data tree pointers (self-referential FKs to content_data)
	{Comment: "content_data tree pointers (self-referential FKs to content_data)", Column: "content_data.first_child_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableContentID"},
	{Column: "content_data.next_sibling_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableContentID"},
	{Column: "content_data.prev_sibling_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableContentID"},
	// admin_content_data tree pointers (self-referential FKs to admin_content_data)
	{Comment: "admin_content_data tree pointers (self-referential FKs to admin_content_data)", Column: "admin_content_data.first_child_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminContentID"},
	{Column: "admin_content_data.next_sibling_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminContentID"},
	{Column: "admin_content_data.prev_sibling_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminContentID"},
	// FIX: datatypes.parent_id references datatypes, not content_data
	{Comment: "FIX: datatypes.parent_id references datatypes, not content_data", Column: "datatypes.parent_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableDatatypeID"},
	// FIX: admin_datatypes.parent_id references admin_datatypes, not content_data
	{Comment: "FIX: admin_datatypes.parent_id references admin_datatypes, not content_data", Column: "admin_datatypes.parent_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminDatatypeID"},
	// FIX: fields.parent_id references datatypes, not content_data
	{Comment: "FIX: fields.parent_id references datatypes, not content_data", Column: "fields.parent_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableDatatypeID"},
	// FIX: admin_fields.parent_id references admin_datatypes, not content_data
	{Comment: "FIX: admin_fields.parent_id references admin_datatypes, not content_data", Column: "admin_fields.parent_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminDatatypeID"},
	// FIX: admin_content_data.parent_id references admin_content_data, not content_data
	{Comment: "FIX: admin_content_data.parent_id references admin_content_data, not content_data", Column: "admin_content_data.parent_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminContentID"},
	{Column: "*.parent_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableContentID"},
	{Column: "*.field_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableFieldID"},
	{Column: "*.field_id", Nullable: boolPtr(false), Import: typesImport, Type: "FieldID"},
	{Column: "*.media_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableMediaID"},
	{Column: "*.media_id", Nullable: boolPtr(false), Import: typesImport, Type: "MediaID"},
	{Column: "*.route_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableRouteID"},
	{Column: "*.route_id", Nullable: boolPtr(false), Import: typesImport, Type: "RouteID"},
	{Column: "*.admin_datatype_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminDatatypeID"},
	{Column: "*.admin_datatype_id", Nullable: boolPtr(false), Import: typesImport, Type: "AdminDatatypeID"},
	{Column: "*.admin_field_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminFieldID"},
	{Column: "*.admin_field_id", Nullable: boolPtr(false), Import: typesImport, Type: "AdminFieldID"},
	{Column: "*.admin_route_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminRouteID"},
	{Column: "*.admin_route_id", Nullable: boolPtr(false), Import: typesImport, Type: "AdminRouteID"},
	{Column: "*.admin_content_data_id", Nullable: boolPtr(true), Import: typesImport, Type: "NullableAdminContentID"},
	{Column: "*.admin_content_data_id", Nullable: boolPtr(false), Import: typesImport, Type: "AdminContentID"},
	// TIMESTAMPS
	{Comment: "TIMESTAMPS", Column: "*.date_created", Import: typesImport, Type: "Timestamp"},
	{Column: "*.date_modified", Import: typesImport, Type: "Timestamp"},
	{Column: "*.expires_at", Import: typesImport, Type: "Timestamp"},
	// DISTRIBUTED SYSTEM
	{Comment: "DISTRIBUTED SYSTEM", Column: "*.node_id", Import: typesImport, Type: "NodeID"},
	{Column: "*.hlc_timestamp", Import: typesImport, Type: "HLC"},
	{Column: "datatypes.conflict_policy", Import: typesImport, Type: "ConflictPolicy"},
	{Column: "admin_datatypes.conflict_policy", Import: typesImport, Type: "ConflictPolicy"},
	// CHANGE EVENTS
	{Comment: "CHANGE EVENTS", Column: "change_events.event_id", Import: typesImport, Type: "EventID"},
	{Column: "change_events.operation", Import: typesImport, Type: "Operation"},
	{Column: "change_events.action", Import: typesImport, Type: "Action"},
	// BACKUPS
	{Comment: "BACKUPS", Column: "backups.backup_id", Import: typesImport, Type: "BackupID"},
	{Column: "backups.backup_type", Import: typesImport, Type: "BackupType"},
	{Column: "backups.status", Import: typesImport, Type: "BackupStatus"},
	{Column: "backups.started_at", Import: typesImport, Type: "Timestamp"},
	{Column: "backups.completed_at", Import: typesImport, Type: "Timestamp"},
	{Column: "backups.duration_ms", Import: typesImport, Type: "NullableInt64"},
	{Column: "backups.record_count", Import: typesImport, Type: "NullableInt64"},
	{Column: "backups.size_bytes", Import: typesImport, Type: "NullableInt64"},
	{Column: "backups.replication_lsn", Import: typesImport, Type: "NullableString"},
	{Column: "backups.checksum", Import: typesImport, Type: "NullableString"},
	{Column: "backups.triggered_by", Import: typesImport, Type: "NullableString"},
	{Column: "backups.error_message", Import: typesImport, Type: "NullableString"},
	{Column: "backups.metadata", Import: typesImport, Type: "JSONData"},
	{Column: "backup_verifications.verification_id", Import: typesImport, Type: "VerificationID"},
	{Column: "backup_verifications.backup_id", Import: typesImport, Type: "BackupID"},
	{Column: "backup_verifications.status", Import: typesImport, Type: "VerificationStatus"},
	{Column: "backup_verifications.verified_at", Import: typesImport, Type: "Timestamp"},
	{Column: "backup_verifications.verified_by", Import: typesImport, Type: "NullableString"},
	{Column: "backup_verifications.restore_tested", Import: typesImport, Type: "NullableBool"},
	{Column: "backup_verifications.checksum_valid", Import: typesImport, Type: "NullableBool"},
	{Column: "backup_verifications.record_count_match", Import: typesImport, Type: "NullableBool"},
	{Column: "backup_verifications.error_message", Import: typesImport, Type: "NullableString"},
	{Column: "backup_verifications.duration_ms", Import: typesImport, Type: "NullableInt64"},
	{Column: "backup_sets.backup_set_id", Import: typesImport, Type: "BackupSetID"},
	{Column: "backup_sets.status", Import: typesImport, Type: "BackupSetStatus"},
	{Column: "backup_sets.backup_ids", Import: typesImport, Type: "JSONData"},
	{Column: "backup_sets.completed_count", Import: typesImport, Type: "NullableInt64"},
	{Column: "backup_sets.error_message", Import: typesImport, Type: "NullableString"},
	// CHANGE EVENTS
	{Comment: "CHANGE EVENTS", Column: "change_events.wall_timestamp", Import: typesImport, Type: "Timestamp"},
	{Column: "change_events.old_values", Import: typesImport, Type: "JSONData"},
	{Column: "change_events.new_values", Import: typesImport, Type: "JSONData"},
	{Column: "change_events.metadata", Import: typesImport, Type: "JSONData"},
	{Column: "change_events.request_id", Import: typesImport, Type: "NullableString"},
	{Column: "change_events.ip", Import: typesImport, Type: "NullableString"},
	{Column: "change_events.synced_at", Import: typesImport, Type: "Timestamp"},
	{Column: "change_events.consumed_at", Import: typesImport, Type: "Timestamp"},
	// SAFE BOOLEANS
	{Comment: "SAFE BOOLEANS", Column: "roles.system_protected", Import: typesImport, Type: "SafeBool"},
	{Column: "permissions.system_protected", Import: typesImport, Type: "SafeBool"},
	// NULLABLE INTEGERS
	{Comment: "NULLABLE INTEGERS", Column: "media_dimensions.width", Import: typesImport, Type: "NullableInt64"},
	{Column: "media_dimensions.height", Import: typesImport, Type: "NullableInt64"},
	// TIMESTAMP OVERRIDES
	{Comment: "TIMESTAMP OVERRIDES", Column: "tokens.issued_at", Import: typesImport, Type: "Timestamp"},
	{Column: "sessions.last_access", Import: typesImport, Type: "Timestamp"},
	{Column: "user_oauth.token_expires_at", Import: typesImport, Type: "Timestamp"},
	// ENUMS
	{Comment: "ENUMS", Column: "content_data.status", Import: typesImport, Type: "ContentStatus"},
	{Column: "admin_content_data.status", Import: typesImport, Type: "ContentStatus"},
	{Column: "fields.type", Import: typesImport, Type: "FieldType"},
	{Column: "admin_fields.type", Import: typesImport, Type: "FieldType"},
	{Column: "fields.translatable", Import: typesImport, Type: "SafeBool"},
	{Column: "admin_fields.translatable", Import: typesImport, Type: "SafeBool"},
	{Column: "fields.roles", Import: typesImport, Type: "NullableString"},
	{Column: "admin_fields.roles", Import: typesImport, Type: "NullableString"},
	{Column: "locales.is_default", Type: "int64"},
	{Column: "locales.is_enabled", Type: "int64"},
	{Column: "locales.sort_order", Type: "int64"},
	{Column: "routes.type", Import: typesImport, Type: "RouteType"},
	// VALIDATION
	{Comment: "VALIDATION", Column: "*.slug", Import: typesImport, Type: "Slug"},
	{Column: "users.email", Import: typesImport, Type: "Email"},
	{Column: "media.url", Import: typesImport, Type: "URL"},
	{Column: "media.focal_x", Import: typesImport, Type: "NullableFloat64"},
	{Column: "media.focal_y", Import: typesImport, Type: "NullableFloat64"},
	// PLUGINS
	{Comment: "PLUGINS", Column: "plugins.plugin_id", Import: typesImport, Type: "PluginID"},
	{Column: "plugins.status", Import: typesImport, Type: "PluginStatus"},
	{Column: "plugins.capabilities", Import: typesImport, Type: "JSONData"},
	{Column: "plugins.approved_access", Import: typesImport, Type: "JSONData"},
	{Column: "plugins.date_installed", Import: typesImport, Type: "Timestamp"},
	// WEBHOOKS
	{Comment: "WEBHOOKS", Column: "webhooks.webhook_id", Import: typesImport, Type: "WebhookID"},
	{Column: "webhooks.author_id", Import: typesImport, Type: "UserID"},
	{Column: "webhooks.is_active", Type: "int64"},
	{Column: "webhook_deliveries.delivery_id", Import: typesImport, Type: "WebhookDeliveryID"},
	{Column: "webhook_deliveries.webhook_id", Import: typesImport, Type: "WebhookID"},
	// PIPELINES
	{Comment: "PIPELINES", Column: "pipelines.pipeline_id", Import: typesImport, Type: "PipelineID"},
	{Column: "pipelines.plugin_id", Import: typesImport, Type: "PluginID"},
	{Column: "pipelines.config", Import: typesImport, Type: "JSONData"},
	// CONTENT VERSIONS
	{Comment: "CONTENT VERSIONS", Column: "content_versions.content_version_id", Import: typesImport, Type: "ContentVersionID"},
	{Column: "content_versions.published_by", Nullable: boolPtr(true), Import: typesImport, Type: "NullableUserID"},
	{Column: "content_versions.date_created", Import: typesImport, Type: "Timestamp"},
	{Column: "admin_content_versions.admin_content_version_id", Import: typesImport, Type: "AdminContentVersionID"},
	{Column: "admin_content_versions.published_by", Nullable: boolPtr(true), Import: typesImport, Type: "NullableUserID"},
	{Column: "admin_content_versions.date_created", Import: typesImport, Type: "Timestamp"},
	// PUBLISH METADATA on content_data / admin_content_data
	{Comment: "PUBLISH METADATA on content_data / admin_content_data", Column: "content_data.published_at", Nullable: boolPtr(true), Import: typesImport, Type: "Timestamp"},
	{Column: "content_data.published_by", Nullable: boolPtr(true), Import: typesImport, Type: "NullableUserID"},
	{Column: "content_data.publish_at", Nullable: boolPtr(true), Import: typesImport, Type: "Timestamp"},
	{Column: "admin_content_data.published_at", Nullable: boolPtr(true), Import: typesImport, Type: "Timestamp"},
	{Column: "admin_content_data.published_by", Nullable: boolPtr(true), Import: typesImport, Type: "NullableUserID"},
	{Column: "admin_content_data.publish_at", Nullable: boolPtr(true), Import: typesImport, Type: "Timestamp"},
}
