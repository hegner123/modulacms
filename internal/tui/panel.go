package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
)

// FocusPanel identifies which panel currently has focus.
type FocusPanel int

const (
	TreePanel    FocusPanel = iota
	ContentPanel
	RoutePanel
)

// String returns the display name of the focused panel.
func (f FocusPanel) String() string {
	switch f {
	case TreePanel:
		return "Tree"
	case ContentPanel:
		return "Content"
	case RoutePanel:
		return "Route"
	default:
		return "Unknown"
	}
}

// Panel represents a bordered UI section with a title.
type Panel struct {
	Title   string
	Width   int
	Height  int
	Content string
	Focused bool
}

// Render draws the panel as a bordered box with a title bar.
func (p Panel) Render() string {
	borderColor := config.DefaultStyle.Tertiary
	if p.Focused {
		borderColor = config.DefaultStyle.Accent
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent)

	// Inner width is panel width minus border columns (2 chars for left+right border)
	innerWidth := p.Width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	// Inner height is panel height minus border rows (2 for top+bottom) minus title row
	innerHeight := p.Height - 3
	if innerHeight < 0 {
		innerHeight = 0
	}

	// Build title line
	title := titleStyle.Render(p.Title)

	// Pad or truncate content to fill the inner area
	content := padContent(p.Content, innerWidth, innerHeight)

	body := title + "\n" + content

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight + 1). // +1 for the title line
		Render(body)

	return box
}

// padContent ensures content fills exactly the given width and height.
func padContent(content string, width, height int) string {
	lines := strings.Split(content, "\n")

	// Truncate or pad lines to fill height
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	// Truncate each line to width
	for i, line := range lines {
		runeCount := len([]rune(line))
		if runeCount > width {
			lines[i] = string([]rune(line)[:width])
		}
	}

	return strings.Join(lines, "\n")
}
