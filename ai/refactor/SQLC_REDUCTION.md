# SQLC Configuration Refactor: Type Unification

## Status: Active Plan (HQ Multi-Agent)

## Problem Statement

The current database layer has **34,589 lines of code**:
- `internal/db/`: 16,795 lines (manual wrapper/mapping layer)
- `internal/db-sqlite/`: 6,168 lines (sqlc generated)
- `internal/db-mysql/`: 5,859 lines (sqlc generated)
- `internal/db-psql/`: 5,767 lines (sqlc generated)

**Root causes:**
1. sqlc generates different Go types per engine (int32 vs int64, sql.NullTime vs sql.NullString)
2. Duplicate struct variants: `*JSON`, `*FormParams` for each model
3. Complex type conversion functions throughout mapping code
4. No compile-time type safety for IDs (can pass UserID where DatatypeID expected)
5. No centralized validation for API/CLI inputs
6. Inconsistent null handling and JSON serialization

## Goals

1. **Simplify** - Reduce wrapper boilerplate by 45-50%
2. **Type Safety** - Compile-time prevention of ID mixups
3. **Validation** - Centralized input validation for API and CLI
4. **Reliability** - Enterprise-grade error handling and logging
5. **Consistency** - Unified behavior across all three database engines
6. **Defense in Depth** - Database constraints mirror Go validation
7. **Distribution-Ready** - ULIDs, HLC timestamps, change events for future multi-node support
8. **Audit Trail** - change_events table serves as audit log, replication log, and webhook source

## Architecture Constraint

Go uses **nominal typing**. Even if sqlc generates byte-identical structs across engines, they remain different types. **The wrapper layer is architecturally necessary** for runtime database switching.

However, by using **custom types defined in `internal/db/`**, all three sqlc packages will use the SAME Go types for fields, enabling:
- Compile-time type safety
- Centralized validation
- Consistent JSON serialization
- Specific error messages

```
Application Code (API, CLI, etc.)
         │
         ▼ uses db.DbDriver, db.Datatypes, db.DatatypeID
┌─────────────────────────────────────────────────────────┐
│                    internal/db/                          │
│                                                          │
│  Custom Types (shared by ALL packages):                  │
│  ├── DatatypeID, UserID, ContentID, etc.                │
│  ├── NullableDatatypeID, NullableUserID, etc.           │
│  ├── Timestamp (unified datetime)                        │
│  ├── ContentStatus, FieldType (enums)                   │
│  └── Slug, Email, URL, History (validation types)       │
│                                                          │
│  Wrapper Layer:                                          │
│  ├── db.Datatypes (unified struct)                      │
│  ├── db.DbDriver (interface)                            │
│  └── Database, MysqlDatabase, PsqlDatabase              │
└─────────────────────────────────────────────────────────┘
         │                │                │
         ▼                ▼                ▼
    mdb.Queries     mdbm.Queries     mdbp.Queries
      (sqlc)          (sqlc)           (sqlc)
         │                │                │
         └────── All use db.DatatypeID, db.Timestamp, etc.
```

---

## Schema Improvements (Do Now While Cost Is Zero)

These schema changes should be implemented in Phase 0/1 before the type system work. Adding them now is free; adding them later with production data requires migrations.

### 1. ULID Primary Keys (Distribution-Ready)

**Why ULIDs instead of auto-increment integers:**
- Globally unique without database coordination
- Sortable by creation time (useful for listings)
- 26 characters as string (reasonable URL length)
- No ID conflicts when merging databases
- Can generate on application server before INSERT

```sql
-- PostgreSQL: Use CHAR(26) for ULIDs
CREATE TABLE datatypes (
    datatype_id CHAR(26) PRIMARY KEY,  -- ULID as string
    -- ... other columns
);

-- MySQL: Use CHAR(26)
CREATE TABLE datatypes (
    datatype_id CHAR(26) PRIMARY KEY,
    -- ... other columns
);

-- SQLite: Use TEXT
CREATE TABLE datatypes (
    datatype_id TEXT PRIMARY KEY CHECK (length(datatype_id) = 26),
    -- ... other columns
);
```

**Apply to ALL tables:** datatypes, users, roles, permissions, fields, content_data, media, sessions, tokens, routes, etc.

### 2. Node Identity Column (Multi-Node Ready)

Add `node_id` to all entity tables for future distributed support:

```sql
-- Add to every entity table (nullable for now, required in multi-node mode)
node_id CHAR(26),  -- ULID of the node that created/owns this record

CREATE INDEX idx_<table>_node ON <table>(node_id);
```

### 3. Change Events Table (Audit + Replication + Webhooks)

Replaces `change_events`. Serves three purposes:
1. **Audit trail** - Who changed what, when
2. **Replication log** - What changes to sync to other nodes
3. **Webhook source** - What events to fire

```sql
-- PostgreSQL
CREATE TABLE change_events (
    event_id CHAR(26) PRIMARY KEY,           -- ULID
    hlc_timestamp BIGINT NOT NULL,           -- Hybrid Logical Clock (for ordering)
    wall_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    node_id CHAR(26) NOT NULL,               -- Which node generated this event
    table_name VARCHAR(64) NOT NULL,
    record_id CHAR(26) NOT NULL,             -- ULID of affected record
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    action VARCHAR(20),                       -- Business action: 'publish', 'archive', etc.
    user_id CHAR(26),                         -- Who performed the action (ULID)
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,                           -- Webhook data, conflict info, etc.
    synced_at TIMESTAMP WITH TIME ZONE,      -- NULL = not yet synced to other nodes
    consumed_at TIMESTAMP WITH TIME ZONE     -- NULL = not yet processed by webhooks
);

CREATE INDEX idx_events_record ON change_events(table_name, record_id);
CREATE INDEX idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX idx_events_node ON change_events(node_id);
CREATE INDEX idx_events_user ON change_events(user_id);
CREATE INDEX idx_events_unsynced ON change_events(synced_at) WHERE synced_at IS NULL;
CREATE INDEX idx_events_unconsumed ON change_events(consumed_at) WHERE consumed_at IS NULL;

-- MySQL
CREATE TABLE change_events (
    event_id CHAR(26) PRIMARY KEY,
    hlc_timestamp BIGINT NOT NULL,
    wall_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    node_id CHAR(26) NOT NULL,
    table_name VARCHAR(64) NOT NULL,
    record_id CHAR(26) NOT NULL,
    operation VARCHAR(20) NOT NULL,
    action VARCHAR(20),
    user_id CHAR(26),
    old_values JSON,
    new_values JSON,
    metadata JSON,
    synced_at TIMESTAMP NULL,
    consumed_at TIMESTAMP NULL,
    CONSTRAINT chk_operation CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE'))
);

CREATE INDEX idx_events_record ON change_events(table_name, record_id);
CREATE INDEX idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX idx_events_node ON change_events(node_id);
CREATE INDEX idx_events_user ON change_events(user_id);
CREATE INDEX idx_events_unsynced ON change_events((synced_at IS NULL));
CREATE INDEX idx_events_unconsumed ON change_events((consumed_at IS NULL));

-- SQLite
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
    old_values TEXT,  -- JSON
    new_values TEXT,  -- JSON
    metadata TEXT,    -- JSON
    synced_at TEXT,
    consumed_at TEXT
);

CREATE INDEX idx_events_record ON change_events(table_name, record_id);
CREATE INDEX idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX idx_events_node ON change_events(node_id);
CREATE INDEX idx_events_user ON change_events(user_id);
```

**Note:** Remove `history` TEXT column from all entity tables (no migration needed - no production data).

### 4. HLC Timestamp Column

Add Hybrid Logical Clock timestamp alongside `date_modified` for distributed ordering:

```sql
-- Add to every entity table
hlc_modified BIGINT NOT NULL DEFAULT 0,  -- Hybrid Logical Clock for ordering

-- HLC encodes: (wall_time_ms << 16) | logical_counter
-- Allows total ordering even with clock skew between nodes
```

### 5. Conflict Policy (Per-Datatype)

Add to datatypes table for configurable conflict resolution:

```sql
-- Add to datatypes and admin_datatypes tables
conflict_policy VARCHAR(20) NOT NULL DEFAULT 'lww'
    CHECK (conflict_policy IN ('lww', 'manual')),
    -- 'lww' = Last Write Wins (simple, possible data loss)
    -- 'manual' = Flag conflicts for human resolution
```

### 6. Foreign Key Indexes

**Problem:** Missing FK indexes cause full table scans on JOINs. Adding indexes later on tables with millions of rows is expensive.

**Add to every schema file:**

```sql
-- content_data (note: FKs are now CHAR(26) ULIDs)
CREATE INDEX idx_content_data_author ON content_data(author_id);
CREATE INDEX idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX idx_content_data_parent ON content_data(parent_id);
CREATE INDEX idx_content_data_node ON content_data(node_id);

-- content_fields
CREATE INDEX idx_content_fields_content ON content_fields(content_data_id);
CREATE INDEX idx_content_fields_field ON content_fields(field_id);

-- datatypes_fields
CREATE INDEX idx_datatypes_fields_datatype ON datatypes_fields(datatype_id);
CREATE INDEX idx_datatypes_fields_field ON datatypes_fields(field_id);

-- media
CREATE INDEX idx_media_author ON media(author_id);
CREATE INDEX idx_media_node ON media(node_id);

-- sessions
CREATE INDEX idx_sessions_user ON sessions(user_id);

-- tokens
CREATE INDEX idx_tokens_user ON tokens(user_id);

-- user_oauth
CREATE INDEX idx_user_oauth_user ON user_oauth(user_id);

-- user_ssh_keys
CREATE INDEX idx_user_ssh_keys_user ON user_ssh_keys(user_id);

-- routes
CREATE INDEX idx_routes_content ON routes(content_data_id);

-- admin_* tables (same pattern)
CREATE INDEX idx_admin_content_data_author ON admin_content_data(author_id);
CREATE INDEX idx_admin_content_data_datatype ON admin_content_data(admin_datatype_id);
CREATE INDEX idx_admin_content_data_node ON admin_content_data(node_id);
-- etc.
```

### 7. Database Constraints (Defense in Depth)

**Principle:** Database should enforce same rules as Go types. If Go rejects it, DB should too.

```sql
-- PostgreSQL: Create DOMAINs matching Go types
CREATE DOMAIN slug AS VARCHAR(255)
    CHECK (VALUE ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$');

CREATE DOMAIN email AS VARCHAR(254)
    CHECK (VALUE ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$');

-- MySQL/SQLite: Use CHECK constraints inline
-- In table definitions:
slug VARCHAR(255) CHECK (slug REGEXP '^[a-z0-9]+(-[a-z0-9]+)*$'),
email VARCHAR(254) CHECK (email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$'),

-- Enum constraints (all databases)
status VARCHAR(20) NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft', 'published', 'archived', 'pending')),

field_type VARCHAR(20) NOT NULL
    CHECK (field_type IN ('text', 'textarea', 'number', 'date', 'datetime',
           'boolean', 'select', 'media', 'relation', 'json', 'richtext',
           'slug', 'email', 'url')),

route_type VARCHAR(20) NOT NULL
    CHECK (route_type IN ('static', 'dynamic', 'api', 'redirect')),
```

### 8. Timestamp Standardization

**Rule:** All timestamps stored as UTC. One input format accepted.

```sql
-- PostgreSQL: Use TIMESTAMP WITH TIME ZONE
date_created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
date_modified TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

-- MySQL: Use TIMESTAMP (auto-converts to UTC)
date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
date_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

-- SQLite: Store as ISO8601 UTC string, use triggers for ON UPDATE
date_created TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
date_modified TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),

-- SQLite trigger for ON UPDATE behavior
CREATE TRIGGER update_<table>_modified
    AFTER UPDATE ON <table>
    FOR EACH ROW
    BEGIN
        UPDATE <table> SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE rowid = NEW.rowid;
    END;
```

**Go Timestamp type update:** Accept only RFC3339 with timezone:

```go
// Strict parsing - only accept RFC3339 with timezone
var timestampFormats = []string{
    time.RFC3339,      // "2006-01-02T15:04:05Z07:00" - PRIMARY FORMAT
    time.RFC3339Nano,  // "2006-01-02T15:04:05.999999999Z07:00"
}

// Legacy formats only for database reads, NOT for API input
var legacyTimestampFormats = []string{
    "2006-01-02 15:04:05",  // MySQL without TZ (assume UTC)
    "2006-01-02",           // Date only (assume 00:00:00 UTC)
}
```

### 9. Backup Tracking Tables (Distributed Restore)

Essential for distributed systems. Without this, you can't answer:
- "What's the latest backup for the London node?"
- "Can I restore NYC to match London's state from 6 hours ago?"
- "Did last night's backup job actually run?"
- "Which backup contains transaction X?"

```sql
-- PostgreSQL
CREATE TABLE backups (
    backup_id       CHAR(26) PRIMARY KEY,  -- ULID
    node_id         CHAR(26) NOT NULL,
    backup_type     VARCHAR(20) NOT NULL CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    status          VARCHAR(20) NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'verified')),

    -- Timing
    started_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at    TIMESTAMP WITH TIME ZONE,
    duration_ms     INTEGER,

    -- What's in it
    record_count    BIGINT,
    size_bytes      BIGINT,

    -- Replication position (critical for distributed restore)
    replication_lsn VARCHAR(64),           -- PostgreSQL LSN or MySQL binlog position
    hlc_timestamp   BIGINT,                -- HLC at backup time

    -- Where it lives
    storage_path    TEXT NOT NULL,         -- s3://bucket/backups/2026-01-20/...
    checksum        VARCHAR(64),           -- SHA256 of backup file

    -- Metadata
    triggered_by    VARCHAR(64),           -- 'scheduled', 'manual', 'pre-deploy', user_id
    error_message   TEXT,
    metadata        JSONB
);

CREATE INDEX idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX idx_backups_hlc ON backups(hlc_timestamp);

-- Backup verification tracking (backups that aren't tested aren't backups)
CREATE TABLE backup_verifications (
    verification_id  CHAR(26) PRIMARY KEY,
    backup_id        CHAR(26) NOT NULL REFERENCES backups(backup_id),
    verified_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_by      VARCHAR(64),          -- 'automated', user_id

    -- What was checked
    restore_tested   BOOLEAN DEFAULT FALSE,
    checksum_valid   BOOLEAN DEFAULT FALSE,
    record_count_match BOOLEAN DEFAULT FALSE,

    -- Results
    status           VARCHAR(20) NOT NULL CHECK (status IN ('passed', 'failed')),
    error_message    TEXT,
    duration_ms      INTEGER
);

CREATE INDEX idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

-- Coordinated backup sets (consistent snapshots across nodes)
CREATE TABLE backup_sets (
    backup_set_id    CHAR(26) PRIMARY KEY,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    hlc_timestamp    BIGINT NOT NULL,      -- Consistent point across all nodes
    status           VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed')),
    backup_ids       JSONB NOT NULL,       -- Array of backup_id ULIDs
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT
);

CREATE INDEX idx_backup_sets_time ON backup_sets(created_at DESC);
CREATE INDEX idx_backup_sets_hlc ON backup_sets(hlc_timestamp);

-- MySQL
CREATE TABLE backups (
    backup_id       CHAR(26) PRIMARY KEY,
    node_id         CHAR(26) NOT NULL,
    backup_type     VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP NULL,
    duration_ms     INTEGER,
    record_count    BIGINT,
    size_bytes      BIGINT,
    replication_lsn VARCHAR(64),
    hlc_timestamp   BIGINT,
    storage_path    TEXT NOT NULL,
    checksum        VARCHAR(64),
    triggered_by    VARCHAR(64),
    error_message   TEXT,
    metadata        JSON,
    CONSTRAINT chk_backup_type CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    CONSTRAINT chk_backup_status CHECK (status IN ('started', 'completed', 'failed', 'verified'))
);

CREATE INDEX idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE backup_verifications (
    verification_id  CHAR(26) PRIMARY KEY,
    backup_id        CHAR(26) NOT NULL,
    verified_at      TIMESTAMP NOT NULL,
    verified_by      VARCHAR(64),
    restore_tested   BOOLEAN DEFAULT FALSE,
    checksum_valid   BOOLEAN DEFAULT FALSE,
    record_count_match BOOLEAN DEFAULT FALSE,
    status           VARCHAR(20) NOT NULL,
    error_message    TEXT,
    duration_ms      INTEGER,
    CONSTRAINT chk_verification_status CHECK (status IN ('passed', 'failed')),
    FOREIGN KEY (backup_id) REFERENCES backups(backup_id)
);

CREATE INDEX idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE backup_sets (
    backup_set_id    CHAR(26) PRIMARY KEY,
    created_at       TIMESTAMP NOT NULL,
    hlc_timestamp    BIGINT NOT NULL,
    status           VARCHAR(20) NOT NULL,
    backup_ids       JSON NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT,
    CONSTRAINT chk_set_status CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_backup_sets_time ON backup_sets(created_at DESC);
CREATE INDEX idx_backup_sets_hlc ON backup_sets(hlc_timestamp);

-- SQLite
CREATE TABLE backups (
    backup_id       TEXT PRIMARY KEY CHECK (length(backup_id) = 26),
    node_id         TEXT NOT NULL CHECK (length(node_id) = 26),
    backup_type     TEXT NOT NULL CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    status          TEXT NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'verified')),
    started_at      TEXT NOT NULL,
    completed_at    TEXT,
    duration_ms     INTEGER,
    record_count    INTEGER,
    size_bytes      INTEGER,
    replication_lsn TEXT,
    hlc_timestamp   INTEGER,
    storage_path    TEXT NOT NULL,
    checksum        TEXT,
    triggered_by    TEXT,
    error_message   TEXT,
    metadata        TEXT  -- JSON
);

CREATE INDEX idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE backup_verifications (
    verification_id  TEXT PRIMARY KEY CHECK (length(verification_id) = 26),
    backup_id        TEXT NOT NULL REFERENCES backups(backup_id),
    verified_at      TEXT NOT NULL,
    verified_by      TEXT,
    restore_tested   INTEGER DEFAULT 0,
    checksum_valid   INTEGER DEFAULT 0,
    record_count_match INTEGER DEFAULT 0,
    status           TEXT NOT NULL CHECK (status IN ('passed', 'failed')),
    error_message    TEXT,
    duration_ms      INTEGER
);

CREATE INDEX idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE backup_sets (
    backup_set_id    TEXT PRIMARY KEY CHECK (length(backup_set_id) = 26),
    created_at       TEXT NOT NULL,
    hlc_timestamp    INTEGER NOT NULL,
    status           TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'failed')),
    backup_ids       TEXT NOT NULL,  -- JSON array
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT
);

CREATE INDEX idx_backup_sets_time ON backup_sets(created_at DESC);
CREATE INDEX idx_backup_sets_hlc ON backup_sets(hlc_timestamp);
```

**Replication Position** is critical for distributed restore - this tells you "This backup represents the state at replication position X":

```go
type Backup struct {
    BackupID       BackupID  `json:"backup_id"`
    NodeID         NodeID    `json:"node_id"`
    ReplicationLSN string    `json:"replication_lsn"` // PostgreSQL: "0/16B3748", MySQL: "mysql-bin.000003:12345"
    HLCTimestamp   HLC       `json:"hlc_timestamp"`
    // ...
}

// Restore node B to match node A's state at a specific point
func RestoreToPosition(targetNode string, sourceBackupID BackupID) error {
    backup, _ := db.GetBackup(sourceBackupID)
    // Restore the backup
    // Then replay replication stream up to backup.ReplicationLSN
    // Now targetNode matches sourceNode at that exact point
    return nil
}
```

---

### Primary ID Types (ULID-based)

Using [github.com/oklog/ulid/v2](https://github.com/oklog/ulid) for ULID generation.

```go
// internal/db/types/ids.go
package db

import (
    "crypto/rand"
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/oklog/ulid/v2"
)

// Thread-safe entropy source for ULID generation
var (
    entropyMu sync.Mutex
    entropy   = ulid.Monotonic(rand.Reader, 0)
)

// NewULID generates a new ULID (thread-safe)
func NewULID() ulid.ULID {
    entropyMu.Lock()
    defer entropyMu.Unlock()
    return ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
}

// DatatypeID uniquely identifies a datatype (26-char ULID string)
type DatatypeID string

// NewDatatypeID generates a new DatatypeID
func NewDatatypeID() DatatypeID {
    return DatatypeID(NewULID().String())
}

func (id DatatypeID) Validate() error {
    if id == "" {
        return fmt.Errorf("DatatypeID: cannot be empty")
    }
    if len(id) != 26 {
        return fmt.Errorf("DatatypeID: invalid length %d (expected 26)", len(id))
    }
    // Validate ULID format
    _, err := ulid.Parse(string(id))
    if err != nil {
        return fmt.Errorf("DatatypeID: invalid ULID format %q: %w", id, err)
    }
    return nil
}

func (id DatatypeID) String() string {
    return string(id)
}

// ParseDatatypeID parses a string into a DatatypeID (validates format)
func ParseDatatypeID(s string) (DatatypeID, error) {
    id := DatatypeID(s)
    if err := id.Validate(); err != nil {
        return "", err
    }
    return id, nil
}

// ULID returns the underlying ulid.ULID (for time extraction, comparison)
func (id DatatypeID) ULID() (ulid.ULID, error) {
    return ulid.Parse(string(id))
}

// Time returns the timestamp encoded in the ULID
func (id DatatypeID) Time() (time.Time, error) {
    u, err := id.ULID()
    if err != nil {
        return time.Time{}, err
    }
    return ulid.Time(u.Time()), nil
}

func (id DatatypeID) Value() (driver.Value, error) {
    if id == "" {
        return nil, fmt.Errorf("DatatypeID: cannot be empty")
    }
    return string(id), nil
}

func (id *DatatypeID) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("DatatypeID: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *id = DatatypeID(v)
    case []byte:
        *id = DatatypeID(string(v))
    default:
        return fmt.Errorf("DatatypeID: cannot scan %T", value)
    }
    return id.Validate()
}

func (id DatatypeID) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(id))
}

func (id *DatatypeID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return fmt.Errorf("DatatypeID: %w", err)
    }
    *id = DatatypeID(s)
    return id.Validate()
}

// IsZero returns true if the ID is empty
func (id DatatypeID) IsZero() bool {
    return id == ""
}

// Repeat pattern for all ID types:
// - UserID, RoleID, PermissionID, FieldID
// - ContentID, ContentFieldID, MediaID, MediaDimensionID
// - SessionID, TokenID, RouteID, AdminRouteID, TableID
// - UserOauthID, UserSshKeyID
// - AdminDatatypeID, AdminFieldID, AdminContentID, AdminContentFieldID
// - DatatypeFieldID, AdminDatatypeFieldID
// - EventID (for change_events)
// - NodeID (for multi-node support)
```

### Nullable ID Types (ULID-based)

```go
// internal/db/types/nullable_ids.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
)

// NullableDatatypeID represents a nullable foreign key to datatypes
type NullableDatatypeID struct {
    ID    DatatypeID
    Valid bool
}

func (n NullableDatatypeID) Validate() error {
    if n.Valid {
        return n.ID.Validate()
    }
    return nil
}

func (n NullableDatatypeID) String() string {
    if !n.Valid {
        return "NullableDatatypeID(null)"
    }
    return fmt.Sprintf("NullableDatatypeID(%s)", n.ID)
}

func (n NullableDatatypeID) Value() (driver.Value, error) {
    if !n.Valid {
        return nil, nil
    }
    return string(n.ID), nil  // Store as string
}

func (n *NullableDatatypeID) Scan(value any) error {
    if value == nil {
        n.Valid = false
        n.ID = ""
        return nil
    }
    n.Valid = true
    return n.ID.Scan(value)
}

func (n NullableDatatypeID) MarshalJSON() ([]byte, error) {
    if !n.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(n.ID)
}

func (n *NullableDatatypeID) UnmarshalJSON(data []byte) error {
    if string(data) == "null" {
        n.Valid = false
        n.ID = ""
        return nil
    }
    n.Valid = true
    return json.Unmarshal(data, &n.ID)
}

// IsZero returns true if null or empty
func (n NullableDatatypeID) IsZero() bool {
    return !n.Valid || n.ID == ""
}

// Repeat pattern for all nullable ID types:
// - NullableUserID
// - NullableRoleID
// - NullableContentID
// - NullableFieldID
// - NullableMediaID
// - NullableDatatypeID
// etc.
```

### Unified Timestamp

**Standardization rule:** All timestamps stored as UTC. One input format accepted (RFC3339 with timezone).

```go
// internal/db/types/timestamp.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
)

// Strict formats for API input - only RFC3339 with timezone
var strictTimestampFormats = []string{
    time.RFC3339,     // "2006-01-02T15:04:05Z07:00" - PRIMARY FORMAT
    time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
}

// Legacy formats for database reads only (historical data compatibility)
// These are NOT accepted for API input
var legacyTimestampFormats = []string{
    "2006-01-02 15:04:05",     // MySQL without TZ (assume UTC)
    "2006-01-02T15:04:05Z",    // UTC shorthand
    "2006-01-02T15:04:05",     // No TZ (assume UTC)
    "2006-01-02",              // Date only (assume 00:00:00 UTC)
}

// Timestamp handles datetime columns across SQLite (TEXT), MySQL (DATETIME), PostgreSQL (TIMESTAMP)
// All times are stored and returned in UTC.
type Timestamp struct {
    Time  time.Time
    Valid bool
}

func NewTimestamp(t time.Time) Timestamp {
    return Timestamp{Time: t.UTC(), Valid: true}
}

func TimestampNow() Timestamp {
    return Timestamp{Time: time.Now().UTC(), Valid: true}
}

func (t Timestamp) String() string {
    if !t.Valid {
        return "Timestamp(null)"
    }
    return fmt.Sprintf("Timestamp(%s)", t.Time.UTC().Format(time.RFC3339))
}

func (t Timestamp) Value() (driver.Value, error) {
    if !t.Valid {
        return nil, nil
    }
    return t.Time.UTC(), nil  // Always store as UTC
}

// Scan reads from database - accepts legacy formats for compatibility
func (t *Timestamp) Scan(value any) error {
    if value == nil {
        t.Valid = false
        return nil
    }
    switch v := value.(type) {
    case time.Time:
        t.Time, t.Valid = v.UTC(), true
        return nil
    case string:
        if v == "" {
            t.Valid = false
            return nil
        }
        // Try strict formats first
        for _, format := range strictTimestampFormats {
            if parsed, err := time.Parse(format, v); err == nil {
                t.Time, t.Valid = parsed.UTC(), true
                return nil
            }
        }
        // Fall back to legacy formats for database reads
        for _, format := range legacyTimestampFormats {
            if parsed, err := time.Parse(format, v); err == nil {
                t.Time, t.Valid = parsed.UTC(), true
                return nil
            }
        }
        return fmt.Errorf("Timestamp: cannot parse %q", v)
    case []byte:
        return t.Scan(string(v))
    default:
        return fmt.Errorf("Timestamp: cannot scan %T", value)
    }
}

// MarshalJSON always outputs RFC3339 in UTC
func (t Timestamp) MarshalJSON() ([]byte, error) {
    if !t.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(t.Time.UTC().Format(time.RFC3339))
}

// UnmarshalJSON ONLY accepts RFC3339 format with timezone - strict API input validation
func (t *Timestamp) UnmarshalJSON(data []byte) error {
    if string(data) == "null" {
        t.Valid = false
        return nil
    }
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return fmt.Errorf("Timestamp: expected string, got %s", string(data))
    }

    // STRICT: Only accept RFC3339 formats for API input
    for _, format := range strictTimestampFormats {
        if parsed, err := time.Parse(format, s); err == nil {
            t.Time, t.Valid = parsed.UTC(), true
            return nil
        }
    }

    // Reject legacy formats for API input
    return fmt.Errorf("Timestamp: invalid format %q (must be RFC3339: 2006-01-02T15:04:05Z or 2006-01-02T15:04:05-07:00)", s)
}

// IsZero returns true if timestamp is null or zero time
func (t Timestamp) IsZero() bool {
    return !t.Valid || t.Time.IsZero()
}

// Before reports whether t is before u
func (t Timestamp) Before(u Timestamp) bool {
    if !t.Valid || !u.Valid {
        return false
    }
    return t.Time.Before(u.Time)
}

// After reports whether t is after u
func (t Timestamp) After(u Timestamp) bool {
    if !t.Valid || !u.Valid {
        return false
    }
    return t.Time.After(u.Time)
}

// UTC returns the time in UTC (already stored as UTC, but explicit for clarity)
func (t Timestamp) UTC() time.Time {
    return t.Time.UTC()
}
```

### Hybrid Logical Clock (HLC)

For distributed ordering of events across nodes with clock skew.

```go
// internal/db/types/hlc.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

// HLC represents a Hybrid Logical Clock timestamp
// Format: (wall_time_ms << 16) | logical_counter
// - Upper 48 bits: milliseconds since Unix epoch
// - Lower 16 bits: logical counter (for ordering within same millisecond)
type HLC int64

var (
    hlcMu      sync.Mutex
    hlcLast    HLC
    hlcCounter uint16
)

// Now returns a new HLC timestamp greater than any previously issued
func HLCNow() HLC {
    hlcMu.Lock()
    defer hlcMu.Unlock()

    wallMs := time.Now().UnixMilli()
    physical := HLC(wallMs << 16)

    if physical > hlcLast {
        hlcLast = physical
        hlcCounter = 0
    } else {
        hlcCounter++
        if hlcCounter == 0 {
            // Counter overflow, wait for wall clock to advance
            time.Sleep(time.Millisecond)
            return HLCNow()
        }
    }

    hlcLast = physical | HLC(hlcCounter)
    return hlcLast
}

// Update merges a received HLC with local time (for receiving events from other nodes)
func HLCUpdate(received HLC) HLC {
    hlcMu.Lock()
    defer hlcMu.Unlock()

    wallMs := time.Now().UnixMilli()
    physical := HLC(wallMs << 16)

    maxPhysical := physical
    if received > maxPhysical {
        maxPhysical = received & ^HLC(0xFFFF) // Extract physical part
    }
    if hlcLast > maxPhysical {
        maxPhysical = hlcLast & ^HLC(0xFFFF)
    }

    if maxPhysical == hlcLast&^HLC(0xFFFF) {
        hlcCounter++
    } else {
        hlcCounter = 0
    }

    hlcLast = maxPhysical | HLC(hlcCounter)
    return hlcLast
}

func (h HLC) Physical() time.Time {
    ms := int64(h >> 16)
    return time.UnixMilli(ms)
}

func (h HLC) Logical() uint16 {
    return uint16(h & 0xFFFF)
}

func (h HLC) String() string {
    return fmt.Sprintf("HLC(%d:%d)", h>>16, h&0xFFFF)
}

func (h HLC) Value() (driver.Value, error) {
    return int64(h), nil
}

func (h *HLC) Scan(value any) error {
    if value == nil {
        *h = 0
        return nil
    }
    switch v := value.(type) {
    case int64:
        *h = HLC(v)
    case int:
        *h = HLC(v)
    default:
        return fmt.Errorf("HLC: cannot scan %T", value)
    }
    return nil
}

func (h HLC) MarshalJSON() ([]byte, error) {
    return json.Marshal(int64(h))
}

func (h *HLC) UnmarshalJSON(data []byte) error {
    var v int64
    if err := json.Unmarshal(data, &v); err != nil {
        return fmt.Errorf("HLC: %w", err)
    }
    *h = HLC(v)
    return nil
}

// Before returns true if h happened before other
func (h HLC) Before(other HLC) bool {
    return h < other
}

// After returns true if h happened after other
func (h HLC) After(other HLC) bool {
    return h > other
}
```

### Domain Enums

```go
// internal/db/types/enums.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
)

// ContentStatus represents the publication status of content
type ContentStatus string

const (
    ContentStatusDraft     ContentStatus = "draft"
    ContentStatusPublished ContentStatus = "published"
    ContentStatusArchived  ContentStatus = "archived"
    ContentStatusPending   ContentStatus = "pending"
)

func (s ContentStatus) Validate() error {
    switch s {
    case ContentStatusDraft, ContentStatusPublished, ContentStatusArchived, ContentStatusPending:
        return nil
    case "":
        return fmt.Errorf("ContentStatus: cannot be empty")
    default:
        return fmt.Errorf("ContentStatus: invalid value %q", s)
    }
}

func (s ContentStatus) String() string {
    return string(s)
}

func (s ContentStatus) Value() (driver.Value, error) {
    if err := s.Validate(); err != nil {
        return nil, err
    }
    return string(s), nil
}

func (s *ContentStatus) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("ContentStatus: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *s = ContentStatus(v)
    case []byte:
        *s = ContentStatus(string(v))
    default:
        return fmt.Errorf("ContentStatus: cannot scan %T", value)
    }
    return s.Validate()
}

func (s ContentStatus) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(s))
}

func (s *ContentStatus) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return fmt.Errorf("ContentStatus: %w", err)
    }
    *s = ContentStatus(str)
    return s.Validate()
}

// FieldType represents the type of a content field
type FieldType string

const (
    FieldTypeText     FieldType = "text"
    FieldTypeTextarea FieldType = "textarea"
    FieldTypeNumber   FieldType = "number"
    FieldTypeDate     FieldType = "date"
    FieldTypeDatetime FieldType = "datetime"
    FieldTypeBoolean  FieldType = "boolean"
    FieldTypeSelect   FieldType = "select"
    FieldTypeMedia    FieldType = "media"
    FieldTypeRelation FieldType = "relation"
    FieldTypeJSON     FieldType = "json"
    FieldTypeRichText FieldType = "richtext"
    FieldTypeSlug     FieldType = "slug"
    FieldTypeEmail    FieldType = "email"
    FieldTypeURL      FieldType = "url"
)

func (t FieldType) Validate() error {
    switch t {
    case FieldTypeText, FieldTypeTextarea, FieldTypeNumber, FieldTypeDate,
        FieldTypeDatetime, FieldTypeBoolean, FieldTypeSelect, FieldTypeMedia,
        FieldTypeRelation, FieldTypeJSON, FieldTypeRichText, FieldTypeSlug,
        FieldTypeEmail, FieldTypeURL:
        return nil
    case "":
        return fmt.Errorf("FieldType: cannot be empty")
    default:
        return fmt.Errorf("FieldType: invalid value %q", t)
    }
}

// Similar implementation for Scan, Value, MarshalJSON, UnmarshalJSON...

// RouteType represents the type of a route
type RouteType string

const (
    RouteTypeStatic  RouteType = "static"
    RouteTypeDynamic RouteType = "dynamic"
    RouteTypeAPI     RouteType = "api"
    RouteTypeRedirect RouteType = "redirect"
)

func (t RouteType) Validate() error {
    switch t {
    case RouteTypeStatic, RouteTypeDynamic, RouteTypeAPI, RouteTypeRedirect:
        return nil
    case "":
        return fmt.Errorf("RouteType: cannot be empty")
    default:
        return fmt.Errorf("RouteType: invalid value %q", t)
    }
}

// Similar implementation for Scan, Value, MarshalJSON, UnmarshalJSON...
```

### Validation Types

```go
// internal/db/types/validation.go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "net/url"
    "regexp"
    "strings"
    "unicode"
)

// Slug represents a URL-safe identifier (lowercase, hyphens, no spaces)
type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s Slug) Validate() error {
    if s == "" {
        return fmt.Errorf("Slug: cannot be empty")
    }
    if len(s) > 255 {
        return fmt.Errorf("Slug: too long (max 255 chars)")
    }
    if !slugRegex.MatchString(string(s)) {
        return fmt.Errorf("Slug: invalid format %q (must be lowercase alphanumeric with hyphens)", s)
    }
    return nil
}

func (s Slug) String() string {
    return string(s)
}

// Slugify converts a string to a valid slug
func Slugify(input string) Slug {
    // Lowercase
    result := strings.ToLower(input)
    // Replace spaces and underscores with hyphens
    result = strings.ReplaceAll(result, " ", "-")
    result = strings.ReplaceAll(result, "_", "-")
    // Remove non-alphanumeric except hyphens
    var sb strings.Builder
    for _, r := range result {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
            sb.WriteRune(r)
        }
    }
    result = sb.String()
    // Collapse multiple hyphens
    for strings.Contains(result, "--") {
        result = strings.ReplaceAll(result, "--", "-")
    }
    // Trim hyphens from ends
    result = strings.Trim(result, "-")
    return Slug(result)
}

func (s Slug) Value() (driver.Value, error) {
    return string(s), nil
}

func (s *Slug) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("Slug: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *s = Slug(v)
    case []byte:
        *s = Slug(string(v))
    default:
        return fmt.Errorf("Slug: cannot scan %T", value)
    }
    return s.Validate()
}

func (s Slug) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(s))
}

func (s *Slug) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return fmt.Errorf("Slug: %w", err)
    }
    *s = Slug(str)
    return s.Validate()
}

// Email represents a validated email address
type Email string

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (e Email) Validate() error {
    if e == "" {
        return fmt.Errorf("Email: cannot be empty")
    }
    if len(e) > 254 {
        return fmt.Errorf("Email: too long (max 254 chars)")
    }
    if !emailRegex.MatchString(string(e)) {
        return fmt.Errorf("Email: invalid format %q", e)
    }
    return nil
}

func (e Email) String() string {
    return string(e)
}

func (e Email) Domain() string {
    parts := strings.Split(string(e), "@")
    if len(parts) != 2 {
        return ""
    }
    return parts[1]
}

func (e Email) Value() (driver.Value, error) {
    return string(e), nil
}

func (e *Email) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("Email: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *e = Email(v)
    case []byte:
        *e = Email(string(v))
    default:
        return fmt.Errorf("Email: cannot scan %T", value)
    }
    return e.Validate()
}

func (e Email) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(e))
}

func (e *Email) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return fmt.Errorf("Email: %w", err)
    }
    *e = Email(str)
    return e.Validate()
}

// URL represents a validated URL
type URL string

func (u URL) Validate() error {
    if u == "" {
        return fmt.Errorf("URL: cannot be empty")
    }
    parsed, err := url.Parse(string(u))
    if err != nil {
        return fmt.Errorf("URL: invalid format %q: %w", u, err)
    }
    if parsed.Scheme == "" {
        return fmt.Errorf("URL: missing scheme in %q", u)
    }
    if parsed.Host == "" {
        return fmt.Errorf("URL: missing host in %q", u)
    }
    return nil
}

func (u URL) String() string {
    return string(u)
}

func (u URL) Parse() (*url.URL, error) {
    return url.Parse(string(u))
}

func (u URL) Value() (driver.Value, error) {
    return string(u), nil
}

func (u *URL) Scan(value any) error {
    if value == nil {
        return fmt.Errorf("URL: cannot be null")
    }
    switch v := value.(type) {
    case string:
        *u = URL(v)
    case []byte:
        *u = URL(string(v))
    default:
        return fmt.Errorf("URL: cannot scan %T", value)
    }
    return u.Validate()
}

func (u URL) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(u))
}

func (u *URL) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return fmt.Errorf("URL: %w", err)
    }
    *u = URL(str)
    return u.Validate()
}

// NullableSlug, NullableEmail, NullableURL for optional fields
type NullableSlug struct {
    Slug  Slug
    Valid bool
}

type NullableEmail struct {
    Email Email
    Valid bool
}

type NullableURL struct {
    URL   URL
    Valid bool
}

// Similar Scan, Value, MarshalJSON, UnmarshalJSON implementations...
```

### Change Events (Audit + Replication + Webhooks)

**Note:** The `change_events` table replaces JSON `history` columns and serves as:
1. **Audit trail** - Who changed what, when
2. **Replication log** - What to sync to other nodes
3. **Webhook source** - What events to fire

```go
// internal/db/types/events.go
package db

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
)

// Operation represents database operations
type Operation string

const (
    OpInsert Operation = "INSERT"
    OpUpdate Operation = "UPDATE"
    OpDelete Operation = "DELETE"
)

func (o Operation) Validate() error {
    switch o {
    case OpInsert, OpUpdate, OpDelete:
        return nil
    default:
        return fmt.Errorf("Operation: invalid value %q", o)
    }
}

// Action represents business-level actions (optional, more specific than Operation)
type Action string

const (
    ActionCreate  Action = "create"
    ActionUpdate  Action = "update"
    ActionDelete  Action = "delete"
    ActionPublish Action = "publish"
    ActionArchive Action = "archive"
)

// EventID uniquely identifies a change event (ULID)
type EventID string

func NewEventID() EventID {
    return EventID(NewULID().String())
}

// (Same methods as other ULID-based ID types)

// NodeID identifies a node in a distributed deployment (ULID)
type NodeID string

func NewNodeID() NodeID {
    return NodeID(NewULID().String())
}

// (Same methods as other ULID-based ID types)

// ChangeEvent represents a row in change_events table
type ChangeEvent struct {
    EventID       EventID        `json:"event_id"`
    HLCTimestamp  HLC            `json:"hlc_timestamp"`
    WallTimestamp Timestamp      `json:"wall_timestamp"`
    NodeID        NodeID         `json:"node_id"`
    TableName     string         `json:"table_name"`
    RecordID      string         `json:"record_id"`  // ULID of affected record
    Operation     Operation      `json:"operation"`
    Action        Action         `json:"action,omitempty"`
    UserID        NullableUserID `json:"user_id,omitempty"`
    OldValues     any            `json:"old_values,omitempty"`
    NewValues     any            `json:"new_values,omitempty"`
    Metadata      any            `json:"metadata,omitempty"`
    SyncedAt      Timestamp      `json:"synced_at,omitempty"`
    ConsumedAt    Timestamp      `json:"consumed_at,omitempty"`
}

// ConflictPolicy defines how conflicts are resolved for a datatype
type ConflictPolicy string

const (
    ConflictLWW    ConflictPolicy = "lww"    // Last Write Wins (simple, possible data loss)
    ConflictManual ConflictPolicy = "manual" // Flag conflicts for human resolution
)

func (c ConflictPolicy) Validate() error {
    switch c {
    case ConflictLWW, ConflictManual:
        return nil
    default:
        return fmt.Errorf("ConflictPolicy: invalid value %q", c)
    }
}

// EventLogger interface for change event operations
type EventLogger interface {
    // LogEvent records a change event
    LogEvent(ctx context.Context, event ChangeEvent) error

    // GetEventsByRecord retrieves events for a specific record
    GetEventsByRecord(ctx context.Context, tableName string, recordID string) ([]ChangeEvent, error)

    // GetEventsSince retrieves events after an HLC timestamp (for replication)
    GetEventsSince(ctx context.Context, hlc HLC, limit int) ([]ChangeEvent, error)

    // GetUnsyncedEvents retrieves events not yet synced to other nodes
    GetUnsyncedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

    // GetUnconsumedEvents retrieves events not yet processed by webhooks
    GetUnconsumedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

    // MarkSynced marks events as synced
    MarkSynced(ctx context.Context, eventIDs []EventID) error

    // MarkConsumed marks events as consumed by webhooks
    MarkConsumed(ctx context.Context, eventIDs []EventID) error
}

// NewChangeEvent creates a change event for the current node
func NewChangeEvent(nodeID NodeID, tableName string, recordID string, op Operation, action Action, userID UserID) ChangeEvent {
    return ChangeEvent{
        EventID:       NewEventID(),
        HLCTimestamp:  HLCNow(),
        WallTimestamp: TimestampNow(),
        NodeID:        nodeID,
        TableName:     tableName,
        RecordID:      recordID,
        Operation:     op,
        Action:        action,
        UserID:        NullableUserID{ID: userID, Valid: userID != ""},
    }
}

// WithChanges adds old/new values to the event
func (e ChangeEvent) WithChanges(oldVal, newVal any) ChangeEvent {
    e.OldValues = oldVal
    e.NewValues = newVal
    return e
}

// WithMetadata adds metadata to the event
func (e ChangeEvent) WithMetadata(meta any) ChangeEvent {
    e.Metadata = meta
    return e
}
```

**SQLC Queries for change_events:**

```sql
-- name: LogEvent :exec
INSERT INTO change_events (event_id, hlc_timestamp, node_id, table_name, record_id, operation, action, user_id, old_values, new_values, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetEventsByRecord :many
SELECT * FROM change_events
WHERE table_name = $1 AND record_id = $2
ORDER BY hlc_timestamp DESC;

-- name: GetEventsSince :many
SELECT * FROM change_events
WHERE hlc_timestamp > $1
ORDER BY hlc_timestamp ASC
LIMIT $2;

-- name: GetUnsyncedEvents :many
SELECT * FROM change_events
WHERE synced_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT $1;

-- name: GetUnconsumedEvents :many
SELECT * FROM change_events
WHERE consumed_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT $1;

-- name: MarkSynced :exec
UPDATE change_events SET synced_at = CURRENT_TIMESTAMP WHERE event_id = ANY($1);

-- name: MarkConsumed :exec
UPDATE change_events SET consumed_at = CURRENT_TIMESTAMP WHERE event_id = ANY($1);
```

### Transaction Helper

```go
// internal/db/types/transaction.go
package db

import (
    "context"
    "database/sql"
    "fmt"
)

// TxFunc executes within a transaction
type TxFunc func(tx *sql.Tx) error

// WithTransaction executes fn within a transaction with automatic commit/rollback
func WithTransaction(ctx context.Context, db *sql.DB, fn TxFunc) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback() // no-op if already committed

    if err := fn(tx); err != nil {
        return err
    }
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}

// WithTransactionResult executes fn within a transaction and returns a result
func WithTransactionResult[T any](ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) (T, error)) (T, error) {
    var result T
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return result, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    result, err = fn(tx)
    if err != nil {
        return result, err
    }
    if err := tx.Commit(); err != nil {
        return result, fmt.Errorf("commit transaction: %w", err)
    }
    return result, nil
}
```

---

## New sqlc.yml Configuration

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

---

## Implementation Phases

### Phase 0: Baseline Validation (Step 0)

#### Step 0: Validate Current SQLC Works
- **Scope:** validation
- **Tasks:**
  - Run `make sqlc` on current configuration
  - Document any existing generation errors
  - Run `make check` to verify current compile status
  - Document baseline line counts:
    ```bash
    wc -l internal/db/*.go internal/db-sqlite/*.go internal/db-mysql/*.go internal/db-psql/*.go
    ```
  - Create baseline commit marker
- **Artifact:** Baseline metrics documented in this plan (update "Before" column in Expected Results)

### Phase 0.5: Schema Improvements (Steps 0a-0d) **DO NOW - ZERO COST**

These schema changes are free now but expensive with production data. Complete before type system work.

#### Step 0a: Create change_events Table
- **Scope:** schema
- **Depends on:** [0]
- **Files:**
  - `sql/schema/0_change_events/schema.sql` (SQLite)
  - `sql/schema/0_change_events/schema_mysql.sql`
  - `sql/schema/0_change_events/schema_psql.sql`
  - `sql/schema/0_change_events/queries.sql` (SQLite)
  - `sql/schema/0_change_events/queries_mysql.sql`
  - `sql/schema/0_change_events/queries_psql.sql`
- **Tasks:**
  - Create change_events table per "Schema Improvements" section (audit + replication + webhooks)
  - Add all indexes (idx_change_hlc, idx_change_table_record, idx_change_node, idx_change_unsynced, idx_change_unconsumed)
  - Add SQLC queries: RecordChangeEvent, GetChangeEventsByRecord, GetUnsyncedEvents, MarkEventSynced, GetUnconsumedEvents, MarkEventConsumed
  - Run `make sqlc` to verify generation
- **Commit message:** "feat(schema): add change_events table for audit, replication, and webhooks"

#### Step 0b: Define Schema Without history Columns
- **Scope:** schema
- **Depends on:** [0a]
- **Files:** All schema files in `sql/schema/*/`
- **Tasks:**
  - Do NOT include `history TEXT` column in entity table schemas
  - Tables affected: datatypes, content_data, admin_content_data, users, media, etc.
  - History/audit tracking now via change_events table
  - No migration needed (no production data)
  - Run `make sqlc` to regenerate code
  - Run `make check` to find compilation errors from missing field
- **Note:** Wrapper layer will need updates to use change_events instead of history (handled in Phase 3)
- **Commit message:** "refactor(schema): define schema without history TEXT columns (use change_events)"

#### Step 0c: Add Foreign Key Indexes
- **Scope:** schema
- **Depends on:** [0]
- **Files:** All schema files
- **Tasks:**
  - Add indexes per "Foreign Key Indexes" section:
    - content_data: author_id, datatype_id, parent_id
    - content_fields: content_data_id, field_id
    - datatypes_fields: datatype_id, field_id
    - media: author_id
    - sessions: user_id
    - tokens: user_id
    - user_oauth: user_id
    - user_ssh_keys: user_id
    - routes: content_data_id
    - All admin_* tables: corresponding FK columns
  - Run `make sqlc` to verify
- **Commit message:** "perf(schema): add FK indexes for query optimization"

#### Step 0d: Add Database Constraints
- **Scope:** schema
- **Depends on:** [0]
- **Files:** All schema files
- **Tasks:**
  - Add CHECK constraints for:
    - `status` column: `CHECK (status IN ('draft', 'published', 'archived', 'pending'))`
    - `field_type` column: `CHECK (field_type IN (...))`
    - `route_type` column: `CHECK (route_type IN (...))`
  - Add PostgreSQL DOMAINs for slug, email (optional - CHECK constraints work too)
  - Add SQLite triggers for `date_modified` ON UPDATE behavior
  - Run `make sqlc` to verify
- **Commit message:** "feat(schema): add CHECK constraints matching Go validation types"

#### Step 0e: Create Backup Tables
- **Scope:** schema
- **Depends on:** [0]
- **Files:**
  - `sql/schema/0_backups/schema.sql` (SQLite)
  - `sql/schema/0_backups/schema_mysql.sql`
  - `sql/schema/0_backups/schema_psql.sql`
  - `sql/schema/0_backups/queries.sql` (SQLite)
  - `sql/schema/0_backups/queries_mysql.sql`
  - `sql/schema/0_backups/queries_psql.sql`
- **Tasks:**
  - Create backups, backup_verifications, backup_sets tables per "Schema Improvements" section
  - Add all indexes for node/status/hlc queries
  - Add SQLC queries: CreateBackup, UpdateBackupStatus, GetBackupsByNode, GetLatestBackup, CreateVerification, GetBackupSet, CreateBackupSet, etc.
  - Run `make sqlc` to verify generation
- **Commit message:** "feat(schema): add backup tables for distributed restore coordination"

### Phase 1: Custom Type System (Steps 5-12)

#### Step 5: Create Feature Branch
- **Scope:** setup
- **Depends on:** [1, 2, 3, 4, 0e]
- **Tasks:** Create `feature/sqlc-type-unification` branch from main (includes schema changes)

#### Step 6: Create Primary ID Types (ULID-based)
- **Scope:** types
- **Depends on:** [5]
- **File:** `internal/db/types_ids.go`
- **Tasks:** Implement all primary ID types as ULID-based strings (26 characters) with validation, Scan, Value, JSON methods
- **IDs to create:**
  - DatatypeID, UserID, RoleID, PermissionID, FieldID
  - ContentID, ContentFieldID, MediaID, MediaDimensionID
  - SessionID, TokenID, RouteID, AdminRouteID, TableID
  - UserOauthID, UserSshKeyID
  - AdminDatatypeID, AdminFieldID, AdminContentID, AdminContentFieldID
  - DatatypeFieldID, AdminDatatypeFieldID
  - **EventID** (for change_events table)
  - **NodeID** (for distributed node identity)
- **ULID Implementation:**
  - All IDs are `type XID string` (not int64)
  - `NewXID()` generates using ULID with monotonic entropy
  - `Validate()` checks length (26) and ULID format
  - Thread-safe generation with sync.Mutex

#### Step 7: Create Nullable ID Types
- **Scope:** types
- **Depends on:** [5]
- **File:** `internal/db/types_nullable_ids.go`
- **Tasks:** Implement nullable variants for all FK relationships
- **Types to create:** NullableXID for each ID type that appears as a foreign key

#### Step 8: Create Timestamp Type
- **Scope:** types
- **Depends on:** [5]
- **File:** `internal/db/types_timestamp.go`
- **Tasks:** Implement unified Timestamp type (UTC only, strict RFC3339 input)

#### Step 9: Create Enum Types
- **Scope:** types
- **Depends on:** [5]
- **File:** `internal/db/types_enums.go`
- **Tasks:** Implement ContentStatus, FieldType, RouteType with validation
- **Distributed System Enums:**
  - **Operation**: INSERT, UPDATE, DELETE (for change_events)
  - **Action**: semantic action names (create_datatype, update_content, publish, etc.)
  - **ConflictPolicy**: lww, manual (per-datatype conflict resolution)

#### Step 10: Create Validation Types
- **Scope:** types
- **Depends on:** [5]
- **File:** `internal/db/types_validation.go`
- **Tasks:** Implement Slug, Email, URL, NullableSlug, NullableEmail, NullableURL

#### Step 11: Create Change Event and HLC Types
- **Scope:** types
- **Depends on:** [5]
- **Files:**
  - `internal/db/types_hlc.go` - Hybrid Logical Clock
  - `internal/db/types_change_events.go` - Change event struct and logger
- **HLC Tasks:**
  - Implement HLC as int64 (48-bit wall time ms + 16-bit counter)
  - HLCNow() with thread-safe monotonic counter
  - Physical() to extract wall time
  - Compare() for event ordering
- **Change Event Tasks:**
  - Implement ChangeEvent struct (audit + replication + webhooks)
  - ChangeEventLogger interface with RecordEvent(), GetByRecord(), GetUnsynced(), etc.
  - Helper functions for creating events from entity operations

#### Step 12: Create Transaction Helper
- **Scope:** types
- **Depends on:** [5]
- **File:** `internal/db/transaction.go`
- **Tasks:** Implement WithTransaction and WithTransactionResult

### Phase 2: Configuration & Generation (Steps 13-16)

#### Step 13: Update sqlc.yml
- **Scope:** config
- **Depends on:** [6, 7, 8, 9, 10, 11]
- **File:** `sql/sqlc.yml`
- **Tasks:** Replace with new configuration above

#### Step 14: Check for Circular Imports
- **Scope:** validation
- **Depends on:** [13]
- **Tasks:**
  - Verify `internal/db/` types can import `database/sql/driver` and `encoding/json`
  - Verify `internal/db/` does NOT import `internal/db-sqlite`, `internal/db-mysql`, or `internal/db-psql`
  - If circular import detected, move types to `internal/db/types/` subpackage
  - Run `go build ./internal/db/...` to verify
- **Contingency:** If circular imports exist, create `internal/db/types/` package and update sqlc.yml import paths

#### Step 15: Regenerate All SQLC Code
- **Scope:** codegen
- **Depends on:** [14]
- **Tasks:**
  - Run `make sqlc`
  - Fix any generation errors (may need to adjust column names in config)
  - Verify all three packages generate: `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/`
  - Commit generated code

#### Step 16: Validate Type Usage
- **Scope:** validation
- **Depends on:** [15]
- **Tasks:**
  - Create validation script to verify custom types are used
  - Check that no raw int64 appears for ID columns
  - Check that Timestamp is used for date columns
  - Check that history columns are gone (change_events used instead)
  - Document any columns that need manual override

### Phase 3: Simplify Wrapper Layer (Steps 17a-38)

#### Step 17a: Update Struct Definitions to Use Custom Types
- **Scope:** wrapper
- **Depends on:** [16]
- **File:** `internal/db/` - all table files
- **Tasks:**
  - Update `db.Datatypes` etc. to use custom types:
    ```go
    // BEFORE
    type Datatypes struct {
        DatatypeID   int64          `json:"datatype_id"`
        ParentID     sql.NullInt64  `json:"parent_id"`
        AuthorID     int64          `json:"author_id"`
        DateCreated  sql.NullString `json:"date_created"`
        History      sql.NullString `json:"history"`  // REMOVED
    }

    // AFTER
    type Datatypes struct {
        DatatypeID   DatatypeID          `json:"datatype_id"`
        ParentID     NullableDatatypeID  `json:"parent_id"`
        AuthorID     UserID              `json:"author_id"`
        DateCreated  Timestamp           `json:"date_created"`
        // History field REMOVED - use AuditLogger interface
    }
    ```
  - Apply same transformation to all table structs
  - Run `make check` to verify compile

#### Step 17b: Remove *JSON Struct Variants
- **Scope:** wrapper
- **Depends on:** [17a]
- **File:** `internal/db/` - all table files
- **Tasks:**
  - Delete all `*JSON` struct definitions (e.g., `DatatypesJSON`, `UsersJSON`)
  - Delete all functions returning `*JSON` variants
  - Update any code referencing `*JSON` types
  - Run `make check` to verify compile

#### Step 17c: Remove *FormParams Struct Variants
- **Scope:** wrapper
- **Depends on:** [17a]
- **File:** `internal/db/` - all table files
- **Tasks:**
  - Delete all `*FormParams` struct definitions (e.g., `CreateDatatypeFormParams`)
  - Delete all functions accepting `*FormParams` types
  - Update any code referencing `*FormParams` types
  - Run `make check` to verify compile

**Note:** Steps 17b and 17c can run in parallel after 17a completes.

#### Steps 18-38: Simplify Each Table Wrapper (Parallel - 21 agents)
- **Scope:** wrapper
- **Depends on:** [17b, 17c]

**Table-to-step mapping:**

| Step | File | Table |
|------|------|-------|
| 18 | permission.go | permissions |
| 19 | role.go | roles |
| 20 | media_dimension.go | media_dimension |
| 21 | user.go | users |
| 22 | admin_route.go | admin_routes |
| 23 | route.go | routes |
| 24 | datatype.go | datatypes |
| 25 | field.go | fields |
| 26 | admin_datatype.go | admin_datatypes |
| 27 | admin_field.go | admin_fields |
| 28 | token.go | tokens |
| 29 | user_oauth.go | user_oauth |
| 30 | table.go | tables |
| 31 | media.go | media |
| 32 | session.go | sessions |
| 33 | content_data.go | content_data |
| 34 | content_field.go | content_fields |
| 35 | admin_content_data.go | admin_content_data |
| 36 | admin_content_field.go | admin_content_fields |
| 37 | datatype_field.go | datatypes_fields |
| 38 | admin_datatype_field.go | admin_datatypes_fields |

### Phase 4: Cleanup & Integration (Steps 39-45)

#### Step 39: Remove Dead Code
- **Scope:** cleanup
- **Depends on:** [18-38]
- **Tasks:**
  - Delete `internal/db/convert.go` - type conversion utilities
  - Delete `internal/db/json.go` - NullString/NullInt64 JSON wrappers
  - Gut `internal/db/utility.go` - remove null conversion helpers
  - Remove all Map* functions
  - Remove any History-related code (replaced by AuditLogger)

#### Step 40: Update API Handlers
- **Scope:** integration
- **Depends on:** [39]
- **Tasks:**
  - Update handlers to use `ParseDatatypeID(r.PathValue("id"))`
  - Update request structs to use typed IDs
  - Update response structs to use typed IDs
  - Validation now happens automatically via UnmarshalJSON
  - Add audit logging calls where entities are created/updated/deleted

#### Step 41: Update CLI Operations
- **Scope:** integration
- **Depends on:** [39]
- **Tasks:**
  - Update CLI input parsing to use `ParseXID` functions
  - Error messages now include specific type information
  - Add audit logging for CLI operations

#### Step 42: Run Test Suite
- **Scope:** testing
- **Depends on:** [40, 41]
- **Tasks:**
  - `make test`
  - Test JSON serialization/deserialization with custom types
  - Test validation (invalid IDs, emails, slugs, etc.)
  - Test Timestamp parsing (strict RFC3339 for input, legacy for DB reads)
  - Test change_events operations
  - Test all three database engines

#### Step 43: Add Type Validation Tests
- **Scope:** testing
- **Depends on:** [42]
- **Tasks:**
  - Unit tests for each custom type's Validate() method
  - Unit tests for Parse functions
  - Unit tests for edge cases (zero values, max values, invalid formats)
  - Unit tests for AuditLogger interface

#### Step 44: Update DbDriver Interface Comments
- **Scope:** docs
- **Depends on:** [43]
- **File:** `internal/db/db.go`
- **Tasks:**
  - Update interface to use custom types in signatures
  - Document type safety guarantees
  - Add AuditLogger to DbDriver interface (or separate interface)

#### Step 45: Final Documentation
- **Scope:** docs
- **Depends on:** [44]
- **Tasks:**
  - Update CLAUDE.md database section
  - Document custom type system (ULID-based IDs, HLC, change_events)
  - Document distributed system foundation (node_id, change_events, backups)
  - Document backup tables and restore coordination
  - Create API migration guide
  - Update this plan with final metrics

---

## HQ Project Configuration

```json
{
  "name": "sqlc-type-unification",
  "base_commit": "<current HEAD>",
  "steps": [
    {"step_num": 0, "branch": "baseline/validate-current", "scope": "validation", "depends_on": []},
    {"step_num": 1, "branch": "schema/change-events-table", "scope": "schema", "depends_on": [0]},
    {"step_num": 2, "branch": "schema/backup-tables", "scope": "schema", "depends_on": [0]},
    {"step_num": 3, "branch": "schema/fk-indexes", "scope": "schema", "depends_on": [0]},
    {"step_num": 4, "branch": "schema/check-constraints", "scope": "schema", "depends_on": [0]},
    {"step_num": 5, "branch": "foundation/feature-branch", "scope": "setup", "depends_on": [1, 2, 3, 4]},
    {"step_num": 6, "branch": "types/primary-ids-ulid", "scope": "types", "depends_on": [5]},
    {"step_num": 7, "branch": "types/nullable-ids", "scope": "types", "depends_on": [5]},
    {"step_num": 8, "branch": "types/timestamp-hlc", "scope": "types", "depends_on": [5]},
    {"step_num": 9, "branch": "types/enums-distributed", "scope": "types", "depends_on": [5]},
    {"step_num": 10, "branch": "types/validation", "scope": "types", "depends_on": [5]},
    {"step_num": 11, "branch": "types/change-events-hlc", "scope": "types", "depends_on": [5]},
    {"step_num": 12, "branch": "types/transaction", "scope": "types", "depends_on": [5]},
    {"step_num": 13, "branch": "config/sqlc-yml", "scope": "config", "depends_on": [6, 7, 8, 9, 10, 11]},
    {"step_num": 14, "branch": "validate/circular-imports", "scope": "validation", "depends_on": [13]},
    {"step_num": 15, "branch": "generate/sqlc-regen", "scope": "codegen", "depends_on": [14]},
    {"step_num": 16, "branch": "validate/type-check", "scope": "validation", "depends_on": [15]},
    {"step_num": 171, "branch": "simplify/struct-definitions", "scope": "wrapper", "depends_on": [16]},
    {"step_num": 172, "branch": "simplify/remove-json-structs", "scope": "wrapper", "depends_on": [171]},
    {"step_num": 173, "branch": "simplify/remove-formparams", "scope": "wrapper", "depends_on": [171]},
    {"step_num": 18, "branch": "simplify/permission", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 19, "branch": "simplify/role", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 20, "branch": "simplify/media-dimension", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 21, "branch": "simplify/user", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 22, "branch": "simplify/admin-route", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 23, "branch": "simplify/route", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 24, "branch": "simplify/datatype", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 25, "branch": "simplify/field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 26, "branch": "simplify/admin-datatype", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 27, "branch": "simplify/admin-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 28, "branch": "simplify/token", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 29, "branch": "simplify/user-oauth", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 30, "branch": "simplify/table", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 31, "branch": "simplify/media", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 32, "branch": "simplify/session", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 33, "branch": "simplify/content-data", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 34, "branch": "simplify/content-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 35, "branch": "simplify/admin-content-data", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 36, "branch": "simplify/admin-content-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 37, "branch": "simplify/datatype-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 38, "branch": "simplify/admin-datatype-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 39, "branch": "cleanup/dead-code", "scope": "cleanup", "depends_on": [18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38]},
    {"step_num": 40, "branch": "integration/api-handlers", "scope": "integration", "depends_on": [39]},
    {"step_num": 41, "branch": "integration/cli-ops", "scope": "integration", "depends_on": [39]},
    {"step_num": 42, "branch": "testing/test-suite", "scope": "testing", "depends_on": [40, 41]},
    {"step_num": 43, "branch": "testing/type-validation", "scope": "testing", "depends_on": [42]},
    {"step_num": 44, "branch": "docs/interface-docs", "scope": "docs", "depends_on": [43]},
    {"step_num": 45, "branch": "docs/final-docs", "scope": "docs", "depends_on": [44]}
  ]
}
```

**Note on step numbering:**
- Steps 1-4 are schema improvements (Phase 0.5)
- Steps 171, 172, 173 are "sub-steps" of conceptual step 17 (17a, 17b, 17c in the plan)
- Schema steps 3 and 4 can run in parallel after step 0

---

## Custom Types Summary

| Category | Types | Purpose |
|----------|-------|---------|
| **Primary IDs** | DatatypeID, UserID, RoleID, PermissionID, FieldID, ContentID, ContentFieldID, MediaID, MediaDimensionID, SessionID, TokenID, RouteID, AdminRouteID, TableID, UserOauthID, UserSshKeyID, AdminDatatypeID, AdminFieldID, AdminContentID, AdminContentFieldID, DatatypeFieldID, AdminDatatypeFieldID, **EventID, NodeID, BackupID, VerificationID, BackupSetID** | ULID-based (26-char string), compile-time type safety, globally unique |
| **Nullable IDs** | NullableXID for each FK relationship | Type-safe nullable foreign keys (ULID or empty string) |
| **Timestamp** | Timestamp, NullableTimestamp | Unified datetime across SQLite/MySQL/PostgreSQL (UTC only, strict RFC3339 input) |
| **HLC** | HLC (Hybrid Logical Clock) | int64, distributed event ordering, encodes wall time + counter |
| **Enums** | ContentStatus, FieldType, RouteType, **Operation, Action, ConflictPolicy** | Domain constraint validation (mirrors DB CHECK constraints) |
| **Backup Enums** | BackupType, BackupStatus, VerificationStatus, BackupSetStatus | Backup system domain constraints |
| **Validation** | Slug, Email, URL | Format validation at boundary |
| **Nullable Validation** | NullableSlug, NullableEmail, NullableURL | Optional validated fields |
| **Change Events** | ChangeEvent, ChangeEventLogger interface | Combined audit trail + replication log + webhook source |
| **Backup Types** | Backup, BackupVerification, BackupSet | Distributed backup tracking and restore coordination |
| **Transaction** | TxFunc, WithTransaction, WithTransactionResult | Consistent transaction handling |

---

## Expected Results

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| `internal/db/` lines | 16,795 | ~7,000-8,000 | **50-55% reduction** |
| Structs per table | 9 | 3 | **67% reduction** |
| Mapping functions per table | 16 | 0 | **100% reduction** |
| Type conversion utilities | ~30 funcs | 0 | **100% reduction** |
| Compile-time type safety | None | Full | **New capability** |
| Validation coverage | Manual | Automatic | **New capability** |
| Error specificity | Generic | Type-specific | **New capability** |
| Total HQ Steps | N/A | 45 | Including sub-steps |

**New Capabilities:**
- Compile-time prevention of ID mixups
- Automatic validation on JSON unmarshal
- Automatic validation on DB scan
- Type-specific error messages
- Consistent datetime handling across all DBs
- **Distribution-ready foundation:**
  - ULID primary keys (globally unique, sortable, no coordination)
  - Node identity tracking (node_id on all rows)
  - HLC timestamps for distributed event ordering
  - Change events table (audit + replication + webhooks)
  - Per-datatype conflict resolution policies

---

## Parallelization Summary

```
Timeline:

Step 0 ─┬─ 1 (change_events) ─┬─ 5 ─┬─ 6 ─┬──────────────────────────────────────────────────────────────┐
        ├─ 2 (backups) ───────┤     ├─ 7 ─┤                                                               │
        ├─ 3 (fk-indexes) ────┤     ├─ 8 ─┤                                                               │
        └─ 4 (constraints) ───┘     ├─ 9 ─┼─ 13 ─ 14 ─ 15 ─ 16 ─ 17a ─┬─ 17b ─┬─ Steps 18-38 (21) ─┬─ 39 ─┬─ 40 ─┬─ 42 ─ 43 ─ 44 ─ 45
                                    ├─ 10 ┤                            └─ 17c ─┘                    │      └─ 41 ─┘
                                    ├─ 11 ┤                                                         │
                                    └─ 12 ┘                                                         │

Phase 0 (Baseline):      1 step (must complete first)
Phase 0.5 (Schema):      4 steps (all parallel after Step 0 - change_events, backups, fk-indexes, constraints)
Phase 1 (Types):         7 steps (parallel after step 5)
Phase 2 (Config):        4 sequential steps (13, 14, 15, 16)
Phase 3 (Wrappers):      24 steps (17a → 17b||17c → 21 parallel)
Phase 4 (Cleanup):       7 steps (some parallel)
Total:                   49 HQ steps (including sub-steps and schema)
```

**Critical Path:** 0 → 1 → 2 → 5 → types → 13 → 14 → 15 → 16 → 17a → 17b/c → wrappers → 39 → integration → tests → docs

---

## Agent Instructions Template

### For schema agents (Steps 1-4):

```
You are modifying database schema files to add enterprise-grade improvements.

**Files to modify:**
- `sql/schema/<dir>/schema.sql` (SQLite)
- `sql/schema/<dir>/schema_mysql.sql`
- `sql/schema/<dir>/schema_psql.sql`
- `sql/schema/<dir>/queries*.sql` (if adding queries)

**For Step 1 (change_events table):**
- Create new directory `sql/schema/0_audit/`
- Add schema files per "Schema Improvements" section
- Add all indexes (record, time, user, action)
- Add SQLC queries: LogAudit, GetAuditHistory, GetAuditHistorySince, GetAuditByUser, GetRecentAudit

**For Step 2 (remove history columns):**
- Remove `history TEXT` from all entity table schemas
- Tables: datatypes, content_data, admin_content_data, users, media, fields, etc.
- DO NOT remove from change_events (that's the replacement)

**For Step 3 (FK indexes):**
- Add `CREATE INDEX idx_<table>_<column>` for every foreign key column
- See "Foreign Key Indexes" section for complete list

**For Step 4 (CHECK constraints):**
- Add CHECK constraints for: status, field_type, route_type
- Add SQLite triggers for date_modified ON UPDATE behavior
- PostgreSQL can use DOMAINs (optional)

**Verification:**
1. `make sqlc` succeeds for all three engines
2. `make check` compiles (may have errors from removed history field - expected)
3. Syntax correct for all three database dialects

**Commit message:** "feat(schema): <specific change>"
```

### For type creation agents (Steps 6-12):

```
You are creating custom types for the database layer.

**Your file:** internal/db/<types_file>.go

**Requirements for each type:**
1. Implement `Validate() error` - returns specific error with type name
2. Implement `String() string` - includes type name for debugging
3. Implement `Value() (driver.Value, error)` - for database writes
4. Implement `Scan(value any) error` - for database reads, calls Validate()
5. Implement `MarshalJSON() ([]byte, error)` - for API responses
6. Implement `UnmarshalJSON(data []byte) error` - for API requests, calls Validate()

**For ID types, also implement:**
- `ParseXID(s string) (XID, error)` - for parsing path params/query strings

**Error message format:**
- Always include type name: `fmt.Errorf("DatatypeID: must be positive, got %d", id)`

**Verification:**
1. `go build ./internal/db/...` passes
2. All methods implemented
3. Validation is called in Scan and UnmarshalJSON

**Commit message:** "feat(types): add <TypeName> with validation"
```

### For wrapper simplification agents (Steps 18-38):

```
You are simplifying the database wrapper for the <TABLE> table.

**Context:**
- Custom types are now used for all ID, timestamp, and validated fields
- All three sqlc packages use the SAME custom types
- No type conversion needed - direct field assignment

**Your file:** internal/db/<table>.go

**Transformation:**

1. DELETE these structs:
   - <Table>JSON
   - Create<Table>ParamsJSON
   - Update<Table>ParamsJSON
   - Create<Table>FormParams
   - Update<Table>FormParams

2. UPDATE the main struct to use custom types:
   - int64 (for IDs) → DatatypeID, UserID, etc.
   - sql.NullInt64 (for FK) → NullableDatatypeID, NullableUserID, etc.
   - sql.NullString (for dates) → Timestamp
   - sql.NullString (for history) → History
   - string (for slug) → Slug
   - string (for email) → Email

3. DELETE all Map* functions

4. SIMPLIFY query wrappers:
   ```go
   // BEFORE
   func (d MysqlDatabase) GetDatatype(id int64) (*Datatypes, error) {
       row, err := mdbm.New(d.Connection).GetDatatype(d.Context, int32(id))
       res := d.MapDatatype(row)
       return &res, nil
   }

   // AFTER
   func (d MysqlDatabase) GetDatatype(id DatatypeID) (*Datatypes, error) {
       row, err := mdbm.New(d.Connection).GetDatatype(d.Context, id)
       if err != nil {
           return nil, err
       }
       return &Datatypes{
           DatatypeID:   row.DatatypeID,   // Both are db.DatatypeID
           ParentID:     row.ParentID,      // Both are db.NullableDatatypeID
           AuthorID:     row.AuthorID,      // Both are db.UserID
           DateCreated:  row.DateCreated,   // Both are db.Timestamp
           History:      row.History,       // Both are db.History
       }, nil
   }
   ```

**Verification:**
1. `make check` passes
2. No Map* functions remain
3. No *JSON or *FormParams structs remain
4. All IDs use typed ID types
5. All timestamps use Timestamp type

**Commit message:** "refactor(<table>): use custom types, eliminate mapping"
```

---

## Verification Checklist

### Type System
- [ ] All ID types are ULID-based (26-char string), have Validate, String, Value, Scan, MarshalJSON, UnmarshalJSON
- [ ] All nullable ID types have corresponding methods
- [ ] Timestamp handles SQLite TEXT, MySQL DATETIME, PostgreSQL TIMESTAMP WITH TIME ZONE
- [ ] HLC type correctly encodes wall time + counter (int64)
- [ ] Enums validate against allowed values (including Operation, Action, ConflictPolicy, BackupType, etc.)
- [ ] Slug, Email, URL validate format
- [ ] ChangeEvent struct correctly handles all fields

### Code Generation
- [ ] `make sqlc` succeeds for all three engines
- [ ] Generated code uses custom types (check models.go in each package)
- [ ] No raw int64 for ID columns
- [ ] No sql.NullX types in generated code

### Wrapper Layer
- [ ] All `*JSON` structs removed
- [ ] All `*FormParams` structs removed
- [ ] All `Map*` functions removed
- [ ] Type conversion utilities removed
- [ ] Query wrappers use direct field assignment

### Integration
- [ ] `make check` passes
- [ ] `make test` passes
- [ ] API handlers use ParseXID functions
- [ ] CLI operations use ParseXID functions
- [ ] Error messages include type names

### Runtime
- [ ] JSON serialization works (custom types → JSON)
- [ ] JSON deserialization validates (JSON → custom types)
- [ ] Database reads validate (DB → custom types)
- [ ] All three database engines work
- [ ] Invalid input rejected with specific errors

### Distributed System Foundation
- [ ] All tables have ULID primary keys (CHAR(26)/TEXT)
- [ ] All entity tables have node_id column
- [ ] change_events table created with all indexes
- [ ] backups, backup_verifications, backup_sets tables created
- [ ] HLC generation is thread-safe and monotonic
- [ ] ULID generation is thread-safe with monotonic entropy
- [ ] datatypes table has conflict_policy column

---

## Enterprise Reliability Features

| Feature | Implementation |
|---------|----------------|
| **Compile-time safety** | ULID-based typed IDs prevent mixups |
| **Input validation** | UnmarshalJSON validates API input (strict RFC3339) |
| **DB validation** | Scan validates database reads |
| **Specific errors** | All errors include type name |
| **Consistent datetime** | Timestamp stored as UTC, strict RFC3339 input |
| **Enum constraints** | Go types + DB CHECK constraints (defense in depth) |
| **Format validation** | Slug, Email, URL validated (Go + DB constraints) |
| **Change events** | Centralized change_events table (audit + replication + webhooks) |
| **Transaction safety** | WithTransaction helper |
| **FK performance** | All foreign keys indexed |
| **Distribution-ready** | ULID PKs (no coordination), node_id tracking, HLC timestamps |
| **Backup tracking** | backups, backup_verifications, backup_sets tables |
| **Conflict resolution** | Per-datatype conflict_policy (lww/manual) |

---

## Remaining Valid Concerns

Issues noted by principal engineer review that require attention during implementation:

| Issue | Risk Level | Status | Recommendation |
|-------|------------|--------|----------------|
| Wildcard sqlc overrides (`*.column`) | Medium | **Monitor** | Test thoroughly; be explicit if wildcards cause issues |
| 21 parallel agents in Phase 3 | Medium | **Planned** | Serialize if merge conflicts; isolate to separate files |
| Performance benchmarks | Low | **Deferred** | Add after initial implementation is working |
| Enum sync (Go ↔ DB) | Medium | **Addressed** | DB CHECK constraints mirror Go validation |
| Timestamp format ambiguity | Medium | **Addressed** | Strict RFC3339 input, legacy only for DB reads |

**Action items deferred to Phase 4:**
- Add performance benchmarks comparing before/after
- Document enum sync process (if Go enum changes, update DB constraint)

---

## Rollback Plan

### If SQLC Generation Fails After Type Changes

1. **Preserve work in feature branch**
   ```bash
   git add -A && git commit -m "WIP: sqlc type changes - generation failing"
   ```

2. **Return to working state**
   ```bash
   git checkout main -- sql/sqlc.yml
   make sqlc  # Verify original config still works
   ```

3. **Diagnose the issue**
   - Check `sqlc generate` error output
   - Common issues:
     - Typo in column name (e.g., `datatypes.id` vs `datatypes.datatype_id`)
     - Missing import path
     - Type not implementing required interface (Scan, Value)
   - Fix one override at a time, regenerating between each

4. **Incremental approach**
   - Add overrides in batches of 5-10 columns
   - Run `make sqlc` after each batch
   - Commit working batches
   - This isolates which override causes failure

### If Application Fails at Runtime

1. **Check type interface implementations**
   ```go
   // Verify these compile - if they don't, the type is missing a method
   var _ sql.Scanner = (*DatatypeID)(nil)
   var _ driver.Valuer = DatatypeID(0)
   var _ json.Marshaler = DatatypeID(0)
   var _ json.Unmarshaler = (*DatatypeID)(nil)
   ```

2. **Test types in isolation**
   ```bash
   go test ./internal/db/... -run TestDatatypeID
   ```

3. **Rollback specific type**
   - Revert to `int64` for that column in sqlc.yml
   - Regenerate
   - File issue to track the problematic type

### Git Branch Strategy

```
main ─────────────────────────────────────────────────────► production
  │
  └── feature/sqlc-type-unification ─┬─ types-working ──────► (checkpoint)
                                     ├─ sqlc-generating ────► (checkpoint)
                                     └─ wrappers-simplified ► (checkpoint)
```

- Create checkpoint tags at each major phase completion
- If later phase fails, can reset to last checkpoint
- Never force-push to feature branch after sharing with team

---

## SQLite Query Strategy

**Current approach:** sqlc generates queries for all three databases including SQLite.

**Files involved:**
- `sql/schema/**/schema.sql` - SQLite schema
- `sql/schema/**/queries.sql` - SQLite queries
- `internal/db-sqlite/` - Generated SQLite code (package `mdb`)

**Why this is correct:**
- The `sql/sqlc.yml` configuration already handles SQLite as an engine
- SQLite queries use `?` parameter placeholders (handled by sqlc)
- SQLite returns `int64` for INTEGER columns (handled by custom types)
- The `Timestamp.Scan()` method handles SQLite's TEXT datetime format

**No changes needed** - SQLite is already fully integrated with sqlc.

---

## References

- [SQLC Type Overrides](../sqlc/OVERRIDE_REFERENCE.md)
- [SQLC Configuration](../sqlc/CONFIG_REFERENCE.md)
- [SQLC Transactions](../sqlc/TRANSACTION_REFERENCE.md)
- Current sqlc.yml: `sql/sqlc.yml`
- Example wrapper: `internal/db/datatype.go`
