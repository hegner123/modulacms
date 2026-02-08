package plugin

import (
	"fmt"
	"strings"

	db "github.com/hegner123/modulacms/internal/db"
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
// The context for DDL calls comes from L.Context() so that execution timeouts apply.
func luaDefineTable(L *lua.LState, pluginName string, exec db.Executor, dialect db.Dialect) lua.LGFunction {
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
