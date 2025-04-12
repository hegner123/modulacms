package cli

import tea "github.com/charmbracelet/bubbletea"

type updateMaxCursorMsg struct {
	cursorMax int
}

func (m Model) UpdateMaxCursor() tea.Cmd {
	return func() tea.Msg {
		start, end := m.paginator.GetSliceBounds(len(m.rows))
		currentView := m.rows[start:end]
		return updateMaxCursorMsg{cursorMax: len(currentView)}
	}
}
