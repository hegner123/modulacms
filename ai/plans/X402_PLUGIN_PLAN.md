# x402 Payment Plugin Plan

Pure Lua plugin using read lifecycle hooks on the standard content delivery endpoint.

Follows the same structure and patterns as the `forms` plugin (`plugins/forms/`).

---

## Architecture

The x402 plugin hooks into the **standard content delivery endpoint** (`GET /api/v1/content/{slug}`) via `before_read` and `after_read` lifecycle hooks. No separate plugin delivery route needed — clients don't need to know about x402.

```
Client → GET /api/v1/content/{slug}
         ↓
    before_read hook fires (synchronous, db/request ALLOWED)
    ├── Plugin queries rules table for this slug
    ├── Not gated? → return nil (let delivery proceed normally)
    ├── Gated, no PAYMENT-SIGNATURE header?
    │   → return structured 402 response + PAYMENT-REQUIRED header
    └── Gated, PAYMENT-SIGNATURE present?
        → POST facilitator /verify (via request.post)
        → If invalid: return structured 402 response
        → If valid: store verification in hook context, return nil (proceed)
         ↓
    Content delivery happens normally (slug lookup, snapshot, tree build, transform)
         ↓
    after_read hook fires (synchronous, NOT fire-and-forget)
    ├── Plugin checks hook context for verified payment
    ├── POST facilitator /settle (via request.post)
    ├── Record transaction in plugin tables
    └── Return headers to append (PAYMENT-RESPONSE)
         ↓
Client ← 200 + content + PAYMENT-RESPONSE header
```

---

## Prerequisite: Read Lifecycle Hook Infrastructure (Phase 1)

The current hook system supports mutation events only (create, update, delete, publish). This plugin requires two new hook events with **fundamentally different execution semantics** than mutation hooks.

### Why read hooks differ from mutation hooks

Mutation before-hooks run inside a database transaction. The `inBeforeHook` flag on `DatabaseAPI` blocks `db.*` calls directly, and `core.*` calls indirectly (because `CoreTableAPI` delegates to `dbAPI.checkOpLimit()` which checks the same flag). The `inBeforeHook` flag on `requestAPIState` blocks `request.send()` and all shorthands (`request.post`, `request.get`, `request.put`, `request.delete`, `request.patch`) since they all route through `doSend()`. This prevents deadlocks (SQLite file-level lock + plugin's separate connection pool). Read hooks have no transaction — they fire in the delivery handler before any DB work starts. The execution policy must be different.

### Go changes required (8 items)

**1. New hook event types** in `internal/db/audited/hooks.go`:
- Add `HookBeforeRead` and `HookAfterRead` to the `HookEvent` constants and `ValidHookEvents` map.

**2. New execution path in hook engine** in `internal/plugin/hook_engine.go`:
- Add a new method separate from `RunBeforeHooks` with this exact signature:
  ```go
  func (e *HookEngine) RunBeforeReadHooks(ctx context.Context, table string, data map[string]any) (*ReadHookResponse, map[string]any, error)
  ```
  Where `table` is the hook table name (e.g., `"content_data"`), `data` contains the request context including `slug`, and the second return value is the state map for passing to `RunAfterReadHooks`.
- This method does NOT set `inBeforeHook = true`. It does NOT run inside a transaction context.
- `db.*`, `core.*`, and `request.*` (send + all shorthands) are all permitted during read hook execution because `inBeforeHook` is never set.
- Must not weaken existing mutation hook guards. The `inBeforeHook` flag must remain set for `before_create`, `before_update`, `before_delete`, `before_publish`. The differentiation is by event type, not a global toggle.

**3. Structured hook return value protocol**:
- Current before-hooks: `L.CallByParam` with `NRet: 0`. Hooks abort via `error()` → 422. No return value.
- Read before-hooks: `L.CallByParam` with `NRet: 1`. Read the return value from the Lua stack.
  - If `nil`: proceed with delivery.
  - If table with `status` field: abort delivery, use the table as the HTTP response. Parse into this Go struct:
    ```go
    // ReadHookResponse lives in internal/db/audited/hooks.go alongside ReadHookRunner.
    type ReadHookResponse struct {
        Status  int               // HTTP status code (e.g., 402)
        Headers map[string]string // HTTP headers to set
        Body    map[string]any    // JSON body to marshal and write
    }
    ```
    The Lua table shape is `{ status, headers, json }` — same as plugin route handler responses parsed in `internal/plugin/http_bridge.go`'s response handling. Reuse that parsing logic where possible.
- Read after-hooks: `L.CallByParam` with `NRet: 1`. Read the return value.
  - If `nil`: no headers to append.
  - If table with `headers` field: merge those headers into the HTTP response.

**4. Synchronous after-read dispatch** in `internal/plugin/hook_engine.go`:
- Add a new method with this exact signature:
  ```go
  func (e *HookEngine) RunAfterReadHooks(ctx context.Context, table string, data map[string]any, state map[string]any) (map[string]string, error)
  ```
  Returns collected headers to append to the HTTP response.
- Current `RunAfterHooks` dispatches in goroutines (`go func()`) and returns void. After-read hooks must run synchronously before the response is written, and their return values (headers) must be collected and returned to the caller.
- After-read hooks are permitted to call `db.*` and `request.send()` (same policy as before-read).

**5. State passing between before_read and after_read**:
- Current hook engine creates independent Lua tables for before and after invocations via `audited.StructToMap`. Mutations to `data` in before_read do not appear in after_read.
- **Decision**: State passes via the `data` table itself. After `L.CallByParam` completes in `RunBeforeReadHooks`, read the Lua `data` table back into a Go `map[string]any` (capturing any `_`-prefixed keys the hook set). Return this map as the `state` return value. `RunAfterReadHooks` uses this same map as the input `data` table for after-read hooks. Keys prefixed with `_` are reserved for plugin state (e.g., `_x402_tx_id`).
- The implementing agent must ensure a `LuaTableToMap` inverse of `MapToLuaTable` exists or is created. If it exists but does not handle nested tables, extend it — the x402 plugin stores a nested `requirements` table in `data._x402_requirements`.

**6. `json` module for the Lua sandbox** in `internal/plugin/sandbox.go`:
- Register a `json` module with at minimum `json.encode(table) -> string` and `json.decode(string) -> table`.
- The Go side has `encoding/json` — wrap it as a frozen Lua module.
- This benefits all plugins, not just x402.

**7. Wiring: make read hook methods accessible from `slugs.go`**:
- Currently `slugs.go` receives `*service.Registry` which has no access to `*plugin.HookEngine`. The existing `audited.HookRunner` interface (defined in `internal/db/audited/hooks.go`) only has `HasHooks`, `RunBeforeHooks`, and `RunAfterHooks` — it does not include the new read hook methods.
- Define a new interface `ReadHookRunner` in `internal/db/audited/hooks.go` with the two methods above. Implement it on `*plugin.HookEngine`. Do NOT add these methods to the existing `HookRunner` interface — that would require updating all mock/test implementations across 147+ importing files.
- Inject `ReadHookRunner` into request context via a new `middleware.ReadHookRunnerMiddleware` function. Follow the exact pattern of `middleware.HookRunnerMiddleware` in `internal/middleware/audit_helpers.go`. Define a `readHookRunnerKey` context key in the same file. Add an extraction function `ReadHookRunnerFromContext(ctx) ReadHookRunner`. Wire it in `cmd/serve.go` at line ~423, adjacent to the existing `HookRunnerMiddleware` call. Do NOT modify `service.Registry`.

**8. Integration in delivery handler** in `internal/router/slugs.go`:
- The slug is extracted via `strings.TrimPrefix(r.URL.Path, "/api/v1/content")` (not a Go 1.22+ `{slug}` path parameter). Build the hook `data` table using this same extraction.
- In `apiGetSlugContent` (or `apiGetSlugContentPublished`), before the slug lookup:
  - Build the `data` table from the request (slug from TrimPrefix, headers with all keys lowercased, client_ip, query params). When building `data.headers` from `r.Header`, lowercase all header names before inserting into the map — Go's `http.Header` uses `CanonicalHeaderKey` (e.g., `Payment-Signature`) but Lua code expects lowercase (e.g., `payment-signature`).
  - Get the `ReadHookRunner` via `middleware.ReadHookRunnerFromContext(r.Context())`. If nil (plugins disabled), skip all hook calls and proceed with normal delivery. Do not return an error.
  - Call `readHookRunner.RunBeforeReadHooks(ctx, "content_data", data)`.
  - If it returns a non-nil response: write that response directly (status, headers, JSON body) and return. Do not proceed with delivery.
  - If it returns nil + state: proceed with delivery, keep state for after-read.
- After content is fetched and tree is built, before `applyFormatAndTransform`:
  - Call `readHookRunner.RunAfterReadHooks(ctx, "content_data", data, state)`.
  - Apply any returned headers to `w.Header()`.
  - Then call `applyFormatAndTransform` as normal.

### Files that change in Phase 1

| File | Change |
|------|--------|
| `internal/db/audited/hooks.go` | Add `HookBeforeRead`, `HookAfterRead` constants |
| `internal/plugin/hook_engine.go` | Add `RunBeforeReadHooks`, `RunAfterReadHooks` methods |
| `internal/plugin/sandbox.go` | Register `json` module (encode/decode) and freeze it via `FreezeModule` |
| `internal/plugin/hooks_api.go` | Allow `before_read`/`after_read` in `hooks.on()` registration |
| `internal/db/audited/hooks.go` | Define `ReadHookRunner` interface and `ReadHookResponse` struct |
| `internal/middleware/audit_helpers.go` | Add `ReadHookRunnerMiddleware`, `readHookRunnerKey`, `ReadHookRunnerFromContext` |
| `cmd/serve.go` | Wire `ReadHookRunnerMiddleware` (adjacent to existing `HookRunnerMiddleware` at line ~423) |
| `internal/plugin/db_api.go` | No change — `inBeforeHook` stays as-is for mutations |
| `internal/plugin/request_api.go` | No change — `inBeforeHook` stays as-is for mutations |
| `internal/plugin/core_api.go` | No change — `core.*` inherits block from `dbAPI.checkOpLimit()`, unblocked by not setting `inBeforeHook` |
| `internal/router/slugs.go` | Add hook invocation in delivery handler |

### Test scenarios for Phase 1

- Read hooks do NOT block `db.*`, `core.*`, or `request.send()`.
- Mutation hooks still block `db.*`, `core.*`, and `request.send()` (regression test).
- `before_read` returning `{ status = 402, ... }` aborts delivery with that status.
- `before_read` returning `nil` lets delivery proceed.
- State set on `data._foo` in `before_read` is accessible in `after_read`.
- `after_read` returning `{ headers = { ... } }` appends those headers to the response.
- `after_read` returning `nil` appends no headers.
- Multiple plugins registering `before_read` execute in priority order; first non-nil return aborts.
- Multiple plugins registering `after_read` execute in priority order; all returned headers are merged.
- `json.encode` and `json.decode` work correctly in the sandbox.

---

## Protocol Summary

Key terms:
- **Resource server**: ModulaCMS (via the x402 plugin)
- **Facilitator**: External service that verifies and settles payments (e.g., Coinbase at `x402.org`)
- **Scheme**: How money moves (`exact` = fixed price)
- **Network**: Which chain (CAIP-2 identifier, e.g., `eip155:8453` for Base)

### Replay protection

The x402 protocol's `exact` scheme uses EIP-3009 `transferWithAuthorization`, which includes a nonce. The facilitator's `/verify` endpoint checks that the authorization has not been used. The `/settle` endpoint submits it on-chain, which consumes the nonce — a second settlement attempt with the same nonce is rejected by the smart contract. This means replay protection is handled at the protocol and blockchain level, not by the plugin.

The plugin does NOT need to track used signatures. If the facilitator verifies a signature, it is valid. If it has already been settled, the facilitator will reject it on `/settle`.

### Facilitator availability

The facilitator (`x402.org` or operator-configured) is a hard dependency for all gated content. If the facilitator is unreachable:

- **Verify timeout**: `request.post` with `timeout = 10` will fail. The `before_read` hook treats verification failure as rejection and returns 402 to the client. Gated content becomes inaccessible. Non-gated content is unaffected.
- **Settle timeout**: `request.post` with `timeout = 15` will fail. The `after_read` hook records the transaction as `"failed"` and logs a warning. Content is still served (verification passed — the client proved they can pay). The operator must monitor failed settlements and retry or reconcile manually.

There is no circuit breaker at the plugin level. The `request` engine has per-domain circuit breakers (after N consecutive failures, the domain is disabled). If the facilitator circuit breaker trips, ALL verify calls fail immediately, and all gated content returns 402. The operator is alerted via the plugin-level circuit breaker mechanism.

**Defined behavior**:
| Facilitator state | Verify result | Settle result | Client sees |
|---|---|---|---|
| Healthy | Pass | Pass | 200 + content + PAYMENT-RESPONSE header |
| Healthy | Fail | N/A | 402 (payment rejected) |
| Unreachable | Fail (timeout) | N/A | 402 (verification failed) |
| Verify OK, settle down | Pass | Fail | 200 + content (no PAYMENT-RESPONSE header). Transaction marked "failed". |

**Settlement failure is a revenue leak.** If verify passes but settle fails, the client gets content without payment. The `transactions.stats` endpoint surfaces failed settlement counts. Operators MUST monitor `status = "failed"` transactions. The plan does not add automated alerting (out of scope for v0.1.0), but the data is available for external monitoring.

---

## Plugin Structure

```
plugins/
  x402/
    init.lua              -- Manifest, request.register, hooks, admin routes, on_init, on_shutdown
    lib/
      utils.lua           -- paginate, response builders (same pattern as forms)
      payment.lua         -- build_requirements, verify, settle, get_rules helpers
      rules.lua           -- Rules CRUD handlers
      transactions.lua    -- Transaction list/get/stats handlers
      settings.lua        -- Settings get/update handlers
```

No `deliver.lua` — the delivery logic lives in the `before_read` and `after_read` hooks in `init.lua`.

---

## Manifest & Domain Registration

```lua
local utils        = require("utils")
local payment      = require("payment")
local rules        = require("rules")
local transactions = require("transactions")
local settings_mod = require("settings")

plugin_info = {
    name        = "x402",
    version     = "0.1.0",
    description = "x402 payment gating for content delivery",
    author      = "ModulaCMS",
    license     = "MIT",

    core_access = {
        content_data = {"read"},
    },
}

-- Register facilitator domain for outbound HTTP (admin must approve)
request.register("x402.org", {
    description = "x402 facilitator for payment verification and settlement",
})
```

---

## Hooks — The x402 Flow

```lua
-- before_read: gate content delivery with x402 payment verification
--
-- This hook runs OUTSIDE any database transaction.
-- db.*, core.*, and request.send() are all permitted.
hooks.on("before_read", "content_data", function(data)
    local wallet = payment.setting("wallet_address")
    if not wallet or wallet == "" then
        return nil  -- x402 not configured, pass through
    end

    local rules_list = payment.get_rules(data.slug)
    if #rules_list == 0 then
        return nil  -- not gated, pass through
    end

    -- Gated: check for payment header
    local payment_header = data.headers["payment-signature"]

    if not payment_header or payment_header == "" then
        -- No payment: return 402 with requirements
        local requirements = payment.build_requirements(rules_list)
        return {
            status  = 402,
            headers = {
                ["payment-required"] = json.encode(requirements),
            },
            json = {
                error        = "payment required",
                slug         = data.slug,
                requirements = requirements,
            },
        }
    end

    -- Verify payment with facilitator
    local requirements = payment.build_requirements(rules_list)
    local verified, verify_result = payment.verify(payment_header, requirements)

    if not verified then
        -- Record rejected transaction
        db.insert("transactions", {
            id            = db.ulid(),
            rule_id       = rules_list[1].id,
            content_id    = rules_list[1].content_id,
            slug          = data.slug,
            client_ip     = data.client_ip,
            status        = "rejected",
            amount        = rules_list[1].price,
            currency      = rules_list[1].currency,
            network       = rules_list[1].network,
            scheme        = rules_list[1].scheme,
            error_message = verify_result,
        })

        return {
            status = 402,
            json   = { error = verify_result },
        }
    end

    -- Payment verified — record and let delivery proceed
    local tx_id = db.ulid()
    db.insert("transactions", {
        id          = tx_id,
        rule_id     = rules_list[1].id,
        content_id  = rules_list[1].content_id,
        slug        = data.slug,
        client_ip   = data.client_ip,
        status      = "verified",
        amount      = rules_list[1].price,
        currency    = rules_list[1].currency,
        network     = rules_list[1].network,
        scheme      = rules_list[1].scheme,
        verified_at = db.timestamp(),
    })

    -- Store state for after_read to settle.
    -- The hook engine reads back the data table after this hook returns,
    -- capturing these _-prefixed keys, and passes them to after_read.
    data._x402_tx_id          = tx_id
    data._x402_payment_header = payment_header
    data._x402_requirements   = requirements

    return nil  -- proceed with delivery
end, { priority = 10 })


-- after_read: settle payment and append response header
--
-- Runs synchronously BEFORE the response is written.
-- db.* and request.send() are permitted.
-- Return value: { headers = { ... } } to append headers, or nil.
hooks.on("after_read", "content_data", function(data)
    local tx_id          = data._x402_tx_id
    local payment_header = data._x402_payment_header
    local requirements   = data._x402_requirements

    if not tx_id then
        return nil  -- not a gated request, nothing to do
    end

    -- Settle with facilitator
    local settled, settle_result = payment.settle(payment_header, requirements)

    if settled then
        local tx_hash = ""
        if settle_result and settle_result.txHash then
            tx_hash = settle_result.txHash
        end
        db.update("transactions", {
            set   = { status = "settled", tx_hash = tx_hash, settled_at = db.timestamp() },
            where = { id = tx_id },
        })
        return {
            headers = {
                ["payment-response"] = json.encode(settle_result),
            },
        }
    else
        db.update("transactions", {
            set   = { status = "failed", error_message = settle_result },
            where = { id = tx_id },
        })
        log.error("x402 settlement FAILED — content served without payment", {
            tx_id  = tx_id,
            slug   = data.slug,
            error  = settle_result,
            amount = data._x402_requirements and data._x402_requirements.accepts
                     and data._x402_requirements.accepts[1]
                     and data._x402_requirements.accepts[1].maxAmountRequired or "unknown",
        })
        return nil  -- content still served; operator must monitor failed settlements
    end
end, { priority = 10 })
```

---

## Admin Routes

```lua
-- Admin: Rules CRUD (5) -------------------------------------------------------
http.handle("GET",    "/rules",        rules.list)
http.handle("POST",   "/rules",        rules.create)
http.handle("GET",    "/rules/{id}",   rules.get)
http.handle("PUT",    "/rules/{id}",   rules.update)
http.handle("DELETE", "/rules/{id}",   rules.delete)

-- Admin: Transactions (2) ----------------------------------------------------
http.handle("GET",    "/transactions",      transactions.list)
http.handle("GET",    "/transactions/{id}", transactions.get)

-- Admin: Settings (2) --------------------------------------------------------
http.handle("GET",    "/settings",     settings_mod.get)
http.handle("PUT",    "/settings",     settings_mod.update)

-- Admin: Stats (1) -----------------------------------------------------------
http.handle("GET",    "/stats",        transactions.stats)

-- Public: Check gating status without delivery (1) ---------------------------
http.handle("GET",    "/check/{slug}",  function(req)
    local slug = req.params.slug
    if not slug or slug == "" then
        return utils.bad_request("slug required")
    end
    local rules_list = payment.get_rules(slug)
    if #rules_list == 0 then
        return utils.success_response({ gated = false, slug = slug })
    end
    local requirements = payment.build_requirements(rules_list)
    return utils.success_response({
        gated        = true,
        slug         = slug,
        requirements = requirements,
    })
end, { public = true })
```

Admin routes under `/api/v1/plugins/x402/`. The delivery itself goes through the standard `/api/v1/content/{slug}`.

The `stats` endpoint includes failed settlement counts so operators can monitor revenue leaks:

```lua
-- transactions.stats returns:
{
    active_rules       = N,
    total_transactions = N,
    verified           = N,  -- verified but not yet settled (in-flight)
    settled            = N,  -- successfully settled
    failed             = N,  -- settlement failed (REVENUE LEAK — monitor this)
    rejected           = N,  -- verification failed (no content served)
}
```

---

## Tables (on_init)

```lua
function on_init()
    db.define_table("rules", {
        columns = {
            { name = "content_id",  type = "text",    not_null = true },
            { name = "slug",        type = "text",    not_null = true },
            { name = "price",       type = "text",    not_null = true },  -- decimal string "0.01"
            { name = "currency",    type = "text",    not_null = true, default = "USDC" },
            { name = "network",     type = "text",    not_null = true },  -- CAIP-2: "eip155:8453"
            { name = "scheme",      type = "text",    not_null = true, default = "exact" },
            { name = "description", type = "text",    default = "" },
            { name = "active",      type = "boolean", not_null = true, default = 1 },
        },
        indexes = {
            { columns = {"slug"} },
            { columns = {"content_id"} },
            { columns = {"slug", "active"} },
            { columns = {"content_id", "network", "currency"}, unique = true },
        },
    })

    db.define_table("transactions", {
        columns = {
            { name = "rule_id",       type = "text",    not_null = true },
            { name = "content_id",    type = "text",    not_null = true },
            { name = "slug",          type = "text",    not_null = true },
            { name = "client_ip",     type = "text",    not_null = true },
            { name = "status",        type = "text",    not_null = true },
            { name = "amount",        type = "text",    not_null = true },
            { name = "currency",      type = "text",    not_null = true },
            { name = "network",       type = "text",    not_null = true },
            { name = "scheme",        type = "text",    not_null = true },
            { name = "tx_hash",       type = "text" },
            { name = "error_message", type = "text" },
            { name = "verified_at",   type = "timestamp" },
            { name = "settled_at",    type = "timestamp" },
        },
        indexes = {
            { columns = {"content_id"} },
            { columns = {"status"} },
            { columns = {"slug"} },
            { columns = {"created_at"} },
        },
    })

    db.define_table("settings", {
        columns = {
            { name = "key",   type = "text", not_null = true, unique = true },
            { name = "value", type = "text", not_null = true },
        },
    })

    local function set_default(key, value)
        if not db.exists("settings", { where = { key = key } }) then
            db.insert("settings", { key = key, value = value })
        end
    end

    set_default("facilitator_url", "https://x402.org/facilitator")
    set_default("wallet_address", "")
    set_default("default_network", "eip155:8453")
    set_default("default_currency", "USDC")

    log.info("x402 plugin initialized")
end

function on_shutdown()
    local rule_count = db.count("rules", { where = { active = 1 } })
    local tx_count   = db.count("transactions", {})
    log.info("x402 plugin shutting down", {
        active_rules = rule_count,
        transactions = tx_count,
    })
end
```

---

## Content Mutation Hooks

Keep payment rules in sync with content lifecycle:

```lua
hooks.on("after_delete", "content_data", function(data)
    local rules_list = db.query("rules", { where = { content_id = data.id } })
    for _, rule in ipairs(rules_list) do
        db.update("rules", {
            set   = { active = 0 },
            where = { id = rule.id },
        })
    end
    if #rules_list > 0 then
        log.info("x402: deactivated rules for deleted content", {
            content_id = data.id,
            count      = #rules_list,
        })
    end
end)

-- Sync slug changes. The after_update hook receives the full entity
-- after the update is committed. data.slug contains the current slug
-- value. If it differs from what the rule has stored, update the rule.
hooks.on("after_update", "content_data", function(data)
    if not data.slug or data.slug == "" then
        return  -- slug not present in entity data, skip
    end
    local rules_list = db.query("rules", { where = { content_id = data.id } })
    for _, rule in ipairs(rules_list) do
        if rule.slug ~= data.slug then
            db.update("rules", {
                set   = { slug = data.slug },
                where = { id = rule.id },
            })
        end
    end
end)
```

---

## lib/ Modules

### lib/payment.lua

`setting()`, `get_rules()`, `build_requirements()`, `verify()`, `settle()`. These helpers are used by the hooks.

### lib/utils.lua

Copy from `forms/lib/utils.lua`, trim to what's needed: `paginate`, `paginated_response`, `error_response`, `success_response`, `not_found`, `bad_request`.

### lib/rules.lua, lib/transactions.lua, lib/settings.lua

Standard CRUD, same patterns as `forms/lib/forms.lua` and `forms/lib/entries.lua`:

- `rules.create` — validates content exists via `core.query`, applies defaults from settings
- `transactions.stats` — aggregate counts by status + active rule count, with `failed` count prominently surfaced
- Everything else is paginated list / get by ID / update / delete

---

## Admin Approval Checklist

```bash
modulacms plugin approve x402 --all-routes
modulacms plugin approve x402 --all-hooks
modulacms plugin approve x402 --request "x402.org"
```

---

## Implementation Order

| Phase | What | Depends On |
|-------|------|------------|
| 1a | Add `HookBeforeRead`, `HookAfterRead` event types to `audited/hooks.go` | Nothing |
| 1b | Add `json` module (encode/decode) to Lua sandbox in `sandbox.go` | Nothing |
| 1c | Add `RunBeforeReadHooks` to hook engine — synchronous, no `inBeforeHook` flag, returns structured response or nil + state. Note: end-to-end testing requires Phase 1e (hooks_api.go validation). Unit test by calling the method directly with pre-populated hookIndex entries, bypassing Lua registration. | 1a |
| 1d | Add `RunAfterReadHooks` to hook engine — synchronous, returns collected headers, receives state from before-read | 1a |
| 1e | Allow `before_read`/`after_read` in `hooks.on()` registration in `hooks_api.go` | 1a |
| 1f | Define `ReadHookRunner` interface, add to `service.Registry` or context middleware, wire in `cmd/serve.go` | 1c, 1d |
| 1g | Integrate read hooks in `slugs.go` delivery handler — extract slug via TrimPrefix, call before-read, pass state, call after-read, apply headers | 1c, 1d, 1f |
| 1h | Tests: read hooks permit db/request access, mutation hooks still block, structured returns work, state passes, multi-plugin priority ordering | 1c, 1d, 1e, 1g |
| 2 | `init.lua` — manifest, request.register, on_init, on_shutdown | Nothing |
| 3 | `lib/utils.lua` + `lib/payment.lua` | 1b (json.encode) |
| 4 | Read hooks in `init.lua` (before_read + after_read) | 1h, 3 |
| 5 | Mutation hooks in `init.lua` (after_delete + after_update) | 2 |
| 6 | `lib/rules.lua` + `lib/transactions.lua` + `lib/settings.lua` | 3 |
| 7 | Admin routes + check endpoint in `init.lua` | 6 |
| 8 | Test: deploy, approve, exercise delivery + admin endpoints | 4-7 |

Phase 1 (a-h) is Go work — a significant expansion of the plugin system's hook infrastructure. Phases 2-7 are Lua files in `plugins/x402/`.

---

## Multi-Plugin Read Hook Interaction

When multiple plugins register `before_read` hooks on `content_data`:

- Hooks execute in priority order (lower number first). The x402 plugin uses `priority = 10`.
- The **first** hook to return a non-nil structured response aborts delivery. Remaining hooks do not fire.
- If all hooks return nil, delivery proceeds.
- For `after_read`: all registered hooks fire in priority order. Returned headers are merged (later hooks overwrite same-named headers from earlier hooks).

Plugins should use distinct priority values. If two plugins need to interact (e.g., analytics before x402), the operator controls execution order via priority.

---

## Dependents to Update

- `audited.HookRunner` interface — NOT modified (new `ReadHookRunner` interface created instead). Existing implementors unaffected: `*plugin.HookEngine`, `trackingHookRunner` in `audited_integration_test.go`, `mockPluginManager` in `plugins_test.go`.
- `ValidHookEvents` map in `audited/hooks.go` — adding entries is purely additive. Only consumer is `hooks_api.go:70` (validation).
- `sandbox.go` loaded libraries — adding `json` module does not conflict with existing `base`, `table`, `string`, `math`. Must call `FreezeModule(L, "json")` after registration.
- `SlugHandler` / delivery functions in `slugs.go` — only called from `mux.go:700`. Blast radius is contained.

---

## Constraints

- **No wildcard table matching for read hooks.** Mutation hooks support `table = "*"` (wildcard). Read hooks fire on every content delivery request, so wildcard matching is rejected. `hooks_api.go` validation in Phase 1e must reject `"*"` for `before_read` and `after_read` events.
- **The `log` module is already registered in the plugin sandbox** (via `RegisterLogAPI`). No new work needed for `log.info`, `log.warn`, `log.error` calls in the x402 plugin.

---

## Open Questions

1. **Which facilitator to default?** Seed `x402.org` as default; admin changes via settings.

2. **Multi-rule per content?** Schema supports it (unique on `content_id + network + currency`). The 402 `accepts` array includes all active rules — client picks network/currency.

3. **Settlement timing?** Sync in `after_read` (before response is written). Settlement latency adds to content delivery time. Expected facilitator latency is ~1-3s for on-chain settlement. If this becomes a bottleneck, settlement could be deferred to an async background job that updates transaction status — but that requires additional infrastructure (job queue). Sync is correct for v0.1.0.
