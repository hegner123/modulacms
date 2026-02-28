# Review Notes: Phase 4 Plan + Admin Panel (2026-02-27)

Three-agent review of the Phase 4 production hardening plan (`~/.claude/plans/elegant-enchanting-tarjan.md`) and the `internal/admin/` HTMX panel.

---

## Phase 4 Plan Issues (Skeptic + Go Backend Reviews)

### Must Fix Before Implementing

1. **SyncCapabilities is not transactional.** Delete old pipelines + create new pipelines are separate DB calls. A crash or concurrent `LoadRegistry` (Watcher, Coordinator, admin API) between them sees zero pipelines for that plugin. Wrap in a DB transaction or hold the Manager write lock for the full duration.

2. **SyncCapabilities missing AuditContext parameter.** Method signature is `(ctx, name)` but audited operations need user ID, IP, etc. Follow the `ActivatePlugin(ctx, name, adminUser)` pattern — accept `adminUser string` and construct audit context internally.

3. **Validate that `manifest_hash` is populated in existing DB records.** `computeChecksum(dir)` is used for comparison but if `manifest_hash` was never stored (empty string in legacy records), drift detection fires false positives for every plugin on every load. Need a migration or first-run population step.

4. **DryRunAll key parsing must split on the last dot.** Keys are `"table.before_op"` — naive `strings.SplitN(key, ".", 2)` breaks if table names contain dots. Use `strings.LastIndex` or find first dot followed by `before_`/`after_`.

5. **Add cooldown to sync-capabilities endpoint.** Existing reload handler has 10s cooldown. Sync is heavier (DB writes, pipeline recreation, registry reload) but has none. Admin hammering the endpoint causes DB churn and temporary pipeline outages.

### Should Fix

6. **SyncCapabilities holds `m.mu` write lock during DB I/O.** Under slow/unreachable DB, this blocks all Manager reads. Consider: hold lock only for in-memory state mutation (steps 8-9), let DB ops run without lock. `LoadRegistry` uses build-then-swap so concurrent runs are safe.

7. **No concurrency guard on sync-capabilities.** Two simultaneous POSTs both delete and recreate pipelines. Need per-plugin sync lock or Manager write lock.

8. **No input validation for plugin not in DB yet.** A discovered-but-not-installed plugin causes `GetPluginByName` to return `sql.ErrNoRows`. Not handled in the plan.

9. **PipelineListHandler queries `m.driver` directly.** Only admin handler to bypass Manager. Add `ListAllPipelines()` method to Manager for consistency with existing patterns.

10. **Coordinator feedback cycle.** `SyncCapabilities` bumps `date_modified`. Coordinator on the *same instance* sees it next tick and calls `LoadRegistry` again (redundant). Not harmful but wasteful. Update `c.lastSeen[name].DateModified` after local sync.

11. **PipelineListHandler vs DryRunHandler data source confusion.** DB list includes disabled pipelines; dry-run shows only active in-memory state. Admin sees different results from the two endpoints with no explanation. Document or unify.

### Open Questions

- What happens to in-flight pipeline executions during `SyncCapabilities`?
- Should sync also re-run `on_init()` / `ExtractManifest` for metadata changes?
- Does the Coordinator watch pipeline table changes or only plugin status? Instance B may not reconcile after Instance A syncs.
- Multi-database test plan? `SyncCapabilities` hits `m.driver` methods that differ across SQLite/MySQL/PostgreSQL.

### Verified Accurate

- All file references and line numbers checked out within 1-5 lines of actual positions
- `ManifestDrift` confirmed declared but never set — real latent bug
- Dependency ordering (Gap 1 -> 2 -> 3) is correct
- Test list is comprehensive for new code
- TUI section correctly follows all existing patterns

---

## Admin Panel HTMX Issues

### Critical

1. **CSRF token injection has dead code + staleness risk.** `<body hx-headers='{"X-CSRF-Token": ""}'>` is dead code. Meta tag token only refreshes on full page loads; goes stale during long SPA sessions. Fix: remove dead `hx-headers`, add cookie-fallback to `htmx:configRequest`.

2. **Delete handlers return 200 empty body — row removal works by coincidence.** `outerHTML` swap with empty body removes elements only by accident. Detail page deletes with `hx-target="#main-content"` + `hx-swap="outerHTML"` **destroy the entire content area**. Fix: use `hx-swap="delete"` for table rows, `HX-Redirect` for detail pages.

3. **ContentUpdateHandler returns empty body — form destroyed after save.** Returns `showToast` + 200 with no body, but form targets `#content-edit-form` with `innerHTML` swap. User loses edit context after every save. Fix: `hx-swap="none"` on form or `HX-Redirect` back.

4. **Route edit forms missing `hx-target`/`hx-swap`.** Validation errors render form partial inside existing `<form>` — creates invalid nested forms. Fix: add explicit `hx-target` to a wrapper div.

### Important

5. **No `hx-indicator` or `hx-disabled-elt` anywhere.** No per-element loading feedback. Users can double-submit forms and click delete buttons multiple times.

6. **No OOB updates after mutations.** Deleting items doesn't update counts, pagination, or related panels. Stale data across panels.

7. **Media upload may be missing `hx-encoding="multipart/form-data"`.** Without it, HTMX serializes as URL-encoded and file data is silently dropped.

8. **`RoleUpdateHandler` calls `w.WriteHeader` twice** (`roles.go`:222-229). Headers set after first `WriteHeader` may not be sent. Toast notification lost on validation errors.

9. **`pageTitle` event handler reads `e.detail.value` but HTMX passes string as `e.detail` directly** (`admin.js`:433-437). Browser tab title never updates during SPA navigation.

### Minor

- Keyboard nav bypasses `hx-push-url` — sidebar active state doesn't update
- Login form has artificial 750ms delay
- Content search has no explicit `hx-swap`
- No `aria-live` regions for swapped content
- Content search doesn't push URL — results lost on back button
- External CDN for Lucide icons (`unpkg.com/lucide@latest`) — pin version, consider bundling

### What's Good

- `RenderNav`/`IsNavHTMX` pattern correctly differentiates sidebar nav from partial HTMX requests
- Buffer-first rendering with atomic error handling (`Render` sends `HX-Retarget` to `#none` on failure)
- `RenderWithOOB` centralizes OOB fragment wrapping; dialog containers sent via OOB during nav
- CSRF cookie reuse prevents classic HTMX desync where meta tag goes stale
- Consistent handler structure across all endpoints
