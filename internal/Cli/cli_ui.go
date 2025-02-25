package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

func (m model) Init() tea.Cmd {
	return m.LaunchCms()
}

func (m model) LaunchCms() tea.Cmd {
	m.tables = GetTables("")
	return func() tea.Msg {
		m.Update("Launch")
		return m.View()

	}

}

func (m model) RenderUI() string {
	m.footer += fmt.Sprintf("%v\nPress q to quit.\n%v", utility.REDF, utility.RESET)
	return m.header + m.body + m.footer
}

func (m model) SelectTableUI(action string) string {
	m.header += fmt.Sprintf("\nSelect table to %s\n\n", action)

	for i, choice := range m.tables {

		cursor := " " // no cursor
		if m.cursor == i {
			cursor = "->" // cursor!
		}

		m.body += fmt.Sprintf("%s  %s\n", cursor, choice)
	}

	return m.RenderUI()
}
