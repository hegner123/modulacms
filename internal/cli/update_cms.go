package cli

import (

	tea "github.com/charmbracelet/bubbletea"
)

type CmsUpdate struct{}

func NewCmsUpdate() tea.Cmd {
	return func() tea.Msg {
		return CmsUpdate{}
	}
}

func (m Model) UpdateCms(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GetFullTreeResMsg:
		r := m.BuildTree(msg.Rows)
		return m, tea.Batch(
			r,
		)
	case BuildTreeFromRouteMsg:
		return m, tea.Batch()
	default:
		return m, nil
	}

}
