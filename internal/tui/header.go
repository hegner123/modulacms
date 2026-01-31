package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// renderHeader renders the top action bar with app title and action buttons.
func renderHeader(width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent).
		PaddingRight(2)

	buttonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(config.DefaultStyle.Secondary).
		Background(config.DefaultStyle.SecondaryBG)

	actions := []string{"New", "Save", "Copy", "Duplicate", "Export"}
	buttons := make([]string, len(actions))
	for i, a := range actions {
		buttons[i] = buttonStyle.Render(a)
	}

	title := titleStyle.Render("ModulaCMS")
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)

	row := lipgloss.JoinHorizontal(lipgloss.Center, title, buttonBar)

	container := lipgloss.NewStyle().
		Width(width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.Tertiary).
		Padding(0, 1).
		Render(row)

	return container
}
