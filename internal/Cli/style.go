package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	width     = 96
	White     = lipgloss.Color("#ffffff")
	LightGray = lipgloss.Color("#a3a3a3")
	Gray      = lipgloss.Color("#939393")
	Black     = lipgloss.Color("#000000")

	Purple      = lipgloss.Color("#6612e3")
	LightPurple = lipgloss.Color("#8347de")
	Emerald     = lipgloss.Color("#00CC66")
	Rose        = lipgloss.Color("#D90368")
	Yellow      = lipgloss.Color("#F1C40F")
	Orange      = lipgloss.Color("#F75C03")
)

func Active(s string) string {
	a := lipgloss.NewStyle().Foreground(Rose)
	return a.Render(s)
}

func StyledTable(hdrs []string, r [][]string, index int) *table.Table {
	var headers []string
	var rows [][]string
	max := 10
	if len(hdrs) > max {
		headers = hdrs[:max]
		for _, row := range r {
			ro := row[:max]
			rows = append(rows, ro)
		}
	} else {
		headers = hdrs
		rows = r
	}

	// if headers pass limit
	// create new variables of slices up to limit
	var (
		headerStyle  = lipgloss.NewStyle().Foreground(Purple).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().MaxWidth(10)
		activeStyle  = cellStyle.Background(LightGray).Foreground(Black).Bold(true)
		oddRowStyle  = cellStyle.Foreground(Gray)
		evenRowStyle = cellStyle.Foreground(LightGray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(Purple)).
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

func RenderEntry(h []string, e []string) string {
	var row []string
	keyStyle := lipgloss.NewStyle().Width(20).Align(lipgloss.Left,lipgloss.Top).Foreground(LightGray)
	valueStyle := lipgloss.NewStyle().Foreground(LightGray).Width(50)

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

func RenderTitle(s string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(Rose)

	return titleStyle.Render(s)
}
func RenderFooter(s string) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(Gray)
	return footerStyle.Render(s)
}

func RenderHeading(s string) string {
	headingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(LightPurple)
	return headingStyle.Render(s)
}

func RenderBorder(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Purple).
		Padding(1)
	return style.Render(s)
}

func RenderBlock(s string) string {
	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(Purple).
		Width(24)
	return blockStyle.Render(s)

}

func (m model) RenderStatusBar() string {
	doc := strings.Builder{}
	status := []string{"Page", "Form", "Dialog"}
	statusNugget := lipgloss.NewStyle().
		Foreground(LightGray).
		Padding(0, 1)

	statusBarStyle := lipgloss.NewStyle().
		Foreground(LightGray).
		Background(Rose)

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
	var v string

	v = m.page.Label
	if m.table != "" {
		v = m.table
	}

	statusKey := statusStyle.Render("STATUS")
	encoding := encodingStyle.Render("UTF-8")
	fishCake := fishCakeStyle.Render(v)
	statusVal := statusText.
		Width(width - w(statusKey) - w(encoding) - w(fishCake)).
		Render(status[m.focus])

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
		BorderForeground(Rose).
		Padding(1)

	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(Black).
		Background(Yellow).
		Width(24)
	return borderStyle.Render(blockStyle.Render(s))

}

func RenderButton(s string) string {
	style := lipgloss.NewStyle().
		Foreground(White).
		Background(Purple).
		Padding(1, 1).
		Margin(0, 1)
	return style.Render(s)
}

func RenderActiveButton(s string) string {
	style := lipgloss.NewStyle().
		Foreground(White).
		Background(Rose).
		Padding(1, 1).
		Margin(0, 1)
	return style.Render(s)
}
