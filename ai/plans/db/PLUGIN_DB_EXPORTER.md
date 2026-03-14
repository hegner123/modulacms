 Plan: Plugin Table Deploy Export/Import

 Context

 The deploy sync system exports/imports 12 hardcoded content tables as JSON. Plugin tables — created at runtime by plugins and registered in the tables CMS table — are excluded. Plugin
 tables hold user data that must be deployable between environments. The plugins already register their tables; the deploy system just needs to read them.

 No DDL export needed. Plugins create tables on init — if the plugin isn't installed on the destination, the data has nowhere to go (warn and skip). No general-purpose SQL dumper.

 Files Modified

 ┌─────────────────────────────┬───────────────────────────────────────────────────────────────────────────┐
 │            File             │                                  Change                                   │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/db/consts.go       │ IsValidTable + ValidateTableName accept plugin_ tables                    │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/db/deploy_ops.go   │ Add ColumnMeta, IntrospectColumns, QueryAllRows to DeployOps              │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/types.go    │ Add ExportOptions struct, PluginTables field on SyncManifest              │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/export.go   │ discoverPluginTables, isPluginTable, modify ExportPayload signature       │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/import.go   │ Plugin table truncate/insert after core tables, catalog-based coercion    │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/validate.go │ Scope ULID checks to core tables; basic row-width check for plugin tables │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/deploy.go   │ Update ExportToFile, Pull, Push signatures                                │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/server.go   │ Add include_plugins to exportRequest                                      │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/deploy/client.go   │ Forward IncludePlugins in Export                                          │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ cmd/deploy.go               │ Add --include-plugins flag to export/push/pull                            │
 ├─────────────────────────────┼───────────────────────────────────────────────────────────────────────────┤
 │ internal/tui/deploy_update.go│ Update Pull/Push calls to pass ExportOptions instead of []db.DBTable    │
 └─────────────────────────────┴───────────────────────────────────────────────────────────────────────────┘

 Steps

 1. Extend table validation (internal/db/consts.go)

 The IsValidTable check in TruncateTable and BulkInsert rejects plugin tables. Fix by accepting names with plugin_ prefix that pass the identifier safety check (ValidTableName in
 query_builder.go).

 // Add new function. Minimum length 9 is intentional: plugin tables follow the naming
 // convention plugin_<pluginname>_<tablename>, so the shortest valid name is "plugin_x_y" (10 chars).
 // len >= 9 allows "plugin_xy" as the absolute floor (single-segment names from tests).
 func IsValidPluginTableName(name string) bool {
     return len(name) >= 9 && name[:7] == "plugin_" && ValidTableName(name) == nil
 }

 // Modify IsValidTable to also accept plugin tables:
 func IsValidTable(t DBTable) bool {
     if _, ok := allTables[t]; ok {
         return true
     }
     return IsValidPluginTableName(string(t))
 }

 // Modify ValidateTableName to also accept plugin tables:
 func ValidateTableName(name string) (DBTable, error) {
     t := DBTable(name)
     if _, ok := allTables[t]; ok {
         return t, nil
     }
     if IsValidPluginTableName(name) {
         return t, nil
     }
     return "", fmt.Errorf("unknown table name: %q", name)
 }

 Also add a shared system plugin table set (used by both deploy discovery and plugin manager):
 var SystemPluginTables = map[string]bool{
     "plugin_routes": true, "plugin_hooks": true, "plugin_requests": true,
 }

 This unblocks TruncateTable and BulkInsert for plugin tables on all three backends without any changes to deploy_ops.go's existing methods.

 2. Add catalog introspection to DeployOps (internal/db/deploy_ops.go)

 Two new methods on the DeployOps interface + ColumnMeta type. Each method requires three
 implementations (sqliteDeployOps, psqlDeployOps, mysqlDeployOps) = 6 new functions total:

 type ColumnMeta struct {
     Name      string
     IsInteger bool // INTEGER, INT, BIGINT, SMALLINT, TINYINT, SERIAL
 }

 IntrospectColumns(ctx, table) ([]ColumnMeta, error) — query catalog for column names and types.
 - SQLite: PRAGMA table_info(<table>) — columns: cid, name, type, notnull, dflt_value, pk. Check type for INTEGER/INT.
 - PostgreSQL: SELECT column_name, data_type FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 ORDER BY ordinal_position
 - MySQL: SELECT COLUMN_NAME, DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? ORDER BY ORDINAL_POSITION

 Returns error if table doesn't exist (used as existence check on import).

 QueryAllRows(ctx, table) ([]string, [][]any, error) — SELECT * and return column names + all row values.
 - All backends: SELECT * FROM <table> with rows.Columns() for column names
 - Validate table via IsValidTable before executing
 - Scan strategy: allocate a []any slice of *sql.NullString targets for every column. After scanning
   each row, convert values: for columns where IntrospectColumns reports IsInteger=true, parse the
   NullString.String via strconv.ParseInt and store as int64. For all other columns, store as string
   (or nil if NullString.Valid is false). This produces consistent Go types across all three backends
   regardless of driver-specific scan behavior. The JSON round-trip is: int64 columns serialize as
   JSON numbers and are coerced back from float64 on import via coerceWithIntMap. Text columns
   serialize as JSON strings and pass through unchanged.
 - Each concrete QueryAllRows (e.g., sqliteDeployOps.QueryAllRows) calls its own IntrospectColumns
   method on the same receiver to get the IsInteger map for scan conversion.
 - Rationale for dual coercion: QueryAllRows converts integer columns to int64 during scan so the
   exported JSON contains numbers (not strings). On import, JSON decoding turns those numbers back
   to float64, so coerceWithIntMap must convert float64 -> int64 using the destination catalog.
   Both sides are required; neither is redundant.

 The catalog introspection patterns already exist in internal/plugin/schema_api.go at line 467
 (introspectTableColumns function, used for schema drift detection via checkSchemaDrift at line 546).

 3. Add types (internal/deploy/types.go)

 type ExportOptions struct {
     Tables         []db.DBTable // core tables; nil = DefaultTableSet
     IncludePlugins bool         // discover and include registered plugin tables
 }

 Add field to SyncManifest:
 PluginTables []string `json:"plugin_tables,omitempty"` // subset of Tables that are plugin tables

 4. Modify export (internal/deploy/export.go)

 Add helper (delegates to db package for single source of truth on plugin table name validation):
 func isPluginTable(name string) bool {
     return db.IsValidPluginTableName(name)
 }

 Add discovery:
 func discoverPluginTables(driver db.DbDriver) ([]db.DBTable, error) {
     tables, err := driver.ListTables()
     if err != nil { return nil, err }

     // SystemPluginTables is defined as a package-level var in db/consts.go so both
     // the plugin manager (ListOrphanedTables) and deploy discovery share the same list.
     // var SystemPluginTables = map[string]bool{
     //     "plugin_routes": true, "plugin_hooks": true, "plugin_requests": true,
     // }
     system := db.SystemPluginTables

     var result []db.DBTable
     for _, t := range *tables {
         if !isPluginTable(t.Label) || system[t.Label] { continue }
         result = append(result, db.DBTable(t.Label))
     }
     sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
     return result, nil
 }

 Change ExportPayload signature from (ctx, driver, []db.DBTable) to (ctx, driver, ExportOptions):

 After the existing core table export loop, add:
 if opts.IncludePlugins {
     ops, err := db.NewDeployOps(driver)
     if err != nil {
         return nil, fmt.Errorf("export: create deploy ops for plugin tables: %w", err)
     }
     pluginTables, err := discoverPluginTables(driver)
     if err != nil {
         return nil, fmt.Errorf("export: discover plugin tables: %w", err)
     }
     var pluginNames []string

     for _, pt := range pluginTables {
         cols, rows, err := ops.QueryAllRows(ctx, pt)
         if err != nil { /* warn and skip */ continue }

         td := TableData{Columns: cols, Rows: rows}
         name := string(pt)
         tableDataMap[name] = td
         tableNames = append(tableNames, name)
         rowCounts[name] = len(rows)
         pluginNames = append(pluginNames, name)
     }

     manifest.PluginTables = pluginNames
 }

 Move the collectUserRefs call (currently at export.go:69) to AFTER the plugin table export block
 so that plugin table data is included in user ref collection. Inside collectUserRefs, use
 isPluginTable(tableName) to skip user_id columns for plugin tables — only collect author_id.
 See Step 6 for rationale on why user_id must be skipped for plugin tables.

 ImportFromFile (deploy.go:56) requires no changes — it calls ImportPayload which handles plugin
 tables via the payload content.

 5. Modify import (internal/deploy/import.go)

 Inside the ImportPayload function's callback passed to ops.ImportAtomic (import.go:76-124),
 after the core table BulkInsert loop (line ~113) and before VerifyForeignKeys (line ~116),
 add plugin table handling:

 // Identify plugin tables in payload.
 var pluginTableNames []string
 for name := range payload.Tables {
     if isPluginTable(name) {
         pluginTableNames = append(pluginTableNames, name)
     }
 }
 sort.Strings(pluginTableNames)

 // Truncate existing plugin tables in reverse sorted order (matches core table convention;
 // FKs are disabled within the atomic block so order is not critical).
 for i := len(pluginTableNames) - 1; i >= 0; i-- {
     name := pluginTableNames[i]
     t := db.DBTable(name)
     // Check if table exists on destination.
     if _, iErr := ops.IntrospectColumns(ctx, t); iErr != nil {
         warnings = append(warnings, fmt.Sprintf("plugin table %q not on destination (plugin not installed); skipping", name))
         continue
     }
     ops.TruncateTable(ctx, ex, t)
 }

 // Insert plugin table data.
 for _, name := range pluginTableNames {
     td := payload.Tables[name]
     if len(td.Rows) == 0 { continue }
     t := db.DBTable(name)

     // Check destination has this table + get column types for coercion.
     destCols, iErr := ops.IntrospectColumns(ctx, t)
     if iErr != nil { continue } // already warned above

     // Schema mismatch check.
     destColNames := extractColNames(destCols)
     if !columnsMatch(td.Columns, destColNames) {
         warnings = append(warnings, fmt.Sprintf("plugin table %q schema mismatch; skipping", name))
         continue
     }

     // Catalog-based coercion (float64 -> int64 for integer columns).
     intCols := buildIntColumnMap(td.Columns, destCols)
     rows := coerceWithIntMap(td.Rows, intCols)

     ops.BulkInsert(ctx, ex, t, td.Columns, rows)
 }

 New helper functions (all in internal/deploy/import.go):

 // extractColNames returns just the Name field from each ColumnMeta.
 func extractColNames(cols []db.ColumnMeta) []string

 // columnsMatch returns true if both slices contain the same column names
 // in the same order. Length and order must match exactly.
 func columnsMatch(a, b []string) bool

 // buildIntColumnMap returns a map of column indices (in cols) that correspond
 // to integer columns according to destMeta. Used for JSON float64 -> int64 coercion.
 func buildIntColumnMap(cols []string, destMeta []db.ColumnMeta) map[int]bool

 // coerceWithIntMap converts float64 values to int64 at the column indices in intCols.
 // Extracted from the existing coerceRows function's inner loop.
 func coerceWithIntMap(rows [][]any, intCols map[int]bool) [][]any

 Refactor coerceRows: Extract the coercion loop into coerceWithIntMap. The existing coerceRows
 builds intCols from TableStructMap then calls coerceWithIntMap. Do not change coerceRows behavior
 — only extract the shared logic.

 6. Modify validation (internal/deploy/validate.go)

 In ValidatePayload:
 - Check 4 (ULID validation): Skip plugin tables. Plugin id columns are TEXT ULIDs, but user-defined _id columns may not be ULIDs.
 for tableName, td := range payload.Tables {
     if isPluginTable(tableName) { continue } // skip ULID check for plugin tables
     // ... existing ULID check
 }
 - Checks 5-6 (content FK + tree pointers): Already scoped to specific core table names. No change needed.
 - Check 7 (user refs): `validateUserRefs` (validate.go:321) already only checks `author_id` — no
   `user_id` handling exists there, so no validation change is needed. However, `collectUserRefs`
   (export.go:183) collects both `author_id` and `user_id` columns during export. For plugin tables,
   `collectUserRefs` must skip `user_id` columns — plugins may define `user_id` columns that
   reference external user systems, not CMS users. Keep `author_id` collection for plugin tables
   (plugin tables may or may not have `author_id` — it is NOT auto-injected; the auto-injected
   reserved columns are `id`, `created_at`, `updated_at` only).
 - Add check 8: Row width consistency for plugin tables.
 for tableName, td := range payload.Tables {
     if !isPluginTable(tableName) { continue }
     for i, row := range td.Rows {
         if len(row) != len(td.Columns) {
             errs = append(errs, SyncError{Table: tableName, Phase: "validate", ...})
         }
     }
 }

 In VerifyImport: IsValidTable now accepts plugin tables (Step 1), so the post-import row count verification works for plugin tables without changes.

 7. Update callers (internal/deploy/deploy.go)

 Update function signatures:
 - ExportToFile(ctx, driver, tables, outPath) → ExportToFile(ctx, driver, opts ExportOptions, outPath)
 - Pull(ctx, cfg, driver, envName, tables, skipBackup, dryRun) → Pull(ctx, cfg, driver, envName, opts ExportOptions, skipBackup, dryRun)
 - Push(ctx, cfg, driver, envName, tables, dryRun) → Push(ctx, cfg, driver, envName, opts ExportOptions, dryRun)

 Each passes opts through to ExportPayload.

 Also update internal/tui/deploy_update.go which calls Pull (line 108) and Push (line 153).
 Both currently pass nil for the tables parameter — change to deploy.ExportOptions{} (empty options,
 meaning DefaultTableSet with no plugin tables, preserving current TUI behavior).

 8. Update server (internal/deploy/server.go)

 type exportRequest struct {
     Tables         []string `json:"tables"`
     IncludePlugins bool     `json:"include_plugins"`
 }

 In DeployExportHandler, build ExportOptions from request and pass to ExportPayload.

 9. Update client (internal/deploy/client.go)

 Change Export(ctx, tables) to Export(ctx, opts ExportOptions). Serialize include_plugins in the request body.

 10. Update CLI (cmd/deploy.go)

 Add --include-plugins flag to deployExportCmd, deployPullCmd, deployPushCmd:
 deployExportCmd.Flags().Bool("include-plugins", false, "Include plugin table data in export")

 Read the flag and build ExportOptions in each RunE.

 11. Tests

 Test in internal/db/deploy_ops_test.go:
 - TestIntrospectColumns — SQLite with a test plugin_test_items table, verify column names and IsInteger detection
 - TestQueryAllRows — verify row scanning with TEXT, INTEGER, REAL columns
 - TestIsValidPluginTableName — valid/invalid names, edge cases

 Test in internal/deploy/ (new file deploy_plugin_test.go):
 - TestExportWithPlugins — create plugin table, export with IncludePlugins: true, verify payload
 - TestExportWithoutPlugins — export without flag, verify no plugin tables
 - TestImportPluginTable_Exists — import payload with plugin table, verify data
 - TestImportPluginTable_Missing — verify warning when table doesn't exist on destination
 - TestImportPluginTable_SchemaMismatch — column differences, verify skip + warning
 - TestImportPluginTable_IntCoercion — float64 -> int64 for INTEGER columns via catalog

 Edge Cases

 ┌─────────────────────────────────────────────┬─────────────────────────────────────────────────────────┐
 │                    Case                     │                        Behavior                         │
 ├─────────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
 │ Empty plugin table (0 rows)                 │ Exported with empty rows; imported as truncate-only     │
 ├─────────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
 │ Plugin on source, not destination           │ Warning emitted, table skipped during import            │
 ├─────────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
 │ Plugin version mismatch (different columns) │ Warning emitted, table skipped during import            │
 ├─────────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
 │ System plugin tables (plugin_routes, etc.)  │ Filtered out during discovery                           │
 ├─────────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
 │ Plugin table with INTEGER columns           │ Catalog-based introspection for float64->int64 coercion │
 ├─────────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
 │ No plugins installed                        │ discoverPluginTables returns empty list; no-op          │
 ├─────────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
 │ Plugin table with user_id column            │ user_id NOT collected as user ref (may be external);    │
 │ (non-CMS user reference)                    │ author_id collected if present (not auto-injected)      │
 └─────────────────────────────────────────────┴─────────────────────────────────────────────────────────┘

 Do Not Modify

 These functions must not be changed beyond what this plan specifies:
 - structSliceToTableData, columnsFromType, serializeField (core table serialization — untouched)
 - DefaultTableSet, tableListFuncs (core table enumeration — untouched)
 - coerceRows (refactor only — extract coerceWithIntMap helper, do not change existing behavior)
 - resolveUserRefs, ImportFromFile, BuildDryRunResult (no changes needed)

 Verification

 1. just check — compile verification after each step
 2. just test — run full test suite
 3. Manual: create a plugin table via Lua, populate it, export with --include-plugins, verify JSON contains plugin table data, import into a fresh DB with the plugin installed, verify data
