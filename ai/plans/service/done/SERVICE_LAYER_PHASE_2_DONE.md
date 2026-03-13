──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 Ready to code?

 Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Service Layer Phase 2: ContentService

 Parent: SERVICE_LAYER_ROADMAP.md
 Assumes: Phase 0 (foundation) complete. Phase 1 (SchemaService) will be done separately.

 ---
 Context

 Content is the largest and most complex domain in ModulaCMS (~60 DbDriver methods, ~25 API handler files, ~29 MCP tools). The sibling-pointer tree manipulation logic is copy-pasted across 4
  locations (API public, API admin, HTMX admin, tree save handler) with inconsistent behavior — the HTMX admin handlers skip sibling repair on delete and ignore position on move. Extracting
 this into a service layer eliminates duplication, enforces consistent tree integrity, and gives all three consumers (admin panel, REST API, MCP) identical business logic.

 ---
 Architecture

 Tree Package (internal/tree/)

 A standalone package containing the sibling-pointer tree algorithms. No dependency on db, service, or config — pure logic against a generic Backend interface parameterized by ID type. This
 preserves the compile-time type safety of types.ContentID vs types.AdminContentID.

 package tree

 type Backend[ID ~string] interface {
     GetNode(ctx context.Context, id ID) (*Node[ID], error)
     UpdatePointers(ctx context.Context, ac audited.AuditContext, id ID, ptrs Pointers[ID]) error
 }

 type Node[ID ~string] struct {
     ID            ID
     ParentID      NullableID[ID]
     FirstChildID  NullableID[ID]
     NextSiblingID NullableID[ID]
     PrevSiblingID NullableID[ID]
 }

 type NullableID[ID ~string] struct {
     Value ID
     Valid bool
 }

 All algorithm functions are generic over the ID type:

 func Unlink[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], node *Node[ID]) error
 func Move[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], params MoveParams[ID]) (*MoveResult[ID], error)
 // etc.

 This means ContentService implements tree.Backend[types.ContentID] and AdminContentService implements tree.Backend[types.AdminContentID]. A ContentID cannot be accidentally passed to an
 admin tree operation — the compiler rejects it.

 Performance note: UpdatePointers hides a read-modify-write pattern. The existing UpdateContentData requires all 12 fields, not just pointers. Each UpdatePointers implementation must fetch
 the current row, merge pointer changes, then call the full update. A Reorder of N nodes = N Gets + N Updates + 1 parent update. This matches the current handler behavior. A future
 optimization would be adding a pointer-only UPDATE query to DbDriver, but that is out of scope for Phase 2.

 Algorithms (each is a standalone generic function taking Backend[ID]):
 - Unlink(ctx, ac, b, node) — detach node from sibling chain (3 steps: repair prev, repair next, repair parent first_child)
 - AppendChild(ctx, ac, b, parentID, childID) — walk to end of sibling chain, link
 - InsertAt(ctx, ac, b, parentID, childID, position) — insert at specific position in chain
 - Move(ctx, ac, b, params) — cycle detection + unlink + insert (full algorithm from apiMoveAdminContentData)
 - Reorder(ctx, ac, b, parentID, orderedIDs) — rewrite all sibling pointers + parent first_child
 - DetectCycle(ctx, b, nodeID, proposedParentID) — walk parent chain upward
 - Save(ctx, ac, b, params) — multi-phase bulk: create with null pointers, remap IDs, update pointers, delete

 The tree package has its own test file using a mock Backend (in-memory map). This validates all algorithms independently of the database.

 Content Services

 Two separate service structs — ContentService (public) and AdminContentService (admin). Same method names, parallel types. Both implement tree.Backend for their respective DbDriver methods.

 internal/service/
     content.go              # ContentService struct, constructor, tree.Backend impl, CRUD
     content_admin.go        # AdminContentService struct, constructor, tree.Backend impl, CRUD
     content_fields.go       # ContentService field methods
     content_fields_admin.go # AdminContentService field methods
     content_publish.go      # ContentService publish/unpublish/schedule
     content_publish_admin.go # AdminContentService publish/unpublish/schedule
     content_versions.go     # ContentService version methods
     content_versions_admin.go # AdminContentService version methods
     content_batch.go        # ContentService batch update
     content_heal.go         # ContentService heal
     content_relations.go    # ContentService relation CRUD
     content_relations_admin.go # AdminContentService relation CRUD
     content_params.go       # Shared param/result types

 Both registered in Registry:

 // service.go additions
 type Registry struct {
     // ... existing fields ...
     Content      *ContentService
     AdminContent *AdminContentService
 }

 Service Dependencies

 Each service holds:

 type ContentService struct {
     driver     db.DbDriver
     mgr        *config.Manager
     dispatcher publishing.WebhookDispatcher
 }

 Services call publishing.PublishContent, publishing.BuildSnapshot, etc. directly — the publishing package is already well-factored. No need to absorb it; just call through.

 For schema lookups (field auto-creation on content create, heal logic), services call driver.ListFieldsByDatatypeID directly. When Phase 1 delivers SchemaService, these calls can migrate.
 No blocker.

 ---
 Sub-Phases

 2a: Tree Package + Service Skeleton + Error Mapping

 Scope: Build internal/tree/, define both service structs, implement tree.Backend[ID] on each, wire into Registry, add shared error mapping helper, and design the transaction boundary.

 New files:

 ┌───────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │               File                │                                                          Content                                                          │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tree/tree.go             │ Node[ID], NullableID[ID], Pointers[ID], Backend[ID] interface, SaveParams[ID], SaveResult, MoveParams[ID], MoveResult[ID] │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tree/unlink.go           │ Unlink[ID] — 3-step sibling chain detach                                                                                  │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tree/insert.go           │ AppendChild[ID], InsertAt[ID] — sibling chain splice                                                                      │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tree/move.go             │ Move[ID], DetectCycle[ID] — full move with cycle guard                                                                    │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tree/reorder.go          │ Reorder[ID] — rewrite all sibling pointers for ordered list                                                               │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tree/save.go             │ Save[ID] — multi-phase bulk (create/remap/update/delete)                                                                  │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tree/tree_test.go        │ Table-driven tests against in-memory mock Backend[string]                                                                 │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/service/content.go       │ ContentService struct, constructor, tree.Backend[types.ContentID] impl                                                    │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/service/content_admin.go │ AdminContentService struct, constructor, tree.Backend[types.AdminContentID] impl                                          │
 ├───────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/service/errors_http.go   │ HandleServiceError(w, r, err) — shared mapping from service errors to HTTP responses                                      │
 └───────────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Error mapping helper (errors_http.go):

 All 16 rewired handler files need to translate service errors to HTTP responses. Without a shared helper, each handler re-implements the mapping — replacing one form of duplication with
 another. The helper lives in the service package and handles both API (JSON) and admin panel (HTMX) response formats:

 func HandleServiceError(w http.ResponseWriter, r *http.Request, err error) {
     switch {
     case IsNotFound(err):     // 404
     case IsValidation(err):   // 422, JSON body with FieldErrors
     case IsConflict(err):     // 409
     case IsForbidden(err):    // 403
     default:                  // 500, log internal error
     }
 }

 For HTMX requests (HX-Request header present), errors are returned via HX-Trigger toast events instead of JSON. The helper detects this automatically.

 Transaction boundary design:

 Phase 0 convention 11 deferred WithTx to Phase 2. Multi-step tree operations (Move = ~5 DB calls, Reorder = ~N+1 calls, Save = many) are not atomic — an intermediate failure leaves the tree
  in a partially-modified state. The current handler code has this same problem, so this is inherited, not introduced.

 The Backend[ID] interface reserves the transaction API shape without requiring real transactions initially:

 // tree.go — optional transactional backend
 type TxBackend[ID ~string] interface {
     Backend[ID]
     WithTx(ctx context.Context, fn func(Backend[ID]) error) error
 }

 Algorithm functions accept Backend[ID]. If the caller passes a TxBackend[ID], it can wrap the call in WithTx. The tree functions themselves do not manage transactions — the service layer
 decides whether to wrap a tree operation in a transaction.

 Initial implementation: ContentService.WithTx is a no-op passthrough (calls fn(self)). When DbDriver gains a transactional variant (requires adding BeginTx to the interface and modifying
 sqlc wrappers), the service layer enables real transactions without changing any tree algorithm code.

 Known limitation: Until real transactions are wired, a mid-operation failure during Move/Reorder/Save can leave orphaned or double-linked nodes. The Heal operation (Phase 2d) recovers from
 this. The failure window matches the current handler behavior.

 Modified files:

 ┌─────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │            File             │                                                Change                                                │
 ├─────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/service/service.go │ Add Content *ContentService, AdminContent *AdminContentService to Registry; construct in NewRegistry │
 └─────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────┘

 No handler changes. Services exist but are unused until 2b.

 Verification: go build ./... and just test pass.

 ---
 2b: Content Data CRUD + Content Fields

 Scope: Implement Get/List/Create/Update/Delete on both services for content data and content fields. Create uses tree package for sibling splicing + auto-field creation. Delete uses tree
 package for sibling repair. Rewire all API handlers and admin handlers to call service.

 Methods on ContentService (AdminContentService mirrors with admin types):

 Content data:
 - Get(ctx, id) → *db.ContentData, error
 - List(ctx, params) → []db.ContentData, error
 - ListPaginated(ctx, params) → []db.ContentDataTopLevel, int64, error
 - Create(ctx, ac, params) → *db.ContentData, error — if ParentID is set: sibling splice via tree.AppendChild/tree.InsertAt; if no parent (root-level): skip splicing. Auto-create empty
 fields for all fields in the assigned datatype
 - Update(ctx, ac, params) → error — includes optimistic locking path (UpdateWithRevision)
 - Delete(ctx, ac, id) → error — tree.Unlink before delete

 Content fields:
 - GetField(ctx, id) → *db.ContentFields, error
 - ListFields(ctx, params) → []db.ContentFields, error
 - ListFieldsPaginated(ctx, params) → []db.ContentFields, int64, error
 - CreateField(ctx, ac, params) → *db.ContentFields, error — runs validation.ValidateField
 - UpdateField(ctx, ac, params) → error — runs validation.ValidateField
 - DeleteField(ctx, ac, id) → error

 New files:

 ┌──────────────────────────────────────────┬───────────────────────────────────┐
 │                   File                   │              Content              │
 ├──────────────────────────────────────────┼───────────────────────────────────┤
 │ internal/service/content_fields.go       │ ContentService field methods      │
 ├──────────────────────────────────────────┼───────────────────────────────────┤
 │ internal/service/content_fields_admin.go │ AdminContentService field methods │
 └──────────────────────────────────────────┴───────────────────────────────────┘

 Modified files (handler rewiring):

 ┌───────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────┐
 │                 File                  │                                              Change                                               │
 ├───────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/contentData.go        │ All api* functions → call svc.Content.* instead of db.ConfigDB(c) + direct DbDriver               │
 ├───────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/contentFields.go      │ Same pattern                                                                                      │
 ├───────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/adminContentData.go   │ → call svc.AdminContent.*                                                                         │
 ├───────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/adminContentFields.go │ Same pattern                                                                                      │
 ├───────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/handlers/content.go    │ CRUD handlers → call service. Fixes missing sibling repair on delete and missing position on move │
 ├───────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/mux.go                │ Pass svc to handler closures that need it                                                         │
 └───────────────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────────┘

 Key fix: HTMX admin ContentDeleteHandler currently does a raw delete without sibling repair. After migration, it calls svc.Content.Delete() which always runs tree.Unlink first. Same for
 ContentCreateHandler (now gets sibling splicing) and ContentMoveHandler (now respects position and does cycle detection).

 Handler pattern after migration (API example):
 func apiGetContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
     id := types.ContentID(r.PathValue("id"))
     cd, err := svc.Content.Get(r.Context(), id)
     if err != nil {
         return err // mapped to HTTP status by error handler
     }
     writeJSON(w, cd)
     return nil
 }

 Testing: Service-level tests against SQLite. Note: there are essentially no existing handler-level tests for content data operations in internal/router/. The new service-level tests
 establish baseline coverage for the first time — they are not supplementing existing tests. This means the service tests must be thorough: they are the primary verification that behavior is
  correct, not just that it hasn't changed.

 ---
 2c: Tree Operations + Publishing + Versioning + Scheduling

 Scope: Add Move, Reorder, SaveTree, Publish, Unpublish, Schedule, and version CRUD to both services. Rewire all remaining content handlers.

 Tree methods:
 - Move(ctx, ac, params) → *MoveResult, error — delegates to tree.Move
 - Reorder(ctx, ac, parentID, orderedIDs) → error — delegates to tree.Reorder
 - SaveTree(ctx, ac, params) → *tree.SaveResult, error — delegates to tree.Save
 - GetTree(ctx, routeID) → []TreeRow, error — thin wrapper on GetContentTreeByRoute

 Publishing methods (wraps internal/publishing/):
 - Publish(ctx, ac, contentID, locale, userID) → *db.ContentVersion, error — calls publishing.PublishContent
 - Unpublish(ctx, ac, contentID, locale, userID) → error — calls publishing.UnpublishContent
 - Schedule(ctx, contentID, publishAt) → error — validates future time, calls UpdateContentDataSchedule

 Version methods:
 - ListVersions(ctx, contentID) → []db.ContentVersion, error
 - GetVersion(ctx, versionID) → *db.ContentVersion, error
 - CreateVersion(ctx, ac, params) → *db.ContentVersion, error — snapshot + create + prune
 - DeleteVersion(ctx, ac, versionID) → error — rejects if published
 - RestoreVersion(ctx, ac, contentID, versionID) → *publishing.RestoreResult, error — calls publishing.RestoreContent

 New files:

 ┌────────────────────────────────────────────┬────────────────────────────────────────────────┐
 │                    File                    │                    Content                     │
 ├────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ internal/service/content_publish.go        │ ContentService publish/unpublish/schedule      │
 ├────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ internal/service/content_publish_admin.go  │ AdminContentService publish/unpublish/schedule │
 ├────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ internal/service/content_versions.go       │ ContentService version CRUD                    │
 ├────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ internal/service/content_versions_admin.go │ AdminContentService version CRUD               │
 └────────────────────────────────────────────┴────────────────────────────────────────────────┘

 Modified files (handler rewiring):

 ┌────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────┐
 │                File                │                                            Change                                            │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/contentTree.go     │ → svc.Content.SaveTree                                                                       │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/contentTreeGet.go  │ → svc.Content.GetTree                                                                        │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/publish.go         │ → svc.Content.Publish/Unpublish/Schedule                                                     │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/publish_admin.go   │ → svc.AdminContent.Publish/Unpublish/Schedule                                                │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/versions.go        │ → svc.Content.*Version*                                                                      │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/versions_admin.go  │ → svc.AdminContent.*Version*                                                                 │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/scheduler.go       │ → svc.Content.Publish / svc.AdminContent.Publish instead of direct publishing.* calls        │
 ├────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/handlers/content.go │ Remaining handlers: reorder, move, tree save, publish, unpublish, versions, restore, compare │
 └────────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────┘

 Scheduler migration: StartPublishScheduler currently takes raw (driver, cfg, dispatcher). After migration, it takes *service.Registry and calls svc.Content.Publish() /
 svc.AdminContent.Publish(). The audit context for scheduled publishes uses Registry.SystemAuditCtx(item.AuthorID, "scheduled-publish") — the authorID comes from the content_data row's
 AuthorID field (the user who created or last edited the content).

 ---
 2d: Batch + Heal + Relations

 Scope: Remaining content operations.

 Batch method:
 - BatchUpdate(ctx, ac, params) → *BatchResult, error — optional content_data update + field upserts with validation. Consolidates the logic from ContentBatchHandler.

 Heal method:
 - Heal(ctx, ac, dryRun bool) → *HealReport, error — 4-pass repair: content data IDs, content field IDs, duplicate fields, missing fields. Consolidates from apiHealContent.

 Relation methods:
 - CreateRelation(ctx, ac, params) → *db.ContentRelations, error
 - DeleteRelation(ctx, ac, id) → error
 - GetRelation(ctx, id) → *db.ContentRelations, error
 - ListRelationsBySource(ctx, sourceID) → []db.ContentRelations, error
 - ListRelationsByTarget(ctx, targetID) → []db.ContentRelations, error
 - ListRelationsBySourceAndField(ctx, sourceID, fieldID) → []db.ContentRelations, error
 - UpdateRelationSortOrder(ctx, ac, params) → error

 New files:

 ┌─────────────────────────────────────────────┬───────────────────────────────────┐
 │                    File                     │              Content              │
 ├─────────────────────────────────────────────┼───────────────────────────────────┤
 │ internal/service/content_batch.go           │ BatchUpdate method                │
 ├─────────────────────────────────────────────┼───────────────────────────────────┤
 │ internal/service/content_heal.go            │ Heal method                       │
 ├─────────────────────────────────────────────┼───────────────────────────────────┤
 │ internal/service/content_relations.go       │ ContentService relation CRUD      │
 ├─────────────────────────────────────────────┼───────────────────────────────────┤
 │ internal/service/content_relations_admin.go │ AdminContentService relation CRUD │
 └─────────────────────────────────────────────┴───────────────────────────────────┘

 Modified files:

 ┌─────────────────────────────────┬───────────────────────────┐
 │              File               │          Change           │
 ├─────────────────────────────────┼───────────────────────────┤
 │ internal/router/contentBatch.go │ → svc.Content.BatchUpdate │
 ├─────────────────────────────────┼───────────────────────────┤
 │ internal/router/contentHeal.go  │ → svc.Content.Heal        │
 └─────────────────────────────────┴───────────────────────────┘

 Note: Content relations currently have no HTTP handlers or MCP tools — only DbDriver methods. The service methods are added for completeness and to support future endpoint exposure. No
 handler rewiring needed for relations.

 ---
 Conventions

 1. Services return service.NotFoundError, service.ValidationError, service.ConflictError (from Phase 0 errors.go). Handlers map these to HTTP status codes (404, 422, 409).
 2. Optimistic locking: Update checks revision. Returns *service.ConflictError{Detail: "revision mismatch"} on conflict. Handlers return 409.
 3. publishing.* functions are called directly — no absorption. If publishing returns a string-prefixed conflict error, the service wraps it in *ConflictError.
 4. Tree operations always go through tree.Backend → tree.* functions. No direct sibling-pointer manipulation in service code.
 5. All mutating methods take audited.AuditContext as second parameter (after ctx).
 6. The tree package imports only internal/db/audited (for the AuditContext type in Backend.UpdatePointers). No other internal/ imports. The audited package is a leaf with no further
 dependencies — this is acceptable.
 7. ContentService.Create() conditionally splices into the sibling chain only when ParentID is set. Root-level creates (no parent) skip splicing entirely.
 8. Each sub-phase is a separate git commit that can be independently reverted if a behavioral regression is discovered.

 ---
 Files Summary

 New files (22):

 internal/tree/ (7):
 - tree.go — Node[ID], NullableID[ID], Pointers[ID], Backend[ID] interface, TxBackend[ID], SaveParams[ID], SaveResult, MoveParams[ID], MoveResult[ID]
 - unlink.go — Unlink (3-step sibling chain detach)
 - insert.go — AppendChild, InsertAt
 - move.go — Move, DetectCycle
 - reorder.go — Reorder
 - save.go — Save (multi-phase bulk)
 - tree_test.go — table-driven tests against in-memory mock Backend

 internal/service/ (15 — 1 replaced, 14 new):
 - content.go (replaced) — ContentService struct, constructor, tree.Backend[types.ContentID] impl, CRUD
 - content_admin.go — AdminContentService struct, constructor, tree.Backend[types.AdminContentID] impl, CRUD
 - content_fields.go — public field methods
 - content_fields_admin.go — admin field methods
 - content_publish.go — public publish/unpublish/schedule
 - content_publish_admin.go — admin publish/unpublish/schedule
 - content_versions.go — public version CRUD
 - content_versions_admin.go — admin version CRUD
 - content_batch.go — batch update
 - content_heal.go — heal
 - content_relations.go — public relation CRUD
 - content_relations_admin.go — admin relation CRUD
 - content_params.go — shared param/result types (BatchParams, BatchResult, HealReport, etc.)
 - errors_http.go — HandleServiceError(w, r, err) shared error-to-HTTP mapping

 Modified files (16):
 - internal/service/service.go — Registry additions
 - internal/router/contentData.go
 - internal/router/contentFields.go
 - internal/router/contentTree.go
 - internal/router/contentTreeGet.go
 - internal/router/contentBatch.go
 - internal/router/contentHeal.go
 - internal/router/adminContentData.go
 - internal/router/adminContentFields.go
 - internal/router/publish.go
 - internal/router/publish_admin.go
 - internal/router/versions.go
 - internal/router/versions_admin.go
 - internal/router/scheduler.go
 - internal/router/mux.go
 - internal/admin/handlers/content.go

 Not touched: MCP server (stays as HTTP client, deferred to Phase 7), internal/publishing/ (called from service, not modified), internal/db/ (no changes).

 ---
 Parallelization

 Sub-phases are sequential: 2a → 2b → 2c → 2d. Each depends on the prior sub-phase's service struct and methods. The overlap between 2b and 2c is minimal (2c's publishing methods call
 svc.Content.Get from 2b).

 Within each sub-phase, public and admin handler rewiring can run in parallel (different files).

 Recommended agent split for the largest sub-phases:
 - 2b: Agent A rewires API handlers (internal/router/content*.go, adminContent*.go), Agent B rewires admin handlers (internal/admin/handlers/content.go)
 - 2c: Agent A does publishing handlers, Agent B does version handlers + scheduler

 ---
 Verification

 After each sub-phase:
 1. go build ./... — compiles
 2. just test — all existing tests pass
 3. go vet ./... — no vet warnings

 After Phase 2 complete:
 4. Tree package tests (in-memory mock): all algorithms (unlink, insert, move with cycle detection, reorder, save with ID remapping) with edge cases (empty chain, single node, cycle at root,
  concurrent pointer states)
 5. Service-level tests (SQLite): CRUD with sibling splicing, field auto-creation, optimistic locking conflict, publish/unpublish, version create/delete/restore, batch update with
 validation, heal dry-run and apply
 6. Note: there are no existing handler-level tests for content operations. The service tests are establishing baseline correctness, not verifying behavioral equivalence. They must be
 comprehensive.
