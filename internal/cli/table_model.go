package cli

import "database/sql"

// TableModel encapsulates all table-related state extracted from Model.
// This groups database table viewing and navigation fields together,
// following the FormModel pattern established in Phase 1.
type TableModel struct {
	Table       string                 // Current table name
	Headers     []string               // Table column headers for display
	Rows        [][]string             // Table data rows
	Columns     *[]string              // Column names from database
	ColumnTypes *[]*sql.ColumnType     // Column type metadata from database
	Selected    map[int]struct{}       // Selected rows (for multi-select operations)
	Row         *[]string              // Currently selected single row data
}

// NewTableModel creates a new TableModel with safe defaults.
// Headers, Rows are initialized as empty slices.
// Selected is initialized as an empty map.
// Pointer fields (Columns, ColumnTypes, Row) are nil by default.
func NewTableModel() *TableModel {
	return &TableModel{
		Table:    "",
		Headers:  make([]string, 0),
		Rows:     make([][]string, 0),
		Selected: make(map[int]struct{}),
	}
}
