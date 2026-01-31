package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the TUI model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.Focus = (m.Focus + 1) % 3
		case "shift+tab":
			m.Focus = (m.Focus + 2) % 3
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}
