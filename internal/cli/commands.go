package cli

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type DatabaseCMD string

const (
	INSERT DatabaseCMD = "insert"
	SELECT DatabaseCMD = "select"
	UPDATE DatabaseCMD = "update"
	DELETE DatabaseCMD = "delete"
	BATCH  DatabaseCMD = "batch"
)

type FetchErrMsg struct {
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
			return FetchErrMsg{Error: err}
		}

		for _, table := range *tables {
			labels = append(labels, table.Label)
		}
		return TablesSet{Tables: labels}
	}
}

// TODO Add default case for generic operations
func (m Model) DatabaseInsert(c *config.Config, table db.DBTable, columns []string, values []*string) tea.Cmd {
	d := db.ConfigDB(*c)
	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	valuesMap := make(map[string]any, 0)
	for i, v := range values {
		if i == 0 || v == nil {
			continue
		} else {
			valuesMap[columns[i]] = *v
		}
	}
	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildInsertQuery(string(table), valuesMap)
	if err != nil {
		return tea.Batch(
			LogMessageCmd(err.Error()),
			LogMessageCmd(fmt.Sprintln(valuesMap)),
		)
	}
	res, err := sqb.SecureExecuteModifyQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}


	return tea.Batch(
		DbResultCmd(res, string(table)),
	)
}

func (m Model) DatabaseUpdate(c *config.Config, table db.DBTable) tea.Cmd {
	id := m.GetCurrentRowId()
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	valuesMap := make(map[string]any, 0)
	for i, v := range m.FormState.FormValues {
		valuesMap[m.Headers[i]] = *v
	}

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildUpdateQuery(string(table), id, valuesMap)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	res, err := sqb.SecureExecuteModifyQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	// Reset the form values after update
	m.FormState.FormValues = nil

	utility.DefaultLogger.Finfo("CLI Update successful", nil)
	return DbResultCmd(res, string(table))
}
func (m Model) DatabaseGet(c *config.Config, source FetchSource, table db.DBTable, id int64) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildSelectQuery(string(table), id)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	r, err := sqb.SecureExecuteSelectQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	defer utility.HandleRowsCloseDeferErr(r)
	out, err := Parse(r, table)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DatabaseGetRowsCmd(source, out, table)

}

func (m Model) DatabaseList(c *config.Config, source FetchSource, table db.DBTable) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildListQuery(string(table))
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	r, err := sqb.SecureExecuteSelectQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	defer utility.HandleRowsCloseDeferErr(r)
	out, err := Parse(r, table)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DatabaseListRowsCmd(source, out, table)

}

func (m Model) DatabaseFilteredList(c *config.Config, source FetchSource, table db.DBTable, columns []string, whereColumn string, value any) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildSelectWithColumnsQuery(string(table), columns, whereColumn, value)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	r, err := sqb.SecureExecuteSelectQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	defer utility.HandleRowsCloseDeferErr(r)
	out, err := Parse(r, table)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DatabaseListRowsCmd(source, out, table)
}

func (m Model) DatabaseDelete(c *config.Config, table db.DBTable) tea.Cmd {
	id := m.GetCurrentRowId()
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildDeleteQuery(string(table), id)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	res, err := sqb.SecureExecuteModifyQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DbResultCmd(res, string(table))
}

func (m Model) GetContentField(node *string) []byte {
	row := m.Rows[m.Cursor]
	j, err := json.Marshal(row)
	if err != nil {
		utility.DefaultLogger.Ferror("", err)
	}
	return j
}

func (m Model) GetFullTree(c *config.Config, id int64) tea.Cmd {
	// TODO: Implement tree retrieval logic
	d := db.ConfigDB(*c)
	res, err := d.GetRouteTreeByRouteID(1)
	if err != nil {
		return ErrorSetCmd(err)
	}
	out := db.LogRouteTree("GetFullTree", res)
	return GetFullTreeResCMD(out, *res)
}

func (m Model) GetContentInstances(c *config.Config) tea.Cmd {
	//	d := db.ConfigDB(*c)
	//TODO JOIN STATEMENTS FOR CONTENT DATA AND DATATYPES
	//TODO JOIN STATEMENTS FOR CONTENT FIELDS AND FIELDS

	return tea.Batch()
}
