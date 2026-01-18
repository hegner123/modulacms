package cli

import (
	"strconv"

	"github.com/hegner123/modulacms/internal/utility"
)

func (m Model) GetCurrentRowId() int64 {
	rows := m.TableState.Rows
	if len(rows) == 0 {
		utility.DefaultLogger.Ferror("No rows available", nil)
		return 0
	}
	if m.Cursor >= len(rows) {
		utility.DefaultLogger.Ferror("Cursor out of range", nil)
		return 0
	}
	row := rows[m.Cursor]
	if len(row) == 0 {
		utility.DefaultLogger.Ferror("Row has no columns", nil)
		return 0
	}
	rowCol := row[0]
	utility.DefaultLogger.Finfo("rowCOl", rowCol)
	id, err := strconv.ParseInt(rowCol, 10, 64)
	if err != nil {
		utility.DefaultLogger.Ferror("", err)
		return 0
	}
	return id
}
