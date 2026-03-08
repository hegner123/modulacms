package plugin

import (
	"fmt"
	"math"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	lua "github.com/yuin/gopher-lua"
)

// luaInsertMany implements db.insert_many(table, columns, rows).
// Auto-injects id, created_at, updated_at per row.
// Operation budget: counts as ceil(len(rows) / batchSize) operations.
// Returns nothing on success, nil+errmsg on error.
func (api *DatabaseAPI) luaInsertMany(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	colsTbl := L.CheckTable(2)
	rowsTbl := L.CheckTable(3)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse columns from Lua sequence.
	var userCols []string
	colsTbl.ForEach(func(key, value lua.LValue) {
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		if s, ok := value.(lua.LString); ok {
			userCols = append(userCols, string(s))
		}
	})
	if len(userCols) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("insert_many: columns cannot be empty"))
		return 2
	}

	// Parse rows from Lua sequence of sequences.
	var userRows [][]any
	var parseErr error
	rowsTbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		rowTbl, ok := value.(*lua.LTable)
		if !ok {
			parseErr = fmt.Errorf("insert_many: each row must be a table")
			return
		}
		var vals []any
		rowTbl.ForEach(func(k, v lua.LValue) {
			if _, ok := k.(lua.LNumber); !ok {
				return
			}
			vals = append(vals, LuaValueToGo(v))
		})
		if len(vals) != len(userCols) {
			parseErr = fmt.Errorf("insert_many: row has %d values, expected %d (matching columns)", len(vals), len(userCols))
			return
		}
		userRows = append(userRows, vals)
	})
	if parseErr != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(parseErr.Error()))
		return 2
	}
	if len(userRows) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("insert_many: rows cannot be empty"))
		return 2
	}
	if len(userRows) > db.MaxBulkInsertRows {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("insert_many: too many rows: %d (max %d)", len(userRows), db.MaxBulkInsertRows)))
		return 2
	}

	// Auto-prepend id, created_at, updated_at to columns.
	columns := append([]string{"id", "created_at", "updated_at"}, userCols...)

	// Auto-prepend auto-generated values to each row.
	now := time.Now().UTC().Format(time.RFC3339)
	rows := make([][]any, len(userRows))
	for i, userRow := range userRows {
		row := make([]any, 0, len(columns))
		row = append(row, types.NewULID().String(), now, now)
		row = append(row, userRow...)
		rows[i] = row
	}

	// Operation budget: charge ceil(rows / batchSize) - 1 additional ops
	// (the initial checkOpLimit already counted 1).
	batchSize := bulkBatchSizeForDialect(api.dialect, len(columns))
	totalBatches := int(math.Ceil(float64(len(rows)) / float64(batchSize)))
	if totalBatches > 1 {
		additionalOps := totalBatches - 1
		if api.opCount+additionalOps > api.maxOpsPerExec {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf(
				"insert_many: would consume %d operations (%d rows / %d batch size), exceeding budget (%d remaining of %d)",
				totalBatches, len(rows), batchSize, api.maxOpsPerExec-api.opCount, api.maxOpsPerExec,
			)))
			return 2
		}
		api.opCount += additionalOps
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.insert_many: no context set on Lua state (VM setup error)")
		return 0
	}

	_, err = db.QBulkInsert(ctx, api.currentExec, api.dialect, db.BulkInsertParams{
		Table:   prefixed,
		Columns: columns,
		Rows:    rows,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaCreateIndex implements db.create_index(table, opts).
// opts: {columns = {"col1", "col2"}, unique = true/false}
// Returns nothing on success, nil+errmsg on error.
func (api *DatabaseAPI) luaCreateIndex(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	optsTbl := L.CheckTable(2)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse columns.
	colsVal := L.GetField(optsTbl, "columns")
	colsTbl, ok := colsVal.(*lua.LTable)
	if !ok || colsVal == lua.LNil {
		L.Push(lua.LNil)
		L.Push(lua.LString("create_index: 'columns' table is required"))
		return 2
	}

	var columns []string
	colsTbl.ForEach(func(key, value lua.LValue) {
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		if s, ok := value.(lua.LString); ok {
			columns = append(columns, string(s))
		}
	})
	if len(columns) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("create_index: columns cannot be empty"))
		return 2
	}

	unique := parseBoolField(L, optsTbl, "unique")

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.create_index: no context set on Lua state (VM setup error)")
		return 0
	}

	err = db.DDLCreateIndex(ctx, api.currentExec, api.dialect, db.DDLCreateIndexParams{
		Table:       prefixed,
		Columns:     columns,
		Unique:      unique,
		IfNotExists: true,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// bulkBatchSizeForDialect mirrors db.bulkBatchSize but is accessible from the plugin package.
// This avoids exporting the internal batch size function from the db package.
func bulkBatchSizeForDialect(d db.Dialect, numCols int) int {
	if numCols <= 0 {
		return 1
	}
	switch d {
	case db.DialectSQLite:
		perCol := 999 / numCols
		if perCol > 100 {
			return 100
		}
		if perCol < 1 {
			return 1
		}
		return perCol
	case db.DialectMySQL:
		return 500
	case db.DialectPostgres:
		return 1000
	default:
		return 100
	}
}
