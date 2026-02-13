# Dependency Injection Refactor Plan

**Goal:** Restructure ModulaCMS to follow idiomatic Go patterns that demonstrate expertise to senior backend engineers at distributed systems companies.

**Outcome:**
- Consumer-defined interfaces with narrow adapters (not passing god interface)
- Shared domain types in `db/types` (no duplicate struct definitions)
- Atomic audited operations via `db/audited` package (mutation + audit in same transaction)
- Interface-based command pattern for audited CRUD (idiomatic Go)
- Context propagation with cancellation, timeout handling, and audit metadata
- Connection lifecycle management (health checks, graceful shutdown)
- Structured error handling with sentinel errors
- Observability hooks (request correlation, structured logging, metrics)
- AuditContext flowing from HTTP request through to database operations
- Testable, maintainable code

**Related Documents:**
- `AUDIT_PACKAGE_REVISED_DESIGN.md` - Detailed audit package design with command pattern

---

## Current State

```
cmd/main.go
    └── imports router, cli, db, middleware, etc.

router/
    └── imports db, auth, media, backup, middleware, model, transform

auth/
    └── imports db
    └── creates own db connection via db.ConfigDB()
    └── has global state (globalStateStore, globalVerifierStore)

middleware/
    └── imports db, auth

model/
    └── imports db

cli/
    └── imports db, model

9 packages import db directly
```

## Target State

```
cmd/main.go
    └── imports all packages
    └── creates db driver once
    └── creates adapter types for narrow interfaces
    └── creates services with specific adapters
    └── creates handler groups with specific adapters
    └── wires routes with observability

router/
    └── defines interfaces it needs
    └── imports db/types for domain types (allowed)
    └── NO import of db driver, auth, media packages
    └── handler groups accept narrow interfaces via constructors

auth/
    └── defines interfaces it needs
    └── imports db/types for domain types (allowed)
    └── NO import of db driver
    └── no global state

middleware/
    └── defines interfaces it needs
    └── imports db/types (allowed)
    └── NO import of db driver, auth

model/
    └── imports db/types (allowed)
    └── NO import of db driver
    └── pure domain logic

cli/
    └── defines interfaces it needs
    └── imports db/types (allowed)
    └── NO import of db driver
    └── CLIServices container for Bubbletea model
```

**Clarification:** "NO db import" means no import of the database driver package (`internal/db`). Importing `internal/db/types` for shared domain types (User, Session, IDs, enums) is allowed and encouraged to avoid duplicate struct definitions.

---

## Phases

### Phase 0: Establish Test Baseline

**Why:** You need verification that the refactor doesn't break existing behavior. Without tests, you're flying blind.

**Files:** All packages

**Changes:**

1. Audit existing test coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

2. Identify critical paths with no coverage:
   - User authentication flow
   - OAuth provisioning
   - Content CRUD operations
   - Media upload
   - Session management

3. Write integration tests for critical paths before refactoring:
```go
// internal/auth/auth_integration_test.go
func TestProvisionUser_NewUser(t *testing.T) { ... }
func TestProvisionUser_ExistingEmail_LinksOAuth(t *testing.T) { ... }
func TestProvisionUser_ExistingOAuth_UpdatesTokens(t *testing.T) { ... }
```

4. Create test fixtures that can be reused across phases

**Verification:**
- [ ] Coverage report generated
- [ ] Critical paths identified and documented
- [ ] Integration tests written for auth, middleware auth check, content CRUD
- [ ] All new tests pass against current code

---

### Phase 1: Enhance Database Layer Foundation

**Why:** Distributed systems require context propagation, transaction support, connection lifecycle management, and proper error handling.

**Files:** `internal/db/db.go`, `internal/db/types/`, all entity files

**Changes:**

#### 1.1 Add Domain Types to db/types

Move shared structs to `db/types` so all packages can import them without importing the driver:

```go
// internal/db/types/entities.go
package types

// User represents a user account
type User struct {
    UserID       UserID
    Username     string
    Name         string
    Email        Email
    Hash         string
    Role         int64
    DateCreated  Timestamp
    DateModified Timestamp
}

// Session represents a user session
type Session struct {
    SessionID SessionID
    UserID    NullableUserID
    Token     string
    ExpiresAt Timestamp
    CreatedAt Timestamp
}

// UserOauth represents OAuth credentials
type UserOauth struct {
    UserOauthID         UserOauthID
    UserID              NullableUserID
    OauthProvider       string
    OauthProviderUserID string
    AccessToken         string
    RefreshToken        string
    TokenExpiresAt      string
    DateCreated         Timestamp
}

// ... other entities (ContentData, Media, etc.)
```

#### 1.2 Add Sentinel Errors

```go
// internal/db/types/errors.go
package types

import "errors"

var (
    ErrNotFound      = errors.New("record not found")
    ErrDuplicate     = errors.New("duplicate record")
    ErrForeignKey    = errors.New("foreign key constraint violated")
    ErrInvalidInput  = errors.New("invalid input")
    ErrTxClosed      = errors.New("transaction already closed")
)

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
    return errors.Is(err, ErrNotFound)
}
```

#### 1.3 Add Context to All Methods

```go
// Before
GetUser(types.UserID) (*User, error)

// After
GetUser(ctx context.Context, id types.UserID) (*types.User, error)
```

#### 1.4 Add Transaction Support

```go
// internal/db/types/transaction.go
package types

import (
    "context"
    "database/sql"
)

// Tx represents a database transaction
type Tx interface {
    Commit() error
    Rollback() error
}

// TxBeginner can begin transactions
type TxBeginner interface {
    BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

// TxFunc is a function that runs within a transaction
type TxFunc func(ctx context.Context, tx Tx) error

// WithTransaction executes fn within a transaction, handling commit/rollback
func WithTransaction(ctx context.Context, db TxBeginner, fn TxFunc) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    if err := fn(ctx, tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback failed: %v (original: %w)", rbErr, err)
        }
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}
```

#### 1.5 Add Connection Lifecycle Methods

```go
// internal/db/db.go

type DbDriver interface {
    // Lifecycle
    Ping(ctx context.Context) error
    Close() error
    Stats() sql.DBStats

    // Transactions
    BeginTx(ctx context.Context, opts *sql.TxOptions) (types.Tx, error)

    // ... existing methods with context added
}
```

#### 1.6 Implement Health Check

```go
// internal/db/health.go
package db

import (
    "context"
    "time"
)

func (d *Database) Ping(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return d.Connection.PingContext(ctx)
}

func (d *Database) Close() error {
    if d.Connection != nil {
        return d.Connection.Close()
    }
    return nil
}

func (d *Database) Stats() sql.DBStats {
    return d.Connection.Stats()
}
```

**Verification:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] All db methods accept context as first parameter
- [ ] `types.ErrNotFound` returned for missing records
- [ ] `BeginTx` implemented for all three drivers
- [ ] `Ping`, `Close`, `Stats` implemented
- [ ] Domain types moved to `db/types/entities.go`

---

### Phase 1b: Create Audited Operations Package

**Why:** Atomic audited operations (mutation + change event in same transaction) are critical for data integrity and audit compliance. The command pattern is idiomatic Go and provides clean call sites.

**Reference:** See `AUDIT_PACKAGE_REVISED_DESIGN.md` for full design details.

**Files:**
- `internal/db/audited/interfaces.go`
- `internal/db/audited/context.go`
- `internal/db/audited/change_event.go`
- `internal/db/audited/audited.go`

**Changes:**

#### 1b.1 Create Command Interfaces

```go
// internal/db/audited/interfaces.go
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

// CreateCommand bundles context, audit context, and params for audited creates.
type CreateCommand[T any] interface {
    Context() context.Context
    AuditContext() AuditContext
    Connection() *sql.DB
    TableName() string
    Execute(context.Context, DBTX) (T, error)
    GetID(T) string
    Params() any
}

// UpdateCommand bundles context, audit context, and params for audited updates.
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

// DeleteCommand bundles context, audit context, and params for audited deletes.
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

#### 1b.2 Create AuditContext

```go
// internal/db/audited/context.go
package audited

import "github.com/hegner123/modulacms/internal/db/types"

// AuditContext carries metadata for audit records.
// Flows from HTTP request through handlers to database operations.
type AuditContext struct {
    NodeID    types.NodeID // Distributed system node ID
    UserID    types.UserID // Authenticated user
    RequestID string       // For distributed tracing (from observability)
    IP        string       // Client IP for security audits
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

#### 1b.3 Create Generic Audited Functions

```go
// internal/db/audited/audited.go
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

    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }

    err := types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
        created, err := cmd.Execute(ctx, tx)
        if err != nil {
            return fmt.Errorf("execute: %w", err)
        }
        result = created

        newValues, err := json.Marshal(created)
        if err != nil {
            return fmt.Errorf("marshal: %w", err)
        }

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
    // Similar pattern - see AUDIT_PACKAGE_REVISED_DESIGN.md
}

// Delete executes an audited delete operation.
// Captures before-state, executes delete, records deletion atomically.
func Delete[T any](cmd DeleteCommand[T]) error {
    // Similar pattern - see AUDIT_PACKAGE_REVISED_DESIGN.md
}
```

#### 1b.4 Add RecordChangeEventTx to sqlc Queries

Add to each driver's queries.sql:

```sql
-- sqlite/queries.sql
-- name: RecordChangeEventTx :exec
INSERT INTO change_events (
    event_id, hlc_timestamp, node_id, table_name, record_id,
    operation, user_id, old_values, new_values, request_id, ip,
    wall_timestamp
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'));
```

#### 1b.5 Create Command Structs for Users (Template)

```go
// internal/db/user.go - add command structs

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
    return queries.CreateUser(ctx, mdb.CreateUserParams{...})
}

// Factory method on Database
func (d Database) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
    return NewUserCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// Similar patterns for UpdateUserCmd, DeleteUserCmd
// Similar patterns for MySQL (NewUserCmdMysql) and PostgreSQL variants
```

**Verification:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/db/audited/...` passes
- [ ] `audited.Create` commits both user and change_event on success
- [ ] `audited.Create` rolls back both on failure
- [ ] Users command structs implemented for all three drivers
- [ ] Change events recorded with correct operation, timestamps, user_id

---

### Phase 2: Add Observability Infrastructure

**Why:** Senior engineers at distributed systems companies expect request correlation, structured logging, and metrics. This separates "clean code" from "production-ready."

**Files:** New `internal/observability/` package, updates to `internal/utility/`

**Changes:**

#### 2.1 Create Request Context

```go
// internal/observability/context.go
package observability

import (
    "context"
    "crypto/rand"
    "encoding/hex"
)

type contextKey string

const (
    requestIDKey contextKey = "request_id"
    userIDKey    contextKey = "user_id"
)

// RequestID generates a new request ID
func RequestID() string {
    b := make([]byte, 8)
    rand.Read(b)
    return hex.EncodeToString(b)
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}

// WithUserID adds user ID to context (after auth)
func WithUserID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, userIDKey, id)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
    if id, ok := ctx.Value(userIDKey).(string); ok {
        return id
    }
    return ""
}
```

#### 2.2 Create Structured Logger Interface

```go
// internal/observability/logger.go
package observability

import "context"

// Logger interface for structured logging
type Logger interface {
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Warn(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, msg string, err error, fields ...Field)
}

// Field is a key-value pair for structured logging
type Field struct {
    Key   string
    Value any
}

func F(key string, value any) Field {
    return Field{Key: key, Value: value}
}

// DefaultLogger wraps existing charmbracelet logger with context support
type DefaultLogger struct {
    // wrap utility.DefaultLogger
}

func (l *DefaultLogger) Info(ctx context.Context, msg string, fields ...Field) {
    // Extract request_id, user_id from ctx
    // Add to log output
    reqID := GetRequestID(ctx)
    // ... log with correlation
}
```

#### 2.3 Create Request ID Middleware

```go
// internal/middleware/request_id.go
package middleware

import (
    "net/http"
    "github.com/hegner123/modulacms/internal/observability"
)

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = observability.RequestID()
        }

        ctx := observability.WithRequestID(r.Context(), requestID)
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### 2.4 Helper to Extract AuditContext from Request

```go
// internal/observability/audit.go
package observability

import (
    "net/http"
    "github.com/hegner123/modulacms/internal/db/audited"
    "github.com/hegner123/modulacms/internal/db/types"
)

// AuditContextFromRequest extracts audit metadata from HTTP request.
// Combines observability (request ID) with auth (user ID) and request info (IP).
func AuditContextFromRequest(r *http.Request, nodeID types.NodeID) audited.AuditContext {
    return audited.Ctx(
        nodeID,
        types.UserID(GetUserID(r.Context())), // From auth middleware
        GetRequestID(r.Context()),             // From request ID middleware
        GetClientIP(r),                        // From request
    )
}

// GetClientIP extracts client IP from request, handling proxies.
func GetClientIP(r *http.Request) string {
    // Check X-Forwarded-For, X-Real-IP, then RemoteAddr
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        // Take first IP if comma-separated
        if idx := strings.Index(xff, ","); idx != -1 {
            return strings.TrimSpace(xff[:idx])
        }
        return xff
    }
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }
    // Strip port from RemoteAddr
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}
```

**Verification:**
- [ ] `observability` package created
- [ ] Request ID middleware works
- [ ] Logger extracts correlation IDs from context
- [ ] Request IDs appear in logs
- [ ] `AuditContextFromRequest` helper works
- [ ] Client IP extraction handles proxied requests

---

### Phase 3: Refactor auth Package

**Why:** auth is mid-level complexity, used by router and middleware. Good proof of concept.

**Files:**
- `internal/auth/auth.go`
- `internal/auth/state_store.go`
- `internal/auth/token_refresh.go`
- `internal/auth/user_provision.go`

**Changes:**

#### 3.1 Create Consumer Interfaces

```go
// internal/auth/interfaces.go
package auth

import (
    "context"
    "github.com/hegner123/modulacms/internal/db/types"
)

// UserReader defines read operations auth needs for users
type UserReader interface {
    GetUser(ctx context.Context, id types.UserID) (*types.User, error)
    GetUserByEmail(ctx context.Context, email types.Email) (*types.User, error)
}

// UserWriter defines write operations auth needs for users
type UserWriter interface {
    CreateUser(ctx context.Context, params types.CreateUserParams) (*types.User, error)
}

// OAuthReader defines read operations for OAuth
type OAuthReader interface {
    GetUserOauthByUserId(ctx context.Context, id types.NullableUserID) (*types.UserOauth, error)
    GetUserOauthByProviderID(ctx context.Context, provider, providerUserID string) (*types.UserOauth, error)
}

// OAuthWriter defines write operations for OAuth
type OAuthWriter interface {
    CreateUserOauth(ctx context.Context, params types.CreateUserOauthParams) (*types.UserOauth, error)
    UpdateUserOauth(ctx context.Context, params types.UpdateUserOauthParams) error
}

// TxBeginner for atomic operations
type TxBeginner interface {
    BeginTx(ctx context.Context, opts *sql.TxOptions) (types.Tx, error)
}
```

Note: Uses `types.User`, `types.UserOauth` from `db/types` - no duplicate struct definitions.

#### 3.2 Create Auth Service

```go
// internal/auth/service.go
package auth

import (
    "context"
    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/db/types"
    "github.com/hegner123/modulacms/internal/observability"
)

// Service provides authentication and user provisioning
type Service struct {
    config    *config.Config
    logger    observability.Logger
    users     UserReader
    userWrite UserWriter
    oauth     OAuthReader
    oauthWrite OAuthWriter
    txBeginner TxBeginner
    states    *StateStore
    verifiers *VerifierStore
}

// New creates an Auth service with injected dependencies
func New(
    cfg *config.Config,
    logger observability.Logger,
    users UserReader,
    userWrite UserWriter,
    oauth OAuthReader,
    oauthWrite OAuthWriter,
    txBeginner TxBeginner,
) *Service {
    return &Service{
        config:     cfg,
        logger:     logger,
        users:      users,
        userWrite:  userWrite,
        oauth:      oauth,
        oauthWrite: oauthWrite,
        txBeginner: txBeginner,
        states:     NewStateStore(),
        verifiers:  NewVerifierStore(),
    }
}

// ProvisionUser creates or links a user via OAuth (atomic operation)
func (s *Service) ProvisionUser(
    ctx context.Context,
    info *UserInfo,
    token *oauth2.Token,
    provider string,
) (*types.User, error) {
    s.logger.Info(ctx, "provisioning user",
        observability.F("provider", provider),
        observability.F("email", info.Email),
    )

    // Check existing OAuth link
    existingOauth, err := s.oauth.GetUserOauthByProviderID(ctx, provider, info.ProviderUserID)
    if err != nil && !types.IsNotFound(err) {
        return nil, fmt.Errorf("check existing oauth: %w", err)
    }

    if existingOauth != nil {
        // Update tokens and return user
        if err := s.oauthWrite.UpdateUserOauth(ctx, types.UpdateUserOauthParams{
            UserOauthID:  existingOauth.UserOauthID,
            AccessToken:  token.AccessToken,
            RefreshToken: token.RefreshToken,
        }); err != nil {
            s.logger.Warn(ctx, "failed to update tokens", observability.F("error", err))
        }
        return s.users.GetUser(ctx, existingOauth.UserID.ID)
    }

    // Check existing user by email
    existingUser, err := s.users.GetUserByEmail(ctx, types.Email(info.Email))
    if err != nil && !types.IsNotFound(err) {
        return nil, fmt.Errorf("check existing user: %w", err)
    }

    if existingUser != nil {
        // Link OAuth to existing user
        return s.linkOAuth(ctx, existingUser, info, token, provider)
    }

    // Create new user with OAuth (atomic)
    return s.createUserWithOAuth(ctx, info, token, provider)
}

// createUserWithOAuth creates user and OAuth link atomically
func (s *Service) createUserWithOAuth(
    ctx context.Context,
    info *UserInfo,
    token *oauth2.Token,
    provider string,
) (*types.User, error) {
    var user *types.User

    err := types.WithTransaction(ctx, s.txBeginner, func(ctx context.Context, tx types.Tx) error {
        var err error
        user, err = s.userWrite.CreateUser(ctx, types.CreateUserParams{
            Username:     info.Username,
            Name:         info.Name,
            Email:        types.Email(info.Email),
            Role:         4,
            DateCreated:  types.TimestampNow(),
            DateModified: types.TimestampNow(),
        })
        if err != nil {
            return fmt.Errorf("create user: %w", err)
        }

        _, err = s.oauthWrite.CreateUserOauth(ctx, types.CreateUserOauthParams{
            UserID:              types.NullableUserID{ID: user.UserID, Valid: true},
            OauthProvider:       provider,
            OauthProviderUserID: info.ProviderUserID,
            AccessToken:         token.AccessToken,
            RefreshToken:        token.RefreshToken,
            DateCreated:         types.TimestampNow(),
        })
        if err != nil {
            return fmt.Errorf("create oauth: %w", err)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    s.logger.Info(ctx, "created new user via oauth",
        observability.F("user_id", user.UserID),
        observability.F("provider", provider),
    )
    return user, nil
}
```

#### 3.3 Remove Global State

```go
// internal/auth/state_store.go

// Before
var globalStateStore = &StateStore{...}

// After - no global, created in New()
func NewStateStore() *StateStore {
    return &StateStore{
        states: make(map[string]time.Time),
    }
}
```

#### 3.4 Delete Old Files

- Delete `token_refresh.go` (consolidated)
- Delete `user_provision.go` (consolidated)

**Verification:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/auth/...` passes
- [ ] No `db` import in auth package (only `db/types`)
- [ ] No global variables
- [ ] Transaction used for user+oauth creation
- [ ] Structured logging with context
- [ ] Integration tests still pass

---

### Phase 4: Refactor middleware Package

**Files:** `internal/middleware/*.go`

**Changes:**

#### 4.1 Create Consumer Interfaces

```go
// internal/middleware/interfaces.go
package middleware

import (
    "context"
    "github.com/hegner123/modulacms/internal/db/types"
)

// SessionReader defines session lookup
type SessionReader interface {
    GetSessionByToken(ctx context.Context, token string) (*types.Session, error)
}

// UserReader defines user lookup
type UserReader interface {
    GetUser(ctx context.Context, id types.UserID) (*types.User, error)
}
```

#### 4.2 Update Auth Middleware

```go
// internal/middleware/auth.go
package middleware

import (
    "net/http"
    "github.com/hegner123/modulacms/internal/observability"
)

type AuthMiddleware struct {
    logger   observability.Logger
    sessions SessionReader
    users    UserReader
}

func NewAuthMiddleware(
    logger observability.Logger,
    sessions SessionReader,
    users UserReader,
) *AuthMiddleware {
    return &AuthMiddleware{
        logger:   logger,
        sessions: sessions,
        users:    users,
    }
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        token := extractToken(r)
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        session, err := m.sessions.GetSessionByToken(ctx, token)
        if err != nil {
            m.logger.Debug(ctx, "invalid session token", observability.F("error", err))
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        // Add user ID to context for downstream logging
        ctx = observability.WithUserID(ctx, session.UserID.ID.String())

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Verification:**
- [ ] No `db` import (only `db/types`)
- [ ] No `auth` import
- [ ] User ID added to context after auth
- [ ] Structured logging

---

### Phase 5: Refactor media Package

**Files:** `internal/media/*.go`

**Changes:**

Similar pattern:
- Define `MediaReader`, `MediaWriter`, `DimensionWriter` interfaces
- Use `types.Media`, `types.MediaDimension` from `db/types`
- Accept interfaces in constructor
- Add structured logging
- Remove `db` import

**Verification:**
- [ ] No `db` import (only `db/types`)
- [ ] Integration tests pass

---

### Phase 6: Refactor backup Package

**Files:** `internal/backup/*.go`

**Changes:**

Similar pattern. Backup likely needs transaction support for consistent snapshots.

**Verification:**
- [ ] No `db` import (only `db/types`)
- [ ] Transaction support for consistent backup

---

### Phase 7: Refactor model Package

**Files:** `internal/model/*.go`

**Changes:**

1. Model should contain pure domain logic
2. If model needs data access, define interfaces
3. Remove `db` import (keep `db/types`)

**Verification:**
- [ ] No `db` import (only `db/types`)
- [ ] Pure functions where possible

---

### Phase 8: Refactor router Package

**Why:** Largest package. Split into sub-phases to reduce risk.

**Files:** 25 files, ~4600 lines

#### Phase 8a: Create Infrastructure

Create shared router infrastructure:

```go
// internal/router/router.go
package router

import "net/http"

// buildChain creates middleware chain
func buildChain(mw ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
    return func(final http.Handler) http.Handler {
        for i := len(mw) - 1; i >= 0; i-- {
            final = mw[i](final)
        }
        return final
    }
}

// respond is a helper for JSON responses
func respond(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// respondError is a helper for error responses
func respondError(w http.ResponseWriter, status int, message string) {
    respond(w, status, map[string]string{"error": message})
}
```

#### Phase 8b: AuthHandlers

```go
// internal/router/auth_handlers.go
package router

import (
    "context"
    "net/http"
    "github.com/hegner123/modulacms/internal/db/types"
    "github.com/hegner123/modulacms/internal/observability"
)

// AuthService defines what auth handlers need
type AuthService interface {
    ProvisionUser(ctx context.Context, info *UserInfo, token *oauth2.Token, provider string) (*types.User, error)
    ValidateState(state string) error
    GenerateState() (string, error)
    StoreVerifier(state, verifier string)
    GetVerifier(state string) (string, error)
}

// SessionWriter for creating sessions after login
type SessionWriter interface {
    CreateSession(ctx context.Context, params types.CreateSessionParams) (*types.Session, error)
}

// SessionDeleter for logout
type SessionDeleter interface {
    DeleteSessionByToken(ctx context.Context, token string) error
}

type AuthHandlers struct {
    config   *config.Config
    logger   observability.Logger
    auth     AuthService
    sessions SessionWriter
    sessionsDelete SessionDeleter
}

func NewAuthHandlers(
    cfg *config.Config,
    logger observability.Logger,
    auth AuthService,
    sessions SessionWriter,
    sessionsDelete SessionDeleter,
) *AuthHandlers {
    return &AuthHandlers{
        config:         cfg,
        logger:         logger,
        auth:           auth,
        sessions:       sessions,
        sessionsDelete: sessionsDelete,
    }
}

func (h *AuthHandlers) RegisterRoutes(mux *http.ServeMux, mw ...func(http.Handler) http.Handler) {
    chain := buildChain(mw...)

    mux.Handle("POST /api/v1/auth/login", chain(http.HandlerFunc(h.Login)))
    mux.Handle("POST /api/v1/auth/logout", chain(http.HandlerFunc(h.Logout)))
    mux.Handle("GET /api/v1/auth/me", chain(http.HandlerFunc(h.Me)))
    mux.Handle("POST /api/v1/auth/register", chain(http.HandlerFunc(h.Register)))
    mux.Handle("GET /api/v1/auth/oauth/login", chain(http.HandlerFunc(h.OAuthInitiate)))
    mux.Handle("GET /api/v1/auth/oauth/callback", chain(http.HandlerFunc(h.OAuthCallback)))
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    h.logger.Info(ctx, "login attempt")
    // ... implementation using h.auth, h.sessions
}
```

**Verification for 8b:**
- [ ] AuthHandlers compiles
- [ ] No `db` import
- [ ] Auth flow works end-to-end

#### Phase 8b-1: UserHandlers with Audited Operations

This shows the pattern for handlers that perform audited CRUD:

```go
// internal/router/user_handlers.go
package router

import (
    "encoding/json"
    "net/http"

    "github.com/hegner123/modulacms/internal/db/audited"
    "github.com/hegner123/modulacms/internal/db/types"
    "github.com/hegner123/modulacms/internal/observability"
)

// UserCommands defines the command factory methods handlers need.
// These are satisfied by db.Database, db.MysqlDatabase, db.PsqlDatabase.
type UserCommands interface {
    NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params types.CreateUserParams) audited.CreateCommand[any]
    UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params types.UpdateUserParams) audited.UpdateCommand[any]
    DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) audited.DeleteCommand[any]
}

// UserReader for non-audited reads
type UserReader interface {
    GetUser(ctx context.Context, id types.UserID) (*types.User, error)
    ListUsers(ctx context.Context) ([]*types.User, error)
}

type UserHandlers struct {
    config   *config.Config
    logger   observability.Logger
    nodeID   types.NodeID
    commands UserCommands
    reader   UserReader
}

func NewUserHandlers(
    cfg *config.Config,
    logger observability.Logger,
    nodeID types.NodeID,
    commands UserCommands,
    reader UserReader,
) *UserHandlers {
    return &UserHandlers{
        config:   cfg,
        logger:   logger,
        nodeID:   nodeID,
        commands: commands,
        reader:   reader,
    }
}

func (h *UserHandlers) RegisterRoutes(mux *http.ServeMux, mw ...func(http.Handler) http.Handler) {
    chain := buildChain(mw...)

    mux.Handle("GET /api/v1/users", chain(http.HandlerFunc(h.List)))
    mux.Handle("GET /api/v1/users/{id}", chain(http.HandlerFunc(h.Get)))
    mux.Handle("POST /api/v1/users", chain(http.HandlerFunc(h.Create)))
    mux.Handle("PUT /api/v1/users/{id}", chain(http.HandlerFunc(h.Update)))
    mux.Handle("DELETE /api/v1/users/{id}", chain(http.HandlerFunc(h.Delete)))
}

func (h *UserHandlers) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Parse and validate params (types validate during unmarshal)
    var params types.CreateUserParams
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        h.logger.Debug(ctx, "invalid request body", observability.F("error", err))
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // 2. Set server-generated fields
    params.UserID = types.NewUserID()
    params.DateCreated = types.TimestampNow()
    params.DateModified = types.TimestampNow()

    // 3. Build audit context from request (request ID, user ID, IP flow through)
    auditCtx := observability.AuditContextFromRequest(r, h.nodeID)

    // 4. Create command - bundles ctx, auditCtx, params
    cmd := h.commands.NewUserCmd(ctx, auditCtx, params)

    // 5. Execute - atomic (user + change_event in same transaction)
    user, err := audited.Create(cmd)
    if err != nil {
        h.logger.Error(ctx, "create user failed", err)
        respondError(w, http.StatusInternalServerError, "failed to create user")
        return
    }

    h.logger.Info(ctx, "user created",
        observability.F("user_id", user.UserID),
    )

    respond(w, http.StatusCreated, user)
}

func (h *UserHandlers) Update(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := types.UserID(r.PathValue("id"))

    if err := id.Validate(); err != nil {
        respondError(w, http.StatusBadRequest, "invalid user id")
        return
    }

    var params types.UpdateUserParams
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    params.UserID = id
    params.DateModified = types.TimestampNow()

    auditCtx := observability.AuditContextFromRequest(r, h.nodeID)
    cmd := h.commands.UpdateUserCmd(ctx, auditCtx, params)

    if err := audited.Update(cmd); err != nil {
        if types.IsNotFound(err) {
            respondError(w, http.StatusNotFound, "user not found")
            return
        }
        h.logger.Error(ctx, "update user failed", err)
        respondError(w, http.StatusInternalServerError, "failed to update user")
        return
    }

    h.logger.Info(ctx, "user updated", observability.F("user_id", id))
    w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandlers) Delete(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := types.UserID(r.PathValue("id"))

    if err := id.Validate(); err != nil {
        respondError(w, http.StatusBadRequest, "invalid user id")
        return
    }

    auditCtx := observability.AuditContextFromRequest(r, h.nodeID)
    cmd := h.commands.DeleteUserCmd(ctx, auditCtx, id)

    if err := audited.Delete(cmd); err != nil {
        if types.IsNotFound(err) {
            respondError(w, http.StatusNotFound, "user not found")
            return
        }
        h.logger.Error(ctx, "delete user failed", err)
        respondError(w, http.StatusInternalServerError, "failed to delete user")
        return
    }

    h.logger.Info(ctx, "user deleted", observability.F("user_id", id))
    w.WriteHeader(http.StatusNoContent)
}

// List and Get use reader interface (no audit needed for reads)
func (h *UserHandlers) List(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    users, err := h.reader.ListUsers(ctx)
    if err != nil {
        h.logger.Error(ctx, "list users failed", err)
        respondError(w, http.StatusInternalServerError, "failed to list users")
        return
    }
    respond(w, http.StatusOK, users)
}

func (h *UserHandlers) Get(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := types.UserID(r.PathValue("id"))

    user, err := h.reader.GetUser(ctx, id)
    if err != nil {
        if types.IsNotFound(err) {
            respondError(w, http.StatusNotFound, "user not found")
            return
        }
        h.logger.Error(ctx, "get user failed", err)
        respondError(w, http.StatusInternalServerError, "failed to get user")
        return
    }
    respond(w, http.StatusOK, user)
}
```

**Key patterns demonstrated:**
- `UserCommands` interface for audited operations (command factories)
- `UserReader` interface for non-audited reads
- `observability.AuditContextFromRequest` extracts request ID, user ID, IP
- `audited.Create/Update/Delete` execute atomic operations
- Structured logging with context correlation
- Sentinel error handling (`types.IsNotFound`)

#### Phase 8c: ContentHandlers

Convert contentData.go, contentFields.go

**Verification:** Same as 8b

#### Phase 8d: AdminHandlers

Convert adminContentData.go, adminContentFields.go, adminDatatypes.go, adminFields.go, adminRoutes.go

**Verification:** Same as 8b

#### Phase 8e: MediaHandlers

Convert media.go, mediaDimensions.go, mediaUpload.go

**Verification:** Same as 8b

#### Phase 8f: UserHandlers

Convert users.go, roles.go, sessions.go, tokens.go, userOauth.go, ssh_keys.go

**Verification:** Same as 8b

#### Phase 8g: SchemaHandlers

Convert datatypes.go, fields.go

**Verification:** Same as 8b

#### Phase 8h: RouteHandlers

Convert routes.go, slugs.go

**Verification:** Same as 8b

#### Phase 8i: ImportHandlers

Convert import.go

**Verification:** Same as 8b

#### Phase 8j: SystemHandlers

Convert tables.go, restore.go

**Verification:** Same as 8b

#### Phase 8k: Wire and Clean Up

- Update mux.go to use new handler groups
- Delete old standalone handler functions
- Final verification

**Verification:**
- [ ] All routes work
- [ ] No `db` import in router (only `db/types`)
- [ ] All integration tests pass

---

### Phase 9: Create Adapter Types in main

**Why:** Avoid passing god interface. Create narrow adapters that only expose what each consumer needs.

**File:** `cmd/main.go`

**Changes:**

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/hegner123/modulacms/internal/auth"
    "github.com/hegner123/modulacms/internal/backup"
    "github.com/hegner123/modulacms/internal/bucket"
    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/db"
    "github.com/hegner123/modulacms/internal/media"
    "github.com/hegner123/modulacms/internal/middleware"
    "github.com/hegner123/modulacms/internal/observability"
    "github.com/hegner123/modulacms/internal/router"
)

// Adapter types - expose only what each consumer needs
// This prevents consumers from type-asserting back to DbDriver

type authUserReader struct{ driver db.DbDriver }
func (a authUserReader) GetUser(ctx context.Context, id types.UserID) (*types.User, error) {
    return a.driver.GetUser(ctx, id)
}
func (a authUserReader) GetUserByEmail(ctx context.Context, email types.Email) (*types.User, error) {
    return a.driver.GetUserByEmail(ctx, email)
}

type authUserWriter struct{ driver db.DbDriver }
func (a authUserWriter) CreateUser(ctx context.Context, params types.CreateUserParams) (*types.User, error) {
    return a.driver.CreateUser(ctx, params)
}

type authOAuthReader struct{ driver db.DbDriver }
func (a authOAuthReader) GetUserOauthByUserId(ctx context.Context, id types.NullableUserID) (*types.UserOauth, error) {
    return a.driver.GetUserOauthByUserId(ctx, id)
}
func (a authOAuthReader) GetUserOauthByProviderID(ctx context.Context, provider, providerUserID string) (*types.UserOauth, error) {
    return a.driver.GetUserOauthByProviderID(ctx, provider, providerUserID)
}

// ... more adapters for each consumer interface

func main() {
    // Load configuration
    cfg := config.Load()

    // Create logger
    logger := observability.NewLogger()

    // Create infrastructure (single db connection)
    driver := db.ConfigDB(cfg)
    defer driver.Close()

    // Verify database connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    if err := driver.Ping(ctx); err != nil {
        log.Fatalf("database ping failed: %v", err)
    }
    cancel()

    bucketClient := bucket.New(&cfg)

    // Create services with narrow adapters
    authSvc := auth.New(
        &cfg,
        logger,
        authUserReader{driver},
        authUserWriter{driver},
        authOAuthReader{driver},
        authOAuthWriter{driver},
        driver, // TxBeginner
    )

    mediaSvc := media.New(
        &cfg,
        logger,
        mediaStoreAdapter{driver},
        dimensionStoreAdapter{driver},
        bucketClient,
    )

    backupSvc := backup.New(&cfg, logger, driver)

    // Create middleware
    requestIDMW := middleware.RequestID
    corsMW := middleware.CORS(&cfg)
    rateLimitMW := middleware.RateLimit(10, 60)
    authMW := middleware.NewAuthMiddleware(
        logger,
        sessionReaderAdapter{driver},
        userReaderAdapter{driver},
    )

    // Create handler groups with narrow interfaces
    authHandlers := router.NewAuthHandlers(
        &cfg,
        logger,
        authSvc,
        sessionWriterAdapter{driver},
        sessionDeleterAdapter{driver},
    )
    mediaHandlers := router.NewMediaHandlers(&cfg, logger, mediaSvc)
    contentHandlers := router.NewContentHandlers(&cfg, logger, contentAdapter{driver})
    // ... more handler groups

    // Build mux
    mux := http.NewServeMux()

    // All routes get request ID middleware first
    wrappedMux := requestIDMW(mux)

    // Public routes
    authHandlers.RegisterRoutes(mux, corsMW, rateLimitMW)
    routeHandlers.RegisterPublicRoutes(mux, corsMW)

    // Protected routes
    mediaHandlers.RegisterRoutes(mux, corsMW, authMW.Handler)
    contentHandlers.RegisterRoutes(mux, corsMW, authMW.Handler)
    // ... more routes

    // Create server with timeouts
    srv := &http.Server{
        Addr:         cfg.HTTPPort,
        Handler:      wrappedMux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    go func() {
        logger.Info(context.Background(), "starting server", observability.F("addr", cfg.HTTPPort))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info(context.Background(), "shutting down server")

    ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    logger.Info(context.Background(), "server stopped")
}
```

**Verification:**
- [ ] Application starts
- [ ] All routes work
- [ ] Graceful shutdown works
- [ ] Request IDs in all logs
- [ ] Health check passes

---

### Phase 10: Refactor cli Package

**Why:** CLI has different architecture (Bubbletea). Needs its own design.

**Files:** `internal/cli/*.go` (~20 files, ~3000 lines)

**Architecture:**

Bubbletea uses the Elm architecture:
- `Model` holds state
- `Update(msg) (Model, Cmd)` handles events
- `View() string` renders

The challenge: Models can't easily accept interfaces in constructors because they're created/recreated during state transitions.

**Solution: CLIServices container**

```go
// internal/cli/services.go
package cli

import (
    "context"
    "github.com/hegner123/modulacms/internal/db/types"
)

// Services holds all dependencies the CLI needs
type Services struct {
    Logger      Logger
    Users       UserService
    Content     ContentService
    Datatypes   DatatypeService
    Media       MediaService
    // ... etc
}

// UserService defines user operations for CLI
type UserService interface {
    List(ctx context.Context) ([]*types.User, error)
    Get(ctx context.Context, id types.UserID) (*types.User, error)
    Create(ctx context.Context, params types.CreateUserParams) (*types.User, error)
    Update(ctx context.Context, params types.UpdateUserParams) error
    Delete(ctx context.Context, id types.UserID) error
}

// ContentService defines content operations for CLI
type ContentService interface {
    List(ctx context.Context) ([]*types.ContentData, error)
    // ...
}
```

```go
// internal/cli/model.go
package cli

type Model struct {
    services *Services  // Pointer to services, shared across all models
    // ... other fields
}

func NewModel(services *Services, config *config.Config) Model {
    return Model{
        services: services,
        config:   config,
        // ...
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Access services via m.services.Users.List(ctx)
}
```

```go
// cmd/main.go (CLI startup)
cliServices := &cli.Services{
    Logger:    logger,
    Users:     userServiceAdapter{driver},
    Content:   contentServiceAdapter{driver},
    Datatypes: datatypeServiceAdapter{driver},
    Media:     mediaServiceAdapter{driver},
}

model := cli.NewModel(cliServices, &cfg)
p := tea.NewProgram(model)
```

**Sub-phases for CLI:**

1. **10a:** Create Services struct and interfaces
2. **10b:** Create adapter types
3. **10c:** Update Model to use Services
4. **10d:** Update all view/update functions
5. **10e:** Wire in main

**Verification:**
- [ ] CLI starts
- [ ] All screens work
- [ ] No `db` import (only `db/types`)
- [ ] Services container passed correctly

---

### Phase 11: Organize db Package (Optional)

**Files:** `internal/db/db.go`

**Changes:**

Organize `DbDriver` with embedded interfaces for documentation purposes:

```go
type UserOperations interface {
    GetUser(ctx context.Context, id types.UserID) (*types.User, error)
    GetUserByEmail(ctx context.Context, email types.Email) (*types.User, error)
    CreateUser(ctx context.Context, params types.CreateUserParams) (*types.User, error)
    UpdateUser(ctx context.Context, params types.UpdateUserParams) error
    DeleteUser(ctx context.Context, id types.UserID) error
    ListUsers(ctx context.Context) ([]*types.User, error)
}

type SessionOperations interface { ... }
type ContentOperations interface { ... }
type MediaOperations interface { ... }

type DbDriver interface {
    // Lifecycle
    Ping(ctx context.Context) error
    Close() error
    Stats() sql.DBStats
    BeginTx(ctx context.Context, opts *sql.TxOptions) (types.Tx, error)

    // Domain operations
    UserOperations
    SessionOperations
    ContentOperations
    MediaOperations
    // ... etc
}
```

This is purely organizational - doesn't change behavior.

---

## Execution Order Summary

| Phase | Package | Effort | Risk | Depends On |
|-------|---------|--------|------|------------|
| 0 | tests | Medium | Low | None |
| 1 | db foundation | High | Medium | Phase 0 |
| 2 | observability | Medium | Low | None |
| 3 | auth | Medium | Low | Phases 1, 2 |
| 4 | middleware | Low | Low | Phases 1, 2 |
| 5 | media | Medium | Low | Phases 1, 2 |
| 6 | backup | Low | Low | Phases 1, 2 |
| 7 | model | Low | Low | Phase 1 |
| 8a-k | router (11 sub-phases) | High | Medium | Phases 3-7 |
| 9 | main adapters | Medium | Low | Phase 8 |
| 10a-e | cli (5 sub-phases) | High | Medium | Phases 1-7 |
| 11 | db organize | Low | Low | Any time |

**Recommended approach:**
1. Complete Phases 0-3 as proof of concept
2. Verify everything works
3. Continue with remaining phases
4. Each sub-phase of 8 and 10 is a separate PR

---

## Verification Checklist (Per Phase)

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] No import cycles: `go list -f '{{.ImportPath}} -> {{.Imports}}' ./internal/... | grep -v vendor`
- [ ] Target package has no `db` driver import (only `db/types` allowed)
- [ ] No global variables in refactored package
- [ ] Context propagated through all operations
- [ ] Structured logging with request correlation
- [ ] Sentinel errors used for not found, etc.

---

## Final Import Graph

```
cmd/main.go
    ├── config         (load configuration)
    ├── db             (create driver - ONLY place that imports db driver)
    ├── db/types       (shared domain types)
    ├── bucket         (create S3 client)
    ├── observability  (create logger)
    ├── auth           (create service)
    ├── media          (create service)
    ├── backup         (create service)
    ├── middleware     (create middleware)
    └── router         (create handlers, register routes)

router/
    └── config
    └── db/types       (for User, Session, etc.)
    └── observability
    └── NO db driver import

auth/
    └── config
    └── db/types
    └── observability
    └── NO db driver import

middleware/
    └── config
    └── db/types
    └── observability
    └── NO db driver import

media/
    └── config
    └── db/types
    └── observability
    └── NO db driver import

cli/
    └── config
    └── db/types
    └── observability
    └── NO db driver import
```

---

## What This Demonstrates to Senior Engineers

| Pattern | What It Shows |
|---------|---------------|
| Consumer-defined interfaces | Understands Go's interface philosophy |
| Adapter types in main | Knows how to prevent interface pollution |
| Shared types in db/types | Avoids duplicate definitions pragmatically |
| Context propagation | Distributed systems awareness |
| Transaction support | Data integrity understanding |
| Sentinel errors | Proper error handling patterns |
| Structured logging with correlation | Production observability |
| Graceful shutdown | Reliability engineering |
| Phase-by-phase refactor | Manages risk in large changes |

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking changes during refactor | Phase 0 tests, feature branch, phase-by-phase PRs |
| Interface signature mismatches | Compiler catches these |
| Missing context propagation | Grep for db calls without ctx |
| Performance regression | Benchmark before/after Phase 1 |
| CLI complexity | Dedicated sub-phases with own design |
| Adapter boilerplate | Accept as cost of proper DI |
| Test coverage gaps | Phase 0 identifies and fills gaps |
