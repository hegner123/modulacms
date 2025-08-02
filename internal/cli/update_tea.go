package cli

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TeaUpdate struct{}

func NewTea() tea.Cmd {
	return func() tea.Msg {
		return TeaUpdate{}
	}
}

func (m Model) UpdateTea(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
		headerHeight := lipgloss.Height(m.headerView() + RenderTitle(m.Titles[m.TitleFont]) + RenderHeading(m.Header))
		footerHeight := lipgloss.Height(m.footerView() + RenderFooter(m.Footer))
		verticalMarginHeight := headerHeight + footerHeight

		if !m.Ready {
			m.Viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight)
			m.Viewport.YPosition = headerHeight
			m.Ready = true
		} else {
			m.Viewport.YPosition = headerHeight
			m.Viewport.Width = msg.Width - 4
			m.Viewport.Height = msg.Height - verticalMarginHeight - 10
		}
    return m, NewTea()
	default:
		return m, nil
	}
}
