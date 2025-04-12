package cli

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
	config "github.com/hegner123/modulacms/internal/config"
)

//go:embed titles
var TitleFile embed.FS

func (m Model) RenderUI() string {
	app := strings.Builder{}
	column := []string{}
	docStyle := lipgloss.NewStyle()
	docStyle = docStyle.Width(m.width).Height(m.height)
	if m.footer == "" {
		m.footer = "Press q to quit."
	}

	title := RenderTitle(m.titles[m.titleFont])
	header := RenderHeading(m.header)
	footer := RenderFooter(m.footer)
	column = append(column, title)
	column = append(column, header)
	body := m.body
	if m.verbose {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.body, m.RenderStatusTable())

	}
	column = append(column, body)

	app.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		column...,
	))
	h := m.RenderSpace(app.String() + RenderFooter(m.footer))
	doc := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Padding(0, 2).Render(app.String()),
		h,
		footer,
		m.RenderStatusBar(),
	)

	return docStyle.Render(doc)
}

func formatJSON(b *config.Config) (string, error) {
	formatted, err := json.MarshalIndent(*b, "", "    ")
	if err != nil {
		return "", err
	}
	nulled := strings.ReplaceAll(string(formatted), "\"\",", "null")
	trimmed := strings.ReplaceAll(nulled, "\"", "")
	return string(trimmed), nil
}
