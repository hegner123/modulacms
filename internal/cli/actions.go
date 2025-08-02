package cli

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// TODO Add default case for generic operations
func (m *Model) DatabaseInsert(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
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
		query, args, err := sqb.SecureBuildInsertQuery(string(table), valuesMap)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}
		_, err = sqb.SecureExecuteModifyQuery(query, args)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}

		// Reset the form values after creation
		m.FormValues = nil

		utility.DefaultLogger.Finfo("CLI Create successful", nil)
		return DbErrMsg{Error: nil}
	}
}

func (m *Model) DatabaseUpdate(c *config.Config, table db.DBTable) tea.Cmd {
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
		_, err = sqb.SecureExecuteModifyQuery(query, args)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}

		// Reset the form values after update
		m.FormValues = nil

		utility.DefaultLogger.Finfo("CLI Update successful", nil)
		return DbErrMsg{Error: nil}
	}
}

func (m *Model) DatabaseRead(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*c)

		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ReadMsg{Error: err, Result: nil, RType: nil}
		}

		// Using secure query builder
		sqb := db.NewSecureQueryBuilder(con)
		query, args, err := sqb.SecureBuildListQuery(string(table))
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ReadMsg{Error: err, Result: nil, RType: nil}
		}
		r, err := sqb.SecureExecuteSelectQuery(query, args)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return ReadMsg{Error: err, Result: nil, RType: nil}
		}
		defer r.Close()

		out := make([]db.Datatypes, 0)

		for r.Next() {
			res := &db.Datatypes{}
			err := r.Scan(
				&res.DatatypeID,
				&res.ParentID,
				&res.Label,
				&res.Type,
				&res.AuthorID,
				&res.DateCreated,
				&res.DateModified,
				&res.History,
			)
			if err != nil {
				return DataFetchErrorMsg{Error: err}
			}
			out = append(out, *res)
			utility.DefaultLogger.Finfo("CLI Read successful", out)

		}
		return DatatypesFetchedMsg{data: out}
	}

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
		_, err = sqb.SecureExecuteModifyQuery(query, args)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return DbErrMsg{Error: err}
		}

		return DbErrMsg{Error: nil}
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
