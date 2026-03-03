# Plan: `modula connect` — Same TUI, Swappable Backend + Project Registry

## Context

The TUI runs on the server (directly or via SSH), so the file picker browses the server's filesystem, terminal colors degrade, and keystrokes have latency. The user wants the same TUI running locally, connecting to any project — local or remote.

A minimal registry at `~/.modula/configs.json` maps project names to `config.json` paths. Everything else — DB driver, remote URL, API keys, environments — lives in each project's own `config.json`. The TUI codebase is shared across modes with mode-aware guards; only the backend driver changes based on what the loaded config says.

## TUI Target

`modula connect` targets the **v1 TUI** (`InitialModel` + `CliRun`) because it accepts a `db.DbDriver` parameter. The panel model (`NewPanelModel` + `PanelRun`) does NOT accept an injected driver — its constructor signature is `NewPanelModel(v *bool, c *config.Config) (PanelModel, error)` with no driver parameter. Panel model support is out of scope for this plan. If panel model support is desired later, it requires modifying `NewPanelModel` to accept an optional `db.DbDriver` and is a separate change.

## Project Registry

`~/.modula/configs.json` — just a name->path map:

```json
{
  "projects": {
    "local": "/Users/home/Documents/Code/Go_dev/modulacms/config.json",
    "staging": "/var/www/staging/config.json",
    "production": "/etc/modulacms/prod/config.json"
  },
  "default": "local"
}
```

That's it. No URLs, keys, or types in the registry. The project's `config.json` has everything. The registry is machine-local (`~/.modula/`) and stores absolute paths, so it is not portable across machines.

### Resolution flow

1. `modula connect production` -> read `configs.json` -> get `/etc/modulacms/prod/config.json`
2. Project root = `filepath.Dir(configPath)` -> `/etc/modulacms/prod/`
3. `os.Chdir(projectRoot)` so relative paths (e.g., `Db_URL: "./modula.db"`) resolve correctly
   - NOTE: `os.Chdir` is a process-global side effect. Acceptable because `modula connect` is a CLI command that runs one TUI per process. Do not use the registry package as a library expecting concurrent multi-project access.
4. Load `config.json` from that path
5. If config has `Remote_URL` set -> launch TUI with `RemoteDriver` (Go SDK over HTTPS)
6. If config has `Db_Driver` set -> launch TUI with local database driver (existing behavior)
7. If neither is set -> error: "config.json must have either remote_url or db_driver"

### Remote project config

A remote project's `config.json` would look like:

```json
{
  "remote_url": "https://cms.example.com",
  "remote_api_key": "mod_abc123...",
  "environment": "production"
}
```

The same `config.json` format, just with remote fields instead of (or alongside) database fields. All credentials, URLs, and settings stay in the project's own config — not in the global registry.

**Obtaining an API key:** The operator must create a token on the remote instance first (via TUI Actions > "Generate API Token", admin panel, or REST API `POST /api/v1/tokens`). The `remote_api_key` value goes in the project's config.json. This is a prerequisite documented in the CLI help text and error messages.

## Architecture

```
~/.modula/configs.json          project config.json
   name -> path                     |-- db_driver + db_url (local)
                                    +-- remote_url + remote_api_key (remote)
                                          |
modula connect <name>                     |
    |                                     |
    |-- Resolve path from registry        |
    |-- chdir to project root             |
    |-- Load config.json -----------------+
    |-- Route driver:
    |   |-- cfg.Remote_URL set? -> remote.NewDriver() + db.SetInstance()
    |   +-- cfg.Db_Driver set?  -> db.ConfigDB() (SQLite/MySQL/PostgreSQL)
    |
    +-- Launch v1 TUI (InitialModel + CliRun)
        +-- File picker browses LOCAL filesystem
```

**Driver routing bypasses `db.InitDB()`.** The `connect` command constructs the driver directly:
- Remote config -> `remote.NewDriver(url, apiKey)` returns a `db.DbDriver`, then `db.SetInstance(driver)` sets the singleton so the 83 `db.ConfigDB(*cfg)` calls in TUI code return the RemoteDriver
- Local config -> `db.ConfigDB(cfg)` returns a local driver (existing fallback path, no singleton needed)

The TUI receives the driver as a parameter to `InitialModel()` and stores it in `m.DB`.

### The `db.ConfigDB()` problem in remote mode

The TUI has **83 call sites** that call `db.ConfigDB(*cfg)` to get a driver. In remote mode, `cfg.Db_Driver` is empty (the config has `remote_url` instead). The `ConfigDB()` fallback switch matches on `Db_Driver` and returns `nil` when no case matches. Every one of those 83 call sites would nil-dereference and panic.

**Solution: Add `db.SetInstance(driver)` so `ConfigDB()` returns the RemoteDriver.**

Phase 2a adds a public `db.SetInstance(DbDriver)` function that sets the `dbInstance` singleton. The `connect` command calls `db.SetInstance(remoteDriver)` before launching the TUI. This way, all 83 `db.ConfigDB(*cfg)` calls return the RemoteDriver via the existing `if dbInstance != nil { return dbInstance }` fast path. Zero TUI call sites need modification.

```go
// internal/db/init.go — new function added in Phase 2a
func SetInstance(d DbDriver) {
    dbInstance = d
}
```

This is preferred over replacing all 83 call sites because:
- It is a 1-line change vs. modifying 9 TUI files
- It uses the existing singleton path that the server already relies on
- The TUI already calls `db.ConfigDB(*cfg)` in `tea.Cmd` closures where `m.DB` is not available (closures capture `cfg`, not `m`)

## TUI Method Audit

The DbDriver interface has ~330 methods across 36 entity types. The TUI calls **83 distinct methods**. The RemoteDriver must implement all 83.

### Methods requiring special handling in remote mode

**7 `GetConnection()` calls — raw SQL access (MUST be guarded):**

| File | Line | Purpose |
|------|------|---------|
| `commands.go` | 77 | `DatabaseInsert` — generic query builder `db.QInsert()` |
| `commands.go` | 110 | `DatabaseUpdate` — generic query builder |
| `commands.go` | 133 | `DatabaseGet` — generic query builder |
| `commands.go` | 160 | `DatabaseList` — generic query builder |
| `commands.go` | 186 | `DatabaseFilteredList` — generic query builder |
| `commands.go` | 214 | `DatabaseDelete` — generic query builder |
| `fields.go` | 11 | Column value suggestions for form autocomplete |

`RemoteDriver.GetConnection()` returns `nil, nil, ErrRemoteMode`. The 6 generic query builder commands (`DatabaseInsert` through `DatabaseDelete`) are used by the Table Browser screen. In remote mode, the Table Browser is hidden from navigation — the TUI skips it when `m.IsRemote` is true. The `fields.go` autocomplete gracefully degrades: when `GetConnection()` errors, it returns an empty suggestion list.

**5 admin-only methods — local database management (disabled in remote mode):**

| Method | File | Line | Purpose |
|--------|------|------|---------|
| `DropAllTables()` | actions.go | 169, 205 | DB Wipe, DB Wipe & Redeploy |
| `CreateAllTables()` | actions.go | 212 | DB Wipe & Redeploy |
| `CreateBootstrapData()` | actions.go | 219 | DB Wipe & Redeploy |
| `ValidateBootstrapData()` | actions.go | 226 | DB Wipe & Redeploy |
| `DumpSql()` | actions.go | 262 | DB Export |

These return `ErrRemoteMode`. The Actions page filters the menu: when `m.IsRemote` is true, only remote-safe actions are shown (Check for Updates, Validate Config, Generate API Token). DB Init, DB Wipe, DB Reset, DB Export, Register SSH Key, Create Backup, and Restore Backup are hidden. Generate API Token works remotely because it calls `CreateToken` via the SDK.

**Audited mutation methods — `audited.AuditContext` is accepted but discarded:**

Many DbDriver methods take `(context.Context, audited.AuditContext, Params)`. The `RemoteDriver` accepts these parameters to satisfy the interface but does not pass them to the SDK. The remote server creates its own audit records from the authenticated request context. This is documented in `RemoteDriver` with a package-level comment explaining the pattern.

### API coverage gaps

| TUI Method | REST API | SDK | RemoteDriver Strategy |
|------------|----------|-----|----------------------|
| `GetContentTreeByRoute()` | No endpoint | `ContentTree.Save` (different operation) | **New endpoint needed**: `GET /api/v1/content/tree/{routeID}` — returns the content tree for a route. Add to router + SDK before Phase 2. |
| `ListContentFieldsByContentDataAndLocale()` | No endpoint | `ContentFields.List` (no locale filter) | Use `ContentFields.RawList` with query params `content_data_id={id}&locale={code}`. Requires adding locale filter to the existing list endpoint. |
| `ListContentVersionsByContent()` | No endpoint | No resource | **New endpoint needed**: `GET /api/v1/contentversions?content_id={id}`. |
| `ListEnabledLocales()` | No endpoint | `Locales.List` (no status filter) | Use `Locales.RawList` with query param `enabled=true`. Requires adding filter to existing endpoint. |
| `GetMaxSortOrderByParentID()` | No endpoint | No method | **New endpoint needed**: `GET /api/v1/fields/max-sort-order?parent_id={id}`. |
| `GetUserByEmail()` | No endpoint | `Users.List` (no email filter) | Use `Users.RawList` with query param `email={email}`. Requires adding filter. |
| `GetUserSshKeyByFingerprint()` | No endpoint | `SSHKeys` resource | Check if SSHKeys has fingerprint lookup; if not, add query param filter. |
| `UpdateFieldSortOrder()` | No endpoint | No method | **New endpoint needed**: `PUT /api/v1/fields/{id}/sort-order`. |

**Required new endpoints (before Phase 2):** 4 new endpoints, 4 filter additions to existing endpoints. This is a prerequisite for Phase 2, not part of it.

## Type Conversion Layer

The plan originally estimated ~20 entity types. The actual count is **36 entity types** with bidirectional conversion needed for entities the TUI mutates. The conversion layer is organized as:

```
internal/remote/
    convert_content.go      # ContentData, ContentFields, ContentRelations, ContentVersions
    convert_admin.go        # Admin* variants of the above + AdminRoutes, AdminDatatypes, AdminFields, AdminFieldTypes
    convert_core.go         # Datatypes, Fields, FieldTypes, Routes, Media, MediaDimensions
    convert_users.go        # Users, Roles, Permissions, RolePermissions, Tokens, Sessions, UserOauths, UserSshKeys
    convert_system.go       # Backups, Locales, Webhooks, Plugins, Pipelines, ChangeEvents, Tables
```

Each file contains:
- `entityToDb(sdk) -> db.Entity` (SDK response -> db type, for reads)
- `entityFromDb(db.CreateEntityParams) -> sdk.CreateEntity` (db params -> SDK request, for mutations)

**Nullable handling rules (no exceptions):**
- `*string` (SDK) <-> `sql.NullString` (db): nil -> `{Valid: false}`, non-nil -> `{String: *v, Valid: true}`
- `*ContentID` (SDK) <-> `types.NullableContentID` (db): same pattern
- `*time.Time` (SDK) <-> `types.Timestamp` (db): parse/format via `types.NewTimestamp()`

Estimated: **~70 conversion functions** across all files (1 read + 1 write per mutated entity, 1 read-only for read-only entities).

## Error Handling Strategy (Remote Mode)

Every SDK call over HTTPS can fail with timeouts, connection resets, TLS errors, HTTP 4xx/5xx, or rate limits. The strategy:

**1. Timeout:** SDK client configured with 15s timeout (not the default 30s). Long enough for normal operations, short enough that the TUI stays responsive.

**2. Error presentation:** All `tea.Cmd` closures that call RemoteDriver methods already return error messages via `tea.Msg`. Network errors are wrapped with context: `fmt.Errorf("remote: %s: %w", operationName, err)`. The TUI displays these in the existing error dialog.

**3. No automatic retry for mutations.** Mutations are not idempotent (e.g., `CreateContentData` generates a server-side ULID). If a mutation times out, the TUI shows an error; the user retries manually. This avoids the "did it apply or not?" problem.

**4. Automatic retry for reads.** Read operations (List*, Get*) retry once after a 1s delay on timeout or 5xx. If the retry fails, show the error.

**5. Connection check at startup.** `modula connect` pings the remote server (`GET /api/v1/health`) before launching the TUI. If unreachable, error immediately with a clear message: "Cannot reach {url}: {error}".

**6. Status bar indicator.** Remote mode shows `[remote: cms.example.com]` in the TUI status bar. On network error, it changes to `[remote: disconnected]` until the next successful request.

## Changes

### Phase 1: Registry + `modula connect` for local projects

**Scope:** 2 new files, 1 modified file. No TUI changes.

**New files:**

| File | Purpose |
|------|---------|
| `cmd/connect.go` | Cobra command: `connect [name]`, `add`, `list`, `remove`, `default` |
| `internal/registry/registry.go` | Load/save `~/.modula/configs.json` |

**Modified files:**

| File | Change |
|------|--------|
| `cmd/root.go` | Register `connectCmd` |

**`internal/registry/registry.go`:**

```go
type Registry struct {
    Projects map[string]string `json:"projects"` // name -> absolute config path
    Default  string            `json:"default"`
}

func Path() string                                       // ~/.modula/configs.json
func Load() (*Registry, error)                           // read from disk; returns empty Registry if file missing
func (r *Registry) Save() error                          // write to disk; creates ~/.modula/ if needed
func (r *Registry) Resolve(name string) (string, error)  // name -> config path; error if not found
func (r *Registry) Add(name, configPath string) error    // stores filepath.Abs(configPath)
func (r *Registry) Remove(name string) error
func (r *Registry) SetDefault(name string) error
```

**`cmd/connect.go`** core flow:

```go
// modula connect [name]
configPath, err := reg.Resolve(name) // get absolute path from registry
projectDir := filepath.Dir(configPath)
os.Chdir(projectDir)                 // process-global; one TUI per process
mgr, driver, err := loadConfigAndDB()  // same helper as cmd/tui.go
cfg, _ := mgr.Config()
model, _ := tui.InitialModel(&verbose, cfg, driver, utility.DefaultLogger, nil, mgr, nil, nil)
tui.CliRun(&model)
```

**CLI:**

```
modula connect                           # Launch TUI for default project
modula connect <name>                    # Launch TUI for named project
modula connect add <name> <config-path>  # Register a project (stores absolute path)
modula connect list                      # List registered projects
modula connect remove <name>            # Remove a project
modula connect default <name>            # Set default project
```

**Done when:**
1. `go build ./...` succeeds
2. `just test` passes (existing tests unchanged)
3. `modula connect add dev ./config.json` from project root -> no error
4. `cd /tmp && modula connect dev` -> TUI opens, file picker browses local filesystem
5. `modula connect list` -> shows `dev` with absolute path
6. `modula connect remove dev` -> removes it, `modula connect list` -> empty
7. `modula connect default dev` -> sets default, `modula connect` (no args) uses it

---

### Phase 1.5: API gap endpoints

**Scope:** 4 new router handlers, 4 filter additions to existing handlers, corresponding SDK methods. No TUI or RemoteDriver changes.

**Why this phase exists:** The RemoteDriver cannot implement 8 of the 83 TUI-used DbDriver methods without REST API endpoints backing them. This is prerequisite work.

**New router endpoints:**

| Endpoint | Method | Handler | Purpose |
|----------|--------|---------|---------|
| `/api/v1/content/tree/{routeID}` | GET | `ContentTreeGetHandler` | Return content tree for a route |
| `/api/v1/contentversions` | GET | `ContentVersionsListHandler` | List versions with `content_id` filter |
| `/api/v1/fields/{id}/sort-order` | PUT | `FieldSortOrderHandler` | Update field sort order |
| `/api/v1/fields/max-sort-order` | GET | `FieldMaxSortOrderHandler` | Get max sort order for parent |

**Filter additions to existing endpoints:**

| Endpoint | New query param | Behavior |
|----------|-----------------|----------|
| `GET /api/v1/contentfields` | `locale` | Filter by locale code |
| `GET /api/v1/locales` | `enabled` | `enabled=true` returns only enabled locales |
| `GET /api/v1/users` | `email` | Return single user matching email |
| `GET /api/v1/usersshkeys` | `fingerprint` | Return single key matching fingerprint |

**DbDriver interface coverage:** All 8 methods in the API coverage gaps table (`GetContentTreeByRoute`, `ListContentFieldsByContentDataAndLocale`, `ListContentVersionsByContent`, `ListEnabledLocales`, `GetMaxSortOrderByParentID`, `GetUserByEmail`, `GetUserSshKeyByFingerprint`, `UpdateFieldSortOrder`) already exist in the `DbDriver` interface and are implemented on all three local drivers. No new `DbDriver` methods are needed — only new HTTP handlers that call the existing driver methods, and corresponding SDK methods.

**Handler pattern** (follow existing patterns in `internal/router/`):

```go
// internal/router/contentTree.go — new handler
// ContentTreeGetHandler returns the content tree for a route.
// Registered as: mux.Handle("GET /api/v1/content/tree/{routeID}",
//   middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//     ContentTreeGetHandler(w, r, driver)
//   })))
func ContentTreeGetHandler(w http.ResponseWriter, r *http.Request, d db.DbDriver) {
    routeID := types.RouteID(r.PathValue("routeID"))
    tree, err := d.GetContentTreeByRoute(routeID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tree)
}
```

Register all new endpoints in `internal/router/mux.go` following the existing pattern: permission guard wrapping a handler that receives `driver` from the closure.

**Filter addition pattern** (for existing handlers):

```go
// In existing list handler, add query param check:
// locale := r.URL.Query().Get("locale")
// if locale != "" {
//     results, err = driver.ListContentFieldsByContentDataAndLocale(contentDataID, locale)
// } else {
//     results, err = driver.ListContentFields(...)
// }
```

**SDK additions (`sdks/go/`):**

- `ContentTree.Get(ctx, routeID)` -> tree data (new method on existing `ContentTreeResource`)
- `ContentVersions` resource (new) with `List` + `content_id` filter
- `Fields.UpdateSortOrder(ctx, id, order)` (new method on existing `Fields` resource)
- `Fields.MaxSortOrder(ctx, parentID)` (new method on existing `Fields` resource)
- Query param support on existing resources for the 4 filter additions (use `RawList` with params)

**Done when:**
1. `go build ./...` succeeds
2. `just test` passes
3. Start server with `just run`, then for each new endpoint:
   - `curl localhost:8080/api/v1/content/tree/{routeID}` -> 200 with tree JSON (or 404 if route doesn't exist)
   - `curl localhost:8080/api/v1/contentversions?content_id={id}` -> 200 with versions array
   - `curl -X PUT localhost:8080/api/v1/fields/{id}/sort-order -d '{"sort_order": 5}'` -> 200
   - `curl localhost:8080/api/v1/fields/max-sort-order?parent_id={id}` -> 200 with `{"max_sort_order": N}`
   - `curl localhost:8080/api/v1/locales?enabled=true` -> returns only enabled locales
   - `curl localhost:8080/api/v1/users?email=system@modula.local` -> returns matching user
4. SDK: `go build ./sdks/go/...` succeeds, new methods exist with correct signatures

---

### Phase 2a: RemoteDriver skeleton + config wiring

**Scope:** 2 new files, 2 modified files. Every method is a stub. The goal is proving the plumbing works end-to-end: config loads, driver constructs, health check passes, TUI launches.

**New files:**

| File | Purpose |
|------|---------|
| `internal/remote/driver.go` | `RemoteDriver` struct, `NewDriver()`, all ~330 DbDriver methods as `ErrNotSupported` stubs |
| `internal/remote/errors.go` | `ErrRemoteMode`, `ErrNotSupported` sentinel errors |

**Modified files:**

| File | Change |
|------|--------|
| `cmd/connect.go` | Route to `remote.NewDriver()` when `cfg.Remote_URL` is set; call `db.SetInstance(driver)` |
| `internal/config/config.go` | Add `Remote_URL string`, `Remote_API_Key string` fields to Config struct |
| `internal/db/init.go` | Add `func SetInstance(d DbDriver)` to set the singleton for remote mode |

**`internal/db/init.go` addition:**

```go
// SetInstance sets the singleton database driver directly.
// Used by modula connect to inject a RemoteDriver so that the 83
// db.ConfigDB(*cfg) call sites in TUI code return the remote driver
// via the existing dbInstance fast path.
func SetInstance(d DbDriver) {
    dbInstance = d
}
```

**`internal/config/config.go` addition:**

```go
// Remote connection (mutually exclusive with Db_Driver for connect command)
Remote_URL     string `json:"remote_url"`
Remote_API_Key string `json:"remote_api_key"`
```

**`internal/remote/errors.go`:**

```go
package remote

import "errors"

// ErrRemoteMode is returned by methods that require a local database connection
// (GetConnection, DropAllTables, CreateAllTables, etc.).
var ErrRemoteMode = errors.New("operation not available in remote mode")

// ErrNotSupported is returned by DbDriver methods that have no remote equivalent.
// The error message includes the method name for debuggability.
type ErrNotSupported struct {
    Method string
}

func (e ErrNotSupported) Error() string {
    return "remote: method not supported: " + e.Method
}
```

**`internal/remote/driver.go`:**

```go
// Package remote implements db.DbDriver over the ModulaCMS Go SDK.
//
// Audited mutation methods accept audited.AuditContext to satisfy the DbDriver
// interface but discard it — the remote server creates audit records from the
// authenticated request context (Bearer token identity + request metadata).
package remote

type RemoteDriver struct {
    client *modula.Client
    url    string
}

func NewDriver(url, apiKey string) (*RemoteDriver, error) {
    client, err := modula.NewClient(modula.ClientConfig{
        BaseURL: url,
        APIKey:  apiKey,
        HTTPClient: &http.Client{Timeout: 15 * time.Second},
    })
    if err != nil {
        return nil, fmt.Errorf("invalid client config for %s: %w", url, err)
    }
    // Verify connectivity before returning
    if _, err := client.Health.Check(context.Background()); err != nil {
        return nil, fmt.Errorf("cannot reach %s: %w", url, err)
    }
    return &RemoteDriver{client: client, url: url}, nil
}

// All ~330 DbDriver methods stubbed as:
// func (d *RemoteDriver) MethodName(...) (...) {
//     return ..., ErrNotSupported{Method: "MethodName"}
// }
//
// GetConnection specifically returns ErrRemoteMode (not ErrNotSupported)
// because TUI code checks for this to hide the table browser.
//
// Stub generation: Copy each method signature from db.DbDriver interface in
// internal/db/db.go (lines 75-542). For each method, return zero values for
// all non-error returns and ErrNotSupported{Method: "MethodName"} for the
// error return. Example: a method returning (*[]db.Routes, error) returns
// (nil, ErrNotSupported{Method: "ListRoutes"}).
```

**Driver routing in `cmd/connect.go`** (NOT in `db.ConfigDB`):

```go
// After loading config:
var driver db.DbDriver
if cfg.Remote_URL != "" {
    driver, err = remote.NewDriver(cfg.Remote_URL, cfg.Remote_API_Key)
    if err != nil {
        return fmt.Errorf("remote connection failed: %w", err)
    }
    // Set singleton so the 83 db.ConfigDB(*cfg) calls in TUI code return this driver
    db.SetInstance(driver)
} else {
    driver = db.ConfigDB(*cfg)
}
model, _ := tui.InitialModel(&verbose, cfg, driver, utility.DefaultLogger, nil, mgr, nil, nil)
tui.CliRun(&model)
```

**Done when:**
1. `go build ./...` succeeds — RemoteDriver satisfies the `db.DbDriver` interface at compile time
2. `just test` passes (existing tests unchanged)
3. Create `test-remote-config.json` with `{"remote_url": "http://localhost:8080", "remote_api_key": "test"}`
4. `modula connect add remote-test ./test-remote-config.json`
5. Start a local server with `just run` in another terminal
6. `modula connect remote-test` -> TUI launches (operations will fail with ErrNotSupported, but the TUI opens)
7. Clean up: `modula connect remove remote-test`, delete `test-remote-config.json`

---

### Phase 2b: Conversion layer (pure functions, no driver methods)

**Scope:** 5 new files containing ~70 conversion functions + 1 test file. No driver methods are implemented yet — this phase produces pure functions that are tested in isolation.

**New files:**

| File | Entity types covered |
|------|---------------------|
| `internal/remote/convert_content.go` | ContentData, ContentFields, ContentRelations, ContentVersions |
| `internal/remote/convert_admin.go` | AdminContentData, AdminContentFields, AdminDatatypes, AdminFields, AdminFieldTypes, AdminRoutes |
| `internal/remote/convert_core.go` | Datatypes, Fields, FieldTypes, Routes, Media, MediaDimensions |
| `internal/remote/convert_users.go` | Users, Roles, Permissions, RolePermissions, Tokens, Sessions, UserOauths, UserSshKeys |
| `internal/remote/convert_system.go` | Backups, Locales, Webhooks, Plugins, Tables |

Each file contains:
- `entityToDb(sdk *modula.Entity) db.Entity` — SDK response -> db type (for reads)
- `entityFromDb(params db.CreateEntityParams) modula.CreateEntityParams` — db params -> SDK request (for mutations)
- Only for entity types the TUI actually uses (see TUI Method Audit above)

**SDK type -> db type mapping** (reference `sdks/go/types.go` and `internal/db/db.go`):

| SDK type (`modula.`) | db type (`db.`) | Direction | File |
|----------------------|-----------------|-----------|------|
| `Route` | `Routes` | bidirectional | convert_core.go |
| `ContentData` | `ContentData` | bidirectional | convert_content.go |
| `ContentField` | `ContentFields` | bidirectional | convert_content.go |
| `ContentVersion` | `ContentVersions` | read-only | convert_content.go |
| `Datatype` | `Datatypes` | bidirectional | convert_core.go |
| `Field` | `Fields` | bidirectional | convert_core.go |
| `FieldType` | `FieldTypes` | bidirectional | convert_core.go |
| `Media` | `Media` | read-only (delete only, no create via SDK) | convert_core.go |
| `User` | `Users` | bidirectional | convert_users.go |
| `Role` | `Roles` | read-only | convert_users.go |
| `Permission` | `Permissions` | read-only | convert_users.go |
| `Token` | `Tokens` | write-only (create) | convert_users.go |
| `UserSshKey` | `UserSshKeys` | write-only (create) | convert_users.go |
| `AdminRoute` | `AdminRoutes` | bidirectional | convert_admin.go |
| `AdminDatatype` | `AdminDatatypes` | bidirectional | convert_admin.go |
| `AdminField` | `AdminFields` | bidirectional | convert_admin.go |
| `AdminFieldType` | `AdminFieldTypes` | bidirectional | convert_admin.go |
| `AdminContentData` | `AdminContentData` | bidirectional | convert_admin.go |
| `Locale` | `Locales` | read-only | convert_system.go |
| `Webhook` | `Webhooks` | read-only | convert_system.go |
| `Backup` | `Backups` | write-only | convert_system.go |
| `Table` | `Tables` | read-only | convert_system.go |

Note: SDK type names use singular (`Route`), db type names often use plural (`Routes`). Check exact names in `sdks/go/types.go` and the `DbDriver` method return types in `internal/db/db.go`.

**Nullable handling rules (enforced in all conversion functions, no exceptions):**

| SDK type | db type | nil/zero handling |
|----------|---------|-------------------|
| `*string` | `sql.NullString` | nil -> `{Valid: false}`, non-nil -> `{String: *v, Valid: true}` |
| `*ContentID` | `types.NullableContentID` | nil -> `{Valid: false}`, non-nil -> `{ID: ContentID(*v), Valid: true}` |
| `*UserID` | `types.NullableUserID` | same pattern |
| `*time.Time` | `types.Timestamp` | nil -> zero `Timestamp`, non-nil -> `types.NewTimestamp(*v)` |
| `*int64` | `types.NullableInt64` | nil -> `{Valid: false}`, non-nil -> `{Int64: *v, Valid: true}` |

**Test file:** `internal/remote/convert_test.go`

Test strategy: table-driven tests for each conversion function. For each entity type:
- Round-trip test: create SDK entity with all fields populated -> `toDb()` -> `fromDb()` -> compare with original
- Null test: create SDK entity with all nullable fields nil -> `toDb()` -> verify all `Valid: false`
- This catches silent data corruption from incorrect nullable handling

**Done when:**
1. `go build ./...` succeeds
2. `go test ./internal/remote/...` passes — all conversion round-trip and null tests pass
3. `just test` passes (existing tests unchanged)
4. Every conversion function used by the 83 TUI-used methods exists (cross-reference against TUI Method Audit)

---

### Phase 2c: RemoteDriver read methods

**Scope:** Modify `internal/remote/driver.go` only. Replace stubs with real implementations for all read-only methods the TUI calls. Uses conversion functions from Phase 2b.

**Methods to implement (all List*, Get*, Count* from the TUI audit):**

| Category | Methods |
|----------|---------|
| Routes | `ListRoutes`, `ListRoutesByDatatype`, `GetRoute`, `GetRouteID`, `GetRouteTreeByRouteID`, `GetContentTreeByRoute` |
| Datatypes | `ListDatatypes`, `ListDatatypesRoot`, `GetDatatype` |
| Fields | `ListFieldsByDatatypeID`, `ListFieldTypes`, `GetField`, `GetFieldType`, `GetMaxSortOrderByParentID` |
| ContentData | `ListContentDataTopLevelPaginated`, `GetContentData` |
| ContentFields | `ListContentFieldsByContentData`, `ListContentFieldsByContentDataAndLocale`, `GetContentField` |
| ContentVersions | `ListContentVersionsByContent` |
| Locales | `ListEnabledLocales` |
| Users | `ListUsers`, `ListUsersWithRoleLabel`, `GetUser`, `GetUserByEmail` |
| UserSshKeys | `GetUserSshKeyByFingerprint` |
| Roles | `ListRoles` |
| Media | `ListMedia` |
| Tables | `ListTables` |
| Webhooks | `ListWebhooks` |
| Admin | `ListAdminRoutes`, `ListAdminDatatypes`, `ListAdminFieldsByDatatypeID`, `ListAdminContentDataTopLevelPaginated`, `ListAdminFieldTypes`, `ListFieldTypes` |
| Admin Get | `GetAdminRoute`, `GetAdminDatatypeById`, `GetAdminField`, `GetAdminFieldType` |

**Pattern for each method:**

```go
func (d *RemoteDriver) ListRoutes() (*[]db.Routes, error) {
    routes, err := d.client.Routes.List(context.Background())
    if err != nil {
        return nil, fmt.Errorf("remote: ListRoutes: %w", err)
    }
    result := make([]db.Routes, len(routes))
    for i, r := range routes {
        result[i] = routeToDb(&r)
    }
    return &result, nil
}
```

All errors wrapped with `fmt.Errorf("remote: %s: %w", methodName, err)` for debuggability.

**Done when:**
1. `go build ./...` succeeds
2. `just test` passes
3. Start local server (`just run`), register remote config, run `modula connect remote-test`
4. Navigate to Routes screen -> routes load and display
5. Navigate to Datatypes screen -> datatypes load and display
6. Navigate to Content screen -> content list loads
7. Navigate to Users screen -> users load and display
8. Navigate to Media screen -> media list loads
9. All navigation screens show data from the remote server without errors

---

### Phase 2d: RemoteDriver write methods + TUI guards

**Scope:** Implement Create/Update/Delete methods in `driver.go`. Add `IsRemote` flag and guards to TUI files. This is the final phase that makes remote mode fully functional.

**Modified files:**

| File | Change |
|------|--------|
| `internal/remote/driver.go` | Implement Create*, Update*, Delete* methods for all TUI-used entities |
| `internal/tui/model.go` | Add `IsRemote bool` field to Model |
| `internal/tui/actions.go` | Add `ActionsMenuForMode(isRemote bool)`, filter menu in remote mode |
| `internal/tui/commands.go` | Guard 6 `GetConnection()` callers; add remote media upload path |
| `internal/tui/fields.go` | Graceful degradation when `GetConnection()` returns error |
| `cmd/connect.go` | Set `model.IsRemote = cfg.Remote_URL != ""` after creating model |

**Write methods to implement:**

| Category | Methods |
|----------|---------|
| ContentData | `CreateContentData`, `UpdateContentData`, `DeleteContentData` |
| ContentFields | `CreateContentField`, `UpdateContentField`, `DeleteContentField` |
| Datatypes | `CreateDatatype`, `UpdateDatatype`, `DeleteDatatype` |
| Fields | `CreateField`, `UpdateField`, `DeleteField`, `UpdateFieldSortOrder` |
| FieldTypes | `CreateFieldType`, `UpdateFieldType`, `DeleteFieldType` |
| Routes | `CreateRoute`, `UpdateRoute`, `DeleteRoute` |
| Media | `DeleteMedia` |
| Users | `CreateUser`, `UpdateUser`, `DeleteUser` |
| UserSshKeys | `CreateUserSshKey` |
| Tokens | `CreateToken` |
| Backups | `CreateBackup`, `UpdateBackupStatus` |
| Admin entities | `CreateAdminRoute`, `UpdateAdminRoute`, `DeleteAdminRoute`, `CreateAdminDatatype`, `UpdateAdminDatatype`, `DeleteAdminDatatype`, `CreateAdminField`, `UpdateAdminField`, `DeleteAdminField`, `CreateAdminFieldType`, `UpdateAdminFieldType`, `DeleteAdminFieldType`, `DeleteAdminContentData` |

**Audited methods pattern** (methods that take `audited.AuditContext`):

```go
func (d *RemoteDriver) CreateContentData(ctx context.Context, ac audited.AuditContext, params db.CreateContentDataParams) (*db.ContentData, error) {
    // ac is discarded — server creates audit records from Bearer token context
    sdkParams := contentDataFromDb(params)
    result, err := d.client.ContentData.Create(ctx, sdkParams)
    if err != nil {
        return nil, fmt.Errorf("remote: CreateContentData: %w", err)
    }
    row := contentDataToDb(result)
    return &row, nil
}
```

**TUI guards:**

```go
// model.go — add field
type Model struct {
    // ... existing fields
    IsRemote bool // true when connected to a remote server via Go SDK
}

// actions.go — filter menu
func ActionsMenuForMode(isRemote bool) []ActionItem {
    all := ActionsMenu()
    if !isRemote {
        return all
    }
    // Remote-safe: Check for Updates (idx 6), Validate Config (idx 7), Generate API Token (idx 8)
    return []ActionItem{all[6], all[7], all[8]}
}

// commands.go — guard GetConnection callers
// The 6 Database* functions (Insert/Update/Get/List/FilteredList/Delete)
// already call GetConnection() which returns ErrRemoteMode.
// The TUI's Table Browser screen must be hidden when IsRemote is true.
// Add a check in the navigation code that skips the Tables page.

// fields.go — graceful degradation
// GetColumnRowsString calls GetConnection(). When it returns error,
// return empty string slice (no autocomplete suggestions).

// commands.go — remote media upload
// In HandleMediaUpload, check m.IsRemote:
//   true  -> open file, call d.client.MediaUpload.Upload(ctx, file, filename)
//   false -> existing local pipeline (media.HandleMediaUpload)
```

**Done when:**
1. `go build ./...` succeeds
2. `just test` passes
3. Start local server, register remote config, `modula connect remote-test`
4. Create a new route -> appears on remote server
5. Create a new datatype -> appears on remote server
6. Create content under a route -> appears on remote server
7. Edit content fields -> changes saved on remote server
8. Delete content -> removed on remote server
9. Upload a local file as media -> server receives, optimizes, stores
10. Actions menu shows only "Check for Updates", "Validate Config", and "Generate API Token"
11. Table Browser is not accessible in navigation
12. `fields.go` autocomplete returns empty list (no crash)

---

### Phase 3a: Connection health indicator

**Scope:** 3 modified files. Adds a status indicator to the TUI status bar showing remote connection state.

**Modified files:**

| File | Change |
|------|--------|
| `internal/remote/driver.go` | Add `RemoteStatus` tracking (update on every SDK call success/failure) |
| `internal/tui/model.go` | Add `RemoteStatus` field, expose via accessor |
| `internal/tui/view.go` (or equivalent status bar renderer) | Render `[remote: cms.example.com]` or `[remote: disconnected]` |

**RemoteStatus values:** `Connected`, `Disconnected`, `Unknown`

**Update logic:** After every SDK call in RemoteDriver, update status. On success -> `Connected`. On network error (timeout, connection refused, 5xx) -> `Disconnected`. On 4xx -> keep current status (4xx is not a connectivity issue).

**Done when:**
1. `go build ./...` succeeds
2. Connect to remote -> status bar shows `[remote: localhost:8080]`
3. Kill remote server -> next operation shows error, status bar changes to `[remote: disconnected]`
4. Restart server -> next successful operation changes status back to `[remote: localhost:8080]`

---

### Phase 3b: Read retry logic

**Scope:** 1 new file, 1 modified file. Adds automatic single-retry for read operations on transient failures.

**New files:**

| File | Purpose |
|------|---------|
| `internal/remote/retry.go` | `retryOnce()` helper function |

**Modified files:**

| File | Change |
|------|--------|
| `internal/remote/driver.go` | Wrap all read methods (List*, Get*, Count*) with `retryOnce()` |

**Retry rules:**
- Retry on: timeout, 502, 503, 504
- No retry on: 4xx (client error), connection refused (server down), unknown errors
- 1s delay between attempts
- Maximum 1 retry (2 total attempts)

```go
// retryRead wraps a read operation with a single retry on transient failures.
// Uses Go generics so it works with any return type (*[]db.Routes, *db.ContentData, etc.).
func retryRead[T any](fn func() (T, error)) (T, error) {
    result, err := fn()
    if err == nil || !isRetryable(err) {
        return result, err
    }
    time.Sleep(1 * time.Second)
    return fn()
}

// isRetryable checks if an error is a transient failure worth retrying.
// The Go SDK wraps HTTP errors as *modula.ApiError (see sdks/go/errors.go).
func isRetryable(err error) bool {
    // Check for Go SDK API errors with retryable status codes
    var apiErr *modula.ApiError
    if errors.As(err, &apiErr) {
        switch apiErr.StatusCode {
        case 502, 503, 504:
            return true
        }
        return false
    }
    // Check for network timeouts (net.Error interface)
    var netErr net.Error
    if errors.As(err, &netErr) && netErr.Timeout() {
        return true
    }
    // Check for context deadline exceeded
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    return false
}
```

**Usage in driver.go read methods:**

```go
func (d *RemoteDriver) ListRoutes() (*[]db.Routes, error) {
    return retryRead(func() (*[]db.Routes, error) {
        routes, err := d.client.Routes.List(context.Background())
        if err != nil {
            return nil, fmt.Errorf("remote: ListRoutes: %w", err)
        }
        result := make([]db.Routes, len(routes))
        for i, r := range routes {
            result[i] = routeToDb(&r)
        }
        return &result, nil
    })
}
```

**Done when:**
1. `go build ./...` succeeds
2. `go test ./internal/remote/...` passes (add test for retryOnce with mock errors)
3. Manual test: connect to remote, briefly kill and restart server during a list operation -> data loads on retry without user-visible error

---

### Phase 3c: Remaining RemoteDriver methods

**Scope:** `internal/remote/driver.go` only. Implement additional DbDriver methods beyond the 83 TUI-used ones that have REST API equivalents.

**Target:** ~40 additional methods that the SDK covers but the TUI doesn't currently call. These provide coverage for future TUI screens or direct SDK usage.

**Exclusions (remain as `ErrNotSupported` permanently):**
- `CreateAllTables`, `DropAllTables`, `CreateBootstrapData`, `ValidateBootstrapData` — server-internal DDL
- `DumpSql` — requires raw SQL connection
- `GetConnection`, `Ping`, `Query`, `ExecuteQuery` — raw SQL access
- `GetForeignKeys`, `ScanForeignKeyQueryRows`, `SelectColumnFromTable`, `SortTables` — schema introspection
- All `Create*Table` methods — DDL operations
- All `Drop*Table` methods — DDL operations

**Done when:**
1. `go build ./...` succeeds
2. `go test ./internal/remote/...` passes
3. Grep for `ErrNotSupported` in `driver.go` -> only the excluded methods above remain as stubs

---

### Phase 3d: Auto-detect config

**Scope:** `cmd/connect.go` only.

When `modula connect` is run with no name and no default is set in the registry:
1. Check `./config.json` in the current working directory
2. If found, use it directly (no registry entry needed)
3. If not found, error: "no project specified and no config.json in current directory"

**Done when:**
1. `go build ./...` succeeds
2. `cd` to project root (which has `config.json`), run `modula connect` -> TUI opens using local config
3. `cd /tmp`, run `modula connect` -> error message about no project and no config.json
4. Existing registry-based commands still work (no regression)

---

### Phase 3e: Upload progress bar

**Scope:** 2 modified files. Adds progress feedback for media uploads in remote mode.

**Modified files:**

| File | Change |
|------|--------|
| `sdks/go/media_upload.go` | Add `UploadWithProgress(ctx, file, filename, progressFn)` method |
| `internal/tui/commands.go` | Use `UploadWithProgress` in remote mode, emit progress messages to TUI |

**Done when:**
1. `go build ./...` succeeds
2. Upload a large file (>1MB) in remote mode -> progress bar renders in TUI
3. Upload completes -> progress bar disappears, success message shown

## Key Files Reference

| File | Role |
|------|------|
| `internal/db/init.go:211` | `ConfigDB()` — fallback driver creation (83 call sites from TUI); add `SetInstance()` in Phase 2a |
| `internal/db/db.go` | `DbDriver` interface definition (~330 methods, 36 entity types) |
| `internal/tui/commands.go` | 36 `db.ConfigDB()` calls in tea.Cmd closures + 6 `GetConnection()` calls |
| `internal/tui/actions.go` | 5 admin-only methods (DropAllTables, CreateAllTables, etc.) |
| `internal/tui/update_dialog.go` | 9 `db.ConfigDB()` calls |
| `internal/tui/admin_update_dialog.go` | 16 `db.ConfigDB()` calls |
| `internal/tui/update_fetch.go` | List/fetch commands for all entity screens |
| `internal/tui/fields.go` | 1 `GetConnection()` call (autocomplete suggestions) |
| `sdks/go/modula.go` | SDK Client + NewClient (15 standard resources, 20+ specialized) |
| `sdks/go/types.go` | SDK entity types (conversion reference) |
| `sdks/go/media_upload.go` | SDK media upload (multipart over HTTP) |
| `sdks/go/content_tree.go` | SDK content tree save (bulk pointer updates) |
| `internal/config/config.go` | Config struct (add Remote_URL, Remote_API_Key) |
| `internal/router/media.go:184` | Server-side media optimization pipeline (same as local TUI) |
| `internal/router/mux.go` | All REST API route registrations |
| `sdks/go/errors.go` | `ApiError` struct with `StatusCode int` — used by Phase 3b `isRetryable()` |
| `sdks/go/health.go` | `HealthResource.Check(ctx) (*HealthResponse, error)` — connectivity check |

## Verification Summary

Each phase has a "Done when" checklist inline with its description above. The gate for every phase is:

1. `go build ./...` succeeds (compile check)
2. `just test` passes (no regressions)
3. Phase-specific acceptance criteria (listed per phase)

**Phase dependency chain:**

```
Phase 1 (Registry)
    |
Phase 1.5 (API gaps)
    |
Phase 2a (RemoteDriver skeleton)
    |
Phase 2b (Conversion layer)  -- can run in parallel with 2a if interface is stable
    |
Phase 2c (Read methods)      -- depends on 2a + 2b
    |
Phase 2d (Write methods + TUI guards) -- depends on 2c
    |
    +-- Phase 3a (Health indicator)  \
    +-- Phase 3b (Read retry)         |-- all independent, any order
    +-- Phase 3c (Remaining methods)  |
    +-- Phase 3d (Auto-detect config) |
    +-- Phase 3e (Upload progress)   /
```

**Agent assignment notes:**
- Phases 1 and 1.5 can be done by separate agents (different packages, no overlap)
- Phase 2a must complete before 2c/2d (they need the struct to exist)
- Phase 2b can run in parallel with 2a (pure functions, no dependencies on driver.go beyond types)
- Phase 2c must complete before 2d (reads prove the conversion layer works before mutations use it)
- All Phase 3 sub-phases are independent and can run in parallel
