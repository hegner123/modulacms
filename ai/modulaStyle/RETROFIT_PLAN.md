# ModulaStyle Assertion Retrofit Plan

All new code follows MODULA_STYLE.md immediately. This plan tracks the eventual retrofit of existing packages, ordered by boundary risk. This is not a priority -- features and v1 milestones come first. Packages are retrofitted opportunistically when being actively modified, or as dedicated work when time permits.

## Prerequisites

### 0. Panic Recovery Middleware

**Before adding assertion panics to any handler or service code, build and wire panic recovery middleware.**

Without recovery middleware, a single assertion failure in request-scoped code crashes the entire process. This is a denial-of-service vector even during development.

**What to build:**

Create `internal/middleware/recovery.go`:
- Wrap each request in a `defer func() { recover() }`.
- On panic, log the panic value and stack trace via `slog.Error`.
- Respond with 500 and a generic error body (never leak the panic message to the client).
- In debug mode, include the panic message in the response for developer visibility.

Wire recovery as the **outermost** middleware in every HTTP chain (`internal/middleware/http_chain.go`), so it catches panics from all downstream handlers and middleware.

The existing `internal/plugin/recovery.go` handles Lua execution panics within the plugin package. The new middleware is separate -- it catches panics in the Go request path.

**Exit criterion:** `curl` a handler that contains `panic("test")` and receive a 500 response. The process must not crash.

### 0b. Rename `convert.go` Assert Functions

The functions `AssertString`, `AssertInteger`, `AssertInt32`, `AssertInt64` in `internal/db/convert.go` silently return zero values on type mismatch. They do not assert anything. Under ModulaStyle, "Assert" means "panic on violation" -- these names are actively misleading.

**Before retrofitting `internal/db/`**, rename these to `CoerceString`, `CoerceInteger`, `CoerceInt32`, `CoerceInt64` (or similar) to make the silent-fallback behavior explicit. Then decide per call site whether coercion is appropriate or whether a real assertion (panic on mismatch) is needed.

## Priority Order

### 1. internal/plugin/ (59 files)

**Boundary type:** Untrusted code execution (Go/Lua VM boundary).

**Why first:** Every value crossing the Lua boundary is untyped. A Lua script can return nil, wrong type, out-of-range numbers, or strings where tables are expected. The plugin system exposes DB operations, HTTP requests, schema manipulation, and hook registration to third-party code. A single unchecked Lua return value can corrupt application state.

**Sub-phases:** This package is too large for a single pass. Break into independently shippable phases:

#### 1a. Lua Boundary Conversions

Focus: `lua_helpers.go` (6KB) and every call site that pulls `LValue` from Lua.

Every `LValue` pulled from Lua must be type-checked and range-checked before use. The three failure modes are nil, wrong type, and overflow. Every Go value pushed to Lua must match the documented plugin API contract.

This sub-phase establishes the boundary assertion pattern that the rest of the plugin package depends on.

#### 1b. DB API Sandboxing

Focus: `db_api.go` (36KB), `db_api_bulk.go` (6KB), `db_conditions.go` (9KB).

Assert that table names and column names are from the allowed set -- no SQL injection via Lua string manipulation. Assert row count limits on bulk operations. Assert that query conditions are well-formed.

#### 1c. Hook Engine Invariants

Focus: `hooks_api.go` (6KB), `hook_engine.go` (28KB), `hook_handlers.go` (5KB).

Assert that hook names match the allowed set. Assert execution order invariants. Assert that hook registration does not exceed per-plugin limits.

#### 1d. HTTP Bridge and Request Engine

Focus: `http_bridge.go` (36KB), `request_engine.go` (24KB), `http_request.go` (15KB), `request_api.go` (13KB), `request_ssrf.go` (3KB).

SSRF guards: pair assertion on URL validation (validate the URL, then validate the response didn't redirect to a blocked host). Request/response shape validation. Resource limits on HTTP operations from plugins.

#### 1e. Manager and Pool

Focus: `manager.go` (69KB), `pool.go` (17KB), `sandbox.go` (6KB), `coordinator.go` (6KB).

Plugin lifecycle preconditions on every public method. Pool checkout must enforce max wait time. Lua execution must respect context deadlines. This sub-phase comes last because manager methods call into all the subsystems above -- assertions in the subsystems inform what the manager can assume.

**Note on function length:** Many functions in this package (especially `manager.go`) will exceed the 70-line limit after assertions are added. Expect mandatory splits as part of the retrofit. This roughly doubles the scope for large files -- account for it.

**Key files (all sub-phases):**
- `manager.go` (69KB) -- plugin lifecycle
- `db_api.go` (36KB) -- database operations exposed to Lua
- `core_api.go` (23KB) -- core CMS operations exposed to Lua
- `request_engine.go` (24KB) -- HTTP request pipeline for plugins
- `http_bridge.go` (36KB) -- HTTP handler bridge for plugin routes
- `hook_engine.go` (28KB) -- hook execution engine
- `pool.go` (17KB) -- Lua VM pool, resource bounds
- `sandbox.go` (6KB) -- Lua sandbox configuration
- `lua_helpers.go` (6KB) -- Lua/Go type conversions

### 2. internal/service/ (40 files)

**Boundary type:** Business logic convergence point. All paths (router, TUI, plugin, CLI) flow through service methods.

**Why second:** Service methods are the narrowest point where domain invariants can be enforced once and protect all callers. A service method that doesn't validate its inputs relies on every caller doing it correctly. One missed caller path means a bug.

**Caller trust table:**

The service layer is reachable from multiple callers with different trust levels. The response to invalid input depends on who is calling:

| Caller | Trust | Invalid input response |
|--------|-------|----------------------|
| HTTP handler (post-validation) | Trusted | Panic -- handler should have validated. |
| TUI command (post-validation) | Trusted | Panic -- TUI should have validated. |
| Lua plugin (`core_api.go`) | Semi-trusted | Return error -- plugin code is third-party. |
| RemoteDriver (`internal/remote/`) | Semi-trusted | Return error -- data arrived over network. |
| Internal package (deploy, backup) | Trusted | Panic -- caller is first-party code. |

**Consequence:** Service methods that are reachable from both trusted and semi-trusted callers must **return errors** (the safer default). Only panic in methods exclusively called from trusted, post-validation code paths. When in doubt, return an error.

Before retrofitting this package, enumerate the actual call paths for each public service method. A table mapping `method -> callers -> trust level` will make the panic-vs-error decision concrete.

**What to assert:**

- **Preconditions on every public method**: Valid IDs, non-empty required fields, legal state transitions. Panic or return error per the caller trust table above.
- **Postconditions on structural mutations**: After tree moves, verify parent pointers. After publish, verify status changed. After permission assignment, verify the junction row exists. **Important:** Until real transactions are implemented, postcondition read-backs must log warnings (`slog.Warn`), not panic. Without transactions, concurrent requests can modify rows between write and read-back, causing false-positive panics. See MODULA_STYLE.md for details.
- **Cross-entity consistency**: Content creation requires a valid datatype. Field creation requires a valid content parent. Route creation requires a valid content target. Assert these relationships exist before proceeding, not after the DB constraint fails with a cryptic FK error.
- **State machine transitions**: Content status (draft/published/archived), plugin state (disabled/enabled/error). Assert the transition is legal.

**Key files:**
- `content.go` (16KB), `content_admin.go` (16KB) -- content CRUD, tree operations
- `schema.go` (34KB) -- datatype/field schema management
- `routes.go` (15KB) -- route management
- `users.go` (15KB) -- user management with password hashing
- `rbac.go` (14KB) -- role/permission management
- `auth.go` (13KB) -- authentication flows
- `plugins.go` (14KB) -- plugin lifecycle management
- `media.go` (17KB) -- media upload/optimization
- `content_heal.go` (13KB) -- tree healing (structural repair)
- `locales.go` (23KB) -- locale management
- `webhooks.go` (12KB) -- webhook management

### 3. internal/router/ (handlers)

**Boundary type:** External HTTP input/output.

**Why third:** HTTP handlers are the primary external boundary. JSON unmarshaling provides some type safety, but path parameters, query parameters, and the shape of nested JSON bodies are unchecked.

**What to assert:**

- **Input validation**: Path params parsed and validated before use. Query params bounds-checked (pagination limits, sort field allowlists). Request bodies size-limited via `MaxBytesReader`.
- **Auth context**: Assert that middleware has populated the auth context before handler code accesses it. This catches route registration bugs where a handler is wired without its middleware chain.
- **Response postconditions**: In debug/test mode, assert that response bodies match the expected shape (no nil fields where the API contract says non-null, no leaked internal error messages).
- **Consistent error format**: All error responses go through a single path that enforces the error response shape.

**Key files:** All handler files in `internal/router/`. Focus on the mux registration (`mux.go`) to verify middleware chains are complete.

### 4. internal/model/ (6 files)

**Boundary type:** Structural invariants (content tree).

**Why fourth:** Small package, high leverage. Tree pointer corruption cascades silently. Six files means this can be done in a single pass.

**What to assert:**

- **Tree build invariants**: After building a tree from flat rows, assert: every node's parent exists in the tree (or is root). Every first_child's parent points back. Every next_sibling's prev_sibling points back. No cycles (visited set during traversal).
- **Mutation postconditions**: After insert, move, or reorder, re-verify the affected subtree's pointer integrity. These are in-memory operations on the tree struct, not DB read-backs, so the transaction caveat does not apply.
- **Node validity**: A node must have a valid ID. A non-root node must have a valid parent. Sort order must be non-negative.

**Key files:**
- `model.go` (8KB) -- types and tree structure
- `build.go` (7KB) -- tree construction from flat rows

### 5. internal/db/ (wrapper methods)

**Boundary type:** Persistence boundary (Go types to/from SQL).

**Why fifth:** The typed ID system (`db/types/`) provides compile-time safety on which ID goes where. The gap is runtime safety on what's in the values -- the mapper functions that convert between sqlc types and application types.

**Prerequisite:** Rename the `Assert*` functions in `convert.go` (see step 0b above). Do not retrofit this package while functions named "Assert" silently swallow type mismatches.

**What to assert:**

- **Mapper postconditions**: After mapping a sqlc row to an application type, assert the ID is valid (non-empty, 26 chars for ULIDs). Assert timestamps are non-zero. Assert required string fields are non-empty.
- **Write preconditions**: Before passing params to sqlc, assert the ID is valid, required fields are present. This is the pair to the mapper postcondition.
- **NULL conversion safety**: The `convert.go` helpers (`NullStringToString`, etc.) are trusted interior code. Assert at the call sites that the conversion result makes sense in context (e.g., a parent_id that came back as empty string when the row is not a root node).
- **Type width boundaries**: SQLite uses int64, MySQL/PostgreSQL use int32. The wrapper methods cast between them. Assert no truncation on the int64-to-int32 path.

**Key files:** All `*_gen.go` and `*_custom.go` wrapper files. `convert.go` for NULL helpers. `types/` for ID validation methods (these may need to be extended).

### 6. internal/validation/ (5 files)

**Boundary type:** Input validation rules.

**Why sixth:** This package is already assertion-like -- it exists to validate. The retrofit here is not adding assertions to the validation package, but auditing that validation is called at every entry point.

**What to do:**

- **Coverage audit**: For every field type in `type_validators.go` and every rule in `rules.go`, trace the call paths to verify they are invoked from all public-facing input paths: HTTP handler params, TUI input, and plugin API params.
- **Negative space**: For each validation rule, verify that tests cover both valid and invalid inputs. The validation test file is 48KB which suggests good coverage, but confirm edge cases at type boundaries (empty strings, max-length strings, zero values, negative numbers).

**Exit criterion:** All public-facing input paths (HTTP handlers, TUI commands, plugin API entry points) call validation for every user-supplied value. Verified by a call-path trace document listing each input path and its validation coverage.

**Key files:**
- `validate.go` (2KB) -- entry point
- `rules.go` (10KB) -- validation rules
- `type_validators.go` (4KB) -- per-type validators
- `validation_test.go` (48KB) -- test coverage

### 7. internal/middleware/ (auth wiring)

**Boundary type:** Trust transition (unauthenticated to authenticated, unprivileged to privileged).

**Why last:** The middleware already implements fail-closed semantics (missing PermissionSet = 403). The retrofit here is adding pair assertions: middleware asserts it set the context value, handler asserts the context value exists. This catches route wiring bugs.

**What to assert:**

- **Context population**: Each middleware that sets a context value (session, user, permissions) asserts the value is non-nil before setting it.
- **Context consumption**: Each handler or downstream middleware that reads a context value asserts it exists. Currently some handlers check and return 403, but a panic on missing auth context is more appropriate since it indicates a wiring bug, not a user error.
- **PermissionCache refresh**: Assert the cache is non-nil and was loaded within the refresh interval. A stale or nil cache is a programmer error.

**Key files:**
- `authorization.go` -- RBAC middleware, PermissionCache
- `sessions.go` -- session middleware
- `audit.go` -- audit logging middleware

## Boundaries Not In Scope

### internal/tui/ (40+ files)

The SSH TUI is another external input boundary -- users type commands that manipulate content and schema. The TUI is **not a separate retrofit target** because it calls service methods, and the service layer assertions (package #2) cover the trust boundary. The TUI is a trusted caller (same trust level as HTTP handlers post-validation). If TUI code passes invalid data to the service layer, the service preconditions will catch it.

If the TUI is later found to have input paths that bypass the service layer entirely (direct DB calls), it should be added to this plan.

### internal/remote/

The RemoteDriver is `DbDriver` over HTTPS. It is a semi-trusted caller of the DB interface. The service layer caller trust table (package #2) already accounts for it. The RemoteDriver itself should validate responses from the network, but this is part of its existing error handling, not a ModulaStyle retrofit.

## How to Retrofit a Package

1. Read MODULA_STYLE.md.
2. Read the package. Identify every function that sits at a boundary (accepts external input, returns external output, reads/writes persistent state, crosses a trust boundary).
3. For each boundary function, determine the caller trust level (see the trust table in MODULA_STYLE.md). Decide panic vs return error accordingly.
4. For each boundary function, add:
   - Precondition assertions (validate inputs at the top).
   - Postcondition assertions (verify outputs before returning), where the cost of a read-back is justified. Use read-backs for structural pointers and security state. Skip for simple leaf CRUD. Log warnings instead of panicking until real transactions are implemented.
5. For structural invariants (tree, permissions), add pair assertions (validate before write, validate after read).
6. Run `just test` to verify assertions don't fire on existing tests. If they do, that's a bug -- fix the bug, not the assertion.
7. Add test cases that exercise the negative space: invalid inputs, impossible states, boundary values. Use the `requirePanic` test helper (see MODULA_STYLE.md) for tests that verify expected panics.
8. Run `just check` to compile-verify.
9. If adding assertions pushed functions over 70 lines, split them. Keep control flow in the parent, push non-branching logic into helpers.

## Tracking

This is not a sprint. Check off items as they are completed:

- [ ] **Step 0:** Panic recovery middleware (`internal/middleware/recovery.go`)
- [ ] **Step 0b:** Rename `convert.go` Assert functions
- [ ] internal/plugin/ phase 1a: Lua boundary conversions
- [ ] internal/plugin/ phase 1b: DB API sandboxing
- [ ] internal/plugin/ phase 1c: Hook engine invariants
- [ ] internal/plugin/ phase 1d: HTTP bridge and request engine
- [ ] internal/plugin/ phase 1e: Manager and pool
- [ ] internal/service/ (with caller trust table)
- [ ] internal/router/
- [ ] internal/model/
- [ ] internal/db/
- [ ] internal/validation/ (audit)
- [ ] internal/middleware/
