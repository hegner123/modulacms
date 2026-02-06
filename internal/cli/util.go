package cli

import (
	"fmt"
	"strconv"
)

// Validation functions

func Required(s string) error {
	if len(s) < 1 {
		return fmt.Errorf("\nInput Cannot Be Null")
	} else {
		return nil
	}
}

// Table initialization

func InitTables(tables []string) map[string]string {
	out := make(map[string]string, 0)
	for _, v := range tables {
		out[v] = v
	}
	return out
}

// Row handling

func (m Model) GetCurrentRowId() int64 {
	rows := m.TableState.Rows
	if len(rows) == 0 {
		m.Logger.Ferror("No rows available", nil)
		return 0
	}
	if m.Cursor >= len(rows) {
		m.Logger.Ferror("Cursor out of range", nil)
		return 0
	}
	row := rows[m.Cursor]
	if len(row) == 0 {
		m.Logger.Ferror("Row has no columns", nil)
		return 0
	}
	rowCol := row[0]
	m.Logger.Finfo("rowCOl", rowCol)
	id, err := strconv.ParseInt(rowCol, 10, 64)
	if err != nil {
		m.Logger.Ferror("", err)
		return 0
	}
	return id
}
