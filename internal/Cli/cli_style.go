package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	width     = 96
	white     = lipgloss.Color("#ffffff")
	lightGray = lipgloss.Color("#7d7d7d")
	gray      = lipgloss.Color("#3d3d3d")
	black     = lipgloss.Color("#000000")

	purple  = lipgloss.Color("#6612e3")
	emerald = lipgloss.Color("#00CC66")
	rose    = lipgloss.Color("#D90368")
	yellow  = lipgloss.Color("#F1C40F")
	orange  = lipgloss.Color("#F75C03")
)

func Active(s string) string {
	a := lipgloss.NewStyle().Foreground(rose)
	return a.Render(s)
}

func StyledTable(hdrs []string, rows [][]string, index int) *table.Table {
	var (
		headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().Padding(0, 1).Width(8)
		activeStyle  = cellStyle.Background(white).Foreground(black).Bold(true)
		oddRowStyle  = cellStyle.Foreground(gray)
		evenRowStyle = cellStyle.Foreground(lightGray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(yellow)).
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
	white := lipgloss.Color("#ffffff")
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(white).
		Background(purple)

	return titleStyle.Render(s)
}

func RenderHeading(s string) string {
	headingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(purple).
		MaxWidth(20)
	return headingStyle.Render(s)
}

func RenderBorder(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(purple).
		Padding(1)
	return style.Render(s)
}

func RenderBlock(s string) string {
	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(purple).
		Width(24)
	return blockStyle.Render(s)

}

func RenderStatusBar() string {
	doc := strings.Builder{}
	statusNugget := lipgloss.NewStyle().
		Foreground(lightGray).
		Padding(0, 1)

	statusBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
		Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle := lipgloss.NewStyle().
		Inherit(statusBarStyle).
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#FF5F87")).
		Padding(0, 1).
		MarginRight(1)

	encodingStyle := statusNugget.
		Background(lipgloss.Color("#A550DF")).
		Align(lipgloss.Right)

	statusText := lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle := statusNugget.Background(lipgloss.Color("#6124DF"))
	w := lipgloss.Width

	statusKey := statusStyle.Render("STATUS")
	encoding := encodingStyle.Render("UTF-8")
	fishCake := fishCakeStyle.Render("🍥 Fish Cake")
	statusVal := statusText.
		Width(width - w(statusKey) - w(encoding) - w(fishCake)).
		Render("Ravishing")

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		encoding,
		fishCake,
	)

	doc.WriteString(statusBarStyle.Width(width).Render(bar))
	return statusBarStyle.Render(doc.String())

}

func RenderBorderBlock(s string) string {
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(rose).
		Padding(1)

	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(black).
		Background(yellow).
		Width(24)
	return borderStyle.Render(blockStyle.Render(s))

}

func RenderButton(s string) string {
	style := lipgloss.NewStyle().
        Foreground(white).
        Background(purple).
        Padding(2, 1).
        Margin(0, 1)
	return style.Render(s)
}
func RenderActiveButton(s string) string {
	style := lipgloss.NewStyle().
        Foreground(white).
        Background(rose).
        Padding(2, 1).
        Margin(0, 1)
	return style.Render(s)
}
