package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	width       = 96
	White       = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "15"}
	LightGray   = lipgloss.CompleteColor{TrueColor: "#a3a3a3", ANSI256: "254", ANSI: "7"}
	Gray        = lipgloss.CompleteColor{TrueColor: "#939393", ANSI256: "250", ANSI: "8"}
	Black       = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	Purple      = lipgloss.CompleteColor{TrueColor: "#6612e3", ANSI256: "129", ANSI: "5"}
	LightPurple = lipgloss.CompleteColor{TrueColor: "#8347de", ANSI256: "98", ANSI: "13"}
	Emerald     = lipgloss.CompleteColor{TrueColor: "#00CC66", ANSI256: "41", ANSI: "2"}
	Rose        = lipgloss.CompleteColor{TrueColor: "#D90368", ANSI256: "161", ANSI: "1"}
	Yellow      = lipgloss.CompleteColor{TrueColor: "#F1C40F", ANSI256: "220", ANSI: "11"}
	Orange      = lipgloss.CompleteColor{TrueColor: "#F75C03", ANSI256: "202", ANSI: "3"}
)

func Active(s string) string {
	a := lipgloss.NewStyle().Foreground(Rose)
	return a.Render(s)
}

func StyledTable(hdrs []string, r [][]string, index int) *table.Table {
	var headers []string
	var rows [][]string

	// if headers pass limit
	// create new variables of slices up to limit
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
	keyStyle := lipgloss.NewStyle().Width(20).Align(lipgloss.Left, lipgloss.Top).Foreground(LightGray)
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
		Foreground(Purple)
	return headingStyle.Render(s)
}

func RenderBorderFlex(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Purple).
		Padding(1)
	return style.Render(s)
}
func RenderBlockFlex(s string) string {
	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Purple).
		Foreground(LightGray).
		Background(Rose).
        Padding(1)
	return blockStyle.Render(s)
}
func RenderBorderFixed(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Purple).
		Padding(1).
        Width(75).
        Height(20)
	return style.Render(s)
}
func RenderBlockFixed(s string) string {
	blockStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Purple).
		Foreground(LightGray).
		Background(Rose).
        Padding(1).
        Width(75).
        Height(20)
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
		Foreground(LightGray).
		Background(Rose).
		Padding(0, 1).
		MarginRight(1)

	encodingStyle := statusNugget.
		Background(LightPurple).
		Align(lipgloss.Right)

	statusText := lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle := statusNugget.Background(Purple)
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

	doc.WriteString(statusBarStyle.Width(m.width).Render(bar))
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
