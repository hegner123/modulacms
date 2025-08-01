package cli

import tea "github.com/charmbracelet/bubbletea"

type updateMaxCursorMsg struct {
	cursorMax int
}

func (m Model) UpdateMaxCursor() tea.Cmd {
	return func() tea.Msg {
		start, end := m.Paginator.GetSliceBounds(len(m.Rows))
		currentView := m.Rows[start:end]
		return updateMaxCursorMsg{cursorMax: len(currentView)}
	}
}