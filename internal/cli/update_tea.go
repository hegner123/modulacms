package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TeaUpdate signals a Tea framework update, triggered by window size changes.
type TeaUpdate struct{}

// NewTea returns a command that creates a TeaUpdate message.
func NewTea() tea.Cmd {
	return func() tea.Msg {
		return TeaUpdate{}
	}
}

// UpdateTea handles window resize messages from Tea framework.
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

// titleStyle and infoStyle are Lipgloss styles for rendering the header and footer.
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

// headerView renders the viewport header with title and horizontal line.
func (m Model) headerView() string {
	var titleText string

	if m.Page.Index == CONFIGPAGE {
		titleText = "Configuration"
	}
	title := titleStyle.Render(titleText)
	line := strings.Repeat("─", max(0, m.Viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

// footerView renders the viewport footer showing scroll percentage.
func (m Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.Viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.Viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
