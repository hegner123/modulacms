Deploy Sync Engine - Implementation Plan

 Context

 ModulaCMS lacks the ability to sync CMS schema and content between environments. Developers running local databases cannot easily pull content from dev/staging, and pushing schema changes from local to dev requires manual effort.
 Existing CMS solutions (Umbraco) fail silently with opaque dependency graph errors. This package replaces the current stubs in internal/deploy/ with a selective, configurable sync engine that provides clear error reporting and
 git-friendly snapshots.

 Architecture Overview

 The deploy engine is a table-level sync system. It exports rows from a configurable set of tables on one instance and imports them on another. Transport is REST API with Bearer token auth. Wire format is JSON (ZIP if > 1GB). V1 supports
 overwrite merge strategy only (nuke target, replace with source).

 Default Sync Set (14 tables)

 ┌──────┬────────────────────────────────────────────┐
 │ Tier │           Tables (insert order)            │
 ├──────┼────────────────────────────────────────────┤
 │ 1    │ datatypes, admin_datatypes                 │
 ├──────┼────────────────────────────────────────────┤
 │ 2    │ fields, admin_fields, routes, admin_routes │
 ├──────┼────────────────────────────────────────────┤
 │ 3    │ datatypes_fields, admin_datatypes_fields   │
 ├──────┼────────────────────────────────────────────┤
 │ 4    │ content_data, admin_content_data           │
 ├──────┼────────────────────────────────────────────┤
 │ 5    │ content_fields, admin_content_fields       │
 ├──────┼────────────────────────────────────────────┤
 │ 6    │ content_relations, admin_content_relations │
 └──────┴────────────────────────────────────────────┘

 Truncate order is reverse (tier 6 first). Junction tables (datatypes_fields) and relation tables (content_relations) are included by default because without them the schema associations and content references are broken.

 NOTE: These tier numbers are specific to the deploy sync set and do NOT correspond to the tier
 numbers in wipe.go's DropAllTables. The truncate order (tier 6 first) mirrors the reverse
 dependency logic in wipe.go, but with different tier labels because the sync set is a subset
 of all tables. When implementing, use the tier ordering defined here, not wipe.go's.

 Admin tables use distinct ID types (AdminContentID, AdminDatatypeID, etc.) but are handled
 by the same BulkInsert/TruncateTable pipeline — at the SQL level they are just table names
 and column values. Validation logic that touches typed IDs must be aware of both families.

 Opt-in tables: users, roles, permissions, role_permissions, media, media_dimensions, sessions, tokens, user_oauth, user_ssh_keys, field_types, admin_field_types, change_events, backups, tables.

 Files to Create/Modify

 New Files

 ┌─────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │            File             │                                                 Purpose                                                  │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/types.go    │ SyncPayload, SyncManifest, SyncConfig, TableSet, MergeStrategy, SyncResult, SyncError, DeployEnvironment │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/export.go   │ Export from DbDriver into SyncPayload using existing List* methods                                       │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/import.go   │ Apply SyncPayload to DbDriver (overwrite strategy with FK handling)                                      │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/validate.go │ Pre-sync payload validation, post-sync integrity checks, clear error reporting                           │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/snapshot.go │ Save/list/load/restore local snapshots (JSON files in deploy/snapshots/)                                 │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/client.go   │ HTTP client for remote ModulaCMS instances (modeled on sdks/go/httpclient.go)                            │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/server.go   │ Handler functions for deploy API endpoints                                                               │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/deploy_ops.go   │ DeployOps interface + 3 backend implementations (FK toggle, truncate, bulk insert)                       │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/deploy.go   │ API endpoint registration                                                                                │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ cmd/deploy.go               │ Cobra CLI commands                                                                                       │
 └─────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Modified Files

 ┌────────────────────────────────────────────────────┬────────────────────────────────────────────────────────────────────────────┐
 │                        File                        │                                   Change                                   │
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/deploy.go                          │ Replace stubs with orchestrator (Push, Pull, ExportToFile, ImportFromFile) │
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ internal/config/config.go                          │ Add Deploy_Environments, Deploy_Snapshot_Dir fields                        │
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/db.go                                  │ Add ListContentRelations, ListAdminContentRelations to DbDriver interface  │
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/content_relation*.go (3 wrappers)      │ Implement ListContentRelations on Database, MysqlDatabase, PsqlDatabase    │
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/admin_content_relation*.go (3 wrappers)│ Implement ListAdminContentRelations on Database, MysqlDatabase, PsqlDatabase│
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/mux.go                             │ Register deploy endpoints                                                  │
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ sql/schema/24_content_relations/queries*.sql       │ Add ListContentRelations query                                             │
 ├────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────┤
 │ sql/schema/25_admin_content_relations/queries*.sql │ Add ListAdminContentRelations query                                        │
 └────────────────────────────────────────────────────┴────────────────────────────────────────────────────────────────────────────┘

 Core Types (internal/deploy/types.go)

 type MergeStrategy string
 const StrategyOverwrite MergeStrategy = "overwrite"

 // TableData is the per-table wire format inside SyncPayload.
 // Columns defines the ordered column names. Rows are positional, matching Columns order.
 // Typed IDs serialize as their 26-char ULID string. Timestamps serialize as RFC3339 strings.
 // Nullable fields serialize as JSON null or their string value.
 type TableData struct {
     Columns []string `json:"columns"`
     Rows    [][]any  `json:"rows"`
 }

 type SyncPayload struct {
     Manifest SyncManifest          `json:"manifest"`
     Tables   map[string]TableData  `json:"tables"`    // keyed by DBTable string value
     UserRefs map[string]string     `json:"user_refs"` // user_id -> username
 }

 type SyncManifest struct {
     SchemaVersion string         `json:"schema_version"` // SHA256 of sorted table:columns
     Timestamp     string         `json:"timestamp"`      // RFC3339 UTC
     SourceNodeID  string         `json:"source_node_id"`
     SourceURL     string         `json:"source_url"`
     Version       string         `json:"version"`        // ModulaCMS version
     Strategy      MergeStrategy  `json:"strategy"`
     Tables        []string       `json:"tables"`         // table names included
     RowCounts     map[string]int `json:"row_counts"`     // table -> count
     PayloadHash   string         `json:"payload_hash"`   // SHA256 of Tables map JSON
 }

 type SyncResult struct {
     Success        bool           `json:"success"`
     DryRun         bool           `json:"dry_run"`
     Strategy       MergeStrategy  `json:"strategy"`
     TablesAffected []string       `json:"tables_affected"`
     RowCounts      map[string]int `json:"row_counts"`
     BackupPath     string         `json:"backup_path"`
     SnapshotID     string         `json:"snapshot_id"`
     Duration       string         `json:"duration"`
     Errors         []SyncError    `json:"errors,omitempty"`
     Warnings       []string       `json:"warnings,omitempty"`
 }

 type SyncError struct {
     Table   string `json:"table"`
     Phase   string `json:"phase"`   // "export", "validate", "truncate", "insert", "verify"
     Message string `json:"message"`
     RowID   string `json:"row_id,omitempty"`
 }

 type SyncConfig struct {
     Source     string        // "local" or environment name
     Target     string        // "local" or environment name
     Strategy   MergeStrategy
     Tables     []string      // empty = DefaultTableSet
     DryRun     bool
     SkipBackup bool
     Timeout    time.Duration // default 5 minutes
 }

 Export serialization: List* methods return typed Go structs. The export function converts each
 struct to a TableData by reflecting on json tags to get column names, and extracting values
 in matching order. Typed IDs → string (ULID), Timestamps → RFC3339 string, Nullable* → nil or
 string. Import reverses this: TableData.Rows values are passed directly to BulkInsert as []any,
 which database/sql handles via driver.Valuer for the target column types.

 Key Design Decisions

 DeployOps: Separate Interface (not on DbDriver)

 Deploy needs FK toggle, truncate, and bulk insert - operations that bypass auditing. Adding them to DbDriver (150+ methods) bloats the interface. Instead, a separate DeployOps interface in internal/db/deploy_ops.go with 3 implementations
  (sqlite, mysql, psql). Constructed via NewDeployOps(driver) which type-switches on the concrete driver and uses GetConnection() to get the raw *sql.DB.

 type DeployOps interface {
     // ImportAtomic runs fn inside a backend-appropriate atomic context.
     // Handles connection pinning, FK disabling, transaction, and cleanup.
     // If fn returns error, transaction rolls back. See deploy_ops.go for per-backend details.
     ImportAtomic(ctx context.Context, fn ImportFunc) error

     // TruncateTable removes all rows. Must be called inside ImportAtomic.
     // Validates table name against allTables whitelist.
     TruncateTable(ctx context.Context, ex Executor, table DBTable) error

     // BulkInsert writes rows using multi-row INSERT. Must be called inside ImportAtomic.
     // Batches dynamically: min(500, maxVars/len(columns)) for SQLite.
     BulkInsert(ctx context.Context, ex Executor, table DBTable, columns []string, rows [][]any) error

     // VerifyForeignKeys checks all FK constraints, returns violations.
     VerifyForeignKeys(ctx context.Context, ex Executor) ([]FKViolation, error)
 }

 type ImportFunc func(ctx context.Context, ex Executor) error

 // Constructed via NewDeployOps(driver DbDriver) (DeployOps, error).
 // Returns one of: sqliteDeployOps, mysqlDeployOps, psqlDeployOps (unexported;
 // callers only need the DeployOps interface).
 // Already implemented in internal/db/deploy_ops.go.

 Connection Isolation (CRITICAL):

 DeployOps must NOT use the shared *sql.DB connection pool for FK manipulation. Each backend
 implementation acquires a dedicated connection at construction time via db.Conn(ctx) and holds
 it for the entire import operation. This prevents FK-off state from leaking to other goroutines.

 - SQLite: PRAGMA foreign_keys is per-connection state. A pooled connection with FKs off will be
   reused by normal request handlers. Solution: db.Conn(ctx) → dedicated connection. Note that
   PRAGMA foreign_keys cannot be changed inside a transaction, so the sequence is:
   PRAGMA foreign_keys=OFF → BEGIN → truncate+insert → COMMIT → PRAGMA foreign_keys=ON → close conn.
 - MySQL: SET FOREIGN_KEY_CHECKS=0 is session-scoped. Same pool leak risk. Solution: db.Conn(ctx).
 - PostgreSQL: SET session_replication_role='replica' is session-scoped. Solution: db.Conn(ctx).
   PostgreSQL can wrap TRUNCATE in a transaction, so: BEGIN → SET → truncate+insert → COMMIT → close conn.

 NewDeployOps(driver) calls GetConnection() to get the *sql.DB. Connection pinning happens
 inside ImportAtomic — each call pins a connection for the duration of the callback, then
 releases it. All TruncateTable/BulkInsert/VerifyForeignKeys calls run on the Executor
 passed to the ImportFunc callback, which is the pinned transaction.

 Transaction Safety:

 - SQLite: Entire truncate+insert wrapped in a single transaction on the dedicated connection.
   If crash occurs mid-import, transaction is never committed → DB unchanged.
 - PostgreSQL: Same — TRUNCATE and INSERT are transactional. Full rollback on failure.
 - MySQL: TRUNCATE TABLE is DDL (implicit commit, non-transactional). Cannot wrap in transaction.
   The pre-import backup (step 3 of import flow) is the ONLY recovery mechanism for MySQL.
   This is documented in SyncResult.Warnings when running against MySQL.

 Table Name Whitelist:

 TruncateTable and BulkInsert validate the table name against a hardcoded allowlist before
 constructing any SQL. The allowlist is the union of DefaultTableSet + all opt-in table names
 defined as constants in types.go. Any table name not in the allowlist returns an error
 immediately — no SQL is constructed or executed. This prevents injection via malformed payloads.

 Backend-specific SQL:
 - FK off: SQLite PRAGMA foreign_keys=OFF, MySQL SET FOREIGN_KEY_CHECKS=0, PostgreSQL SET session_replication_role='replica'
 - Truncate: SQLite DELETE FROM, MySQL TRUNCATE TABLE, PostgreSQL TRUNCATE TABLE CASCADE
 - Bulk insert: Multi-row INSERT, batched dynamically. Batch size = min(500, 999/len(columns))
   for SQLite (variable limit = 999). content_data has 11 columns → 90 rows/batch.
   MySQL/PostgreSQL use higher limits (batch of 1000 rows).
   PostgreSQL uses $N placeholders, others use ?.

 VerifyForeignKeys — per-backend SQL:
 - SQLite: PRAGMA foreign_key_check — returns (table, rowid, referenced_table, fk_index) for
   every row with a dangling FK reference. Handles self-referential content_data correctly
   because it checks actual data state, not insertion order.
 - MySQL: Query information_schema.KEY_COLUMN_USAGE to get all FK definitions for sync-set
   tables, then for each FK run: SELECT t.pk FROM t LEFT JOIN ref ON t.fk_col = ref.pk
   WHERE ref.pk IS NULL. Report each orphaned row.
 - PostgreSQL: Temporarily SET session_replication_role='origin' (re-enables constraint checks),
   then for each sync-set table run: ALTER TABLE t VALIDATE CONSTRAINT <fk_name>. Catch any
   constraint violation errors and translate to FKViolation structs.

 User Reference Handling

 Users aren't in the default sync set, but most tables have author_id FK to users. Strategy:
 1. Export collects all referenced user IDs into UserRefs map (user_id -> username)
 2. Import checks each referenced user_id on target
 3. If user exists (same ULID), use as-is
 4. If not, create placeholder user with original ULID, [sync] username, viewer role, locked (no password hash)
 5. No ID remapping needed - ULIDs are globally unique

 Wire Format

 - JSON by default (Content-Type: application/json)
 - If serialized payload > 1GB, compress to ZIP (Content-Type: application/zip)
 - Receiver checks Content-Type header to decide decoding

 Snapshots

 Stored in deploy/snapshots/ (configurable via Deploy_Snapshot_Dir). Each snapshot is a
 SyncPayload JSON file. Git-committable for version control. Gzipped if > 1GB.

 Snapshot naming: snap_YYYYMMDD_HHMMSS.json (or .json.gz if compressed). Timestamp is UTC.
 Example: snap_20260223_143022.json. The snapshot list command sorts by filename, which sorts
 chronologically due to the format. Snapshot IDs used in CLI commands are the filename stem
 (e.g., "snap_20260223_143022").

 API Endpoints

 ┌────────┬───────────────────────┬───────────────┬────────────────────────────┐
 │ Method │         Path          │  Permission   │          Purpose           │
 ├────────┼───────────────────────┼───────────────┼────────────────────────────┤
 │ GET    │ /api/v1/deploy/health │ deploy:read   │ Health + version check     │
 ├────────┼───────────────────────┼───────────────┼────────────────────────────┤
 │ POST   │ /api/v1/deploy/export │ deploy:read   │ Export data as SyncPayload │
 ├────────┼───────────────────────┼───────────────┼────────────────────────────┤
 │ POST   │ /api/v1/deploy/import │ deploy:create │ Apply SyncPayload          │
 └────────┴───────────────────────┴───────────────┴────────────────────────────┘

 New permissions: deploy:read, deploy:create (admin role only).

 Authentication: Uses the existing middleware chain (internal/middleware/middleware.go).
 AuthRequest() tries cookie auth first, then falls back to APIKeyAuth() which reads
 Authorization: Bearer <token>. No new auth mechanism needed — the deploy client sends
 a standard Bearer token and the existing middleware handles it. The deploy:read and
 deploy:create permissions are checked by RequirePermission() after authentication succeeds.

 Export request body:
 {"tables": ["datatypes", "fields", ...]}

 Import request body: SyncPayload JSON (or ZIP).

 Import response:
 {
   "success": true,
   "tables_affected": [...],
   "row_counts": {"datatypes": 12},
   "errors": [],
   "warnings": ["Created 2 placeholder users"]
 }

 Error response:
 {
   "error": "insert failed",
   "details": [{"table": "content_data", "phase": "insert", "message": "...", "row_id": "01ABC..."}],
   "recovery": "Use snapshot snap_20260223_143022 to restore"
 }

 CLI Commands

 modulacms deploy pull <source>           Pull from remote, apply locally
 modulacms deploy push <target>           Export local, push to remote
 modulacms deploy export [--file path]    Export to local file
 modulacms deploy import <file>           Import from local file
 modulacms deploy snapshot list           List snapshots
 modulacms deploy snapshot restore <id>   Restore a snapshot
 modulacms deploy snapshot show <id>      Show snapshot metadata
 modulacms deploy env list                List configured environments
 modulacms deploy env test <name>         Test connectivity

 Flags: --tables, --strategy, --dry-run, --skip-backup, --json

 Dry-Run Mode:

 --dry-run performs all read-only steps without modifying the target database:
 1. Export the payload from source (or read from file)
 2. Run full pre-sync validation (schema version, FK consistency, ULID format, row counts)
 3. Connect to target and verify health + schema compatibility
 4. Report what WOULD happen: tables to be truncated, row counts to be inserted,
    user references that would need placeholder creation
 5. Return a DryRunResult with all findings

 Dry-run does NOT: truncate tables, insert rows, create placeholder users, modify any data.
 Output format matches SyncResult but with success=false and a "dry_run": true flag.

 Config Additions

 // In config.go Config struct
 Deploy_Environments []DeployEnvironmentConfig `json:"deploy_environments"`
 Deploy_Snapshot_Dir string                    `json:"deploy_snapshot_dir"`

 // In deploy/types.go or config package
 type DeployEnvironmentConfig struct {
     Name   string `json:"name"`
     URL    string `json:"url"`
     APIKey string `json:"api_key"` // supports ${VAR} expansion via existing config system
 }

 API keys MUST use the existing ${VAR} config expansion pattern. Example config:
 {"deploy_environments": [{"name": "prod", "url": "https://cms.example.com", "api_key": "${DEPLOY_PROD_KEY}"}]}
 The deploy package resolves these via the same expansion logic in internal/config/file_provider.go.
 Document that hardcoding API keys is unsupported — the system should warn at startup if a
 deploy_environments entry has an api_key value that does not use ${} expansion.

 Overwrite Import Flow

 1. Validate payload (hash, row counts, FK integrity within payload, schema version compatibility)
 2. Save pre-import snapshot (safety net)
 3. Create backup via backup.CreateFullBackup(cfg) — cfg is the current config.Config.
    Backup is stored per existing backup configuration (local or S3).
    The returned path is included in SyncResult.BackupPath for recovery reference.
 4. Acquire import lock — sync.Mutex.TryLock() (non-blocking). If locked, return 409 immediately.
 5. Create context with timeout (default 5 minutes, configurable via SyncConfig.Timeout)
 6. ops.ImportAtomic() wraps steps 7-12 — acquires dedicated connection, disables FKs,
    begins transaction (SQLite/PostgreSQL), executes fn, commits, re-enables FKs, releases conn.
    The entire sequence runs inside the ImportAtomic callback:
 7.   Truncate tables tier 6 → tier 1 (reverse dependency order)
 8.   Resolve user references (find or create placeholder users)
 9.   Bulk insert tier 1 → tier 6 (dependency order)
 10.  VerifyForeignKeys — report any violations
 11. (ImportAtomic handles commit + FK re-enable + connection release internally)
 12. Count rows per table, compare to manifest
 13. Record synthetic change_event for audit trail (see Audit Trail below)
 14. Release import lock
 15. Return SyncResult with row counts, warnings, errors

 NOTE: The import flow uses ImportAtomic (already implemented in deploy_ops.go), NOT the
 separate BeginImport/CommitImport/AbortImport methods listed in the old DeployOps interface.
 ImportAtomic handles the correct backend-specific sequencing internally:
 - SQLite: PRAGMA OFF → BEGIN → fn → COMMIT → PRAGMA ON → close conn
 - PostgreSQL: BEGIN → SET replica → fn → SET origin → COMMIT
 - MySQL: BEGIN → SET FK_CHECKS=0 → fn → SET FK_CHECKS=1 → COMMIT
 If fn returns an error, the transaction rolls back automatically.

 On error: ImportAtomic rolls back the transaction (SQLite/PostgreSQL restore to pre-import
 state). The pre-import backup (step 3) provides additional recovery.

 Import Timeout:

 The import operation uses a context with a configurable timeout (default: 5 minutes).
 If the context deadline expires, ImportAtomic returns ctx.Err(), the transaction rolls back,
 and the import lock is released. SyncResult reports the timeout as a SyncError.

 Import Concurrency Guard:

 A sync.Mutex at the package level prevents concurrent imports. Uses TryLock() for immediate
 failure — does not block. If a second import request arrives while one is in progress, it
 returns HTTP 409 Conflict: "Import already in progress."

 V1 limitation: the lock is process-scoped. If multiple ModulaCMS instances share the same
 database (horizontal scaling), concurrent imports from different processes are NOT prevented.
 This is acceptable for V1 because ModulaCMS is primarily a single-binary deployment and
 deploy import is a rare administrative operation. Document this limitation. V2 can add a
 deploy_locks table with a single row claimed via UPDATE ... WHERE locked_by IS NULL.

 Audit Trail:

 The import bypasses the audited command pattern (uses raw SQL via DeployOps). To prevent a
 gap in the audit log, record a single synthetic change_event after successful import:
 - Operation: "deploy_sync"
 - TableName: "_deploy_sync" (sentinel value, not a real table)
 - NewValues: JSON of the SyncManifest (tables, row counts, source URL, schema version hash)
 - OldValues: null (or the pre-import snapshot ID for reference)
 - UserID: the authenticated user who triggered the import
 This gives the audit log a clear marker: "everything changed by a deploy sync from source Z
 at time T." Without it, a site admin sees a mysterious gap where content changed but nothing
 was recorded.

 Validation (Clear Error Reporting)

 SyncManifest Schema Version:

 SyncManifest includes a SchemaVersion field — a hash of the column names for each table in the
 sync set. Computed as: for each table in sorted order, join column names with commas, then
 SHA256 the concatenated result. On import, the target computes its own schema hash and compares.
 If they differ, the import fails immediately with: "Schema mismatch: source and target have
 different table structures. Source schema version: abc123, target: def456. Run database
 migrations on the target before syncing." This catches column additions/removals/renames
 before they produce cryptic bulk insert failures.

 Pre-sync:
 - Schema version matches between payload manifest and target database
 - Manifest hash matches payload
 - Row counts match slice lengths
 - All ULIDs are valid format
 - FK references within payload are consistent (every content_data.datatype_id exists in payload datatypes)
 - Content tree pointers reference existing content_data IDs
 - All author_id values present in UserRefs map

 Post-sync:
 - Row counts match expected
 - FK integrity check via DeployOps.VerifyForeignKeys()
 - Sample verification (spot-check rows by ID)

 Every failure produces a SyncError{Table, Phase, Message, RowID} - the user always knows exactly what table, what row, and why.

 Implementation Order

 Phase 1: Foundation
   1a. sqlc queries for ListContentRelations/ListAdminContentRelations (6 query files across
       3 dialects each for content_relations and admin_content_relations) → just sqlc

       sqlc annotations:
         -- name: ListContentRelations :many
         SELECT * FROM content_relations ORDER BY date_created;

         -- name: ListAdminContentRelations :many
         SELECT * FROM admin_content_relations ORDER BY date_created;

   1b. Add to DbDriver interface + implement on all 3 wrapper structs (6 new wrapper methods):

       ListContentRelations() (*[]ContentRelations, error)
       ListAdminContentRelations() (*[]AdminContentRelations, error)

       These follow the same pattern as ListDatatypeField() and ListAdminDatatypeField().
   1c. DeployOps interface + 3 backend implementations (internal/db/deploy_ops.go)
       PRIORITY: Prototype DisableForeignKeys isolation on SQLite first using db.Conn(ctx).
       Verify with a test that FK-off does not leak to other pool connections.
   1d. Core types (internal/deploy/types.go)
   1e. DeployOps tests — table-driven tests for each backend:
       - FK toggle isolation (verify other connections still enforce FKs)
       - TruncateTable with valid table name succeeds
       - TruncateTable with invalid table name returns error (whitelist enforcement)
       - BulkInsert with dynamic batch sizing
       - VerifyForeignKeys detects deliberately broken FK references

 Phase 2: Local sync (no network)
   2a. Export (internal/deploy/export.go)
   2b. Validate (internal/deploy/validate.go) including schema version computation
   2c. Import with overwrite strategy (internal/deploy/import.go) including:
       - Transaction wrapping (SQLite/PostgreSQL)
       - MySQL non-transactional path with backup-only recovery
       - Import concurrency guard (sync.Mutex)
   2d. Export/Import roundtrip tests — written alongside the code, not deferred:
       - SQLite roundtrip: export known data → import to fresh DB → verify all rows match
       - Self-referential content_data tree survives roundtrip (parent/child/sibling pointers)
       - Placeholder user creation when author_id doesn't exist on target
       - Schema version mismatch produces clear error
       - Overwrite correctly nukes previous data
   2e. Snapshot save/list/load/restore (internal/deploy/snapshot.go)
   2f. Orchestrator - ExportToFile, ImportFromFile (internal/deploy/deploy.go)
   2g. CLI: export, import, snapshot commands (cmd/deploy.go)

 Phase 3: Config + API
   3a. Config additions (internal/config/config.go) — ensure ${VAR} expansion works for api_key
   3b. API handlers (internal/deploy/server.go, internal/router/deploy.go)
   3c. Permission bootstrap (deploy:read, deploy:create)
   3d. Register endpoints (internal/router/mux.go)

 Phase 4: Network
   4a. HTTP client (internal/deploy/client.go)
   4b. Orchestrator - Push, Pull (internal/deploy/deploy.go)
   4c. CLI: pull, push, env commands (cmd/deploy.go)

 Phase 5: Polish
   5a. ZIP/gzip compression for large payloads
   5b. Dry-run mode (read-only validation + impact report)

 Verification

 1. just sqlc - regenerate after adding content relation queries
 2. go build ./... - compile check after each phase
 3. just test - run full test suite
 4. Manual test: export → import roundtrip with SQLite dev database
 5. Manual test: snapshot save → snapshot restore preserves data
 6. Manual test: deploy pull from a running dev instance
