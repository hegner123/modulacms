# ModulaCMS Database Schema

Database schema for ModulaCMS headless CMS supporting SQLite, MySQL, and PostgreSQL.

## Overview

The ModulaCMS database uses a numbered migration system with schema organized into directories under schema/. Each directory contains table definitions and sqlc-annotated queries that generate type-safe Go code for three database engines.

Schema directories are numbered for migration ordering: 0_backups, 1_permissions, 2_roles, etc. The sqlc tool generates three driver packages: mdb for SQLite, mdbm for MySQL, and mdbp for PostgreSQL.

All primary keys use ULID format: 26-character lexicographically sortable unique identifiers. Foreign keys reference these ULIDs. Timestamps use TEXT columns in SQLite storing RFC3339 format, DATETIME in MySQL, and TIMESTAMP in PostgreSQL.

## Core Tables

### users

User accounts with authentication and role assignment.

```sql
CREATE TABLE users (
    user_id TEXT PRIMARY KEY CHECK (length(user_id) = 26),
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL REFERENCES roles ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: user_id is a 26-char ULID primary key, username must be unique, email is indexed separately with unique constraint, hash stores bcrypt password hash, role references roles table, date_created and date_modified use automatic triggers.

Foreign keys: role references roles.role_id with SET NULL on delete.

### roles

Role definitions for access control.

```sql
CREATE TABLE roles (
    role_id TEXT PRIMARY KEY CHECK (length(role_id) = 26),
    role_name TEXT NOT NULL UNIQUE,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: role_id is ULID primary key, role_name is unique string identifier, timestamps track creation and modification.

### permissions

Permission grants linking roles to specific permissions.

```sql
CREATE TABLE permissions (
    permission_id TEXT PRIMARY KEY CHECK (length(permission_id) = 26),
    role_id TEXT NOT NULL REFERENCES roles ON DELETE CASCADE,
    permission_name TEXT NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: permission_id is ULID primary key, role_id references roles table with CASCADE delete, permission_name is string identifier for the granted permission.

Foreign keys: role_id references roles.role_id with CASCADE on delete.

### datatypes

Content type definitions forming a hierarchy with parent relationships.

```sql
CREATE TABLE datatypes (
    datatype_id TEXT PRIMARY KEY CHECK (length(datatype_id) = 26),
    parent_id TEXT DEFAULT NULL REFERENCES datatypes ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: datatype_id is ULID primary key, parent_id allows hierarchical organization, label is display name, type categorizes datatype as GLOBAL or ROOT, author_id tracks who created it.

Foreign keys: parent_id self-references datatypes.datatype_id, author_id references users.user_id.

Indexes: idx_datatypes_parent on parent_id, idx_datatypes_author on author_id.

### fields

Field definitions for datatypes describing content structure.

```sql
CREATE TABLE fields (
    field_id TEXT PRIMARY KEY CHECK (length(field_id) = 26),
    parent_id TEXT REFERENCES datatypes ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: field_id is ULID primary key, parent_id references the datatype this field belongs to, label is display name, type specifies field type like text, textarea, number, date, boolean, select, media, relation, json, richtext, slug, email, url.

Foreign keys: parent_id references datatypes.datatype_id, author_id references users.user_id.

### routes

URL routes for published content.

```sql
CREATE TABLE routes (
    route_id TEXT PRIMARY KEY CHECK (length(route_id) = 26),
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: route_id is ULID primary key, slug is URL-safe unique path, title is display name, status is integer flag for route state, author_id tracks who created the route.

Foreign keys: author_id references users.user_id.

Indexes: idx_routes_author on author_id.

### content_data

Content records organized in a tree structure using sibling pointers for O(1) operations.

```sql
CREATE TABLE content_data (
    content_data_id TEXT PRIMARY KEY CHECK (length(content_data_id) = 26),
    parent_id TEXT REFERENCES content_data ON DELETE SET NULL,
    first_child_id TEXT REFERENCES content_data ON DELETE SET NULL,
    next_sibling_id TEXT REFERENCES content_data ON DELETE SET NULL,
    prev_sibling_id TEXT REFERENCES content_data ON DELETE SET NULL,
    route_id TEXT NOT NULL REFERENCES routes ON DELETE CASCADE,
    datatype_id TEXT NOT NULL REFERENCES datatypes ON DELETE SET NULL,
    author_id TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: content_data_id is ULID primary key, parent_id points to parent node, first_child_id points to first child, next_sibling_id and prev_sibling_id form doubly-linked sibling list, route_id associates with URL route, datatype_id specifies content type, author_id tracks creator, status is draft, published, archived, or pending.

Foreign keys: tree pointers reference content_data self, route_id references routes.route_id with RESTRICT on delete to prevent orphaned content, datatype_id references datatypes.datatype_id with RESTRICT, author_id references users.user_id.

Indexes: idx_content_data_parent on parent_id, idx_content_data_route on route_id, idx_content_data_datatype on datatype_id, idx_content_data_author on author_id.

### content_fields

Field values for content records.

```sql
CREATE TABLE content_fields (
    content_field_id TEXT PRIMARY KEY CHECK (length(content_field_id) = 26),
    content_data_id TEXT NOT NULL REFERENCES content_data ON DELETE CASCADE,
    field_id TEXT NOT NULL REFERENCES fields ON DELETE CASCADE,
    value TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: content_field_id is ULID primary key, content_data_id references the content record, field_id references the field definition, value stores the field data as TEXT.

Foreign keys: content_data_id references content_data.content_data_id with CASCADE delete, field_id references fields.field_id with CASCADE delete.

### content_relations

Many-to-many relationships between content records through relation fields.

```sql
CREATE TABLE content_relations (
    content_relation_id TEXT PRIMARY KEY CHECK (length(content_relation_id) = 26),
    source_content_id TEXT NOT NULL REFERENCES content_data ON DELETE CASCADE,
    target_content_id TEXT NOT NULL REFERENCES content_data ON DELETE CASCADE,
    field_id TEXT NOT NULL REFERENCES fields ON DELETE CASCADE,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: content_relation_id is ULID primary key, source_content_id is the content initiating the relation, target_content_id is the related content, field_id specifies which relation field defines this link.

Foreign keys: all foreign keys CASCADE on delete to remove relations when source, target, or field definition is deleted.

### media

Media assets with metadata and S3 storage references.

```sql
CREATE TABLE media (
    media_id TEXT PRIMARY KEY CHECK (length(media_id) = 26),
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT UNIQUE,
    srcset TEXT,
    author_id TEXT NOT NULL REFERENCES users ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: media_id is ULID primary key, name is filename, display_name is human-readable label, alt is image alt text, caption and description provide metadata, class supports CSS classification, mimetype stores MIME type, dimensions stores image dimensions, url is S3 object URL with unique constraint, srcset stores responsive image URLs, author_id tracks uploader.

Foreign keys: author_id references users.user_id.

Indexes: idx_media_author on author_id.

### media_dimension

Responsive image size variants for media assets.

```sql
CREATE TABLE media_dimension (
    media_dimension_id TEXT PRIMARY KEY CHECK (length(media_dimension_id) = 26),
    media_id TEXT NOT NULL REFERENCES media ON DELETE CASCADE,
    width INTEGER,
    height INTEGER,
    url TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: media_dimension_id is ULID primary key, media_id references parent media record, width and height specify variant dimensions, url points to resized image.

Foreign keys: media_id references media.media_id with CASCADE delete.

### sessions

User session tracking for authentication.

```sql
CREATE TABLE sessions (
    session_id TEXT PRIMARY KEY CHECK (length(session_id) = 26),
    user_id TEXT NOT NULL REFERENCES users ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    last_activity TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: session_id is ULID primary key, user_id references authenticated user, token is unique session token, expires_at specifies expiration timestamp, last_activity tracks most recent session use.

Foreign keys: user_id references users.user_id with CASCADE delete.

### tokens

API tokens for programmatic access.

```sql
CREATE TABLE tokens (
    token_id TEXT PRIMARY KEY CHECK (length(token_id) = 26),
    user_id TEXT NOT NULL REFERENCES users ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    description TEXT,
    expires_at TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: token_id is ULID primary key, user_id references owning user, token is unique API token string, description labels the token purpose, expires_at is optional expiration timestamp.

Foreign keys: user_id references users.user_id with CASCADE delete.

### user_oauth

OAuth provider associations for users.

```sql
CREATE TABLE user_oauth (
    user_oauth_id TEXT PRIMARY KEY CHECK (length(user_oauth_id) = 26),
    user_id TEXT NOT NULL REFERENCES users ON DELETE CASCADE,
    provider TEXT NOT NULL,
    oauth_provider_user_id TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: user_oauth_id is ULID primary key, user_id references ModulaCMS user, provider specifies OAuth provider name, oauth_provider_user_id is provider's user ID, access_token and refresh_token store OAuth credentials, expires_at tracks token expiration.

Foreign keys: user_id references users.user_id with CASCADE delete.

### user_ssh_keys

SSH public keys for TUI access.

```sql
CREATE TABLE user_ssh_keys (
    user_ssh_key_id TEXT PRIMARY KEY CHECK (length(user_ssh_key_id) = 26),
    user_id TEXT NOT NULL REFERENCES users ON DELETE CASCADE,
    label TEXT NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint TEXT NOT NULL UNIQUE,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: user_ssh_key_id is ULID primary key, user_id references key owner, label is human-readable key name, public_key stores SSH public key, fingerprint is unique SSH key fingerprint.

Foreign keys: user_id references users.user_id with CASCADE delete.

## Admin Tables

Admin tables mirror content tables for draft/staging environments. They use admin_ prefix and separate IDs but follow identical structure to their production counterparts.

### admin_routes

Staging routes mirroring routes table structure.

### admin_datatypes

Staging datatypes mirroring datatypes table structure.

### admin_fields

Staging fields mirroring fields table structure.

### admin_content_data

Staging content mirroring content_data table structure including tree pointers.

### admin_content_fields

Staging content field values mirroring content_fields table structure.

### admin_content_relations

Staging content relations mirroring content_relations table structure.

## Join Tables

### datatypes_fields

Junction table associating fields with datatypes.

```sql
CREATE TABLE datatypes_fields (
    datatype_field_id TEXT PRIMARY KEY CHECK (length(datatype_field_id) = 26),
    datatype_id TEXT NOT NULL REFERENCES datatypes ON DELETE CASCADE,
    field_id TEXT NOT NULL REFERENCES fields ON DELETE CASCADE,
    display_order INTEGER DEFAULT 0,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: datatype_field_id is ULID primary key, datatype_id references datatype, field_id references field, display_order controls field ordering in UI.

Foreign keys: both foreign keys CASCADE on delete.

### admin_datatypes_fields

Junction table for admin datatypes and fields.

## Distributed System Tables

### change_events

Audit trail and replication log using Hybrid Logical Clock ordering.

```sql
CREATE TABLE change_events (
    event_id TEXT PRIMARY KEY CHECK (length(event_id) = 26),
    hlc_timestamp INTEGER NOT NULL,
    wall_timestamp TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    node_id TEXT NOT NULL CHECK (length(node_id) = 26),
    table_name TEXT NOT NULL,
    record_id TEXT NOT NULL CHECK (length(record_id) = 26),
    operation TEXT NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    action TEXT,
    user_id TEXT CHECK (user_id IS NULL OR length(user_id) = 26),
    old_values TEXT,
    new_values TEXT,
    metadata TEXT,
    request_id TEXT,
    ip TEXT,
    synced_at TEXT,
    consumed_at TEXT
);
```

Columns: event_id is ULID primary key, hlc_timestamp provides distributed ordering via HLC, wall_timestamp is physical timestamp, node_id identifies which node recorded the event, table_name and record_id specify affected record, operation is INSERT UPDATE or DELETE, action provides semantic context, user_id tracks who made the change, old_values and new_values store JSON snapshots, metadata holds additional context, request_id correlates with HTTP requests, ip tracks client address, synced_at and consumed_at track replication.

Indexes: idx_events_record on table_name and record_id, idx_events_hlc on hlc_timestamp, idx_events_node on node_id, idx_events_user on user_id.

## Backup Tables

### backups

Backup metadata tracking full, incremental, and differential backups.

```sql
CREATE TABLE backups (
    backup_id TEXT PRIMARY KEY CHECK (length(backup_id) = 26),
    node_id TEXT NOT NULL CHECK (length(node_id) = 26),
    backup_type TEXT NOT NULL CHECK (backup_type IN ('full', 'incremental', 'differential')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    started_at TEXT NOT NULL,
    completed_at TEXT,
    duration_ms INTEGER,
    record_count INTEGER,
    size_bytes INTEGER,
    replication_lsn TEXT,
    hlc_timestamp INTEGER,
    storage_path TEXT NOT NULL,
    checksum TEXT,
    triggered_by TEXT,
    error_message TEXT,
    metadata TEXT
);
```

Columns: backup_id is ULID primary key, node_id identifies backup source, backup_type is full incremental or differential, status tracks backup progress, started_at and completed_at bound execution time, duration_ms measures performance, record_count and size_bytes track backup size, replication_lsn and hlc_timestamp support distributed systems, storage_path points to backup file, checksum verifies integrity, triggered_by tracks initiator, error_message captures failures, metadata stores JSON context.

Indexes: idx_backups_node on node_id and started_at DESC, idx_backups_status on status and started_at DESC, idx_backups_hlc on hlc_timestamp.

### backup_verifications

Backup verification results.

```sql
CREATE TABLE backup_verifications (
    verification_id TEXT PRIMARY KEY CHECK (length(verification_id) = 26),
    backup_id TEXT NOT NULL REFERENCES backups ON DELETE CASCADE,
    verified_at TEXT NOT NULL,
    verified_by TEXT,
    restore_tested INTEGER DEFAULT 0,
    checksum_valid INTEGER DEFAULT 0,
    record_count_match INTEGER DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('pending', 'verified', 'failed')),
    error_message TEXT,
    duration_ms INTEGER
);
```

Columns: verification_id is ULID primary key, backup_id references verified backup, verified_at is verification timestamp, verified_by identifies verifier, restore_tested checksum_valid record_count_match are boolean flags stored as INTEGER, status indicates verification outcome, error_message captures failures, duration_ms measures verification time.

Foreign keys: backup_id references backups.backup_id with CASCADE delete.

Indexes: idx_verifications_backup on backup_id and verified_at DESC.

### backup_sets

Coordinated multi-node backup sets.

```sql
CREATE TABLE backup_sets (
    backup_set_id TEXT PRIMARY KEY CHECK (length(backup_set_id) = 26),
    created_at TEXT NOT NULL,
    hlc_timestamp INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'complete', 'partial')),
    backup_ids TEXT NOT NULL,
    node_count INTEGER NOT NULL,
    completed_count INTEGER DEFAULT 0,
    error_message TEXT
);
```

Columns: backup_set_id is ULID primary key, created_at is set creation time, hlc_timestamp provides distributed ordering, status tracks completion, backup_ids stores JSON array of backup IDs in this set, node_count is expected backup count, completed_count tracks actual completions, error_message captures failures.

Indexes: idx_backup_sets_time on created_at DESC, idx_backup_sets_hlc on hlc_timestamp.

## Utility Tables

### tables

Metadata tracking for all tables in the schema.

```sql
CREATE TABLE tables (
    table_id TEXT PRIMARY KEY CHECK (length(table_id) = 26),
    table_name TEXT NOT NULL UNIQUE,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Columns: table_id is ULID primary key, table_name stores table name, timestamps track schema changes.

## Query Patterns

All tables support standard CRUD operations through sqlc-generated functions. Query patterns are consistent across tables.

### Count Queries

`CountTableName` returns total record count.

```sql
-- name: CountUser :one
SELECT COUNT(*) FROM users;
```

### Get Queries

`GetTableName` retrieves single record by primary key, returns one row or error.

```sql
-- name: GetUser :one
SELECT * FROM users WHERE user_id = ? LIMIT 1;
```

### List Queries

`ListTableName` retrieves all records ordered by primary key, returns many rows.

```sql
-- name: ListUser :many
SELECT * FROM users ORDER BY user_id;
```

### List Paginated Queries

`ListTableNamePaginated` supports pagination with LIMIT and OFFSET.

```sql
-- name: ListUserPaginated :many
SELECT * FROM users ORDER BY user_id LIMIT ? OFFSET ?;
```

### Create Queries

`CreateTableName` inserts new record and returns created row with generated timestamps.

```sql
-- name: CreateUser :one
INSERT INTO users (...) VALUES (...) RETURNING *;
```

### Update Queries

`UpdateTableName` modifies existing record by primary key.

```sql
-- name: UpdateUser :exec
UPDATE users SET ... WHERE user_id = ?;
```

### Delete Queries

`DeleteTableName` removes record by primary key.

```sql
-- name: DeleteUser :exec
DELETE FROM users WHERE user_id = ?;
```

### Specialized Queries

Many tables include domain-specific queries. Examples: ListDatatypeGlobal filters by type, ListContentDataByRoute filters by route_id, GetLatestBackup finds most recent completed backup.

## Database-Specific Files

Each schema directory contains three variants for database compatibility.

SQLite files: schema.sql and queries.sql use TEXT for IDs and timestamps, integer for booleans, no AUTO_INCREMENT.

MySQL files: schema_mysql.sql and queries_mysql.sql use VARCHAR for IDs and timestamps, DATETIME for timestamps, TINYINT for booleans.

PostgreSQL files: schema_psql.sql and queries_psql.sql use TEXT for IDs, TIMESTAMP for timestamps, BOOLEAN for booleans.

The sqlc configuration maps all variants to unified Go types defined in internal/db/types.

## Type Overrides

SQLC configuration maps database columns to custom Go types. All ID columns map to ULID-based types from internal/db/types. All timestamp columns map to types.Timestamp. Enum columns map to validated string types. Nullable foreign keys map to Nullable wrapper types.

Example overrides: users.user_id maps to types.UserID, users.email maps to types.Email, users.role_id maps to types.NullableRoleID, content_data.status maps to types.ContentStatus, change_events.hlc_timestamp maps to types.HLC.

## Tree Structure

The content_data table implements a tree using sibling pointers for O(1) operations. Each node stores parent_id first_child_id next_sibling_id prev_sibling_id. The Go code builds a NodeIndex map for direct lookup and implements three-phase loading: create nodes, assign hierarchy, resolve orphans.

Circular reference detection prevents infinite loops during orphan resolution. Lazy loading supports large trees by loading children on demand.

## Triggers

All tables with date_modified columns include UPDATE triggers that automatically set date_modified to current UTC timestamp on row modification.

Example trigger for users table:

```sql
CREATE TRIGGER update_users_modified
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    UPDATE users SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE user_id = NEW.user_id;
END;
```

## Migration Workflow

Schema changes follow numbered directory convention. To add a new table, create schema/N_tablename/ with schema.sql schema_mysql.sql schema_psql.sql queries.sql queries_mysql.sql queries_psql.sql. Add type overrides to sqlc.yml. Run just sqlc to regenerate Go code. Add DbDriver interface methods in internal/db/db.go. Implement mapper functions in internal/db/tablename.go for all three database drivers.
