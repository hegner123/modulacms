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
		return m, LogMessage(fmt.Sprintf("ID: %d | Table: %s\n", msg.Id, msg.Table))
	case DatabaseCreateEntry:
		return m, LogMessage(string(msg.Table))
	case DatabaseUpdateEntry:
		return m, LogMessage(string(msg.Table))
	default:
		return m, nil

	}

}
