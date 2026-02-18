package cli

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/hegner123/modulacms/internal/config"
)

// Active renders a string in the primary color.
func Active(s string) string {
	a := lipgloss.NewStyle().Foreground(config.DefaultStyle.Primary)
	return a.Render(s)
}

// TableRender creates a styled table with headers and rows, highlighting the active row.
func TableRender(hdrs []string, r [][]string, index int) *table.Table {
	var headers []string
	var rows [][]string

	// if headers pass limit
	// create new variables of slices up to limit
	max := 10
	if len(hdrs) > max {
		headers = hdrs[:max]
		for _, row := range r {
			// Use minimum of max and row length to avoid panic
			rowMax := max
			if len(row) < max {
				rowMax = len(row)
			}
			ro := row[:rowMax]
			rows = append(rows, ro)
		}
	} else {
		headers = hdrs
		rows = r
	}

	var (
		headerStyle  = lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().MaxWidth(10)
		activeStyle  = cellStyle.Foreground(config.DefaultStyle.Active).Background(config.DefaultStyle.ActiveBG).Bold(true)
		oddRowStyle  = cellStyle.Foreground(config.DefaultStyle.Secondary).Background(config.DefaultStyle.SecondaryBG)
		evenRowStyle = cellStyle.Foreground(config.DefaultStyle.Tertiary).Background(config.DefaultStyle.TertiaryBG)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(config.DefaultStyle.PrimaryBorder)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == index:
				return activeStyle
			case row == table.HeaderRow:
				return headerStyle
			case row%2 == 0:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		}).
		Headers(headers...).
		Rows(rows...)

	return t
}

// RenderEntry renders a list of key-value pairs vertically.
func RenderEntry(h []string, e []string) string {
	var row []string
	keyStyle := lipgloss.NewStyle().Width(20).Align(lipgloss.Left, lipgloss.Top).Foreground(config.DefaultStyle.Primary)
	valueStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Secondary).Width(50)

	for i, en := range e {
		s := lipgloss.JoinHorizontal(lipgloss.Center,
			keyStyle.Render(h[i]),
			valueStyle.Render(en),
		)
		row = append(row, s)
	}
	doc := lipgloss.JoinVertical(lipgloss.Top,
		row...,
	)

	return doc

}

// RenderTitle renders a string in the accent color.
func RenderTitle(s string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent2)

	return titleStyle.Render(s)
}

// RenderFooter renders a string in the secondary color.
func RenderFooter(s string) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary)
	return footerStyle.Render(s)
}

// RenderHeading renders a bold string in the secondary color.
func RenderHeading(s string) string {
	headingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Secondary)
	return headingStyle.Render(s)
}

// RenderBorderFlex renders a bordered string with flexible sizing.
func RenderBorderFlex(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.PrimaryBorder).
		Padding(1)
	return style.Render(s)
}

// RenderBlockFlex renders a bordered block with flexible sizing and background color.
func RenderBlockFlex(s string) string {
	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.PrimaryBorder).
		Foreground(config.DefaultStyle.Tertiary).
		Background(config.DefaultStyle.TertiaryBG).
		Padding(1)
	return blockStyle.Render(s)
}

// RenderBorderFixed renders a bordered string with fixed dimensions.
func RenderBorderFixed(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.PrimaryBorder).
		Padding(1).
		Width(75).
		Height(20)
	return style.Render(s)
}

// RenderBlockFixed renders a bordered block with fixed dimensions and background color.
func RenderBlockFixed(s string) string {
	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.Secondary).
		Foreground(config.DefaultStyle.Tertiary).
		Background(config.DefaultStyle.TertiaryBG).
		Padding(1).
		Width(75).
		Height(20)
	return blockStyle.Render(s)
}

// RenderSpace renders vertical space to fill remaining viewport height.
func (m Model) RenderSpace(content string) string {
	spaceStyle := lipgloss.NewStyle().Height(m.Height - lipgloss.Height(content) - 2)
	return spaceStyle.Render("")
}

// RenderStatusBar renders the application status bar.
func (m Model) RenderStatusBar() string {
	doc := strings.Builder{}
	status := []string{"Page", "Table", "Form", "Dialog"}
	statusNugget := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Status1).
		Padding(0, 1)

	statusBarStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Status2).
		Background(config.DefaultStyle.Status2BG)

	nuggetStyle := statusNugget.
		Background(config.DefaultStyle.Status3BG).
		Align(lipgloss.Right)

	statusText := lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle := statusNugget.Background(config.DefaultStyle.Status3BG).Foreground(config.DefaultStyle.Status3)
	var v string

	v = m.Page.Label
	if m.TableState.Table != "" {
		v = m.TableState.Table
	}

	statusKey := m.GetStatus()
	p := m.Page.Label
	cm := strconv.FormatInt(int64(m.CursorMax), 10)

	nugget := nuggetStyle.Render("Page: " + p + "  CursorMax: " + cm)
	fishCake := fishCakeStyle.Render(v)

	w := lipgloss.Width
	statusVal := statusText.
		Width(m.Width - w(statusKey) - w(nugget) - w(fishCake) - 34).
		Render(status[m.Focus])

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		nugget,
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			HorizontalSpace(34-(lipgloss.Width(fishCake))/2),
			fishCake,
			HorizontalSpace((34-lipgloss.Width(fishCake))/2),
		),
	)

	doc.WriteString(statusBarStyle.Width(m.Width).Render(bar))
	return statusBarStyle.Render(doc.String())

}

// HorizontalSpace returns a string of repeated spaces for layout.
func HorizontalSpace(i int) string {
	s := strings.Repeat(" ", i)
	return s
}

// RenderBorderBlock renders a string in a styled block.
func RenderBorderBlock(s string) string {
	borderStyle := lipgloss.NewStyle().
		BorderForeground(config.DefaultStyle.Tertiary)

	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Padding(1).
		Foreground(config.DefaultStyle.Secondary).
		Background(config.DefaultStyle.SecondaryBG).
		Width(24)
	return borderStyle.Render(blockStyle.Render(s))

}

