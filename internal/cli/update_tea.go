package cli

import (
	"fmt"
	"strings"

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
		headerHeight := lipgloss.Height(m.headerView() + RenderTitle(m.Titles[m.TitleFont]))
		footerHeight := lipgloss.Height(m.footerView())
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

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

func (m Model) headerView() string {
	var titleText string

	if m.Page.Index == CONFIGPAGE {
		titleText = "Configuration"
	}
	title := titleStyle.Render(titleText)
	line := strings.Repeat("─", max(0, m.Viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.Viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.Viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
