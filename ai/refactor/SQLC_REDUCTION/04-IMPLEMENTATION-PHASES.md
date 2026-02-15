# Implementation Phases

**Last Updated:** 2026-01-23

---

## Status Summary

| Phase | Steps | Status |
|-------|-------|--------|
| Phase 0: Baseline | 0 | Complete |
| Phase 0.5: Schema | 0a-0e | Complete |
| Phase 1: Types | 5-12 | Complete |
| Phase 2: Config | 13-16 | Complete |
| Phase 3: Wrapper | 17a-38 | Complete |
| Phase 4: Cleanup | 39-45 | Complete |

**Progress:** 100% complete - Build passes with zero errors (2026-01-23)

---

## Phase 0: Baseline Validation (Step 0) - COMPLETE

### Step 0: Validate Current SQLC Works ✅
- **Scope:** validation
- **Status:** Complete - `e9da301`
- **Tasks:**
  - Run `just sqlc` on current configuration
  - Document any existing generation errors
  - Run `make check` to verify current compile status
  - Document baseline line counts:
    ```bash
    wc -l internal/db/*.go internal/db-sqlite/*.go internal/db-mysql/*.go internal/db-psql/*.go
    ```
  - Create baseline commit marker
- **Artifact:** Baseline metrics documented in this plan (update "Before" column in Expected Results)

---

## Phase 0.5: Schema Improvements (Steps 0a-0e) - COMPLETE

These schema changes are free now but expensive with production data. Complete before type system work.

### Step 0a: Create change_events Table ✅
- **Scope:** schema
- **Status:** Complete - `567755d`
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
  - Run `just sqlc` to verify generation
- **Commit message:** "feat(schema): add change_events table for audit, replication, and webhooks"

### Step 0b: Define Schema Without history Columns ✅
- **Scope:** schema
- **Status:** Complete - `6263a5b`
- **Depends on:** [0a]
- **Files:** All schema files in `sql/schema/*/`
- **Tasks:**
  - Do NOT include `history TEXT` column in entity table schemas
  - Tables affected: datatypes, content_data, admin_content_data, users, media, etc.
  - History/audit tracking now via change_events table
  - No migration needed (no production data)
  - Run `just sqlc` to regenerate code
  - Run `make check` to find compilation errors from missing field
- **Note:** Wrapper layer will need updates to use change_events instead of history (handled in Phase 3)
- **Commit message:** "refactor(schema): define schema without history TEXT columns (use change_events)"

### Step 0c: Add Foreign Key Indexes ✅
- **Scope:** schema
- **Status:** Complete - `aec5cf0`
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
  - Run `just sqlc` to verify
- **Commit message:** "perf(schema): add FK indexes for query optimization"

### Step 0d: Add Database Constraints ✅
- **Scope:** schema
- **Status:** Complete - `f8d4524`
- **Depends on:** [0]
- **Files:** All schema files
- **Tasks:**
  - Add CHECK constraints for:
    - `status` column: `CHECK (status IN ('draft', 'published', 'archived', 'pending'))`
    - `field_type` column: `CHECK (field_type IN (...))`
    - `route_type` column: `CHECK (route_type IN (...))`
  - Add PostgreSQL DOMAINs for slug, email (optional - CHECK constraints work too)
  - Add SQLite triggers for `date_modified` ON UPDATE behavior
  - Run `just sqlc` to verify
- **Commit message:** "feat(schema): add CHECK constraints matching Go validation types"

### Step 0e: Create Backup Tables ✅
- **Scope:** schema
- **Status:** Complete - `df8956d`
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
  - Run `just sqlc` to verify generation
- **Commit message:** "feat(schema): add backup tables for distributed restore coordination"

---

## Phase 1: Custom Type System (Steps 5-12) - COMPLETE

### Step 5: Create Feature Branch ✅
- **Scope:** setup
- **Status:** Complete - branch `feature/sqlc-type-unification` exists
- **Depends on:** [1, 2, 3, 4, 0e]
- **Tasks:** Create `feature/sqlc-type-unification` branch from main (includes schema changes)

### Step 6: Create Primary ID Types (ULID-based) ✅
- **Scope:** types
- **Status:** Complete - `e663efe`
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

### Step 7: Create Nullable ID Types ✅
- **Scope:** types
- **Status:** Complete - `e663efe`
- **Depends on:** [5]
- **File:** `internal/db/types_nullable_ids.go`
- **Tasks:** Implement nullable variants for all FK relationships
- **Types to create:** NullableXID for each ID type that appears as a foreign key

### Step 8: Create Timestamp Type ✅
- **Scope:** types
- **Status:** Complete - `e663efe`
- **Depends on:** [5]
- **File:** `internal/db/types_timestamp.go`
- **Tasks:** Implement unified Timestamp type (UTC only, strict RFC3339 input)

### Step 9: Create Enum Types ✅
- **Scope:** types
- **Status:** Complete - `e663efe`
- **Depends on:** [5]
- **File:** `internal/db/types_enums.go`
- **Tasks:** Implement ContentStatus, FieldType, RouteType with validation
- **Distributed System Enums:**
  - **Operation**: INSERT, UPDATE, DELETE (for change_events)
  - **Action**: semantic action names (create_datatype, update_content, publish, etc.)
  - **ConflictPolicy**: lww, manual (per-datatype conflict resolution)

### Step 10: Create Validation Types ✅
- **Scope:** types
- **Status:** Complete - `e663efe`
- **Depends on:** [5]
- **File:** `internal/db/types_validation.go`
- **Tasks:** Implement Slug, Email, URL, NullableSlug, NullableEmail, NullableURL

### Step 11: Create Change Event and HLC Types ✅
- **Scope:** types
- **Status:** Complete - `e663efe`
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

### Step 12: Create Transaction Helper ✅
- **Scope:** types
- **Status:** Complete - `e663efe`
- **Depends on:** [5]
- **File:** `internal/db/transaction.go`
- **Tasks:** Implement WithTransaction and WithTransactionResult

---

## Phase 2: Configuration & Generation (Steps 13-16) - IN PROGRESS

### Step 13: Update sqlc.yml 🔄
- **Scope:** config
- **Status:** In Progress - uncommitted changes in `sql/sqlc.yml`
- **Depends on:** [6, 7, 8, 9, 10, 11]
- **File:** `sql/sqlc.yml`
- **Tasks:** Replace with new configuration above

### Step 14: Check for Circular Imports ⬜
- **Scope:** validation
- **Status:** Not Started
- **Depends on:** [13]
- **Tasks:**
  - Verify `internal/db/` types can import `database/sql/driver` and `encoding/json`
  - Verify `internal/db/` does NOT import `internal/db-sqlite`, `internal/db-mysql`, or `internal/db-psql`
  - If circular import detected, move types to `internal/db/types/` subpackage
  - Run `go build ./internal/db/...` to verify
- **Contingency:** If circular imports exist, create `internal/db/types/` package and update sqlc.yml import paths

### Step 15: Regenerate All SQLC Code 🔄
- **Scope:** codegen
- **Status:** In Progress - uncommitted regenerated files
- **Depends on:** [14]
- **Tasks:**
  - Run `just sqlc`
  - Fix any generation errors (may need to adjust column names in config)
  - Verify all three packages generate: `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/`
  - Commit generated code

### Step 16: Validate Type Usage ⬜
- **Scope:** validation
- **Status:** Not Started
- **Depends on:** [15]
- **Tasks:**
  - Create validation script to verify custom types are used
  - Check that no raw int64 appears for ID columns
  - Check that Timestamp is used for date columns
  - Check that history columns are gone (change_events used instead)
  - Document any columns that need manual override

---

## Phase 3: Simplify Wrapper Layer (Steps 17a-38) - NOT STARTED

### Step 17a: Update Struct Definitions to Use Custom Types ⬜
- **Scope:** wrapper
- **Status:** Not Started
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

### Step 17b: Remove *JSON Struct Variants ⬜
- **Scope:** wrapper
- **Status:** Not Started
- **Depends on:** [17a]
- **File:** `internal/db/` - all table files
- **Tasks:**
  - Delete all `*JSON` struct definitions (e.g., `DatatypesJSON`, `UsersJSON`)
  - Delete all functions returning `*JSON` variants
  - Update any code referencing `*JSON` types
  - Run `make check` to verify compile

### Step 17c: Remove *FormParams Struct Variants ⬜
- **Scope:** wrapper
- **Status:** Not Started
- **Depends on:** [17a]
- **File:** `internal/db/` - all table files
- **Tasks:**
  - Delete all `*FormParams` struct definitions (e.g., `CreateDatatypeFormParams`)
  - Delete all functions accepting `*FormParams` types
  - Update any code referencing `*FormParams` types
  - Run `make check` to verify compile

**Note:** Steps 17b and 17c can run in parallel after 17a completes.

### Steps 18-38: Simplify Each Table Wrapper (Parallel - 21 agents) ⬜
- **Scope:** wrapper
- **Status:** Not Started
- **Depends on:** [17b, 17c]

**Table-to-step mapping:**

| Step | File | Table | Status |
|------|------|-------|--------|
| 18 | permission.go | permissions | ⬜ |
| 19 | role.go | roles | ⬜ |
| 20 | media_dimension.go | media_dimension | ⬜ |
| 21 | user.go | users | ⬜ |
| 22 | admin_route.go | admin_routes | ⬜ |
| 23 | route.go | routes | ⬜ |
| 24 | datatype.go | datatypes | ⬜ |
| 25 | field.go | fields | ⬜ |
| 26 | admin_datatype.go | admin_datatypes | ⬜ |
| 27 | admin_field.go | admin_fields | ⬜ |
| 28 | token.go | tokens | ⬜ |
| 29 | user_oauth.go | user_oauth | ⬜ |
| 30 | table.go | tables | ⬜ |
| 31 | media.go | media | ⬜ |
| 32 | session.go | sessions | ⬜ |
| 33 | content_data.go | content_data | ⬜ |
| 34 | content_field.go | content_fields | ⬜ |
| 35 | admin_content_data.go | admin_content_data | ⬜ |
| 36 | admin_content_field.go | admin_content_fields | ⬜ |
| 37 | datatype_field.go | datatypes_fields | ⬜ |
| 38 | admin_datatype_field.go | admin_datatypes_fields | ⬜ |

---

## Phase 4: Cleanup & Integration (Steps 39-45) - NOT STARTED

### Step 39: Remove Dead Code ⬜
- **Scope:** cleanup
- **Status:** Not Started
- **Depends on:** [18-38]
- **Tasks:**
  - Delete `internal/db/convert.go` - type conversion utilities
  - Delete `internal/db/json.go` - NullString/NullInt64 JSON wrappers
  - Gut `internal/db/utility.go` - remove null conversion helpers
  - Remove all Map* functions
  - Remove any History-related code (replaced by AuditLogger)

### Step 40: Update API Handlers ⬜
- **Scope:** integration
- **Status:** Not Started
- **Depends on:** [39]
- **Tasks:**
  - Update handlers to use `ParseDatatypeID(r.PathValue("id"))`
  - Update request structs to use typed IDs
  - Update response structs to use typed IDs
  - Validation now happens automatically via UnmarshalJSON
  - Add audit logging calls where entities are created/updated/deleted

### Step 41: Update CLI Operations ⬜
- **Scope:** integration
- **Status:** Not Started
- **Depends on:** [39]
- **Tasks:**
  - Update CLI input parsing to use `ParseXID` functions
  - Error messages now include specific type information
  - Add audit logging for CLI operations

### Step 42: Run Test Suite ⬜
- **Scope:** testing
- **Status:** Not Started
- **Depends on:** [40, 41]
- **Tasks:**
  - `just test`
  - Test JSON serialization/deserialization with custom types
  - Test validation (invalid IDs, emails, slugs, etc.)
  - Test Timestamp parsing (strict RFC3339 for input, legacy for DB reads)
  - Test change_events operations
  - Test all three database engines

### Step 43: Add Type Validation Tests ⬜
- **Scope:** testing
- **Status:** Not Started
- **Depends on:** [42]
- **Tasks:**
  - Unit tests for each custom type's Validate() method
  - Unit tests for Parse functions
  - Unit tests for edge cases (zero values, max values, invalid formats)
  - Unit tests for AuditLogger interface

### Step 44: Update DbDriver Interface Comments ⬜
- **Scope:** docs
- **Status:** Not Started
- **Depends on:** [43]
- **File:** `internal/db/db.go`
- **Tasks:**
  - Update interface to use custom types in signatures
  - Document type safety guarantees
  - Add AuditLogger to DbDriver interface (or separate interface)

### Step 45: Final Documentation ⬜
- **Scope:** docs
- **Status:** Not Started
- **Depends on:** [44]
- **Tasks:**
  - Update CLAUDE.md database section
  - Document custom type system (ULID-based IDs, HLC, change_events)
  - Document distributed system foundation (node_id, change_events, backups)
  - Document backup tables and restore coordination
  - Create API migration guide
  - Update this plan with final metrics
