package tui

import (
	"database/sql"

	"github.com/hegner123/modulacms/internal/db"
)

// DatabaseGetMsg requests fetching a single database record by ID.
type DatabaseGetMsg struct {
	Source FetchSource
	Table  db.DBTable
	ID     int64
}

// DatabaseListMsg requests listing all records from a table.
type DatabaseListMsg struct {
	Source FetchSource
	Table  db.DBTable
}

// DatabaseListFilteredMsg requests listing database records with a WHERE filter.
type DatabaseListFilteredMsg struct {
	Source      FetchSource
	Table       db.DBTable
	Columns     []string
	WhereColumn string
	Value       any
}

// DatabaseGetRowMsg returns a single database row result.
type DatabaseGetRowMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}

// DatabaseListFilteredRowsMsg returns filtered database rows.
type DatabaseListFilteredRowsMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}

// DatabaseListRowsMsg returns all database rows from a table.
type DatabaseListRowsMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}

// DatabaseDeleteEntry requests deletion of a database entry.
type DatabaseDeleteEntry struct {
	Column string // PK column name (e.g. "datatype_id", "id")
	Value  string // PK value (integer or ULID string)
	Table  string
}

// DatabaseInsertEntry requests insertion of a new database entry.
type DatabaseInsertEntry struct {
	Table   db.DBTable
	Columns []string
	Values  []*string
}

// DatabaseUpdateEntry requests update of an existing database entry.
type DatabaseUpdateEntry struct {
	Table  db.DBTable
	RowID  string
	Values map[string]string
}

// DbResMsg returns a database operation result.
type DbResMsg struct {
	Result sql.Result
	Table  string
}

// DbErrMsg reports a database operation error.
type DbErrMsg struct {
	Error error
}

// ReadMsg returns database read results with optional error.
type ReadMsg struct {
	Result *sql.Rows
	Error  error
	RType  any
}

// DatabaseTreeMsg requests building the database tree view.
type DatabaseTreeMsg struct{}

// BuildTreeFromRows requests building a content tree from database rows.
type BuildTreeFromRows struct {
	Rows []db.GetRouteTreeByRouteIDRow
}

// GetFullTreeResMsg returns the full tree query result rows.
type GetFullTreeResMsg struct {
	Rows []db.GetRouteTreeByRouteIDRow
}

// TableSet sets the currently selected database table.
type TableSet struct {
	Table string
}

// SetDatabaseModeMsg sets the database operation mode on the Model.
// Emitted by DatabaseScreen before navigating to READPAGE so that
// screenForPage receives the correct mode.
type SetDatabaseModeMsg struct {
	Mode DatabaseMode
}
