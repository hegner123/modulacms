package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type DatabaseUpdate struct{}

func NewDatabaseUpdate() tea.Cmd {
	return func() tea.Msg {
		return DatabaseUpdate{}
	}
}

func (m Model) UpdateDatabase(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DatabaseDeleteEntry:
		return m, LogMessageCmd(fmt.Sprintf("Database delete requested: ID %d from table %s", msg.Id, msg.Table))
	case DatabaseInsertEntry:
		return m, tea.Batch(
			m.DatabaseInsert(m.Config, msg.Table, msg.Columns, msg.Values),
			LogMessageCmd(fmt.Sprintf("Database create initiated: table %s with %d fields", msg.Table, len(msg.Columns))),
		)
	case DatabaseUpdateEntry:
		return m, LogMessageCmd(fmt.Sprintf("Database update initiated: table %s with %d fields", msg.Table, len(msg.Values)))
	default:
		return m, nil

	}

}
