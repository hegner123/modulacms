package cli

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type ErrMsg struct {
	Error error
}

type ForeignKeyReference struct {
	From   string
	Table  string // Referenced table name.
	Column string // Referenced column name.
}

type tableFetchedMsg struct {
	Tables []string
}

type columnFetchedMsg struct {
	Columns     *[]string
	ColumnTypes *[]*sql.ColumnType
}

type headersRowsFetchedMsg struct {
	Headers []string
	Rows    [][]string
}

// ForeignKeyReference holds the referenced table and column information.

func GetTablesCMD(c *config.Config) tea.Cmd {
	return func() tea.Msg {
		var (
			d      db.DbDriver
			labels []string
		)
		d = db.ConfigDB(*c)
		con, ctx, _ := d.GetConnection()
		q := "SELECT * FROM tables;"
		rows, err := con.QueryContext(ctx, q)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ErrMsg{Error: err}
		}

		for rows.Next() {
			var (
				id        int
				label     string
				author_id int
			)
			err = rows.Scan(&id, &label, &author_id)
			if err != nil {
				return ErrMsg{Error: err}
			}
			labels = append(labels, label)
		}
		if err := rows.Err(); err != nil {
			return ErrMsg{Error: err}
		}
		return tableFetchedMsg{Tables: labels}
	}
}

func GetColumns(c *config.Config, t string) tea.Cmd {
	return func() tea.Msg {
		dbt := db.StringDBTable(t)
		query := "SELECT * FROM"
		d := db.ConfigDB(*c)
		rows, err := d.ExecuteQuery(query, dbt)
		if err != nil {
			return ErrMsg{Error: err}
		}
		clm, err := rows.Columns()
		if err != nil {
			return ErrMsg{Error: err}
		}
		ct, err := rows.ColumnTypes()
		if err != nil {
			return ErrMsg{Error: err}

		}
		return columnFetchedMsg{Columns: &clm, ColumnTypes: &ct}
	}
}

func FetchHeadersRows(c *config.Config, t string) tea.Cmd {
	return func() tea.Msg {
		utility.DefaultLogger.Finfo("call to get columns rows")
		dbt := db.StringDBTable(t)
		query := "SELECT * FROM"
		d := db.ConfigDB(*c)
		rows, err := d.ExecuteQuery(query, dbt)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ErrMsg{Error: err}
		}
		columns, err := rows.Columns()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ErrMsg{Error: err}
		}
		listRows, err := db.GenericList(dbt, d)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ErrMsg{Error: err}
		}
		return headersRowsFetchedMsg{Headers: columns, Rows: listRows}
	}

}
