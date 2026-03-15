Composite/Cascade API Endpoints Plan

 Context

 ModulaCMS currently exposes low-level CRUD endpoints that operate on single rows. Several common workflows require multiple sequential API calls that should be atomic or at least
 coordinated:

 - Creating content requires a separate call per field
 - Deleting a user fails if they authored any content (FK RESTRICT), with no way to reassign
 - Deleting a datatype fails if content exists, with no cascade option
 - Deleting content doesn't touch children (they become orphaned)
 - Deleting media leaves dangling references in content fields
 - Adding/removing a field from a datatype doesn't propagate to existing content

 This plan adds 6 composite endpoints that handle these workflows, plus a protected system user for ownership reassignment. All three SDKs (TypeScript, Go, Swift) are updated to match.

 ---
 Phase 1: Foundation (SQL + DbDriver + System User)

 1a. New SQL Queries

 Add to all three dialects (SQLite queries.sql, MySQL queries_mysql.sql, PostgreSQL queries_psql.sql):

 sql/schema/16_content_data/
 -- name: ListContentDataByDatatypeID :many
 SELECT * FROM content_data WHERE datatype_id = ?;

 -- name: ReassignContentDataAuthor :exec
 UPDATE content_data SET author_id = ? WHERE author_id = ?;

 -- name: CountContentDataByAuthor :one
 SELECT COUNT(*) FROM content_data WHERE author_id = ?;

 sql/schema/7_datatypes/
 -- name: ReassignDatatypeAuthor :exec
 UPDATE datatypes SET author_id = ? WHERE author_id = ?;

 -- name: CountDatatypesByAuthor :one
 SELECT COUNT(*) FROM datatypes WHERE author_id = ?;

 sql/schema/18_admin_content_data/
 -- name: ReassignAdminContentDataAuthor :exec
 UPDATE admin_content_data SET author_id = ? WHERE author_id = ?;

 -- name: CountAdminContentDataByAuthor :one
 SELECT COUNT(*) FROM admin_content_data WHERE author_id = ?;

 Run just sqlc to regenerate. Then add all new methods to DbDriver interface in internal/db/db.go and implement on all three wrapper structs (Database, MysqlDatabase, PsqlDatabase).

 1b. Protected System User

 Add to bootstrap data creation in internal/install/ (or wherever CreateBootstrapData lives):

 - Create a user with username system, email system@localhost, role admin
 - Mark as protected: add a system_protected boolean column to users table, or use a convention (e.g., user_id is a well-known constant ULID like 00000000000000000000000000). Prefer
 convention -- use a sentinel ULID constant types.SystemUserID set to 00000000000000000000SYSTEM (or a valid ULID generated once and hardcoded). No schema change needed.
 - Add types.SystemUserID constant in internal/db/types/
 - Guard in ApiDeleteUser: reject deletion if user_id == types.SystemUserID
 - Guard in ApiUpdateUser: reject role changes or disabling of system user
 - The system user is created during bootstrap with a random 64-char password (unguessable, no login)

 Files:
 - internal/db/types/ids.go -- add SystemUserID constant
 - internal/install/bootstrap.go (or equivalent) -- create system user during setup
 - internal/router/users.go -- add guards to delete/update handlers

 1c. Extract Sibling Helpers

 Extract from internal/router/contentData.go into new internal/router/contentHelpers.go:

 // appendToSiblingChain links a newly created content node to its parent's sibling chain.
 func appendToSiblingChain(ctx context.Context, ac audited.AuditContext, d db.DbDriver, node *db.ContentData) (*db.ContentData, error)

 // deleteContentWithSiblingRepair removes a content node and repairs the sibling linked list.
 func deleteContentWithSiblingRepair(ctx context.Context, ac audited.AuditContext, d db.DbDriver, cdID types.ContentID) error

 // collectDescendants returns all descendant content IDs in delete-safe order (leaves first).
 // maxDepth caps recursion (default 100). maxCount caps total nodes (default 1000).
 func collectDescendants(d db.DbDriver, nodeID types.ContentID, maxDepth, maxCount int) ([]types.ContentID, error)

 Update contentData.go to call these helpers instead of inlining the logic.

 ---
 Phase 2: Endpoint 1 -- Content Create with Fields

 Route: POST /api/v1/content/create
 Permission: content:create
 File: New internal/router/contentComposite.go

 Request

 {
   "parent_id": "<content_id or null>",
   "route_id": "<route_id or null>",
   "datatype_id": "<datatype_id>",         // required
   "status": "draft",
   "fields": {                              // optional initial values
     "<field_id>": "value",
     "<field_id>": "value"
   }
 }

 author_id is derived from the authenticated user (same as batch endpoint pattern).

 Response

 {
   "content_data": { /* full ContentData */ },
   "fields": [ /* all created ContentFields */ ],
   "fields_created": 5,
   "fields_failed": 0,
   "errors": []
 }

 Logic

 1. Validate datatype_id exists via d.GetDatatype()
 2. Build CreateContentDataParams from request + authenticated user
 3. Call d.CreateContentData(ctx, ac, params)
 4. Call appendToSiblingChain() if parent_id is set
 5. Fetch datatype's fields via d.ListDatatypeFieldByDatatypeID(datatypeID)
 6. For each field: create content_field with value from fields map (or empty string "")
 7. Return content_data + all created fields + counts

 Mux Registration

 mux.Handle("POST /api/v1/content/create",
     middleware.RequirePermission("content:create")(handler))

 ---
 Phase 3: Endpoint 2 -- User Reassign-Delete

 Route: POST /api/v1/users/reassign-delete
 Permission: users:delete
 File: New internal/router/userComposite.go

 Request

 {
   "user_id": "<ulid of user to delete>",
   "reassign_to": "<ulid or omit for system user>"
 }

 When reassign_to is omitted or empty, defaults to types.SystemUserID.

 Response

 {
   "deleted_user_id": "<ulid>",
   "reassigned_to": "<ulid>",
   "content_data_reassigned": 12,
   "datatypes_reassigned": 3,
   "admin_content_data_reassigned": 0
 }

 Logic

 1. Validate user_id exists. Reject if user_id == types.SystemUserID.
 2. Resolve reassign_to (default to system user). Verify target exists.
 3. Reject if user_id == reassign_to.
 4. Call bulk reassignment queries (no per-row audit):
   - d.ReassignContentDataAuthor(ctx, oldID, newID)
   - d.ReassignDatatypeAuthor(ctx, oldID, newID)
   - d.ReassignAdminContentDataAuthor(ctx, oldID, newID)
 5. Get counts via the Count queries (called before reassignment) for the response.
 6. Call d.DeleteUser(ctx, ac, userID) -- FK RESTRICT constraints now satisfied.
 7. Return counts.

 If any reassignment fails, abort before deleting. Return the error.

 ---
 Phase 4: Endpoint 3 -- Datatype Cascade Delete

 Route: DELETE /api/v1/datatypes/?q=<id>&cascade=true
 Permission: datatypes:delete
 File: Modify internal/router/datatypes.go

 Response (cascade=true)

 {
   "deleted_datatype_id": "<ulid>",
   "content_deleted": 15,
   "errors": []
 }

 Logic

 1. Parse cascade query param. If false/absent, existing behavior.
 2. If cascade=true:
 a. d.ListContentDataByDatatypeID(dtID) to get all content using this datatype
 b. For each content node, call deleteContentWithSiblingRepair(ctx, ac, d, cdID) (content_fields cascade via FK)
 c. Delete the datatype via d.DeleteDatatype(ctx, ac, dtID) (datatypes_fields cascade via FK)
 3. Return counts. Stop on first error.

 ---
 Phase 5: Endpoint 4 -- Recursive Content Delete

 Route: DELETE /api/v1/contentdata/?q=<id>&recursive=true
 Permission: content:delete
 File: Modify internal/router/contentData.go

 Response (recursive=true)

 {
   "deleted_root": "<ulid>",
   "total_deleted": 8,
   "deleted_ids": ["<ulid>", ...]
 }

 Logic

 1. Parse recursive query param. If false/absent, existing behavior.
 2. If recursive=true:
 a. collectDescendants(d, nodeID, 100, 1000) -- depth-first, leaves first
 b. For descendants whose parent is also in the delete set: skip sibling repair (entire level is being removed)
 c. For descendants whose parent is NOT in the delete set: use deleteContentWithSiblingRepair
 d. Delete the root node itself with sibling repair
 3. Safety limits: 100 depth, 1000 nodes. Return 400 if exceeded.

 ---
 Phase 6: Endpoint 5 -- Media Reference Scan + Cleanup

 New route: GET /api/v1/media/references?q=<media_id>
 Enhanced: DELETE /api/v1/media/?q=<id>&clean_refs=true
 Permission: media:read / media:delete
 File: Modify internal/router/media.go

 Reference Scan Response

 {
   "media_id": "<ulid>",
   "references": [
     {
       "content_field_id": "<ulid>",
       "content_data_id": "<ulid>",
       "field_id": "<ulid>"
     }
   ],
   "reference_count": 3
 }

 Logic

 1. Get media record to obtain the media URL/ID string.
 2. d.ListContentFields() and filter where field_value contains the media ID or URL. This is a string containment check, not regex.
 3. For scan endpoint: return the reference list.
 4. For delete with clean_refs=true: update each matched content_field's field_value to "" via d.UpdateContentField(), then proceed with existing S3 + DB delete.
 5. For delete without clean_refs: existing behavior (just delete, dangling refs remain).

 ---
 Phase 7: Endpoint 6 -- Datatype Field Cascade

 Enhanced: POST /api/v1/datatypefields?cascade=true and DELETE /api/v1/datatypefields/?q=<id>&cascade=true
 Permission: fields:create / fields:delete
 File: Modify internal/router/datatype_fields.go

 Add Field with Cascade

 After creating the datatype_field junction row:
 1. d.ListContentDataByDatatypeID(dtID) -- all content using this datatype
 2. For each content_data: d.CreateContentField(ctx, ac, {contentDataID, fieldID, value: ""}) -- create empty field
 3. Return junction + counts.

 Remove Field with Cascade

 Before deleting the datatype_field junction:
 1. Get the junction to extract datatype_id and field_id
 2. d.ListContentDataByDatatypeID(dtID) -- all content using this datatype
 3. For each content_data: find content_fields matching field_id, delete them via d.DeleteContentField()
 4. Delete the junction row.
 5. Return counts.

 ---
 Phase 8: SDK Updates

 TypeScript Admin SDK (sdks/typescript/modulacms-admin-sdk/)

 Add to src/types/content.ts:
 export interface ContentCreateParams {
   parent_id?: string | null;
   route_id?: string | null;
   datatype_id: string;
   status?: string;
   fields?: Record<string, string>;
 }

 export interface ContentCreateResponse {
   content_data: ContentData;
   fields: ContentField[];
   fields_created: number;
   fields_failed: number;
   errors: string[];
 }

 Add to src/types/user.ts (or equivalent):
 export interface UserReassignDeleteParams {
   user_id: string;
   reassign_to?: string;
 }

 export interface UserReassignDeleteResponse {
   deleted_user_id: string;
   reassigned_to: string;
   content_data_reassigned: number;
   datatypes_reassigned: number;
   admin_content_data_reassigned: number;
 }

 Add methods to the appropriate resource classes:
 - contentData.createWithFields(params) -- POST /content/create
 - users.reassignDelete(params) -- POST /users/reassign-delete
 - contentData.deleteRecursive(id) -- DELETE /contentdata/?q=id&recursive=true
 - datatypes.deleteCascade(id) -- DELETE /datatypes/?q=id&cascade=true
 - media.getReferences(id) -- GET /media/references?q=id
 - media.deleteWithCleanup(id) -- DELETE /media/?q=id&clean_refs=true
 - datatypeFields.createCascade(params) -- POST /datatypefields?cascade=true
 - datatypeFields.deleteCascade(id) -- DELETE /datatypefields/?q=id&cascade=true

 Also add types to sdks/typescript/types/ for the shared request/response shapes.

 Go SDK (sdks/go/)

 Add equivalent methods and types:
 - content_composite.go -- CreateContentWithFields(ctx, params), DeleteContentRecursive(ctx, id)
 - user_composite.go -- ReassignDeleteUser(ctx, params)
 - Extend existing resource files for cascade params on datatypes, media, datatypefields

 Swift SDK (sdks/swift/Sources/ModulaCMS/)

 Add equivalent methods and types following the existing Swift SDK patterns:
 - ContentCompositeResource.swift
 - UserCompositeResource.swift
 - Extend existing resources for cascade/recursive options

 ---
 Implementation Order

 ┌──────┬───────────────────────────────────────────┬────────────┐
 │ Step │                   What                    │ Depends On │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 1a   │ SQL queries + sqlc + DbDriver             │ --         │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 1b   │ System user constant + bootstrap + guards │ 1a         │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 1c   │ Extract sibling helpers                   │ --         │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 2    │ Content create with fields                │ 1a, 1c     │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 3    │ User reassign-delete                      │ 1a, 1b     │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 4    │ Datatype cascade delete                   │ 1a, 1c     │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 5    │ Recursive content delete                  │ 1c         │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 6    │ Media reference scan + cleanup            │ 1a         │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 7    │ Datatype field cascade                    │ 1a         │
 ├──────┼───────────────────────────────────────────┼────────────┤
 │ 8    │ SDK updates (TS, Go, Swift)               │ 2-7        │
 └──────┴───────────────────────────────────────────┴────────────┘

 Steps 1a and 1c are parallelizable. Steps 2-7 are parallelizable after their dependencies. Step 8 follows all backend work.

 ---
 Files to Create

 ┌─────────────────────────────────────┬──────────────────────────────────────────────────┐
 │                File                 │                     Purpose                      │
 ├─────────────────────────────────────┼──────────────────────────────────────────────────┤
 │ internal/router/contentComposite.go │ Content create with fields handler               │
 ├─────────────────────────────────────┼──────────────────────────────────────────────────┤
 │ internal/router/userComposite.go    │ User reassign-delete handler                     │
 ├─────────────────────────────────────┼──────────────────────────────────────────────────┤
 │ internal/router/contentHelpers.go   │ Extracted sibling chain + tree traversal helpers │
 └─────────────────────────────────────┴──────────────────────────────────────────────────┘

 Files to Modify

 ┌─────────────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────┐
 │                            File                             │                                       Changes                                        │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sql/schema/16_content_data/queries*.sql (x3 dialects)       │ Add ListContentDataByDatatypeID, ReassignContentDataAuthor, CountContentDataByAuthor │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sql/schema/7_datatypes/queries*.sql (x3 dialects)           │ Add ReassignDatatypeAuthor, CountDatatypesByAuthor                                   │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sql/schema/18_admin_content_data/queries*.sql (x3 dialects) │ Add ReassignAdminContentDataAuthor, CountAdminContentDataByAuthor                    │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/db.go                                           │ Add new methods to DbDriver interface                                                │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/content_data_gen.go                             │ Implement new methods on Database                                                    │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/content_data_gen_mysql.go                       │ Implement on MysqlDatabase                                                           │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/content_data_gen_psql.go                        │ Implement on PsqlDatabase                                                            │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/datatype_gen.go + mysql/psql variants           │ Implement reassign methods                                                           │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/admin_content_data_gen.go + variants            │ Implement reassign methods                                                           │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/db/types/ids.go                                    │ Add SystemUserID constant                                                            │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/mux.go                                      │ Register 4 new routes                                                                │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/contentData.go                              │ Use extracted helpers; add recursive delete                                          │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/datatypes.go                                │ Add cascade delete option                                                            │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/datatype_fields.go                          │ Add cascade create/delete options                                                    │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/media.go                                    │ Add reference scan; add clean_refs delete                                            │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/router/users.go                                    │ Add system user guards                                                               │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ internal/install/ (bootstrap file)                          │ Create system user during setup                                                      │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sdks/typescript/types/src/*.ts                              │ New shared types                                                                     │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sdks/typescript/modulacms-admin-sdk/src/types/*.ts          │ New param/response types                                                             │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sdks/typescript/modulacms-admin-sdk/src/index.ts            │ New methods on resources                                                             │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sdks/go/*.go                                                │ New composite methods + types                                                        │
 ├─────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────┤
 │ sdks/swift/Sources/ModulaCMS/*.swift                        │ New composite resources + types                                                      │
 └─────────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────┘

 Verification

 1. just sqlc -- regenerate after SQL changes
 2. just check -- compile check
 3. just test -- run existing tests (should not break)
 4. Manual API testing with curl against each new endpoint
 5. just sdk-build && just sdk-typecheck -- verify SDK builds
 6. just sdk-go-test && just sdk-go-vet -- verify Go SDK
 7. just sdk-swift-build -- verify Swift SDK
