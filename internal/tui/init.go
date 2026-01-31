package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/utility"
)

// Init returns nil; no initial async work needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Run creates and runs a tea.Program with alt screen.
// Returns the program and whether the TUI exited normally.
func Run(m *Model) (*tea.Program, bool) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		utility.DefaultLogger.Ffatal("", err)
	}
	return nil, false
}
