package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

func CliRun(m tea.Model) *tea.Program {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		ErrLog.Fatal("", err)
	}
	return p
}

func (m *model) Init() tea.Cmd {
    m.tables = GetTables("")
	return nil
}

