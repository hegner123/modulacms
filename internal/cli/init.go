package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/utility"
)

func CliRun(m *Model) (*tea.Program, bool) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		utility.DefaultLogger.Ffatal("", err)
	}
	return nil, CliContinue
}

func (m Model) Init() tea.Cmd {
    return m.spinner.Tick
}
