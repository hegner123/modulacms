# Phase 5: PluginService + WebhookService + LocaleService

**Parent:** [SERVICE_LAYER_ROADMAP.md](SERVICE_LAYER_ROADMAP.md)
**Assumes:** Phases 0–3 complete. Phase 4 (Users + RBAC) may or may not be complete — no dependency.

---

## Context

Three independent domains with moderate business logic beyond simple CRUD. Currently:

- **Plugins** — lifecycle management (enable/disable/reload) delegated to `plugin.Manager`; route and hook approval delegated to `plugin.HTTPBridge`; admin panel handlers are placeholders; API handlers live inside `internal/plugin/` as `http.Handler` factories taking `*Manager`.
- **Webhooks** — full CRUD with SSRF validation, secret generation, test dispatch, delivery tracking with exponential backoff retry, and old delivery pruning. API handlers use `config.Config` pattern. Admin handlers use `(driver, mgr)` closures.
- **Locales** — CRUD with BCP 47 validation, fallback chain cycle detection, default locale atomicity, enabled/disabled filtering, and translation creation (clone translatable fields for a new locale). API handlers use `config.Config` pattern. Admin handlers use `(driver, mgr)` closures.

All three are independent — they can be built in parallel by separate agents.

---

## Key Design Decisions

### PluginService wraps plugin.Manager, does not replace it

The `plugin.Manager` is a 2200+ line subsystem managing Lua VM pools, circuit breakers, hook engines, request engines, pipeline registries, file watchers, and coordinators. Absorbing this into the service layer would be a rewrite, not an extraction. Instead, `PluginService` is a thin facade that:
1. Delegates lifecycle operations (enable/disable/reload/install) to Manager methods
2. Delegates route/hook/request approval to Manager.Bridge() and Manager.HookEngine()
3. Adds service-layer error mapping (Manager errors → `NotFoundError`/`ValidationError`/`InternalError`)
4. Provides a unified interface for all consumers (admin panel, API, future MCP direct calls)

The existing `http.Handler` factories in `internal/plugin/cli_commands.go`, `hook_handlers.go`, and `request_handlers.go` remain unchanged in Phase 5. They are already well-factored — each takes `*Manager` and returns `http.Handler`. The service layer wraps these same Manager calls but returns domain types instead of writing HTTP responses. Rewiring the bridge's `MountAdminEndpoints` to call service methods (instead of Manager directly) is optional in Phase 5 — the handlers inside `internal/plugin/` already achieve the same goal.

### PluginService does NOT touch HTTPBridge mount or plugin route serving

The `HTTPBridge.MountOn()` and `ServeHTTP()` are infrastructure-level concerns (dynamic route registration, per-request rate limiting, Lua VM dispatch). These stay as-is. The service wraps only the management API: list, info, enable, disable, reload, cleanup, route/hook/request approval.

### Admin panel plugin handlers gain real data

The current `PluginsListHandler` and `PluginDetailHandler` are placeholders that render static templates with no data. After Phase 5, they call `svc.Plugins.List()` and `svc.Plugins.Get(name)` to render actual plugin state, health, and VM pool info.

### WebhookService encapsulates SSRF validation and secret generation

Currently handlers inline `webhooks.ValidateWebhookURL()`, secret generation (`crypto/rand` + `hex.EncodeToString`), event parsing, and header marshaling. The service absorbs all of this. Handlers become: parse input → call service → format response.

### WebhookService does NOT absorb the Dispatcher

The `webhooks.Dispatcher` is a long-running goroutine pool with channels, retry tickers, and graceful shutdown. It implements the `publishing.WebhookDispatcher` interface and is constructed at startup in `cmd/serve.go`. The service holds a reference to it for test dispatch, but does not manage its lifecycle. `Dispatch()` calls from publishing remain unchanged — they call `dispatcher.Dispatch()` directly, not through the service.

### LocaleService absorbs translation creation

The `TranslationHandler` and `AdminTranslationHandler` in `internal/router/translations.go` contain significant business logic (locale validation, field enumeration, default-locale value copying, duplicate detection). This logic moves into `LocaleService.CreateTranslation()` and `LocaleService.CreateAdminTranslation()`. The handlers become thin dispatchers.

### LocaleService absorbs locale resolution

`ResolveLocale()` and `WalkFallback()` in `internal/router/locale_resolve.go` contain business logic (Accept-Language parsing, fallback chain walking, locale metadata building). These move to `LocaleService` methods. Content delivery handlers call the service instead of the standalone functions.

### Services are separate structs on Registry

Matching the established pattern (ContentService, MediaService, RouteService):

```go
type Registry struct {
    // ... existing fields ...
    Plugins  *PluginService
    Webhooks *WebhookService
    Locales  *LocaleService
}
```

### MCP stays as-is

MCP tools call the Go SDK over HTTP. Rewiring to call services directly is Phase 7.

---

## PluginService — internal/service/plugins.go

### Struct

```go
type PluginService struct {
    mgr PluginManager
}

func NewPluginService(mgr PluginManager) *PluginService
```

### PluginManager Interface

Consumer-defined interface accepting only the methods PluginService actually calls. The `plugin.Manager` satisfies it implicitly.

**Coupling note:** The `Bridge()`, `HookEngine()`, and `RequestEngine()` methods return concrete pointer types from the `plugin` package, which means the service package imports `internal/plugin`. This is a pragmatic tradeoff — defining narrower interfaces for each sub-component (route listing, hook approval, request approval) would mean 3 additional interfaces and adapters for no real benefit in a codebase with a single implementation. The cost is that mock `PluginManager` implementations in tests must also construct or mock these sub-components. In practice, tests mock the top-level `PluginManager` and provide nil for sub-components (the service nil-guards all sub-component access).

```go
type PluginManager interface {
    // Lifecycle
    InstallPlugin(ctx context.Context, name string) (*db.Plugin, error)
    EnablePlugin(ctx context.Context, name string) error
    DisablePlugin(ctx context.Context, name string) error
    ReloadPlugin(ctx context.Context, name string) error
    ActivatePlugin(ctx context.Context, name string, adminUser string) error
    DeactivatePlugin(ctx context.Context, name string) error

    // Queries
    GetPlugin(name string) *plugin.PluginInstance
    ListPlugins() []*plugin.PluginInstance
    GetPluginState(name string) (plugin.PluginState, bool)
    PluginHealth() plugin.PluginHealthStatus

    // Discovery & Registry
    ListDiscovered() []string
    SyncCapabilities(ctx context.Context, name string, adminUser string) error
    ListAllPipelines() (*[]db.Pipeline, error)
    DryRunPipeline(table, op string) []plugin.DryRunResult
    DryRunAllPipelines() []plugin.DryRunResult

    // Cleanup
    ListOrphanedTables(ctx context.Context) ([]string, error)
    DropOrphanedTables(ctx context.Context, requestedTables []string) ([]string, error)

    // Sub-components
    Bridge() *plugin.HTTPBridge
    HookEngine() *plugin.HookEngine
    RequestEngine() *plugin.RequestEngine
}
```

### Methods

| Method | Signature | Logic |
|--------|-----------|-------|
| List | `(ctx) ([]PluginSummary, error)` | `mgr.ListPlugins()` → map to summary structs (name, version, description, state, circuit breaker). If Manager is nil → return empty slice, no error |
| Get | `(ctx, name string) (*PluginDetail, error)` | `mgr.GetPlugin(name)` → NotFoundError if nil. Maps full instance state to detail struct |
| Enable | `(ctx, name string) error` | Validate name non-empty. `mgr.EnablePlugin(ctx, name)` → map errors |
| Disable | `(ctx, name string) error` | Validate name non-empty. `mgr.DisablePlugin(ctx, name)` → map errors |
| Reload | `(ctx, name string) error` | Validate name non-empty. `mgr.ReloadPlugin(ctx, name)` → map errors |
| Install | `(ctx, name string) (*db.Plugin, error)` | `mgr.InstallPlugin(ctx, name)` → map errors |
| SyncCapabilities | `(ctx, name, adminUser string) error` | `mgr.SyncCapabilities(ctx, name, adminUser)` → map errors |
| Health | `(ctx) (*plugin.PluginHealthStatus, error)` | `mgr.PluginHealth()` — return pointer, nil error |
| CleanupDryRun | `(ctx) ([]string, error)` | `mgr.ListOrphanedTables(ctx)` |
| CleanupDrop | `(ctx, tables []string) ([]string, error)` | Validate tables non-empty. `mgr.DropOrphanedTables(ctx, tables)` |
| ListRoutes | `(ctx) ([]plugin.RouteRegistration, error)` | `mgr.Bridge().ListRoutes()`. If Bridge is nil → empty slice |
| ApproveRoutes | `(ctx, routes []RouteApprovalInput, approvedBy string) error` | Loop: `mgr.Bridge().ApproveRoute(ctx, r.Plugin, r.Method, r.Path, approvedBy)`. Collect errors, return first |
| RevokeRoutes | `(ctx, routes []RouteApprovalInput) error` | Loop: `mgr.Bridge().RevokeRoute(ctx, r.Plugin, r.Method, r.Path)` |
| ListHooks | `(ctx) ([]HookInfo, error)` | `mgr.HookEngine().ListHooks()` → map to service types |
| ApproveHooks | `(ctx, hooks []HookApprovalInput, approvedBy string) error` | Loop: `mgr.HookEngine().ApproveHook(ctx, h.PluginName, h.Event, h.Table, approvedBy)`. Collect errors, return first |
| RevokeHooks | `(ctx, hooks []HookApprovalInput) error` | Loop: `mgr.HookEngine().RevokeHook(ctx, h.PluginName, h.Event, h.Table)` |
| ListRequests | `(ctx) ([]RequestInfo, error)` | `mgr.RequestEngine().ListRequests(ctx)` → map to service types. If RequestEngine is nil → empty slice |
| ApproveRequests | `(ctx, requests []RequestApprovalInput, approvedBy string) error` | Loop: `mgr.RequestEngine().ApproveRequest(ctx, r.PluginName, r.Domain, approvedBy)` |
| RevokeRequests | `(ctx, requests []RequestApprovalInput) error` | Loop: `mgr.RequestEngine().RevokeRequest(ctx, r.PluginName, r.Domain)` |
| ListPipelines | `(ctx) ([]db.Pipeline, error)` | `mgr.ListAllPipelines()` |
| DryRunPipelines | `(ctx) ([]plugin.DryRunResult, error)` | `mgr.DryRunAllPipelines()` |

### Service-Level Types

```go
type PluginSummary struct {
    Name               string
    Version            string
    Description        string
    State              string
    CircuitBreakerState string
}

type PluginDetail struct {
    PluginSummary
    Author       string
    License      string
    FailedReason string
    VMsAvailable int
    VMsTotal     int
    Dependencies []string
    SchemaDrift  bool
}

type RouteApprovalInput struct {
    Plugin string
    Method string
    Path   string
}

type HookApprovalInput struct {
    PluginName string
    Event      string
    Table      string
}

type HookInfo struct {
    PluginName string
    Event      string
    Table      string
    Priority   int
    Approved   bool
    IsWildcard bool
}

type RequestApprovalInput struct {
    PluginName string
    Domain     string
}

type RequestInfo struct {
    PluginName string
    Domain     string
    Approved   bool
}
```

### Error Mapping

Manager methods return plain `error` values (often `fmt.Errorf`). The service maps these:
- "plugin not found" / "not installed" → `*NotFoundError{Resource: "plugin", ID: name}`
- "already enabled" / "already disabled" → `*ConflictError{Resource: "plugin", Detail: ...}`
- All other errors → `*InternalError{Err: err}`

Manager nil check: if `PluginManager` is nil (plugins not enabled), all methods return empty results and no error for reads, or `*ValidationError{field: "plugins", message: "plugin system is not enabled"}` for mutations. This handles the `bridge != nil` guard currently in mux.go.

---

## WebhookService — internal/service/webhooks.go

### Struct

```go
type WebhookService struct {
    driver     db.DbDriver
    mgr        *config.Manager
    dispatcher publishing.WebhookDispatcher
}

func NewWebhookService(driver db.DbDriver, mgr *config.Manager, dispatcher publishing.WebhookDispatcher) *WebhookService
```

### Param Types (internal/service/webhooks.go)

```go
type CreateWebhookInput struct {
    Name     string
    URL      string
    Secret   string            // empty → auto-generate
    Events   []string
    IsActive bool
    Headers  map[string]string
}

type UpdateWebhookInput struct {
    WebhookID types.WebhookID
    Name      string
    URL       string
    Secret    string
    Events    []string
    IsActive  bool
    Headers   map[string]string
}

type WebhookTestResult struct {
    Status     string // "success" or "failed"
    StatusCode int
    Error      string
    Duration   string
}

type DeliveryRetryResult struct {
    DeliveryID types.WebhookDeliveryID
    Status     string
}
```

### Methods

| Method | Signature | Logic Beyond DbDriver |
|--------|-----------|----------------------|
| CreateWebhook | `(ctx, ac, CreateWebhookInput) (*db.Webhook, error)` | Validate name non-empty, URL non-empty, SSRF check via `webhooks.ValidateWebhookURL(url, allowHTTP)`, validate events non-empty. Auto-generate secret if empty (`crypto/rand` 32 bytes → hex). Marshal events to JSON, headers to JSON. Set timestamps UTC. |
| UpdateWebhook | `(ctx, ac, UpdateWebhookInput) (*db.Webhook, error)` | Fetch existing → NotFoundError if nil. SSRF check on URL if changed. Validate events non-empty. Marshal events/headers. Preserve immutable fields (DateCreated, AuthorID). Update + re-fetch. |
| DeleteWebhook | `(ctx, ac, WebhookID) error` | Validate ID. Delete. |
| GetWebhook | `(ctx, WebhookID) (*db.Webhook, error)` | NotFoundError if nil |
| ListWebhooks | `(ctx) ([]db.Webhook, error)` | Passthrough |
| ListWebhooksPaginated | `(ctx, PaginationParams) ([]db.Webhook, int64, error)` | Paginate + count |
| TestWebhook | `(ctx, WebhookID) (*WebhookTestResult, error)` | Fetch webhook → NotFoundError if nil. SSRF re-check URL (DNS rebinding defense). Build `webhooks.Payload{Event: "webhook.test", ...}` — refer to the existing `WebhookTestHandler` in `internal/router/webhooks.go` for exact payload construction. Sign with `webhooks.Sign(secret, payloadBytes)`. Execute synchronous HTTP POST with `cfg.WebhookTimeout()` (default 10s). Return status, code, duration. Does NOT create a delivery record. |
| ListDeliveries | `(ctx, WebhookID) ([]db.WebhookDelivery, error)` | Passthrough |
| RetryDelivery | `(ctx, DeliveryID) (*DeliveryRetryResult, error)` | Fetch delivery → NotFoundError if nil. Fetch parent webhook → NotFoundError if nil. Re-dispatch via `dispatcher.Dispatch()` (fire-and-forget, async). Return `DeliveryRetryResult{Status: "queued"}` — the retry is enqueued, not completed. The actual delivery result is tracked asynchronously via the delivery record's status field. |
| PruneDeliveries | `(ctx, olderThan time.Time) error` | `driver.PruneOldDeliveries(ctx, types.Timestamp(olderThan))` |

### SSRF Validation

The service calls `webhooks.ValidateWebhookURL(url, allowHTTP)` where `allowHTTP` comes from `svc.mgr.Config().WebhookAllowHTTP()`. This check runs at:
1. Create time — prevents registering malicious URLs
2. Update time — re-validates on URL change
3. Test time — defends against DNS rebinding (URL resolved again)

Delivery-time SSRF validation stays in the Dispatcher (it already does this). The service adds no new validation points.

### Secret Generation

Moved from handler to service. If `CreateWebhookInput.Secret` is empty:
```go
secretBytes := make([]byte, 32)
io.ReadFull(crypto/rand.Reader, secretBytes)
secret = hex.EncodeToString(secretBytes)
```

### Event Validation

The service validates that `Events` is non-empty but does not validate event names against a whitelist. Unknown events simply never match — this is the current behavior and allows forward compatibility (new events don't require schema changes).

---

## LocaleService — internal/service/locales.go

### Struct

```go
type LocaleService struct {
    driver db.DbDriver
    mgr    *config.Manager
}

func NewLocaleService(driver db.DbDriver, mgr *config.Manager) *LocaleService
```

### Param Types

```go
type CreateLocaleInput struct {
    Code         string
    Label        string
    IsDefault    bool
    IsEnabled    bool
    FallbackCode string
    SortOrder    int64
}

type UpdateLocaleInput struct {
    LocaleID     types.LocaleID
    Code         string
    Label        string
    IsDefault    bool
    IsEnabled    bool
    FallbackCode string
    SortOrder    int64
}

type TranslationResult struct {
    Locale        string
    FieldsCreated int
}

type LocaleVersionInfo struct {
    Published   bool
    PublishedAt string
}

type LocaleMetadata struct {
    Locales          map[string]LocaleVersionInfo
    AvailableLocales []string
}
```

### Methods

| Method | Signature | Logic Beyond DbDriver |
|--------|-----------|----------------------|
| CreateLocale | `(ctx, ac, CreateLocaleInput) (*db.Locale, error)` | Validate code via `validateLocaleCode()` (BCP 47, no regex). Validate label non-empty. Validate fallback chain if FallbackCode non-empty (exists, no self-reference, no cycle — max 2 hops). If IsDefault, call `ClearDefaultLocale()` first. Auto-generate ID, set DateCreated UTC. |
| UpdateLocale | `(ctx, ac, UpdateLocaleInput) (*db.Locale, error)` | Fetch existing → NotFoundError. Validate code + label. Validate fallback chain (exclude self). If toggling IsDefault ON, clear others first. Preserve DateCreated. Update + re-fetch. |
| DeleteLocale | `(ctx, ac, LocaleID) error` | Fetch existing → NotFoundError. Guard: cannot delete default locale → ForbiddenError. Delete. |
| GetLocale | `(ctx, LocaleID) (*db.Locale, error)` | NotFoundError if nil |
| GetLocaleByCode | `(ctx, code string) (*db.Locale, error)` | NotFoundError if nil |
| GetDefaultLocale | `(ctx) (*db.Locale, error)` | NotFoundError if nil |
| ListLocales | `(ctx) ([]db.Locale, error)` | Passthrough (all locales — admin use) |
| ListEnabledLocales | `(ctx) ([]db.Locale, error)` | Passthrough (enabled only — public use) |
| ListLocalesPaginated | `(ctx, PaginationParams) ([]db.Locale, int64, error)` | Paginate + count |
| CreateTranslation | `(ctx, ac, contentDataID ContentID, locale string) (*TranslationResult, error)` | Check i18n enabled → ValidationError if not. Validate locale exists and is enabled. Get content data node → get datatype fields. List existing fields for target locale (skip duplicates). Copy default-locale field values for each translatable field. Create content field rows. Return count. |
| CreateAdminTranslation | `(ctx, ac, adminContentDataID AdminContentID, locale string) (*TranslationResult, error)` | Same logic as CreateTranslation but uses admin types (AdminContentID, AdminFieldID, etc.) |
| ResolveLocale | `(r *http.Request) string` | Calls `s.mgr.Config()` to obtain `cfg` for `I18nEnabled()` and `I18nDefaultLocale()`, and `s.driver` for locale lookups. If config load fails, returns empty string (same as i18n-disabled behavior). Priority: query param `?locale=` → Accept-Language header (via private `parseAcceptLanguage` helper) → config default. Returns resolved locale code or empty string if i18n disabled. NOTE: This method accepts `*http.Request` because it reads headers. This is a pragmatic exception to the "no net/http in services" rule — the alternative (parsing Accept-Language in every handler) creates more duplication than it saves. |
| WalkFallback | `(ctx, contentDataID ContentID, requestedLocale string) (string, *db.ContentVersion, error)` | Try requested locale for published snapshot. Walk fallback chain (max 2 hops). Return resolved locale + snapshot. Note: The current standalone function does not take `ctx`. The service adds it for convention consistency. Driver calls (`GetPublishedSnapshot`, `GetLocaleByCode`) do not currently accept `ctx` — pass the same arguments they currently take. |
| BuildLocaleMetadata | `(ctx, contentDataID ContentID) (*LocaleMetadata, error)` | Check each enabled locale for published snapshot. Build availability map. Same `ctx` convention note as WalkFallback — driver calls don't take `ctx` yet. |

### Validation: Locale Code (BCP 47)

Moved from `internal/router/locales.go` into the service as a private helper:

```go
func validateLocaleCode(code string) error
```

Character-by-character validation (no regex — per CLAUDE.md rules):
- Primary subtag: 2–5 lowercase ASCII letters
- Optional: hyphen + subtag of 2–3 alphanumeric chars
- Examples: "en" ✓, "en-US" ✓, "zh-Hans" ✓, "E" ✗, "en_US" ✗

### Validation: Fallback Chain

Moved from `internal/router/locales.go` into the service as a private helper:

```go
func (s *LocaleService) validateFallbackChain(ctx context.Context, fallbackCode, selfCode string) error
```

Rules:
1. Self-reference forbidden: fallbackCode != selfCode → ValidationError
2. Direct fallback must exist: `GetLocaleByCode(fallbackCode)` → ValidationError if not found
3. Max 2 hops: if fallback's fallback == selfCode → ValidationError (cycle detected)

### ResolveLocale: HTTP Coupling

`ResolveLocale` needs `*http.Request` to read query params and `Accept-Language` header. Two options considered:

**A. Accept `*http.Request` in service** — pragmatic, keeps resolution logic in one place. Violates the "no net/http" convention but this method is specifically about HTTP request locale negotiation.

**B. Accept parsed inputs** — `func ResolveLocale(ctx, queryLocale string, acceptLanguage string) string`. Keeps service pure but pushes parsing to every caller.

**Decision: Option A.** The method is inherently tied to HTTP semantics (content negotiation). Moving it to the service unifies the logic without adding parse-and-pass boilerplate. The config check (`I18nEnabled()`) and DB query (`GetLocaleByCode`) also live in the service.

---

## Handler Rewiring

### Admin Handlers

**internal/admin/handlers/plugins.go** — signatures change from `(driver db.DbDriver)` to `(svc *service.Registry)`:

| Handler | Before | After |
|---------|--------|-------|
| PluginsListHandler | Renders static template, ignores driver | Calls `svc.Plugins.List()`, renders actual plugin data |
| PluginDetailHandler | Renders static template with name only | Calls `svc.Plugins.Get(name)`, renders state/health/VM info |

**internal/admin/handlers/webhooks.go** — signatures change from `(driver db.DbDriver, mgr *config.Manager)` to `(svc *service.Registry)`:

| Handler | Before | After |
|---------|--------|-------|
| WebhookSettingsHandler | `driver.ListWebhooks()` | `svc.Webhooks.ListWebhooks(ctx)` |
| WebhookDetailHandler | `driver.GetWebhook()` + `driver.ListWebhookDeliveriesByWebhook()` | `svc.Webhooks.GetWebhook()` + `svc.Webhooks.ListDeliveries()` |
| WebhookCreateHandler | Inline SSRF check, secret gen, audit ctx, create | Parse form → `svc.Webhooks.CreateWebhook(ctx, ac, input)` |
| WebhookUpdateHandler | Inline SSRF check, overlay, audit ctx, update | Parse form → `svc.Webhooks.UpdateWebhook(ctx, ac, input)` |
| WebhookDeleteHandler | `driver.DeleteWebhook()` | `svc.Webhooks.DeleteWebhook(ctx, ac, id)` |
| WebhookTestHandler | Inline test dispatch (~60 lines) | `svc.Webhooks.TestWebhook(ctx, id)` |

Private helper `renderWebhookTableRows(w, r, driver)` changes to `renderWebhookTableRows(w, r, svc *service.Registry)` — calls `svc.Webhooks.ListWebhooks(ctx)` instead of `driver.ListWebhooks()`.

**internal/admin/handlers/locale_settings.go** — signatures change from `(driver db.DbDriver, mgr *config.Manager)` to `(svc *service.Registry)`:

| Handler | Before | After |
|---------|--------|-------|
| LocaleSettingsHandler | `driver.ListLocales()` + `cfg.I18nEnabled()` | `svc.Locales.ListLocales(ctx)` + `svc.Config()` for i18n check |
| LocaleEditDialogHandler | `driver.GetLocale()` + `driver.ListLocales()` | `svc.Locales.GetLocale()` + `svc.Locales.ListLocales()` |
| LocaleCreateHandler | Inline validation, ClearDefaultLocale, audit ctx | Parse form → `svc.Locales.CreateLocale(ctx, ac, input)` |
| LocaleUpdateHandler | Inline validation, ClearDefaultLocale, audit ctx | Parse form → `svc.Locales.UpdateLocale(ctx, ac, input)` |
| LocaleDeleteHandler | Inline default guard, `driver.DeleteLocale()` | `svc.Locales.DeleteLocale(ctx, ac, id)` (ForbiddenError if default) |

Private helper `renderLocaleTableRows(w, r, driver)` changes to `renderLocaleTableRows(w, r, svc *service.Registry)` — calls `svc.Locales.ListLocales(ctx)` instead of `driver.ListLocales()`.

### API Handlers

**internal/router/webhooks.go** — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`:

| Function | Key Change |
|----------|-----------|
| WebhookListHandler | `db.ConfigDB(c)` → `svc.Webhooks.ListWebhooks(ctx)` |
| WebhookCreateHandler | Remove ~40 lines inline validation/secret gen → `svc.Webhooks.CreateWebhook(ctx, ac, input)` |
| WebhookGetHandler | `svc.Webhooks.GetWebhook(ctx, id)` |
| WebhookUpdateHandler | Remove inline SSRF/overlay → `svc.Webhooks.UpdateWebhook(ctx, ac, input)` |
| WebhookDeleteHandler | `svc.Webhooks.DeleteWebhook(ctx, ac, id)` |
| WebhookTestHandler | Remove ~100 lines synchronous test logic → `svc.Webhooks.TestWebhook(ctx, id)` |
| WebhookDeliveryListHandler | `svc.Webhooks.ListDeliveries(ctx, id)` |
| WebhookDeliveryRetryHandler | `svc.Webhooks.RetryDelivery(ctx, deliveryID)` |

API handler signature changes: `WebhookTestHandler` and `WebhookDeliveryRetryHandler` currently take `(w, r, c config.Config, dispatcher publishing.WebhookDispatcher)`. After migration both become `(w, r, svc *service.Registry)` — the dispatcher is accessed through the service.

**internal/router/locales.go** — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`:

| Function | Key Change |
|----------|-----------|
| LocalesHandler | Collection CRUD → `svc.Locales.*` |
| LocaleHandler | Item CRUD → `svc.Locales.*` |
| LocalesPublicHandler | `svc.Locales.ListEnabledLocales(ctx)` |

Internal functions `apiCreateLocale`, `apiUpdateLocale`, `apiDeleteLocale`, `apiGetLocale`, `apiListLocales`: remove inline validation (locale code, fallback chain, default clearing) → call service methods.

**internal/router/translations.go** — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`:

| Function | Key Change |
|----------|-----------|
| TranslationHandler | `svc.Locales.CreateTranslation(ctx, ac, contentID, locale)` |
| AdminTranslationHandler | `svc.Locales.CreateAdminTranslation(ctx, ac, adminContentID, locale)` |

Remove ~150 lines of inline translation logic per handler (field enumeration, duplicate checking, default value copying).

**internal/router/locale_resolve.go** — business logic moves to LocaleService. The standalone functions (`ResolveLocale`, `WalkFallback`, `BuildLocaleMetadata`) are deleted after their logic moves to `LocaleService` methods.

**internal/router/slugs.go** — migrated from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)` in the same mechanical pattern used for all other handler migrations in Phases 2–3. `SlugHandler` is the only caller of the `locale_resolve.go` functions, so migrating it eliminates the need for wrapper functions entirely. The migration is straightforward: replace `db.ConfigDB(c)` with `svc.Driver()`, replace `ResolveLocale(r, &c, d)` with `svc.Locales.ResolveLocale(r)`, replace `WalkFallback(d, id, locale)` with `svc.Locales.WalkFallback(ctx, id, locale)`, replace `BuildLocaleMetadata(d, id)` with `svc.Locales.BuildLocaleMetadata(ctx, id)`. All other handler logic (slug parsing, tree walking, response formatting) stays unchanged.

### Mux Wiring (internal/router/mux.go)

**Webhook API routes** — change closures from `*c` to `svc`:
```go
// Before
WebhookListHandler(w, r, *c)
WebhookTestHandler(w, r, *c, dispatcher)
WebhookDeliveryRetryHandler(w, r, *c, dispatcher)

// After
WebhookListHandler(w, r, svc)
WebhookTestHandler(w, r, svc)
WebhookDeliveryRetryHandler(w, r, svc)
```

**Locale API routes** — change closures from `*c` to `svc`:
```go
// Before
LocalesHandler(w, r, *c)
LocaleHandler(w, r, *c)
LocalesPublicHandler(w, r, *c)
TranslationHandler(w, r, *c)
AdminTranslationHandler(w, r, *c)

// After
LocalesHandler(w, r, svc)
LocaleHandler(w, r, svc)
LocalesPublicHandler(w, r, svc)
TranslationHandler(w, r, svc)
AdminTranslationHandler(w, r, svc)
```

**Webhook admin routes** — change from `(driver, mgr)` to `(svc)`:
```go
// Before
adminhandlers.WebhookSettingsHandler(driver, mgr)
adminhandlers.WebhookCreateHandler(driver, mgr)
adminhandlers.WebhookTestHandler(driver, mgr)

// After
adminhandlers.WebhookSettingsHandler(svc)
adminhandlers.WebhookCreateHandler(svc)
adminhandlers.WebhookTestHandler(svc)
```

**Locale admin routes** — change from `(driver, mgr)` to `(svc)`:
```go
// Before
adminhandlers.LocaleSettingsHandler(driver, mgr)
adminhandlers.LocaleCreateHandler(driver, mgr)

// After
adminhandlers.LocaleSettingsHandler(svc)
adminhandlers.LocaleCreateHandler(svc)
```

**Plugin admin routes** — change from `(driver)` to `(svc)`:
```go
// Before
adminhandlers.PluginsListHandler(driver)
adminhandlers.PluginDetailHandler(driver)

// After
adminhandlers.PluginsListHandler(svc)
adminhandlers.PluginDetailHandler(svc)
```

**Plugin API routes** — no change. The `MountAdminEndpoints(mux, authChain, readPerm, adminPerm)` call on HTTPBridge stays as-is. These handlers take `*Manager` directly and are already well-factored. They will be rewired in a future phase if the plugin admin panel needs unified service calls, or in Phase 7 when MCP is rewired.

---

## Registry Changes

```go
type Registry struct {
    // ... existing fields ...

    Schema       *SchemaService
    Content      *ContentService
    AdminContent *AdminContentService
    Media        *MediaService
    Routes       *RouteService
    Plugins      *PluginService   // Phase 5
    Webhooks     *WebhookService  // Phase 5
    Locales      *LocaleService   // Phase 5
}
```

NewRegistry additions:

```go
// Phase 5 — PluginService requires a PluginManager which may be nil
// (plugins not enabled). PluginService handles nil gracefully.
reg.Plugins = NewPluginService(pluginMgr)  // pluginMgr may be nil
reg.Webhooks = NewWebhookService(driver, mgr, dispatcher)
reg.Locales = NewLocaleService(driver, mgr)
```

**NewRegistry signature is unchanged.** PluginService is the first service that depends on infrastructure outside the existing `(driver, mgr, pc, emailSvc, dispatcher)` set. Rather than adding a 6th positional parameter (which would break `routes_test.go` and trend toward parameter sprawl), PluginService is constructed *after* `NewRegistry` returns and assigned to the exported `Plugins` field:

```go
// cmd/serve.go — NewRegistry signature unchanged
reg := service.NewRegistry(driver, mgr, pc, emailSvc, dispatcher)
reg.Plugins = service.NewPluginService(pluginMgr)  // pluginMgr may be nil
reg.Webhooks = service.NewWebhookService(driver, mgr, dispatcher)
reg.Locales = service.NewLocaleService(driver, mgr)
```

WebhookService and LocaleService are also constructed externally for consistency — all Phase 5 services use the same post-construction pattern. This keeps `NewRegistry` stable across phases and avoids breaking existing callers.

**PluginService handles nil PluginManager internally** — reads return empty results, mutations return `*ValidationError{field: "plugins", message: "plugin system is not enabled"}`. No nil checks needed at the handler level.

---

## Files Changed

### New Files

**Note:** `plugins.go`, `webhooks.go`, and `locales.go` may exist as Phase 5 stubs (package declaration + comment only). Replace entire contents.

| File | Content |
|------|---------|
| internal/service/plugins.go | PluginService struct, PluginManager interface, methods, service types |
| internal/service/webhooks.go | WebhookService struct, store interface, methods, param types |
| internal/service/locales.go | LocaleService struct, methods, param types, validation helpers |
| internal/service/plugins_test.go | Service tests (mock PluginManager) |
| internal/service/webhooks_test.go | Service tests (SQLite) |
| internal/service/locales_test.go | Service tests (SQLite) |

### Modified Files

| File | Change | Scope |
|------|--------|-------|
| internal/service/service.go | Add Plugins, Webhooks, Locales fields to Registry struct | Small |
| cmd/serve.go | Construct Phase 5 services post-NewRegistry, assign to exported fields | Small |
| internal/router/webhooks.go | Change all 8 handlers from `(w, r, c)` to `(w, r, svc)` | Major |
| internal/router/locales.go | Change 3 handlers from `(w, r, c)` to `(w, r, svc)` | Moderate |
| internal/router/translations.go | Change 2 handlers from `(w, r, c)` to `(w, r, svc)`, remove inline logic | Major |
| internal/router/locale_resolve.go | Move all logic (including `parseAcceptLanguage`) to service, delete file entirely | Moderate |
| internal/admin/handlers/webhooks.go | Change 6 handlers from `(driver, mgr)` to `(svc)` | Major |
| internal/admin/handlers/locale_settings.go | Change 5 handlers from `(driver, mgr)` to `(svc)` | Major |
| internal/admin/handlers/plugins.go | Change 2 handlers from `(driver)` to `(svc)`, add real data rendering | Moderate |
| internal/router/mux.go | Update ~25 handler registrations | Moderate |
| internal/router/slugs.go | Migrate from `(w, r, c)` to `(w, r, svc)`, call LocaleService directly | Moderate |
| internal/admin/pages/plugins_list.templ | Accept `[]service.PluginSummary` data. Render table rows: Name, Version, Description, State (with status badge), CircuitBreakerState. Existing `plugins_table_rows.templ` has a 4-column layout already scaffolded — extend it | Moderate |
| internal/admin/pages/plugin_detail.templ | Accept `*service.PluginDetail` data. Render: summary fields + Author, License, FailedReason (if non-empty), VM pool stats (Available/Total), Dependencies list, SchemaDrift warning. Standard detail page layout matching other detail pages | Moderate |

### Not Changed

- `internal/plugin/manager.go` — reused as-is via interface
- `internal/plugin/http_bridge.go` — MountOn, MountAdminEndpoints unchanged
- `internal/plugin/cli_commands.go` — handler factories unchanged
- `internal/plugin/hook_handlers.go` — handler factories unchanged
- `internal/plugin/request_handlers.go` — handler factories unchanged
- `internal/webhooks/dispatcher.go` — lifecycle stays in cmd/serve.go
- `internal/webhooks/signing.go` — called from service
- `internal/publishing/` — unchanged, still calls dispatcher directly
- `mcp/` — deferred to Phase 7

---

## Testing

### internal/service/plugins_test.go

Uses a mock `PluginManager` (in-memory state):

- List when no plugins → empty slice
- List with plugins → correct summaries
- Get existing plugin → correct detail
- Get non-existent → NotFoundError
- Enable/Disable/Reload → delegates to mock, returns nil error
- Enable non-existent → NotFoundError
- CleanupDryRun → returns orphaned table list
- CleanupDrop → returns dropped table list
- All methods with nil manager → graceful (empty/ValidationError)

### internal/service/webhooks_test.go

SQLite-based, table-driven:

- Create with valid input → success, secret auto-generated
- Create with explicit secret → preserves provided secret
- Create with empty name → ValidationError
- Create with empty URL → ValidationError
- Create with SSRF URL (127.0.0.1) → ValidationError
- Create with empty events → ValidationError
- Update rename → preserves immutable fields
- Update with SSRF URL → ValidationError
- Delete → success
- Delete non-existent → NotFoundError
- Get existing → correct fields
- Get non-existent → NotFoundError
- List paginated → correct counts
- Test webhook → requires HTTP mock (httptest.NewServer), verify signature header present
- Delivery list by webhook → returns deliveries

### internal/service/locales_test.go

SQLite-based, table-driven:

- Create with valid BCP 47 code → success
- Create with invalid code → ValidationError (various: too short, underscore, trailing hyphen)
- Create with empty label → ValidationError
- Create with is_default → clears other defaults
- Create with duplicate code → ConflictError (DB unique constraint)
- Create with fallback self-reference → ValidationError
- Create with fallback cycle → ValidationError
- Create with fallback to non-existent → ValidationError
- Update rename code → success
- Update set default → clears others
- Delete non-default → success
- Delete default locale → ForbiddenError
- Delete non-existent → NotFoundError
- ListEnabledLocales → returns only enabled
- ResolveLocale → query param priority > Accept-Language > config default
- CreateTranslation → correct field count, skips non-translatable, skips duplicates
- CreateTranslation with i18n disabled → ValidationError
- CreateTranslation with disabled locale → ValidationError
- WalkFallback → resolves through chain, respects max 2 hops

---

## Implementation Order

All three services are independent and can be built by parallel agents after a shared prerequisite.

### Step 0 — Registry Wiring (prerequisite, do first)

Add `Plugins`, `Webhooks`, `Locales` fields to Registry. Update `cmd/serve.go` to construct Phase 5 services post-`NewRegistry` and assign to exported fields. All three `New*Service` constructors can initially be stubs returning empty structs.

Verify: `go build ./...` + `just test` must pass before streams diverge.

### Stream A — PluginService (~1 session)

1. Define `PluginManager` interface in `internal/service/plugins.go`
2. Implement `PluginService` struct + all methods (including request approval methods)
3. Write `internal/service/plugins_test.go` with mock manager
4. Update templ templates: `plugins_list.templ` to accept `[]PluginSummary`, `plugin_detail.templ` to accept `*PluginDetail`
5. Rewire `internal/admin/handlers/plugins.go` to `(svc)` — add real data rendering using updated templates
6. Update mux.go plugin admin registrations
7. `go build ./...` + `just test`

### Stream B — WebhookService (~1–2 sessions)

1. Implement `WebhookService` struct + all methods in `internal/service/webhooks.go`
2. Write `internal/service/webhooks_test.go`
3. Rewire `internal/admin/handlers/webhooks.go` to `(svc)`
4. Rewire `internal/router/webhooks.go` to `(w, r, svc)` pattern
5. Update mux.go webhook registrations (API + admin)
6. `go build ./...` + `just test`

### Stream C — LocaleService (~1–2 sessions)

1. Implement `LocaleService` struct + all methods in `internal/service/locales.go`
2. Move `validateLocaleCode()` and `validateFallbackChain()` from router to service
3. Move `ResolveLocale()`, `WalkFallback()`, `BuildLocaleMetadata()` and the private `parseAcceptLanguage()` helper from locale_resolve.go to the service. Delete `internal/router/locale_resolve.go` entirely (no remaining functions after migration)
4. Write `internal/service/locales_test.go`
5. Rewire `internal/admin/handlers/locale_settings.go` to `(svc)` (including `renderLocaleTableRows` helper)
6. Rewire `internal/router/locales.go` + `translations.go` to `(w, r, svc)` pattern
7. Migrate `slugs.go` from `(w, r, c)` to `(w, r, svc)` — replace `ResolveLocale`/`WalkFallback`/`BuildLocaleMetadata` calls with `svc.Locales.*` calls (same mechanical pattern as all other handler migrations)
8. Update mux.go locale registrations (API + admin + translations + slugs handler closure)
9. `go build ./...` + `just test`

### Final — Integration verification

After all three streams merge:
```
go build ./...
just test
go vet ./...
```

---

## Verification

After each stream:
1. `go build ./...` — compiles clean
2. `just test` — all existing tests pass
3. `go vet ./...` — no vet warnings

After Phase 5 complete:
4. `go test -v ./internal/service/` — all new service tests pass
5. Manual smoke: `just run` → admin panel:
   - Plugins page renders actual plugin data (or empty state if no plugins)
   - Webhook settings page: create/edit/delete/test webhooks work
   - Locale settings page: create/edit/delete locales work, fallback validation works
   - Translation creation via API works
