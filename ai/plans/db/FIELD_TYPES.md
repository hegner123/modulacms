 Plan: Add field_types and admin_field_types Lookup Tables

 Context

 Field types in ModulaCMS are hardcoded as a Go enum (FieldType with 14 values) and enforced via SQL CHECK constraints. Users need to dynamically add field type values that appear in the fields type column at runtime. This plan adds two
 lookup tables (field_types and admin_field_types) that make field types queryable and extensible.

 This is step 1 of a two-step effort. Step 2 (future, separate plan) will:
 - Wire fields.type to FK on field_types.field_type_id (ULID)
 - Wire admin_fields.type to FK on admin_field_types.admin_field_type_id (ULID)
 - Remove CHECK constraints from the 6 fields/admin_fields schema files
 - Update FieldType.Validate() to load valid types from the DB instead of the hardcoded switch

 Two tables are intentional -- they serve two separate consumers (fields and admin_fields), consistent with the existing admin/public table pattern throughout the codebase.

 No migration plan is needed -- there are no existing deployed databases.

 This plan covers the full stack: database, HTTP API, TUI, admin panel, and all SDKs. HTTP routes and admin UI are included here so that field types are queryable and manageable at runtime immediately, not deferred to step 2.

 Table Structure

 Both tables: 3 columns, no timestamps, no FKs.

 ┌─────────────────────────────────────┬───────────────────────┬──────────────────┐
 │               Column                │         Type          │    Constraint    │
 ├─────────────────────────────────────┼───────────────────────┼──────────────────┤
 │ field_type_id / admin_field_type_id │ TEXT (ULID, 26 chars) │ PRIMARY KEY      │
 ├─────────────────────────────────────┼───────────────────────┼──────────────────┤
 │ type                                │ TEXT                  │ NOT NULL, UNIQUE │
 ├─────────────────────────────────────┼───────────────────────┼──────────────────┤
 │ label                               │ TEXT                  │ NOT NULL         │
 └─────────────────────────────────────┴───────────────────────┴──────────────────┘

 ═══════════════════════════════════════════════════════════════
 PHASE 1: Database Layer
 ═══════════════════════════════════════════════════════════════

 1. Add ID types

 File: internal/db/types/types_ids.go

 Append FieldTypeID and AdminFieldTypeID blocks following the RolePermissionID pattern (~70 lines each): type declaration, New*(), String(), IsZero(), Validate(), ULID(), Time(), Parse*(), Value(), Scan(), MarshalJSON(), UnmarshalJSON().

 2. Create SQL schema directories (12 new files)

 sql/schema/27_field_types/ -- 6 files (schema + queries x 3 dialects)
 sql/schema/28_admin_field_types/ -- 6 files (same shape, admin_ prefixed columns/queries)

 Queries per table: Create*Table, Drop*Table, Get* (by ID), Get*ByType (by type string), List*, Create*, Update*, Delete*, Count*.

 Dialect differences:
 - SQLite: ? placeholders, RETURNING *, CHECK(length(...)=26)
 - MySQL: ? placeholders, :exec for create (no RETURNING), VARCHAR(26)/VARCHAR(255), named CONSTRAINT
 - PostgreSQL: $1,$2,$3 placeholders, RETURNING *, CHECK(length(...)=26)

 3. Update sqlc config

 File: sql/sqlc.yml

 In all 3 engine sections, add rename entries and override entries for the new typed IDs.

 4. Run just sqlc

 Regenerates internal/db-sqlite/, internal/db-mysql/, internal/db-psql/.

 5. Add entity definitions

 File: tools/dbgen/definitions.go

 Add FieldTypes and AdminFieldTypes entities in the "FULL GENERATION ENTITIES" section. 3 fields each:
 - ID: types.FieldTypeID / types.AdminFieldTypeID (toString conversion)
 - Type: string
 - Label: string
 No nullable fields, no per-driver type differences. Include GetFieldTypeByType / GetAdminFieldTypeByType as ExtraQueries.

 6. Run dbgen

 go run ./tools/dbgen -- generates internal/db/field_type_gen.go and internal/db/admin_field_type_gen.go.

 7. Update DbDriver interface

 File: internal/db/db.go

 Add methods for both entities: Count, Create, CreateTable, Delete, Get, GetByType, List, Update -- for both FieldTypes and AdminFieldTypes.

 8. Update CreateAllTables -- all three drivers

 File: internal/db/db.go

 Add CreateFieldTypesTable() and CreateAdminFieldTypesTable() calls to ALL THREE driver implementations:
 - (d Database).CreateAllTables() -- SQLite
 - (d MysqlDatabase).CreateAllTables() -- MySQL
 - (d PsqlDatabase).CreateAllTables() -- PostgreSQL

 Place in Tier 0: Foundation tables (no dependencies), alongside CreatePermissionTable(), CreateRoleTable(), CreateMediaDimensionTable().

 9. Update DropAllTables -- all three drivers

 File: internal/db/wipe.go

 Add DropFieldTypesTable and DropAdminFieldTypesTable to ALL THREE driver implementations:
 - (d Database).DropAllTables()
 - (d MysqlDatabase).DropAllTables()
 - (d PsqlDatabase).DropAllTables()

 Place in Tier 0: Foundation tables section.

 10. Add bootstrap seed data -- all three drivers

 File: internal/db/db.go

 In ALL THREE driver-specific CreateBootstrapData() methods, seed all 14 type records into field_types AND all 14 identical type records into admin_field_types. Both tables get the same 14 rows:

 ┌──────────┬─────────────┐
 │   type   │    label    │
 ├──────────┼─────────────┤
 │ text     │ Text Input  │
 ├──────────┼─────────────┤
 │ textarea │ Text Area   │
 ├──────────┼─────────────┤
 │ number   │ Number      │
 ├──────────┼─────────────┤
 │ date     │ Date        │
 ├──────────┼─────────────┤
 │ datetime │ Date & Time │
 ├──────────┼─────────────┤
 │ boolean  │ Boolean     │
 ├──────────┼─────────────┤
 │ select   │ Select      │
 ├──────────┼─────────────┤
 │ media    │ Media       │
 ├──────────┼─────────────┤
 │ relation │ Relation    │
 ├──────────┼─────────────┤
 │ json     │ JSON        │
 ├──────────┼─────────────┤
 │ richtext │ Rich Text   │
 ├──────────┼─────────────┤
 │ slug     │ Slug        │
 ├──────────┼─────────────┤
 │ email    │ Email       │
 ├──────────┼─────────────┤
 │ url      │ URL         │
 └──────────┴─────────────┘

 11. Update ValidateBootstrapData -- all three drivers

 File: internal/db/db.go

 Add validation to all three driver-specific ValidateBootstrapData() methods: each table (field_types and admin_field_types) has >= 1 record. Uses >= 1 for consistency with existing validation checks (permissions, roles, etc.).

 12. Update combined schemas

 Files: sql/all_schema*.sql

 Append new table DDL to all_schema.sql, all_schema_mysql.sql, all_schema_psql.sql.

 13. Add bootstrap permission labels

 File: internal/db/db.go

 Add to the rbacPermissionLabels slice in CreateBootstrapData():
 - "field_types:read", "field_types:create", "field_types:update", "field_types:delete", "field_types:admin"
 - "admin_field_types:read", "admin_field_types:create", "admin_field_types:update", "admin_field_types:delete", "admin_field_types:admin"

 Add the 10 new labels to admin (all) and editor (read/create/update/delete) role assignments. Viewer gets read only.

 14. Add database tests

 Following the permission_crud_test.go / permission_test.go pattern:

 internal/db/field_type_crud_test.go -- Integration CRUD lifecycle tests:
 - Full CRUD cycle (Count/Create/Get/List/Update/Delete)
 - Multiple records test
 - GetByType query test
 - Uses testIntegrationDB(t) (Tier 0: no FK dependencies)

 internal/db/field_type_test.go -- White-box mapper tests:
 - MapStringFieldType tests (all fields, zero value)
 - Per-driver mapper tests: MapFieldType, MapCreateFieldTypeParams, MapUpdateFieldTypeParams for SQLite, MySQL, PostgreSQL
 - Cross-database mapper consistency
 - Audited command accessor tests (Create/Update/Delete) for all 3 drivers
 - Cross-database audited command consistency
 - Compile-time interface assertions

 internal/db/admin_field_type_crud_test.go -- Same structure as field_type_crud_test.go
 internal/db/admin_field_type_test.go -- Same structure as field_type_test.go

 ═══════════════════════════════════════════════════════════════
 PHASE 2: HTTP API
 ═══════════════════════════════════════════════════════════════

 15. Add HTTP handler -- field types

 File: internal/router/field_types.go (new)

 Follow the routes.go handler pattern (3 params: w, r, c). Same dispatch structure as permissions.go but WITHOUT the *middleware.PermissionCache parameter (field_types don't affect RBAC cache) and WITHOUT SystemProtected guards (field_types have no system_protected column).

 Two dispatcher functions, five operation functions:

 FieldTypesHandler(w http.ResponseWriter, r *http.Request, c config.Config) -- collection endpoint
   GET  -> apiListFieldTypes(w, c)
   POST -> apiCreateFieldType(w, r, c)

 FieldTypeHandler(w http.ResponseWriter, r *http.Request, c config.Config) -- item endpoint
   GET    -> apiGetFieldType(w, r, c)
   PUT    -> apiUpdateFieldType(w, r, c)
   DELETE -> apiDeleteFieldType(w, r, c)

 Operation details:
 - apiListFieldTypes: d.ListFieldTypes(), return 200
 - apiGetFieldType: parse ?q=, validate FieldTypeID, d.GetFieldType(), return 200
 - apiCreateFieldType: decode CreateFieldTypeParams, AuditContextFromRequest, d.CreateFieldType(), return 201
 - apiUpdateFieldType: decode UpdateFieldTypeParams, AuditContextFromRequest, d.UpdateFieldType(), fetch updated, return 200
 - apiDeleteFieldType: parse ?q=, validate FieldTypeID, AuditContextFromRequest, d.DeleteFieldType(), return 200

 16. Add HTTP handler -- admin field types

 File: internal/router/admin_field_types.go (new)

 Same structure as step 15 but for AdminFieldTypes:

 AdminFieldTypesHandler(w http.ResponseWriter, r *http.Request, c config.Config) -- collection
 AdminFieldTypeHandler(w http.ResponseWriter, r *http.Request, c config.Config) -- item

 Uses AdminFieldTypeID, d.ListAdminFieldTypes(), d.CreateAdminFieldType(), etc.

 17. Register routes

 File: internal/router/mux.go

 Add 4 route registrations (following the roles/permissions pattern at lines ~217-231):

   // Field types
   mux.Handle("/api/v1/fieldtypes", middleware.RequireResourcePermission("field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       FieldTypesHandler(w, r, *c)
   })))
   mux.Handle("/api/v1/fieldtypes/", middleware.RequireResourcePermission("field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       FieldTypeHandler(w, r, *c)
   })))

   // Admin field types
   mux.Handle("/api/v1/adminfieldtypes", middleware.RequireResourcePermission("admin_field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       AdminFieldTypesHandler(w, r, *c)
   })))
   mux.Handle("/api/v1/adminfieldtypes/", middleware.RequireResourcePermission("admin_field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       AdminFieldTypeHandler(w, r, *c)
   })))

 API endpoints summary:

 ┌────────┬───────────────────────────────────┬───────────────────────────────┐
 │ Method │             URL                   │          Operation            │
 ├────────┼───────────────────────────────────┼───────────────────────────────┤
 │ GET    │ /api/v1/fieldtypes                │ List all field types          │
 │ POST   │ /api/v1/fieldtypes                │ Create field type             │
 │ GET    │ /api/v1/fieldtypes/?q={id}        │ Get field type by ID          │
 │ PUT    │ /api/v1/fieldtypes/               │ Update field type             │
 │ DELETE │ /api/v1/fieldtypes/?q={id}        │ Delete field type             │
 │ GET    │ /api/v1/adminfieldtypes           │ List all admin field types    │
 │ POST   │ /api/v1/adminfieldtypes           │ Create admin field type       │
 │ GET    │ /api/v1/adminfieldtypes/?q={id}   │ Get admin field type by ID    │
 │ PUT    │ /api/v1/adminfieldtypes/          │ Update admin field type       │
 │ DELETE │ /api/v1/adminfieldtypes/?q={id}   │ Delete admin field type       │
 └────────┴───────────────────────────────────┴───────────────────────────────┘

 ═══════════════════════════════════════════════════════════════
 PHASE 3: TUI
 ═══════════════════════════════════════════════════════════════

 Naming convention: The dbgen entity struct is FieldTypes (plural). HTTP handlers use singular for the item endpoint (FieldTypeHandler) following the existing PermissionHandler/RouteHandler pattern. Test file names use singular (field_type_crud_test.go) following the existing permission_crud_test.go pattern. TUI message types use "FieldType" (singular) for individual entity actions and "FieldTypes" (plural) for collection operations.

 The TUI exposes both field_types and admin_field_types as separate screens. Steps 18-25 cover field_types; steps 18b-25b mirror the same structure for admin_field_types.

 18. Add page constants

 File: internal/cli/pages.go

 Add FIELDTYPES and ADMINFIELDTYPES to the PageIndex iota block (after QUICKSTARTPAGE).
 In InitPages():
   fieldTypesPage := NewPage(FIELDTYPES, "Field Types")
   adminFieldTypesPage := NewPage(ADMINFIELDTYPES, "Admin Field Types")
 Add both to the page map:
   p[FIELDTYPES] = fieldTypesPage
   p[ADMINFIELDTYPES] = adminFieldTypesPage

 19. Add model state

 File: internal/cli/model.go

 Add to the Model struct:
   FieldTypesList      []db.FieldTypes
   AdminFieldTypesList []db.AdminFieldTypes

 20. Add message types

 File: internal/cli/admin_message_types.go

 Add message structs for field_types:
   FieldTypesFetchMsg struct{}
   FieldTypesFetchResultsMsg struct{ Data []db.FieldTypes }
   FieldTypesSetMsg struct{ Data []db.FieldTypes }
   CreateFieldTypeFromDialogRequestMsg struct{ Type, Label string }
   FieldTypeCreatedFromDialogMsg struct{ FieldType db.FieldTypes }
   UpdateFieldTypeFromDialogRequestMsg struct{ ID types.FieldTypeID; Type, Label string }
   FieldTypeUpdatedFromDialogMsg struct{ FieldType db.FieldTypes }
   DeleteFieldTypeRequestMsg struct{ ID types.FieldTypeID }
   FieldTypeDeletedMsg struct{ ID types.FieldTypeID }

 Add mirrored message structs for admin_field_types:
   AdminFieldTypesFetchMsg struct{}
   AdminFieldTypesFetchResultsMsg struct{ Data []db.AdminFieldTypes }
   AdminFieldTypesSetMsg struct{ Data []db.AdminFieldTypes }
   CreateAdminFieldTypeFromDialogRequestMsg struct{ Type, Label string }
   AdminFieldTypeCreatedFromDialogMsg struct{ AdminFieldType db.AdminFieldTypes }
   UpdateAdminFieldTypeFromDialogRequestMsg struct{ ID types.AdminFieldTypeID; Type, Label string }
   AdminFieldTypeUpdatedFromDialogMsg struct{ AdminFieldType db.AdminFieldTypes }
   DeleteAdminFieldTypeRequestMsg struct{ ID types.AdminFieldTypeID }
   AdminFieldTypeDeletedMsg struct{ ID types.AdminFieldTypeID }

 21. Add command constructors

 File: internal/cli/admin_constructors.go (or commands.go)

 Add Cmd functions for both entities:
   FieldTypesFetchCmd() tea.Msg -- returns FieldTypesFetchMsg{}
   FieldTypesSetCmd(data []db.FieldTypes) tea.Cmd -- returns FieldTypesSetMsg{Data: data}
   AdminFieldTypesFetchCmd() tea.Msg -- returns AdminFieldTypesFetchMsg{}
   AdminFieldTypesSetCmd(data []db.AdminFieldTypes) tea.Cmd -- returns AdminFieldTypesSetMsg{Data: data}

 22. Add panel view rendering

 File: internal/cli/panel_view.go

 Add FIELDTYPES and ADMINFIELDTYPES to isCMSPanelPage() check.
 Add case FIELDTYPES to cmsPanelTitles(): return "Field Types", "Details", "Actions".
 Add case ADMINFIELDTYPES to cmsPanelTitles(): return "Admin Field Types", "Details", "Actions".
 Add case FIELDTYPES to cmsPanelContent(): dispatch to renderFieldTypesList, renderFieldTypeDetail, renderFieldTypeActions.
 Add case ADMINFIELDTYPES to cmsPanelContent(): dispatch to renderAdminFieldTypesList, renderAdminFieldTypeDetail, renderAdminFieldTypeActions.

 File: internal/cli/admin_panel_view.go (or panel_view.go)

 Add six rendering functions (three per entity):
   renderFieldTypesList(m Model) string -- iterate m.FieldTypesList, show cursor + type + label
   renderFieldTypeDetail(m Model) string -- show selected field type details (ID, type, label)
   renderFieldTypeActions(m Model) string -- show available actions: n=new, e=edit, d=delete
   renderAdminFieldTypesList(m Model) string -- iterate m.AdminFieldTypesList, same layout
   renderAdminFieldTypeDetail(m Model) string -- same layout
   renderAdminFieldTypeActions(m Model) string -- same layout

 23. Add keyboard controls

 File: internal/cli/admin_controls.go (or update_controls.go)

 Add FieldTypesControls(m Model, msg tea.Msg) (Model, tea.Cmd):
   up/down: cursor navigation through FieldTypesList
   'n': show create dialog (ShowFormDialogCmd with FORMDIALOGCREATEFIELDTYPE)
   'e': show edit dialog (ShowFormDialogCmd with FORMDIALOGEDITFIELDTYPE)
   'd': show delete confirmation dialog

 Add AdminFieldTypesControls(m Model, msg tea.Msg) (Model, tea.Cmd):
   Same structure, using AdminFieldTypesList and FORMDIALOGCREATEADMINFIELDTYPE / FORMDIALOGEDITADMINFIELDTYPE.

 Register both in PageSpecificMsgHandlers():
   case FIELDTYPES -> FieldTypesControls
   case ADMINFIELDTYPES -> AdminFieldTypesControls

 24. Add form dialog actions

 File: internal/cli/form_dialog.go

 Add FormDialogAction constants:
   FORMDIALOGCREATEFIELDTYPE       FormDialogAction = "create_field_type"
   FORMDIALOGEDITFIELDTYPE         FormDialogAction = "edit_field_type"
   FORMDIALOGCREATEADMINFIELDTYPE  FormDialogAction = "create_admin_field_type"
   FORMDIALOGEDITADMINFIELDTYPE    FormDialogAction = "edit_admin_field_type"

 25. Add fetch/update handlers

 File: internal/cli/admin_update_fetch.go

 Handle FieldTypesFetchMsg: call d.ListFieldTypes(), return FieldTypesFetchResultsMsg.
 Handle FieldTypesFetchResultsMsg: call FieldTypesSetCmd(msg.Data), LoadingStopCmd().
 Handle AdminFieldTypesFetchMsg: call d.ListAdminFieldTypes(), return AdminFieldTypesFetchResultsMsg.
 Handle AdminFieldTypesFetchResultsMsg: call AdminFieldTypesSetCmd(msg.Data), LoadingStopCmd().

 File: internal/cli/admin_update_cms.go

 Handle CRUD request messages for field_types:
   CreateFieldTypeFromDialogRequestMsg -> HandleCreateFieldTypeFromDialog() -> FieldTypeCreatedFromDialogMsg
   UpdateFieldTypeFromDialogRequestMsg -> HandleUpdateFieldTypeFromDialog() -> FieldTypeUpdatedFromDialogMsg
   DeleteFieldTypeRequestMsg -> HandleDeleteFieldType() -> FieldTypeDeletedMsg

 Handle CRUD request messages for admin_field_types:
   CreateAdminFieldTypeFromDialogRequestMsg -> HandleCreateAdminFieldTypeFromDialog() -> AdminFieldTypeCreatedFromDialogMsg
   UpdateAdminFieldTypeFromDialogRequestMsg -> HandleUpdateAdminFieldTypeFromDialog() -> AdminFieldTypeUpdatedFromDialogMsg
   DeleteAdminFieldTypeRequestMsg -> HandleDeleteAdminFieldType() -> AdminFieldTypeDeletedMsg

 Handle result messages (both entities reload their respective lists):
   FieldTypeCreatedFromDialogMsg -> FieldTypesFetchCmd()
   FieldTypeUpdatedFromDialogMsg -> FieldTypesFetchCmd()
   FieldTypeDeletedMsg -> FieldTypesFetchCmd()
   AdminFieldTypeCreatedFromDialogMsg -> AdminFieldTypesFetchCmd()
   AdminFieldTypeUpdatedFromDialogMsg -> AdminFieldTypesFetchCmd()
   AdminFieldTypeDeletedMsg -> AdminFieldTypesFetchCmd()

 File: internal/cli/admin_update_dialog.go

 Handle dialog handlers that perform DB operations for field_types:
   HandleCreateFieldTypeFromDialog: d.CreateFieldType(ctx, ac, params)
   HandleUpdateFieldTypeFromDialog: d.UpdateFieldType(ctx, ac, params)
   HandleDeleteFieldType: d.DeleteFieldType(ctx, ac, id)

 Handle dialog handlers for admin_field_types:
   HandleCreateAdminFieldTypeFromDialog: d.CreateAdminFieldType(ctx, ac, params)
   HandleUpdateAdminFieldTypeFromDialog: d.UpdateAdminFieldType(ctx, ac, params)
   HandleDeleteAdminFieldType: d.DeleteAdminFieldType(ctx, ac, id)

 26. Add menu entries

 File: internal/cli/menus.go

 Add m.PageMap[FIELDTYPES] to CmsMenuInit() (alongside DATATYPES, ROUTES, MEDIA, USERSADMIN).
 Add m.PageMap[ADMINFIELDTYPES] to AdminCmsMenuInit() (alongside ADMINCONTENT, ADMINDATATYPES, ADMINROUTES).

 ═══════════════════════════════════════════════════════════════
 PHASE 4: TypeScript SDKs
 ═══════════════════════════════════════════════════════════════

 27. Add branded IDs to @modulacms/types

 File: sdks/typescript/types/src/ids.ts

 Add:
   export type FieldTypeID = Brand<string, 'FieldTypeID'>
   export type AdminFieldTypeID = Brand<string, 'AdminFieldTypeID'>

 File: sdks/typescript/types/src/index.ts

 Add FieldTypeID and AdminFieldTypeID to the ids.js re-export block.

 28. Add entity types to @modulacms/types

 File: sdks/typescript/types/src/entities/schema.ts (add alongside existing Datatype, Field, DatatypeField definitions)

 Add entity types:
   export type FieldTypeInfo = {
     field_type_id: FieldTypeID
     type: string
     label: string
   }

   export type AdminFieldTypeInfo = {
     admin_field_type_id: AdminFieldTypeID
     type: string
     label: string
   }

 File: sdks/typescript/types/src/index.ts

 Add FieldTypeInfo and AdminFieldTypeInfo to the entity re-exports.

 29. Add types to admin SDK

 File: sdks/typescript/modulacms-admin-sdk/src/types/schema.ts

 Add re-exports and param types:
   Re-export: export type { FieldTypeInfo, AdminFieldTypeInfo } from '@modulacms/types'

   export type CreateFieldTypeParams = { type: string; label: string }
   export type UpdateFieldTypeParams = { field_type_id: FieldTypeID; type: string; label: string }
   export type CreateAdminFieldTypeParams = { type: string; label: string }
   export type UpdateAdminFieldTypeParams = { admin_field_type_id: AdminFieldTypeID; type: string; label: string }

 File: sdks/typescript/modulacms-admin-sdk/src/types/index.ts

 Add all new types to the schema.js re-export block.

 File: sdks/typescript/modulacms-admin-sdk/src/types/common.ts

 Add FieldTypeID and AdminFieldTypeID to the common re-exports from @modulacms/types.

 30. Wire resources to admin SDK client

 File: sdks/typescript/modulacms-admin-sdk/src/index.ts

 Add to the createAdminClient() return object (around line 748, alongside roles/permissions):

   fieldTypes: createResource<FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID>(http, 'fieldtypes'),
   adminFieldTypes: createResource<AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams, AdminFieldTypeID>(http, 'adminfieldtypes'),

 Add types to the ModulaCMSAdminClient type interface.
 Add new type imports and exports at the top-level export block.

 31. Build and verify

 Run: just sdk-build
 Run: just sdk-typecheck

 ═══════════════════════════════════════════════════════════════
 PHASE 5: Go SDK
 ═══════════════════════════════════════════════════════════════

 32. Add Go SDK types

 File: sdks/go/ids.go

 Add:
   type FieldTypeID string
   type AdminFieldTypeID string

 File: sdks/go/types.go

 Add entity and param types:
   type FieldTypeInfo struct {
       FieldTypeID FieldTypeID `json:"field_type_id"`
       Type        string      `json:"type"`
       Label       string      `json:"label"`
   }

   type CreateFieldTypeParams struct {
       Type  string `json:"type"`
       Label string `json:"label"`
   }

   type UpdateFieldTypeParams struct {
       FieldTypeID FieldTypeID `json:"field_type_id"`
       Type        string      `json:"type"`
       Label       string      `json:"label"`
   }

   Same pattern for AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams (using AdminFieldTypeID).

 33. Wire resources to Go SDK client

 File: sdks/go/modula.go

 Add to Client struct:
   FieldTypes      *Resource[FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID]
   AdminFieldTypes *Resource[AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams, AdminFieldTypeID]

 Add to NewClient() return:
   FieldTypes:      newResource[FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID](h, "/api/v1/fieldtypes"),
   AdminFieldTypes: newResource[AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams, AdminFieldTypeID](h, "/api/v1/adminfieldtypes"),

 34. Build and verify

 Run: just sdk-go-vet

 ═══════════════════════════════════════════════════════════════
 PHASE 6: Swift SDK
 ═══════════════════════════════════════════════════════════════

 35. Add Swift SDK types

 File: sdks/swift/Sources/Modula/IDs.swift

 Add:
   public struct FieldTypeID: ResourceID, Sendable { public let rawValue: String; public init(rawValue: String) { self.rawValue = rawValue } }
   public struct AdminFieldTypeID: ResourceID, Sendable { public let rawValue: String; public init(rawValue: String) { self.rawValue = rawValue } }

 File: sdks/swift/Sources/Modula/Types.swift

 Add entity and param types with CodingKeys:

   public struct FieldTypeInfo: Codable, Sendable {
       public let fieldTypeID: FieldTypeID
       public let type: String
       public let label: String
       enum CodingKeys: String, CodingKey {
           case fieldTypeID = "field_type_id"
           case type, label
       }
   }

   public struct CreateFieldTypeParams: Encodable, Sendable {
       public let type: String
       public let label: String
       public init(type: String, label: String) { self.type = type; self.label = label }
   }

   public struct UpdateFieldTypeParams: Encodable, Sendable {
       public let fieldTypeID: FieldTypeID
       public let type: String
       public let label: String
       public init(fieldTypeID: FieldTypeID, type: String, label: String) {
           self.fieldTypeID = fieldTypeID; self.type = type; self.label = label
       }
       enum CodingKeys: String, CodingKey {
           case fieldTypeID = "field_type_id"
           case type, label
       }
   }

   Same pattern for AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams.

 36. Wire resources to Swift SDK client

 File: sdks/swift/Sources/Modula/Client.swift

 Add to ModulaClient properties:
   public let fieldTypes: Resource<FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID>
   public let adminFieldTypes: Resource<AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams, AdminFieldTypeID>

 Add to init(config:):
   fieldTypes = Resource(path: "/api/v1/fieldtypes", http: http)
   adminFieldTypes = Resource(path: "/api/v1/adminfieldtypes", http: http)

 37. Build and verify

 Run: just sdk-swift-build

 ═══════════════════════════════════════════════════════════════
 PHASE 7: Admin Panel (React)
 ═══════════════════════════════════════════════════════════════

 The admin panel exposes field_types only. Admin field types are managed exclusively through the TUI and HTTP API -- the opinionated admin panel does not surface admin-prefixed schema tables. React Query hooks for admin_field_types are included (step 39) for programmatic use but no UI page is built.

 38. Add query key factory entries

 File: admin/src/lib/query-keys.ts

 Add:
   fieldTypes: {
     all: ['fieldTypes'] as const,
     list: () => [...queryKeys.fieldTypes.all, 'list'] as const,
   },
   adminFieldTypes: {
     all: ['adminFieldTypes'] as const,
     list: () => [...queryKeys.adminFieldTypes.all, 'list'] as const,
   },

 39. Add React Query hooks

 File: admin/src/queries/field-types.ts (new)

 Following the roles pattern from admin/src/queries/users.ts:

   useFieldTypes() -- sdk.fieldTypes.list()
   useCreateFieldType() -- sdk.fieldTypes.create(), invalidate fieldTypes.all
   useUpdateFieldType() -- sdk.fieldTypes.update(), invalidate fieldTypes.all
   useDeleteFieldType() -- sdk.fieldTypes.remove(), invalidate fieldTypes.all

   useAdminFieldTypes() -- sdk.adminFieldTypes.list()
   useCreateAdminFieldType() -- sdk.adminFieldTypes.create(), invalidate adminFieldTypes.all
   useUpdateAdminFieldType() -- sdk.adminFieldTypes.update(), invalidate adminFieldTypes.all
   useDeleteAdminFieldType() -- sdk.adminFieldTypes.remove(), invalidate adminFieldTypes.all

 All mutations use onError: (err) => toast.error(err.message).

 40. Add route page

 File: admin/src/routes/_admin/schema/field-types/index.tsx (new)

 A list page following the roles page pattern:

 - Fetch data with useFieldTypes()
 - DataTable with columns: type (string), label (string), actions (dropdown with edit/delete)
 - Create button opens a Dialog with two fields: Type (Input) and Label (Input)
 - Edit via inline dialog (pre-populated with selected row data)
 - Delete via ConfirmDialog
 - EmptyState fallback when list is empty

 Form state managed via useState (matching codebase convention, not React Hook Form).

 41. Add sidebar navigation

 File: admin/src/components/admin/sidebar.tsx

 Add a NavItem under the "Schema" section header (after Fields, before Media):

   <NavItem
     to="/schema/field-types"
     icon={List}
     label="Field Types"
     collapsed={collapsed}
   />

 Import List (or Tags) from lucide-react.

 ═══════════════════════════════════════════════════════════════
 FILES SUMMARY
 ═══════════════════════════════════════════════════════════════

 ┌────────────────┬──────────────────────────────────────────────────────────────────────────┐
 │     Action     │                                  Files                                  │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Create (SQL)   │ 12 SQL files in sql/schema/27_*/ and sql/schema/28_*/                   │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Create (Go)    │ internal/router/field_types.go                                          │
 │                │ internal/router/admin_field_types.go                                    │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Create (Test)  │ 4 test files in internal/db/                                            │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Create (React) │ admin/src/routes/_admin/schema/field-types/index.tsx                    │
 │                │ admin/src/queries/field-types.ts                                        │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Auto-generated │ internal/db-{sqlite,mysql,psql}/ (sqlc)                                 │
 │                │ internal/db/{field_type,admin_field_type}_gen.go (dbgen)                 │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Modify (Go)    │ internal/db/types/types_ids.go                                          │
 │                │ sql/sqlc.yml                                                            │
 │                │ tools/dbgen/definitions.go                                              │
 │                │ internal/db/db.go                                                       │
 │                │ internal/db/wipe.go                                                     │
 │                │ sql/all_schema*.sql (x3)                                                │
 │                │ internal/router/mux.go                                                  │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Modify (TUI)   │ internal/cli/pages.go                                                   │
 │                │ internal/cli/model.go                                                   │
 │                │ internal/cli/menus.go                                                   │
 │                │ internal/cli/admin_message_types.go                                     │
 │                │ internal/cli/admin_constructors.go                                      │
 │                │ internal/cli/panel_view.go                                              │
 │                │ internal/cli/admin_panel_view.go                                        │
 │                │ internal/cli/admin_controls.go                                          │
 │                │ internal/cli/form_dialog.go                                             │
 │                │ internal/cli/admin_update_fetch.go                                      │
 │                │ internal/cli/admin_update_cms.go                                        │
 │                │ internal/cli/admin_update_dialog.go                                     │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Modify (TS)    │ sdks/typescript/types/src/ids.ts                                        │
 │                │ sdks/typescript/types/src/index.ts                                      │
 │                │ sdks/typescript/types/src/entities/schema.ts                             │
 │                │ sdks/typescript/modulacms-admin-sdk/src/types/schema.ts                  │
 │                │ sdks/typescript/modulacms-admin-sdk/src/types/index.ts                   │
 │                │ sdks/typescript/modulacms-admin-sdk/src/types/common.ts                  │
 │                │ sdks/typescript/modulacms-admin-sdk/src/index.ts                         │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Modify (Go SDK)│ sdks/go/ids.go                                                          │
 │                │ sdks/go/types.go                                                        │
 │                │ sdks/go/modula.go                                                       │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Modify (Swift) │ sdks/swift/Sources/Modula/IDs.swift                                     │
 │                │ sdks/swift/Sources/Modula/Types.swift                                   │
 │                │ sdks/swift/Sources/Modula/Client.swift                                  │
 ├────────────────┼──────────────────────────────────────────────────────────────────────────┤
 │ Modify (React) │ admin/src/lib/query-keys.ts                                             │
 │                │ admin/src/components/admin/sidebar.tsx                                   │
 └────────────────┴──────────────────────────────────────────────────────────────────────────┘

 ═══════════════════════════════════════════════════════════════
 VERIFICATION
 ═══════════════════════════════════════════════════════════════

 Phase 1 (Database):
   1. just sqlc -- no errors
   2. go run ./tools/dbgen -- generates clean files
   3. just check -- compiles without errors
   4. just test -- all tests pass (existing + 4 new test files)

 Phase 2 (HTTP API):
   5. just check -- compiles with new handlers
   6. Manual: start server, curl all 10 endpoints, verify CRUD lifecycle

 Phase 3 (TUI):
   7. just check -- compiles with TUI changes
   8. Manual: SSH into TUI, navigate to CMS > Field Types, create/edit/delete a field type
   9. Manual: SSH into TUI, navigate to Admin CMS > Admin Field Types, create/edit/delete an admin field type

 Phase 4-6 (SDKs):
  10. just sdk-build -- all TypeScript packages build
  11. just sdk-typecheck -- no type errors
  12. just sdk-go-vet -- Go SDK passes vet
  13. just sdk-swift-build -- Swift SDK compiles

 Phase 7 (Admin Panel):
  14. cd admin && npm run build -- production build succeeds
  15. Manual: dev server, navigate to Schema > Field Types, verify CRUD UI
