package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/utility"
)

// Init returns nil; no initial async work needed.
func (m PanelModel) Init() tea.Cmd {
	return nil
}

// PanelRun creates and runs a tea.Program with alt screen.
// Returns the program and whether the TUI exited normally.
func PanelRun(m *PanelModel) (*tea.Program, bool) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		utility.DefaultLogger.Ffatal("", err)
	}
	return nil, false
}
