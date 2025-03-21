package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

func CliRun(m *model) (*tea.Program,bool) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		ErrLog.Fatal("", err)
	}
	return nil,CliContinue
}

func (m *model) Init() tea.Cmd {
    m.tables = GetTables("")
	return nil
}

