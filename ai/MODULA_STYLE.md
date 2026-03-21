# ModulaStyle

Adapted from [TigerStyle](https://tigerbeetle.com/blog/2023-03-28-tigerstyle) for Go. These rules apply to all new code and to any code being actively refactored. They are not a mandate to rewrite the codebase.

Our design goals are **safety**, **performance**, and **developer experience**. In that order. When they conflict: safety > performance > developer experience. None taken to extremes.

> "the simple and elegant systems tend to be easier and faster to design and get right, more
> efficient in execution, and much more reliable" -- Edsger Dijkstra

## Boundaries

ModulaStyle is primarily concerned with **boundaries** -- places where data enters, exits, or changes shape. These are where bugs live:

- **External input**: HTTP requests, Lua plugin values, SSH commands, config files.
- **External output**: HTTP responses, database writes, plugin return values.
- **Trust transitions**: Untrusted (user/plugin) to trusted (service/db). Trusted to untrusted (API responses that might leak internal state).
- **Shape transitions**: JSON to Go struct. Go struct to SQL params. SQL row to Go struct. Go value to Lua value.
- **Structural invariants**: Tree pointer consistency, sibling link bidirectionality, parent/child integrity.

Code in the interior -- pure transformations on already-validated data -- benefits less from these rules. Focus effort where data crosses a boundary.

## Assertions

> "Specifically, we found that almost all (92%) of the catastrophic system failures are the result
> of incorrect handling of non-fatal errors explicitly signaled in software."
> -- [Yuan et al., OSDI '14](https://www.usenix.org/system/files/conference/osdi14/osdi14-paper-yuan.pdf)

### Recovery Middleware (Prerequisite)

Assertion panics in request-scoped code must be caught by recovery middleware that converts them to 500 responses. Without this, a single assertion failure crashes the entire process. The recovery middleware in `internal/middleware/recovery.go` must be wired into every HTTP chain before any assertion panics are added to handler or service code.

### Programmer Errors vs Operating Errors

Go distinguishes these already: `error` for operating errors (expected, handled), `panic` for programmer errors (unexpected, crash). ModulaStyle reinforces the distinction:

- **Operating errors** (network timeout, invalid user input, missing record): Return `error`. Handle at every call site. Never ignore with `_`.
- **Programmer errors** (nil where non-nil is required, violated invariant, impossible state): `panic` with a descriptive message. The only correct response to corrupt program state is to crash. In request-scoped code, the recovery middleware converts panics to 500s, but the intent is the same -- this request is unsalvageable.

### Caller Trust Levels

The service layer is reachable from multiple callers with different trust levels. Whether a service method panics or returns an error on bad input depends on who is calling:

| Caller | Trust level | Invalid input response |
|--------|-------------|----------------------|
| HTTP handler (post-validation) | Trusted | Panic -- handler should have validated. |
| TUI command (post-validation) | Trusted | Panic -- TUI should have validated. |
| Lua plugin (`core_api.go`) | Semi-trusted | Return error -- plugin code is third-party. |
| RemoteDriver (`internal/remote/`) | Semi-trusted | Return error -- data arrived over the network. |
| Internal package (e.g., deploy, backup) | Trusted | Panic -- caller is first-party code. |

For service methods reachable by both trusted and semi-trusted callers, **return errors** (the safer default). Only panic in service methods that are exclusively called from trusted, post-validation code paths. When in doubt, return an error.

### Preconditions

Assert function inputs at the boundary. A function must not operate blindly on data it has not checked.

```go
// Good: preconditions at the top.
func (s *Service) MoveContent(ctx context.Context, id types.ContentID, targetParentID types.ContentID) error {
    if !id.Valid() {
        panic("MoveContent: invalid content ID")
    }
    if !targetParentID.Valid() {
        panic("MoveContent: invalid target parent ID")
    }
    if id == targetParentID {
        panic("MoveContent: cannot move content to itself")
    }
    // ...
}
```

For functions that accept external input (handlers, plugin API), return errors instead of panicking -- the caller is untrusted:

```go
// Good: handler validates and returns an error response.
func (h *Handler) handleMoveContent(w http.ResponseWriter, r *http.Request) {
    id, err := types.ParseContentID(r.PathValue("id"))
    if err != nil {
        respondError(w, http.StatusBadRequest, "invalid content ID")
        return
    }
    // ...
}
```

The rule: **panic on programmer errors (broken invariants inside trusted code), return errors on operating errors (bad input from untrusted sources).**

### Postconditions

After a critical operation, verify the result before returning it.

```go
// Good: verify structural integrity after tree mutation.
func (s *Service) MoveContent(ctx context.Context, id, targetID types.ContentID) error {
    // ... perform the move ...

    // Postcondition: verify the move actually took effect.
    moved, err := s.driver.GetContentData(ctx, id)
    if err != nil {
        return fmt.Errorf("MoveContent postcondition: failed to read moved node: %w", err)
    }
    if moved.ParentID.String() != targetID.String() {
        panic(fmt.Sprintf("MoveContent postcondition: parent_id is %s, expected %s", moved.ParentID, targetID))
    }
    return nil
}
```

Postconditions are not free. Use them at structural boundaries (tree mutations, permission changes, publish operations) where silent corruption is worse than the cost of a read-back.

**Transaction caveat:** Postcondition read-backs are only meaningful when the write and read-back are atomic -- either inside a transaction or operating on a single-writer resource. Without a transaction, another request can modify the row between the write and the read-back, causing a false-positive panic on valid concurrent behavior. Until real transactions are implemented, postcondition read-backs should **log a warning** (via `slog.Warn`), not panic. This caveat applies to the service layer; the db wrapper layer's pair assertions (validate before write, validate after read) do not have this problem because they operate on the same row within the same function call.

**Cost heuristic:** Use read-back postconditions when the operation modifies structural pointers (tree parent/child/sibling), security-sensitive state (permissions, roles), or publish status. Skip read-backs for simple CRUD on leaf entities (create/update a field, update a route label). For networked backends (MySQL/PostgreSQL), a read-back is a round trip -- factor that into the decision.

### Pair Assertions

For every property you want to enforce, assert it in at least two code paths. The classic pair: validate before writing, validate after reading.

```go
// Before write: assert the data we are about to persist.
func (d Database) CreateDatatype(ctx context.Context, params CreateDatatypeParams) (Datatypes, error) {
    if !params.DatatypeID.Valid() {
        panic("CreateDatatype: invalid ID")
    }
    if params.Label == "" {
        panic("CreateDatatype: empty label")
    }
    // ... write to db ...
}

// After read: assert the data we just loaded.
func MapDatatype(row mdb.Datatypes) Datatypes {
    if row.DatatypeID == "" {
        panic("MapDatatype: empty ID from database")
    }
    // ... map fields ...
}
```

### Split Assertions

Prefer separate assertions over compound ones. Each assertion should test one thing so failures are precise.

```go
// Good: separate, each failure is specific.
if !id.Valid() {
    panic("invalid content ID")
}
if parentID == id {
    panic("content cannot be its own parent")
}

// Bad: compound, which condition failed?
if !id.Valid() || parentID == id {
    panic("invalid content or self-referencing parent")
}
```

### Assertion Density

Aim for at least two assertions per function at trust boundaries. Interior helper functions that operate on already-validated data need fewer. This is a guideline, not a counting exercise -- the goal is that functions do not operate blindly.

### Assertions in Tests

Go panics crash the entire test binary. Tests that exercise invalid inputs need to handle expected panics explicitly.

Use a `requirePanic` test helper:

```go
func requirePanic(t *testing.T, name string, fn func()) {
    t.Helper()
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("%s: expected panic, got none", name)
        }
    }()
    fn()
}
```

Usage:

```go
func TestMoveContent_SelfReference(t *testing.T) {
    // Verify that moving content to itself panics (programmer error).
    id := types.NewContentID()
    requirePanic(t, "self-reference", func() {
        svc.MoveContent(ctx, id, id)
    })
}
```

Tests that verify happy paths should not use `recover()`. If an assertion fires during a happy-path test, that is a real bug -- let it crash and fix the root cause.

## Control Flow

### Early Returns

Use early returns to shed invalid states at the top. This is idiomatic Go and reduces nesting. Do not use `else` after a `return`.

```go
// Good.
func (s *Service) GetContent(ctx context.Context, id types.ContentID) (Content, error) {
    if !id.Valid() {
        return Content{}, ErrInvalidID
    }
    content, err := s.driver.GetContentData(ctx, id)
    if err != nil {
        return Content{}, fmt.Errorf("GetContent: %w", err)
    }
    return content, nil
}
```

### Split Compound Conditions

When a branch depends on multiple booleans, split into separate checks so the reader can verify each case.

```go
// Good: each condition is clear.
if !user.Active {
    return ErrInactive
}
if !permSet.Has("content:create") {
    return ErrForbidden
}

// Bad: compound.
if !user.Active || !permSet.Has("content:create") {
    return ErrForbidden
}
```

### State Invariants Positively

Prefer checking the positive condition. Negations are harder to reason about.

```go
// Good: positive framing.
if index < length {
    // Invariant holds.
} else {
    // Invariant violated.
}

// Harder to read.
if index >= length {
    // Not true that the invariant holds.
}
```

## Limits

Put a limit on everything. Unbounded operations are unbounded risk.

- **Loops**: Every loop that processes external data must have a maximum iteration count or a context deadline.
- **Channels**: Buffered channels must have an explicit capacity based on expected load, not a magic number.
- **Goroutines**: Worker pools must be bounded. Document the bound and why.
- **Query results**: Paginate. Never `SELECT *` without `LIMIT` in production code paths.
- **Plugin execution**: Time-boxed via context deadlines. Memory-bounded via pool limits.
- **Request bodies**: `http.MaxBytesReader` on every handler that reads a body.

Where a loop is intentionally unbounded (event loop, server accept loop), document why and ensure it respects a shutdown signal.

## Functions

### Length

Hard limit: **70 lines per function** for new code. This is the point where a function stops fitting on a screen. When splitting:

- Keep control flow (switch/if) in the parent function. Push non-branching logic into helpers.
- Keep state manipulation in the parent. Let helpers compute what needs to change, not apply it.
- Leaf functions should be pure where possible.

Existing functions over 70 lines are not a priority to split, but should be split when actively modified.

### Signatures

- Fewer parameters, simpler return types. If a function takes more than 4 parameters, consider a params struct (we already do this for Create/Update operations).
- `context.Context` is always the first parameter for functions that do I/O.
- Callbacks and function parameters go last.

## Scope

- Declare variables at the smallest possible scope.
- Calculate or check values close to where they are used. Do not introduce variables before they are needed. The distance between check and use is where bugs hide (POCPOU -- place-of-check to place-of-use).
- Do not duplicate state. One source of truth. If two variables must stay in sync, one of them should not exist.

## Resource Cleanup

Group resource allocation with its deferred cleanup. Newline before the allocation, `defer` immediately after.

```go
func readConfig(path string) (Config, error) {

    f, err := os.Open(path)
    if err != nil {
        return Config{}, fmt.Errorf("open config: %w", err)
    }
    defer f.Close()

    // ... use f ...
}
```

For write handles where the close error matters:

```go
func writeBackup(path string, data []byte) (err error) {

    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create backup: %w", err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = cerr
        }
    }()

    // ... write to f ...
}
```

## Comments

- Comments are sentences. Capital letter, full stop. Space after `//`.
- Inline end-of-line comments can be phrases without punctuation.
- **Always say why.** Code shows what. Comments show why. If the why is obvious from the what, no comment is needed.
- Test functions get a comment at the top explaining the goal and methodology.

## Errors

- All errors must be handled. No `_` for error returns.
- Wrap errors with context: `fmt.Errorf("operation context: %w", err)`.
- Compare errors with `errors.Is` / `errors.As`, not `==`.
- At trust boundaries, translate internal errors to safe external representations. Never leak stack traces, internal paths, or SQL errors to API clients.

## Naming

Follow Go conventions (PascalCase exported, camelCase unexported), but apply TigerStyle principles within those constraints:

- **No abbreviations** except established Go idioms (`ctx`, `err`, `req`, `resp`, `cfg`).
- **Units and qualifiers last**, sorted by descending significance: `timeoutMs`, `limitMax`, `retryCountMax` -- not `maxTimeout`, `maxLimit`, `maxRetryCount`.
- **Do not overload names.** If a term means something in one context, do not reuse it with a different meaning elsewhere.
- **Prefer nouns** for identifiers that will appear in documentation or logs. `pipeline` over `preparing`.

## Performance

Think about performance from the outset. The biggest wins come from design, not profiling.

- Optimize for the slowest resource first: network > disk > memory > CPU.
- Batch operations where possible. One query returning 50 rows beats 50 queries returning 1 row.
- Distinguish control plane (admin operations, config changes) from data plane (content delivery). The data plane is the hot path.
- Be explicit about allocations in hot paths. Reuse slices. Pre-size maps.

## What This Document Is Not

- **Not a mandate to rewrite.** Existing code is refactored to ModulaStyle when it is being actively modified. We are pre-v1 -- features ship first.
- **Not TigerStyle verbatim.** We do not adopt: static-only memory allocation, zero dependencies, no recursion, snake_case, Zig-specific patterns. These are TigerBeetle constraints that do not apply to a Go CMS.
- **Not absolute.** Every rule has exceptions. When a rule conflicts with shipping, shipping wins -- but document the exception with a comment explaining why, so the debt is visible.

## Attribution

Inspired by [TigerStyle](https://github.com/tigerbeetle/tigerbeetle/blob/main/docs/TIGER_STYLE.md) from the TigerBeetle project. Adapted for Go and the ModulaCMS domain.
