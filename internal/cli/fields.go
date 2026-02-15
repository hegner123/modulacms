package cli

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// GetSuggestionsString retrieves column value suggestions from the database.
func (m Model) GetSuggestionsString(c *config.Config, column string) []string {
	d := db.ConfigDB(*c)
	con, ctx, _ := d.GetConnection()
	if column == "NIll" {
		return nil
	} else {
		r, err := db.GetColumnRowsString(con, ctx, m.TableState.Table, column)
		if err != nil {
			m.Logger.Error("ERROR", err)
			return nil
		}
		return r

	}
}
