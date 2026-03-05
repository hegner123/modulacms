╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Service Layer Phase 1: SchemaService

 Context

 Three consumers (admin panel, REST API, MCP server) independently implement CRUD + business logic for datatypes, fields, field types, and datatype-field associations. This results in
 duplicated validation, audit context construction, and default-value logic across ~15 handler files. Phase 1 extracts this into a unified SchemaService in internal/service/schema.go,
 proving the service layer pattern established in Phase 0.

 Scope: Datatypes, Fields, FieldTypes (both public and admin variants). MCP tools are excluded (they go through the Go SDK over HTTP — that's Phase 7).

 ---
 Architecture Decisions

 Single Service, Not Split

 The admin and public DbDriver methods operate on different entity types (AdminDatatypes vs Datatypes) with different ID types — no branching logic needed, just separate methods namespaced
 by prefix (CreateDatatype vs CreateAdminDatatype). A single struct avoids duplicating the dependency injection.

 Pass-through db.*Params types

 Service methods accept db.CreateDatatypeParams, db.UpdateFieldParams, etc. directly — no service-layer param types. The handlers already construct these types today. The service adds
 validation and defaults on top. If a future phase requires input validation that conflicts with the database param shape (e.g., fields the caller must not set, or transforms before persistence), introduce service-layer param types at that point.

 Composable store interfaces

 Split the monolithic SchemaStore into embedded sub-interfaces for testability:

 type DatatypeStore interface { ... }      // ~12 methods
 type FieldStore interface { ... }         // ~10 methods
 type FieldTypeStore interface { ... }     // ~7 methods
 type AdminDatatypeStore interface { ... } // ~7 methods
 type AdminFieldStore interface { ... }    // ~9 methods
 type AdminFieldTypeStore interface { ... }// ~7 methods

 type SchemaStore interface {
     DatatypeStore
     FieldStore
     FieldTypeStore
     AdminDatatypeStore
     AdminFieldStore
     AdminFieldTypeStore
 }

 Only include methods that in-scope consumers (admin panel, REST API) actually call. Exclude TUI-only methods (ListDatatypesRoot, ListDatatypeChildren, ListDatatypeChildrenPaginated) — they
 can be added when the TUI migrates.

 ---
 Step 1: Define SchemaService struct and store interfaces

 File: internal/service/schema.go

 type SchemaService struct {
     store      SchemaStore
     fullDriver db.DbDriver
 }

 func NewSchemaService(store SchemaStore, driver db.DbDriver) *SchemaService {
     return &SchemaService{store: store, fullDriver: driver}
 }

 The SchemaStore interface (composed from sub-interfaces above) includes only methods called by admin panel or REST API handlers. The full DbDriver satisfies it implicitly.

 **GetDatatypeFull dependency:** `AssembleDatatypeFullView` in `internal/db/assemble.go` requires `DbDriver` methods outside SchemaStore's scope (notably `GetUser`). Add a `fullDriver db.DbDriver` field to SchemaService, set from the same driver passed to `NewSchemaService`. `GetDatatypeFull` delegates to `db.AssembleDatatypeFullView(s.fullDriver, id)`. Do not use type assertions or inline the assembly logic.

 Wire into Registry (internal/service/service.go):
 // Add to Registry struct:
 Schema *SchemaService

 // In NewRegistry, after constructing:
 reg.Schema = NewSchemaService(driver, driver)

 ---
 Step 2: Implement SchemaService methods

 Each method: validate → set defaults → call store → wrap errors → return.

 Error wrapping strategy

 ┌────────────────────────────────┬───────────────────────────────────────────────────────┐
 │          Store error           │                     Service error                     │
 ├────────────────────────────────┼───────────────────────────────────────────────────────┤
 │ Returns nil entity (not found) │ *NotFoundError{Resource: "datatype", ID: id.String()} │
 ├────────────────────────────────┼───────────────────────────────────────────────────────┤
 │ Unique constraint violation    │ *ConflictError{Resource, ID, Detail}                  │
 ├────────────────────────────────┼───────────────────────────────────────────────────────┤
 │ Any other DB error             │ *InternalError{Err: fmt.Errorf("context: %w", err)}   │
 ├────────────────────────────────┼───────────────────────────────────────────────────────┤
 │ Validation failures            │ *ValidationError with collected []FieldError          │
 └────────────────────────────────┴───────────────────────────────────────────────────────┘

 **Unique constraint detection:** ConflictError mapping is deferred from Phase 1. SQLite, MySQL, and PostgreSQL surface unique violations differently (SQLITE_CONSTRAINT_UNIQUE, MySQL error 1062, PostgreSQL SQLSTATE 23505). Until a cross-backend `db.IsUniqueViolation(err) bool` helper is added in a future phase, unique constraint violations will surface as `*InternalError`. Add `db.IsUniqueViolation` to the Phase 0 backlog.

 Public Datatype Methods

 ┌────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │         Method         │                     Signature                      │                                              Business Logic                                              │
 ├────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ListDatatypes          │ (ctx) ([]Datatypes, error)                         │ List all, nil-safe                                                                                       │
 ├────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ GetDatatype            │ (ctx, DatatypeID) (*Datatypes, error)              │ NotFoundError if nil                                                                                     │
 ├────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ GetDatatypeFull        │ (ctx, DatatypeID) (*DatatypeFullView, error)       │ Delegates to db.AssembleDatatypeFullView                                                                 │
 ├────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ CreateDatatype         │ (ctx, ac, CreateDatatypeParams) (*Datatypes,       │ Validate: label required, type required, ValidateUserDatatypeType. Defaults: generate ID if zero, set    │
 │                        │ error)                                             │ timestamps if zero                                                                                       │
 ├────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ UpdateDatatype         │ (ctx, ac, UpdateDatatypeParams) (*Datatypes,       │ Validate type if non-empty. Fetch existing first to preserve immutable fields (ParentID, AuthorID,       │
 │                        │ error)                                             │ DateCreated). Update, then re-fetch and return                                                           │
 ├────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ DeleteDatatype         │ (ctx, ac, DatatypeID) error                        │ Validate ID format                                                                                       │
 ├────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ListDatatypesPaginated │ (ctx, PaginationParams)                            │ Paginate + count                                                                                         │
 │                        │ (PaginatedResponse[Datatypes], error)              │                                                                                                          │
 └────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Note on UpdateDatatype: The admin handler currently fetches the existing entity to preserve ParentID, AuthorID, DateCreated. The service's UpdateDatatype should do this fetch internally —
 the handler passes in user-editable fields only, and the service merges them with the existing entity's immutable fields before calling store.UpdateDatatype. This eliminates the pre-fetch
 from every handler.

 Public Field Methods

 ┌────────────────────────┬─────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │         Method         │                    Signature                    │                                               Business Logic                                                │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ListFields             │ (ctx) ([]Fields, error)                         │ Nil-safe                                                                                                    │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ListFieldsFiltered     │ (ctx, roleID string, isAdmin bool) ([]Fields,   │ List + db.FilterFieldsByRole                                                                                │
 │                        │ error)                                          │                                                                                                             │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ GetField               │ (ctx, FieldID, roleID string, isAdmin bool)     │ Get + db.IsFieldAccessible → ForbiddenError                                                                 │
 │                        │ (*Fields, error)                                │                                                                                                             │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │                        │                                                 │ Validate: label required, type required, field type exists (GetFieldTypeByType), validation config JSON     │
 │ CreateField            │ (ctx, ac, CreateFieldParams) (*Fields, error)   │ (types.ParseValidationConfig + types.ValidateValidationConfig). Defaults: empty JSON for                    │
 │                        │                                                 │ data/validation/ui_config, generate ID, set timestamps                                                      │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ UpdateField            │ (ctx, ac, UpdateFieldParams) (*Fields, error)   │ Same validation. Fetch existing to preserve immutable fields. Update + re-fetch                             │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ DeleteField            │ (ctx, ac, FieldID) error                        │ Validate ID                                                                                                 │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ListFieldsPaginated    │ (ctx, PaginationParams, roleID string, isAdmin  │ Paginate + count + role filter                                                                              │
 │                        │ bool) (PaginatedResponse[Fields], error)        │                                                                                                             │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ UpdateFieldSortOrder   │ (ctx, ac, UpdateFieldSortOrderParams) error     │ Validate field ID                                                                                           │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ GetMaxSortOrder        │ (ctx, NullableDatatypeID) (int64, error)        │ Pass-through                                                                                                │
 ├────────────────────────┼─────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ ListFieldsByDatatypeID │ (ctx, NullableDatatypeID) ([]Fields, error)     │ Nil-safe                                                                                                    │
 └────────────────────────┴─────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 `NullableDatatypeID` is `types.NullableDatatypeID` from `internal/db/types/`.

 Note on roles/translatable: These are fields in db.CreateFieldParams / db.UpdateFieldParams. Since we pass through db.*Params types, roles and translatable are set by the handler before
 calling the service. The service doesn't need special handling — it validates and persists whatever the handler provides.

 Public FieldType Methods

 ┌─────────────────┬───────────────────────────────────────────────────────┬───────────────────────────────────────────────────┐
 │     Method      │                       Signature                       │                  Business Logic                   │
 ├─────────────────┼───────────────────────────────────────────────────────┼───────────────────────────────────────────────────┤
 │ ListFieldTypes  │ (ctx) ([]FieldTypes, error)                           │ Nil-safe                                          │
 ├─────────────────┼───────────────────────────────────────────────────────┼───────────────────────────────────────────────────┤
 │ GetFieldType    │ (ctx, FieldTypeID) (*FieldTypes, error)               │ NotFoundError if nil                              │
 ├─────────────────┼───────────────────────────────────────────────────────┼───────────────────────────────────────────────────┤
 │ CreateFieldType │ (ctx, ac, CreateFieldTypeParams) (*FieldTypes, error) │ Pass-through (FieldTypes have minimal validation) │
 ├─────────────────┼───────────────────────────────────────────────────────┼───────────────────────────────────────────────────┤
 │ UpdateFieldType │ (ctx, ac, UpdateFieldTypeParams) (*FieldTypes, error) │ Update + re-fetch                                 │
 ├─────────────────┼───────────────────────────────────────────────────────┼───────────────────────────────────────────────────┤
 │ DeleteFieldType │ (ctx, ac, FieldTypeID) error                          │ Validate ID                                       │
 └─────────────────┴───────────────────────────────────────────────────────┴───────────────────────────────────────────────────┘

 Admin variants

 Admin variants follow the same validate-defaults-store-wrap pattern as public methods. Signatures use `AdminDatatypeID`, `AdminFieldID`, `AdminFieldTypeID` and return `AdminDatatypes`, `AdminFields`, `AdminFieldTypes` respectively. Named `CreateAdminDatatype`, `ListAdminFields`, `UpdateAdminFieldType`, etc.

 Key difference: Admin datatype methods do NOT call `ValidateUserDatatypeType`. Admin datatypes use system-defined types (Page, Component, etc.) that may not pass user-type validation. All other validation (label required, ID format, field type existence for fields) applies identically.

 Behavioral change: unified validation (public datatypes only)

 Currently, `ValidateUserDatatypeType` is only called by the REST API handler, not the admin panel handler. After migration, the service's public `CreateDatatype`/`UpdateDatatype` calls it for all consumers. This is an intentional improvement — the admin panel should enforce the same type validation for user-facing datatypes. Admin variant methods (`CreateAdminDatatype`, etc.) are exempt.

 ---
 Step 3: Rewire admin panel handlers

 Files to modify:
 - internal/admin/handlers/datatypes.go — all 7 handlers + `DatatypesJSONHandler` and `DatatypeFieldsJSONHandler`
 - internal/admin/handlers/fields.go — FieldDetailHandler, FieldUpdateHandler, FieldDeleteHandler only
 - internal/admin/handlers/field_types.go — no change (hardcoded list, no DB calls)

 Dead code removal: Remove FieldsListHandler and FieldCreateHandler from fields.go — they are not registered in mux.go and are dead code.

 Signature change: All schema handlers change to `(svc *service.Registry)` regardless of their current parameter list. Handlers that previously took only `driver` gain access to config through `svc.Config()` if needed.

 What moves to service:
 - Validation (label required, type required, field type existence check, validation config parsing)
 - Default value assignment (empty JSON, ID generation, timestamps)
 - Audit context construction → replaced by svc.AuditCtx(ctx)
 - Pre-fetch for updates (service handles immutable field preservation internally)

 What stays in handler:
 - HTTP form parsing (r.ParseForm(), r.FormValue())
 - HTMX-aware rendering (partial vs full page)
 - Fuzzy search, sorting, pagination slicing (presentation logic)
 - CSRF token from context
 - Role extraction from context (passed as params to service)
 - HTTP response formatting (status codes, headers, JSON encoding)
 - Config reads for presentation (cfg.I18nEnabled() in FieldDetailHandler)
 - Roles multi-select parsing (JSON marshal of form values → types.NullableString)

 ValidationError → map[string]string mapping in admin handlers:
 var ve *service.ValidationError
 if errors.As(err, &ve) {
     errs := make(map[string]string, len(ve.Errors))
     for _, fe := range ve.Errors {
         errs[fe.Field] = fe.Message
     }
     // render form partial with errs
 }

 IP resolution behavior change: Admin handlers currently parse r.RemoteAddr with net.SplitHostPort to get client IP. After migration, svc.AuditCtx(ctx) uses
 middleware.ClientIPFromContext(ctx), which resolves X-Forwarded-For → X-Real-IP → RemoteAddr. This is an intentional improvement — the admin handlers were bypassing the proxy-aware
 middleware.

 ---
 Step 4: Rewire REST API handlers

 Files to modify:
 - internal/router/datatypes.go — includes `DatatypeFullHandler`: replace `db.ConfigDB(c)` + `db.AssembleDatatypeFullView` with `svc.Schema.GetDatatypeFull`
 - internal/router/adminDatatypes.go
 - internal/router/fields.go
 - internal/router/adminFields.go
 - internal/router/field_types.go
 - internal/router/admin_field_types.go
 - internal/router/fieldSortOrder.go — both FieldSortOrderHandler and FieldMaxSortOrderHandler

 Migration strategy (Option A — incremental): Keep existing func(w, r, config.Config) signatures, add svc *service.Registry as additional parameter. Mux closures capture svc:
 mux.Handle("/api/v1/datatype", perm(http.HandlerFunc(func(w, r) {
     DatatypesHandler(w, r, *c, svc)
 })))

 Special case: FieldSortOrderHandler takes (w, r, db.DbDriver, config.Config) and FieldMaxSortOrderHandler takes (w, r, db.DbDriver). Both change to (w, r, svc *service.Registry) — the
 driver and config are accessed through svc.

 What moves to service:
 - db.ConfigDB(c) lookup (eliminated)
 - Validation (label == "", type == "", ValidateUserDatatypeType)
 - Default values (empty JSON, ID generation, timestamps)
 - middleware.AuditContextFromRequest(r, c) → svc.AuditCtx(r.Context())
 - Update-then-refetch pattern (service returns updated entity)

 What stays in handler:
 - JSON decode of request body
 - HTTP status code mapping
 - JSON encode of response
 - Method routing (GET/POST/PUT/DELETE switch)
 - API internal functions (e.g., `apiCreateDatatype`) retain their `error` return types even though callers ignore the return value. Do not change these signatures in this phase.

 Cleanup while touching these files:
 - Remove fmt.Println(err) on line 25 of adminDatatypes.go (debug output in production code)
 - API handler internal functions return error but callers ignore it. Leave return types as-is for now — removing them would be a larger refactor beyond Phase 1 scope.

 ---
 Step 5: Write service-level tests

 File: internal/service/schema_test.go

 Table-driven tests against SQLite. Test setup:
 1. Create a temporary SQLite database using `t.TempDir()` to ensure automatic cleanup regardless of how tests are invoked
 2. Call CreateAllTables()
 3. Construct SchemaService with the test driver

 Test categories:
 - Validation: label required, type required, invalid field type → ValidationError
 - CRUD round-trip: create → get → update → get → delete → get (NotFoundError)
 - Pagination: create N items → list paginated → verify counts and offsets
 - Field role filtering: create field with roles JSON → ListFieldsFiltered with matching/non-matching role
 - Field access check: GetField with restricted role → ForbiddenError
 - Validation config: valid and invalid JSON → appropriate error
 - Sort order: update sort order → verify, get max sort order
 - Full view: create datatype + create fields with parent → GetDatatypeFull → verify fields present
 - Error wrapping: verify DB errors are wrapped as InternalError, not-founds as NotFoundError

 ---
 Step 6: Update mux.go route registrations

 File: internal/router/mux.go

 Admin handler registrations in registerAdminRoutes change from:
 adminhandlers.DatatypeCreateHandler(driver, mgr)
 adminhandlers.DatatypeDetailHandler(driver)
 adminhandlers.FieldDetailHandler(driver, mgr)
 to:
 adminhandlers.DatatypeCreateHandler(svc)
 adminhandlers.DatatypeDetailHandler(svc)
 adminhandlers.FieldDetailHandler(svc)

 API route closures in NewModulacmsMux pass svc:
 DatatypesHandler(w, r, *c, svc)
 FieldSortOrderHandler(w, r, svc)
 FieldMaxSortOrderHandler(w, r, svc)

 Also update DatatypesJSONHandler(driver) → DatatypesJSONHandler(svc) and DatatypeFieldsJSONHandler(driver) → DatatypeFieldsJSONHandler(svc).

 ---
 Files Changed Summary

 ┌────────────────────────────────────────┬─────────────┬────────────────────────────────────────────────────────────────────────────────┐
 │                  File                  │ Change Type │                                  Description                                   │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/service/schema.go             │ Rewrite     │ Composable store interfaces + SchemaService + all methods                      │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/service/service.go            │ Edit        │ Add Schema *SchemaService field + wire in NewRegistry                          │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/service/schema_test.go        │ New         │ Service-level tests against SQLite                                             │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/handlers/datatypes.go   │ Edit        │ Change all 7 handlers + DatatypesJSONHandler + DatatypeFieldsJSONHandler to (svc) │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/handlers/fields.go      │ Edit        │ Remove dead FieldsListHandler/FieldCreateHandler, migrate remaining 3 handlers │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/admin/handlers/field_types.go │ No change   │ Hardcoded list, no DB calls                                                    │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/datatypes.go           │ Edit        │ Add svc param, replace db.ConfigDB with service calls                          │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/adminDatatypes.go      │ Edit        │ Same + remove fmt.Println(err)                                                 │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/fields.go              │ Edit        │ Same                                                                           │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/adminFields.go         │ Edit        │ Same                                                                           │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/field_types.go         │ Edit        │ Same                                                                           │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/admin_field_types.go   │ Edit        │ Same                                                                           │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/fieldSortOrder.go      │ Edit        │ Both handlers: change to (w, r, svc)                                           │
 ├────────────────────────────────────────┼─────────────┼────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/mux.go                 │ Edit        │ Update all schema handler registrations                                        │
 └────────────────────────────────────────┴─────────────┴────────────────────────────────────────────────────────────────────────────────┘

 ---
 Execution Order

 Step 1-2: SchemaService implementation (schema.go + service.go)
 Step 5:   Tests (schema_test.go) — verify service standalone
     ├── Step 3: Admin handler rewiring + dead code removal  (parallel)
     ├── Step 4: API handler rewiring                        (parallel)
     └── Step 6: Mux registration updates                    (after 3+4)

 Sequential single-agent order:
 1. Implement SchemaService (schema.go) — interfaces, struct, all methods
 2. Wire into Registry (service.go)
 3. Write tests (schema_test.go) — verify service works standalone
 4. Rewire admin handlers + remove dead code + mux registrations
 5. Rewire API handlers + mux registrations
 6. go build ./... and just test

 ---
 Verification

 1. go build ./... — compiles clean
 2. just test — all existing tests pass (no regressions)
 3. go test -v ./internal/service/ — new service tests pass
 4. Manual smoke: just run → admin panel schema pages work (list/create/edit/delete datatypes and fields)
