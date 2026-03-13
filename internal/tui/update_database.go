package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/db"
)

// FetchSource identifies the context in which a database fetch was initiated.
type FetchSource string

// FetchSource constants indicate the originating context of a database operation.
const (
	DATATYPEMENU    FetchSource = "datatype_menu"
	BUILDTREE       FetchSource = "build_tree"
	PICKCONTENTDATA FetchSource = "fetch_source"
)

// DatabaseUpdate signals a legacy database operation, now being phased out.
type DatabaseUpdate struct{}

// NewDatabaseUpdate returns a command that creates a DatabaseUpdate message.
func NewDatabaseUpdate() tea.Cmd {
	return func() tea.Msg {
		return DatabaseUpdate{}
	}
}

// UpdateDatabase handles legacy database operations using generic query builders.
func (m Model) UpdateDatabase(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DatabaseGetMsg:
		return m, tea.Batch(
			LoadingStartCmd(),
			m.DatabaseGet(m.Config, msg.Source, msg.Table, msg.ID),
		)
	case DatabaseListMsg:
		return m, tea.Batch(
			LoadingStartCmd(),
			m.DatabaseList(m.Config, msg.Source, msg.Table),
		)
	case DatabaseListFilteredMsg:
		return m, tea.Batch(
			LoadingStartCmd(),
			m.DatabaseFilteredList(m.Config, msg.Source, msg.Table, msg.Columns, msg.WhereColumn, msg.Value),
		)
	case DatabaseTreeMsg:
		return m, tea.Batch(
			LoadingStartCmd(),
			m.GetFullTree(m.Config, m.PageRouteId),
		)

	case DatabaseListRowsMsg:
		res := db.CastToTypedSlice(msg.Rows, msg.Table)
		switch msg.Table {
		case db.Datatype:
			data, _ := res.([]db.Datatypes)
			return m, DatatypesFetchResultCmd(data)
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintln(res)),
		)
	case DatabaseDeleteEntry:
		if msg.Column == "" || msg.Value == "" {
			return m, LogMessageCmd("Delete failed: no primary key selected")
		}
		d := db.ConfigDB(*m.Config)
		con, _, err := d.GetConnection()
		if err != nil {
			return m, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Delete failed: %s", err.Error())),
			)
		}
		dialect := db.DialectFromString(string(m.Config.Db_Driver))
		return m, tea.Batch(
			LoadingStartCmd(),
			func() tea.Msg {
				res, err := db.QDelete(context.Background(), con, dialect, db.DeleteParams{
					Table: msg.Table,
					Where: map[string]any{msg.Column: msg.Value},
				})
				if err != nil {
					return LogMsg{Message: fmt.Sprintf("Delete failed: %s", err.Error())}
				}
				return DbResMsg{Result: res, Table: msg.Table}
			},
		)
	case DatabaseInsertEntry:
		return m, tea.Batch(
			LoadingStartCmd(),
			m.DatabaseInsert(m.Config, msg.Table, msg.Columns, msg.Values),
			LogMessageCmd(fmt.Sprintf("Database create initiated: table %s with %d fields", msg.Table, len(msg.Columns))),
		)
	case DatabaseUpdateEntry:
		// Convert string values to any map for the query builder
		valuesMap := make(map[string]any, len(msg.Values))
		for k, v := range msg.Values {
			valuesMap[k] = v
		}
		// Parse rowID as int64 for the update WHERE clause
		var rowID int64
		if _, err := fmt.Sscanf(msg.RowID, "%d", &rowID); err != nil {
			return m, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Invalid row ID %q: %s", msg.RowID, err.Error())),
			)
		}
		return m, tea.Batch(
			LoadingStartCmd(),
			m.DatabaseUpdate(m.Config, msg.Table, rowID, valuesMap),
			LogMessageCmd(fmt.Sprintf("Database update initiated: table %s row %s", msg.Table, msg.RowID)),
		)
	case DbResMsg:
		// Operation completed — re-fetch the table rows so the screen refreshes.
		table := msg.Table
		if table == "" {
			table = m.TableState.Table
		}
		if table == "" {
			return m, LoadingStopCmd()
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			FetchTableHeadersRowsCmd(*m.Config, table, nil),
		)
	default:
		return m, nil
	}
}
