 Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Plan: Decompose DbDriver Interface into Repository Interfaces

 Context

 The DbDriver interface in internal/db/db.go has ~288 methods. This makes it untestable (cannot mock that many methods), creates high coupling (every new entity adds methods to the interface
 and all three wrappers), and creates cognitive load (3,199-line file defining the entire data layer contract). The goal is to decompose it into focused repository interfaces while keeping
 the config-based tri-database selection pattern unchanged.

 Approach

 Phase 1 (this plan): Create repository interfaces, rewrite DbDriver as a composed embedding, and narrow the one consumer that benefits most (PermissionCache). All other consumers continue
 working unchanged.

 Phase 2 (future, separate plan): Narrow handler and TUI consumer signatures to accept repository interfaces (or consumer-defined interfaces where appropriate) instead of DbDriver. Note: GenericList in internal/db/generic_list.go calls methods across nearly every repository group and will be a hard case for Phase 2 narrowing.

 Step 0: Create repository interfaces and rewrite DbDriver

 Create internal/db/repositories.go with repository interfaces. Group every method in the interface until none remain ungrouped. If a method is missed during implementation, add it to the appropriate repository.

 ┌─────────────────────────────┬────────────────────────────────────────────┐
 │          Interface          │               Source Groups                │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ SchemaRepository            │ Connection (lifecycle subset)              │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ ConnectionRepository        │ Connection (raw SQL subset)                │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ ContentDataRepository       │ ContentData + ContentRelations             │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ ContentFieldRepository      │ ContentFields                              │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ AdminContentDataRepository  │ AdminContentData + AdminContentRelations   │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ AdminContentFieldRepository │ AdminContentFields                         │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ DatatypeRepository          │ Datatypes + DatatypeFields                 │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ AdminDatatypeRepository     │ AdminDatatypes + AdminDatatypeFields       │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ FieldRepository             │ Fields                                     │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ AdminFieldRepository        │ AdminFields                                │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ RouteRepository             │ Routes                                     │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ AdminRouteRepository        │ AdminRoutes                                │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ MediaRepository             │ Media + MediaDimensions                    │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ UserRepository              │ Users + UserOauths + UserSshKeys           │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ AuthRepository              │ Sessions + Tokens                          │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ RBACRepository              │ Roles + Permissions + RolePermissions      │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ BackupRepository            │ Backups + BackupSets + BackupVerifications │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ ChangeEventRepository       │ ChangeEvents                               │
 ├─────────────────────────────┼────────────────────────────────────────────┤
 │ TableRepository             │ Tables                                     │
 └─────────────────────────────┴────────────────────────────────────────────┘

 Consolidation rationale:
 - DatatypeFields is never used without Datatypes - every consumer uses both
 - MediaDimensions is exclusively used alongside Media
 - BackupSets/Verifications are only meaningful in backup workflows
 - Roles/Permissions/RolePermissions form one RBAC domain
 - ContentRelations/AdminContentRelations belong with their content data entities
 - Admin + non-admin of the same entity stay separate (different ID types, never mixed in a handler)

 Then rewrite DbDriver in db.go:

 type DbDriver interface {
     SchemaRepository
     ConnectionRepository
     ContentDataRepository
     ContentFieldRepository
     AdminContentDataRepository
     AdminContentFieldRepository
     DatatypeRepository
     AdminDatatypeRepository
     FieldRepository
     AdminFieldRepository
     RouteRepository
     AdminRouteRepository
     MediaRepository
     UserRepository
     AuthRepository
     RBACRepository
     BackupRepository
     ChangeEventRepository
     TableRepository
 }

 CreateXxxTable() methods stay with their entity repository (not SchemaRepository) because CreateAllTables() calls them via the concrete receiver.

 Verification: just check compiles. All three wrapper structs already implement every method, so they automatically satisfy the composed interface. Zero runtime change.

 Step 1: Narrow PermissionCache to RBACRepository

 Change in internal/middleware/authorization.go:
 - func (pc *PermissionCache) Load(driver db.DbDriver) -> Load(driver db.RBACRepository)
 - func (pc *PermissionCache) StartPeriodicRefresh(ctx, driver db.DbDriver, ...) -> StartPeriodicRefresh(ctx, driver db.RBACRepository, ...)

 Callers in cmd/serve.go (lines 193, 198, 348, 353) pass driver which is DbDriver. Since DbDriver embeds RBACRepository, all call sites work without changes.

 Verification: just check && just test

 Files Modified

 ┌──────────────────────────────────────┬──────────────────────────────────────────────────────────┐
 │                 File                 │                          Change                          │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/db/repositories.go          │ NEW - 19 repository interface definitions                │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/db/db.go                    │ Replace method list with 19 embedded interfaces          │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ internal/middleware/authorization.go │ Narrow Load and StartPeriodicRefresh parameter types     │
 └──────────────────────────────────────┴──────────────────────────────────────────────────────────┘

 Files NOT Modified

 - All wrapper files (content_data.go, user.go, etc.) - unchanged
 - internal/db/init.go - unchanged (InitDB/ConfigDB return DbDriver)
 - cmd/serve.go - unchanged (passes DbDriver which satisfies RBACRepository)
 - internal/cli/ - unchanged (Model.DB stays DbDriver)
 - internal/router/ - unchanged (handlers call ConfigDB())
 - All test files - unchanged (use concrete Database struct)
 - sqlc-generated code - unchanged

 Verification

 1. just check - compilation passes (interface satisfaction verified by compiler)
 2. just test - all existing tests pass (no runtime behavior change)
 3. just lint-go - no new warnings
