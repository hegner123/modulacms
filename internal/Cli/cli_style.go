package cli

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	white     = lipgloss.Color("#ffffff")
	lightGray = lipgloss.Color("#7d7d7d")
	gray      = lipgloss.Color("#3d3d3d")
	black     = lipgloss.Color("#000000")

	purple  = lipgloss.Color("#6612e3")
	emerald = lipgloss.Color("#00CC66")
	rose    = lipgloss.Color("#D90368")
	yellow  = lipgloss.Color("#F1C40F")
	orange  = lipgloss.Color("#F75C03")

	statusNugget = lipgloss.NewStyle().
			Foreground(lightGray).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	encodingStyle = statusNugget.
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle = statusNugget.Background(lipgloss.Color("#6124DF"))

)

func StyledTable(hdrs []string, rows [][]string, index int) *table.Table {
	var (
		purple    = lipgloss.Color("#6612e3")
		gray      = lipgloss.Color("#3d3d3d")
		lightGray = lipgloss.Color("#7d7d7d")
		white     = lipgloss.Color("#ffffff")
		black     = lipgloss.Color("#000000")

		headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().Padding(0, 1).Width(8)
		activeStyle  = cellStyle.Background(white).Foreground(black).Bold(true)
		oddRowStyle  = cellStyle.Foreground(gray)
		evenRowStyle = cellStyle.Foreground(lightGray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
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
		Headers(hdrs...).
		Rows(rows...)

	return t
}

func RenderTitle(s string) string {
	purple := lipgloss.Color("#6612e3")
	white := lipgloss.Color("#ffffff")
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(white).
		Background(purple)

	return titleStyle.Render(s)
}

func RenderHeading(s string) string {
	purple := lipgloss.Color("#6612e3")
	headingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(purple).
        MaxWidth(20)
	return headingStyle.Render(s)
}

func RenderBorder(s string) string {
	purple := lipgloss.Color("#6612e3")
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(purple).
		Padding(1)
	return style.Render(s)
}

func RenderBlock(s string) string {
	purple := lipgloss.Color("#6612e3")

	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(purple).
		Width(24)
	return blockStyle.Render(s)

}

func RenderBorderBlock(s string) string {
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(orange).
		Padding(1)

	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(white).
		Background(emerald).
		Width(24)
	return borderStyle.Render(blockStyle.Render(s))

}
