package cli

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)


type dbErrMsg struct {
	Error error
}

// TODO Add default case for generic operations
func (m *Model) CLICreate(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*c)
		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}
		valuesMap := make(map[string]string, 0)
		for i, v := range m.FormValues {
			valuesMap[m.Headers[i]] = *v
		}
		
		// Using generic query method since direct CRUD methods aren't available
		query := db.BuildInsertQuery(string(table), valuesMap)
		_, err = d.Query(con, query)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}
		
		// Reset the form values after creation
		m.FormValues = nil
		
		utility.DefaultLogger.Finfo("CLI Create successful", nil)
		return dbErrMsg{Error: nil}
	}
}

func (m *Model) CLIUpdate(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		id := m.GetIDRow()
		d := db.ConfigDB(*c)
		
		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}
		
		valuesMap := make(map[string]string, 0)
		for i, v := range m.FormValues {
			valuesMap[m.Headers[i]] = *v
		}
		
		// Using generic query method since direct CRUD methods aren't available
		query := db.BuildUpdateQuery(string(table), id, valuesMap)
		_, err = d.Query(con, query)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}
		
		// Reset the form values after update
		m.FormValues = nil
		
		utility.DefaultLogger.Finfo("CLI Update successful", nil)
		return dbErrMsg{Error: nil}
	}
}

func (m *Model) CLIRead(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		id := m.GetIDRow()
		d := db.ConfigDB(*c)

		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}

		// Using generic query method since direct CRUD methods aren't available
		query := db.BuildSelectQuery(string(table), id)
		_, err = d.Query(con, query)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}

		utility.DefaultLogger.Finfo("CLI Read successful", nil)
		return dbErrMsg{Error: nil}
	}
}

func (m *Model) CLIDelete(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		id := m.GetIDRow()
		d := db.ConfigDB(*c)

		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}

		// Using generic query method since direct CRUD methods aren't available
		query := db.BuildDeleteQuery(string(table), id)
		_, err = d.Query(con, query)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}

		return dbErrMsg{Error: nil}
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