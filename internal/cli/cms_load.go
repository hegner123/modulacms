package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)

func LoadedShallowTreeCmd(t *TreeRoot) tea.Cmd {
	return func() tea.Msg {
		return LoadedShallowTreeMsg{
			TreeRoot: t,
		}
	}

}

type LoadedShallowTreeMsg struct {
	TreeRoot *TreeRoot
}

func (m Model) LoadShallowTree() tea.Cmd {
	root := TreeRoot{}
	return LoadedShallowTreeCmd(&root)
}

func (m Model) BuildTree(rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	return tea.Batch()
}
