package cli

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

func (m model) RenderUI() string {
	doc := strings.Builder{}
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	physicalWidth, physicalHeight, _ := term.GetSize(os.Stdout.Fd())
	docStyle = docStyle.Width(physicalWidth).Height(physicalHeight)
	m.title = "ModulaCMS\n"
	m.footer += "\n\nPress q to quit.\n"
	header := RenderHeading(m.header)

	doc.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		m.title,
		header,
		m.body,
		m.footer,
		//m.RenderStatusTable(),
		//m.RenderStatusBar(),
	),
	)

	return docStyle.Render(doc.String())
}
