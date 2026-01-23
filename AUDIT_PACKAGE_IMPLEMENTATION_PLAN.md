# Audit Package Implementation Plan

**Created**: 2026-01-23
**Status**: Planning
**Location**: `internal/db/audited/`

---

## Overview

Add transactional audit logging to all database mutations. Every Create, Update, and Delete operation will atomically record a change event in the same transaction, ensuring no mutation occurs without an audit trail.

### Goals

1. Atomic audit logging - mutation and audit event succeed or fail together
2. Before/after state capture for updates and deletes
3. Clean API using Go generics and command pattern
4. Leverage existing `types.*` JSON validation
5. No changes to existing `DbDriver` interface methods

### Non-Goals

- Retroactive auditing of existing data
- Real-time audit event streaming (future consideration)
- Audit event compaction or archival (separate concern)

---

## Architecture

### Package Structure

```
internal/db/
├── audited/
│   └── audited.go          # Generic Create/Update/Delete + interfaces
├── types/                   # (existing) - validates on JSON unmarshal
├── user.go                  # + NewUserCmd, UpdateUserCmd, DeleteUserCmd
├── route.go                 # + command structs for Route
├── backup.go                # + command structs for Backup
└── ...                      # same pattern for all entities
```

### Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│ HTTP Request                                                         │
│   ↓                                                                  │
│ json.Decode(&params)  ← types.Email, types.Slug validate here       │
│   ↓                                                                  │
│ db.NewUser(ctx, auditCtx, params)  ← creates command struct         │
│   ↓                                                                  │
│ audited.Create(cmd)                                                  │
│   ├─────────────────────────────────────────────────────────────┐   │
│   │ BEGIN TRANSACTION                                            │   │
│   │   1. Execute mutation (CreateUser)                           │   │
│   │   2. Serialize result to JSON                                │   │
│   │   3. Record change event                                     │   │
│   │ COMMIT (or ROLLBACK on any error)                            │   │
│   └─────────────────────────────────────────────────────────────┘   │
│   ↓                                                                  │
│ Return result + nil error (or rollback + error)                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Details

### Phase 1: Core Audited Package

**File**: `internal/db/audited/audited.go`

```go
package audited

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"

    "github.com/hegner123/modulacms/internal/db"
    "github.com/hegner123/modulacms/internal/db/types"
)

// AuditContext carries audit metadata through the request lifecycle
type AuditContext struct {
    NodeID types.NodeID
    UserID types.UserID
}

// ============================================================================
// COMMAND INTERFACES
// ============================================================================

// CreateCommand encapsulates everything needed for an audited create operation
type CreateCommand[T any] interface {
    TableName() string
    Execute() (T, error)
    GetID(T) string
    Params() any
    Driver() db.DbDriver
    Context() context.Context
    AuditContext() AuditContext
}

// UpdateCommand encapsulates everything needed for an audited update operation
type UpdateCommand[T any, ID any] interface {
    TableName() string
    RecordID() ID
    GetBefore() (*T, error)
    Execute() error
    Params() any
    Driver() db.DbDriver
    Context() context.Context
    AuditContext() AuditContext
}

// DeleteCommand encapsulates everything needed for an audited delete operation
type DeleteCommand[T any, ID any] interface {
    TableName() string
    RecordID() ID
    GetBefore() (*T, error)
    Execute() error
    Driver() db.DbDriver
    Context() context.Context
    AuditContext() AuditContext
}

// ============================================================================
// AUDITED OPERATIONS
// ============================================================================

// Create executes a create operation with atomic audit logging
func Create[T any](cmd CreateCommand[T]) (T, error) {
    var result T
    d := cmd.Driver()
    ctx := cmd.Context()
    auditCtx := cmd.AuditContext()

    conn, _, err := d.GetConnection()
    if err != nil {
        return result, fmt.Errorf("audited.Create: get connection: %w", err)
    }

    err = withTx(ctx, conn, func(tx *sql.Tx) error {
        // 1. Execute the create operation
        created, err := cmd.Execute()
        if err != nil {
            return fmt.Errorf("execute: %w", err)
        }
        result = created

        // 2. Serialize new state for audit record
        newValues, err := json.Marshal(created)
        if err != nil {
            return fmt.Errorf("marshal new values: %w", err)
        }

        // 3. Record the change event
        _, err = d.RecordChangeEvent(db.RecordChangeEventParams{
            EventID:      types.NewEventID(),
            HlcTimestamp: types.HLCNow(),
            NodeID:       auditCtx.NodeID,
            TableName:    cmd.TableName(),
            RecordID:     cmd.GetID(created),
            Operation:    types.OperationINSERT,
            UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
            NewValues:    types.JSONData(newValues),
        })
        if err != nil {
            return fmt.Errorf("record change event: %w", err)
        }

        return nil
    })

    if err != nil {
        return result, fmt.Errorf("audited.Create[%s]: %w", cmd.TableName(), err)
    }
    return result, nil
}

// Update executes an update operation with atomic audit logging
// Captures before-state for the audit record
func Update[T any, ID any](cmd UpdateCommand[T, ID]) error {
    d := cmd.Driver()
    ctx := cmd.Context()
    auditCtx := cmd.AuditContext()

    conn, _, err := d.GetConnection()
    if err != nil {
        return fmt.Errorf("audited.Update: get connection: %w", err)
    }

    return withTx(ctx, conn, func(tx *sql.Tx) error {
        // 1. Capture before state
        before, err := cmd.GetBefore()
        if err != nil {
            return fmt.Errorf("get before state: %w", err)
        }
        oldValues, err := json.Marshal(before)
        if err != nil {
            return fmt.Errorf("marshal old values: %w", err)
        }

        // 2. Execute the update operation
        if err := cmd.Execute(); err != nil {
            return fmt.Errorf("execute: %w", err)
        }

        // 3. Serialize new params for audit record
        newValues, err := json.Marshal(cmd.Params())
        if err != nil {
            return fmt.Errorf("marshal new values: %w", err)
        }

        // 4. Record the change event
        _, err = d.RecordChangeEvent(db.RecordChangeEventParams{
            EventID:      types.NewEventID(),
            HlcTimestamp: types.HLCNow(),
            NodeID:       auditCtx.NodeID,
            TableName:    cmd.TableName(),
            RecordID:     fmt.Sprint(cmd.RecordID()),
            Operation:    types.OperationUPDATE,
            UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
            OldValues:    types.JSONData(oldValues),
            NewValues:    types.JSONData(newValues),
        })
        if err != nil {
            return fmt.Errorf("record change event: %w", err)
        }

        return nil
    })
}

// Delete executes a delete operation with atomic audit logging
// Captures before-state for the audit record
func Delete[T any, ID any](cmd DeleteCommand[T, ID]) error {
    d := cmd.Driver()
    ctx := cmd.Context()
    auditCtx := cmd.AuditContext()

    conn, _, err := d.GetConnection()
    if err != nil {
        return fmt.Errorf("audited.Delete: get connection: %w", err)
    }

    return withTx(ctx, conn, func(tx *sql.Tx) error {
        // 1. Capture before state (what we're deleting)
        before, err := cmd.GetBefore()
        if err != nil {
            return fmt.Errorf("get before state: %w", err)
        }
        oldValues, err := json.Marshal(before)
        if err != nil {
            return fmt.Errorf("marshal old values: %w", err)
        }

        // 2. Execute the delete operation
        if err := cmd.Execute(); err != nil {
            return fmt.Errorf("execute: %w", err)
        }

        // 3. Record the change event
        _, err = d.RecordChangeEvent(db.RecordChangeEventParams{
            EventID:      types.NewEventID(),
            HlcTimestamp: types.HLCNow(),
            NodeID:       auditCtx.NodeID,
            TableName:    cmd.TableName(),
            RecordID:     fmt.Sprint(cmd.RecordID()),
            Operation:    types.OperationDELETE,
            UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
            OldValues:    types.JSONData(oldValues),
        })
        if err != nil {
            return fmt.Errorf("record change event: %w", err)
        }

        return nil
    })
}

// ============================================================================
// TRANSACTION HELPER
// ============================================================================

func withTx(ctx context.Context, conn *sql.DB, fn func(*sql.Tx) error) error {
    tx, err := conn.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }

    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
        }
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}
```

### Phase 2: Entity Command Structs

Add command structs to each entity file. Example for `user.go`:

**File**: `internal/db/user.go` (additions)

```go
// ============================================================================
// COMMAND STRUCTS
// ============================================================================

// NewUserCmd encapsulates a create user operation
type NewUserCmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    db       DbDriver
    params   CreateUserParams
}

// NewUser creates a command for creating a user with audit logging
func (d Database) NewUser(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
    return NewUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, params: params}
}

func (d MysqlDatabase) NewUser(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
    return NewUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, params: params}
}

func (d PsqlDatabase) NewUser(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
    return NewUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, params: params}
}

func (c NewUserCmd) TableName() string                  { return "users" }
func (c NewUserCmd) Driver() DbDriver                   { return c.db }
func (c NewUserCmd) Context() context.Context           { return c.ctx }
func (c NewUserCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c NewUserCmd) Params() any                        { return c.params }
func (c NewUserCmd) GetID(u Users) string               { return string(u.UserID) }
func (c NewUserCmd) Execute() (Users, error)            { return c.db.CreateUser(c.params) }

// UpdateUserCmd encapsulates an update user operation
type UpdateUserCmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    db       DbDriver
    id       types.UserID
    params   UpdateUserParams
}

// UpdateUserAudited creates a command for updating a user with audit logging
func (d Database) UpdateUserAudited(ctx context.Context, auditCtx audited.AuditContext, id types.UserID, params UpdateUserParams) UpdateUserCmd {
    return UpdateUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id, params: params}
}

func (d MysqlDatabase) UpdateUserAudited(ctx context.Context, auditCtx audited.AuditContext, id types.UserID, params UpdateUserParams) UpdateUserCmd {
    return UpdateUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id, params: params}
}

func (d PsqlDatabase) UpdateUserAudited(ctx context.Context, auditCtx audited.AuditContext, id types.UserID, params UpdateUserParams) UpdateUserCmd {
    return UpdateUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id, params: params}
}

func (c UpdateUserCmd) TableName() string                  { return "users" }
func (c UpdateUserCmd) Driver() DbDriver                   { return c.db }
func (c UpdateUserCmd) Context() context.Context           { return c.ctx }
func (c UpdateUserCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateUserCmd) RecordID() types.UserID             { return c.id }
func (c UpdateUserCmd) Params() any                        { return c.params }
func (c UpdateUserCmd) GetBefore() (*Users, error)         { return c.db.GetUser(c.id) }
func (c UpdateUserCmd) Execute() error                     { _, err := c.db.UpdateUser(c.params); return err }

// DeleteUserCmd encapsulates a delete user operation
type DeleteUserCmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    db       DbDriver
    id       types.UserID
}

// DeleteUserAudited creates a command for deleting a user with audit logging
func (d Database) DeleteUserAudited(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmd {
    return DeleteUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id}
}

func (d MysqlDatabase) DeleteUserAudited(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmd {
    return DeleteUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id}
}

func (d PsqlDatabase) DeleteUserAudited(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmd {
    return DeleteUserCmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id}
}

func (c DeleteUserCmd) TableName() string                  { return "users" }
func (c DeleteUserCmd) Driver() DbDriver                   { return c.db }
func (c DeleteUserCmd) Context() context.Context           { return c.ctx }
func (c DeleteUserCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c DeleteUserCmd) RecordID() types.UserID             { return c.id }
func (c DeleteUserCmd) GetBefore() (*Users, error)         { return c.db.GetUser(c.id) }
func (c DeleteUserCmd) Execute() error                     { return c.db.DeleteUser(c.id) }
```

### Phase 3: Entity Command Template

Use this template for each entity file. Replace placeholders:
- `{Entity}` → e.g., `Route`, `Backup`, `ContentData`
- `{entity}` → e.g., `route`, `backup`, `content_data`
- `{EntityID}` → e.g., `types.RouteID`, `types.BackupID`, `types.ContentID`

```go
// ============================================================================
// COMMAND STRUCTS - {Entity}
// ============================================================================

type New{Entity}Cmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    db       DbDriver
    params   Create{Entity}Params
}

func (d Database) New{Entity}(ctx context.Context, auditCtx audited.AuditContext, params Create{Entity}Params) New{Entity}Cmd {
    return New{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, params: params}
}

func (d MysqlDatabase) New{Entity}(ctx context.Context, auditCtx audited.AuditContext, params Create{Entity}Params) New{Entity}Cmd {
    return New{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, params: params}
}

func (d PsqlDatabase) New{Entity}(ctx context.Context, auditCtx audited.AuditContext, params Create{Entity}Params) New{Entity}Cmd {
    return New{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, params: params}
}

func (c New{Entity}Cmd) TableName() string                  { return "{entity}" }
func (c New{Entity}Cmd) Driver() DbDriver                   { return c.db }
func (c New{Entity}Cmd) Context() context.Context           { return c.ctx }
func (c New{Entity}Cmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c New{Entity}Cmd) Params() any                        { return c.params }
func (c New{Entity}Cmd) GetID(e {Entity}) string            { return string(e.{Entity}ID) }
func (c New{Entity}Cmd) Execute() ({Entity}, error)         { return c.db.Create{Entity}(c.params) }

type Update{Entity}Cmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    db       DbDriver
    id       {EntityID}
    params   Update{Entity}Params
}

func (d Database) Update{Entity}Audited(ctx context.Context, auditCtx audited.AuditContext, id {EntityID}, params Update{Entity}Params) Update{Entity}Cmd {
    return Update{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id, params: params}
}

func (d MysqlDatabase) Update{Entity}Audited(ctx context.Context, auditCtx audited.AuditContext, id {EntityID}, params Update{Entity}Params) Update{Entity}Cmd {
    return Update{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id, params: params}
}

func (d PsqlDatabase) Update{Entity}Audited(ctx context.Context, auditCtx audited.AuditContext, id {EntityID}, params Update{Entity}Params) Update{Entity}Cmd {
    return Update{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id, params: params}
}

func (c Update{Entity}Cmd) TableName() string                  { return "{entity}" }
func (c Update{Entity}Cmd) Driver() DbDriver                   { return c.db }
func (c Update{Entity}Cmd) Context() context.Context           { return c.ctx }
func (c Update{Entity}Cmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c Update{Entity}Cmd) RecordID() {EntityID}               { return c.id }
func (c Update{Entity}Cmd) Params() any                        { return c.params }
func (c Update{Entity}Cmd) GetBefore() (*{Entity}, error)      { return c.db.Get{Entity}(c.id) }
func (c Update{Entity}Cmd) Execute() error                     { _, err := c.db.Update{Entity}(c.params); return err }

type Delete{Entity}Cmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    db       DbDriver
    id       {EntityID}
}

func (d Database) Delete{Entity}Audited(ctx context.Context, auditCtx audited.AuditContext, id {EntityID}) Delete{Entity}Cmd {
    return Delete{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id}
}

func (d MysqlDatabase) Delete{Entity}Audited(ctx context.Context, auditCtx audited.AuditContext, id {EntityID}) Delete{Entity}Cmd {
    return Delete{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id}
}

func (d PsqlDatabase) Delete{Entity}Audited(ctx context.Context, auditCtx audited.AuditContext, id {EntityID}) Delete{Entity}Cmd {
    return Delete{Entity}Cmd{ctx: ctx, auditCtx: auditCtx, db: d, id: id}
}

func (c Delete{Entity}Cmd) TableName() string                  { return "{entity}" }
func (c Delete{Entity}Cmd) Driver() DbDriver                   { return c.db }
func (c Delete{Entity}Cmd) Context() context.Context           { return c.ctx }
func (c Delete{Entity}Cmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c Delete{Entity}Cmd) RecordID() {EntityID}               { return c.id }
func (c Delete{Entity}Cmd) GetBefore() (*{Entity}, error)      { return c.db.Get{Entity}(c.id) }
func (c Delete{Entity}Cmd) Execute() error                     { return c.db.Delete{Entity}(c.id) }
```

---

## Entities Requiring Command Structs

| Entity | File | ID Type | Table Name |
|--------|------|---------|------------|
| Users | user.go | types.UserID | users |
| Routes | route.go | types.RouteID | routes |
| Datatypes | datatype.go | types.DatatypeID | datatypes |
| DatatypeFields | datatype_field.go | int64 | datatype_fields |
| Fields | field.go | types.FieldID | fields |
| ContentData | content_data.go | types.ContentID | content_data |
| ContentFields | content_field.go | types.ContentFieldID | content_fields |
| Media | media.go | types.MediaID | media |
| MediaDimensions | media_dimension.go | int64 | media_dimensions |
| Permissions | permission.go | types.PermissionID | permissions |
| Roles | role.go | types.RoleID | roles |
| Sessions | session.go | types.SessionID | sessions |
| Tables | table.go | int64 | tables |
| Tokens | token.go | int64 | tokens |
| UserOauth | user_oauth.go | types.UserOauthID | user_oauth |
| UserSshKeys | user_ssh_keys.go | int64 | user_ssh_keys |
| AdminRoutes | admin_route.go | types.AdminRouteID | admin_routes |
| AdminDatatypes | admin_datatype.go | types.AdminDatatypeID | admin_datatypes |
| AdminDatatypeFields | admin_datatype_field.go | int64 | admin_datatype_fields |
| AdminFields | admin_field.go | types.AdminFieldID | admin_fields |
| AdminContentData | admin_content_data.go | types.AdminContentID | admin_content_data |
| AdminContentFields | admin_content_field.go | types.AdminContentFieldID | admin_content_fields |
| Backups | backup.go | types.BackupID | backups |
| BackupSets | backup.go | types.BackupSetID | backup_sets |
| BackupVerifications | backup.go | types.VerificationID | backup_verifications |

**Total**: 25 entities × 3 commands = 75 command structs

---

## Usage Examples

### HTTP Handler - Create

```go
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
    // 1. Parse + validate (types.Email, types.Timestamp validate on unmarshal)
    var params db.CreateUserParams
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    // 2. Build audit context from request
    auditCtx := audited.AuditContext{
        NodeID: config.NodeID,
        UserID: middleware.GetUserID(r.Context()),
    }

    // 3. Create command and execute
    newUser := db.NewUser(r.Context(), auditCtx, params)
    user, err := audited.Create(newUser)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // 4. Return result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

### HTTP Handler - Update

```go
func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
    userID := types.UserID(chi.URLParam(r, "id"))

    var params db.UpdateUserParams
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    auditCtx := audited.AuditContext{
        NodeID: config.NodeID,
        UserID: middleware.GetUserID(r.Context()),
    }

    updateCmd := db.UpdateUserAudited(r.Context(), auditCtx, userID, params)
    if err := audited.Update(updateCmd); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

### HTTP Handler - Delete

```go
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
    userID := types.UserID(chi.URLParam(r, "id"))

    auditCtx := audited.AuditContext{
        NodeID: config.NodeID,
        UserID: middleware.GetUserID(r.Context()),
    }

    deleteCmd := db.DeleteUserAudited(r.Context(), auditCtx, userID)
    if err := audited.Delete(deleteCmd); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

---

## Implementation Order

### Step 1: Create Audited Package (Day 1)
- [ ] Create `internal/db/audited/audited.go`
- [ ] Implement `AuditContext` struct
- [ ] Implement command interfaces
- [ ] Implement `Create`, `Update`, `Delete` functions
- [ ] Implement `withTx` helper
- [ ] Add import to `internal/db/audited` (avoid circular imports)

### Step 2: Implement User Commands (Day 1)
- [ ] Add command structs to `user.go`
- [ ] Write test for audited create
- [ ] Write test for audited update
- [ ] Write test for audited delete
- [ ] Verify transaction rollback on failure

### Step 3: Implement Remaining Entities (Day 2-3)
- [ ] Route commands
- [ ] ContentData commands
- [ ] ContentField commands
- [ ] Datatype commands
- [ ] DatatypeField commands
- [ ] Field commands
- [ ] Media commands
- [ ] MediaDimension commands
- [ ] Permission commands
- [ ] Role commands
- [ ] Session commands
- [ ] Table commands
- [ ] Token commands
- [ ] UserOauth commands
- [ ] UserSshKey commands
- [ ] AdminRoute commands
- [ ] AdminDatatype commands
- [ ] AdminDatatypeField commands
- [ ] AdminField commands
- [ ] AdminContentData commands
- [ ] AdminContentField commands
- [ ] Backup commands
- [ ] BackupSet commands
- [ ] BackupVerification commands

### Step 4: Integration (Day 4)
- [ ] Update HTTP handlers to use audited operations
- [ ] Add audit context middleware
- [ ] Test end-to-end with real database
- [ ] Verify change_events table populated correctly

---

## Testing Strategy

### Unit Tests

```go
func TestAuditedCreate_Success(t *testing.T) {
    // Setup: mock db that succeeds
    // Execute: audited.Create(cmd)
    // Assert: result returned, change event recorded
}

func TestAuditedCreate_MutationFails_NoAuditEvent(t *testing.T) {
    // Setup: mock db where CreateUser fails
    // Execute: audited.Create(cmd)
    // Assert: error returned, no change event recorded (rolled back)
}

func TestAuditedCreate_AuditFails_MutationRolledBack(t *testing.T) {
    // Setup: mock db where CreateUser succeeds but RecordChangeEvent fails
    // Execute: audited.Create(cmd)
    // Assert: error returned, user not in database (rolled back)
}

func TestAuditedUpdate_CapturesBeforeState(t *testing.T) {
    // Setup: existing user in db
    // Execute: audited.Update(cmd)
    // Assert: change event has old_values with previous state
}

func TestAuditedDelete_CapturesDeletedRecord(t *testing.T) {
    // Setup: existing user in db
    // Execute: audited.Delete(cmd)
    // Assert: change event has old_values with deleted record
}
```

### Integration Tests

```go
func TestAuditedCreate_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    ctx := context.Background()
    auditCtx := audited.AuditContext{
        NodeID: types.NewNodeID(),
        UserID: types.NewUserID(),
    }

    params := db.CreateUserParams{
        Username: "testuser",
        Email:    types.Email("test@example.com"),
        // ...
    }

    cmd := db.NewUser(ctx, auditCtx, params)
    user, err := audited.Create(cmd)
    require.NoError(t, err)

    // Verify user exists
    fetched, err := db.GetUser(user.UserID)
    require.NoError(t, err)
    assert.Equal(t, "testuser", fetched.Username)

    // Verify change event exists
    events, err := db.GetChangeEventsByRecord("users", string(user.UserID))
    require.NoError(t, err)
    require.Len(t, events, 1)
    assert.Equal(t, types.OperationINSERT, events[0].Operation)
}
```

---

## Considerations

### Circular Import Prevention

The `audited` package imports `db` for `DbDriver` and `RecordChangeEventParams`. Entity files import `audited` for `AuditContext`. This is a one-way dependency:

```
audited → db (ok)
db/user.go → audited (ok)
audited → db/user.go (NO - would be circular)
```

The command interfaces use generics, so `audited` doesn't need to know about specific entity types.

### Performance

- One additional query per Update/Delete (GetBefore)
- JSON serialization for audit values
- Transaction overhead

For high-throughput systems, consider:
- Async audit logging with outbox pattern
- Batch audit events
- Sampling for non-critical operations

### Future Enhancements

1. **Audit event streaming**: Push events to message queue for real-time processing
2. **Audit compaction**: Archive old events, keep recent for quick access
3. **Audit queries**: API endpoints for querying change history
4. **Undo support**: Use audit events to implement undo functionality

---

## Files to Create/Modify

| Action | File | Description |
|--------|------|-------------|
| Create | `internal/db/audited/audited.go` | Core audit package |
| Modify | `internal/db/user.go` | Add command structs |
| Modify | `internal/db/route.go` | Add command structs |
| Modify | `internal/db/datatype.go` | Add command structs |
| Modify | `internal/db/datatype_field.go` | Add command structs |
| Modify | `internal/db/field.go` | Add command structs |
| Modify | `internal/db/content_data.go` | Add command structs |
| Modify | `internal/db/content_field.go` | Add command structs |
| Modify | `internal/db/media.go` | Add command structs |
| Modify | `internal/db/media_dimension.go` | Add command structs |
| Modify | `internal/db/permission.go` | Add command structs |
| Modify | `internal/db/role.go` | Add command structs |
| Modify | `internal/db/session.go` | Add command structs |
| Modify | `internal/db/table.go` | Add command structs |
| Modify | `internal/db/token.go` | Add command structs |
| Modify | `internal/db/user_oauth.go` | Add command structs |
| Modify | `internal/db/user_ssh_keys.go` | Add command structs |
| Modify | `internal/db/admin_route.go` | Add command structs |
| Modify | `internal/db/admin_datatype.go` | Add command structs |
| Modify | `internal/db/admin_datatype_field.go` | Add command structs |
| Modify | `internal/db/admin_field.go` | Add command structs |
| Modify | `internal/db/admin_content_data.go` | Add command structs |
| Modify | `internal/db/admin_content_field.go` | Add command structs |
| Modify | `internal/db/backup.go` | Add command structs |
| Create | `internal/db/audited/audited_test.go` | Unit tests |
| Create | `internal/db/audited/integration_test.go` | Integration tests |

**Total**: 2 new files, 23 modified files
