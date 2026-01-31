package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI.
func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	header := renderHeader(m.Width)
	statusBar := renderStatusBar(m)

	headerH := lipgloss.Height(header)
	statusH := lipgloss.Height(statusBar)

	// Body height is whatever remains after header and status bar
	bodyH := m.Height - headerH - statusH
	if bodyH < 3 {
		bodyH = 3
	}

	// Column widths: 25% / 50% / 25%
	leftW := m.Width / 4
	centerW := m.Width / 2
	rightW := m.Width - leftW - centerW

	treePanel := Panel{
		Title:   "Tree",
		Width:   leftW,
		Height:  bodyH,
		Content: "Content Tree\n\n  (no data loaded)",
		Focused: m.Focus == TreePanel,
	}

	contentPanel := Panel{
		Title:   "Content",
		Width:   centerW,
		Height:  bodyH,
		Content: "Fields\n\n  Select a node",
		Focused: m.Focus == ContentPanel,
	}

	routePanel := Panel{
		Title:   "Route",
		Width:   rightW,
		Height:  bodyH,
		Content: "Route\n\n  (none)",
		Focused: m.Focus == RoutePanel,
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		treePanel.Render(),
		contentPanel.Render(),
		routePanel.Render(),
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)
}
