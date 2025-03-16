package cli

import tea "github.com/charmbracelet/bubbletea"

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
