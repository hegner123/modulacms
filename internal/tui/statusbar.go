package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// renderStatusBar renders a two-line status bar at the bottom of the
// three-panel layout. Line 1 shows the focused panel and navigation
// keys. Line 2 shows action keys.
func renderStatusBar(m PanelModel) string {
	// High-contrast: white text on black background.
	barFG := config.DefaultStyle.Primary
	barBG := config.DefaultStyle.PrimaryBG

	barStyle := lipgloss.NewStyle().
		Foreground(barFG).
		Background(barBG)

	keyStyle := barStyle.Bold(true)

	// --- helpers ---

	// key renders a single  key:label  pair.
	key := func(k, label string) string {
		return keyStyle.Render(k) + barStyle.Render(":"+label)
	}

	// padLine fills a line to exactly m.Width so the background color
	// spans the full terminal width.
	padLine := func(content string) string {
		w := lipgloss.Width(content)
		if w >= m.Width {
			return content
		}
		return content + barStyle.Render(strings.Repeat(" ", m.Width-w))
	}

	// --- line 1: focus indicator + navigation ---
	focusLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent).
		Background(barBG).
		Padding(0, 1).
		Render("[" + m.Focus.String() + "]")

	nav := strings.Join([]string{
		key("tab", "next panel"),
		key("shift+tab", "prev panel"),
	}, barStyle.Render("  "))

	line1 := focusLabel + barStyle.Render("  ") + nav

	// --- line 2: action keys ---
	actions := strings.Join([]string{
		key("n", "new"),
		key("s", "save"),
		key("d", "duplicate"),
		key("e", "export"),
		key("q", "quit"),
		key("?", "help"),
	}, barStyle.Render("  "))

	line2 := barStyle.Render(" ") + actions

	return padLine(line1) + "\n" + padLine(line2)
}
