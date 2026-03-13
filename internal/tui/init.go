package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/utility"
)

// CliRun runs the CLI TUI with the given model and returns a program pointer and the continue flag.
func CliRun(m *Model) (*tea.Program, bool) {
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		utility.DefaultLogger.Ffatal("", err)
	}
	return nil, CliContinue
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		GetTablesCMD(m.Config),
	}
	if m.Page.Index == HOMEPAGE {
		cmds = append(cmds, HomeDashboardFetchCmd(m.DB))
	}
	if m.Loading {
		cmds = append(cmds, m.Spinner.Tick)
	}
	return tea.Batch(cmds...)
}
