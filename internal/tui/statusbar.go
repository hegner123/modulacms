package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// renderStatusBar renders the bottom status bar with focus indicator and key hints.
func renderStatusBar(m Model) string {
	barStyle := lipgloss.NewStyle().
		Background(config.DefaultStyle.Status2BG).
		Foreground(config.DefaultStyle.Status1)

	focusLabel := barStyle.
		Bold(true).
		Padding(0, 1).
		Render("[" + m.Focus.String() + "]")

	hints := barStyle.
		Padding(0, 1).
		Render("tab: switch panel  q: quit")

	// Calculate gap to push hints to the right
	focusWidth := lipgloss.Width(focusLabel)
	hintsWidth := lipgloss.Width(hints)
	gap := m.Width - focusWidth - hintsWidth
	if gap < 0 {
		gap = 0
	}

	spacer := barStyle.Render(strings.Repeat(" ", gap))

	return focusLabel + spacer + hints
}
