package cli

import (
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

type FetchTables struct {
	Tables []string
}

type headersRowsFetchedMsg struct {
	Headers []string
	Rows    [][]string
}

func GetTablesCMD(c *config.Config) tea.Cmd {
	return func() tea.Msg {
		var (
			d      db.DbDriver
			labels []string
		)
		d = db.ConfigDB(*c)
		tables, err := d.ListTables()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ErrMsg{Error: err}
		}

		for _, table := range *tables {
			labels = append(labels, table.Label)
		}
		return TablesSet{Tables: labels}
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
		defer rows.Close()
		clm, err := rows.Columns()
		if err != nil {
			return ErrMsg{Error: err}
		}
		ct, err := rows.ColumnTypes()
		if err != nil {
			return ErrMsg{Error: err}
		}
		return ColumnsFetched{Columns: &clm, ColumnTypes: &ct}
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
		defer rows.Close()
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
