package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/utility"
)

// CliRun runs the CLI TUI with the given model and returns a program pointer and the continue flag.
func CliRun(m *Model) (*tea.Program, bool) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		utility.DefaultLogger.Ffatal("", err)
	}
	return nil, CliContinue
}

func (m Model) Init() tea.Cmd {
    // Only return spinner tick if we're in a loading state
    if m.Loading {
        return m.Spinner.Tick
    }
    return nil
}
