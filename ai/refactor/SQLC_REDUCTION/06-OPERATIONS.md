# Operations Guide

This document contains agent instructions, verification checklists, and rollback procedures.

---

## Agent Instructions Templates

### For Schema Agents (Steps 1-4)

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
1. `just sqlc` succeeds for all three engines
2. `make check` compiles (may have errors from removed history field - expected)
3. Syntax correct for all three database dialects

**Commit message:** "feat(schema): <specific change>"
```

---

### For Type Creation Agents (Steps 6-12)

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

---

### For Wrapper Simplification Agents (Steps 18-38)

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
- [ ] `just sqlc` succeeds for all three engines
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
- [ ] `just test` passes
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

## Rollback Plan

### If SQLC Generation Fails After Type Changes

1. **Preserve work in feature branch**
   ```bash
   git add -A && git commit -m "WIP: sqlc type changes - generation failing"
   ```

2. **Return to working state**
   ```bash
   git checkout main -- sql/sqlc.yml
   just sqlc  # Verify original config still works
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
   - Run `just sqlc` after each batch
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
