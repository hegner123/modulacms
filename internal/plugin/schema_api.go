package plugin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"
)

// TableDefinition holds the parsed and validated schema for a plugin-defined table.
// Populated by luaDefineTable after parsing the Lua table argument.
type TableDefinition struct {
	PluginName  string
	TableName   string // without prefix
	FullName    string // plugin_<name>_<table>
	Columns     []db.CreateColumnDef
	Indexes     []db.IndexDef
	ForeignKeys []db.ForeignKeyDef
}

// reservedColumns are auto-injected by schema_api.go and cannot be defined by plugins.
// Plugins that include these names in their columns list receive a validation error.
var reservedColumns = map[string]bool{
	"id":         true,
	"created_at": true,
	"updated_at": true,
}

// tablePrefix returns the full table prefix for a plugin: "plugin_<name>_".
func tablePrefix(pluginName string) string {
	return "plugin_" + pluginName + "_"
}

// luaDefineTable returns a Go function that implements db.define_table(tableName, definition).
// It parses the Lua arguments, validates and prefixes the table name, auto-injects
// id/created_at/updated_at columns, validates FK namespace isolation, and executes
// DDLCreateTable with IfNotExists: true.
//
// tableReg may be nil; when non-nil, the table is registered in the CMS tables registry
// after successful DDL. Registration failure is logged but does not fail the define_table call.
//
// The context for DDL calls comes from L.Context() so that execution timeouts apply.
func luaDefineTable(L *lua.LState, pluginName string, exec db.Executor, dialect db.Dialect, tableReg TableRegistrar) lua.LGFunction {
	prefix := tablePrefix(pluginName)

	return func(L *lua.LState) int {
		tableName := L.CheckString(1)
		defTable := L.CheckTable(2)

		// Validate the raw table name before prefixing.
		if tableName == "" {
			L.ArgError(1, "table name cannot be empty")
			return 0
		}

		// Build the full prefixed table name and validate it.
		fullName := prefix + tableName
		if err := db.ValidTableName(fullName); err != nil {
			L.ArgError(1, fmt.Sprintf("invalid table name %q: %s", tableName, err.Error()))
			return 0
		}

		// Parse columns (required).
		columnsVal := L.GetField(defTable, "columns")
		columnsTbl, ok := columnsVal.(*lua.LTable)
		if !ok || columnsVal == lua.LNil {
			L.ArgError(2, "definition must include a 'columns' table")
			return 0
		}

		userColumns, err := parseColumns(L, columnsTbl, pluginName)
		if err != nil {
			L.ArgError(2, err.Error())
			return 0
		}
		if len(userColumns) == 0 {
			L.ArgError(2, "columns cannot be empty")
			return 0
		}

		// Auto-inject id as the first column, created_at and updated_at as the last columns.
		// These three columns are always present on plugin-defined tables.
		columns := make([]db.CreateColumnDef, 0, len(userColumns)+3)
		columns = append(columns, db.CreateColumnDef{
			Name:       "id",
			Type:       db.ColText,
			NotNull:    true,
			PrimaryKey: true,
		})
		columns = append(columns, userColumns...)
		columns = append(columns,
			db.CreateColumnDef{
				Name:    "created_at",
				Type:    db.ColText,
				NotNull: true,
			},
			db.CreateColumnDef{
				Name:    "updated_at",
				Type:    db.ColText,
				NotNull: true,
			},
		)

		// Parse indexes (optional).
		var indexes []db.IndexDef
		indexesVal := L.GetField(defTable, "indexes")
		if indexesTbl, ok := indexesVal.(*lua.LTable); ok && indexesVal != lua.LNil {
			indexes, err = parseIndexes(L, indexesTbl)
			if err != nil {
				L.ArgError(2, err.Error())
				return 0
			}
		}

		// Parse foreign keys (optional).
		var foreignKeys []db.ForeignKeyDef
		fkVal := L.GetField(defTable, "foreign_keys")
		if fkTbl, ok := fkVal.(*lua.LTable); ok && fkVal != lua.LNil {
			foreignKeys, err = parseForeignKeys(L, fkTbl, prefix)
			if err != nil {
				L.ArgError(2, err.Error())
				return 0
			}
		}

		// Build DDL params and execute.
		params := db.DDLCreateTableParams{
			Table:       fullName,
			Columns:     columns,
			Indexes:     indexes,
			ForeignKeys: foreignKeys,
			IfNotExists: true,
		}

		ctx := L.Context()
		if ctx == nil {
			// Defensive: if no context is set, this is a programming error in the
			// VM setup. Use a background context to avoid nil dereference, but this
			// path should not be reached in normal operation.
			L.RaiseError("define_table: no context set on Lua state (VM setup error)")
			return 0
		}

		if err := db.DDLCreateTable(ctx, exec, dialect, params); err != nil {
			L.RaiseError("define_table %q: %s", tableName, err.Error())
			return 0
		}

		// Register the plugin-defined table in the CMS tables registry.
		// Failure is advisory — the DDL succeeded, so the table exists.
		if tableReg != nil {
			if regErr := tableReg.RegisterTable(ctx, fullName); regErr != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("plugin %q: failed to register table %q in tables registry: %s",
						pluginName, fullName, regErr.Error()),
					nil,
				)
			}
		}

		// Phase 4: Schema drift detection (Known Gap #9).
		// After successful DDL, introspect actual columns and compare with
		// the expected definition. Drift is advisory only (logs at slog.Warn,
		// does not error). Drift results are stored on the PluginInstance
		// for surface in the admin PluginInfoHandler (S7).
		expectedCols := make([]string, 0, len(columns))
		for _, col := range columns {
			expectedCols = append(expectedCols, col.Name)
		}
		actualCols, introErr := introspectTableColumns(ctx, exec, dialect, fullName)
		if introErr != nil {
			// Non-fatal: log and continue.
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: schema introspection failed for table %q: %s",
					pluginName, fullName, introErr.Error()),
				nil,
			)
		} else {
			drifts := checkSchemaDrift(pluginName, fullName, expectedCols, actualCols)
			if len(drifts) > 0 {
				// Store drift entries on the registry for the PluginInfoHandler.
				// The registry key "__schema_drift" holds a Lua table of drift
				// descriptions. Since drift is read by Go code (not Lua), we
				// store it via a Go-side mechanism: the pluginName is used to
				// look up the PluginInstance from the Manager, but we do not have
				// a reference to the instance here. Instead, we store the drift
				// entries in a Lua global that the manager reads after loading.
				driftTbl := L.NewTable()
				for i, d := range drifts {
					entry := L.NewTable()
					L.SetField(entry, "table", lua.LString(d.Table))
					L.SetField(entry, "kind", lua.LString(d.Kind))
					L.SetField(entry, "column", lua.LString(d.Column))
					driftTbl.RawSetInt(i+1, entry)
				}
				L.SetGlobal("__schema_drift", driftTbl)
			}
		}

		return 0
	}
}

// parseColumns parses the Lua columns sequence table into Go CreateColumnDef slices.
// Each entry must be a table with:
//
//	name (string, required) -- validated via db.ValidColumnName()
//	type (string, required) -- validated via db.ValidateColumnType()
//	not_null (bool, optional)
//	default (string or number, optional)
//	unique (bool, optional)
//
// Reserved column names (id, created_at, updated_at) are rejected.
func parseColumns(L *lua.LState, tbl *lua.LTable, pluginName string) ([]db.CreateColumnDef, error) {
	var columns []db.CreateColumnDef

	var parseErr error
	tbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}

		colTbl, ok := value.(*lua.LTable)
		if !ok {
			parseErr = fmt.Errorf("each column must be a table, got %s", value.Type())
			return
		}

		// Extract name (required).
		nameVal := L.GetField(colTbl, "name")
		name, ok := nameVal.(lua.LString)
		if !ok || nameVal == lua.LNil {
			parseErr = fmt.Errorf("column missing required 'name' field")
			return
		}
		colName := string(name)

		// Reject reserved column names.
		if reservedColumns[colName] {
			parseErr = fmt.Errorf("column %q is auto-injected and cannot be defined manually", colName)
			return
		}

		// Validate column name.
		if err := db.ValidColumnName(colName); err != nil {
			parseErr = fmt.Errorf("invalid column name %q: %s", colName, err.Error())
			return
		}

		// Extract type (required).
		typeVal := L.GetField(colTbl, "type")
		typStr, ok := typeVal.(lua.LString)
		if !ok || typeVal == lua.LNil {
			parseErr = fmt.Errorf("column %q missing required 'type' field", colName)
			return
		}
		colType := string(typStr)

		// Validate column type.
		if err := db.ValidateColumnType(colType); err != nil {
			parseErr = fmt.Errorf("column %q: %s", colName, err.Error())
			return
		}

		// Extract optional fields.
		notNull := luaBoolField(L, colTbl, "not_null")
		unique := luaBoolField(L, colTbl, "unique")
		defaultVal := luaStringOrNumberField(L, colTbl, "default")

		columns = append(columns, db.CreateColumnDef{
			Name:    colName,
			Type:    db.ColumnType(colType),
			NotNull: notNull,
			Unique:  unique,
			Default: defaultVal,
		})
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return columns, nil
}

// parseIndexes parses the Lua indexes sequence table into Go IndexDef slices.
// Each entry must be a table with:
//
//	columns (sequence of strings, required)
//	unique (bool, optional)
func parseIndexes(L *lua.LState, tbl *lua.LTable) ([]db.IndexDef, error) {
	var indexes []db.IndexDef

	var parseErr error
	tbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}

		idxTbl, ok := value.(*lua.LTable)
		if !ok {
			parseErr = fmt.Errorf("each index must be a table, got %s", value.Type())
			return
		}

		// Extract columns (required).
		colsVal := L.GetField(idxTbl, "columns")
		colsTbl, ok := colsVal.(*lua.LTable)
		if !ok || colsVal == lua.LNil {
			parseErr = fmt.Errorf("index missing required 'columns' field")
			return
		}

		var colNames []string
		colsTbl.ForEach(func(_, v lua.LValue) {
			if parseErr != nil {
				return
			}
			s, ok := v.(lua.LString)
			if !ok {
				parseErr = fmt.Errorf("index column name must be a string, got %s", v.Type())
				return
			}
			colNames = append(colNames, string(s))
		})
		if parseErr != nil {
			return
		}

		if len(colNames) == 0 {
			parseErr = fmt.Errorf("index columns cannot be empty")
			return
		}

		unique := luaBoolField(L, idxTbl, "unique")

		indexes = append(indexes, db.IndexDef{
			Columns: colNames,
			Unique:  unique,
		})
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return indexes, nil
}

// parseForeignKeys parses the Lua foreign_keys sequence table into Go ForeignKeyDef slices.
// Each entry must be a table with:
//
//	column (string, required) -- local column name
//	ref_table (string, required) -- must start with the plugin's table prefix for namespace isolation
//	ref_column (string, required) -- referenced column name
//	on_delete (string, optional) -- CASCADE, SET NULL, or RESTRICT
//
// FK namespace isolation: ref_table must start with the same plugin_<name>_ prefix.
// This prevents plugins from referencing core CMS tables or other plugins' tables.
func parseForeignKeys(L *lua.LState, tbl *lua.LTable, prefix string) ([]db.ForeignKeyDef, error) {
	var fks []db.ForeignKeyDef

	var parseErr error
	tbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}

		fkTbl, ok := value.(*lua.LTable)
		if !ok {
			parseErr = fmt.Errorf("each foreign_key must be a table, got %s", value.Type())
			return
		}

		// Extract column (required).
		colVal := L.GetField(fkTbl, "column")
		colStr, ok := colVal.(lua.LString)
		if !ok || colVal == lua.LNil {
			parseErr = fmt.Errorf("foreign_key missing required 'column' field")
			return
		}

		// Extract ref_table (required).
		refTableVal := L.GetField(fkTbl, "ref_table")
		refTableStr, ok := refTableVal.(lua.LString)
		if !ok || refTableVal == lua.LNil {
			parseErr = fmt.Errorf("foreign_key missing required 'ref_table' field")
			return
		}
		refTable := string(refTableStr)

		// FK namespace isolation: ref_table must have the same plugin prefix.
		if !strings.HasPrefix(refTable, prefix) {
			parseErr = fmt.Errorf("foreign_key ref_table %q must start with plugin prefix %q (namespace isolation)", refTable, prefix)
			return
		}

		// Extract ref_column (required).
		refColVal := L.GetField(fkTbl, "ref_column")
		refColStr, ok := refColVal.(lua.LString)
		if !ok || refColVal == lua.LNil {
			parseErr = fmt.Errorf("foreign_key missing required 'ref_column' field")
			return
		}

		// Extract on_delete (optional).
		onDelete := ""
		onDeleteVal := L.GetField(fkTbl, "on_delete")
		if s, ok := onDeleteVal.(lua.LString); ok {
			onDelete = strings.ToUpper(string(s))
		}

		fks = append(fks, db.ForeignKeyDef{
			Column:    string(colStr),
			RefTable:  refTable,
			RefColumn: string(refColStr),
			OnDelete:  onDelete,
		})
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return fks, nil
}

// luaBoolField reads an optional boolean field from a Lua table.
// Returns false if the field is absent or not a boolean.
func luaBoolField(L *lua.LState, tbl *lua.LTable, field string) bool {
	val := L.GetField(tbl, field)
	if b, ok := val.(lua.LBool); ok {
		return bool(b)
	}
	return false
}

// luaStringOrNumberField reads an optional field that may be a string or number.
// Returns the value as a string suitable for the DDL Default field.
// Returns empty string if the field is absent or nil.
func luaStringOrNumberField(L *lua.LState, tbl *lua.LTable, field string) string {
	val := L.GetField(tbl, field)
	switch v := val.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return v.String()
	default:
		return ""
	}
}

// DriftEntry represents a single schema drift finding for a plugin table.
// Stored on PluginInstance.SchemaDrift and surfaced via the admin API (S7).
type DriftEntry struct {
	Table  string `json:"table"`  // full prefixed table name
	Kind   string `json:"kind"`   // "missing" or "extra"
	Column string `json:"column"` // the column name that differs
}

// introspectTableColumns queries the database for actual column names of the
// given table. Returns the column names in the order reported by the database.
func introspectTableColumns(ctx context.Context, exec db.Executor, dialect db.Dialect, tableName string) ([]string, error) {
	var query string
	var args []any

	switch dialect {
	case db.DialectSQLite:
		// PRAGMA table_info cannot be parameterized. tableName is already validated
		// via db.ValidTableName before reaching here, so identifier injection is
		// not possible. We construct the query with the validated name directly.
		query = "PRAGMA table_info(" + tableName + ")"
	case db.DialectMySQL:
		query = "SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		args = []any{tableName}
	case db.DialectPostgres:
		query = "SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 ORDER BY ordinal_position"
		args = []any{tableName}
	default:
		return nil, fmt.Errorf("unsupported dialect for introspection: %d", dialect)
	}

	// Use the underlying *sql.DB or *sql.Tx via the Executor interface.
	// Executor only defines ExecContext and QueryContext, but we need QueryContext.
	type queryContexter interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}
	qc, ok := exec.(queryContexter)
	if !ok {
		return nil, fmt.Errorf("executor does not support QueryContext")
	}

	rows, err := qc.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("introspecting table %q: %w", tableName, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("schema introspection rows.Close error: %s", cerr.Error()),
				nil,
			)
		}
	}()

	var columns []string

	if dialect == db.DialectSQLite {
		// PRAGMA table_info returns: cid, name, type, notnull, dflt_value, pk
		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dfltValue sql.NullString
			if scanErr := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); scanErr != nil {
				return nil, fmt.Errorf("scanning PRAGMA table_info: %w", scanErr)
			}
			columns = append(columns, name)
		}
	} else {
		// MySQL and PostgreSQL: single column result (column_name).
		for rows.Next() {
			var name string
			if scanErr := rows.Scan(&name); scanErr != nil {
				return nil, fmt.Errorf("scanning column name: %w", scanErr)
			}
			columns = append(columns, name)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating column names: %w", err)
	}

	return columns, nil
}

// checkSchemaDrift compares expected column names (from the plugin definition)
// with actual column names (from the database) and returns any drift entries.
//
// S7: Drift warnings use slog.Warn (not Info/Debug).
func checkSchemaDrift(pluginName, fullName string, expected, actual []string) []DriftEntry {
	expectedSet := make(map[string]bool, len(expected))
	for _, col := range expected {
		expectedSet[col] = true
	}
	actualSet := make(map[string]bool, len(actual))
	for _, col := range actual {
		actualSet[col] = true
	}

	var drifts []DriftEntry

	// Check for columns expected but missing from actual.
	for _, col := range expected {
		if !actualSet[col] {
			drifts = append(drifts, DriftEntry{
				Table:  fullName,
				Kind:   "missing",
				Column: col,
			})
		}
	}

	// Check for columns in actual that are not expected.
	for _, col := range actual {
		if !expectedSet[col] {
			drifts = append(drifts, DriftEntry{
				Table:  fullName,
				Kind:   "extra",
				Column: col,
			})
		}
	}

	if len(drifts) > 0 {
		// Build column lists for log message.
		var missingCols, extraCols []string
		for _, d := range drifts {
			if d.Kind == "missing" {
				missingCols = append(missingCols, d.Column)
			} else {
				extraCols = append(extraCols, d.Column)
			}
		}
		utility.DefaultLogger.Warn(
			fmt.Sprintf("plugin %q table %q has schema drift -- add migration logic in on_init() or recreate the table",
				pluginName, fullName),
			nil,
			"missing_columns", strings.Join(missingCols, ", "),
			"extra_columns", strings.Join(extraCols, ", "),
		)
	}

	return drifts
}
