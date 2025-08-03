package cli

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type DatabaseAction string

const (
	INSERT DatabaseAction = "insert"
	SELECT DatabaseAction = "select"
	UPDATE DatabaseAction = "update"
	DELETE DatabaseAction = "delete"
)

// TODO Add default case for generic operations
func (m Model) DatabaseInsert(c *config.Config, table db.DBTable, columns []string, values []*string) tea.Cmd {
	d := db.ConfigDB(*c)
	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	valuesMap := make(map[string]any, 0)
	for i, v := range values {
		if i == 0 {
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

	// Reset the form values after creation
	reset := make([]*string, 0)

	return tea.Batch(
		FormSetValuesCmd(reset),
		DbResultCmd(res, string(table)),
	)
}

func (m Model) DatabaseUpdate(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		id := m.GetCurrentRowId()
		d := db.ConfigDB(*c)

		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}

		valuesMap := make(map[string]any, 0)
		for i, v := range m.FormValues {
			valuesMap[m.Headers[i]] = *v
		}

		// Using secure query builder
		sqb := db.NewSecureQueryBuilder(con)
		query, args, err := sqb.SecureBuildUpdateQuery(string(table), id, valuesMap)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}
		res, err := sqb.SecureExecuteModifyQuery(query, args)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}

		// Reset the form values after update
		m.FormValues = nil

		utility.DefaultLogger.Finfo("CLI Update successful", nil)
		return DbResultCmd(res, string(table))
	}
}

func (m Model) DatabaseList(c *config.Config, table db.DBTable) tea.Cmd {
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
	defer r.Close()
	out, err := Parse(r, table)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DatabaseListRowsCmd(out, table)

}

func (m Model) DatabaseDelete(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		id := m.GetCurrentRowId()
		d := db.ConfigDB(*c)

		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}

		// Using secure query builder
		sqb := db.NewSecureQueryBuilder(con)
		query, args, err := sqb.SecureBuildDeleteQuery(string(table), id)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}
		res, err := sqb.SecureExecuteModifyQuery(query, args)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}

		return DbResultCmd(res, string(table))
	}
}

func (m Model) GetContentField(node *string) []byte {
	row := m.Rows[m.Cursor]
	j, err := json.Marshal(row)
	if err != nil {
		utility.DefaultLogger.Ferror("", err)
	}
	return j
}
