# Audit Package - Revised Design

**Revision Date**: 2026-01-23
**Approach**: Interface + Per-Entity Command Structs (Idiomatic Go)
**Addresses**: Critical issues from AUDIT_PACKAGE_REVIEW.md

---

## Design Goals

1. **Atomic transactions**: Mutation and audit event succeed or fail together
2. **Idiomatic Go**: Interface-based command pattern (matches stdlib conventions)
3. **Clean call sites**: Single command argument to audit functions
4. **Leverage typed fields**: JSON unmarshal validates via types.Email, types.Slug, etc.
5. **No changes to existing wrappers**: Audited operations are additive

---

## Package Structure

```
internal/db/
├── audited/
│   ├── audited.go       # Generic Create, Update, Delete functions
│   ├── interfaces.go    # Command interfaces
│   ├── context.go       # AuditContext type and Ctx constructor
│   └── change_event.go  # Transaction-aware change event recording
├── user.go              # NewUserCmd, UpdateUserCmd, DeleteUserCmd structs + factories
├── content_data.go      # Command structs for content_data
└── ...                  # Other entity files
```

---

## Core Types

### audited/interfaces.go

```go
package audited

import (
    "context"
    "database/sql"
)

// DBTX matches sqlc's generated interface - works with both *sql.DB and *sql.Tx
type DBTX interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
}

// CreateCommand is the interface for audited create operations.
// Implementations bundle context, audit context, and params into a single struct.
type CreateCommand[T any] interface {
    Context() context.Context
    AuditContext() AuditContext
    Connection() *sql.DB
    TableName() string
    Execute(context.Context, DBTX) (T, error)
    GetID(T) string
    Params() any // For JSON serialization to audit log
}

// UpdateCommand is the interface for audited update operations.
type UpdateCommand[T any] interface {
    Context() context.Context
    AuditContext() AuditContext
    Connection() *sql.DB
    TableName() string
    GetBefore(context.Context, DBTX) (T, error)
    Execute(context.Context, DBTX) error
    GetID() string
    Params() any
}

// DeleteCommand is the interface for audited delete operations.
type DeleteCommand[T any] interface {
    Context() context.Context
    AuditContext() AuditContext
    Connection() *sql.DB
    TableName() string
    GetBefore(context.Context, DBTX) (T, error)
    Execute(context.Context, DBTX) error
    GetID() string
}
```

### audited/context.go

```go
package audited

import "github.com/hegner123/modulacms/internal/db/types"

// AuditContext carries metadata for audit records.
type AuditContext struct {
    NodeID    types.NodeID
    UserID    types.UserID
    RequestID string // For distributed tracing
    IP        string // Client IP for security audits
}

// Ctx is a brief constructor for AuditContext.
// Usage: auditCtx := audited.Ctx(nodeID, userID, requestID, ip)
func Ctx(nodeID types.NodeID, userID types.UserID, requestID, ip string) AuditContext {
    return AuditContext{
        NodeID:    nodeID,
        UserID:    userID,
        RequestID: requestID,
        IP:        ip,
    }
}
```

### audited/change_event.go

```go
package audited

import (
    "context"
    "database/sql"

    "github.com/hegner123/modulacms/internal/db/types"
)

// RecordChangeEventParams contains all fields for a change event.
type RecordChangeEventParams struct {
    EventID      types.EventID
    HlcTimestamp types.HLC
    NodeID       types.NodeID
    TableName    string
    RecordID     string
    Operation    types.Operation
    UserID       types.NullableUserID
    OldValues    types.JSONData
    NewValues    types.JSONData
    RequestID    string
    IP           string
}

// recordChangeEventTx records a change event within an existing transaction.
// Uses raw SQL to ensure it participates in the same transaction as the mutation.
func recordChangeEventTx(ctx context.Context, tx *sql.Tx, p RecordChangeEventParams) error {
    query := `
        INSERT INTO change_events (
            event_id, hlc_timestamp, node_id, table_name, record_id,
            operation, user_id, old_values, new_values, request_id, ip,
            wall_timestamp
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
    `

    var userID any
    if p.UserID.Valid {
        userID = string(p.UserID.ID)
    }

    _, err := tx.ExecContext(ctx, query,
        string(p.EventID),
        p.HlcTimestamp,
        string(p.NodeID),
        p.TableName,
        p.RecordID,
        string(p.Operation),
        userID,
        []byte(p.OldValues),
        []byte(p.NewValues),
        p.RequestID,
        p.IP,
    )
    return err
}

// Note: For MySQL, use NOW() instead of datetime('now')
// Note: For PostgreSQL, use NOW() and $1, $2, etc. placeholders
// See "Driver-Specific Change Event Recording" section for implementation options.
```

### audited/audited.go

```go
package audited

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/hegner123/modulacms/internal/db/types"
)

// Create executes an audited create operation.
// The mutation and audit record are atomic - both succeed or both fail.
func Create[T any](cmd CreateCommand[T]) (T, error) {
    var result T
    ctx := cmd.Context()

    // Add timeout if context has none
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }

    err := types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
        // 1. Execute the create within transaction
        created, err := cmd.Execute(ctx, tx)
        if err != nil {
            return fmt.Errorf("execute: %w", err)
        }
        result = created

        // 2. Serialize for audit
        newValues, err := json.Marshal(created)
        if err != nil {
            return fmt.Errorf("marshal: %w", err)
        }

        // 3. Record change event within same transaction
        auditCtx := cmd.AuditContext()
        return recordChangeEventTx(ctx, tx, RecordChangeEventParams{
            EventID:      types.NewEventID(),
            HlcTimestamp: types.HLCNow(),
            NodeID:       auditCtx.NodeID,
            TableName:    cmd.TableName(),
            RecordID:     cmd.GetID(created),
            Operation:    types.OperationINSERT,
            UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
            NewValues:    types.JSONData(newValues),
            RequestID:    auditCtx.RequestID,
            IP:           auditCtx.IP,
        })
    })

    return result, err
}

// Update executes an audited update operation.
// Captures before-state, executes update, records both states atomically.
func Update[T any](cmd UpdateCommand[T]) error {
    ctx := cmd.Context()

    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }

    return types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
        // 1. Get before state within transaction
        // Note: For strict consistency, use SELECT ... FOR UPDATE in GetBefore
        // (MySQL/PostgreSQL only - SQLite uses database-level locking)
        before, err := cmd.GetBefore(ctx, tx)
        if err != nil {
            return fmt.Errorf("get before: %w", err)
        }
        oldValues, _ := json.Marshal(before)

        // 2. Execute update within transaction
        if err := cmd.Execute(ctx, tx); err != nil {
            return fmt.Errorf("execute: %w", err)
        }

        // 3. Serialize update params (captures "what changed", not full new state)
        newValues, _ := json.Marshal(cmd.Params())

        // 4. Record change event
        auditCtx := cmd.AuditContext()
        return recordChangeEventTx(ctx, tx, RecordChangeEventParams{
            EventID:      types.NewEventID(),
            HlcTimestamp: types.HLCNow(),
            NodeID:       auditCtx.NodeID,
            TableName:    cmd.TableName(),
            RecordID:     cmd.GetID(),
            Operation:    types.OperationUPDATE,
            UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
            OldValues:    types.JSONData(oldValues),
            NewValues:    types.JSONData(newValues),
            RequestID:    auditCtx.RequestID,
            IP:           auditCtx.IP,
        })
    })
}

// Delete executes an audited delete operation.
// Captures before-state, executes delete, records deletion atomically.
func Delete[T any](cmd DeleteCommand[T]) error {
    ctx := cmd.Context()

    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }

    return types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
        // 1. Get before state
        before, err := cmd.GetBefore(ctx, tx)
        if err != nil {
            return fmt.Errorf("get before: %w", err)
        }
        oldValues, _ := json.Marshal(before)

        // 2. Execute delete
        if err := cmd.Execute(ctx, tx); err != nil {
            return fmt.Errorf("execute: %w", err)
        }

        // 3. Record change event
        auditCtx := cmd.AuditContext()
        return recordChangeEventTx(ctx, tx, RecordChangeEventParams{
            EventID:      types.NewEventID(),
            HlcTimestamp: types.HLCNow(),
            NodeID:       auditCtx.NodeID,
            TableName:    cmd.TableName(),
            RecordID:     cmd.GetID(),
            Operation:    types.OperationDELETE,
            UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
            OldValues:    types.JSONData(oldValues),
            RequestID:    auditCtx.RequestID,
            IP:           auditCtx.IP,
        })
    })
}
```

---

## Entity Command Structs

Each entity defines command structs that implement the interfaces. This is where the per-entity logic lives.

### internal/db/user.go (command structs and factories)

```go
// ========== AUDITED COMMAND TYPES ==========

// ----- CREATE -----

// NewUserCmd implements audited.CreateCommand[mdb.Users].
type NewUserCmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    params   CreateUserParams
    conn     *sql.DB
}

func (c NewUserCmd) Context() context.Context           { return c.ctx }
func (c NewUserCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c NewUserCmd) Connection() *sql.DB                { return c.conn }
func (c NewUserCmd) TableName() string                  { return "users" }
func (c NewUserCmd) Params() any                        { return c.params }
func (c NewUserCmd) GetID(u mdb.Users) string           { return u.UserID }

func (c NewUserCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
    queries := mdb.New(tx)
    return queries.CreateUser(ctx, mdb.CreateUserParams{
        UserID:       string(c.params.UserID),
        Username:     c.params.Username,
        Email:        string(c.params.Email),
        PasswordHash: c.params.PasswordHash,
        RoleID:       string(c.params.RoleID),
        DateCreated:  c.params.DateCreated.String(),
        DateModified: c.params.DateModified.String(),
    })
}

// NewUserCmd factory - bundles ctx, auditCtx, params into single command.
func (d Database) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
    return NewUserCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// ----- UPDATE -----

// UpdateUserCmd implements audited.UpdateCommand[mdb.Users].
type UpdateUserCmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    params   UpdateUserParams
    conn     *sql.DB
}

func (c UpdateUserCmd) Context() context.Context           { return c.ctx }
func (c UpdateUserCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateUserCmd) Connection() *sql.DB                { return c.conn }
func (c UpdateUserCmd) TableName() string                  { return "users" }
func (c UpdateUserCmd) Params() any                        { return c.params }
func (c UpdateUserCmd) GetID() string                      { return string(c.params.UserID) }

func (c UpdateUserCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
    queries := mdb.New(tx)
    return queries.GetUser(ctx, string(c.params.UserID))
}

func (c UpdateUserCmd) Execute(ctx context.Context, tx audited.DBTX) error {
    queries := mdb.New(tx)
    return queries.UpdateUser(ctx, mdb.UpdateUserParams{
        UserID:       string(c.params.UserID),
        Username:     c.params.Username,
        Email:        string(c.params.Email),
        RoleID:       string(c.params.RoleID),
        DateModified: c.params.DateModified.String(),
    })
}

// UpdateUserCmd factory.
func (d Database) UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserParams) UpdateUserCmd {
    return UpdateUserCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// ----- DELETE -----

// DeleteUserCmd implements audited.DeleteCommand[mdb.Users].
type DeleteUserCmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    id       types.UserID
    conn     *sql.DB
}

func (c DeleteUserCmd) Context() context.Context           { return c.ctx }
func (c DeleteUserCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c DeleteUserCmd) Connection() *sql.DB                { return c.conn }
func (c DeleteUserCmd) TableName() string                  { return "users" }
func (c DeleteUserCmd) GetID() string                      { return string(c.id) }

func (c DeleteUserCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
    queries := mdb.New(tx)
    return queries.GetUser(ctx, string(c.id))
}

func (c DeleteUserCmd) Execute(ctx context.Context, tx audited.DBTX) error {
    queries := mdb.New(tx)
    return queries.DeleteUser(ctx, string(c.id))
}

// DeleteUserCmd factory.
func (d Database) DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmd {
    return DeleteUserCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
```

### MySQL variant (internal/db/user.go - MysqlDatabase)

```go
// NewUserCmd for MySQL - same pattern, different sqlc package.
type NewUserCmdMysql struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    params   CreateUserParams
    conn     *sql.DB
}

func (c NewUserCmdMysql) Context() context.Context           { return c.ctx }
func (c NewUserCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }
func (c NewUserCmdMysql) Connection() *sql.DB                { return c.conn }
func (c NewUserCmdMysql) TableName() string                  { return "users" }
func (c NewUserCmdMysql) Params() any                        { return c.params }
func (c NewUserCmdMysql) GetID(u mdbm.Users) string          { return u.UserID }

func (c NewUserCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Users, error) {
    queries := mdbm.New(tx)
    return queries.CreateUser(ctx, mdbm.CreateUserParams{
        UserID:       string(c.params.UserID),
        Username:     c.params.Username,
        Email:        string(c.params.Email),
        PasswordHash: c.params.PasswordHash,
        RoleID:       string(c.params.RoleID),
        DateCreated:  c.params.DateCreated.Time(),  // MySQL uses time.Time
        DateModified: c.params.DateModified.Time(),
    })
}

func (d MysqlDatabase) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmdMysql {
    return NewUserCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// Similar patterns for UpdateUserCmdMysql, DeleteUserCmdMysql...
```

---

## Params Structs with Typed Fields

The params structs use typed fields that validate during JSON unmarshal:

```go
// internal/db/user.go

// CreateUserParams - JSON tags enable direct unmarshal from HTTP request body.
// Typed fields (types.Email, types.UserID, etc.) validate during UnmarshalJSON.
type CreateUserParams struct {
    UserID       types.UserID    `json:"user_id"`
    Username     string          `json:"username"`
    Email        types.Email     `json:"email"`         // Validates email format
    PasswordHash string          `json:"-"`             // Never from JSON
    RoleID       types.RoleID    `json:"role_id"`
    DateCreated  types.Timestamp `json:"date_created"`  // Validates RFC3339
    DateModified types.Timestamp `json:"date_modified"`
}

// UpdateUserParams - partial update, only changed fields serialized to audit log.
type UpdateUserParams struct {
    UserID       types.UserID    `json:"user_id"`
    Username     string          `json:"username,omitempty"`
    Email        types.Email     `json:"email,omitempty"`
    RoleID       types.RoleID    `json:"role_id,omitempty"`
    DateModified types.Timestamp `json:"date_modified"`
}
```

---

## Usage in HTTP Handlers

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 1. JSON -> typed params (validation automatic via types.Email.UnmarshalJSON etc.)
    var params db.CreateUserParams
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 2. Set server-generated fields
    params.UserID = types.NewUserID()
    params.PasswordHash = hashPassword(r.FormValue("password"))
    params.DateCreated = types.TimestampNow()
    params.DateModified = types.TimestampNow()

    // 3. Brief audit context constructor
    auditCtx := audited.Ctx(h.nodeID, getUserID(r), getRequestID(r), getIP(r))

    // 4. Create command - bundles ctx, auditCtx, params
    cmd := h.db.NewUserCmd(r.Context(), auditCtx, params)

    // 5. Execute - single argument
    user, err := audited.Create(cmd)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 6. Map sqlc type to API response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(mapToAPIUser(user))
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    var params db.UpdateUserParams
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    params.DateModified = types.TimestampNow()
    auditCtx := audited.Ctx(h.nodeID, getUserID(r), getRequestID(r), getIP(r))
    cmd := h.db.UpdateUserCmd(r.Context(), auditCtx, params)

    if err := audited.Update(cmd); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    id := types.UserID(chi.URLParam(r, "id"))
    if err := id.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    auditCtx := audited.Ctx(h.nodeID, getUserID(r), getRequestID(r), getIP(r))
    cmd := h.db.DeleteUserCmd(r.Context(), auditCtx, id)

    if err := audited.Delete(cmd); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

---

## Driver-Specific Change Event Recording

The `recordChangeEventTx` function needs driver-specific SQL. Recommended approach:

### Use sqlc-generated code

Add to each driver's queries.sql:

```sql
-- sqlite/queries.sql
-- name: RecordChangeEventTx :exec
INSERT INTO change_events (
    event_id, hlc_timestamp, node_id, table_name, record_id,
    operation, user_id, old_values, new_values, request_id, ip,
    wall_timestamp
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'));

-- mysql/queries.sql
-- name: RecordChangeEventTx :exec
INSERT INTO change_events (
    event_id, hlc_timestamp, node_id, table_name, record_id,
    operation, user_id, old_values, new_values, request_id, ip,
    wall_timestamp
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW());

-- postgres/queries.sql
-- name: RecordChangeEventTx :exec
INSERT INTO change_events (
    event_id, hlc_timestamp, node_id, table_name, record_id,
    operation, user_id, old_values, new_values, request_id, ip,
    wall_timestamp
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW());
```

Then create driver-specific recorder functions in the audited package or pass the recorder as part of the command.

---

## Code Statistics

| Metric | Count |
|--------|-------|
| Generic interfaces | 3 (CreateCommand, UpdateCommand, DeleteCommand) |
| Command structs per entity | 3 (New, Update, Delete) |
| Methods per command struct | 6-7 |
| Lines per entity (all 3 commands) | ~75 |
| Total for 23 entities × 3 drivers | ~5,175 lines |
| Audited package | ~200 lines |

**Trade-off**: More definition code, but call sites are trivial and the pattern is immediately recognizable to any Go developer.

---

## Testing Strategy

### Unit Tests

```go
func TestAuditedCreate_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    auditCtx := audited.Ctx(types.NewNodeID(), types.NewUserID(), "req-123", "127.0.0.1")
    params := db.CreateUserParams{
        UserID:       types.NewUserID(),
        Username:     "testuser",
        Email:        types.Email("test@example.com"),
        PasswordHash: "hashed",
        RoleID:       types.NewRoleID(),
        DateCreated:  types.TimestampNow(),
        DateModified: types.TimestampNow(),
    }

    cmd := db.NewUserCmd(context.Background(), auditCtx, params)
    user, err := audited.Create(cmd)

    require.NoError(t, err)
    require.Equal(t, "testuser", user.Username)

    // Verify change event was recorded
    events, _ := db.ListChangeEventsByTable("users")
    require.Len(t, events, 1)
    require.Equal(t, types.OperationINSERT, events[0].Operation)
}

func TestAuditedCreate_Rollback(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    auditCtx := audited.Ctx(types.NewNodeID(), types.NewUserID(), "", "")
    params := db.CreateUserParams{
        UserID: "", // Invalid - triggers error
    }

    cmd := db.NewUserCmd(context.Background(), auditCtx, params)
    _, err := audited.Create(cmd)

    require.Error(t, err)

    // Verify nothing was created
    userCount, _ := db.CountUsers()
    eventCount, _ := db.CountChangeEvents()
    require.Equal(t, int64(0), *userCount)
    require.Equal(t, int64(0), *eventCount)
}

func TestAuditedUpdate_CapturesBeforeState(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Create user first
    auditCtx := audited.Ctx(types.NewNodeID(), types.NewUserID(), "", "")
    createParams := db.CreateUserParams{
        UserID:   types.NewUserID(),
        Username: "original",
        Email:    types.Email("original@example.com"),
        // ...
    }
    cmd := db.NewUserCmd(context.Background(), auditCtx, createParams)
    user, _ := audited.Create(cmd)

    // Update user
    updateParams := db.UpdateUserParams{
        UserID:       user.UserID,
        Username:     "updated",
        DateModified: types.TimestampNow(),
    }
    updateCmd := db.UpdateUserCmd(context.Background(), auditCtx, updateParams)
    err := audited.Update(updateCmd)

    require.NoError(t, err)

    // Verify change event captures before state
    events, _ := db.ListChangeEventsByTable("users")
    require.Len(t, events, 2) // CREATE + UPDATE

    updateEvent := events[1]
    require.Equal(t, types.OperationUPDATE, updateEvent.Operation)

    var oldValues map[string]any
    json.Unmarshal([]byte(updateEvent.OldValues), &oldValues)
    require.Equal(t, "original", oldValues["username"])
}
```

### Integration Tests (All Drivers)

```go
func TestAuditedOperations_AllDrivers(t *testing.T) {
    drivers := []struct {
        name  string
        setup func(t *testing.T) db.DbDriver
    }{
        {"sqlite", setupSQLiteDB},
        {"mysql", setupMySQLDB},
        {"postgres", setupPostgresDB},
    }

    for _, d := range drivers {
        t.Run(d.name, func(t *testing.T) {
            db := d.setup(t)
            defer cleanup(db)

            t.Run("Create", func(t *testing.T) { testAuditedCreate(t, db) })
            t.Run("Update", func(t *testing.T) { testAuditedUpdate(t, db) })
            t.Run("Delete", func(t *testing.T) { testAuditedDelete(t, db) })
            t.Run("Rollback", func(t *testing.T) { testAuditedRollback(t, db) })
        })
    }
}
```

---

## Implementation Order

1. **Create audited package**
   - `internal/db/audited/interfaces.go`
   - `internal/db/audited/context.go`
   - `internal/db/audited/change_event.go`
   - `internal/db/audited/audited.go`

2. **Add RecordChangeEventTx to sqlc**
   - Add query to all three driver query files
   - Run `make sqlc`

3. **Implement Users entity commands**
   - Add NewUserCmd, UpdateUserCmd, DeleteUserCmd to `internal/db/user.go`
   - Add MySQL variants to MysqlDatabase methods
   - Add PostgreSQL variants to PsqlDatabase methods
   - Write tests

4. **Verify atomicity**
   - Test rollback scenarios
   - Test context cancellation
   - Test concurrent access (if applicable)

5. **Roll out to remaining entities**
   - Use Users as template
   - Prioritize high-traffic entities

---

## Open Questions

1. **Sensitive field stripping**: Should command structs implement a `Redact()` method that returns a copy with sensitive fields zeroed before JSON serialization?

2. **Batch operations**: Should we support batch create/update/delete with a single transaction and one change event per record?

3. **Soft deletes**: Entities with soft delete (status change) - use `audited.Update` with a `SoftDeleteUserCmd`, or create `audited.SoftDelete`?

4. **SELECT FOR UPDATE**: Add `GetBeforeForUpdate` variant for MySQL/PostgreSQL that uses row locking? SQLite can ignore (database-level locking).
