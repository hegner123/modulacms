package tui

import "fmt"

// Required validates that the input string is not empty.
func Required(s string) error {
	if len(s) < 1 {
		return fmt.Errorf("\nInput Cannot Be Null")
	} else {
		return nil
	}
}

// InitTables initializes a map of table names from a slice of strings.
func InitTables(tables []string) map[string]string {
	out := make(map[string]string, 0)
	for _, v := range tables {
		out[v] = v
	}
	return out
}

// GetCurrentRowPK returns the primary key column name and value for the current
// row in the table state. Works with both integer and string (ULID) primary keys.
func (m Model) GetCurrentRowPK() (column string, value string) {
	if m.TableState == nil {
		return "", ""
	}
	rows := m.TableState.Rows
	if len(rows) == 0 || m.Cursor >= len(rows) {
		return "", ""
	}
	row := rows[m.Cursor]
	if len(row) == 0 {
		return "", ""
	}
	// First header is the PK column name
	if len(m.TableState.Headers) > 0 {
		column = m.TableState.Headers[0]
	}
	return column, row[0]
}
