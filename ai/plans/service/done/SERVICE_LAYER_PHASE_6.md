# Phase 6: Thin CRUD Sweep

## Context

Phases 0-4 established the service layer and extracted the high-value domains: Schema, Content, AdminContent, Media, Routes, Users, RBAC. Phase 5 (Plugins, Webhooks, Locales) is tracked separately. Phase 6 extracts the remaining 10 Tier 3 services — all thin CRUD with limited business logic beyond decode-call-encode.

**Goal:** After Phase 6, all 10 remaining Tier 3 services delegate to the service layer. No Tier 3 handler calls `db.ConfigDB(c)` or takes `config.Config` as a dispatcher parameter.

**Out of scope — `config.Config` references that remain after Phase 6:**
- `auth.go` — Phase 7 scope (auth refactor)
- `globals.go` (`GlobalsHandler`) — content delivery straggler from Phase 2
- `contentComposite.go` (`ContentCreateHandler`) — content delivery straggler from Phase 2
- `query.go` (`QueryHandler`) — content delivery straggler from Phase 2
- `adminTree.go` (`AdminTreeHandler`) — content tree straggler from Phase 2
- `health.go` (`HealthHandler`) — infrastructure, not a domain service; takes `PluginHealthChecker`, not raw driver calls

These are content delivery / system health handlers missed in earlier phases. They work fine with `*c` since the mux closure has access. A post-Phase 6 cleanup pass or Phase 7 can address them. `slugs.go` and `scheduler.go` are already migrated to `*service.Registry`.

### Current State (as of Phase 4 completion)

**Registry struct** (`internal/service/service.go`):
- Fields: `driver`, `mgr`, `pc`, `emailSvc`, `dispatcher`
- Domain services: `Schema`, `Content`, `AdminContent`, `Media`, `Routes`, `Users`, `RBAC`
- Getter methods: `Driver()`, `Config()`, `Manager()`, `PermissionCache()`, `EmailService()`, `Dispatcher()`

**Migrated handlers** (use `*service.Registry`):
- Content, AdminContent, Schema (Phase 1-2)
- Media, Routes (Phase 3)
- Users, Roles, Permissions, RolePermissions (Phase 4)

**Not yet migrated** (still use `config.Config` dispatcher or `(driver, mgr)` closures):
- Sessions, Tokens, SSH Keys, OAuth, Tables, Config, Import, Deploy, Audit, Backup

**Placeholder stubs exist** for all 10 services in `internal/service/`.

---

## Key Design Decisions

**Services are thin wrappers for thin CRUD.** Sessions, OAuth, Tables have zero business logic beyond passthrough. Their services exist only for: (a) consistent error mapping (`NotFoundError`, `ValidationError`), (b) uniform audit context threading, (c) single entry point for all consumers. Service methods may be as short as 3 lines.

**No service-layer input types for pure passthrough.** Services like SessionService and OAuthService accept `db.UpdateSessionParams` / `db.CreateUserOauthParams` directly. Creating wrapper types adds boilerplate with zero benefit when the service doesn't transform the data. Exception: TokenService needs `CreateTokenInput` because it generates and hashes the token.

**Config service wraps `config.Manager`, not `DbDriver`.** ConfigService is the only service that doesn't use the driver. It exposes redacted reads, validated updates, and field metadata. The router's existing `ConfigGetHandler(mgr)` / `ConfigUpdateHandler(mgr)` closure pattern becomes `ConfigGetHandler(svc)` / `ConfigUpdateHandler(svc)`.

**Import service extracts the core logic, not the HTTP layer.** The 480-line `apiImportContent` + `importRootToDatabase` complex in `router/import.go` moves to `internal/service/import_svc.go`. The handler becomes: parse body → call `svc.Import.ImportContent(ctx, ac, body, format, routeID)`. Format-specific handlers (`ImportContentfulHandler`, etc.) become one-liner dispatchers.

**Deploy service wraps existing `internal/deploy/` functions.** The deploy package already has clean separation: `ExportPayload()`, `ImportPayload()`, `BuildDryRunResult()`. DeployService wraps these with dependency injection (driver, config) and consistent error types. The handlers in `deploy/server.go` stay in the deploy package but delegate to the service. Alternative: move handlers to `internal/router/` — but this is a refactor beyond Phase 6 scope. Accept the split.

**Backup service has no HTTP handlers.** BackupService wraps `backup.CreateFullBackup()` and `backup.RestoreFromBackup()` with config injection. Only consumed by TUI and CLI. No router changes needed.

**Admin import handler stays unimplemented.** `ImportSubmitHandler` currently returns "Processing is not yet implemented." Phase 6 wires it to the service but keeps it as a TODO — implementing the multipart-to-service bridge is straightforward once the service exists, but the format detection from file upload is new feature work, not extraction.

**`auth.go` is out of scope.** The auth handlers (login, register, OAuth flows, password reset) are deeply coupled to session management, cookie handling, and OAuth provider logic. They'll be addressed in Phase 7 (MCP alignment) or a dedicated auth refactor.

---

## Service Specifications

### A. SessionService (`sessions.go`)

**Dependencies:** `db.DbDriver`

**Methods:**
```
CreateSession(ctx, ac, params db.CreateSessionParams) (*db.Sessions, error)
GetSession(ctx, sessionID) (*db.Sessions, error)
ListSessions(ctx) (*[]db.Sessions, error)
UpdateSession(ctx, ac, params db.UpdateSessionParams) (*db.Sessions, error)
DeleteSession(ctx, ac, sessionID) error
```

**Business logic:** None. Pure delegation with NotFoundError mapping on Get/Delete. `CreateSession` is included for completeness — currently called by auth handlers (Phase 7 scope), but available for any consumer (TUI, tests) that needs to create sessions programmatically.

**DbDriver methods used:** `CreateSession`, `GetSession`, `ListSessions`, `UpdateSession`, `DeleteSession`

---

### B. TokenService (`tokens.go`)

**Dependencies:** `db.DbDriver`

**Input types:**
```go
// CreateTokenInput holds fields for API token creation.
// The service generates the raw token, hashes it, and stores the hash.
type CreateTokenInput struct {
    Label   string
    Expiry  types.Timestamp // caller provides; admin handler defaults to 365 days
}

// CreateTokenResult returns the created token record plus the raw token
// (shown once, never stored).
type CreateTokenResult struct {
    Token    *db.Tokens
    RawToken string // "mcms_" + hex(32 random bytes)
}
```

**Methods:**
```
GetToken(ctx, tokenID) (*db.Tokens, error)
ListTokens(ctx) (*[]db.Tokens, error)
CreateToken(ctx, ac, input CreateTokenInput) (*CreateTokenResult, error)
UpdateToken(ctx, ac, params db.UpdateTokenParams) (*db.Tokens, error)
DeleteToken(ctx, ac, tokenID) error
```

**Business logic:**
- `CreateToken`: generates 32 random bytes → hex encode → prepend `mcms_` prefix → hash via `utility.HashToken()` → store hash, return raw token in result. **Note:** This is a deliberate behavior change for the API handler. The current API handler (`router/tokens.go`) accepts a client-provided raw token string and hashes it. The admin handler already generates tokens server-side. The service unifies both to server-side generation. Existing API consumers that generate their own token strings will need to update.
- Fetch-after-update on UpdateToken (driver returns `*string`, not token)

**DbDriver methods used:** `GetToken`, `ListTokens`, `CreateToken`, `UpdateToken`, `DeleteToken`

---

### C. SSHKeyService (`ssh_keys.go`)

**Dependencies:** `db.DbDriver`

**Input types:**
```go
type AddSSHKeyInput struct {
    UserID    types.UserID
    PublicKey string
    Label     string
}
```

**Methods:**
```
AddKey(ctx, ac, input AddSSHKeyInput) (*db.UserSshKeys, error)
ListKeys(ctx, userID types.UserID) (*[]db.UserSshKeys, error)
ListKeysByFingerprint(ctx, userID, fingerprint string) (*db.UserSshKeys, error)
DeleteKey(ctx, ac, userID types.UserID, keyID types.UserSshKeyID) error
```

**Business logic:**
- `AddKey`: validates public key format via `middleware.ParseSSHPublicKey()` → extracts keyType + fingerprint → checks duplicate by fingerprint → creates record
- `DeleteKey`: verifies ownership (key.UserID == requested userID) before deleting
- `ListKeys`: filters by authenticated user (no cross-user access)

**DbDriver methods used:** `CreateUserSshKey`, `ListUserSshKeys`, `GetUserSshKey`, `GetUserSshKeyByFingerprint`, `DeleteUserSshKey`

**Note:** `middleware.ParseSSHPublicKey` is a middleware function. The service imports middleware for this utility. This is acceptable — it's a stateless parser, not request-scoped middleware.

---

### D. OAuthService (`oauth.go`)

**Dependencies:** `db.DbDriver`

**Methods:**
```
CreateUserOauth(ctx, ac, params db.CreateUserOauthParams) (*db.UserOauth, error)
UpdateUserOauth(ctx, ac, params db.UpdateUserOauthParams) (*db.UserOauth, error)
DeleteUserOauth(ctx, ac, oauthID types.UserOauthID) error
```

**Business logic:** None. Pure delegation with error mapping. Accepts db params directly.

**DbDriver methods used:** `CreateUserOauth`, `UpdateUserOauth`, `DeleteUserOauth`

---

### E. TableService (`tables.go`)

**Dependencies:** `db.DbDriver`

**Methods:**
```
GetTable(ctx, tableID types.TableID) (*db.Tables, error)
ListTables(ctx) (*[]db.Tables, error)
UpdateTable(ctx, ac, params db.UpdateTableParams) (*db.Tables, error)
DeleteTable(ctx, ac, tableID types.TableID) error
```

**Business logic:** None. Fetch-after-update on UpdateTable (driver returns `*string`).

**DbDriver methods used:** `GetTable`, `ListTables`, `UpdateTable`, `DeleteTable`

---

### F. ConfigService (`config_svc.go`)

**Dependencies:** `*config.Manager` (NOT DbDriver)

**Methods:**
```
GetConfig(ctx) (json.RawMessage, error)          // redacted full config as JSON
GetConfigByCategory(ctx, category string) (json.RawMessage, error)  // filtered by category
GetFieldMetadata(ctx) (json.RawMessage, error)    // field registry with labels, types, categories
UpdateConfig(ctx, updates map[string]any) (*ConfigUpdateResult, error)
```

**Types:**
```go
type ConfigUpdateResult struct {
    RestartRequired bool
    Applied         map[string]any  // fields that were changed
    Warnings        []string        // non-fatal issues
}
```

**Business logic:**
- `GetConfig`: calls `config.RedactedConfig()` to mask sensitive fields
- `GetConfigByCategory`: calls `config.FieldsByCategory()` for filtered view
- `UpdateConfig`: delegates to `mgr.Update()` which validates fields and returns restart hints
- `GetFieldMetadata`: calls `config.FieldRegistryJSON()` for schema metadata

**Note:** No audit context needed — config updates are already logged by the Manager.

---

### G. ImportService (`import_svc.go`)

**Dependencies:** `db.DbDriver`, `*config.Manager`

**Input types:**
```go
type ImportContentInput struct {
    Format  string          // "contentful", "sanity", "strapi", "wordpress", "clean"
    Body    io.Reader       // raw JSON body
    RouteID types.RouteID   // optional; zero value = no route association
}

type ImportResult struct {
    DatatypesCreated int
    ContentCreated   int
    FieldsCreated    int
    Errors           []string
}
```

**Methods:**
```
ImportContent(ctx, ac, input ImportContentInput) (*ImportResult, error)
```

**Business logic (moved from `router/import.go`):**
- Format validation → transformer lookup via `transform.NewTransformConfig()` + `GetTransformer()`
- Read and parse body through transformer
- `importRootToDatabase()` — recursive tree import:
  - Datatype deduplication cache (`"label|type"` → DatatypeID)
  - Node creation with parent/child/sibling linking
  - Sibling pointer patching (doubly-linked list) after all children created
  - Error accumulation (partial success)
- Route association (optional)

**Files to move:**
- `importRootToDatabase`, `importNode`, `findOrCreateDatatype`, `createFieldAndContentField`, `patchSiblingPointers` — all move from `router/import.go` to `service/import_svc.go`
- `ImportResult` type moves to service (handler keeps `ImportResult` as response type, or references service type directly)

**DbDriver methods used:** `CreateDatatype`, `CreateContentData`, `CreateField`, `CreateContentField`, `UpdateContentData`, `GetContentData`

---

### H. DeployService (`deploy.go`)

**Dependencies:** `db.DbDriver`, `*config.Manager`

**Methods:**
```
Health(ctx) map[string]any
Export(ctx, tables []string) (*deploy.SyncPayload, error)
Import(ctx, payload *deploy.SyncPayload, dryRun bool) (*deploy.ImportResult, error)
DryRun(ctx, payload *deploy.SyncPayload) (*deploy.DryRunResult, error)
```

**Business logic:**
- `Export`: validates table names via `db.ValidateTableName()`, delegates to `deploy.ExportPayload()`
- `Import`: validates payload, delegates to `deploy.ImportPayload()`, handles concurrent import gating
- `DryRun`: delegates to `deploy.BuildDryRunResult()`
- `Health`: returns version, node_id, status

**Note:** The `deploy` package functions (`ExportPayload`, `ImportPayload`, `BuildDryRunResult`) stay in `internal/deploy/`. The service wraps them with dependency injection and error type mapping. Handlers in `deploy/server.go` are updated to accept `*service.Registry` instead of `config.Config`.

**Implementation:** Create a minimal DeployService that delegates to the deploy package functions and update the 3 handler signatures. The service wraps `deploy.ExportPayload()` / `deploy.ImportPayload()` / `deploy.BuildDryRunResult()` with dependency injection. Handlers in `deploy/server.go` accept `*service.Registry` and call `svc.Deploy.*` methods. Note: the deploy package functions take `config.Config` by value (not `*config.Manager`), so the service calls `s.mgr.Config()` to get a config snapshot at each invocation — correct for hot-reload since config may change between calls. Concurrent import gating remains in the `deploy` package's `importMu` mutex; the service does not add concurrency control.

---

### I. AuditLogService (`audit_log.go`)

**Dependencies:** `db.DbDriver`

**Methods:**
```
ListChangeEvents(ctx, limit, offset int32) (*[]db.ChangeEvents, error)
CountChangeEvents(ctx) (*int64, error)
```

**Business logic:** None. Read-only pagination wrapper with error mapping. No mutations, no audit context.

**DbDriver methods used:** `ListChangeEvents`, `CountChangeEvents`

---

### J. BackupService (`backup.go`)

**Dependencies:** `*config.Manager`

**Methods:**
```
CreateFullBackup(ctx) (path string, sizeBytes int64, err error)
RestoreFromBackup(ctx, backupPath string) error
ReadManifest(backupPath string) (*backup.BackupManifest, error)
```

**Business logic:** Delegates to `backup.CreateFullBackup()` and `backup.RestoreFromBackup()`. Resolves config from manager.

**Note:** No HTTP handlers. Consumed by TUI (`msg_backup.go`) and CLI commands. Registry wiring makes it available to both without global state.

---

## Implementation Order

Phase 6 has 10 independent services with no cross-dependencies. Group into 4 batches by complexity, with each batch fully parallelizable internally.

### Step 0: Registry Wiring + Mux Rewire (prerequisite, all batches)

Two things happen in Step 0 before any batch work begins:

**0a. Registry fields.** Add all 10 service fields to `Registry` struct and initialize in `NewRegistry()`.

```go
// Add to Registry struct:
Sessions  *SessionService
Tokens    *TokenService
SSHKeys   *SSHKeyService
OAuth     *OAuthService
Tables    *TableService
ConfigSvc *ConfigService    // "Config" collides with Config() method
Import    *ImportService
Deploy    *DeployService
AuditLog  *AuditLogService
Backup    *BackupService
```

```go
// Add to NewRegistry():
reg.Sessions = NewSessionService(driver)
reg.Tokens = NewTokenService(driver)
reg.SSHKeys = NewSSHKeyService(driver)
reg.OAuth = NewOAuthService(driver)
reg.Tables = NewTableService(driver)
reg.ConfigSvc = NewConfigService(mgr)
reg.Import = NewImportService(driver, mgr)
reg.Deploy = NewDeployService(driver, mgr)
reg.AuditLog = NewAuditLogService(driver)
reg.Backup = NewBackupService(mgr)
```

**0b. All handler signatures + mux.go registrations.** Change every handler signature to `(w, r, svc)` / `(svc)` and update all ~25 mux.go registrations in a single pass. Handler bodies can temporarily call `svc.Driver()` / `svc.Config()` to get the old dependencies — this compiles immediately and keeps all handlers working. Then batches fill in the actual service method calls without touching mux.go.

**Why upfront:** mux.go is a single file. In Phase 5, concurrent agents editing mux.go caused merge conflicts. By doing all mux changes in Step 0, batches are fully independent — they only modify service files and handler method bodies.

### Batch A: Pure Passthrough (Sessions, OAuth, Tables)

**Why first:** Zero business logic. Each service is ~30-50 lines. Fast to implement, fast to verify. Proves the batch pattern works.

**Per service:**
1. Implement service methods (replace stub)
2. Rewire API handler bodies to call service instead of `svc.Driver()` passthrough
3. Rewire admin handler bodies (sessions.go only — OAuth and Tables have no admin handlers)

**Estimated lines changed per service:** ~80-120

### Batch B: Moderate Logic (Tokens, SSH Keys, AuditLog)

**Tokens:** Token generation + hashing logic moves from both admin and API handlers into service. `CreateTokenInput` / `CreateTokenResult` types needed. Admin handler's crypto/rand + hex encoding moves to service.

**SSH Keys:** Key parsing, fingerprint dedup, ownership check move into service. `AddSSHKeyInput` type needed.

**AuditLog:** Read-only, simplest of the batch. Pagination params passed through.

**Per service:**
1. Implement service with input types where needed
2. Rewire API handler bodies to call service
3. Rewire admin handler bodies (tokens.go, audit.go — SSH keys has no admin handler)

**Estimated lines changed per service:** ~100-200

### Batch C: Config (special pattern)

**Unique:** Wraps `*config.Manager`, not `db.DbDriver`. No audit context. Router handlers use closure-factory pattern (`func(mgr) http.Handler`), not dispatcher.

**Steps:**
1. Implement ConfigService wrapping manager methods
2. Update handler bodies to call `svc.ConfigSvc.*` instead of `mgr.*`

**Estimated lines changed:** ~120

### Batch D: Orchestration (Import, Deploy, Backup)

**Import:** Largest change. ~300 lines of business logic move from `router/import.go` to `service/import_svc.go`. Handler becomes thin dispatch (parse body + call service). Admin handler's `ImportSubmitHandler` stays unimplemented but signature changes to `(svc)`.

**Deploy:** Minimal service wrapper around existing `deploy.ExportPayload()` / `deploy.ImportPayload()`. Update 3 handler signatures. Handlers stay in `deploy/server.go` but accept `*service.Registry`.

**Backup:** No handler changes. Service wraps `backup.CreateFullBackup()` / `backup.RestoreFromBackup()`. TUI consumption is out of scope for Phase 6 (TUI still calls backup package directly; future work to route through service).

**Steps:**
1. Implement ImportService (move logic from router/import.go)
2. Implement DeployService (thin wrapper)
3. Implement BackupService (thin wrapper)
4. Rewire import API handler bodies (6 format handlers + bulk)
5. Rewire import admin handler body (signature already changed in Step 0)
6. Rewire deploy handler bodies (3 handlers)

**Estimated lines changed:** ~500 (mostly import logic relocation)

### Step 5: Cleanup + Verify

- Remove unused imports (`config`, `db` direct) from migrated handler files
- Run `just check` and `just test`
- Verify no Tier 3 handler calls `db.ConfigDB(c)` (grep for it)
- Verify remaining `config.Config` references in `internal/router/` are only the documented out-of-scope handlers (see Goal section)

---

## Handler Signature Changes

### API Handlers (router package)

| File | Current Signature | New Signature | Functions |
|------|-------------------|---------------|-----------|
| `sessions.go` | `(w, r, c config.Config)` | `(w, r, svc *service.Registry)` | SessionsHandler, SessionHandler + 2 internal |
| `tokens.go` | `(w, r, c config.Config)` | `(w, r, svc *service.Registry)` | TokensHandler, TokenHandler + 4 internal |
| `ssh_keys.go` | `(w, r, c config.Config)` | `(w, r, svc *service.Registry)` | AddSSHKeyHandler, ListSSHKeysHandler, DeleteSSHKeyHandler |
| `userOauth.go` | `(w, r, c config.Config)` | `(w, r, svc *service.Registry)` | UserOauthsHandler, UserOauthHandler + 3 internal |
| `tables.go` | `(w, r, c config.Config)` | `(w, r, svc *service.Registry)` | TablesHandler, TableHandler + 4 internal |
| `config.go` | `(mgr *config.Manager)` | `(svc *service.Registry)` | ConfigGetHandler, ConfigUpdateHandler, ConfigMetaHandler |
| `import.go` | `(w, r, c config.Config)` | `(w, r, svc *service.Registry)` | 6 format handlers + ImportBulkHandler |

### Admin Handlers (admin/handlers package)

| File | Current Signature | New Signature | Functions |
|------|-------------------|---------------|-----------|
| `sessions.go` | `(driver)` / `(driver, mgr)` | `(svc *service.Registry)` | SessionsListHandler, SessionDeleteHandler |
| `tokens.go` | `(driver)` / `(driver, mgr)` | `(svc *service.Registry)` | TokensListHandler, TokenCreateHandler, TokenDeleteHandler |
| `audit.go` | `(driver)` | `(svc *service.Registry)` | AuditLogHandler |
| `import_handler.go` | `()` / `(driver)` | `()` / `(svc *service.Registry)` | ImportPageHandler (no change), ImportSubmitHandler |

### Deploy Handlers (deploy package)

| File | Current Signature | New Signature | Functions |
|------|-------------------|---------------|-----------|
| `server.go` | `(w, r, c config.Config)` | `(w, r, svc *service.Registry)` | DeployHealthHandler, DeployExportHandler, DeployImportHandler |

### Mux Registrations (~25 changes)

All remaining `*c` and `pc` references in handler closures become `svc`:
- Sessions: 2 registrations
- Tokens: 2 registrations
- SSH Keys: 3 registrations
- OAuth: 2 registrations
- Tables: 2 registrations
- Config: 3 registrations (auth chain wrapping stays)
- Import: 6 registrations
- Deploy: 3 registrations
- Admin sessions: 2 registrations
- Admin tokens: 3 registrations
- Admin audit: 1 registration
- Admin import: 2 registrations

---

## Files Changed

### New/Rewritten Service Files
| File | Lines (est.) | Notes |
|------|-------------|-------|
| `service/sessions.go` | 40 | Replace stub |
| `service/tokens.go` | 80 | Replace stub; includes generation + hashing |
| `service/ssh_keys.go` | 90 | Replace stub; includes key parsing + ownership |
| `service/oauth.go` | 40 | Replace stub |
| `service/tables.go` | 50 | Replace stub |
| `service/config_svc.go` | 70 | Replace stub; wraps config.Manager |
| `service/import_svc.go` | 350 | Replace stub; logic from router/import.go |
| `service/deploy.go` | 60 | Replace stub; wraps deploy package |
| `service/audit_log.go` | 30 | Replace stub |
| `service/backup.go` | 40 | Replace stub; wraps backup package |
| `service/service.go` | +15 | Registry fields + NewRegistry init |

### Modified Handler Files
| File | Change Type | Notes |
|------|------------|-------|
| `router/sessions.go` | Rewrite | `config.Config` → `svc` |
| `router/tokens.go` | Rewrite | `config.Config` → `svc`; remove hash logic |
| `router/ssh_keys.go` | Rewrite | `config.Config` → `svc`; remove parse/ownership logic |
| `router/userOauth.go` | Rewrite | `config.Config` → `svc` |
| `router/tables.go` | Rewrite | `config.Config` → `svc` |
| `router/config.go` | Rewrite | `config.Manager` → `svc` |
| `router/import.go` | Major rewrite | ~400 lines of logic move to service |
| `deploy/server.go` | Rewrite | `config.Config` → `svc` |
| `admin/handlers/sessions.go` | Rewrite | `(driver, mgr)` → `(svc)` |
| `admin/handlers/tokens.go` | Rewrite | `(driver, mgr)` → `(svc)` |
| `admin/handlers/audit.go` | Rewrite | `(driver)` → `(svc)` |
| `admin/handlers/import_handler.go` | Minor | `(driver)` → `(svc)` on ImportSubmitHandler |
| `router/mux.go` | Edit | ~25 registration changes |

---

## Testing Strategy

**Unit tests per service** (`service/*_test.go`): Table-driven tests against SQLite. Focus on:
- TokenService: verify hash is stored (not raw), raw token format matches `mcms_` + 64 hex chars
- SSHKeyService: verify fingerprint extraction, duplicate rejection, ownership check
- ImportService: verify recursive tree creation, sibling linking, deduplication
- ConfigService: verify redaction masks sensitive fields

**No tests for pure passthrough services** (Sessions, OAuth, Tables): The service adds no logic. Testing would only verify that `service.GetSession(id)` calls `driver.GetSession(id)`. The existing handler/integration tests already cover the end-to-end path. **Boundary:** If any passthrough service later gains business logic (validation, side effects, gating), tests must be added at that time.

**Compile check after each batch:** `just check` + `just test` between batches to catch regressions early.

---

## Risk & Mitigations

**Import logic relocation is the riskiest change.** The 480-line recursive import has no unit tests (only tested via API integration). Mitigation: move the code exactly as-is (no refactoring during the move), verify compilation, then add service-level tests for the core `importRootToDatabase` path.

**Deploy package boundary.** The deploy handlers live in `internal/deploy/`, not `internal/router/`. Updating them to accept `*service.Registry` creates an import from `deploy` → `service`. This is fine architecturally (deploy is a consumer of service, not the reverse). But verify no circular import: `service` must NOT import `deploy`. The DeployService wraps deploy functions by accepting them as parameters or by importing `internal/deploy/` one-directionally.

**Config closure pattern differs from other handlers.** Config handlers return `http.Handler` (not called with `w, r, svc`). The handler factory pattern stays — just replace `mgr` with `svc` in the closure capture. Mux registration already wraps these in middleware chains.

**Backup has no handler coverage in this phase.** The service exists but TUI still calls `backup.CreateFullBackup()` directly. Wiring TUI through the service requires changes to `internal/tui/` which is out of Phase 6 scope. The service is available for future TUI migration.

---

## Parallel Execution Plan

Batches A-D are independent. Step 0 handles all mux.go edits upfront so no batch touches mux.go. With `hq` coordination:

```
Step 0: Registry Wiring + Mux Rewire   (sequential, prerequisite — all signatures + mux.go)
  │
  ├── Batch A: Sessions + OAuth + Tables     (service bodies + handler bodies)
  ├── Batch B: Tokens + SSH Keys + AuditLog  (service bodies + handler bodies)
  ├── Batch C: Config                        (service body + handler body)
  └── Batch D: Import + Deploy + Backup      (service bodies + handler bodies)
  │
Step 5: Cleanup + Verify                (sequential)
```

Maximum parallelism: 10 agents (one per service). Practical parallelism: 4 agents (one per batch), with batch-internal work done sequentially per agent. No merge conflicts on mux.go since it's fully settled in Step 0.
