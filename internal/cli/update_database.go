package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)

type FetchSource string

const (
	DATATYPEMENU    FetchSource = "datatype_menu"
	BUILDTREE       FetchSource = "build_tree"
	PICKCONTENTDATA FetchSource = "fetch_source"
)

type DatabaseUpdate struct{}

func NewDatabaseUpdate() tea.Cmd {
	return func() tea.Msg {
		return DatabaseUpdate{}
	}
}

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
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Database delete requested: ID %d from table %s", msg.Id, msg.Table)),
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
		// Parse rowID as int64 for SecureBuildUpdateQuery
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
	default:
		return m, nil
	}
}
