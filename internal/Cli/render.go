package cli

import (
	"embed"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

//go:embed titles
var TitleFile embed.FS

func (m model) RenderUI() string {
	doc := strings.Builder{}
	column := []string{}
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	physicalWidth, physicalHeight, _ := term.GetSize(os.Stdout.Fd())
	docStyle = docStyle.Width(physicalWidth).Height(physicalHeight)
	if m.footer == "" {
		m.footer = "\n\nPress q to quit.\n"
	}
	title := RenderTitle(m.titles[m.titleFont])
	header := RenderHeading(m.header)
	footer := RenderFooter(m.footer)
	column = append(column, title)
	column = append(column, header)
	column = append(column, m.body)
	column = append(column, footer)
	if m.verbose {
		column = append(column, m.RenderStatusTable())

	}

	doc.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		column...,
	),
	)

	return docStyle.Render(doc.String())
}
