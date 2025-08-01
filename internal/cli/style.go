package cli

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/hegner123/modulacms/internal/config"
)

func Active(s string) string {
	a := lipgloss.NewStyle().Foreground(config.DefaultStyle.Primary)
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

func RenderTitle(s string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent2)

	return titleStyle.Render(s)
}
func RenderFooter(s string) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary)
	return footerStyle.Render(s)
}

func RenderHeading(s string) string {
	headingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Secondary)
	return headingStyle.Render(s)
}

func RenderBorderFlex(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.PrimaryBorder).
		Padding(1)
	return style.Render(s)
}
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
func RenderBorderFixed(s string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.PrimaryBorder).
		Padding(1).
		Width(75).
		Height(20)
	return style.Render(s)
}
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

func (m Model) RenderSpace(content string) string {
	spaceStyle := lipgloss.NewStyle().Height(m.Height - lipgloss.Height(content) - 2)
	return spaceStyle.Render("")
}

func (m Model) RenderStatusBar() string {
	doc := strings.Builder{}
	status := []string{"Page", "Form", "Dialog"}
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
	if m.Table != "" {
		v = m.Table
	}

	statusKey := m.GetStatus()
	c := strconv.FormatInt(int64(m.Cursor), 10)
	cm := strconv.FormatInt(int64(m.CursorMax), 10)

	nugget := nuggetStyle.Render("Cursor: " + c + "  CursorMax: " + cm)
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
func HorizontalSpace(i int) string {
	s := strings.Repeat(" ", i)
	return s
}

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

func RenderButton(s string) string {
	style := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary).
		Background(config.DefaultStyle.SecondaryBG).
		Padding(1, 1).
		Margin(0, 1)
	return style.Render(s)
}

func RenderActiveButton(s string) string {
	style := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Background(config.DefaultStyle.TertiaryBG).
		Padding(1, 1).
		Margin(0, 1)
	return style.Render(s)
}
