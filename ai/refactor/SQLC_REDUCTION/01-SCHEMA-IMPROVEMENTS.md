# Schema Improvements (Do Now While Cost Is Zero)

These schema changes should be implemented in Phase 0/1 before the type system work. Adding them now is free; adding them later with production data requires migrations.

---

## 1. ULID Primary Keys (Distribution-Ready)

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

---

## 2. Node Identity Column (Multi-Node Ready)

Add `node_id` to all entity tables for future distributed support:

```sql
-- Add to every entity table (nullable for now, required in multi-node mode)
node_id CHAR(26),  -- ULID of the node that created/owns this record

CREATE INDEX idx_<table>_node ON <table>(node_id);
```

---

## 3. Change Events Table (Audit + Replication + Webhooks)

Replaces `change_events`. Serves three purposes:
1. **Audit trail** - Who changed what, when
2. **Replication log** - What changes to sync to other nodes
3. **Webhook source** - What events to fire

### PostgreSQL

```sql
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
```

### MySQL

```sql
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
```

### SQLite

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

---

## 4. HLC Timestamp Column

Add Hybrid Logical Clock timestamp alongside `date_modified` for distributed ordering:

```sql
-- Add to every entity table
hlc_modified BIGINT NOT NULL DEFAULT 0,  -- Hybrid Logical Clock for ordering

-- HLC encodes: (wall_time_ms << 16) | logical_counter
-- Allows total ordering even with clock skew between nodes
```

---

## 5. Conflict Policy (Per-Datatype)

Add to datatypes table for configurable conflict resolution:

```sql
-- Add to datatypes and admin_datatypes tables
conflict_policy VARCHAR(20) NOT NULL DEFAULT 'lww'
    CHECK (conflict_policy IN ('lww', 'manual')),
    -- 'lww' = Last Write Wins (simple, possible data loss)
    -- 'manual' = Flag conflicts for human resolution
```

---

## 6. Foreign Key Indexes

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

---

## 7. Database Constraints (Defense in Depth)

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

---

## 8. Timestamp Standardization

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

---

## 9. Backup Tracking Tables (Distributed Restore)

Essential for distributed systems. Without this, you can't answer:
- "What's the latest backup for the London node?"
- "Can I restore NYC to match London's state from 6 hours ago?"
- "Did last night's backup job actually run?"
- "Which backup contains transaction X?"

### PostgreSQL

```sql
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
```

### MySQL

```sql
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
```

### SQLite

```sql
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

---

## Replication Position

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
