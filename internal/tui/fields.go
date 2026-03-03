package tui

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// GetSuggestionsString retrieves column value suggestions from the database.
// Returns nil when the connection is unavailable (e.g., remote mode).
func (m Model) GetSuggestionsString(c *config.Config, column string) []string {
	if column == "NIll" {
		return nil
	}
	d := db.ConfigDB(*c)
	con, ctx, err := d.GetConnection()
	if err != nil {
		return nil
	}
	r, err := db.GetColumnRowsString(con, ctx, m.TableState.Table, column)
	if err != nil {
		m.Logger.Error("ERROR", err)
		return nil
	}
	return r
}
