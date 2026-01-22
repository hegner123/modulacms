# New sqlc.yml Configuration

This configuration replaces the current `sql/sqlc.yml` to use custom types across all three database engines.

```yaml
version: "2"

# ============================================
# GLOBAL OVERRIDES - Custom types for all engines
# ============================================
overrides:
  go:
    rename:
      id: "ID"
      url: "URL"
      oauth: "OAuth"
      ssh: "SSH"

    overrides:
      # ========================================
      # PRIMARY ID COLUMNS - Typed IDs
      # ========================================
      - column: "datatypes.datatype_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "DatatypeID"
      - column: "users.user_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "UserID"
      - column: "roles.role_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "RoleID"
      - column: "permissions.permission_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "PermissionID"
      - column: "fields.field_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "FieldID"
      - column: "content_data.content_data_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "ContentID"
      - column: "content_fields.content_field_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "ContentFieldID"
      - column: "media.media_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "MediaID"
      - column: "media_dimension.media_dimension_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "MediaDimensionID"
      - column: "sessions.session_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "SessionID"
      - column: "tokens.token_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "TokenID"
      - column: "routes.route_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "RouteID"
      - column: "admin_routes.admin_route_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "AdminRouteID"
      - column: "tables.table_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "TableID"
      - column: "user_oauth.user_oauth_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "UserOauthID"
      - column: "user_ssh_keys.user_ssh_key_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "UserSshKeyID"
      - column: "admin_datatypes.admin_datatype_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "AdminDatatypeID"
      - column: "admin_fields.admin_field_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "AdminFieldID"
      - column: "admin_content_data.admin_content_data_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "AdminContentID"
      - column: "admin_content_fields.admin_content_field_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "AdminContentFieldID"
      - column: "datatypes_fields.datatype_field_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "DatatypeFieldID"
      - column: "admin_datatypes_fields.admin_datatype_field_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "AdminDatatypeFieldID"

      # ========================================
      # FOREIGN KEY COLUMNS - Nullable Typed IDs
      # ========================================
      # datatype_id foreign keys
      - column: "*.datatype_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableDatatypeID"
      - column: "*.datatype_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "DatatypeID"

      # user_id foreign keys
      - column: "*.user_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableUserID"
      - column: "*.user_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "UserID"
      - column: "*.author_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableUserID"
      - column: "*.author_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "UserID"

      # role_id foreign keys
      - column: "*.role_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableRoleID"
      - column: "*.role_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "RoleID"

      # content_data_id / parent_id foreign keys
      - column: "*.content_data_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableContentID"
      - column: "*.content_data_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "ContentID"
      - column: "*.parent_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableContentID"

      # field_id foreign keys
      - column: "*.field_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableFieldID"
      - column: "*.field_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "FieldID"

      # media_id foreign keys
      - column: "*.media_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableMediaID"
      - column: "*.media_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "MediaID"

      # route_id foreign keys
      - column: "*.route_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableRouteID"
      - column: "*.route_id"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "RouteID"

      # ========================================
      # TIMESTAMP COLUMNS - Unified Timestamp
      # ========================================
      - column: "*.date_created"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "*.date_modified"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "*.created_at"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "*.updated_at"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "*.expires_at"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "sessions.last_activity"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"

      # ========================================
      # DISTRIBUTED SYSTEM COLUMNS
      # ========================================
      # node_id - present on all tables for multi-node tracking
      - column: "*.node_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NodeID"

      # hlc_timestamp - Hybrid Logical Clock for event ordering
      - column: "*.hlc_timestamp"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "HLC"

      # conflict_policy - per-datatype conflict resolution strategy
      - column: "datatypes.conflict_policy"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "ConflictPolicy"
      - column: "admin_datatypes.conflict_policy"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "ConflictPolicy"

      # ========================================
      # CHANGE EVENTS TABLE (Audit + Replication + Webhooks)
      # ========================================
      - column: "change_events.event_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "EventID"
      - column: "change_events.node_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NodeID"
      - column: "change_events.hlc_timestamp"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "HLC"
      - column: "change_events.wall_timestamp"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "change_events.operation"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Operation"
      - column: "change_events.action"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Action"
      - column: "change_events.user_id"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableUserID"
      - column: "change_events.synced_at"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableTimestamp"
      - column: "change_events.consumed_at"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableTimestamp"

      # ========================================
      # BACKUP TABLES (Distributed Restore)
      # ========================================
      - column: "backups.backup_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "BackupID"
      - column: "backups.node_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NodeID"
      - column: "backups.backup_type"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "BackupType"
      - column: "backups.status"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "BackupStatus"
      - column: "backups.started_at"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "backups.completed_at"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableTimestamp"
      - column: "backups.hlc_timestamp"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "HLC"

      - column: "backup_verifications.verification_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "VerificationID"
      - column: "backup_verifications.backup_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "BackupID"
      - column: "backup_verifications.verified_at"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "backup_verifications.status"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "VerificationStatus"

      - column: "backup_sets.backup_set_id"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "BackupSetID"
      - column: "backup_sets.created_at"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Timestamp"
      - column: "backup_sets.hlc_timestamp"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "HLC"
      - column: "backup_sets.status"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "BackupSetStatus"

      # ========================================
      # ENUM COLUMNS - Domain Types
      # ========================================
      - column: "*.status"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "ContentStatus"
      - column: "fields.type"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "FieldType"
      - column: "admin_fields.type"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "FieldType"
      - column: "routes.type"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "RouteType"

      # ========================================
      # VALIDATION COLUMNS - Specific Types
      # ========================================
      - column: "*.slug"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Slug"
      - column: "*.email"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Email"
      - column: "users.email"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "Email"
      - column: "*.url"
        nullable: true
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "NullableURL"
      - column: "*.url"
        nullable: false
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "URL"
      - column: "media.url"
        go_type:
          import: "github.com/hegner123/modulacms/internal/db"
          type: "URL"

# ============================================
# PER-ENGINE SQL CONFIGURATIONS
# ============================================
sql:
  # ---------- SQLite ----------
  - engine: "sqlite"
    schema: "schema/**/schema.sql"
    queries: "schema/**/queries.sql"
    gen:
      go:
        package: "mdb"
        out: "../internal/db-sqlite"
        emit_json_tags: true
        emit_empty_slices: true
        json_tags_case_style: "snake"
        initialisms: ["ID", "URL", "OAuth", "SSH", "HTTP", "HTTPS", "API"]
        query_parameter_limit: 0

  # ---------- MySQL ----------
  - engine: "mysql"
    schema: "schema/**/schema_mysql.sql"
    queries: "schema/**/queries_mysql.sql"
    gen:
      go:
        package: "mdbm"
        out: "../internal/db-mysql"
        emit_json_tags: true
        emit_empty_slices: true
        json_tags_case_style: "snake"
        initialisms: ["ID", "URL", "OAuth", "SSH", "HTTP", "HTTPS", "API"]
        query_parameter_limit: 0

  # ---------- PostgreSQL ----------
  - engine: "postgresql"
    schema: "schema/**/schema_psql.sql"
    queries: "schema/**/queries_psql.sql"
    gen:
      go:
        package: "mdbp"
        out: "../internal/db-psql"
        emit_json_tags: true
        emit_empty_slices: true
        json_tags_case_style: "snake"
        initialisms: ["ID", "URL", "OAuth", "SSH", "HTTP", "HTTPS", "API"]
        query_parameter_limit: 0
```

**Note:** `emit_pointers_for_null_types` removed - we now use custom `NullableXID` types instead for better type safety.
